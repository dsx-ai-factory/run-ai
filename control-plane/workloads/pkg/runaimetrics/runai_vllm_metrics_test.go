/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRunaiMetrics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Runai metrics suite")
}

var _ = Describe("vLLM metric names", func() {
	Context("per-workload and per-pod wrappers", func() {
		It("uses the recording-rule naming convention", func() {
			names := []VLLMMetricName{
				VLLMNumRequestsRunning,
				VLLMNumRequestsWaiting,
				VLLMRequestSuccess,
				VLLMRequestSuccessRate,
				VLLMKVCacheUsagePerc,
				VLLMPromptTokensPerSecond,
				VLLMGenerationTokensPerSecond,
				VLLMTimeToFirstTokenSecondsBucket,
				VLLME2ERequestLatencySecondsBucket,
				VLLMTimeToFirstTokenSecondsPercentile,
				VLLME2ERequestLatencySecondsPercentile,
				VLLMRequestQueueTimeSecondsPercentile,
				VLLMRequestPrefillTimeSecondsPercentile,
				VLLMRequestDecodeTimeSecondsPercentile,
				VLLMInterTokenLatencySecondsPercentile,
			}

			for _, name := range names {
				workloadName := RunaiVLLMPerWorkloadMetric{Name: name}.GetName()
				podName := RunaiVLLMPerPodMetric{Name: name}.GetName()
				Expect(workloadName).To(Equal("runai_vllm_" + string(name) + "_per_workload"))
				Expect(podName).To(Equal("runai_vllm_" + string(name) + "_per_pod"))
			}
		})
	})

	Context("token usage metrics", func() {
		It("map to the aggregation query keys", func() {
			expected := map[VLLMTokenUsageMetricName]string{
				VLLMPromptTokensTotal:                    "runai_vllm_prompt_tokens_total_per_workload",
				VLLMGenerationTokensTotal:                "runai_vllm_generation_tokens_total_per_workload",
				VLLMCachedPromptTokensTotal:              "runai_vllm_prompt_tokens_cached_total_per_workload",
				VLLMFinishedEngineRequestCount:           "runai_vllm_finished_engine_request_count_per_workload",
				VLLMPromptTokensPerRequestAvg:            "runai_vllm_prompt_tokens_per_request_avg_per_workload",
				VLLMGenerationTokensPerRequestAvg:        "runai_vllm_generation_tokens_per_request_avg_per_workload",
				VLLMPromptTokensPerRequestPercentile:     "runai_vllm_prompt_tokens_per_request_percentile_per_workload",
				VLLMGenerationTokensPerRequestPercentile: "runai_vllm_generation_tokens_per_request_percentile_per_workload",
			}
			for name, queryKey := range expected {
				metric := RunaiVLLMTokenUsagePerWorkloadMetric{Name: name}
				Expect(metric.GetName()).To(Equal(queryKey))
				Expect(metric.IsAggregation()).To(BeTrue())
				Expect(metric.IsQuantile()).To(BeFalse())
			}
		})

		It("leaves the existing vLLM metrics as time series", func() {
			Expect(RunaiVLLMPerWorkloadMetric{Name: VLLMPromptTokensPerSecond}.IsAggregation()).To(BeFalse())
			Expect(RunaiVLLMPerPodMetric{Name: VLLMPromptTokensPerSecond}.IsAggregation()).To(BeFalse())
		})

		It("returns the filters and response labels it was built with, and no quantile", func() {
			filters := RunaiVLLMPerWorkloadFilters{WorkloadId: []string{"workload-1", "workload-2"}}
			responseLabels := RunaiVLLMResponseLabels{WorkloadId: true, QuantileLabel: true}
			metric := RunaiVLLMTokenUsagePerWorkloadMetric{
				Name:           VLLMPromptTokensPerRequestPercentile,
				Filters:        filters,
				ResponseLabels: responseLabels,
			}

			Expect(metric.GetFilters()).To(Equal(filters))
			Expect(metric.GetResponseLabels()).To(Equal(responseLabels))
			Expect(metric.GetQuantile()).To(BeNil())
		})
	})
})
