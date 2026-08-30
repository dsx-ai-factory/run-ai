package restart

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	admissionv1 "k8s.io/api/admission/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
)

// RestartRequestValidatorPath is the path the runai-restartrequest-validation webhook points at.
const RestartRequestValidatorPath = "/validate-restartrequest"

// allowedOverridePath is the only spec override target a RestartRequest may carry
const allowedOverridePath = "/spec/instructions/nodeAffinity"

// supportedGVK is the only workload kind this controller restarts.
var supportedGVK = metav1.GroupVersionKind{
	Group:   dromosv1alpha1.GroupVersion.Group,
	Version: dromosv1alpha1.GroupVersion.Version,
	Kind:    dromosv1alpha1.WorkloadRunnerKind,
}

var rrValidatorLogger = logger.Log.WithName("webhook.restartrequest")

// RestartRequestValidator admits RestartRequests targeting a WorkloadRunner.
type RestartRequestValidator struct {
	client client.Client
}

// NewRestartRequestValidator creates a new RestartRequestValidator.
func NewRestartRequestValidator(client client.Client) *RestartRequestValidator {
	return &RestartRequestValidator{client: client}
}

func (v *RestartRequestValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Operation {
	case admissionv1.Create:
		return v.handleCreate(ctx, req)
	case admissionv1.Update:
		return v.handleUpdate(req)
	default:
		return admission.Allowed("Ok")
	}
}

func (v *RestartRequestValidator) handleUpdate(req admission.Request) admission.Response {
	oldRR := &runv1alpha1.RestartRequest{}
	if err := json.Unmarshal(req.OldObject.Raw, oldRR); err != nil {
		rrValidatorLogger.Error(err, "failed to unmarshal old RestartRequest")
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to unmarshal old RestartRequest: %w", err))
	}
	newRR := &runv1alpha1.RestartRequest{}
	if err := json.Unmarshal(req.Object.Raw, newRR); err != nil {
		rrValidatorLogger.Error(err, "failed to unmarshal new RestartRequest")
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to unmarshal new RestartRequest: %w", err))
	}
	if !reflect.DeepEqual(oldRR.Spec, newRR.Spec) {
		return admission.Denied("RestartRequest spec is immutable")
	}
	return admission.Allowed("Ok")
}

func (v *RestartRequestValidator) handleCreate(ctx context.Context, req admission.Request) admission.Response {
	rr := &runv1alpha1.RestartRequest{}
	if err := json.Unmarshal(req.Object.Raw, rr); err != nil {
		rrValidatorLogger.Error(err, "failed to unmarshal RestartRequest")
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to unmarshal RestartRequest: %w", err))
	}

	if resp := validateWorkloadNameLabel(rr); !resp.Allowed {
		return resp
	}

	if resp := validateGVK(&rr.Spec.WorkloadRef); !resp.Allowed {
		return resp
	}

	workload, resp := v.validateWorkloadExists(ctx, &rr.Spec.WorkloadRef, req.Namespace)
	if !resp.Allowed {
		return resp
	}

	if resp = validateGeneration(&rr.Spec.WorkloadRef, workload.Generation); !resp.Allowed {
		return resp
	}

	if resp = validateSpecOverridePaths(rr.Spec.SpecOverrides); !resp.Allowed {
		return resp
	}

	if resp = v.validateNoActiveRestartRequest(ctx, &rr.Spec.WorkloadRef, req.Namespace); !resp.Allowed {
		return resp
	}

	return admission.Allowed("Ok")
}

func (v *RestartRequestValidator) validateWorkloadExists(ctx context.Context, ref *runv1alpha1.WorkloadReference,
	namespace string) (*dromosv1alpha1.WorkloadRunner, admission.Response) {
	workload := &dromosv1alpha1.WorkloadRunner{}
	if err := v.client.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: namespace}, workload); err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, admission.Denied(fmt.Sprintf("referenced workload %q not found in namespace %q", ref.Name, namespace))
		}
		return nil, admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed to get referenced workload: %w", err))
	}
	return workload, admission.Allowed("")
}

// validateNoActiveRestartRequest keeps at most one active RestartRequest per workload
func (v *RestartRequestValidator) validateNoActiveRestartRequest(ctx context.Context,
	ref *runv1alpha1.WorkloadReference, namespace string) admission.Response {
	list := &runv1alpha1.RestartRequestList{}
	if err := v.client.List(ctx, list,
		client.InNamespace(namespace),
		client.MatchingLabels{constants.RunaiWorkloadNameLabel: ref.Name},
	); err != nil {
		rrValidatorLogger.Error(err, "failed to list RestartRequests")
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed to list RestartRequests: %w", err))
	}

	for i := range list.Items {
		if isActive(list.Items[i].Status.Phase) {
			return admission.Denied(fmt.Sprintf("active RestartRequest already exists: %s", list.Items[i].Name))
		}
	}
	return admission.Allowed("")
}

func validateGVK(ref *runv1alpha1.WorkloadReference) admission.Response {
	if ref.GVK != supportedGVK {
		return admission.Denied(fmt.Sprintf("unsupported workload GVK: %s/%s/%s", ref.GVK.Group, ref.GVK.Version, ref.GVK.Kind))
	}
	return admission.Allowed("")
}

// validateWorkloadNameLabel requires the label the HandoffReconciler's mapper selects by
func validateWorkloadNameLabel(rr *runv1alpha1.RestartRequest) admission.Response {
	val, ok := rr.Labels[constants.RunaiWorkloadNameLabel]
	if !ok {
		return admission.Denied(fmt.Sprintf("label %q is required", constants.RunaiWorkloadNameLabel))
	}
	if val != rr.Spec.WorkloadRef.Name {
		return admission.Denied(fmt.Sprintf("label %q value %q must match workloadRef.name %q",
			constants.RunaiWorkloadNameLabel, val, rr.Spec.WorkloadRef.Name))
	}
	return admission.Allowed("")
}

// validateGeneration pins the generation the RestartReconciler fences on.
func validateGeneration(ref *runv1alpha1.WorkloadReference, workloadGeneration int64) admission.Response {
	if ref.Generation != workloadGeneration {
		return admission.Denied(fmt.Sprintf("workloadRef.generation %d does not match workload generation %d",
			ref.Generation, workloadGeneration))
	}
	return admission.Allowed("")
}

func validateSpecOverridePaths(overrides runv1alpha1.SpecOverrides) admission.Response {
	for _, op := range overrides {
		if op.Path != allowedOverridePath {
			return admission.Denied(fmt.Sprintf("spec override path %q is not allowed; allowed paths: [%s]", op.Path, allowedOverridePath))
		}
	}
	return admission.Allowed("")
}

func isActive(phase runv1alpha1.RestartRequestPhase) bool {
	return phase == "" || phase == runv1alpha1.RestartRequestPhasePending || phase == runv1alpha1.RestartRequestPhaseInProgress
}
