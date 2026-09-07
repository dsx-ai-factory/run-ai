/*
 * SPDX-FileCopyrightText: Copyright (c) 2022 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package v1alpha1

// QueueResourcesConfig defines resource allocations for GPU, CPU, and Memory
type QueueResourcesConfig struct {
	// GPU resources in fractions (e.g. 0.7 = 70% of a GPU)
	GPU SystemResource `json:"gpu,omitempty"`

	// CPU resources in millicpus (e.g. 1000 = 1 CPU)
	CPU SystemResource `json:"cpu,omitempty"`

	// Memory resources in megabytes (1 MB = 10^6 bytes)
	Memory SystemResource `json:"memory,omitempty"`
}

// SystemResource defines resource limits and quotas
type SystemResource struct {
	// +optional
	Deserved float64 `json:"deserved"`
	// +optional
	OverQuotaWeight float64 `json:"overQuotaWeight"`
	// +optional
	Limit float64 `json:"limit"`
}

type QueueConfig struct {
	// Queue name
	Name string `json:"name"`

	// Nodepool associated with the queue
	Nodepool string `json:"nodepool"`

	// Queue resources definition
	Resources *QueueResourcesConfig `json:"resources,omitempty"`

	// Queue priority
	Priority *int32 `json:"priority,omitempty"`
}
