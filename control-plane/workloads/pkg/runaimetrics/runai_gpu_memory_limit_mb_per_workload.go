/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuMemoryLimitBytesPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
	JobId      []string    `mapstructure:"job_uuid,omitempty"`
}

type RunaiGpuMemoryLimitBytesPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
	JobId      bool `mapstructure:"job_uuid,omitempty"`
}

type RunaiGpuMemoryLimitBytesPerWorkloadMetric struct {
	Filters        RunaiGpuMemoryLimitBytesPerWorkloadFilters
	ResponseLabels RunaiGpuMemoryLimitBytesPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiGpuMemoryLimitBytesPerWorkloadMetric) GetName() string {
	return "runai_gpu_memory_limit_bytes_per_workload"
}

func (m RunaiGpuMemoryLimitBytesPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuMemoryLimitBytesPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuMemoryLimitBytesPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuMemoryLimitBytesPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuMemoryLimitBytesPerWorkloadMetric) IsQuantile() bool {
	return false
}
