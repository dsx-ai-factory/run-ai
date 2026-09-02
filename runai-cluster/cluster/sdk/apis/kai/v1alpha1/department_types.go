/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DepartmentSpec defines the desired state of Department
type DepartmentSpec struct {
	// Department's queues
	Queues []QueueConfig `json:"queues"`
}

// DepartmentStatus defines the observed state of Department
type DepartmentStatus struct {
	// Current conditions of the department
	// +optional
	Conditions []DepartmentCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:storageversion

// Department is the Schema for the departments API
type Department struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DepartmentSpec   `json:"spec,omitempty"`
	Status DepartmentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DepartmentList contains a list of Department
type DepartmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Department `json:"items"`
}

type DepartmentCondition struct {
	Type DepartmentConditionType `json:"type"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// Last time we probed the condition.
	// +optional
	// +nullable
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

type DepartmentConditionType string

// These are built-in conditions of department. An application may use a custom condition not listed here.
const (
	DepartmentDeletionBlocked DepartmentConditionType = "DepartmentDeletionBlocked"
)

func init() {
	SchemeBuilder.Register(&Department{}, &DepartmentList{})
}
