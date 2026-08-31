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

// --- GPU Utilization (per nodepool/cluster) ---

func TestGpuUtilizationPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_gpu_utilization_per_gpu{UUID="gpu-1", gpu="0", node="n1", nodepool="np1", modelName="A100"}`, Values: "60x3"},
			{Series: `runai_gpu_utilization_per_gpu{UUID="gpu-2", gpu="1", node="n1", nodepool="np1", modelName="A100"}`, Values: "80x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_utilization_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 70}}},
		},
	})
}

func TestGpuUtilizationPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_gpu_utilization_per_gpu{UUID="gpu-1", gpu="0", node="n1", nodepool="np1", modelName="A100"}`, Values: "40x3"},
			{Series: `runai_gpu_utilization_per_gpu{UUID="gpu-2", gpu="1", node="n2", nodepool="np2", modelName="V100"}`, Values: "60x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_utilization_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 50}}},
		},
	})
}

// --- GPU Memory Used ---

func TestGpuMemoryUsedMebibytesPerGpu(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_dcgm_gpu_used_mebibytes{UUID="gpu-1", gpu="0", node="n1", nodepool="np1", modelName="A100"}`, Values: "10000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_memory_used_mebibytes_per_gpu{UUID="gpu-1"})`, EvalTime: "2m", Samples: []Sample{{Value: 10000}}},
		},
	})
}

func TestGpuMemoryUsedMebibytesPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_gpu_memory_used_mebibytes_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "10000x3"},
			{Series: `runai_gpu_memory_used_mebibytes_per_node{node="n2", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "5000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_memory_used_mebibytes_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 15000}}},
		},
	})
}

func TestGpuMemoryUsedMebibytesPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_gpu_memory_used_mebibytes_per_nodepool{nodepool="np1"}`, Values: "15000x3"},
			{Series: `runai_gpu_memory_used_mebibytes_per_nodepool{nodepool="np2"}`, Values: "8000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_memory_used_mebibytes_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 23000}}},
		},
	})
}

// --- GPU Memory Total ---

func TestGpuMemoryTotalMebibytesPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_gpu_memory_total_mebibytes_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "40000x3"},
			{Series: `runai_gpu_memory_total_mebibytes_per_node{node="n2", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "40000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_memory_total_mebibytes_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 80000}}},
		},
	})
}

func TestGpuMemoryTotalMebibytesPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_gpu_memory_total_mebibytes_per_nodepool{nodepool="np1"}`, Values: "80000x3"},
			{Series: `runai_gpu_memory_total_mebibytes_per_nodepool{nodepool="np2"}`, Values: "40000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_memory_total_mebibytes_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 120000}}},
		},
	})
}

// --- GPU Memory Utilization ---

func TestGpuMemoryUtilizationPerNode(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_gpu_memory_used_mebibytes_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "20000x3"},
			{Series: `runai_gpu_memory_total_mebibytes_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "40000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_memory_utilization_per_node{node="n1"})`, EvalTime: "2m", Samples: []Sample{{Value: 50}}},
		},
	})
}

func TestGpuMemoryUtilizationPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_gpu_memory_used_mebibytes_per_nodepool{nodepool="np1"}`, Values: "30000x3"},
			{Series: `runai_gpu_memory_total_mebibytes_per_nodepool{nodepool="np1"}`, Values: "80000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_memory_utilization_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 37.5}}},
		},
	})
}

func TestGpuMemoryUtilizationPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_gpu_memory_used_mebibytes_per_cluster`, Values: "50000x3"},
			{Series: `runai_gpu_memory_total_mebibytes_per_cluster`, Values: "100000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_memory_utilization_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 50}}},
		},
	})
}

// --- GPU Idle Seconds ---

func TestGpuIdleSecondsPerWorkload(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_last_gpu_utilization_time_per_workload{workload_id="wl-1"}`, Values: "0x5"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_idle_seconds_per_workload{workload_id="wl-1"} >= 0)`, EvalTime: "4m", Samples: []Sample{{Value: 240}}},
		},
	})
}

func TestGpuIdleSecondsPerPod(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_last_gpu_utilization_time_per_pod{pod_uuid="pod-1"}`, Values: "0x5"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_idle_seconds_per_pod{pod_uuid="pod-1"} >= 0)`, EvalTime: "4m", Samples: []Sample{{Value: 240}}},
		},
	})
}

func TestGpuIdleSecondsPerGpu(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_last_gpu_utilization_time_per_gpu{UUID="gpu-1"}`, Values: "0x5"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_idle_seconds_per_gpu{UUID="gpu-1"} >= 0)`, EvalTime: "4m", Samples: []Sample{{Value: 240}}},
		},
	})
}
