/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import "github.com/google/uuid"

// VLLMMetricName is the engine-specific part of a runai_vllm_* metric query name.
// Recording rules and metrics-service query keys use:
// runai_vllm_<name>_per_pod and runai_vllm_<name>_per_workload.
// Percentile query keys add _bucket_percentile; metrics-service computes them
// from the corresponding _bucket_per_pod recording rule, parallel to NIM.
type VLLMMetricName string

const (
	VLLMNumRequestsRunning                  VLLMMetricName = "num_requests_running"
	VLLMNumRequestsWaiting                  VLLMMetricName = "num_requests_waiting"
	VLLMRequestSuccess                      VLLMMetricName = "request_success_per_second"
	VLLMRequestSuccessRate                  VLLMMetricName = "request_success_rate"
	VLLMKVCacheUsagePerc                    VLLMMetricName = "kv_cache_usage_perc"
	VLLMPromptTokensPerSecond               VLLMMetricName = "prompt_tokens_per_second"
	VLLMGenerationTokensPerSecond           VLLMMetricName = "generation_tokens_per_second"
	VLLMTimeToFirstTokenSecondsBucket       VLLMMetricName = "time_to_first_token_seconds_bucket"
	VLLME2ERequestLatencySecondsBucket      VLLMMetricName = "e2e_request_latency_seconds_bucket"
	VLLMTimeToFirstTokenSecondsPercentile   VLLMMetricName = "time_to_first_token_seconds_bucket_percentile"
	VLLME2ERequestLatencySecondsPercentile  VLLMMetricName = "e2e_request_latency_seconds_bucket_percentile"
	VLLMRequestQueueTimeSecondsPercentile   VLLMMetricName = "request_queue_time_seconds_bucket_percentile"
	VLLMRequestPrefillTimeSecondsPercentile VLLMMetricName = "request_prefill_time_seconds_bucket_percentile"
	VLLMRequestDecodeTimeSecondsPercentile  VLLMMetricName = "request_decode_time_seconds_bucket_percentile"
	VLLMInterTokenLatencySecondsPercentile  VLLMMetricName = "inter_token_latency_seconds_bucket_percentile"
)

type RunaiVLLMPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
}

type RunaiVLLMPerPodFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId     []string    `mapstructure:"pod_uuid,omitempty"`
}

type RunaiVLLMResponseLabels struct {
	WorkloadId     bool `mapstructure:"workload_id,omitempty"`
	PodId          bool `mapstructure:"pod_uuid,omitempty"`
	Le             bool `mapstructure:"le,omitempty"`
	QuantileLabel  bool `mapstructure:"quantile_label,omitempty"`
	FinishedReason bool `mapstructure:"finished_reason,omitempty"`
}

type RunaiVLLMPerWorkloadMetric struct {
	Name           VLLMMetricName
	Filters        RunaiVLLMPerWorkloadFilters
	ResponseLabels RunaiVLLMResponseLabels
	Quantile       *float32
}

func (m RunaiVLLMPerWorkloadMetric) GetName() string {
	return "runai_vllm_" + string(m.Name) + "_per_workload"
}

func (m RunaiVLLMPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiVLLMPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiVLLMPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiVLLMPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiVLLMPerWorkloadMetric) IsQuantile() bool {
	return false
}

type RunaiVLLMPerPodMetric struct {
	Name           VLLMMetricName
	Filters        RunaiVLLMPerPodFilters
	ResponseLabels RunaiVLLMResponseLabels
	Quantile       *float32
}

func (m RunaiVLLMPerPodMetric) GetName() string {
	return "runai_vllm_" + string(m.Name) + "_per_pod"
}

func (m RunaiVLLMPerPodMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiVLLMPerPodMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiVLLMPerPodMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiVLLMPerPodMetric) IsAggregation() bool {
	return false
}

func (m RunaiVLLMPerPodMetric) IsQuantile() bool {
	return false
}
