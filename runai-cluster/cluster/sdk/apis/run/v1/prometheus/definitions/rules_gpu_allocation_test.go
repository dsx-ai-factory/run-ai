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

// --- Requested GPUs ---

func TestPendingRequestedGpuCountPerQueue(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_pending_requested_gpu_count_per_workload{queue_name="q1", project="p1", department="d1", nodepool="np1", workload_id="wl-1"}`, Values: "2x3"},
			{Series: `runai_pending_requested_gpu_count_per_workload{queue_name="q1", project="p1", department="d1", nodepool="np1", workload_id="wl-2"}`, Values: "3x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_pending_requested_gpu_count_per_queue{queue_name="q1"})`, EvalTime: "2m", Samples: []Sample{{Value: 5}}},
		},
	})
}

func TestPendingRequestedGpuCountPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_pending_requested_gpu_count_per_queue{queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "3x3"},
			{Series: `runai_pending_requested_gpu_count_per_queue{queue_name="q2", project="p2", department="d1", nodepool="np1"}`, Values: "2x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_pending_requested_gpu_count_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 5}}},
		},
	})
}

func TestPendingRequestedGpuCountPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_pending_requested_gpu_count_per_nodepool{nodepool="np1"}`, Values: "5x3"},
			{Series: `runai_pending_requested_gpu_count_per_nodepool{nodepool="np2"}`, Values: "3x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_pending_requested_gpu_count_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 8}}},
		},
	})
}

// --- Allocated GPUs ---

func TestAllocatedGpuCountPerPod(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_scheduler_allocated_gpus_per_pod_unique{pod_uuid="pod-1"}`, Values: "2x3"},
			{Series: `runai_pod_info_enriched_nodepool_unique{pod_uuid="pod-1", node="n1", nodepool="np1", project="p1", department="d1", queue_name="q1", pod_name="mypod", pod_namespace="ns1", workload_name="wl1", workload_type="train", workload_id="wid-1", gpu="0", modelName="A100", service="svc1"}`, Values: "1x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_allocated_gpu_count_per_pod{pod_uuid="pod-1"})`, EvalTime: "2m", Samples: []Sample{{Value: 2}}},
		},
	})
}

func TestAllocatedGpuCountPerQueue(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_gpu_count_per_pod{pod_uuid="pod-1", queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "2x3"},
			{Series: `runai_allocated_gpu_count_per_pod{pod_uuid="pod-2", queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "1x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_allocated_gpu_count_per_queue{queue_name="q1"})`, EvalTime: "2m", Samples: []Sample{{Value: 3}}},
		},
	})
}

func TestAllocatedGpuCountPerProject(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_gpu_count_per_queue{queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "3x3"},
			{Series: `runai_allocated_gpu_count_per_queue{queue_name="q2", project="p1", department="d1", nodepool="np2"}`, Values: "2x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_allocated_gpu_count_per_project{project="p1"})`, EvalTime: "2m", Samples: []Sample{{Value: 5}}},
		},
	})
}

func TestAllocatedGpuCountPerDepartment(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_gpu_count_per_queue{queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "3x3"},
			{Series: `runai_allocated_gpu_count_per_queue{queue_name="q2", project="p2", department="d1", nodepool="np1"}`, Values: "4x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_allocated_gpu_count_per_department{department="d1"})`, EvalTime: "2m", Samples: []Sample{{Value: 7}}},
		},
	})
}

func TestAllocatedGpuCountPerNode(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_gpu_count_per_pod{pod_uuid="pod-1", node="n1"}`, Values: "2x3"},
			{Series: `runai_allocated_gpu_count_per_pod{pod_uuid="pod-2", node="n1"}`, Values: "1x3"},
			{Series: `runai_node_info_enriched{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "1x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_allocated_gpu_count_per_node{node="n1"})`, EvalTime: "2m", Samples: []Sample{{Value: 3}}},
		},
	})
}

func TestAllocatedGpuCountPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_gpu_count_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "3x3"},
			{Series: `runai_allocated_gpu_count_per_node{node="n2", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "2x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_allocated_gpu_count_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 5}}},
		},
	})
}

func TestAllocatedGpuCountPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_gpu_count_per_nodepool{nodepool="np1"}`, Values: "5x3"},
			{Series: `runai_allocated_gpu_count_per_nodepool{nodepool="np2"}`, Values: "3x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_allocated_gpu_count_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 8}}},
		},
	})
}

