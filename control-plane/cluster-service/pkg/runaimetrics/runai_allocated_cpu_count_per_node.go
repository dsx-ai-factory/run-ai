/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedCpuCoresPerNodeFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiAllocatedCpuCoresPerNodeResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
}

type RunaiAllocatedCpuCoresPerNodeMetric struct {
	Filters        RunaiAllocatedCpuCoresPerNodeFilters
	ResponseLabels RunaiAllocatedCpuCoresPerNodeResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedCpuCoresPerNodeMetric) GetName() string {
	return "runai_allocated_cpu_cores_per_node"
}

func (m RunaiAllocatedCpuCoresPerNodeMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedCpuCoresPerNodeMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedCpuCoresPerNodeMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedCpuCoresPerNodeMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedCpuCoresPerNodeMetric) IsQuantile() bool {
	return false
}
