/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedGpuCountAvgPerDepartmentFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
}

type RunaiAllocatedGpuCountAvgPerDepartmentResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAllocatedGpuCountAvgPerDepartmentMetric struct {
	Filters        RunaiAllocatedGpuCountAvgPerDepartmentFilters
	ResponseLabels RunaiAllocatedGpuCountAvgPerDepartmentResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedGpuCountAvgPerDepartmentMetric) GetName() string {
	return "runai_allocated_gpu_count_avg_per_department"
}

func (m RunaiAllocatedGpuCountAvgPerDepartmentMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedGpuCountAvgPerDepartmentMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedGpuCountAvgPerDepartmentMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedGpuCountAvgPerDepartmentMetric) IsAggregation() bool {
	return true
}

func (m RunaiAllocatedGpuCountAvgPerDepartmentMetric) IsQuantile() bool {
	return false
}
