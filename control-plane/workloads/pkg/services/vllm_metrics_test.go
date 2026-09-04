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
