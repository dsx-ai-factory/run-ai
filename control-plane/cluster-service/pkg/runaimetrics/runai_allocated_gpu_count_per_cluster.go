/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedGpuCountPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiAllocatedGpuCountPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiAllocatedGpuCountPerClusterMetric struct {
	Filters        RunaiAllocatedGpuCountPerClusterFilters
	ResponseLabels RunaiAllocatedGpuCountPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedGpuCountPerClusterMetric) GetName() string {
	return "runai_allocated_gpu_count_per_cluster"
}

func (m RunaiAllocatedGpuCountPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedGpuCountPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedGpuCountPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedGpuCountPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedGpuCountPerClusterMetric) IsQuantile() bool {
	return false
}
