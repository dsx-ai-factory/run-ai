/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiNimTimeToFirstTokenSecondsBucketPercentilePerPodFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId     []string    `mapstructure:"pod_uuid,omitempty"`
}

type RunaiNimTimeToFirstTokenSecondsBucketPercentilePerPodResponseLabels struct {
	PodId         bool `mapstructure:"pod_uuid,omitempty"`
	QuantileLabel bool `mapstructure:"quantile_label,omitempty"`
}

type RunaiNimTimeToFirstTokenSecondsBucketPercentilePerPodMetric struct {
	Filters        RunaiNimTimeToFirstTokenSecondsBucketPercentilePerPodFilters
	ResponseLabels RunaiNimTimeToFirstTokenSecondsBucketPercentilePerPodResponseLabels
	Quantile       *float32
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPercentilePerPodMetric) GetName() string {
	return "runai_nim_time_to_first_token_seconds_bucket_percentile_per_pod"
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPercentilePerPodMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPercentilePerPodMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPercentilePerPodMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPercentilePerPodMetric) IsAggregation() bool {
	return false
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPercentilePerPodMetric) IsQuantile() bool {
	return false
}
