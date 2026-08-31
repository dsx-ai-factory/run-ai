/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuMemoryUtilizationAvgPerDepartmentFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
}

type RunaiGpuMemoryUtilizationAvgPerDepartmentResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiGpuMemoryUtilizationAvgPerDepartmentMetric struct {
	Filters        RunaiGpuMemoryUtilizationAvgPerDepartmentFilters
	ResponseLabels RunaiGpuMemoryUtilizationAvgPerDepartmentResponseLabels
	Quantile       *float32
}

func (m RunaiGpuMemoryUtilizationAvgPerDepartmentMetric) GetName() string {
	return "runai_gpu_memory_utilization_avg_per_department"
}

func (m RunaiGpuMemoryUtilizationAvgPerDepartmentMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuMemoryUtilizationAvgPerDepartmentMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuMemoryUtilizationAvgPerDepartmentMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuMemoryUtilizationAvgPerDepartmentMetric) IsAggregation() bool {
	return true
}

func (m RunaiGpuMemoryUtilizationAvgPerDepartmentMetric) IsQuantile() bool {
	return false
}
