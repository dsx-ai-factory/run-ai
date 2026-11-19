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

func TestVLLMTokenTotalsSurvivePodRestart(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			// pod-1 restarts at 2m, pod-2 keeps counting
			{Series: `vllm:prompt_tokens_total{pod="pod-1",namespace="project",engine="0"}`, Values: "0 120 0 120"},
			{Series: `vllm:prompt_tokens_total{pod="pod-2",namespace="project",engine="0"}`, Values: "0 50 100 150"},
			{Series: `vllm:generation_tokens_total{pod="pod-1",namespace="project",engine="0"}`, Values: "0 60 120 180"},
			{Series: `runai_pod_info_enriched{pod_name="pod-1",pod_namespace="project",pod_uuid="uuid-1",workload_id="workload"}`, Values: "1x4"},
			{Series: `runai_pod_info_enriched{pod_name="pod-2",pod_namespace="project",pod_uuid="uuid-2",workload_id="workload"}`, Values: "1x4"},
		},
		Tests: []Test{
			{Expr: `sum(runai_vllm_prompt_tokens_total_per_pod{pod_uuid="uuid-1"})`, EvalTime: "3m", Samples: []Sample{{Value: 120}}},
			{Expr: `count(runai_vllm_prompt_tokens_total_per_pod{workload_id="workload"})`, EvalTime: "3m", Samples: []Sample{{Value: 2}}},
			{Expr: `sum(runai_vllm_generation_tokens_total_per_pod{pod_uuid="uuid-1"})`, EvalTime: "3m", Samples: []Sample{{Value: 180}}},
			// increase() per pod, then sum: 120 + 120 (pod-1 across restart) + 150 (pod-2)
			{Expr: `sum(increase(runai_vllm_prompt_tokens_total_per_pod{workload_id="workload"}[4m]))`, EvalTime: "3m", Samples: []Sample{{Value: 390}}},
		},
	})
}

func TestVLLMTokenTotalsKeepEnginesSeparate(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			// Two data-parallel engines in one pod. Engine 1 restarts at 2m while engine 0 keeps
			// counting. Summed per pod, the series would read 0 150 160 270 and hide the reset.
			{Series: `vllm:prompt_tokens_total{pod="pod-1",namespace="project",engine="0"}`, Values: "0 50 100 150"},
			{Series: `vllm:prompt_tokens_total{pod="pod-1",namespace="project",engine="1"}`, Values: "0 100 60 120"},
			{Series: `runai_pod_info_enriched{pod_name="pod-1",pod_namespace="project",pod_uuid="uuid-1",workload_id="workload"}`, Values: "1x4"},
		},
		Tests: []Test{
			{Expr: `count(runai_vllm_prompt_tokens_total_per_pod{pod_uuid="uuid-1"})`, EvalTime: "3m", Samples: []Sample{{Value: 2}}},
			// 150 (engine 0) + 100 + 120 (engine 1 across its restart)
			{Expr: `sum(increase(runai_vllm_prompt_tokens_total_per_pod{workload_id="workload"}[4m]))`, EvalTime: "3m", Samples: []Sample{{Value: 370}}},
		},
	})
}

func TestVLLMTokenTotalsCountDeletedPods(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			// pod-2 is scaled down after 2m; its usage before that stays in the workload total.
			{Series: `vllm:prompt_tokens_total{pod="pod-1",namespace="project",engine="0"}`, Values: "0 100 200 300 400 500"},
			{Series: `vllm:prompt_tokens_total{pod="pod-2",namespace="project",engine="0"}`, Values: "0 40 80 stale"},
			{Series: `runai_pod_info_enriched{pod_name="pod-1",pod_namespace="project",pod_uuid="uuid-1",workload_id="workload"}`, Values: "1x6"},
			{Series: `runai_pod_info_enriched{pod_name="pod-2",pod_namespace="project",pod_uuid="uuid-2",workload_id="workload"}`, Values: "1 1 1 stale"},
		},
		Tests: []Test{
			{Expr: `count(runai_vllm_prompt_tokens_total_per_pod{workload_id="workload"})`, EvalTime: "5m", Samples: []Sample{{Value: 1}}},
			// 500 (pod-1) + 80 (pod-2) + 20, the standard increase() extrapolation of half a
			// scrape interval past pod-2's last sample.
			{Expr: `sum(increase(runai_vllm_prompt_tokens_total_per_pod{workload_id="workload"}[6m]))`, EvalTime: "5m", Samples: []Sample{{Value: 600}}},
		},
	})
}

