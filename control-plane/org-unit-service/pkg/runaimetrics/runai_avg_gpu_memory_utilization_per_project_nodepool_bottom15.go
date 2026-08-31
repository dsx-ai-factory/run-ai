/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

type RunaiAvgGpuMemoryUtilizationPerProjectNodePoolBottom15Filters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
	Project    []string `mapstructure:"project,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
}

type RunaiAvgGpuMemoryUtilizationPerProjectNodePoolBottom15Labels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	NodePool   bool `mapstructure:"nodepool,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAvgGpuMemoryUtilizationPerProjectNodePoolBottom15 struct {
	Filters        RunaiAvgGpuMemoryUtilizationPerProjectNodePoolBottom15Filters
	ResponseLabels RunaiAvgGpuMemoryUtilizationPerProjectNodePoolBottom15Labels
	Quantile       *float32
}

func (m RunaiAvgGpuMemoryUtilizationPerProjectNodePoolBottom15) GetName() string {
	return "runai_avg_gpu_memory_utilization_per_project_nodepool_bottom15"
}

func (m RunaiAvgGpuMemoryUtilizationPerProjectNodePoolBottom15) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAvgGpuMemoryUtilizationPerProjectNodePoolBottom15) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAvgGpuMemoryUtilizationPerProjectNodePoolBottom15) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAvgGpuMemoryUtilizationPerProjectNodePoolBottom15) IsAggregation() bool {
	return true
}

func (m RunaiAvgGpuMemoryUtilizationPerProjectNodePoolBottom15) IsQuantile() bool {
	return false
}
