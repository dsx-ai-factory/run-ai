/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiNimRequestSuccessTotalPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
}

type RunaiNimRequestSuccessTotalPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
}

type RunaiNimRequestSuccessTotalPerWorkloadMetric struct {
	Filters        RunaiNimRequestSuccessTotalPerWorkloadFilters
	ResponseLabels RunaiNimRequestSuccessTotalPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiNimRequestSuccessTotalPerWorkloadMetric) GetName() string {
	return "runai_nim_request_success_total_per_workload"
}

func (m RunaiNimRequestSuccessTotalPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiNimRequestSuccessTotalPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiNimRequestSuccessTotalPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiNimRequestSuccessTotalPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiNimRequestSuccessTotalPerWorkloadMetric) IsQuantile() bool {
	return false
}
