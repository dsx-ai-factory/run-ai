/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuOomkillSwapOutOfRamCountPerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuOomkillSwapOutOfRamCountPerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuOomkillSwapOutOfRamCountPerGpuMetric struct {
	Filters        RunaiGpuOomkillSwapOutOfRamCountPerGpuFilters
	ResponseLabels RunaiGpuOomkillSwapOutOfRamCountPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuOomkillSwapOutOfRamCountPerGpuMetric) GetName() string {
	return "runai_gpu_oomkill_swap_out_of_ram_count_per_gpu"
}

func (m RunaiGpuOomkillSwapOutOfRamCountPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuOomkillSwapOutOfRamCountPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuOomkillSwapOutOfRamCountPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuOomkillSwapOutOfRamCountPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuOomkillSwapOutOfRamCountPerGpuMetric) IsQuantile() bool {
	return false
}
