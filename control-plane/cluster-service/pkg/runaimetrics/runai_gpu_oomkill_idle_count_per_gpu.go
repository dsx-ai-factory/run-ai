/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuOomkillIdleCountPerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuOomkillIdleCountPerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuOomkillIdleCountPerGpuMetric struct {
	Filters        RunaiGpuOomkillIdleCountPerGpuFilters
	ResponseLabels RunaiGpuOomkillIdleCountPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuOomkillIdleCountPerGpuMetric) GetName() string {
	return "runai_gpu_oomkill_idle_count_per_gpu"
}

func (m RunaiGpuOomkillIdleCountPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuOomkillIdleCountPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuOomkillIdleCountPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuOomkillIdleCountPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuOomkillIdleCountPerGpuMetric) IsQuantile() bool {
	return false
}
