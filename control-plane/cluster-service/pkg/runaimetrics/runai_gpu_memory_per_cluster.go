/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuMemoryBytesPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiGpuMemoryBytesPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiGpuMemoryBytesPerClusterMetric struct {
	Filters        RunaiGpuMemoryBytesPerClusterFilters
	ResponseLabels RunaiGpuMemoryBytesPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiGpuMemoryBytesPerClusterMetric) GetName() string {
	return "runai_total_gpu_memory_byte_per_cluster"
}

func (m RunaiGpuMemoryBytesPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuMemoryBytesPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuMemoryBytesPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuMemoryBytesPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuMemoryBytesPerClusterMetric) IsQuantile() bool {
	return false
}
