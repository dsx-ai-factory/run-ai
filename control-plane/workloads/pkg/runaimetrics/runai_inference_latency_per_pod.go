/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiInferenceLatencyPerPodFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId     []string    `mapstructure:"pod_uuid,omitempty"`
}

type RunaiInferenceLatencyPerPodResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
}

type RunaiInferenceLatencyPerPodMetric struct {
	Filters        RunaiInferenceLatencyPerPodFilters
	ResponseLabels RunaiInferenceLatencyPerPodResponseLabels
	Quantile       *float32
}

func (m RunaiInferenceLatencyPerPodMetric) GetName() string {
	return "runai_inference_latency_per_pod"
}

func (m RunaiInferenceLatencyPerPodMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiInferenceLatencyPerPodMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiInferenceLatencyPerPodMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiInferenceLatencyPerPodMetric) IsAggregation() bool {
	return false
}

func (m RunaiInferenceLatencyPerPodMetric) IsQuantile() bool {
	return false
}
