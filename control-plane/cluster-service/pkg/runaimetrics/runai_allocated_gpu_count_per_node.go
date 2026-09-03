/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedGpuCountPerNodeFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiAllocatedGpuCountPerNodeResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
}

type RunaiAllocatedGpuCountPerNodeMetric struct {
	Filters        RunaiAllocatedGpuCountPerNodeFilters
	ResponseLabels RunaiAllocatedGpuCountPerNodeResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedGpuCountPerNodeMetric) GetName() string {
	return "runai_allocated_gpu_count_per_node"
}

func (m RunaiAllocatedGpuCountPerNodeMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedGpuCountPerNodeMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedGpuCountPerNodeMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedGpuCountPerNodeMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedGpuCountPerNodeMetric) IsQuantile() bool {
	return false
}
