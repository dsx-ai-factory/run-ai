/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

// InferenceStatus represents the status of an inference workload
// +k8s:deepcopy-gen=true
type InferenceStatus struct {
	// LatestCreatedRevisionName represents the name of the latest created revision name
	LatestCreatedRevisionName *string `json:"latestCreatedRevisionName,omitempty"`

	// LatestReadyRevisionName represents the name of the latest ready revision name
	LatestReadyRevisionName *string `json:"latestReadyRevisionName,omitempty"`

	// ExternalUrls represents the external urls of the workload
	ExternalUrls []string `json:"externalUrls,omitempty"`

	// +kubebuilder:validation:Enum:=Creating;Initializing;Pending;Failed;Ready;Progressing;Deleting;Degraded
	Phase        string `json:"phase,omitempty"`
	PhaseMessage string `json:"phaseMessage,omitempty"`
}

// DistributedInferenceStatus represents the status of a distributed inference workload
// +k8s:deepcopy-gen=true
type DistributedInferenceStatus struct {
	// Replicas is the number of replicas.
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// ReadyReplicas is the number of ready replicas.
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Phase represents the current phase of the workload.
	// +kubebuilder:validation:Enum:=Creating;Pending;Initializing;Ready;Deleting;Degraded;Failed
	// +optional
	Phase string `json:"phase,omitempty"`

	// PhaseMessage provides additional information about the current phase.
	// +optional
	PhaseMessage string `json:"phaseMessage,omitempty"`

	// ServingUrls represents the serving endpoints of the workload, either internal or external.
	// +listType=atomic
	// +optional
	ServingUrls []string `json:"servingUrls,omitempty"`

	// ExternalUrls represents the external urls of the workload.
	// +listType=atomic
	// +optional
	ExternalUrls []string `json:"externalUrls,omitempty"`
}
