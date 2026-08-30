package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"gomodules.xyz/jsonpatch/v2"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	admissionV1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/logging"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	validOk = "Ok"

	// distributed inference fields
	replicasField = "replicas"
)

var (
	distInferenceAllowedPrefixes   = []string{metadataPrefix, statusPrefix}
	distInferenceAllowedSpecFields = []string{replicasField}
	distInferencePreventPrefixes   = []string{kaiAnnotationPrefix}
)

type DistributedInferenceValidator interface {
	Validate(_ context.Context, job *dinfw.DistributedInferenceWorkload) error
}

type DistributedInferenceValidatorHandler struct {
	k8sClient              client.Client
	validators             []DistributedInferenceValidator
	allowedServiceAccounts []string
	elasticEnabled         bool
}

func NewDistributedInferenceValidator(k8sClient client.Client, allowedServiceAccounts []string, elasticEnabled bool, validators ...DistributedInferenceValidator) *DistributedInferenceValidatorHandler {
	return &DistributedInferenceValidatorHandler{
		k8sClient:              k8sClient,
		validators:             validators,
		allowedServiceAccounts: allowedServiceAccounts,
		elasticEnabled:         elasticEnabled,
	}
}

func (dv *DistributedInferenceValidatorHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Operation {
	case admissionV1.Create:
		return dv.HandleCreate(ctx, req.Object, req.RequestKind.Kind, req.UserInfo.Username)
	case admissionV1.Update:
		return dv.HandleUpdate(ctx, req.OldObject, req.Object, req.RequestKind.Kind, req.UserInfo.Username)
	default:
		// in the case of delete, we don't need to do anything
		return admission.Allowed(validOk)
	}
}

func (dv *DistributedInferenceValidatorHandler) HandleCreate(ctx context.Context, newObject runtime.RawExtension, workloadKind string, username string) admission.Response {
	log := logging.FromContext(ctx)

	if workloadKind != constants.KindDistributedInferenceWorkload {
		return admission.Allowed(validOk)
	}

	workload := dinfw.DistributedInferenceWorkload{}
	err := json.Unmarshal(newObject.Raw, &workload)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// check if the user is authorized to create the workload
	err = enforceOwnership(workloadKind, username, workload.Labels, dv.allowedServiceAccounts)
	if err != nil {
		return admission.Errored(http.StatusUnauthorized, err)
	}

	// check if there is service with resource name with the same name, if so, deny
	err = workloadResourceLookupFound(ctx, dv.k8sClient, workload.Name, workload.Namespace)
	if err != nil {
		log.Error(err, "admission failed due resource lookup result")
		return admission.Denied(err.Error())
	}

	// run validators
	for _, validator := range dv.validators {
		err = validator.Validate(ctx, &workload)
		if err != nil {
			return admission.Denied(err.Error())
		}
	}

	return admission.Allowed(validOk)
}

func (dv *DistributedInferenceValidatorHandler) HandleUpdate(ctx context.Context, oldObject runtime.RawExtension, newObject runtime.RawExtension,
	workloadKind string, username string) admission.Response {
	log := logging.FromContext(ctx)

	newWorkload := dinfw.DistributedInferenceWorkload{}
	err := json.Unmarshal(newObject.Raw, &newWorkload)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// only system service account can update the workload
	err = enforceOwnership(workloadKind, username, newWorkload.Labels, dv.allowedServiceAccounts)
	if err != nil {
		return admission.Errored(http.StatusUnauthorized, err)
	}

	// only some fields are allowed to be updated
	oldWorkload := dinfw.DistributedInferenceWorkload{}
	err = json.Unmarshal(oldObject.Raw, &oldWorkload)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	patches, err := jsonpatch.CreatePatch(oldObject.Raw, newObject.Raw)
	if err != nil {
		log.Error(err, "cannot compare old and new objects")
		return admission.Errored(http.StatusServiceUnavailable, err)
	}
	for _, change := range patches {
		if !isAllowedFieldChangeDistributedInference(change) {
			return admission.Denied(fmt.Sprintf("cannot update %s, field is immutable", change.Path))
		}

		// replicas field is immutable if mnnvl is enabled and elastic support is disabled
		if blockReplicaUpdate(change, oldWorkload.Labels, dv.elasticEnabled) {
			return admission.Denied(fmt.Sprintf("cannot update %s, workload cannot be scaled when %s is enabled", change.Path, mnnvl.UserFacingFeatureName))
		}
	}

	// run validators
	for _, validator := range dv.validators {
		err = validator.Validate(ctx, &newWorkload)
		if err != nil {
			return admission.Denied(err.Error())
		}
	}

	return admission.Allowed(validOk)
}

// matchesSpecFieldPath checks if a change path matches a specific spec field using two-fold comparison:
// 1. Exact match against "/spec/fieldName"
// 2. Prefix match against "/spec/fieldName/"
// This is necessary because the json path may look like "/spec/field-name/value" and not just "/spec/field-name".
// Adding "/" at the end of the prefix prevents field name "x" from matching field name "x2" as well.
func matchesSpecFieldPath(changePath, specField string) bool {
	exactPath := "/spec/" + specField
	prefix := exactPath + "/"
	return changePath == exactPath || strings.HasPrefix(changePath, prefix)
}

func isAllowedFieldChangeDistributedInference(change jsonpatch.Operation) bool {
	// Prevent changes
	for _, prefix := range distInferencePreventPrefixes {
		if strings.HasPrefix(change.Path, prefix) {
			return false
		}
	}

	// Allow changes to:
	// - Any field within /metadata, except for /metadata/name and /metadata/namespace
	// - Any field within /status
	// - Specific fields within /spec related to distributed inference editing
	for _, prefix := range distInferenceAllowedPrefixes {
		if strings.HasPrefix(change.Path, prefix) {
			// Disallow changes to /metadata/name and /metadata/namespace
			if change.Path == namePath || change.Path == namespacePath {
				return false
			}
			return true
		}
	}

	for _, specField := range distInferenceAllowedSpecFields {
		if matchesSpecFieldPath(change.Path, specField) {
			return true
		}
	}

	return false
}

func blockReplicaUpdate(change jsonpatch.Operation, labels map[string]string, elasticEnabled bool) bool {
	if !matchesSpecFieldPath(change.Path, replicasField) {
		return false
	}

	// Allow replica changes when elastic support is enabled
	if elasticEnabled {
		return false
	}

	mnnvlMode, ok := labels[mnnvl.FeatureLabelKey]
	if !ok || mnnvlMode == mnnvl.None {
		return false
	}
	return true
}

func (dv *DistributedInferenceValidatorHandler) InjectDecoder(_ *admission.Decoder) error {
	return nil
}
