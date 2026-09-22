/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package remote_write_relabel_configs

import (
	"regexp"
	"slices"
	"testing"
)

func TestVLLMRemoteWriteAllowlist(t *testing.T) {
	const pattern = "runai_vllm_.*"
	if !slices.Contains(genericMetrics, pattern) {
		t.Fatalf("genericMetrics does not contain %q", pattern)
	}

	configs := &ClusterRemoteWriteRelabelConfigs{}
	re := regexp.MustCompile("^(?:" + configs.getRelabelConfigRegex() + ")$")
	for _, test := range []struct {
		metricNamespace string
		wantMatch       bool
	}{
		{metricNamespace: "runai_vllm_num_requests_running_per_pod;project-a", wantMatch: true},
		{metricNamespace: "runai_vllm_request_success_per_second_per_workload;runai", wantMatch: true},
		{metricNamespace: "runai_vllm_time_to_first_token_seconds_bucket_per_pod;default", wantMatch: true},
		{metricNamespace: "runai_vllm_inter_token_latency_seconds_count_per_workload;inference", wantMatch: true},
		{metricNamespace: "vllm:num_requests_running;project-a", wantMatch: false},
		{metricNamespace: "vllm:request_success_total;runai", wantMatch: false},
		{metricNamespace: "unrelated_metric;project-a", wantMatch: false},
		{metricNamespace: "runai_nim_num_requests_running_per_pod;project-a", wantMatch: true},
	} {
		if got := re.MatchString(test.metricNamespace); got != test.wantMatch {
			t.Errorf("generated relabel regex match for %q = %t, want %t", test.metricNamespace, got, test.wantMatch)
		}
	}
}

func TestPreemptedPodsRemoteWriteAllowlist(t *testing.T) {
	const pattern = "runai_pod_group_evicted_pods_total"
	if !slices.Contains(genericMetrics, pattern) {
		t.Fatalf("genericMetrics does not contain %q", pattern)
	}

	configs := &ClusterRemoteWriteRelabelConfigs{}
	re := regexp.MustCompile("^(?:" + configs.getRelabelConfigRegex() + ")$")
	if !re.MatchString("runai_pod_group_evicted_pods_total;runai") {
		t.Errorf("generated relabel regex does not match %q", "runai_pod_group_evicted_pods_total;runai")
	}
}
