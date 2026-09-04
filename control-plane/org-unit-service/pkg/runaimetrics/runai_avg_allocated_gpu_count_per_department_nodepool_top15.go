/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

type RunaiAvgAllocatedGpuCountPerDepartmentNodePoolTop15Filters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
}

type RunaiAvgAllocatedGpuCountPerDepartmentNodePoolTop15Labels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	NodePool   bool `mapstructure:"nodepool,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAvgAllocatedGpuCountPerDepartmentNodePoolTop15 struct {
	Filters        RunaiAvgAllocatedGpuCountPerDepartmentNodePoolTop15Filters `mapstructure:"filters,omitempty"`
	ResponseLabels RunaiAvgAllocatedGpuCountPerDepartmentNodePoolTop15Labels  `mapstructure:"labels,omitempty"`
	Quantile       *float32
}

func (m RunaiAvgAllocatedGpuCountPerDepartmentNodePoolTop15) GetName() string {
	return "runai_avg_allocated_gpu_count_per_department_nodepool_top15"
}

func (m RunaiAvgAllocatedGpuCountPerDepartmentNodePoolTop15) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAvgAllocatedGpuCountPerDepartmentNodePoolTop15) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAvgAllocatedGpuCountPerDepartmentNodePoolTop15) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAvgAllocatedGpuCountPerDepartmentNodePoolTop15) IsAggregation() bool {
	return true
}

func (m RunaiAvgAllocatedGpuCountPerDepartmentNodePoolTop15) IsQuantile() bool {
	return false
}
