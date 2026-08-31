/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1

import (
	rayv1alpha1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RayCluster struct {
	// Standard object metadata.
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the RayCluster.
	Spec   rayv1alpha1.RayClusterSpec   `json:"spec,omitempty"`
	Status rayv1alpha1.RayClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RayClusterList contains a list of RayCluster
type RayClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RayCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RayCluster{}, &RayClusterList{})
}