func TestVLLMTokenTotalsCountEveryNodePool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			// One workload with replicas in two node pools. runai_pod_info_enriched_nodepool_unique
			// keeps one node pool per workload, so the token rules must join runai_pod_info_enriched.
			{Series: `vllm:prompt_tokens_total{pod="pod-1",namespace="project",engine="0"}`, Values: "0 100 200 300"},
			{Series: `vllm:prompt_tokens_total{pod="pod-2",namespace="project",engine="0"}`, Values: "0 10 20 30"},
			{Series: `vllm:request_prompt_tokens_count{pod="pod-1",namespace="project",engine="0"}`, Values: "0 1 2 3"},
			{Series: `vllm:request_prompt_tokens_count{pod="pod-2",namespace="project",engine="0"}`, Values: "0 2 4 6"},
			{Series: `runai_pod_info_enriched{pod_name="pod-1",pod_namespace="project",pod_uuid="uuid-1",workload_id="workload",nodepool="pool-a"}`, Values: "1x4"},
			{Series: `runai_pod_info_enriched{pod_name="pod-2",pod_namespace="project",pod_uuid="uuid-2",workload_id="workload",nodepool="pool-b"}`, Values: "1x4"},
		},
		Tests: []Test{
			{Expr: `count(runai_pod_info_enriched_nodepool_unique{workload_id="workload"})`, EvalTime: "3m", Samples: []Sample{{Value: 1}}},
			{Expr: `count(runai_vllm_prompt_tokens_total_per_pod{workload_id="workload"})`, EvalTime: "3m", Samples: []Sample{{Value: 2}}},
			{Expr: `sum(increase(runai_vllm_prompt_tokens_total_per_pod{workload_id="workload"}[4m]))`, EvalTime: "3m", Samples: []Sample{{Value: 330}}},
			{Expr: `sum(increase(runai_vllm_request_prompt_tokens_count_per_pod{workload_id="workload"}[4m]))`, EvalTime: "3m", Samples: []Sample{{Value: 9}}},
		},
	})
}

func TestVLLMCachedPromptTokensOnlyWhenExposed(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			// pod-1 runs a vLLM version with the cached counter, pod-2 an older one without it.
			{Series: `vllm:prompt_tokens_total{pod="pod-1",namespace="project",engine="0"}`, Values: "0 100 200"},
			{Series: `vllm:prompt_tokens_cached_total{pod="pod-1",namespace="project",engine="0"}`, Values: "0 30 60"},
			{Series: `vllm:prompt_tokens_total{pod="pod-2",namespace="project",engine="0"}`, Values: "0 100 200"},
			{Series: `runai_pod_info_enriched{pod_name="pod-1",pod_namespace="project",pod_uuid="uuid-1",workload_id="workload"}`, Values: "1x3"},
			{Series: `runai_pod_info_enriched{pod_name="pod-2",pod_namespace="project",pod_uuid="uuid-2",workload_id="workload"}`, Values: "1x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_vllm_prompt_tokens_cached_total_per_pod{pod_uuid="uuid-1"})`, EvalTime: "2m", Samples: []Sample{{Value: 60}}},
			// No zero-valued series is invented for the older engine.
			{Expr: `runai_vllm_prompt_tokens_cached_total_per_pod{pod_uuid="uuid-2"}`, EvalTime: "2m"},
		},
	})
}

