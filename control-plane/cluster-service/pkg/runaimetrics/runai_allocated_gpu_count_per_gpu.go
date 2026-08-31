/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedGpuCountPerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiAllocatedGpuCountPerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiAllocatedGpuCountPerGpuMetric struct {
	Filters        RunaiAllocatedGpuCountPerGpuFilters
	ResponseLabels RunaiAllocatedGpuCountPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedGpuCountPerGpuMetric) GetName() string {
	return "runai_allocated_gpu_count_per_gpu"
}

func (m RunaiAllocatedGpuCountPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedGpuCountPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedGpuCountPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedGpuCountPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedGpuCountPerGpuMetric) IsQuantile() bool {
	return false
}
