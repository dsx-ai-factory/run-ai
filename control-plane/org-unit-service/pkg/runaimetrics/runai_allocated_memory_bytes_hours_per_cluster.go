/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedMemoryBytesHoursPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiAllocatedMemoryBytesHoursPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiAllocatedMemoryBytesHoursPerClusterMetric struct {
	Filters        RunaiAllocatedMemoryBytesHoursPerClusterFilters
	ResponseLabels RunaiAllocatedMemoryBytesHoursPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedMemoryBytesHoursPerClusterMetric) GetName() string {
	return "runai_allocated_memory_bytes_hours_per_cluster"
}

func (m RunaiAllocatedMemoryBytesHoursPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedMemoryBytesHoursPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedMemoryBytesHoursPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedMemoryBytesHoursPerClusterMetric) IsAggregation() bool {
	return true
}

func (m RunaiAllocatedMemoryBytesHoursPerClusterMetric) IsQuantile() bool {
	return false
}
