/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

type RunaiAllocatedMemoryMbPerQueuesFilters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Project    []string `mapstructure:"project,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
}

type RunaiAllocatedMemoryMbPerQueuesResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAllocatedMemoryMbPerQueuesMetric struct {
	Filters        RunaiAllocatedMemoryMbPerQueuesFilters
	ResponseLabels RunaiAllocatedMemoryMbPerQueuesResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedMemoryMbPerQueuesMetric) GetName() string {
	return "runai_allocated_memory_mb_per_queues"
}

func (m RunaiAllocatedMemoryMbPerQueuesMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedMemoryMbPerQueuesMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedMemoryMbPerQueuesMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedMemoryMbPerQueuesMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedMemoryMbPerQueuesMetric) IsQuantile() bool {
	return false
}
