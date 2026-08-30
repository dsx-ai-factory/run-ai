package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/run-ai/runai/runai-cluster/cluster/common/utils/validator"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/webhook/validator/workloads"
	"gomodules.xyz/jsonpatch/v2"
	admissionV1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/logging"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	namePath            = "/metadata/name"
	namespacePath       = "/metadata/namespace"
	metadataPrefix      = "/metadata"
	statusPrefix        = "/status"
	kaiAnnotationPrefix = "/metadata/annotations/kai.scheduler"

	// inference fields
	commandField           = "command"
	argumentsField         = "arguments"
	environmentField       = "environment"
	imageField             = "image"
	imagePullPolicyField   = "imagePullPolicy"
	workingDirField        = "workingDir"
	createHomeDirField     = "createHomeDir"
	readinessProbeDirField = "readinessProbe"

	// inference auto-scaling fields
	autoScalingFieldMinScale                     = "minScale"
	autoScalingFieldMaxScale                     = "maxScale"
	autoScalingFieldInitialScale                 = "initialScale"
	autoScalingFieldActivationScale              = "activationScale"
	autoScalingFieldMetric                       = "metric"
	autoScalingFieldMetricName                   = "metricName"
	autoScalingFieldTarget                       = "target"
	autoScalingFieldTargetUtilizationPercentage  = "targetUtilizationPercentage"
	autoScalingFieldConcurrencyHardLimit         = "concurrencyHardLimit"
	autoScalingFieldScaleToZeroRetentionSeconds  = "scaleToZeroRetentionSeconds"
	autoScalingFieldScaleDownDelaySeconds        = "scaleDownDelaySeconds"
	autoScalingFieldInitializationTimeoutSeconds = "initializationTimeoutSeconds"

	// serving configuration fields
	servingConfigurationInitializationTimeoutSeconds = "initializationTimeoutSeconds"
	servingConfigurationRequestTimeoutSeconds        = "requestTimeoutSeconds"

	// compute fields
	gpuField               = "gpu"
	gpuDevicesField        = "gpuDevices"
	gpuLimitField          = "gpuLimit"
	gpuMemoryField         = "gpuMemory"
	gpuMemoryLimitField    = "gpuMemoryLimit"
	cpuField               = "cpu"
	cpuLimitField          = "cpuLimit"
	memoryField            = "memory"
	memoryLimitField       = "memoryLimit"
	extendedResourcesField = "extendedResources"
	largeShmField          = "largeShm"

	// resources fields
	nodeTypeField                  = "nodeType"
	nodePoolsField                 = "nodePools"
	nodeAffinityRequiredField      = "nodeAffinityRequired"
	podAffinityField               = "podAffinity"
	podAffinitySchedulingRuleField = "podAffinitySchedulingRule"
	podAffinityTopologyField       = "podAffinityTopology"
)

var (
	inferenceAllowedPrefixes = []string{
		metadataPrefix, // Allow changes to metadata (except name and namespace)
		statusPrefix,   // Allow changes to status
	}

	inferenceAllowedSpecFields = []string{
		// container fields
		commandField,
		argumentsField,
		environmentField,
		imageField,
		imagePullPolicyField,
		workingDirField,
		createHomeDirField,
		readinessProbeDirField,

		// auto-scaling fields
		autoScalingFieldMinScale,
		autoScalingFieldMaxScale,
		autoScalingFieldInitialScale,
		autoScalingFieldActivationScale,
		autoScalingFieldMetric,
		autoScalingFieldMetricName,
		autoScalingFieldTarget,
		autoScalingFieldTargetUtilizationPercentage,
		autoScalingFieldConcurrencyHardLimit,
		autoScalingFieldScaleToZeroRetentionSeconds,
		autoScalingFieldScaleDownDelaySeconds,
		autoScalingFieldInitializationTimeoutSeconds,

		// serving configuration fields
		servingConfigurationRequestTimeoutSeconds,
		servingConfigurationInitializationTimeoutSeconds,

		// compute fields
		gpuField,
		gpuDevicesField,
		gpuLimitField,
		gpuMemoryField,
		gpuMemoryLimitField,
		cpuField,
		cpuLimitField,
		memoryField,
		memoryLimitField,
		extendedResourcesField,
		largeShmField,

		// resources fields
		nodeTypeField,
		nodePoolsField,
		nodeAffinityRequiredField,
		podAffinityField,
		podAffinitySchedulingRuleField,
		podAffinityTopologyField,
	}

	inferencePreventPrefixes = []string{kaiAnnotationPrefix}
)

