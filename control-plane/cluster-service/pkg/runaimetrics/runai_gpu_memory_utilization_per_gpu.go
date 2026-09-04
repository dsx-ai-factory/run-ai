/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuMemoryUtilizationPerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuMemoryUtilizationPerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuMemoryUtilizationPerGpuMetric struct {
	Filters        RunaiGpuMemoryUtilizationPerGpuFilters
	ResponseLabels RunaiGpuMemoryUtilizationPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuMemoryUtilizationPerGpuMetric) GetName() string {
	return "runai_gpu_memory_utilization_per_gpu"
}

func (m RunaiGpuMemoryUtilizationPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuMemoryUtilizationPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuMemoryUtilizationPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuMemoryUtilizationPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuMemoryUtilizationPerGpuMetric) IsQuantile() bool {
	return false
}
