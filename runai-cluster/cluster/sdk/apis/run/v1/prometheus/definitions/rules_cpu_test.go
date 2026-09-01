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

// --- CPU Cores Total ---

func TestCpuCoresCountPerNode(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `kube_node_status_capacity{node="n1", resource="cpu"}`, Values: "16x3"},
			{Series: `runai_node_info_enriched{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "1x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_cpu_cores_count_per_node{node="n1"})`, EvalTime: "2m", Samples: []Sample{{Value: 16}}},
		},
	})
}

func TestCpuCoresCountPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_cpu_cores_count_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "16x3"},
			{Series: `runai_cpu_cores_count_per_node{node="n2", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "32x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_cpu_cores_count_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 48}}},
		},
	})
}

func TestReadyCpuCoresCountPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_cpu_cores_count_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "16x3"},
			{Series: `runai_cpu_cores_count_per_node{node="n2", nodepool="np1", modelName="A100", ready="false", schedulable="true"}`, Values: "32x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_ready_cpu_cores_count_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 16}}},
		},
	})
}

func TestCpuCoresCountPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_cpu_cores_count_per_nodepool{nodepool="np1"}`, Values: "48x3"},
			{Series: `runai_cpu_cores_count_per_nodepool{nodepool="np2"}`, Values: "64x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_cpu_cores_count_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 112}}},
		},
	})
}

func TestReadyCpuCoresCountPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_ready_cpu_cores_count_per_nodepool{nodepool="np1"}`, Values: "16x3"},
			{Series: `runai_ready_cpu_cores_count_per_nodepool{nodepool="np2"}`, Values: "32x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_ready_cpu_cores_count_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 48}}},
		},
	})
}

// --- CPU Allocation % ---

func TestCpuCoresAllocationPercentagePerNode(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_cpu_cores_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "8x3"},
			{Series: `runai_cpu_cores_count_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "16x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_cpu_cores_allocation_percentage_per_node{node="n1"})`, EvalTime: "2m", Samples: []Sample{{Value: 50}}},
		},
	})
}

func TestCpuCoresAllocationPercentagePerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_cpu_cores_per_nodepool{nodepool="np1"}`, Values: "24x3"},
			{Series: `runai_cpu_cores_count_per_nodepool{nodepool="np1"}`, Values: "48x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_cpu_cores_allocation_percentage_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 50}}},
		},
	})
}

func TestCpuCoresAllocationPercentagePerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_cpu_cores_per_cluster`, Values: "24x3"},
			{Series: `runai_cpu_cores_count_per_cluster`, Values: "48x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_cpu_cores_allocation_percentage_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 50}}},
		},
	})
}

// --- CPU Utilization ---

func TestCpuUtilizationPerQueue(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_used_cpu_cores_per_queue{queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "4x3"},
			{Series: `runai_allocated_cpu_cores_per_queue{queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "8x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_cpu_utilization_per_queue{queue_name="q1"})`, EvalTime: "2m", Samples: []Sample{{Value: 50}}},
		},
	})
}

func TestCpuUtilizationPerNode(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_used_cpu_cores_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "8x3"},
			{Series: `runai_cpu_cores_count_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "16x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_cpu_utilization_per_node{node="n1"})`, EvalTime: "2m", Samples: []Sample{{Value: 50}}},
		},
	})
}

func TestCpuUtilizationPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_used_cpu_cores_per_nodepool{nodepool="np1"}`, Values: "24x3"},
			{Series: `runai_cpu_cores_count_per_nodepool{nodepool="np1"}`, Values: "48x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_cpu_utilization_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 50}}},
		},
	})
}

func TestCpuUtilizationPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_used_cpu_cores_per_cluster`, Values: "24x3"},
			{Series: `runai_cpu_cores_count_per_cluster`, Values: "48x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_cpu_utilization_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 50}}},
		},
	})
}

// --- Allocated CPU Cores ---

func TestAllocatedCpuCoresPerQueue(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_cpu_cores_per_pod{pod_uuid="pod-1", queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "4x3"},
			{Series: `runai_allocated_cpu_cores_per_pod{pod_uuid="pod-2", queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "2x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_allocated_cpu_cores_per_queue{queue_name="q1"})`, EvalTime: "2m", Samples: []Sample{{Value: 6}}},
		},
	})
}

func TestAllocatedCpuCoresPerProject(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_cpu_cores_per_queue{queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "4x3"},
			{Series: `runai_allocated_cpu_cores_per_queue{queue_name="q2", project="p1", department="d1", nodepool="np2"}`, Values: "2x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_allocated_cpu_cores_per_project{project="p1"})`, EvalTime: "2m", Samples: []Sample{{Value: 6}}},
		},
	})
}

func TestAllocatedCpuCoresPerDepartment(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_cpu_cores_per_queue{queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "4x3"},
			{Series: `runai_allocated_cpu_cores_per_queue{queue_name="q2", project="p2", department="d1", nodepool="np1"}`, Values: "3x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_allocated_cpu_cores_per_department{department="d1"})`, EvalTime: "2m", Samples: []Sample{{Value: 7}}},
		},
	})
}

func TestAllocatedCpuCoresPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_cpu_cores_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "8x3"},
			{Series: `runai_allocated_cpu_cores_per_node{node="n2", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "4x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_allocated_cpu_cores_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 12}}},
		},
	})
}

func TestAllocatedCpuCoresPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_cpu_cores_per_nodepool{nodepool="np1"}`, Values: "12x3"},
			{Series: `runai_allocated_cpu_cores_per_nodepool{nodepool="np2"}`, Values: "8x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_allocated_cpu_cores_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 20}}},
		},
	})
}

// --- Used CPU Cores ---

func TestUsedCpuCoresPerQueue(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_used_cpu_cores_per_workload{workload_id="wl-1", queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "3x3"},
			{Series: `runai_used_cpu_cores_per_workload{workload_id="wl-2", queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "2x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_used_cpu_cores_per_queue{queue_name="q1"})`, EvalTime: "2m", Samples: []Sample{{Value: 5}}},
		},
	})
}

func TestUsedCpuCoresPerProject(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_used_cpu_cores_per_queue{queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "5x3"},
			{Series: `runai_used_cpu_cores_per_queue{queue_name="q2", project="p1", department="d1", nodepool="np2"}`, Values: "3x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_used_cpu_cores_per_project{project="p1"})`, EvalTime: "2m", Samples: []Sample{{Value: 8}}},
		},
	})
}

func TestUsedCpuCoresPerDepartment(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_used_cpu_cores_per_queue{queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "5x3"},
			{Series: `runai_used_cpu_cores_per_queue{queue_name="q2", project="p2", department="d1", nodepool="np1"}`, Values: "3x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_used_cpu_cores_per_department{department="d1"})`, EvalTime: "2m", Samples: []Sample{{Value: 8}}},
		},
	})
}

func TestUsedCpuCoresPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_used_cpu_cores_per_node{node="n1", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "8x3"},
			{Series: `runai_used_cpu_cores_per_node{node="n2", nodepool="np1", modelName="A100", ready="true", schedulable="true"}`, Values: "4x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_used_cpu_cores_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 12}}},
		},
	})
}

func TestUsedCpuCoresPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_used_cpu_cores_per_nodepool{nodepool="np1"}`, Values: "12x3"},
			{Series: `runai_used_cpu_cores_per_nodepool{nodepool="np2"}`, Values: "8x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_used_cpu_cores_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 20}}},
		},
	})
}

// --- Used CPU Cores per pod: container="" fallback for sandboxed runtimes ---

// Sandboxed pods (e.g. Kata/VM) expose only container="" cgroups. The rule falls back to
// max(...{container=""}) — max avoids double-counting the hierarchical cgroups. A pod with
// container!="" series ignores the fallback.
func TestUsedCpuCoresPerPodEmptyContainerFallback(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `container_cpu_usage_seconds_total{container="app", pod="leaf-pod", namespace="ns"}`, Values: "0+120x5"},
			{Series: `container_cpu_usage_seconds_total{container="", id="/pod-leaf.slice", pod="leaf-pod", namespace="ns"}`, Values: "0+120x5"},
			{Series: `container_cpu_usage_seconds_total{container="", id="/pod-vm.slice", pod="vm-pod", namespace="ns"}`, Values: "0+180x5"},
			{Series: `container_cpu_usage_seconds_total{container="", id="/pod-vm.slice/child", pod="vm-pod", namespace="ns"}`, Values: "0+120x5"},
			{Series: `container_cpu_usage_seconds_total{container="", id="/overhead-vm", pod="vm-pod", namespace="ns"}`, Values: "0+60x5"},
			{Series: `runai_pod_info_enriched_nodepool_unique{pod_name="leaf-pod", pod_namespace="ns"}`, Values: "1x5"},
			{Series: `runai_pod_info_enriched_nodepool_unique{pod_name="vm-pod", pod_namespace="ns"}`, Values: "1x5"},
		},
		Tests: []Test{
			{Expr: `sum(runai_used_cpu_cores_per_pod{pod_name="leaf-pod"})`, EvalTime: "5m", Samples: []Sample{{Value: 2}}},
			{Expr: `sum(runai_used_cpu_cores_per_pod{pod_name="vm-pod"})`, EvalTime: "5m", Samples: []Sample{{Value: 3}}},
		},
	})
}

// --- Deserved CPU Cores ---

func TestDeservedCpuCoresPerQueue(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_queue_quota_cpu_cores{queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "8x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_deserved_cpu_cores_per_queue{queue_name="q1"})`, EvalTime: "2m", Samples: []Sample{{Value: 8}}},
		},
	})
}

func TestDeservedCpuCoresPerProject(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_deserved_cpu_cores_per_queue{queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "8x3"},
			{Series: `runai_deserved_cpu_cores_per_queue{queue_name="q2", project="p1", department="d1", nodepool="np2"}`, Values: "4x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_deserved_cpu_cores_per_project{project="p1"})`, EvalTime: "2m", Samples: []Sample{{Value: 12}}},
		},
	})
}

func TestDeservedCpuCoresPerDepartment(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_deserved_cpu_cores_per_queue{queue_name="q1", project="", department="d1", nodepool="np1"}`, Values: "8x3"},
			{Series: `runai_deserved_cpu_cores_per_queue{queue_name="q2", project="", department="d1", nodepool="np2"}`, Values: "4x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_deserved_cpu_cores_per_department{department="d1"})`, EvalTime: "2m", Samples: []Sample{{Value: 12}}},
		},
	})
}
