/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

type VLLMTokenUsageFilters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	WorkloadId []string `mapstructure:"workload_id,omitempty"`
}

type VLLMTokenUsageResponseLabels struct {
	WorkloadId    bool `mapstructure:"workload_id,omitempty"`
	QuantileLabel bool `mapstructure:"quantile_label,omitempty"`
}

// VLLMTokenUsageMetric mirrors the workloads service's vLLM token usage metrics: an aggregation
// over the requested period, queried by the metrics-service key name.
type VLLMTokenUsageMetric struct {
	Name           string
	Filters        VLLMTokenUsageFilters
	ResponseLabels VLLMTokenUsageResponseLabels
}

func (m VLLMTokenUsageMetric) GetName() string {
	return m.Name
}

func (m VLLMTokenUsageMetric) GetFilters() interface{} {
	return m.Filters
}

func (m VLLMTokenUsageMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m VLLMTokenUsageMetric) GetQuantile() *float32 {
	return nil
}

func (m VLLMTokenUsageMetric) IsAggregation() bool {
	return true
}

func (m VLLMTokenUsageMetric) IsQuantile() bool {
	return false
}
