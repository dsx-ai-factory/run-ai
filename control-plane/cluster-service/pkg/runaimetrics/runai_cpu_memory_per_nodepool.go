/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiTotalMemoryBytesPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiTotalMemoryBytesPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiTotalMemoryBytesPerNodepoolMetric struct {
	Filters        RunaiTotalMemoryBytesPerNodepoolFilters
	ResponseLabels RunaiTotalMemoryBytesPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiTotalMemoryBytesPerNodepoolMetric) GetName() string {
	return "runai_total_memory_bytes_per_nodepool"
}

func (m RunaiTotalMemoryBytesPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiTotalMemoryBytesPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiTotalMemoryBytesPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiTotalMemoryBytesPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiTotalMemoryBytesPerNodepoolMetric) IsQuantile() bool {
	return false
}
