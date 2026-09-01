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

func TestVLLMGaugesAreEnrichedAndAggregated(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `vllm:num_requests_running{pod="pod-1",namespace="project"}`, Values: "2x3"},
			{Series: `vllm:num_requests_running{pod="pod-2",namespace="project"}`, Values: "3x3"},
			{Series: `vllm:num_requests_waiting{pod="pod-1",namespace="project"}`, Values: "4x3"},
			{Series: `vllm:num_requests_waiting{pod="pod-2",namespace="project"}`, Values: "1x3"},
			{Series: `vllm:kv_cache_usage_perc{pod="pod-1",namespace="project",model_name="a"}`, Values: "0.25x3"},
			{Series: `vllm:kv_cache_usage_perc{pod="pod-1",namespace="project",model_name="b"}`, Values: "0.75x3"},
			{Series: `vllm:kv_cache_usage_perc{pod="pod-2",namespace="project",model_name="a"}`, Values: "0.25x3"},
			{Series: `num_requests_running{pod="pod-1",namespace="project"}`, Values: "100x3"},
			{Series: `runai_pod_info_enriched_nodepool_unique{pod_name="pod-1",pod_namespace="project",pod_uuid="uuid-1",workload_id="workload"}`, Values: "1x3"},
			{Series: `runai_pod_info_enriched_nodepool_unique{pod_name="pod-2",pod_namespace="project",pod_uuid="uuid-2",workload_id="workload"}`, Values: "1x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_vllm_num_requests_running_per_pod)`, EvalTime: "2m", Samples: []Sample{{Value: 5}}},
			{Expr: `count(runai_vllm_num_requests_running_per_pod)`, EvalTime: "2m", Samples: []Sample{{Value: 2}}},
			{Expr: `sum(runai_vllm_num_requests_running_per_workload{workload_id="workload"})`, EvalTime: "2m", Samples: []Sample{{Value: 5}}},
			{Expr: `sum(runai_vllm_num_requests_waiting_per_workload{workload_id="workload"})`, EvalTime: "2m", Samples: []Sample{{Value: 5}}},
			{Expr: `sum(runai_vllm_kv_cache_usage_perc_per_workload{workload_id="workload"})`, EvalTime: "2m", Samples: []Sample{{Value: 0.375}}},
			{Expr: `sum(runai_vllm_kv_cache_usage_perc_max_per_workload{workload_id="workload"})`, EvalTime: "2m", Samples: []Sample{{Value: 0.5}}},
		},
	})
}

func TestVLLMCounterRatesHandleResetAndFinishedReasons(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `vllm:request_success_total{pod="pod-1",namespace="project",finished_reason="stop"}`, Values: "0 60 0 60"},
			{Series: `vllm:request_success_total{pod="pod-1",namespace="project",finished_reason="abort"}`, Values: "0 15 30 45"},
			{Series: `vllm:request_success_total{pod="pod-1",namespace="project",finished_reason="error"}`, Values: "0 15 30 45"},
			{Series: `vllm:prompt_tokens_total{pod="pod-1",namespace="project"}`, Values: "0 120 0 120"},
			{Series: `vllm:generation_tokens_total{pod="pod-1",namespace="project"}`, Values: "0 60 120 180"},
			{Series: `request_success_total{pod="pod-1",namespace="project",finished_reason="stop"}`, Values: "0 6000 12000 18000"},
			{Series: `runai_pod_info_enriched_nodepool_unique{pod_name="pod-1",pod_namespace="project",pod_uuid="uuid-1",workload_id="workload"}`, Values: "1x4"},
		},
		Tests: []Test{
			{Expr: `sum(runai_vllm_request_success_per_second_per_pod{pod_uuid="uuid-1"})`, EvalTime: "3m", Samples: []Sample{{Value: 1}}},
			{Expr: `sum(runai_vllm_request_success_per_second_per_workload{workload_id="workload"})`, EvalTime: "3m", Samples: []Sample{{Value: 1}}},
			{Expr: `sum(runai_vllm_request_success_rate_per_pod{pod_uuid="uuid-1"})`, EvalTime: "3m", Samples: []Sample{{Value: 50}}},
			{Expr: `sum(runai_vllm_request_failure_rate_per_pod{pod_uuid="uuid-1"})`, EvalTime: "3m", Samples: []Sample{{Value: 50}}},
			{Expr: `sum(runai_vllm_request_failure_rate_per_workload{workload_id="workload"})`, EvalTime: "3m", Samples: []Sample{{Value: 50}}},
			{Expr: `sum(runai_vllm_prompt_tokens_per_second_per_pod{pod_uuid="uuid-1"})`, EvalTime: "3m", Samples: []Sample{{Value: 1}}},
			{Expr: `sum(runai_vllm_generation_tokens_per_second_per_workload{workload_id="workload"})`, EvalTime: "3m", Samples: []Sample{{Value: 1}}},
		},
	})
}

func TestVLLMHistogramsPreserveLEAndMergePods(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `vllm:time_to_first_token_seconds_bucket{pod="pod-1",namespace="project",le="1"}`, Values: "0 10 20"},
			{Series: `vllm:time_to_first_token_seconds_bucket{pod="pod-1",namespace="project",le="+Inf"}`, Values: "0 12 24"},
			{Series: `vllm:time_to_first_token_seconds_bucket{pod="pod-2",namespace="project",le="1"}`, Values: "0 5 10"},
			{Series: `vllm:time_to_first_token_seconds_bucket{pod="pod-2",namespace="project",le="+Inf"}`, Values: "0 7 14"},
			{Series: `vllm:request_queue_time_seconds_bucket{pod="pod-1",namespace="project",le="1"}`, Values: "0 8 16"},
			{Series: `time_to_first_token_seconds_bucket{pod="pod-1",namespace="project",le="1"}`, Values: "0 1000 2000"},
			{Series: `runai_pod_info_enriched_nodepool_unique{pod_name="pod-1",pod_namespace="project",pod_uuid="uuid-1",workload_id="workload"}`, Values: "1x3"},
			{Series: `runai_pod_info_enriched_nodepool_unique{pod_name="pod-2",pod_namespace="project",pod_uuid="uuid-2",workload_id="workload"}`, Values: "1x3"},
		},
		Tests: []Test{
			{Expr: `count(runai_vllm_time_to_first_token_seconds_bucket_per_pod)`, EvalTime: "2m", Samples: []Sample{{Value: 4}}},
			{Expr: `sum(runai_vllm_time_to_first_token_seconds_bucket_per_workload{workload_id="workload",le="1"})`, EvalTime: "2m", Samples: []Sample{{Value: 30}}},
			{Expr: `sum(runai_vllm_time_to_first_token_seconds_bucket_per_workload{workload_id="workload",le="+Inf"})`, EvalTime: "2m", Samples: []Sample{{Value: 38}}},
			{Expr: `sum(runai_vllm_request_queue_time_seconds_bucket_per_pod{pod_uuid="uuid-1",le="1"})`, EvalTime: "2m", Samples: []Sample{{Value: 16}}},
		},
	})
}
