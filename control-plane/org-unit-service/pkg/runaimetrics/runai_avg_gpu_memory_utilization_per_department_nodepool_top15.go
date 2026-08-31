/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

type RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolTop15Filters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
}

type RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolTop15Labels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	NodePool   bool `mapstructure:"nodepool,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolTop15 struct {
	Filters        RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolTop15Filters `mapstructure:"filters,omitempty"`
	ResponseLabels RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolTop15Labels  `mapstructure:"labels,omitempty"`
	Quantile       *float32
}

func (m RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolTop15) GetName() string {
	return "runai_avg_gpu_memory_utilization_per_department_nodepool_top15"
}

func (m RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolTop15) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolTop15) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolTop15) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolTop15) IsAggregation() bool {
	return true
}

func (m RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolTop15) IsQuantile() bool {
	return false
}
