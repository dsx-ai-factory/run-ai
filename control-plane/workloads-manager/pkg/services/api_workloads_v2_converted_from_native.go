package services

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"

	manifestconstants "github.com/run-ai/runai/control-plane/workloads-manager/pkg/constants"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

// ConvertedFromNativeLabel marks a workload that was created via the v1 native API
// and later cast to v2 by the control-plane composer.
// Source of truth: control-plane/workloads-and-policies/composer/composer_v1/iw_tw_workloads/iw_tw_to_job.go.
const ConvertedFromNativeLabel = "run.ai/converted-from-native"

// immutableV1WorkloadKinds is the set of workloadKind label values written by the
// v1->v2 composer for v1 workload types that were immutable in v1. Updates to
// workloads carrying ConvertedFromNativeLabel=true and one of these workloadKind
// values are rejected by the v2 update endpoint. Suspend/delete remain allowed
// via their dedicated endpoints. Inference and DistributedInference are
// intentionally absent: those v1 types are mutable and must remain so in v2.
var immutableV1WorkloadKinds = map[string]struct{}{
	"Workspace":   {},
	"Training":    {},
	"Distributed": {},
}

// isConvertedFromNativeWorkload reports whether the given labels mark a workload
// that was created via the v1 native API and cast to v2 (i.e. translated from a
// v1 native type), based on the ConvertedFromNativeLabel.
func isConvertedFromNativeWorkload(labels map[string]string) bool {
	return labels[ConvertedFromNativeLabel] == TrueAsString
}

// nativeWorkloadTypeFromV1Cast returns the original v1 workload type written by
// the v1-to-v2 composer, for example "Training" for a converted batch/v1 Job.
func nativeWorkloadTypeFromV1Cast(translatedFromV1 bool, manifest map[string]any) (string, bool) {
	if !translatedFromV1 || manifest == nil {
		return "", false
	}

	info, err := NewK8sResourceUtil().GetResourceInfo(manifest)
	if err != nil || info == nil || info.Labels == nil {
		return "", false
	}
	if info.Labels[ConvertedFromNativeLabel] != TrueAsString {
		return "", false
	}

	workloadType := info.Labels[constants.LabelWorkloadKind]
	return workloadType, workloadType != ""
}

func distributedFrameworkFromKind(kind string) string {
	switch kind {
	case constants.KindMpiJob:
		return string(workloadsApi.DISTRIBUTEDFRAMEWORK_MPI)
	case constants.KindPyTorchJob:
		return string(workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH)
	case constants.KindTFJob:
		return string(workloadsApi.DISTRIBUTEDFRAMEWORK_TF)
	case constants.KindXGBoostJob:
		return string(workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST)
	case constants.KindJAXJob:
		return string(workloadsApi.DISTRIBUTEDFRAMEWORK_JAX)
	default:
		return ""
	}
}

// isImmutableV1CastWorkload returns (true, workloadKind) when the stored workload
// was converted from a v1 native type that is immutable in v1; otherwise (false, "").
// The translatedFromV1 flag is the trusted source from the database record.
func isImmutableV1CastWorkload(translatedFromV1 bool, rawManifest map[string]any, k8sUtil *K8sResourceUtil) (bool, string) {
	if !translatedFromV1 || rawManifest == nil {
		return false, ""
	}

	info, err := k8sUtil.GetResourceInfo(rawManifest)
	if err != nil || info == nil || info.Labels == nil {
		return false, ""
	}

	workloadKind := info.Labels[constants.LabelWorkloadKind]
	if _, ok := immutableV1WorkloadKinds[workloadKind]; ok {
		return true, workloadKind
	}

	return false, ""
}

// stripConvertedFromNativeLabel removes the ConvertedFromNativeLabel from the manifest
// to prevent clients from forging it on public API requests.
func stripConvertedFromNativeLabel(manifest map[string]any) {
	if manifest == nil {
		return
	}

	metadata, ok := manifest[manifestconstants.FieldMetadata].(map[string]any)
	if !ok {
		return
	}

	labels, ok := metadata[manifestconstants.FieldLabels].(map[string]any)
	if !ok {
		return
	}

	delete(labels, ConvertedFromNativeLabel)
}
