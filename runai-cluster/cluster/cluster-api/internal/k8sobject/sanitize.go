package k8sobject

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

const lastAppliedAnnotation = "kubectl.kubernetes.io/last-applied-configuration"

// Strips server-side-apply bookkeeping (managedFields) and the kubectl
// client-side-apply audit trail — both bloat the payload without being
// actionable for the caller.
func sanitize(u *unstructured.Unstructured) {
	if u == nil {
		return
	}
	u.SetManagedFields(nil)
	annotations := u.GetAnnotations()
	if _, ok := annotations[lastAppliedAnnotation]; ok {
		delete(annotations, lastAppliedAnnotation)
		u.SetAnnotations(annotations)
	}
}
