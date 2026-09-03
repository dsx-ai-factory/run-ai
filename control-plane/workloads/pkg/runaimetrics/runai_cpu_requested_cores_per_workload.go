/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiCpuRequestedCoresPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
	JobId      []string    `mapstructure:"job_uuid,omitempty"`
}

type RunaiCpuRequestedCoresPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
	JobId      bool `mapstructure:"job_uuid,omitempty"`
}

type RunaiCpuRequestedCoresPerWorkloadMetric struct {
	Filters        RunaiCpuRequestedCoresPerWorkloadFilters
	ResponseLabels RunaiCpuRequestedCoresPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiCpuRequestedCoresPerWorkloadMetric) GetName() string {
	return "runai_cpu_requested_cores_per_workload"
}

func (m RunaiCpuRequestedCoresPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiCpuRequestedCoresPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiCpuRequestedCoresPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiCpuRequestedCoresPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiCpuRequestedCoresPerWorkloadMetric) IsQuantile() bool {
	return false
}
