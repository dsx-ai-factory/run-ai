/*
 * SPDX-FileCopyrightText: Copyright (c) 2022 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:deprecatedversion

// Project is the Schema for the projects API
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ProjectSpec   `json:"spec,omitempty"`
	Status            ProjectStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProjectList contains a list of Project
type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Project `json:"items"`
}

type ProjectSpec struct {
	Department                  string             `json:"department,omitempty"`
	DeservedGpus                float64            `json:"deservedGpus,omitempty"`
	MaxAllowedGpus              float64            `json:"maxAllowedGpus,omitempty"`
	GpuOverQuotaWeight          float64            `json:"gpuOverQuotaWeight,omitempty"`
	InteractiveJobTimeLimitSecs int64              `json:"interactiveJobTimeLimitSecs,omitempty"`
	NodeAffinityInteractive     []string           `json:"nodeAffinityInteractive,omitempty"`
	NodeAffinityTrain           []string           `json:"nodeAffinityTrain,omitempty"`
	Permissions                 ProjectPermissions `json:"permissions,omitempty"`
	LimitsRef                   int64              `json:"limitsRef,omitempty"`
	SecretsRef                  map[string]string  `json:"secretsRef,omitempty"`
	CpuDefaultRequest           string             `json:"cpuDefaultRequest,omitempty"`
	MemoryDefaultRequest        string             `json:"memoryDefaultRequest,omitempty"`
	CpuDefaultLimit             string             `json:"cpuDefaultLimit,omitempty"`
	MemoryDefaultLimit          string             `json:"memoryDefaultLimit,omitempty"`
	CpuMaxLimit                 string             `json:"cpuMaxLimit,omitempty"`
	MemoryMaxLimit              string             `json:"memoryMaxLimit,omitempty"`
	SwapEnabled                 bool               `json:"swapEnabled,omitempty"`
}

type ProjectStatus struct {
	//	INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
}

type ProjectPermissions struct {
	Users  []string `json:"users,omitempty"`
	Groups []string `json:"groups,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Project{}, &ProjectList{})
}

// IsOwner returns the index of this Project in the given object's OwnerReferences array if it exists.
func (project Project) IsOwner(obj client.Object) (indexOfOwnerInRefArray int) {
	indexOfOwnerInRefArray = -1
	for index, ownerRef := range obj.GetOwnerReferences() {
		if ownerRef.Kind == "Project" &&
			ownerRef.UID == project.UID &&
			ownerRef.Name == project.Name {
			indexOfOwnerInRefArray = index
		}
	}
	return indexOfOwnerInRefArray
}
