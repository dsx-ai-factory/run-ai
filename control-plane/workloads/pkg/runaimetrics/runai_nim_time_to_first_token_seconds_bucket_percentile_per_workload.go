/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiNimTimeToFirstTokenSecondsBucketPercentilePerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
}

type RunaiNimTimeToFirstTokenSecondsBucketPercentilePerWorkloadResponseLabels struct {
	WorkloadId    bool `mapstructure:"workload_id,omitempty"`
	QuantileLabel bool `mapstructure:"quantile_label,omitempty"`
}

type RunaiNimTimeToFirstTokenSecondsBucketPercentilePerWorkloadMetric struct {
	Filters        RunaiNimTimeToFirstTokenSecondsBucketPercentilePerWorkloadFilters
	ResponseLabels RunaiNimTimeToFirstTokenSecondsBucketPercentilePerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPercentilePerWorkloadMetric) GetName() string {
	return "runai_nim_time_to_first_token_seconds_bucket_percentile_per_workload"
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPercentilePerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPercentilePerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPercentilePerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPercentilePerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPercentilePerWorkloadMetric) IsQuantile() bool {
	return false
}
