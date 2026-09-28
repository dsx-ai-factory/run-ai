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

// runai_pod_gpu_info_fixed is what maps a pod to its GPU. For KAI fraction pods the exporter
// cannot populate it, so it is sourced from metricsd instead. These cover both branches and
// the boundary between them.

func TestPodGpuInfoFixed_FractionPodFromMetricsd(t *testing.T) {
	// No runai_pod_gpu_info at all: the exporter never resolves a GPU for KAI fraction pods.
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_pod_gpu_utilization{pod_uuid="pod-1", gpu="3", gpu_uuid="gpu-3", namespace="team-a", pod="frac-1"}`, Values: "0x3"},
		},
		Tests: []Test{
			{
				Expr:     `runai_pod_gpu_info_fixed`,
				EvalTime: "2m",
				Samples:  []Sample{{Labels: `runai_pod_gpu_info_fixed{pod_uuid="pod-1", gpu="3", UUID="gpu-3"}`, Value: 1}},
			},
		},
	})
}

func TestPodGpuInfoFixed_ExporterPathUnchanged(t *testing.T) {
	// The pre-existing branch: gpu on runai_pod_gpu_info actually holds the UUID, and the real
	// index comes from DCGM.
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_pod_gpu_info{pod_uuid="pod-1", gpu="gpu-0"}`, Values: "1x3"},
			{Series: `runai_dcgm_gpu_utilization{UUID="gpu-0", gpu="0"}`, Values: "50x3"},
		},
		Tests: []Test{
			{
				Expr:     `runai_pod_gpu_info_fixed`,
				EvalTime: "2m",
				Samples:  []Sample{{Labels: `runai_pod_gpu_info_fixed{pod_uuid="pod-1", gpu="0", UUID="gpu-0"}`, Value: 1}},
			},
		},
	})
}

func TestPodGpuInfoFixed_NoDoubleCountWhenBothSourcesReport(t *testing.T) {
	// metricsd reports any GPU process it can attribute to a pod, so a pod the exporter already
	// resolved can appear in both. It must still yield exactly one series.
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_pod_gpu_info{pod_uuid="pod-1", gpu="gpu-0"}`, Values: "1x3"},
			{Series: `runai_dcgm_gpu_utilization{UUID="gpu-0", gpu="0"}`, Values: "50x3"},
			{Series: `runai_pod_gpu_utilization{pod_uuid="pod-1", gpu="0", gpu_uuid="gpu-0", namespace="team-a", pod="frac-1"}`, Values: "10x3"},
		},
		Tests: []Test{
			{Expr: `count(runai_pod_gpu_info_fixed) by (pod_uuid)`, EvalTime: "2m", Samples: []Sample{{Labels: `{pod_uuid="pod-1"}`, Value: 1}}},
		},
	})
}

func TestPodGpuInfoFixed_MixedWholeAndFractionPods(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_pod_gpu_info{pod_uuid="whole-1", gpu="gpu-0"}`, Values: "1x3"},
			{Series: `runai_dcgm_gpu_utilization{UUID="gpu-0", gpu="0"}`, Values: "50x3"},
			{Series: `runai_pod_gpu_utilization{pod_uuid="frac-1", gpu="1", gpu_uuid="gpu-1", namespace="team-a", pod="frac-pod"}`, Values: "20x3"},
		},
		Tests: []Test{
			{Expr: `count(runai_pod_gpu_info_fixed)`, EvalTime: "2m", Samples: []Sample{{Value: 2}}},
			{Expr: `count(runai_pod_gpu_info_fixed{pod_uuid="whole-1", gpu="0", UUID="gpu-0"})`, EvalTime: "2m", Samples: []Sample{{Value: 1}}},
			{Expr: `count(runai_pod_gpu_info_fixed{pod_uuid="frac-1", gpu="1", UUID="gpu-1"})`, EvalTime: "2m", Samples: []Sample{{Value: 1}}},
		},
	})
}

