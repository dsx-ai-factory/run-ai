package constants

const (
	NativeInferenceWorkloadKind = "InferenceWorkload"
	V2InferenceWorkloadKind     = "Inference"
)

// IsManagedInferenceRevision reports whether labels identify a Revision whose
// phase is managed by Run:ai. V2 Knative Revisions inherit the workload name
// from the Service template, while service-only management labels do not
// propagate to Revisions.
func IsManagedInferenceRevision(labels map[string]string) bool {
	switch labels[ControlPlaneWorkloadKindLabelForWorkload] {
	case NativeInferenceWorkloadKind:
		return true
	case V2InferenceWorkloadKind:
		return labels[ControlPlaneWorkloadNameLabelForWorkload] != ""
	default:
		return false
	}
}
