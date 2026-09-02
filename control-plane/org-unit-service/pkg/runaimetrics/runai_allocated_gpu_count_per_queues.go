/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

type RunaiAllocatedGpuCountPerQueuesFilters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Project    []string `mapstructure:"project,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
}

type RunaiAllocatedGpuCountPerQueuesResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAllocatedGpuCountPerQueuesMetric struct {
	Filters        RunaiAllocatedGpuCountPerQueuesFilters
	ResponseLabels RunaiAllocatedGpuCountPerQueuesResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedGpuCountPerQueuesMetric) GetName() string {
	return "runai_allocated_gpu_count_per_queues"
}

func (m RunaiAllocatedGpuCountPerQueuesMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedGpuCountPerQueuesMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedGpuCountPerQueuesMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedGpuCountPerQueuesMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedGpuCountPerQueuesMetric) IsQuantile() bool {
	return false
}
