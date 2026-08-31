/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiCpuCoresPerNodeFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiCpuCoresPerNodeResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
}

type RunaiCpuCoresPerNodeMetric struct {
	Filters        RunaiCpuCoresPerNodeFilters
	ResponseLabels RunaiCpuCoresPerNodeResponseLabels
	Quantile       *float32
}

func (m RunaiCpuCoresPerNodeMetric) GetName() string {
	return "runai_total_cpu_cores_per_node"
}

func (m RunaiCpuCoresPerNodeMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiCpuCoresPerNodeMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiCpuCoresPerNodeMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiCpuCoresPerNodeMetric) IsAggregation() bool {
	return false
}

func (m RunaiCpuCoresPerNodeMetric) IsQuantile() bool {
	return false
}