func TestAllocatedGpuCountPerGpu_FractionalPods(t *testing.T) {
	// OCP/CDI bug scenario: fractional GPU pods without gpu label, mapped via pod_gpu_info_fixed
	// 3 pods each using 0.7 GPU on different GPUs, 1 GPU free
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_gpu_count_per_pod{pod_uuid="pod-1", node="n1", nodepool="np1"}`, Values: "0.7x3"},
			{Series: `runai_allocated_gpu_count_per_pod{pod_uuid="pod-2", node="n1", nodepool="np1"}`, Values: "0.7x3"},
			{Series: `runai_allocated_gpu_count_per_pod{pod_uuid="pod-3", node="n1", nodepool="np1"}`, Values: "0.7x3"},
			{Series: `runai_pod_gpu_info_fixed{pod_uuid="pod-1", gpu="0"}`, Values: "1x3"},
			{Series: `runai_pod_gpu_info_fixed{pod_uuid="pod-2", gpu="1"}`, Values: "1x3"},
			{Series: `runai_pod_gpu_info_fixed{pod_uuid="pod-3", gpu="2"}`, Values: "1x3"},
			{Series: `runai_gpu_memory_total_mebibytes_per_gpu{UUID="gpu-0", gpu="0", node="n1", nodepool="np1", modelName="T4"}`, Values: "16000x3"},
			{Series: `runai_gpu_memory_total_mebibytes_per_gpu{UUID="gpu-1", gpu="1", node="n1", nodepool="np1", modelName="T4"}`, Values: "16000x3"},
			{Series: `runai_gpu_memory_total_mebibytes_per_gpu{UUID="gpu-2", gpu="2", node="n1", nodepool="np1", modelName="T4"}`, Values: "16000x3"},
			{Series: `runai_gpu_memory_total_mebibytes_per_gpu{UUID="gpu-3", gpu="3", node="n1", nodepool="np1", modelName="T4"}`, Values: "16000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_allocated_gpu_count_per_gpu{node="n1", gpu="0"})`, EvalTime: "2m", Samples: []Sample{{Value: 0.7}}},
			{Expr: `sum(runai_allocated_gpu_count_per_gpu{node="n1", gpu="1"})`, EvalTime: "2m", Samples: []Sample{{Value: 0.7}}},
			{Expr: `sum(runai_allocated_gpu_count_per_gpu{node="n1", gpu="2"})`, EvalTime: "2m", Samples: []Sample{{Value: 0.7}}},
			{Expr: `sum(runai_allocated_gpu_count_per_gpu{node="n1", gpu="3"})`, EvalTime: "2m", Samples: []Sample{{Value: 0}}},
		},
	})
}

func TestAllocatedGpuCountPerGpu_MultiGpuPod(t *testing.T) {
	// Multi-GPU pod using 2 GPUs — allocation (2.0) should be split: 1.0 per GPU
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_gpu_count_per_pod{pod_uuid="pod-1", node="n1", nodepool="np1"}`, Values: "2x3"},
			{Series: `runai_pod_gpu_info_fixed{pod_uuid="pod-1", gpu="0"}`, Values: "1x3"},
			{Series: `runai_pod_gpu_info_fixed{pod_uuid="pod-1", gpu="1"}`, Values: "1x3"},
			{Series: `runai_gpu_memory_total_mebibytes_per_gpu{UUID="gpu-0", gpu="0", node="n1", nodepool="np1", modelName="T4"}`, Values: "16000x3"},
			{Series: `runai_gpu_memory_total_mebibytes_per_gpu{UUID="gpu-1", gpu="1", node="n1", nodepool="np1", modelName="T4"}`, Values: "16000x3"},
			{Series: `runai_gpu_memory_total_mebibytes_per_gpu{UUID="gpu-2", gpu="2", node="n1", nodepool="np1", modelName="T4"}`, Values: "16000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_allocated_gpu_count_per_gpu{node="n1", gpu="0"})`, EvalTime: "2m", Samples: []Sample{{Value: 1}}},
			{Expr: `sum(runai_allocated_gpu_count_per_gpu{node="n1", gpu="1"})`, EvalTime: "2m", Samples: []Sample{{Value: 1}}},
			{Expr: `sum(runai_allocated_gpu_count_per_gpu{node="n1", gpu="2"})`, EvalTime: "2m", Samples: []Sample{{Value: 0}}},
		},
	})
}

func TestAllocatedGpuCountPerGpu_BinPackedFractionalPods(t *testing.T) {
	// 2 fractional pods bin-packed on the same GPU (0.3 + 0.5 = 0.8, must be <= 1.0 per GPU)
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_gpu_count_per_pod{pod_uuid="pod-1", node="n1", nodepool="np1"}`, Values: "0.3x3"},
			{Series: `runai_allocated_gpu_count_per_pod{pod_uuid="pod-2", node="n1", nodepool="np1"}`, Values: "0.5x3"},
			{Series: `runai_pod_gpu_info_fixed{pod_uuid="pod-1", gpu="0"}`, Values: "1x3"},
			{Series: `runai_pod_gpu_info_fixed{pod_uuid="pod-2", gpu="0"}`, Values: "1x3"},
			{Series: `runai_gpu_memory_total_mebibytes_per_gpu{UUID="gpu-0", gpu="0", node="n1", nodepool="np1", modelName="T4"}`, Values: "16000x3"},
			{Series: `runai_gpu_memory_total_mebibytes_per_gpu{UUID="gpu-1", gpu="1", node="n1", nodepool="np1", modelName="T4"}`, Values: "16000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_allocated_gpu_count_per_gpu{node="n1", gpu="0"})`, EvalTime: "2m", Samples: []Sample{{Value: 0.8}}},
			{Expr: `sum(runai_allocated_gpu_count_per_gpu{node="n1", gpu="1"})`, EvalTime: "2m", Samples: []Sample{{Value: 0}}},
		},
	})
}

