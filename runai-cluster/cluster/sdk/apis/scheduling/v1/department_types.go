/*
 * SPDX-FileCopyrightText: Copyright (c) 2022 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// Department is a department of Queue.
type Department struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// +optional
	Spec DepartmentSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

//+kubebuilder:object:root=true

// DepartmentList is a collection of deperatments.
type DepartmentList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []Department `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// DepartmentSpec represents the template of Department.
type DepartmentSpec struct {
	DeservedGPUs   float64  `json:"deservedGpus,omitempty" protobuf:"bytes,1,opt,name=deservedGpus"`
	MaxAllowedGPUs *float64 `json:"maxAllowedGpus,omitempty" protobuf:"bytes,1,opt,name=maxAllowedGpus"`
}

func init() {
	SchemeBuilder.Register(&Department{}, &DepartmentList{})
}
