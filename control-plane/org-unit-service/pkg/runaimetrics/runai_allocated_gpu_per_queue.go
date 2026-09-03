/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

type RunaiAllocatedGpuPerQueueFilters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Project    []string `mapstructure:"project,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
}

type RunaiAllocatedGpuPerQueueResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAllocatedGpuPerQueueMetric struct {
	Filters        RunaiAllocatedGpuPerQueueFilters
	ResponseLabels RunaiAllocatedGpuPerQueueResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedGpuPerQueueMetric) GetName() string {
	return "runai_allocated_gpu_per_queue"
}

func (m RunaiAllocatedGpuPerQueueMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedGpuPerQueueMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedGpuPerQueueMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedGpuPerQueueMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedGpuPerQueueMetric) IsQuantile() bool {
	return false
}