func TestAllocatedGpuCountPerGpu_RegularK8sWithGpuLabel(t *testing.T) {
	// Regular K8s (non-OCP): runai_allocated_gpu_count_per_pod already has a gpu label
	// (added during pod info enrichment when DCGM metrics work correctly).
	// This exercises the existing fallback path: sum(runai_allocated_gpu_count_per_pod{gpu!=""}) by(gpu, node)
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_gpu_count_per_pod{pod_uuid="pod-1", node="n1", nodepool="np1", gpu="0"}`, Values: "1x3"},
			{Series: `runai_allocated_gpu_count_per_pod{pod_uuid="pod-2", node="n1", nodepool="np1", gpu="1"}`, Values: "1x3"},
			{Series: `runai_gpu_memory_total_mebibytes_per_gpu{UUID="gpu-0", gpu="0", node="n1", nodepool="np1", modelName="T4"}`, Values: "16000x3"},
			{Series: `runai_gpu_memory_total_mebibytes_per_gpu{UUID="gpu-1", gpu="1", node="n1", nodepool="np1", modelName="T4"}`, Values: "16000x3"},
			{Series: `runai_gpu_memory_total_mebibytes_per_gpu{UUID="gpu-2", gpu="2", node="n1", nodepool="np1", modelName="T4"}`, Values: "16000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_allocated_gpu_count_per_gpu{node="n1", gpu="0"})`, EvalTime: "2m", Samples: []Sample{{Value: 1}}},
			{Expr: `sum(runai_allocated_gpu_count_per_gpu{node="n1", gpu="1"})`, EvalTime: "2m", Samples: []Sample{{Value: 1}}},
			{Expr: `sum(runai_allocated_gpu_count_per_gpu{node="n1", gpu="2"})`, EvalTime: "2m", Samples: []Sample{{Value: 0}}},
		},
	})
}

func TestAllocatedGpuCountPerGpu_NoPodGpuInfoFixed(t *testing.T) {
	// No pod_gpu_info_fixed available — all GPUs default to 0
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_gpu_count_per_pod{pod_uuid="pod-1", node="n1", nodepool="np1"}`, Values: "0.7x3"},
			{Series: `runai_gpu_memory_total_mebibytes_per_gpu{UUID="gpu-0", gpu="0", node="n1", nodepool="np1", modelName="T4"}`, Values: "16000x3"},
			{Series: `runai_gpu_memory_total_mebibytes_per_gpu{UUID="gpu-1", gpu="1", node="n1", nodepool="np1", modelName="T4"}`, Values: "16000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_allocated_gpu_count_per_gpu{node="n1", gpu="0"})`, EvalTime: "2m", Samples: []Sample{{Value: 0}}},
			{Expr: `sum(runai_allocated_gpu_count_per_gpu{node="n1", gpu="1"})`, EvalTime: "2m", Samples: []Sample{{Value: 0}}},
		},
	})
}

// --- GPU Allocation % ---

func TestGpuAllocationPercentagePerNode(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_gpu_count_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "3x3"},
			{Series: `runai_gpu_count_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "4x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_allocation_percentage_per_node{node="n1"})`, EvalTime: "2m", Samples: []Sample{{Value: 75}}},
		},
	})
}

func TestGpuAllocationPercentagePerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_gpu_count_per_nodepool{nodepool="np1"}`, Values: "6x3"},
			{Series: `runai_gpu_count_per_nodepool{nodepool="np1"}`, Values: "8x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_allocation_percentage_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 75}}},
		},
	})
}

func TestGpuAllocationPercentagePerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_gpu_count_per_cluster`, Values: "6x3"},
			{Series: `runai_gpu_count_per_cluster`, Values: "10x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_allocation_percentage_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 60}}},
		},
	})
}

// --- Deserved GPUs ---

func TestDeservedGpuCountPerQueue(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_queue_deserved_gpus{queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "4x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_deserved_gpu_count_per_queue{queue_name="q1"})`, EvalTime: "2m", Samples: []Sample{{Value: 4}}},
		},
	})
}

func TestDeservedGpuCountPerProject(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_deserved_gpu_count_per_queue{queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "4x3"},
			{Series: `runai_deserved_gpu_count_per_queue{queue_name="q2", project="p1", department="d1", nodepool="np2"}`, Values: "3x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_deserved_gpu_count_per_project{project="p1"})`, EvalTime: "2m", Samples: []Sample{{Value: 7}}},
		},
	})
}

func TestDeservedGpuCountPerDepartment(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_deserved_gpu_count_per_queue{queue_name="q1", project="", department="d1", nodepool="np1"}`, Values: "5x3"},
			{Series: `runai_deserved_gpu_count_per_queue{queue_name="q2", project="", department="d1", nodepool="np2"}`, Values: "3x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_deserved_gpu_count_per_department{department="d1"})`, EvalTime: "2m", Samples: []Sample{{Value: 8}}},
		},
	})
}
