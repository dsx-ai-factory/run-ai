/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuUtilizationAvgPerDepartmentFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
}

type RunaiGpuUtilizationAvgPerDepartmentResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiGpuUtilizationAvgPerDepartmentMetric struct {
	Filters        RunaiGpuUtilizationAvgPerDepartmentFilters
	ResponseLabels RunaiGpuUtilizationAvgPerDepartmentResponseLabels
	Quantile       *float32
}

func (m RunaiGpuUtilizationAvgPerDepartmentMetric) GetName() string {
	return "runai_gpu_utilization_avg_per_department"
}

func (m RunaiGpuUtilizationAvgPerDepartmentMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuUtilizationAvgPerDepartmentMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuUtilizationAvgPerDepartmentMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuUtilizationAvgPerDepartmentMetric) IsAggregation() bool {
	return true
}

func (m RunaiGpuUtilizationAvgPerDepartmentMetric) IsQuantile() bool {
	return false
}
