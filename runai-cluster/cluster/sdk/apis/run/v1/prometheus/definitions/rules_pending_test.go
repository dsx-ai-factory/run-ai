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

// --- Pending Requested CPU/Memory ---

func TestPendingRequestedCpuCoresPerQueue(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_pending_requested_cpu_cores_per_workload{workload_id="wl-1", queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "4x3"},
			{Series: `runai_pending_requested_cpu_cores_per_workload{workload_id="wl-2", queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "2x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_pending_requested_cpu_cores_per_queue{queue_name="q1"})`, EvalTime: "2m", Samples: []Sample{{Value: 6}}},
		},
	})
}

func TestPendingRequestedCpuCoresPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_pending_requested_cpu_cores_per_queue{queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "6x3"},
			{Series: `runai_pending_requested_cpu_cores_per_queue{queue_name="q2", project="p2", department="d1", nodepool="np1"}`, Values: "4x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_pending_requested_cpu_cores_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 10}}},
		},
	})
}

func TestPendingRequestedCpuCoresPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_pending_requested_cpu_cores_per_nodepool{nodepool="np1"}`, Values: "10x3"},
			{Series: `runai_pending_requested_cpu_cores_per_nodepool{nodepool="np2"}`, Values: "5x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_pending_requested_cpu_cores_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 15}}},
		},
	})
}

func TestPendingRequestedMemoryBytesPerQueue(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_pending_requested_memory_bytes_per_workload{workload_id="wl-1", queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "4096x3"},
			{Series: `runai_pending_requested_memory_bytes_per_workload{workload_id="wl-2", queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "2048x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_pending_requested_memory_bytes_per_queue{queue_name="q1"})`, EvalTime: "2m", Samples: []Sample{{Value: 6144}}},
		},
	})
}

func TestPendingRequestedMemoryBytesPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_pending_requested_memory_bytes_per_queue{queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "6144x3"},
			{Series: `runai_pending_requested_memory_bytes_per_queue{queue_name="q2", project="p2", department="d1", nodepool="np1"}`, Values: "4096x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_pending_requested_memory_bytes_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 10240}}},
		},
	})
}

func TestPendingRequestedMemoryBytesPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_pending_requested_memory_bytes_per_nodepool{nodepool="np1"}`, Values: "10240x3"},
			{Series: `runai_pending_requested_memory_bytes_per_nodepool{nodepool="np2"}`, Values: "5120x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_pending_requested_memory_bytes_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 15360}}},
		},
	})
}
