/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiNimE2eRequestLatencySecondsBucketPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
}

type RunaiNimE2eRequestLatencySecondsBucketPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
	Le         bool `mapstructure:"le,omitempty"`
}

type RunaiNimE2eRequestLatencySecondsBucketPerWorkloadMetric struct {
	Filters        RunaiNimE2eRequestLatencySecondsBucketPerWorkloadFilters
	ResponseLabels RunaiNimE2eRequestLatencySecondsBucketPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiNimE2eRequestLatencySecondsBucketPerWorkloadMetric) GetName() string {
	return "runai_nim_e2e_request_latency_seconds_bucket_per_workload"
}

func (m RunaiNimE2eRequestLatencySecondsBucketPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiNimE2eRequestLatencySecondsBucketPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiNimE2eRequestLatencySecondsBucketPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiNimE2eRequestLatencySecondsBucketPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiNimE2eRequestLatencySecondsBucketPerWorkloadMetric) IsQuantile() bool {
	return false
}
