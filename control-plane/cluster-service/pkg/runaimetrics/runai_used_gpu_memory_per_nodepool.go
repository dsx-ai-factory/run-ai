/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUsedGpuMemoryBytesPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiUsedGpuMemoryBytesPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiUsedGpuMemoryBytesPerNodepoolMetric struct {
	Filters        RunaiUsedGpuMemoryBytesPerNodepoolFilters
	ResponseLabels RunaiUsedGpuMemoryBytesPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiUsedGpuMemoryBytesPerNodepoolMetric) GetName() string {
	return "runai_used_gpu_memory_bytes_per_nodepool"
}

func (m RunaiUsedGpuMemoryBytesPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUsedGpuMemoryBytesPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUsedGpuMemoryBytesPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUsedGpuMemoryBytesPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiUsedGpuMemoryBytesPerNodepoolMetric) IsQuantile() bool {
	return false
}
