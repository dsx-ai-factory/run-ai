/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuMemoryUtilizationPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiGpuMemoryUtilizationPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiGpuMemoryUtilizationPerNodepoolMetric struct {
	Filters        RunaiGpuMemoryUtilizationPerNodepoolFilters
	ResponseLabels RunaiGpuMemoryUtilizationPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiGpuMemoryUtilizationPerNodepoolMetric) GetName() string {
	return "runai_gpu_memory_utilization_per_nodepool"
}

func (m RunaiGpuMemoryUtilizationPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuMemoryUtilizationPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuMemoryUtilizationPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuMemoryUtilizationPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuMemoryUtilizationPerNodepoolMetric) IsQuantile() bool {
	return false
}
