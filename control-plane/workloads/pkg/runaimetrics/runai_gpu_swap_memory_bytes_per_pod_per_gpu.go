/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuSwapMemoryBytesPerPodPerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId     []string    `mapstructure:"pod_uuid,omitempty"`
}

type RunaiGpuSwapMemoryBytesPerPodPerGpuResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
	Gpu   bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuSwapMemoryBytesPerPodPerGpuMetric struct {
	Filters        RunaiGpuSwapMemoryBytesPerPodPerGpuFilters
	ResponseLabels RunaiGpuSwapMemoryBytesPerPodPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuSwapMemoryBytesPerPodPerGpuMetric) GetName() string {
	return "runai_swap_memory_bytes_per_pod_per_gpu"
}

func (m RunaiGpuSwapMemoryBytesPerPodPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuSwapMemoryBytesPerPodPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuSwapMemoryBytesPerPodPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuSwapMemoryBytesPerPodPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuSwapMemoryBytesPerPodPerGpuMetric) IsQuantile() bool {
	return false
}
