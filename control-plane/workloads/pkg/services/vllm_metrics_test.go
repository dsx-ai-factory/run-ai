/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metricsapi "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
)

var _ = Describe("vLLM metrics", func() {
	Context("API metric enums", func() {
		It("keeps workload and pod metrics aligned with the query mapping", func() {
			workloadTypes := []api.WorkloadMetricType{
				api.WORKLOADMETRICTYPE_VLLM_NUM_REQUESTS_RUNNING,
				api.WORKLOADMETRICTYPE_VLLM_NUM_REQUESTS_WAITING,
				api.WORKLOADMETRICTYPE_VLLM_REQUEST_SUCCESS,
				api.WORKLOADMETRICTYPE_VLLM_REQUEST_SUCCESS_RATE,
				api.WORKLOADMETRICTYPE_VLLM_TIME_TO_FIRST_TOKEN_SECONDS,
				api.WORKLOADMETRICTYPE_VLLM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES,
				api.WORKLOADMETRICTYPE_VLLM_E2_E_REQUEST_LATENCY_SECONDS,
				api.WORKLOADMETRICTYPE_VLLM_E2_E_REQUEST_LATENCY_SECONDS_PERCENTILES,
				api.WORKLOADMETRICTYPE_VLLM_KV_CACHE_USAGE_PERC,
				api.WORKLOADMETRICTYPE_VLLM_PROMPT_TOKENS_PER_SECOND,
				api.WORKLOADMETRICTYPE_VLLM_GENERATION_TOKENS_PER_SECOND,
				api.WORKLOADMETRICTYPE_VLLM_REQUEST_QUEUE_TIME_SECONDS_PERCENTILES,
				api.WORKLOADMETRICTYPE_VLLM_REQUEST_PREFILL_TIME_SECONDS_PERCENTILES,
				api.WORKLOADMETRICTYPE_VLLM_REQUEST_DECODE_TIME_SECONDS_PERCENTILES,
				api.WORKLOADMETRICTYPE_VLLM_INTER_TOKEN_LATENCY_SECONDS_PERCENTILES,
			}
			podTypes := []api.PodMetricType{
				api.PODMETRICTYPE_VLLM_NUM_REQUESTS_RUNNING,
				api.PODMETRICTYPE_VLLM_NUM_REQUESTS_WAITING,
				api.PODMETRICTYPE_VLLM_REQUEST_SUCCESS,
				api.PODMETRICTYPE_VLLM_REQUEST_SUCCESS_RATE,
				api.PODMETRICTYPE_VLLM_TIME_TO_FIRST_TOKEN_SECONDS,
				api.PODMETRICTYPE_VLLM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES,
				api.PODMETRICTYPE_VLLM_E2_E_REQUEST_LATENCY_SECONDS,
				api.PODMETRICTYPE_VLLM_E2_E_REQUEST_LATENCY_SECONDS_PERCENTILES,
				api.PODMETRICTYPE_VLLM_KV_CACHE_USAGE_PERC,
				api.PODMETRICTYPE_VLLM_PROMPT_TOKENS_PER_SECOND,
				api.PODMETRICTYPE_VLLM_GENERATION_TOKENS_PER_SECOND,
				api.PODMETRICTYPE_VLLM_REQUEST_QUEUE_TIME_SECONDS_PERCENTILES,
				api.PODMETRICTYPE_VLLM_REQUEST_PREFILL_TIME_SECONDS_PERCENTILES,
				api.PODMETRICTYPE_VLLM_REQUEST_DECODE_TIME_SECONDS_PERCENTILES,
				api.PODMETRICTYPE_VLLM_INTER_TOKEN_LATENCY_SECONDS_PERCENTILES,
			}

			Expect(workloadTypes).To(HaveLen(len(vllmMetricNames)))
			Expect(podTypes).To(HaveLen(len(vllmMetricNames)))
			for i := range workloadTypes {
				Expect(string(workloadTypes[i])).To(Equal(string(podTypes[i])))
				Expect(isVLLMMetric(string(workloadTypes[i]))).To(BeTrue())
			}
		})
	})

	Context("token usage metrics", func() {
		tokenUsageTypes := []api.WorkloadMetricType{
			api.WORKLOADMETRICTYPE_VLLM_PROMPT_TOKENS_TOTAL,
			api.WORKLOADMETRICTYPE_VLLM_GENERATION_TOKENS_TOTAL,
			api.WORKLOADMETRICTYPE_VLLM_CACHED_PROMPT_TOKENS_TOTAL,
			api.WORKLOADMETRICTYPE_VLLM_FINISHED_ENGINE_REQUEST_COUNT,
			api.WORKLOADMETRICTYPE_VLLM_PROMPT_TOKENS_PER_REQUEST_AVG,
			api.WORKLOADMETRICTYPE_VLLM_GENERATION_TOKENS_PER_REQUEST_AVG,
			api.WORKLOADMETRICTYPE_VLLM_PROMPT_TOKENS_PER_REQUEST_PERCENTILES,
			api.WORKLOADMETRICTYPE_VLLM_GENERATION_TOKENS_PER_REQUEST_PERCENTILES,
		}

		It("are workload-only vLLM metrics that need cluster version 3.1", func() {
			Expect(vllmTokenUsageMetricNames).To(HaveLen(len(tokenUsageTypes)))
			for _, metricType := range tokenUsageTypes {
				Expect(isVLLMMetric(string(metricType))).To(BeTrue())
				Expect(isVLLMTokenUsageMetric(string(metricType))).To(BeTrue())
				Expect(vllmMetricNames).NotTo(HaveKey(string(metricType)))
				Expect(supportedWorkloadsMetricsVersion).To(HaveKeyWithValue(string(metricType), checker.MinVersionVLLMTokenUsageMetrics))
				Expect(supportedPodMetricsVersion).NotTo(HaveKey(string(metricType)))
			}
		})

		It("keep the existing vLLM metrics at their version", func() {
			Expect(isVLLMTokenUsageMetric(string(api.WORKLOADMETRICTYPE_VLLM_PROMPT_TOKENS_PER_SECOND))).To(BeFalse())
			Expect(supportedWorkloadsMetricsVersion).To(HaveKeyWithValue(string(api.WORKLOADMETRICTYPE_VLLM_PROMPT_TOKENS_PER_SECOND), checker.MinVersionVLLMMetrics))
		})

		It("parse the percentile metrics as a histogram", func() {
			Expect(getHistogramMetricKind(string(api.WORKLOADMETRICTYPE_VLLM_PROMPT_TOKENS_PER_REQUEST_PERCENTILES))).To(Equal(histogramMetricPercentiles))
			Expect(getHistogramMetricKind(string(api.WORKLOADMETRICTYPE_VLLM_GENERATION_TOKENS_PER_REQUEST_PERCENTILES))).To(Equal(histogramMetricPercentiles))
			Expect(getHistogramMetricKind(string(api.WORKLOADMETRICTYPE_VLLM_PROMPT_TOKENS_TOTAL))).To(Equal(histogramMetricNone))
		})

		It("validate the period only when token usage metrics are requested", func() {
			now := time.Now()
			Expect(validateRequestedVLLMTokenUsagePeriod([]api.WorkloadMetricType{api.WORKLOADMETRICTYPE_VLLM_PROMPT_TOKENS_PER_SECOND}, now, now.Add(-time.Hour))).To(BeNil())
			Expect(validateRequestedVLLMTokenUsagePeriod(tokenUsageTypes, now.Add(-time.Hour), now)).To(BeNil())

			apiErr := validateRequestedVLLMTokenUsagePeriod(tokenUsageTypes, now, now)
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
		})
	})

	Context("request success labels", func() {
		assertFinishedReasonPreserved := func(metricType string, isPodMetric bool) {
			response := metricsapi.GetMetrics200Response{
				Status:  "success",
				IsRange: true,
				Data: []metricsapi.Result{{
					MetricLabels: map[string]string{"finished_reason": "stop"},
					Measurements: []metricsapi.Measurement{{Timestamp: time.Unix(1, 0), Value: "3"}},
				}},
			}
			composite := api.MetricsCompositeResponse{}
			statusCode := http.StatusOK
			buildMetricsComposite(context.Background(), metricType, &composite, &statusCode, &response, isPodMetric)

			Expect(composite.Measurements).To(HaveLen(1))
			Expect(composite.Measurements[0].Labels).NotTo(BeNil())
			Expect(*composite.Measurements[0].Labels).To(HaveKeyWithValue("finished_reason", "stop"))
		}

		It("preserves the label for workload metrics", func() {
			assertFinishedReasonPreserved(string(api.WORKLOADMETRICTYPE_VLLM_REQUEST_SUCCESS), false)
		})

		It("preserves the label for pod metrics", func() {
			assertFinishedReasonPreserved(string(api.PODMETRICTYPE_VLLM_REQUEST_SUCCESS), true)
		})
	})
})
