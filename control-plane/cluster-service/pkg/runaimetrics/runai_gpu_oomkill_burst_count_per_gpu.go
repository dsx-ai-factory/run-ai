/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuOomkillBurstCountPerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuOomkillBurstCountPerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuOomkillBurstCountPerGpuMetric struct {
	Filters        RunaiGpuOomkillBurstCountPerGpuFilters
	ResponseLabels RunaiGpuOomkillBurstCountPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuOomkillBurstCountPerGpuMetric) GetName() string {
	return "runai_gpu_oomkill_burst_count_per_gpu"
}

func (m RunaiGpuOomkillBurstCountPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuOomkillBurstCountPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuOomkillBurstCountPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuOomkillBurstCountPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuOomkillBurstCountPerGpuMetric) IsQuantile() bool {
	return false
}
