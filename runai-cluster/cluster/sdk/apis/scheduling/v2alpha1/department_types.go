/*
 * SPDX-FileCopyrightText: Copyright (c) 2022 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v2alpha1

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

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

// DepartmentSpec represents the template of Department.
type DepartmentSpec struct {
	DeservedGPUs   float64                    `json:"deservedGpus,omitempty" protobuf:"bytes,1,opt,name=deservedGpus"`
	MaxAllowedGPUs *float64                   `json:"maxAllowedGpus,omitempty" protobuf:"bytes,1,opt,name=maxAllowedGpus"`
	Resources      *v2alpha1.ProjectResources `json:"resources,omitempty" protobuf:"bytes,1,opt,name=resources"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DepartmentList is a collection of deperatments.
type DepartmentList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []Department `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&Department{}, &DepartmentList{})
}
