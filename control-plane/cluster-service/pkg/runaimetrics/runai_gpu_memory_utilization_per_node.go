/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuMemoryUtilizationPerNodeFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuMemoryUtilizationPerNodeResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
}

type RunaiGpuMemoryUtilizationPerNodeMetric struct {
	Filters        RunaiGpuMemoryUtilizationPerNodeFilters
	ResponseLabels RunaiGpuMemoryUtilizationPerNodeResponseLabels
	Quantile       *float32
}

func (m RunaiGpuMemoryUtilizationPerNodeMetric) GetName() string {
	return "runai_gpu_memory_utilization_per_node"
}

func (m RunaiGpuMemoryUtilizationPerNodeMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuMemoryUtilizationPerNodeMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuMemoryUtilizationPerNodeMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuMemoryUtilizationPerNodeMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuMemoryUtilizationPerNodeMetric) IsQuantile() bool {
	return false
}
