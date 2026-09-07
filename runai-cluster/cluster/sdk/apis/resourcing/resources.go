/*
 * SPDX-FileCopyrightText: Copyright (c) 2022 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package resourcing

// QueueResources defines resource allocations for GPU, CPU, and Memory
type QueueResources struct {
	GPU    SystemResource `json:"gpu,omitempty"`
	CPU    SystemResource `json:"cpu,omitempty"`
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
