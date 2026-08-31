/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package limit_range

import (
	"k8s.io/utils/ptr"
)

// LimitRange TODO
type LimitRange struct {
	// CPUDefaultRequestGPUFactor TODO
	// +kubebuilder:validation:Optional
	CPUDefaultRequestGPUFactor string `json:"cpuDefaultRequestGpuFactor,omitempty"`

	// MemoryDefaultRequestGPUFactor TODO
	// +kubebuilder:validation:Optional
	MemoryDefaultRequestGPUFactor string `json:"memoryDefaultRequestGpuFactor,omitempty"`

	// CPUDefaultLimitGPUFactor TODO
	// +kubebuilder:validation:Optional
	CPUDefaultLimitGPUFactor string `json:"cpuDefaultLimitGpuFactor,omitempty"`

	// MemoryDefaultLimitGPUFactor TODO
	// +kubebuilder:validation:Optional
	MemoryDefaultLimitGPUFactor string `json:"memoryDefaultLimitGpuFactor,omitempty"`

	// CPUDefaultRequestCPULimitFactorNoGPU TODO
	// +kubebuilder:validation:Optional
	CPUDefaultRequestCPULimitFactorNoGPU *float32 `json:"cpuDefaultRequestCpuLimitFactorNoGpu,omitempty"`

	// MemoryDefaultRequestMemoryLimitFactorNoGPU TODO
	// +kubebuilder:validation:Optional
	MemoryDefaultRequestMemoryLimitFactorNoGPU *float32 `json:"memoryDefaultRequestMemoryLimitFactorNoGpu,omitempty"`

	// CPUDefaultRequest TODO
	// +kubebuilder:validation:Optional
	CPUDefaultRequest string `json:"cpuDefaultRequest,omitempty"`

	// MemoryDefaultRequest TODO
	// +kubebuilder:validation:Optional
	MemoryDefaultRequest string `json:"memoryDefaultRequest,omitempty"`

	// CPUDefaultLimit TODO
	// +kubebuilder:validation:Optional
	CPUDefaultLimit string `json:"cpuDefaultLimit,omitempty"`

	// MemoryDefaultLimit TODO
	// +kubebuilder:validation:Optional
	MemoryDefaultLimit string `json:"memoryDefaultLimit,omitempty"`

	// CPUMaxLimit TODO
	// +kubebuilder:validation:Optional
	CPUMaxLimit string `json:"cpuMaxLimit,omitempty"`

	// MemoryMaxLimit TODO
	// +kubebuilder:validation:Optional
	MemoryMaxLimit string `json:"memoryMaxLimit,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (l *LimitRange) SetDefaultsWhereNeeded() {
	if l.CPUDefaultRequestGPUFactor == "" {
		l.CPUDefaultRequestGPUFactor = "100m"
	}

	if l.MemoryDefaultRequestGPUFactor == "" {
		l.MemoryDefaultRequestGPUFactor = "100Mi"
	}

	if l.CPUDefaultRequestCPULimitFactorNoGPU == nil {
		l.CPUDefaultRequestCPULimitFactorNoGPU = ptr.To[float32](0.1)
	}

	if l.MemoryDefaultRequestMemoryLimitFactorNoGPU == nil {
		l.MemoryDefaultRequestMemoryLimitFactorNoGPU = ptr.To[float32](0.1)
	}
}
