/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUsedGpuMemoryBytesPerNodeFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiUsedGpuMemoryBytesPerNodeResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
}

type RunaiUsedGpuMemoryBytesPerNodeMetric struct {
	Filters        RunaiUsedGpuMemoryBytesPerNodeFilters
	ResponseLabels RunaiUsedGpuMemoryBytesPerNodeResponseLabels
	Quantile       *float32
}

func (m RunaiUsedGpuMemoryBytesPerNodeMetric) GetName() string {
	return "runai_used_gpu_memory_bytes_per_node"
}

func (m RunaiUsedGpuMemoryBytesPerNodeMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUsedGpuMemoryBytesPerNodeMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUsedGpuMemoryBytesPerNodeMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUsedGpuMemoryBytesPerNodeMetric) IsAggregation() bool {
	return false
}

func (m RunaiUsedGpuMemoryBytesPerNodeMetric) IsQuantile() bool {
	return false
}
