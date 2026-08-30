package drautils

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

// AddResourceClaimToContainer ensures the container has the required resource claim
// and returns the name of the resource claim template.
func AddResourceClaimToContainer(container *corev1.Container, resourceClaimName string) {
	if !resourceClaimExists(container.Resources.Claims, resourceClaimName) {
		container.Resources.Claims = append(
			container.Resources.Claims,
			corev1.ResourceClaim{Name: resourceClaimName},
		)
	}
}

// AddResourceClaimTemplateToPod ensures the Pod’s resource claim template is created if needed.
func AddResourceClaimTemplateToPod(podSpec *corev1.PodSpec, claimName, resourceClaimTemplateName string) {
	AddResourceClaimTemplateToPodClaims(&podSpec.ResourceClaims, claimName, resourceClaimTemplateName)
}

// AddResourceClaimTemplateToPodClaims adds a resource claim template to the pod claims slice
func AddResourceClaimTemplateToPodClaims(claims *[]corev1.PodResourceClaim, claimName, templateName string) {
	if podResourceClaimExists(*claims, claimName) {
		return
	}

	*claims = append(*claims, corev1.PodResourceClaim{
		Name:                      claimName,
		ResourceClaimTemplateName: ptr.To(templateName),
	})
}

// resourceClaimExists checks if the given container resource claim exists.
func resourceClaimExists(claims []corev1.ResourceClaim, claimName string) bool {
	for _, claim := range claims {
		if claim.Name == claimName {
			return true
		}
	}
	return false
}

// podResourceClaimExists checks if the given pod resource claim exists.
func podResourceClaimExists(claims []corev1.PodResourceClaim, claimName string) bool {
	for _, claim := range claims {
		if claim.Name == claimName {
			return true
		}
	}
	return false
}
