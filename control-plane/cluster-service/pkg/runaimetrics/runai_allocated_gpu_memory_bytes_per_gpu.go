/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedGpuMemoryBytesPerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiAllocatedGpuMemoryBytesPerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiAllocatedGpuMemoryBytesPerGpuMetric struct {
	Filters        RunaiAllocatedGpuMemoryBytesPerGpuFilters
	ResponseLabels RunaiAllocatedGpuMemoryBytesPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedGpuMemoryBytesPerGpuMetric) GetName() string {
	return "runai_allocated_gpu_memory_bytes_per_gpu"
}

func (m RunaiAllocatedGpuMemoryBytesPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedGpuMemoryBytesPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedGpuMemoryBytesPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedGpuMemoryBytesPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedGpuMemoryBytesPerGpuMetric) IsQuantile() bool {
	return false
}