type InferenceValidatorHandler struct {
	k8sClient              client.Client
	validators             []workloads.WorkloadValidator
	allowedServiceAccounts []string
}

func NewInferenceValidator(k8sClient client.Client, allowedServiceAccounts []string, validators ...workloads.WorkloadValidator) *InferenceValidatorHandler {
	var workloadValidators []workloads.WorkloadValidator
	workloadValidators = append(workloadValidators, validators...)
	return &InferenceValidatorHandler{
		k8sClient:              k8sClient,
		validators:             workloadValidators,
		allowedServiceAccounts: allowedServiceAccounts,
	}
}

func (wv *InferenceValidatorHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Operation {
	case admissionV1.Create:
		return wv.HandleCreate(ctx, req.Object, req.RequestKind.Kind, req.UserInfo.Username)
	case admissionV1.Update:
		return wv.HandleUpdate(ctx, req.OldObject, req.Object, req.RequestKind.Kind, req.UserInfo.Username)
	default:
		// in the case of delete, we don't need to do anything
		return admission.Allowed(validOk)
	}
}

func (wv *InferenceValidatorHandler) HandleCreate(ctx context.Context, newObject runtime.RawExtension, workloadKind string, username string) admission.Response {
	log := logging.FromContext(ctx)

	workload := jobs.JobWorkload{}
	err := json.Unmarshal(newObject.Raw, &workload)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	err = enforceOwnership(workloadKind, username, workload.Labels, wv.allowedServiceAccounts)
	if err != nil {
		return admission.Errored(http.StatusUnauthorized, err)
	}

	// check if there is service with resource name with the same name, if so, deny
	err = workloadResourceLookupFound(ctx, wv.k8sClient, workload.Name, workload.Namespace)
	if err != nil {
		log.Error(err, "admission failed due resource lookup result")
		return admission.Denied(err.Error())
	}

	for _, validator := range wv.validators {
		err = validator.Validate(ctx, &workload)
		if err != nil {
			return admission.Denied(err.Error())
		}
	}

	return admission.Allowed("Ok")
}

func (wv *InferenceValidatorHandler) InjectDecoder(_ *admission.Decoder) error {
	return nil
}

