/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package runai_container_toolkit_reservation

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
)

// RunAIContainerToolkitReservation is retained only for rollback compatibility.
// Deprecated: the RunAI toolkit reservation stack is not reconciled in KAI-based releases.
type RunAIContainerToolkitReservation struct {
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
	// +kubebuilder:validation:Optional
	Image *Image `json:"image,omitempty"`
	// +kubebuilder:validation:Optional
	ReservationPodImage *common.Image `json:"reservation-pod-image,omitempty"`
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`
}

type Image common.Image
type Resources corev1.ResourceRequirements
