package drautils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("Resource Utils", func() {
	Describe("AddResourceClaimToContainer", func() {
		It("should add a resource claim to a container with no existing claims", func() {
			// Arrange
			container := &corev1.Container{
				Name: "test-container",
			}
			resourceClaimName := "test-claim"

			// Act
			AddResourceClaimToContainer(container, resourceClaimName)

			// Assert
			Expect(container.Resources.Claims).To(HaveLen(1))
			Expect(container.Resources.Claims[0].Name).To(Equal(resourceClaimName))
		})

		It("should not add duplicate resource claim to a container", func() {
			// Arrange
			resourceClaimName := "test-claim"
			container := &corev1.Container{
				Name: "test-container",
				Resources: corev1.ResourceRequirements{
					Claims: []corev1.ResourceClaim{
						{Name: resourceClaimName},
					},
				},
			}

			// Act
			AddResourceClaimToContainer(container, resourceClaimName)

			// Assert
			Expect(container.Resources.Claims).To(HaveLen(1))
			Expect(container.Resources.Claims[0].Name).To(Equal(resourceClaimName))
		})

		It("should add a new resource claim to a container with existing claims", func() {
			// Arrange
			existingClaimName := "existing-claim"
			newClaimName := "new-claim"
			container := &corev1.Container{
				Name: "test-container",
				Resources: corev1.ResourceRequirements{
					Claims: []corev1.ResourceClaim{
						{Name: existingClaimName},
					},
				},
			}

			// Act
			AddResourceClaimToContainer(container, newClaimName)

			// Assert
			Expect(container.Resources.Claims).To(HaveLen(2))
			Expect(container.Resources.Claims[0].Name).To(Equal(existingClaimName))
			Expect(container.Resources.Claims[1].Name).To(Equal(newClaimName))
		})
	})

	Describe("AddResourceClaimTemplateToPod", func() {
		It("should add a resource claim template to a pod with no existing claims", func() {
			// Arrange
			podSpec := &corev1.PodSpec{}
			claimName := "test-claim"
			templateName := "test-template"

			// Act
			AddResourceClaimTemplateToPod(podSpec, claimName, templateName)

			// Assert
			Expect(podSpec.ResourceClaims).To(HaveLen(1))
			Expect(podSpec.ResourceClaims[0].Name).To(Equal(claimName))
			Expect(*podSpec.ResourceClaims[0].ResourceClaimTemplateName).To(Equal(templateName))
		})

		It("should not add duplicate resource claim template to a pod", func() {
			// Arrange
			claimName := "test-claim"
			templateName := "test-template"
			podSpec := &corev1.PodSpec{
				ResourceClaims: []corev1.PodResourceClaim{
					{
						Name:                      claimName,
						ResourceClaimTemplateName: ptr.To(templateName),
					},
				},
			}

			// Act
			AddResourceClaimTemplateToPod(podSpec, claimName, templateName)

			// Assert
			Expect(podSpec.ResourceClaims).To(HaveLen(1))
			Expect(podSpec.ResourceClaims[0].Name).To(Equal(claimName))
			Expect(*podSpec.ResourceClaims[0].ResourceClaimTemplateName).To(Equal(templateName))
		})

		It("should add a new resource claim template to a pod with existing claims", func() {
			// Arrange
			existingClaimName := "existing-claim"
			existingTemplateName := "existing-template"
			newClaimName := "new-claim"
			newTemplateName := "new-template"

			podSpec := &corev1.PodSpec{
				ResourceClaims: []corev1.PodResourceClaim{
					{
						Name:                      existingClaimName,
						ResourceClaimTemplateName: ptr.To(existingTemplateName),
					},
				},
			}

			// Act
			AddResourceClaimTemplateToPod(podSpec, newClaimName, newTemplateName)

			// Assert
			Expect(podSpec.ResourceClaims).To(HaveLen(2))
			Expect(podSpec.ResourceClaims[0].Name).To(Equal(existingClaimName))
			Expect(*podSpec.ResourceClaims[0].ResourceClaimTemplateName).To(Equal(existingTemplateName))
			Expect(podSpec.ResourceClaims[1].Name).To(Equal(newClaimName))
			Expect(*podSpec.ResourceClaims[1].ResourceClaimTemplateName).To(Equal(newTemplateName))
		})
	})

	Describe("resourceClaimExists", func() {
		It("should return true when resource claim exists", func() {
			// Arrange
			claimName := "test-claim"
			claims := []corev1.ResourceClaim{
				{Name: "other-claim"},
				{Name: claimName},
			}

			// Act
			result := resourceClaimExists(claims, claimName)

			// Assert
			Expect(result).To(BeTrue())
		})

		It("should return false when resource claim does not exist", func() {
			// Arrange
			claims := []corev1.ResourceClaim{
				{Name: "other-claim"},
			}

			// Act
			result := resourceClaimExists(claims, "non-existent-claim")

			// Assert
			Expect(result).To(BeFalse())
		})

		It("should return false when claims list is empty", func() {
			// Arrange
			var claims []corev1.ResourceClaim

			// Act
			result := resourceClaimExists(claims, "any-claim")

			// Assert
			Expect(result).To(BeFalse())
		})
	})

	Describe("podResourceClaimExists", func() {
		It("should return true when pod resource claim exists", func() {
			// Arrange
			claimName := "test-claim"
			claims := []corev1.PodResourceClaim{
				{Name: "other-claim"},
				{Name: claimName},
			}

			// Act
			result := podResourceClaimExists(claims, claimName)

			// Assert
			Expect(result).To(BeTrue())
		})

		It("should return false when pod resource claim does not exist", func() {
			// Arrange
			claims := []corev1.PodResourceClaim{
				{Name: "other-claim"},
			}

			// Act
			result := podResourceClaimExists(claims, "non-existent-claim")

			// Assert
			Expect(result).To(BeFalse())
		})

		It("should return false when pod claims list is empty", func() {
			// Arrange
			var claims []corev1.PodResourceClaim

			// Act
			result := podResourceClaimExists(claims, "any-claim")

			// Assert
			Expect(result).To(BeFalse())
		})
	})
})
