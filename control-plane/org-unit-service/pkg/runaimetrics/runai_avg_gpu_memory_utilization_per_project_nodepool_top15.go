/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

type RunaiAvgGpuMemoryUtilizationPerProjectNodePoolTop15Filters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
	Project    []string `mapstructure:"project,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
}

type RunaiAvgGpuMemoryUtilizationPerProjectNodePoolTop15Labels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	NodePool   bool `mapstructure:"nodepool,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAvgGpuMemoryUtilizationPerProjectNodePoolTop15 struct {
	Filters        RunaiAvgGpuMemoryUtilizationPerProjectNodePoolTop15Filters
	ResponseLabels RunaiAvgGpuMemoryUtilizationPerProjectNodePoolTop15Labels
	Quantile       *float32
}

func (m RunaiAvgGpuMemoryUtilizationPerProjectNodePoolTop15) GetName() string {
	return "runai_avg_gpu_memory_utilization_per_project_nodepool_top15"
}

func (m RunaiAvgGpuMemoryUtilizationPerProjectNodePoolTop15) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAvgGpuMemoryUtilizationPerProjectNodePoolTop15) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAvgGpuMemoryUtilizationPerProjectNodePoolTop15) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAvgGpuMemoryUtilizationPerProjectNodePoolTop15) IsAggregation() bool {
	return true
}

func (m RunaiAvgGpuMemoryUtilizationPerProjectNodePoolTop15) IsQuantile() bool {
	return false
}
