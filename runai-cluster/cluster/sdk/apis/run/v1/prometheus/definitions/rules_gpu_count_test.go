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

// --- Total GPUs ---

func TestGpuCountPerNode(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_gpu_memory_total_mebibytes_per_gpu{node="n1", UUID="gpu-1"}`, Values: "40000x3"},
			{Series: `runai_gpu_memory_total_mebibytes_per_gpu{node="n1", UUID="gpu-2"}`, Values: "40000x3"},
			{Series: `runai_node_info_enriched{node="n1", nodepool="default", modelName="A100", ready="true", schedulable="true"}`, Values: "1x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_count_per_node{node="n1"})`, EvalTime: "2m", Samples: []Sample{{Value: 2}}},
		},
	})
}

func TestGpuCountPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_gpu_count_per_node{node="n1", nodepool="default", modelName="A100", ready="true", schedulable="true"}`, Values: "2x3"},
			{Series: `runai_gpu_count_per_node{node="n2", nodepool="default", modelName="A100", ready="true", schedulable="true"}`, Values: "4x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_count_per_nodepool{nodepool="default"})`, EvalTime: "2m", Samples: []Sample{{Value: 6}}},
		},
	})
}

func TestReadyGpuCountPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_gpu_count_per_node{node="n1", nodepool="default", modelName="A100", ready="true", schedulable="true"}`, Values: "2x3"},
			{Series: `runai_gpu_count_per_node{node="n2", nodepool="default", modelName="A100", ready="false", schedulable="true"}`, Values: "4x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_ready_gpu_count_per_nodepool{nodepool="default"})`, EvalTime: "2m", Samples: []Sample{{Value: 2}}},
		},
	})
}

func TestGpuCountPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_gpu_count_per_nodepool{nodepool="np1"}`, Values: "4x3"},
			{Series: `runai_gpu_count_per_nodepool{nodepool="np2"}`, Values: "8x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_count_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 12}}},
		},
	})
}

func TestReadyGpuCountPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_ready_gpu_count_per_nodepool{nodepool="np1"}`, Values: "2x3"},
			{Series: `runai_ready_gpu_count_per_nodepool{nodepool="np2"}`, Values: "3x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_ready_gpu_count_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 5}}},
		},
	})
}

// --- Total GPU Nodes ---

func TestGpuNodesCountPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_gpu_count_per_node{node="n1", nodepool="default", modelName="A100", ready="true", schedulable="true"}`, Values: "2x3"},
			{Series: `runai_gpu_count_per_node{node="n2", nodepool="default", modelName="A100", ready="true", schedulable="true"}`, Values: "4x3"},
			{Series: `runai_gpu_count_per_node{node="n3", nodepool="other", modelName="V100", ready="true", schedulable="true"}`, Values: "1x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_nodes_count_per_nodepool{nodepool="default"})`, EvalTime: "2m", Samples: []Sample{{Value: 2}}},
			{Expr: `sum(runai_gpu_nodes_count_per_nodepool{nodepool="other"})`, EvalTime: "2m", Samples: []Sample{{Value: 1}}},
		},
	})
}

func TestReadyGpuNodesCountPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_gpu_count_per_node{node="n1", nodepool="default", modelName="A100", ready="true", schedulable="true"}`, Values: "2x3"},
			{Series: `runai_gpu_count_per_node{node="n2", nodepool="default", modelName="A100", ready="false", schedulable="true"}`, Values: "4x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_ready_gpu_nodes_count_per_nodepool{nodepool="default"})`, EvalTime: "2m", Samples: []Sample{{Value: 1}}},
		},
	})
}

func TestGpuNodesCountPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_gpu_nodes_count_per_nodepool{nodepool="np1"}`, Values: "2x3"},
			{Series: `runai_gpu_nodes_count_per_nodepool{nodepool="np2"}`, Values: "3x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_gpu_nodes_count_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 5}}},
		},
	})
}

func TestReadyGpuNodesCountPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_ready_gpu_nodes_count_per_nodepool{nodepool="np1"}`, Values: "1x3"},
			{Series: `runai_ready_gpu_nodes_count_per_nodepool{nodepool="np2"}`, Values: "2x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_ready_gpu_nodes_count_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 3}}},
		},
	})
}

// --- Free GPUs ---

func TestFreeGpuCountPerNode(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_allocated_gpu_count_per_gpu{clusterId="c1", node="n1", nodepool="np1", UUID="gpu-1", gpu="0"}`, Values: "1x3"},
			{Series: `runai_allocated_gpu_count_per_gpu{clusterId="c1", node="n1", nodepool="np1", UUID="gpu-2", gpu="1"}`, Values: "0x3"},
			{Series: `runai_allocated_gpu_count_per_gpu{clusterId="c1", node="n1", nodepool="np1", UUID="gpu-3", gpu="2"}`, Values: "1x3"},
			{Series: `runai_allocated_gpu_count_per_node{clusterId="c1", node="n1", nodepool="np1"}`, Values: "2x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_free_gpu_count_per_node{node="n1"})`, EvalTime: "2m", Samples: []Sample{{Value: 1}}},
		},
	})
}
