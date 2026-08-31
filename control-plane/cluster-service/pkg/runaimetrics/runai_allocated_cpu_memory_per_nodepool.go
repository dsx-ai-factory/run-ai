/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedMemoryBytesPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiAllocatedMemoryBytesPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiAllocatedMemoryBytesPerNodepoolMetric struct {
	Filters        RunaiAllocatedMemoryBytesPerNodepoolFilters
	ResponseLabels RunaiAllocatedMemoryBytesPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedMemoryBytesPerNodepoolMetric) GetName() string {
	return "runai_allocated_memory_bytes_per_nodepool"
}

func (m RunaiAllocatedMemoryBytesPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedMemoryBytesPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedMemoryBytesPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedMemoryBytesPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedMemoryBytesPerNodepoolMetric) IsQuantile() bool {
	return false
}
