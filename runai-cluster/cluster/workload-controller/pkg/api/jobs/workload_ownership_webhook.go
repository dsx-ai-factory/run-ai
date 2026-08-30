package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/webhook"
	admissionv1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

//+kubebuilder:webhook:path=/mutate-workload-ownership,mutating=true,failurePolicy=fail,sideEffects=None,groups=run.ai,resources=trainingworkloads;interactiveworkloads;inferenceworkloads,verbs=create;update,versions=v2alpha1,name=runai-workload-ownership-mutation.kb.io,admissionReviewVersions=v1

// WorkloadWrapper abstracts workload operations for the ownership mutator
type WorkloadWrapper interface {
	GetNamespace() string
	GetName() string
	GetUsage() interface{}
	GetOwnerUID() string
	SetOwnerUID(uid string)
	Marshal() ([]byte, error)
	Unmarshal(data []byte) error
}

type WrapperCreator func() WorkloadWrapper

// JobWorkloadWrapper implements WorkloadWrapper for JobWorkload
type JobWorkloadWrapper struct {
	*JobWorkload
}

func (w *JobWorkloadWrapper) GetNamespace() string {
	return w.JobWorkload.Namespace
}

func (w *JobWorkloadWrapper) GetName() string {
	return w.JobWorkload.Name
}

func (w *JobWorkloadWrapper) GetUsage() interface{} {
	return w.JobWorkload.Spec.Usage
}

func (w *JobWorkloadWrapper) GetOwnerUID() string {
	if w.JobWorkload.Spec.UserId == nil {
		return ""
	}
	return *w.JobWorkload.Spec.UserId.Value
}

func (w *JobWorkloadWrapper) SetOwnerUID(uid string) {
	w.JobWorkload.Spec.UserId = &flds.StringField{Value: &uid}
}

func (w *JobWorkloadWrapper) Marshal() ([]byte, error) {
	return json.Marshal(w.JobWorkload)
}

func (w *JobWorkloadWrapper) Unmarshal(data []byte) error {
	return json.Unmarshal(data, w.JobWorkload)
}

func NewJobWorkloadWrapper() WorkloadWrapper {
	return &JobWorkloadWrapper{JobWorkload: &JobWorkload{}}
}

type WorkloadOwnershipMutator struct {
	client.Client
	createWrapper         WrapperCreator
	workloadControllerUID string
}

func NewWorkloadOwnershipMutator(k8sClient client.Client, createWrapper WrapperCreator, workloadControllerUID string) *WorkloadOwnershipMutator {
	return &WorkloadOwnershipMutator{
		Client:                k8sClient,
		workloadControllerUID: workloadControllerUID,
		createWrapper:         createWrapper,
	}
}

func (wom *WorkloadOwnershipMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	wrapper := wom.createWrapper()

	err := wrapper.Unmarshal(req.Object.Raw)
	if err != nil {
		log.Error(err, "Failed to unmarshal workload",
			"operation", req.Operation, "kind", req.Kind, "namespace", req.Namespace, "name", req.Name)
		return admission.Errored(http.StatusBadRequest, err)
	}

	if wrapper.GetUsage() != common.UsedForSubmit {
		// if it's a workload for templates - don't mutate UserId
		return admission.ValidationResponse(true, "")
	}

	log.V(5).Info("WorkloadOwnershipMutator: handling mutation of workload ownership",
		"namespace", wrapper.GetNamespace(), "name", wrapper.GetName(), "requestOperation", req.Operation)

	shouldMutate, err := wom.mutateWorkloadOwnerIfNeeded(ctx, wrapper, req)
	if err != nil {
		log.Error(err, "WorkloadOwnershipMutator: failed to mutate workload",
			"namespace", wrapper.GetNamespace(), "name", wrapper.GetName(), "requestOperation", req.Operation)
		return admission.Errored(http.StatusInternalServerError, err)
	}
	if !shouldMutate {
		return admission.Response{
			AdmissionResponse: admissionv1.AdmissionResponse{
				Allowed:   true,
				PatchType: nil,
			},
		}
	}

	marshaledWorkload, err := wrapper.Marshal()
	if err != nil {
		log.Error(err, "WorkloadOwnershipMutator: failed to json marshal workload",
			"namespace", wrapper.GetNamespace(), "name", wrapper.GetName())
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledWorkload)
}

func (wom *WorkloadOwnershipMutator) InjectDecoder(d *admission.Decoder) error {
	return nil
}

func (wom *WorkloadOwnershipMutator) mutateWorkloadOwnerIfNeeded(
	ctx context.Context,
	wrapper WorkloadWrapper,
	req admission.Request) (bool, error) {
	requesterUID := webhook.GetRequesterUserId(req.UserInfo)

	// some requests are being made by the workload-controller itself,
	// in those requests, the requesterUID equals the workloadControllerUID.
	// we don't want to mutate the workload in this case.
	if requesterUID == wom.workloadControllerUID {
		return false, nil
	}

	if req.Operation == admissionv1.Create {
		return wom.mutateWorkloadOwnerOnCreate(ctx, wrapper, requesterUID)
	} else if req.Operation == admissionv1.Update {
		return wom.mutateWorkloadOwnerOnUpdate(ctx, wrapper, req)
	}

	// shouldn't happen - as we have the webhook only on CREATE and UPDATE operations
	return false, fmt.Errorf("unsupported operation")
}

func (wom *WorkloadOwnershipMutator) mutateWorkloadOwnerOnCreate(ctx context.Context, wrapper WorkloadWrapper, requesterUID string) (bool, error) {
	workloadOwnerUID := wrapper.GetOwnerUID()
	if workloadOwnerUID == requesterUID {
		return false, nil
	}

	log.V(5).Info("Populating UserId on CREATE",
		"updatingToRequesterUID", requesterUID, "originalWorkloadUserId", workloadOwnerUID)

	// populate the UserId field with the correct value
	wrapper.SetOwnerUID(requesterUID)
	return true, nil
}

func (wom *WorkloadOwnershipMutator) mutateWorkloadOwnerOnUpdate(ctx context.Context, wrapper WorkloadWrapper, req admission.Request) (bool, error) {
	oldWrapper := wom.createWrapper()

	err := oldWrapper.Unmarshal(req.OldObject.Raw)
	if err != nil {
		log.Error(err, "Failed to unmarshal old workload",
			"operation", req.Operation, "kind", req.Kind, "namespace", req.Namespace, "name", req.Name)
		return false, err
	}

	workloadOwnerUID := wrapper.GetOwnerUID()
	oldWorkloadOwnerUID := oldWrapper.GetOwnerUID()

	if workloadOwnerUID == oldWorkloadOwnerUID {
		return false, nil
	}

	log.V(2).Info("Warning: Populating UserId on UPDATE",
		"namespace", wrapper.GetNamespace(), "name", wrapper.GetName(),
		"updatingToRequesterUID", oldWorkloadOwnerUID, "workload.Spec.UserId", workloadOwnerUID)

	// populate the UserId field with the correct value - the requester ID from the Old workload object
	wrapper.SetOwnerUID(oldWorkloadOwnerUID)
	return true, nil
}

