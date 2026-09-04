/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedMemoryBytesPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiAllocatedMemoryBytesPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiAllocatedMemoryBytesPerClusterMetric struct {
	Filters        RunaiAllocatedMemoryBytesPerClusterFilters
	ResponseLabels RunaiAllocatedMemoryBytesPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedMemoryBytesPerClusterMetric) GetName() string {
	return "runai_allocated_memory_bytes_per_cluster"
}

func (m RunaiAllocatedMemoryBytesPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedMemoryBytesPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedMemoryBytesPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedMemoryBytesPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedMemoryBytesPerClusterMetric) IsQuantile() bool {
	return false
}
