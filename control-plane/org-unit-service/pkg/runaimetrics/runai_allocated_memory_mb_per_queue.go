/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

type RunaiAllocatedMemoryMBPerQueueFilters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Project    []string `mapstructure:"project,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
}

type RunaiAllocatedMemoryMBPerQueueResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAllocatedMemoryMBPerQueueMetric struct {
	Filters        RunaiAllocatedMemoryMBPerQueueFilters
	ResponseLabels RunaiAllocatedMemoryMBPerQueueResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedMemoryMBPerQueueMetric) GetName() string {
	return "runai_allocated_memory_mb_per_queue"
}

func (m RunaiAllocatedMemoryMBPerQueueMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedMemoryMBPerQueueMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedMemoryMBPerQueueMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedMemoryMBPerQueueMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedMemoryMBPerQueueMetric) IsQuantile() bool {
	return false
}
