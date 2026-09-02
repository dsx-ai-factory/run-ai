/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiNimE2eRequestLatencySecondsBucketPercentilePerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
}

type RunaiNimE2eRequestLatencySecondsBucketPercentilePerWorkloadResponseLabels struct {
	WorkloadId    bool `mapstructure:"workload_id,omitempty"`
	QuantileLabel bool `mapstructure:"quantile_label,omitempty"`
}

type RunaiNimE2eRequestLatencySecondsBucketPercentilePerWorkloadMetric struct {
	Filters        RunaiNimE2eRequestLatencySecondsBucketPercentilePerWorkloadFilters
	ResponseLabels RunaiNimE2eRequestLatencySecondsBucketPercentilePerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiNimE2eRequestLatencySecondsBucketPercentilePerWorkloadMetric) GetName() string {
	return "runai_nim_e2e_request_latency_seconds_bucket_percentile_per_workload"
}

func (m RunaiNimE2eRequestLatencySecondsBucketPercentilePerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiNimE2eRequestLatencySecondsBucketPercentilePerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiNimE2eRequestLatencySecondsBucketPercentilePerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiNimE2eRequestLatencySecondsBucketPercentilePerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiNimE2eRequestLatencySecondsBucketPercentilePerWorkloadMetric) IsQuantile() bool {
	return false
}
