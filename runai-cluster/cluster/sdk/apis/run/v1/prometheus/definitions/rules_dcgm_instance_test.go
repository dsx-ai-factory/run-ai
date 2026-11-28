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

func TestDcgmGpuUtilization_PodIPFromInstance(t *testing.T) {
	for _, tc := range []struct {
		name     string
		instance string
		podIP    string
	}{
		{name: "ipv4", instance: "10.244.0.5:9400", podIP: "10.244.0.5"},
		{name: "ipv6", instance: "[fd00:10:244::24a4]:9400", podIP: "fd00:10:244::24a4"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			newTester(t).Test(TestCase{
				Interval: "1m",
				Inputs: []Input{
					{Series: `DCGM_FI_DEV_GPU_UTIL{instance="` + tc.instance + `", UUID="gpu-1", gpu="0", modelName="T4"}`, Values: "100x3"},
					{Series: `kube_pod_info{created_by_name="nvidia-dcgm-exporter", pod_ip="` + tc.podIP + `", node="n1"}`, Values: "1x3"},
					{Series: `runai_node_nodepool_excluded{node="n1", nodepool="np1"}`, Values: "1x3"},
				},
				Tests: []Test{
					{Expr: `sum(runai_dcgm_gpu_utilization{node="n1", nodepool="np1", UUID="gpu-1"})`, EvalTime: "2m", Samples: []Sample{{Value: 100}}},
				},
			})
		})
	}
}

func TestOomKillMetric_PodIPFromInstance(t *testing.T) {
	for _, tc := range []struct {
		name     string
		instance string
		podIP    string
	}{
		{name: "ipv4", instance: "10.244.0.5:9400", podIP: "10.244.0.5"},
		{name: "ipv6", instance: "[fd00:10:244::24a4]:9400", podIP: "fd00:10:244::24a4"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			newTester(t).Test(TestCase{
				Interval: "1m",
				Inputs: []Input{
					{Series: `runai_gpu_oomkill_burst_count{instance="` + tc.instance + `", UUID="gpu-1", gpu="0"}`, Values: "2x3"},
					{Series: `kube_pod_info{pod_ip="` + tc.podIP + `", node="n1"}`, Values: "1x3"},
					{Series: `runai_node_info_enriched{node="n1", nodepool="np1", modelName="T4"}`, Values: "1x3"},
				},
				Tests: []Test{
					{Expr: `sum(runai_gpu_oomkill_burst_count_per_node{node="n1"})`, EvalTime: "2m", Samples: []Sample{{Value: 2}}},
				},
			})
		})
	}
}
