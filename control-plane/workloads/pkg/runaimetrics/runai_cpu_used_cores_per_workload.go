/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiCpuUsedCoresPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
	JobId      []string    `mapstructure:"job_uuid,omitempty"`
}

type RunaiCpuUsedCoresPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
	JobId      bool `mapstructure:"job_uuid,omitempty"`
}

type RunaiCpuUsedCoresPerWorkloadMetric struct {
	Filters        RunaiCpuUsedCoresPerWorkloadFilters
	ResponseLabels RunaiCpuUsedCoresPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiCpuUsedCoresPerWorkloadMetric) GetName() string {
	return "runai_cpu_used_cores_per_workload"
}

func (m RunaiCpuUsedCoresPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiCpuUsedCoresPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiCpuUsedCoresPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiCpuUsedCoresPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiCpuUsedCoresPerWorkloadMetric) IsQuantile() bool {
	return false
}
