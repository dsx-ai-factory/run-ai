//go:build prometheus_rules
// +build prometheus_rules

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package definitions

import (
	"testing"
)

func TestInferenceThroughputAggregatesHTTPStatusCodes(t *testing.T) {
	newTester(t).Test(inferenceMetricsWithMultipleHTTPStatusCodesTestCase(
		Test{
			Expr:     `sum(runai_inference_throughput_per_pod{pod_uuid="pod-uuid"})`,
			EvalTime: "3m",
			Samples:  []Sample{{Value: 1.5}},
		},
		Test{
			Expr:     `count(runai_inference_throughput_per_pod{pod_uuid="pod-uuid"})`,
			EvalTime: "3m",
			Samples:  []Sample{{Value: 1}},
		},
	))
}

func TestInferenceLatencyIsWeightedAcrossHTTPStatusCodes(t *testing.T) {
	newTester(t).Test(inferenceMetricsWithMultipleHTTPStatusCodesTestCase(
		Test{
			Expr:     `sum(runai_inference_latency_per_pod{pod_uuid="pod-uuid"})`,
			EvalTime: "3m",
			Samples:  []Sample{{Value: 1000}},
		},
	))
}

func TestInferenceMetricsHandleStatusCounterResetBeforeAggregation(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{
				Series: `target_info{instance="pod-instance",k8s_namespace_name="runai-project",k8s_pod_name="inference-pod",kn_revision_name="inference-revision"}`,
				Values: "1x4",
			},
			{
				Series: `kn_serving_invocation_duration_seconds_count{instance="pod-instance",http_response_status_code="200"}`,
				Values: "0 60 0 60",
			},
			{
				Series: `kn_serving_invocation_duration_seconds_count{instance="pod-instance",http_response_status_code="500"}`,
				Values: "0 30 60 90",
			},
			{
				Series: `kn_serving_invocation_duration_seconds_sum{instance="pod-instance",http_response_status_code="200"}`,
				Values: "0 30 0 30",
			},
			{
				Series: `kn_serving_invocation_duration_seconds_sum{instance="pod-instance",http_response_status_code="500"}`,
				Values: "0 60 120 180",
			},
			{
				Series: `runai_pod_info_enriched_nodepool_unique{pod_namespace="runai-project",pod_name="inference-pod",pod_uuid="pod-uuid",workload_id="workload-id"}`,
				Values: "1x4",
			},
		},
		Tests: []Test{
			{
				Expr:     `sum(runai_inference_throughput_per_pod{pod_uuid="pod-uuid"})`,
				EvalTime: "2m",
				Samples:  []Sample{{Value: 0.5}},
			},
			{
				Expr:     `sum(runai_inference_latency_per_pod{pod_uuid="pod-uuid"})`,
				EvalTime: "2m",
				Samples:  []Sample{{Value: 2000}},
			},
		},
	})
}

func TestInferenceMetricsFallBackToLegacyKnativeSeries(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{
				Series: `revision_app_request_count{namespace_name="runai-project",pod_name="inference-pod"}`,
				Values: "0 60 120 180",
			},
			{
				Series: `revision_app_request_latencies_count{namespace_name="runai-project",pod_name="inference-pod"}`,
				Values: "0 60 120 180",
			},
			{
				Series: `revision_app_request_latencies_sum{namespace_name="runai-project",pod_name="inference-pod"}`,
				Values: "0 90000 180000 270000",
			},
			{
				Series: `runai_pod_info_enriched_nodepool_unique{pod_namespace="runai-project",pod_name="inference-pod",pod_uuid="pod-uuid",workload_id="workload-id"}`,
				Values: "1x4",
			},
		},
		Tests: []Test{
			{
				Expr:     `sum(runai_inference_throughput_per_pod{pod_uuid="pod-uuid"})`,
				EvalTime: "3m",
				Samples:  []Sample{{Value: 1}},
			},
			{
				Expr:     `sum(runai_inference_latency_per_pod{pod_uuid="pod-uuid"})`,
				EvalTime: "3m",
				Samples:  []Sample{{Value: 1500}},
			},
		},
	})
}

func inferenceMetricsWithMultipleHTTPStatusCodesTestCase(tests ...Test) TestCase {
	return TestCase{
		Interval: "1m",
		Inputs: []Input{
			{
				Series: `target_info{instance="pod-instance",k8s_namespace_name="runai-project",k8s_pod_name="inference-pod",kn_revision_name="inference-revision"}`,
				Values: "1x5",
			},
			{
				Series: `kn_serving_invocation_duration_seconds_count{instance="pod-instance",http_response_status_code="200"}`,
				Values: "0 60 120 180 240",
			},
			{
				Series: `kn_serving_invocation_duration_seconds_count{instance="pod-instance",http_response_status_code="500"}`,
				Values: "0 30 60 90 120",
			},
			{
				Series: `kn_serving_invocation_duration_seconds_sum{instance="pod-instance",http_response_status_code="200"}`,
				Values: "0 30 60 90 120",
			},
			{
				Series: `kn_serving_invocation_duration_seconds_sum{instance="pod-instance",http_response_status_code="500"}`,
				Values: "0 60 120 180 240",
			},
			{
				Series: `runai_pod_info_enriched_nodepool_unique{pod_namespace="runai-project",pod_name="inference-pod",pod_uuid="pod-uuid",workload_id="workload-id"}`,
				Values: "1x5",
			},
		},
		Tests: tests,
	}
}
