/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiNimE2eRequestLatencySecondsBucketPerPodFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId     []string    `mapstructure:"pod_uuid,omitempty"`
}

type RunaiNimE2eRequestLatencySecondsBucketPerPodResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
	Le    bool `mapstructure:"le,omitempty"`
}

type RunaiNimE2eRequestLatencySecondsBucketPerPodMetric struct {
	Filters        RunaiNimE2eRequestLatencySecondsBucketPerPodFilters
	ResponseLabels RunaiNimE2eRequestLatencySecondsBucketPerPodResponseLabels
	Quantile       *float32
}

func (m RunaiNimE2eRequestLatencySecondsBucketPerPodMetric) GetName() string {
	return "runai_nim_e2e_request_latency_seconds_bucket_per_pod"
}

func (m RunaiNimE2eRequestLatencySecondsBucketPerPodMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiNimE2eRequestLatencySecondsBucketPerPodMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiNimE2eRequestLatencySecondsBucketPerPodMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiNimE2eRequestLatencySecondsBucketPerPodMetric) IsAggregation() bool {
	return false
}

func (m RunaiNimE2eRequestLatencySecondsBucketPerPodMetric) IsQuantile() bool {
	return false
}