func TestVLLMRequestTokenHistogramsKeepEngineAndLE(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			// Engine 0 keeps counting; engine 1 restarts at 2m.
			{Series: `vllm:request_prompt_tokens_bucket{pod="pod-1",namespace="project",engine="0",le="100"}`, Values: "0 2 4 6"},
			{Series: `vllm:request_prompt_tokens_bucket{pod="pod-1",namespace="project",engine="0",le="500"}`, Values: "0 3 6 9"},
			{Series: `vllm:request_prompt_tokens_bucket{pod="pod-1",namespace="project",engine="0",le="+Inf"}`, Values: "0 4 8 12"},
			{Series: `vllm:request_prompt_tokens_sum{pod="pod-1",namespace="project",engine="0"}`, Values: "0 1000 2000 3000"},
			{Series: `vllm:request_prompt_tokens_count{pod="pod-1",namespace="project",engine="0"}`, Values: "0 4 8 12"},
			{Series: `vllm:request_prompt_tokens_bucket{pod="pod-1",namespace="project",engine="1",le="100"}`, Values: "0 2 0 2"},
			{Series: `vllm:request_prompt_tokens_bucket{pod="pod-1",namespace="project",engine="1",le="500"}`, Values: "0 3 0 3"},
			{Series: `vllm:request_prompt_tokens_bucket{pod="pod-1",namespace="project",engine="1",le="+Inf"}`, Values: "0 4 0 4"},
			{Series: `vllm:request_prompt_tokens_sum{pod="pod-1",namespace="project",engine="1"}`, Values: "0 1000 0 1000"},
			{Series: `vllm:request_prompt_tokens_count{pod="pod-1",namespace="project",engine="1"}`, Values: "0 4 0 4"},
			{Series: `vllm:request_generation_tokens_bucket{pod="pod-1",namespace="project",engine="0",le="+Inf"}`, Values: "0 4 8 12"},
			{Series: `vllm:request_generation_tokens_sum{pod="pod-1",namespace="project",engine="0"}`, Values: "0 200 400 600"},
			{Series: `vllm:request_generation_tokens_count{pod="pod-1",namespace="project",engine="0"}`, Values: "0 4 8 12"},
			{Series: `runai_pod_info_enriched{pod_name="pod-1",pod_namespace="project",pod_uuid="uuid-1",workload_id="workload"}`, Values: "1x4"},
		},
		Tests: []Test{
			{Expr: `count(runai_vllm_request_prompt_tokens_bucket_per_pod{pod_uuid="uuid-1"})`, EvalTime: "3m", Samples: []Sample{{Value: 6}}},
			{Expr: `count(runai_vllm_request_prompt_tokens_bucket_per_pod{pod_uuid="uuid-1",engine="1",le="500"})`, EvalTime: "3m", Samples: []Sample{{Value: 1}}},
			// Request count: 12 (engine 0) + 4 + 4 (engine 1 across its restart).
			{Expr: `sum(increase(runai_vllm_request_prompt_tokens_count_per_pod{workload_id="workload"}[4m]))`, EvalTime: "3m", Samples: []Sample{{Value: 20}}},
			// Average prompt tokens per request: 5000 / 20.
			{
				Expr:     `sum(increase(runai_vllm_request_prompt_tokens_sum_per_pod{workload_id="workload"}[4m])) / sum(increase(runai_vllm_request_prompt_tokens_count_per_pod{workload_id="workload"}[4m]))`,
				EvalTime: "3m",
				Samples:  []Sample{{Value: 250}},
			},
			// p50: 10 of the 20 requests are in the le=100 bucket.
			{
				Expr:     `histogram_quantile(0.5, sum by (le) (increase(runai_vllm_request_prompt_tokens_bucket_per_pod{workload_id="workload"}[4m])))`,
				EvalTime: "3m",
				Samples:  []Sample{{Value: 100}},
			},
			{Expr: `sum(increase(runai_vllm_request_generation_tokens_sum_per_pod{workload_id="workload"}[4m]))`, EvalTime: "3m", Samples: []Sample{{Value: 600}}},
			{Expr: `sum(increase(runai_vllm_request_generation_tokens_count_per_pod{workload_id="workload"}[4m]))`, EvalTime: "3m", Samples: []Sample{{Value: 12}}},
			// Average generation tokens per request: 600 / 12.
			{
				Expr:     `sum(increase(runai_vllm_request_generation_tokens_sum_per_pod{workload_id="workload"}[4m])) / sum(increase(runai_vllm_request_generation_tokens_count_per_pod{workload_id="workload"}[4m]))`,
				EvalTime: "3m",
				Samples:  []Sample{{Value: 50}},
			},
			{Expr: `count(runai_vllm_request_generation_tokens_bucket_per_pod{pod_uuid="uuid-1",le="+Inf"})`, EvalTime: "3m", Samples: []Sample{{Value: 1}}},
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
