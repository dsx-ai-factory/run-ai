/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuMemoryRequestedBytesPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
	JobId      []string    `mapstructure:"job_uuid,omitempty"`
}

type RunaiGpuMemoryRequestedBytesPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
	JobId      bool `mapstructure:"job_uuid,omitempty"`
}

type RunaiGpuMemoryRequestedBytesPerWorkloadMetric struct {
	Filters        RunaiGpuMemoryRequestedBytesPerWorkloadFilters
	ResponseLabels RunaiGpuMemoryRequestedBytesPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiGpuMemoryRequestedBytesPerWorkloadMetric) GetName() string {
	return "runai_gpu_memory_requested_bytes_per_workload"
}

func (m RunaiGpuMemoryRequestedBytesPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuMemoryRequestedBytesPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuMemoryRequestedBytesPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuMemoryRequestedBytesPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuMemoryRequestedBytesPerWorkloadMetric) IsQuantile() bool {
	return false
}