func TestGpuUtilizationPerFractionalPod_FromMetricsd(t *testing.T) {
	// The acceptance criterion: a fraction pod known only to metricsd reaches the per-pod rollup.
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_pod_gpu_utilization{pod_uuid="pod-1", gpu="3", gpu_uuid="gpu-3", namespace="team-a", pod="frac-1"}`, Values: "40x3"},
			{Series: `runai_pod_info_enriched_nodepool_unique{pod_uuid="pod-1", pod_name="frac-1", pod_namespace="team-a", workload_name="wl", workload_type="Interactive", workload_id="wl-1", project="p1", department="d1", queue_name="q1", node="n1", nodepool="np1", user="u1"}`, Values: "1x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_utilization_per_fractional_pod)`, EvalTime: "2m", Samples: []Sample{{Value: 40}}},
			{Expr: `sum(runai_gpu_utilization_per_pod)`, EvalTime: "2m", Samples: []Sample{{Value: 40}}},
		},
	})
}

func TestGpuMemoryUsedPerFractionalPod_FromMetricsd(t *testing.T) {
	// metricsd reports bytes; the rule converts to mebibytes. 2 GiB in, 2048 MiB out.
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_pod_gpu_utilization{pod_uuid="pod-1", gpu="3", gpu_uuid="gpu-3", namespace="team-a", pod="frac-1"}`, Values: "0x3"},
			{Series: `runai_pod_gpu_memory_used_bytes{pod_uuid="pod-1", gpu="3", gpu_uuid="gpu-3", namespace="team-a", pod="frac-1"}`, Values: "2147483648x3"},
			{Series: `runai_pod_info_enriched_nodepool_unique{pod_uuid="pod-1", pod_name="frac-1", pod_namespace="team-a", workload_name="wl", workload_type="Interactive", workload_id="wl-1", project="p1", department="d1", queue_name="q1", node="n1", nodepool="np1", user="u1"}`, Values: "1x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_memory_used_mebibytes_per_fractional_pod)`, EvalTime: "2m", Samples: []Sample{{Value: 2048}}},
		},
	})
}

func TestPodGpuInfoFixed_ValueIsOnePerPodGpuRegardlessOfSeriesCount(t *testing.T) {
	// Downstream rules multiply by this value — runai_allocated_gpu_count_per_gpu divides by
	// count(...) by(pod_uuid) and then multiplies the mapping back in — so several metricsd
	// series for one pod and GPU must still yield 1, not their count.
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_pod_gpu_utilization{pod_uuid="pod-1", gpu="3", gpu_uuid="gpu-3", namespace="team-a", pod="frac-1", container="a"}`, Values: "10x3"},
			{Series: `runai_pod_gpu_utilization{pod_uuid="pod-1", gpu="3", gpu_uuid="gpu-3", namespace="team-a", pod="frac-1", container="b"}`, Values: "20x3"},
		},
		Tests: []Test{
			{
				Expr:     `runai_pod_gpu_info_fixed`,
				EvalTime: "2m",
				Samples:  []Sample{{Labels: `runai_pod_gpu_info_fixed{pod_uuid="pod-1", gpu="3", UUID="gpu-3"}`, Value: 1}},
			},
		},
	})
}

func TestPodGpuInfoFixed_WholeGpuPodReportedByMetricsdKeepsNoGpuLabel(t *testing.T) {
	// A whole-GPU pod carries runai_pod_gpu_info with an empty gpu label, which is what
	// wholePodFilter keys on. If metricsd ever reported such a pod, the fallback branch would
	// hand it a gpu label and it would be rolled up as fractional, losing its DCGM-based
	// whole-pod series. Observed on a live cluster: metricsd only reports containers fractiond
	// mapped, so a whole-GPU pod produces nothing — this pins that assumption.
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_pod_gpu_info{pod_uuid="whole-1", gpu=""}`, Values: "1x3"},
			{Series: `runai_dcgm_gpu_utilization{UUID="gpu-0", gpu="0"}`, Values: "50x3"},
		},
		Tests: []Test{
			{
				Expr:     `runai_pod_gpu_info_fixed{pod_uuid="whole-1"}`,
				EvalTime: "2m",
				Samples:  []Sample{},
			},
		},
	})
}
