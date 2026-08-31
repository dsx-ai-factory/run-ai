/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

type RunaiAvgGpuUtilizationPerProjectNodePoolBottom15Filters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
	Project    []string `mapstructure:"project,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
}

type RunaiAvgGpuUtilizationPerProjectNodePoolBottom15Labels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	NodePool   bool `mapstructure:"nodepool,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAvgGpuUtilizationPerProjectNodePoolBottom15 struct {
	Filters        RunaiAvgGpuUtilizationPerProjectNodePoolBottom15Filters
	ResponseLabels RunaiAvgGpuUtilizationPerProjectNodePoolBottom15Labels
	Quantile       *float32
}

func (m RunaiAvgGpuUtilizationPerProjectNodePoolBottom15) GetName() string {
	return "runai_avg_gpu_utilization_per_project_nodepool_bottom15"
}

func (m RunaiAvgGpuUtilizationPerProjectNodePoolBottom15) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAvgGpuUtilizationPerProjectNodePoolBottom15) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAvgGpuUtilizationPerProjectNodePoolBottom15) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAvgGpuUtilizationPerProjectNodePoolBottom15) IsAggregation() bool {
	return true
}

func (m RunaiAvgGpuUtilizationPerProjectNodePoolBottom15) IsQuantile() bool {
	return false
}
