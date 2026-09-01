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

// --- Memory Total ---

func TestTotalMemoryBytesPerNode(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `kube_node_status_capacity{node="n1", resource="memory"}`, Values: "68719476736x3"},
			{Series: `runai_node_info_enriched{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "1x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_total_memory_bytes_per_node{node="n1"})`, EvalTime: "2m", Samples: []Sample{{Value: 68719476736}}},
		},
	})
}

func TestTotalMemoryBytesPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_total_memory_bytes_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "64000x3"},
			{Series: `runai_total_memory_bytes_per_node{node="n2", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "64000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_total_memory_bytes_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 128000}}},
		},
	})
}

func TestTotalMemoryBytesPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_total_memory_bytes_per_nodepool{nodepool="np1"}`, Values: "128000x3"},
			{Series: `runai_total_memory_bytes_per_nodepool{nodepool="np2"}`, Values: "64000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_total_memory_bytes_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 192000}}},
		},
	})
}

// --- Memory Allocation % ---

func TestMemoryAllocationPercentagePerNode(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_memory_bytes_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "32000x3"},
			{Series: `runai_total_memory_bytes_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "64000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_memory_allocation_percentage_per_node{node="n1"})`, EvalTime: "2m", Samples: []Sample{{Value: 50}}},
		},
	})
}

func TestMemoryAllocationPercentagePerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_memory_bytes_per_nodepool{nodepool="np1"}`, Values: "64000x3"},
			{Series: `runai_total_memory_bytes_per_nodepool{nodepool="np1"}`, Values: "128000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_memory_allocation_percentage_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 50}}},
		},
	})
}

func TestMemoryAllocationPercentagePerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_memory_bytes_per_cluster`, Values: "96000x3"},
			{Series: `runai_total_memory_bytes_per_cluster`, Values: "192000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_memory_allocation_percentage_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 50}}},
		},
	})
}

// --- Memory Utilization ---

func TestMemoryUtilizationPerNode(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_used_memory_bytes_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "32000x3"},
			{Series: `runai_total_memory_bytes_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "64000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_memory_utilization_per_node{node="n1"})`, EvalTime: "2m", Samples: []Sample{{Value: 50}}},
		},
	})
}

func TestMemoryUtilizationPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_used_memory_bytes_per_nodepool{nodepool="np1"}`, Values: "64000x3"},
			{Series: `runai_total_memory_bytes_per_nodepool{nodepool="np1"}`, Values: "128000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_memory_utilization_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 50}}},
		},
	})
}

func TestMemoryUtilizationPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_used_memory_bytes_per_cluster`, Values: "96000x3"},
			{Series: `runai_total_memory_bytes_per_cluster`, Values: "192000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_memory_utilization_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 50}}},
		},
	})
}

// --- Allocated Memory Bytes ---

func TestAllocatedMemoryBytesPerQueue(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_memory_per_pod{pod_uuid="pod-1", queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "4096x3"},
			{Series: `runai_allocated_memory_per_pod{pod_uuid="pod-2", queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "2048x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_allocated_memory_bytes_per_queue{queue_name="q1"})`, EvalTime: "2m", Samples: []Sample{{Value: 6144}}},
		},
	})
}

func TestAllocatedMemoryBytesPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_memory_bytes_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "32000x3"},
			{Series: `runai_allocated_memory_bytes_per_node{node="n2", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "16000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_allocated_memory_bytes_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 48000}}},
		},
	})
}

func TestAllocatedMemoryBytesPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_memory_bytes_per_nodepool{nodepool="np1"}`, Values: "48000x3"},
			{Series: `runai_allocated_memory_bytes_per_nodepool{nodepool="np2"}`, Values: "24000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_allocated_memory_bytes_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 72000}}},
		},
	})
}

// --- Used Memory Bytes ---

func TestUsedMemoryBytesPerQueue(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_used_memory_bytes_per_workload{workload_id="wl-1", queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "4096x3"},
			{Series: `runai_used_memory_bytes_per_workload{workload_id="wl-2", queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "2048x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_used_memory_bytes_per_queue{queue_name="q1"})`, EvalTime: "2m", Samples: []Sample{{Value: 6144}}},
		},
	})
}

func TestUsedMemoryBytesPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_used_memory_bytes_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "16000x3"},
			{Series: `runai_used_memory_bytes_per_node{node="n2", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "8000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_used_memory_bytes_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 24000}}},
		},
	})
}

func TestUsedMemoryBytesPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_used_memory_bytes_per_nodepool{nodepool="np1"}`, Values: "24000x3"},
			{Series: `runai_used_memory_bytes_per_nodepool{nodepool="np2"}`, Values: "12000x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_used_memory_bytes_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 36000}}},
		},
	})
}

// --- Used Memory Bytes per pod: container="" fallback for sandboxed runtimes ---

// Memory mirror of the CPU fallback: max(...{container=""}) returns the outermost cgroup
// total without double-counting nested cgroups.
func TestUsedMemoryBytesPerPodEmptyContainerFallback(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `container_memory_usage_bytes{container="app", pod="leaf-pod", namespace="ns"}`, Values: "100x5"},
			{Series: `container_memory_usage_bytes{container="", id="/pod-leaf.slice", pod="leaf-pod", namespace="ns"}`, Values: "100x5"},
			{Series: `container_memory_usage_bytes{container="", id="/pod-vm.slice", pod="vm-pod", namespace="ns"}`, Values: "200x5"},
			{Series: `container_memory_usage_bytes{container="", id="/pod-vm.slice/child", pod="vm-pod", namespace="ns"}`, Values: "150x5"},
			{Series: `container_memory_usage_bytes{container="", id="/overhead-vm", pod="vm-pod", namespace="ns"}`, Values: "50x5"},
			{Series: `runai_pod_info_enriched_nodepool_unique{pod_name="leaf-pod", pod_namespace="ns"}`, Values: "1x5"},
			{Series: `runai_pod_info_enriched_nodepool_unique{pod_name="vm-pod", pod_namespace="ns"}`, Values: "1x5"},
		},
		Tests: []Test{
			{Expr: `sum(runai_used_memory_bytes_per_pod{pod_name="leaf-pod"})`, EvalTime: "3m", Samples: []Sample{{Value: 100}}},
			{Expr: `sum(runai_used_memory_bytes_per_pod{pod_name="vm-pod"})`, EvalTime: "3m", Samples: []Sample{{Value: 200}}},
		},
	})
}

// --- Deserved Memory Bytes ---

func TestDeservedMemoryBytesPerQueue(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_queue_quota_memory_bytes{queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "65536x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_deserved_memory_bytes_per_queue{queue_name="q1"})`, EvalTime: "2m", Samples: []Sample{{Value: 65536}}},
		},
	})
}

func TestDeservedMemoryBytesPerProject(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_deserved_memory_bytes_per_queue{queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "65536x3"},
			{Series: `runai_deserved_memory_bytes_per_queue{queue_name="q2", project="p1", department="d1", nodepool="np2"}`, Values: "32768x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_deserved_memory_bytes_per_project{project="p1"})`, EvalTime: "2m", Samples: []Sample{{Value: 98304}}},
		},
	})
}

func TestDeservedMemoryBytesPerDepartment(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_deserved_memory_bytes_per_queue{queue_name="q1", project="", department="d1", nodepool="np1"}`, Values: "65536x3"},
			{Series: `runai_deserved_memory_bytes_per_queue{queue_name="q2", project="", department="d1", nodepool="np2"}`, Values: "32768x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_deserved_memory_bytes_per_department{department="d1"})`, EvalTime: "2m", Samples: []Sample{{Value: 98304}}},
		},
	})
}
