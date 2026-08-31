/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiIdleAllocatedGpuCountPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiIdleAllocatedGpuCountPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiIdleAllocatedGpuCountPerClusterMetric struct {
	Filters        RunaiIdleAllocatedGpuCountPerClusterFilters
	ResponseLabels RunaiIdleAllocatedGpuCountPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiIdleAllocatedGpuCountPerClusterMetric) GetName() string {
	return "runai_idle_allocated_gpu_count_per_cluster"
}

func (m RunaiIdleAllocatedGpuCountPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiIdleAllocatedGpuCountPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiIdleAllocatedGpuCountPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiIdleAllocatedGpuCountPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiIdleAllocatedGpuCountPerClusterMetric) IsQuantile() bool {
	return false
}
