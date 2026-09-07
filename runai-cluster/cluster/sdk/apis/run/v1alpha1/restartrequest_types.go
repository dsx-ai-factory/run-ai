/*
 * SPDX-FileCopyrightText: Copyright (c) 2022 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// WorkloadReference identifies a workload (e.g. WorkloadRunner or DistributedWorkload) by GVK, name, and generation.
// Used by RestartRequest to target a specific workload instance.
type WorkloadReference struct {
	// GVK is the GroupVersionKind of the workload resource.
	GVK metav1.GroupVersionKind `json:"gvk,omitempty"`

	// Name is the name of the workload resource.
	Name string `json:"name"`

	// Generation is the observed generation of the workload at the time the restart was requested.
	// Used to detect concurrent spec changes and avoid applying stale overrides.
	// +optional
	Generation int64 `json:"generation,omitempty"`
}

// JSONPatchOperationType is the type of JSON Patch operation (RFC 6902).
// +kubebuilder:validation:Enum=add;remove;replace;move;copy;test
type JSONPatchOperationType string

const (
	JSONPatchOperationAdd     JSONPatchOperationType = "add"
	JSONPatchOperationRemove  JSONPatchOperationType = "remove"
	JSONPatchOperationReplace JSONPatchOperationType = "replace"
	JSONPatchOperationMove    JSONPatchOperationType = "move"
	JSONPatchOperationCopy    JSONPatchOperationType = "copy"
	JSONPatchOperationTest    JSONPatchOperationType = "test"
)

// JSONPatchOperation represents a single JSON Patch operation (RFC 6902).
type JSONPatchOperation struct {
	// Op is the operation to perform (add, remove, replace, move, copy, test).
	Op JSONPatchOperationType `json:"op"`

	// Path is the JSON pointer path to the target location.
	Path string `json:"path"`

	// Value is the value to set (required for add, replace, test operations).
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Value *runtime.RawExtension `json:"value,omitempty"`

	// From is the source path (required for move and copy operations).
	// +optional
	From string `json:"from,omitempty"`
}

// SpecOverrides is a JSON Patch array (RFC 6902) applied to the workload spec during restart.
// Each element is a single patch operation.
type SpecOverrides []JSONPatchOperation

// RestartRequestSpec defines the desired state of RestartRequest.
type RestartRequestSpec struct {
	// WorkloadRef identifies the workload to restart (WorkloadRunner, DistributedWorkload, etc.).
	WorkloadRef WorkloadReference `json:"workloadRef"`

	// SpecOverrides is an optional JSON Patch array applied to the workload spec when performing the restart.
	// Enables temporary overrides (e.g. replica count, image) without mutating the original workload spec.
	// +optional
	SpecOverrides SpecOverrides `json:"specOverrides,omitempty"`
}

// RestartRequestPhase is the phase of a RestartRequest.
// +kubebuilder:validation:Enum=Pending;InProgress;Succeeded;Failed
type RestartRequestPhase string

const (
	RestartRequestPhasePending    RestartRequestPhase = "Pending"
	RestartRequestPhaseInProgress RestartRequestPhase = "InProgress"
	RestartRequestPhaseSucceeded  RestartRequestPhase = "Succeeded"
	RestartRequestPhaseFailed     RestartRequestPhase = "Failed"
)

// RestartRequestStatus defines the observed state of RestartRequest.
type RestartRequestStatus struct {
	// Phase is the current phase of the restart (Pending, InProgress, Succeeded, Failed).
	// +optional
	Phase RestartRequestPhase `json:"phase,omitempty"`

	// Conditions represent the latest available observations of the RestartRequest's state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=rr
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// RestartRequest is the Schema for the restartrequests API.
// It requests a coordinated restart of a workload (WorkloadRunner or DistributedWorkload) with optional spec overrides.
type RestartRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RestartRequestSpec   `json:"spec,omitempty"`
	Status RestartRequestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RestartRequestList contains a list of RestartRequest.
type RestartRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RestartRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RestartRequest{}, &RestartRequestList{})
}
