/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvl

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

// addResourceClaimTemplateToPod adds the MNNVL resource claim template to the
// pod's resource-claims slice and attaches the claim to the container.
func addResourceClaimTemplateToPod(workloadInitials, workloadName string, resourceClaims *[]corev1.PodResourceClaim, container *corev1.Container) {
	templateName := composeResourceClaimTemplateName(workloadName, workloadInitials)
	claimName := getResourceClaimName()
	addResourceClaimTemplateToPodClaims(resourceClaims, claimName, templateName)

	if container != nil {
		addResourceClaimToContainer(container, claimName)
	}
}

// addResourceClaimToContainer ensures the container has the required DRA
// resource claim reference.
func addResourceClaimToContainer(container *corev1.Container, claimName string) {
	if containerResourceClaimExists(container.Resources.Claims, claimName) {
		return
	}
	container.Resources.Claims = append(
		container.Resources.Claims,
		corev1.ResourceClaim{Name: claimName},
	)
}

// addResourceClaimTemplateToPodClaims adds a resource claim template entry to
// the pod's resource-claims slice (idempotent on claim name).
func addResourceClaimTemplateToPodClaims(claims *[]corev1.PodResourceClaim, claimName, templateName string) {
	if podResourceClaimExists(*claims, claimName) {
		return
	}

	*claims = append(*claims, corev1.PodResourceClaim{
		Name:                      claimName,
		ResourceClaimTemplateName: ptr.To(templateName),
	})
}

func containerResourceClaimExists(claims []corev1.ResourceClaim, claimName string) bool {
	for _, claim := range claims {
		if claim.Name == claimName {
			return true
		}
	}
	return false
}

func podResourceClaimExists(claims []corev1.PodResourceClaim, claimName string) bool {
	for _, claim := range claims {
		if claim.Name == claimName {
			return true
		}
	}
	return false
}
