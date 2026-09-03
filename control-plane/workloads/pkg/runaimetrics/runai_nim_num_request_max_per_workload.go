/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiNimNumRequestMaxPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
}

type RunaiNimNumRequestMaxPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
}

type RunaiNimNumRequestMaxPerWorkloadMetric struct {
	Filters        RunaiNimNumRequestMaxPerWorkloadFilters
	ResponseLabels RunaiNimNumRequestMaxPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiNimNumRequestMaxPerWorkloadMetric) GetName() string {
	return "runai_nim_num_request_max_per_workload"
}

func (m RunaiNimNumRequestMaxPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiNimNumRequestMaxPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiNimNumRequestMaxPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiNimNumRequestMaxPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiNimNumRequestMaxPerWorkloadMetric) IsQuantile() bool {
	return false
}
