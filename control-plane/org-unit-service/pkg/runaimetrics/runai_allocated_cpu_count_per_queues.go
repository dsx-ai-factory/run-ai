/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

type RunaiAllocatedCpuCountPerQueuesFilters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Project    []string `mapstructure:"project,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
}

type RunaiAllocatedCpuCountPerQueuesResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAllocatedCpuCountPerQueuesMetric struct {
	Filters        RunaiAllocatedCpuCountPerQueuesFilters
	ResponseLabels RunaiAllocatedCpuCountPerQueuesResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedCpuCountPerQueuesMetric) GetName() string {
	return "runai_allocated_cpu_per_queues"
}

func (m RunaiAllocatedCpuCountPerQueuesMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedCpuCountPerQueuesMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedCpuCountPerQueuesMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedCpuCountPerQueuesMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedCpuCountPerQueuesMetric) IsQuantile() bool {
	return false
}
