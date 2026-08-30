package validator

import (
	"context"
	"encoding/json"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"

	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// DistributedInferenceWorkloadWrapper implements WorkloadWrapper for DistributedInferenceWorkload
type DistributedInferenceWorkloadWrapper struct {
	*dinfw.DistributedInferenceWorkload
}

func (w *DistributedInferenceWorkloadWrapper) GetNamespace() string {
	return w.DistributedInferenceWorkload.Namespace
}

func (w *DistributedInferenceWorkloadWrapper) GetName() string {
	return w.DistributedInferenceWorkload.Name
}

func (w *DistributedInferenceWorkloadWrapper) GetUsage() interface{} {
	return w.DistributedInferenceWorkload.Spec.Usage
}

func (w *DistributedInferenceWorkloadWrapper) GetOwnerUID() string {
	if w.DistributedInferenceWorkload.Spec.UserId == nil {
		return ""
	}
	return *w.DistributedInferenceWorkload.Spec.UserId.Value
}

func (w *DistributedInferenceWorkloadWrapper) SetOwnerUID(uid string) {
	w.DistributedInferenceWorkload.Spec.UserId = &flds.StringField{Value: &uid}
}

func (w *DistributedInferenceWorkloadWrapper) Marshal() ([]byte, error) {
	return json.Marshal(w.DistributedInferenceWorkload)
}

func (w *DistributedInferenceWorkloadWrapper) Unmarshal(data []byte) error {
	return json.Unmarshal(data, w.DistributedInferenceWorkload)
}

func NewDistributedInferenceWorkloadWrapper() jobs.WorkloadWrapper {
	return &DistributedInferenceWorkloadWrapper{DistributedInferenceWorkload: &dinfw.DistributedInferenceWorkload{}}
}

// InferenceWorkloadOwnershipProtectionValidator acts as a router between inference and distributed inference workloads
type InferenceWorkloadOwnershipProtectionValidator struct {
	inferenceValidator            *jobs.WorkloadOwnershipProtectionValidator
	distributedInferenceValidator *jobs.WorkloadOwnershipProtectionValidator
}

func NewInferenceWorkloadOwnershipProtectionValidator(k8sClient client.Client, workloadControllerUID string) *InferenceWorkloadOwnershipProtectionValidator {
	return &InferenceWorkloadOwnershipProtectionValidator{
		inferenceValidator:            jobs.NewWorkloadOwnershipProtectionValidator(k8sClient, jobs.NewJobWorkloadWrapper, workloadControllerUID),
		distributedInferenceValidator: jobs.NewWorkloadOwnershipProtectionValidator(k8sClient, NewDistributedInferenceWorkloadWrapper, workloadControllerUID),
	}
}

func (iwopv *InferenceWorkloadOwnershipProtectionValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Kind.Kind {
	case constants.KindDistributedInferenceWorkload:
		return iwopv.distributedInferenceValidator.Handle(ctx, req)
	case constants.KindInferenceWorkload:
		return iwopv.inferenceValidator.Handle(ctx, req)
	default:
		// Allow unknown workload types
		return admission.ValidationResponse(true, "")
	}
}

func (iwopv *InferenceWorkloadOwnershipProtectionValidator) InjectDecoder(d *admission.Decoder) error {
	if err := iwopv.inferenceValidator.InjectDecoder(d); err != nil {
		return err
	}
	return iwopv.distributedInferenceValidator.InjectDecoder(d)
}

// InferenceWorkloadOwnershipMutator acts as a router between inference and distributed inference workloads
type InferenceWorkloadOwnershipMutator struct {
	inferenceMutator            *jobs.WorkloadOwnershipMutator
	distributedInferenceMutator *jobs.WorkloadOwnershipMutator
}

func NewInferenceWorkloadOwnershipMutator(k8sClient client.Client, workloadControllerUID string) *InferenceWorkloadOwnershipMutator {
	return &InferenceWorkloadOwnershipMutator{
		inferenceMutator:            jobs.NewWorkloadOwnershipMutator(k8sClient, jobs.NewJobWorkloadWrapper, workloadControllerUID),
		distributedInferenceMutator: jobs.NewWorkloadOwnershipMutator(k8sClient, NewDistributedInferenceWorkloadWrapper, workloadControllerUID),
	}
}

func (iwom *InferenceWorkloadOwnershipMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Kind.Kind {
	case constants.KindDistributedInferenceWorkload:
		return iwom.distributedInferenceMutator.Handle(ctx, req)
	case constants.KindInferenceWorkload:
		return iwom.inferenceMutator.Handle(ctx, req)
	default:
		// Allow unknown workload types
		return admission.ValidationResponse(true, "")
	}
}

func (iwom *InferenceWorkloadOwnershipMutator) InjectDecoder(d *admission.Decoder) error {
	if err := iwom.inferenceMutator.InjectDecoder(d); err != nil {
		return err
	}
	return iwom.distributedInferenceMutator.InjectDecoder(d)
}
