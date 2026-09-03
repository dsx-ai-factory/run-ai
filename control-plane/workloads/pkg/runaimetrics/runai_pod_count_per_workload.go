/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiPodCountPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
	JobId      []string    `mapstructure:"job_uuid,omitempty"`
}

type RunaiPodCountPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
	JobId      bool `mapstructure:"job_uuid,omitempty"`
}

type RunaiPodCountPerWorkloadMetric struct {
	Filters        RunaiPodCountPerWorkloadFilters
	ResponseLabels RunaiPodCountPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiPodCountPerWorkloadMetric) GetName() string {
	return "runai_pod_count_per_workload"
}

func (m RunaiPodCountPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiPodCountPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiPodCountPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiPodCountPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiPodCountPerWorkloadMetric) IsQuantile() bool {
	return false
}