func (wv *InferenceValidatorHandler) HandleUpdate(ctx context.Context, oldObject runtime.RawExtension, newObject runtime.RawExtension,
	workloadKind string, username string) admission.Response {

	log := logf.FromContext(ctx)

	//
	//   first, let's do the validations on the new workload. this may be sufficient so we won't need
	//   to unmarshal the old object and save time
	//
	newWorkload := jobs.JobWorkload{}

	err := json.Unmarshal(newObject.Raw, &newWorkload)
	if err != nil {
		log.Error(err, "cannot unmarshal new object")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// if the new workload is marked with label of run.ai/workload we allow it only
	// to system service account
	err = enforceOwnership(workloadKind, username, newWorkload.Labels, wv.allowedServiceAccounts)
	if err != nil {
		return admission.Errored(http.StatusUnauthorized, err)
	}

	isTemplate := isUsageTemplate(&newWorkload)

	// check if we need the old workload. we need it in the following cases
	// - if the user is not system service account, we need it to validate ownership based on the labels
	//   of the old workload
	// - if the new workload is a template, we need to check if the old object is also a template. if both
	//   are templates, no need to perform some of the validations that we do if old/new workload is not a template
	needToCheckOldWorkload := !validator.AllowedServiceAccounts(username, wv.allowedServiceAccounts) || isTemplate

	if needToCheckOldWorkload {
		// arriving here, the information from the new workload is not sufficient for us
		// and we have to check also the old object
		oldWorkload := jobs.JobWorkload{}

		err = json.Unmarshal(oldObject.Raw, &oldWorkload)
		if err != nil {
			log.Error(err, "cannot unmarshal old object")
			return admission.Errored(http.StatusBadRequest, err)
		}

		err = enforceOwnership(workloadKind, username, oldWorkload.Labels, wv.allowedServiceAccounts)
		if err != nil {
			return admission.Errored(http.StatusUnauthorized, err)
		}

		isTemplate = isTemplate && isUsageTemplate(&oldWorkload)
	}

	if isTemplate {
		// support workload template is going to be deprecated soon
		// for now, we allow all changes to workload template
		return admission.Allowed(validOk)
	}

	patches, err := jsonpatch.CreatePatch(oldObject.Raw, newObject.Raw)
	if err != nil {
		log.Error(err, "cannot compare old and new objects")
		return admission.Errored(http.StatusServiceUnavailable, err)
	}
	for _, change := range patches {
		if !isAllowedFieldChange(change) {
			return admission.Denied(fmt.Sprintf("cannot update %s, field is immutable", change.Path))
		}
	}

	for _, validator := range wv.validators {
		err = validator.Validate(ctx, &newWorkload)
		if err != nil {
			return admission.Denied(err.Error())
		}
	}

	return admission.Allowed(validOk)
}

func workloadResourceLookupFound(ctx context.Context, k8sClient client.Client, workloadName string, workloadNamespace string) error {
	labels := map[string]string{constants.LabelWorkloadName: workloadName}
	serviceList := &v1.ServiceList{}

	err := k8sClient.List(ctx, serviceList,
		client.InNamespace(workloadNamespace), client.MatchingLabels(labels))
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if len(serviceList.Items) != 0 {
		return fmt.Errorf("a workload with the name %s already exists, please choose a different name for your workload", workloadName)
	}

	return nil
}

func isAllowedFieldChange(change jsonpatch.Operation) bool {
	// Prevent changes
	for _, prefix := range inferencePreventPrefixes {
		if strings.HasPrefix(change.Path, prefix) {
			return false
		}
	}

	// Allow changes to:
	// - Any field within /metadata, except for /metadata/name and /metadata/namespace
	// - Any field within /status
	// - Specific fields within /spec related to inference editing and auto-scaling

	for _, prefix := range inferenceAllowedPrefixes {
		if strings.HasPrefix(change.Path, prefix) {
			// Disallow changes to /metadata/name and /metadata/namespace
			if change.Path == namePath || change.Path == namespacePath {
				return false
			}
			return true
		}
	}

	for _, specField := range inferenceAllowedSpecFields {
		// Fields in /spec should have the exact name, but we are also using prefix check as the json name may look like
		// "/spec/field-name/value" and not just "/spec/field-name". We're adding "/" in the end of the prefix, to
		// prevent that allowing field name "x" won't allow field name "x2" as well.
		exactPath := "/spec/" + specField
		prefix := exactPath + "/"
		if change.Path == exactPath || strings.HasPrefix(change.Path, prefix) {
			return true
		}
	}

	return false
}

func enforceOwnership(workloadKind string, requesterUsername string, labels map[string]string, allowedServiceAccounts []string) error {
	//
	//   we apply service account ownership protection only for run-ai workloads
	//
	_, runaiWorkload := labels[constants.WorkloadLabel]
	if !runaiWorkload {
		return nil
	}

	switch workloadKind {
	case constants.KindInferenceWorkload, constants.KindDistributedInferenceWorkload:
		if !validator.AllowedServiceAccounts(requesterUsername, allowedServiceAccounts) {
			return fmt.Errorf("username %s is not authorized to perform this action on %s", requesterUsername, workloadKind)
		}
	}

	return nil
}

func isUsageTemplate(workload *jobs.JobWorkload) bool {
	return workload != nil && workload.Spec.Usage == common.UsedAsTemplate
}
