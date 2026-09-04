/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiNimNumRequestsWaitingPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
}

type RunaiNimNumRequestsWaitingPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
}

type RunaiNimNumRequestsWaitingPerWorkloadMetric struct {
	Filters        RunaiNimNumRequestsWaitingPerWorkloadFilters
	ResponseLabels RunaiNimNumRequestsWaitingPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiNimNumRequestsWaitingPerWorkloadMetric) GetName() string {
	return "runai_nim_num_requests_waiting_per_workload"
}

func (m RunaiNimNumRequestsWaitingPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiNimNumRequestsWaitingPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiNimNumRequestsWaitingPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiNimNumRequestsWaitingPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiNimNumRequestsWaitingPerWorkloadMetric) IsQuantile() bool {
	return false
}
