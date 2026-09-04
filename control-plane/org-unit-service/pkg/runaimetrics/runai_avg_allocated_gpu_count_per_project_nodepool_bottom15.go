/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

type RunaiAvgAllocatedGpuCountPerProjectNodePoolBottom15Filters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
	Project    []string `mapstructure:"project,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
}

type RunaiAvgAllocatedGpuCountPerProjectNodePoolBottom15Labels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	NodePool   bool `mapstructure:"nodepool,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAvgAllocatedGpuCountPerProjectNodePoolBottom15 struct {
	Filters        RunaiAvgAllocatedGpuCountPerProjectNodePoolBottom15Filters
	ResponseLabels RunaiAvgAllocatedGpuCountPerProjectNodePoolBottom15Labels
	Quantile       *float32
}

func (m RunaiAvgAllocatedGpuCountPerProjectNodePoolBottom15) GetName() string {
	return "runai_avg_allocated_gpu_count_per_project_nodepool_bottom15"
}

func (m RunaiAvgAllocatedGpuCountPerProjectNodePoolBottom15) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAvgAllocatedGpuCountPerProjectNodePoolBottom15) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAvgAllocatedGpuCountPerProjectNodePoolBottom15) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAvgAllocatedGpuCountPerProjectNodePoolBottom15) IsAggregation() bool {
	return true
}

func (m RunaiAvgAllocatedGpuCountPerProjectNodePoolBottom15) IsQuantile() bool {
	return false
}
