/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiIdleAllocatedGpuCountPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiIdleAllocatedGpuCountPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiIdleAllocatedGpuCountPerNodepoolMetric struct {
	Filters        RunaiIdleAllocatedGpuCountPerNodepoolFilters
	ResponseLabels RunaiIdleAllocatedGpuCountPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiIdleAllocatedGpuCountPerNodepoolMetric) GetName() string {
	return "runai_idle_allocated_gpu_count_per_nodepool"
}

func (m RunaiIdleAllocatedGpuCountPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiIdleAllocatedGpuCountPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiIdleAllocatedGpuCountPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiIdleAllocatedGpuCountPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiIdleAllocatedGpuCountPerNodepoolMetric) IsQuantile() bool {
	return false
}
