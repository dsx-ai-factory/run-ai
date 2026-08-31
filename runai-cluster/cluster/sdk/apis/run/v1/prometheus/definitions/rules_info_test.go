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

// --- Info rules ---

func TestNodeNodepoolExcluded(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_node_nodepool{node="n1", nodepool="default"}`, Values: "1x3"},
			{Series: `runai_node_nodepool{node="n2", nodepool="runai-excluded-nodes"}`, Values: "1x3"},
			{Series: `kube_node_info{node="n1"}`, Values: "1x3"},
			{Series: `kube_node_info{node="n2"}`, Values: "1x3"},
		},
		Tests: []Test{
			{Expr: `count(runai_node_nodepool_excluded)`, EvalTime: "2m", Samples: []Sample{{Value: 1}}},
			{Expr: `count(runai_node_nodepool_excluded{node="n1", nodepool="default"})`, EvalTime: "2m", Samples: []Sample{{Value: 1}}},
		},
	})
}

func TestRequestedMigDevicesPerWorkload(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_requested_mig_devices{queue_name="q1", workload_name="wl1", workload_type="train", mig_device_type="1g.5gb", project="p1"}`, Values: "2x3"},
			{Series: `runai_workload_info_enriched{workload_name="wl1", workload_id="wid-1", project="p1"}`, Values: "1x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_requested_mig_devices_per_workload{workload_id="wid-1"})`, EvalTime: "2m", Samples: []Sample{{Value: 2}}},
		},
	})
}

func TestWorkloadInfoUnique(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_workload_info{workload_id="wl-1", workload_name="job1", queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "1x3"},
			{Series: `runai_workload_info{workload_id="wl-1", workload_name="job1", queue_name="q1", project="p1", department="d1", nodepool="np2"}`, Values: "1x3"},
			{Series: `runai_workload_info{workload_id="wl-2", workload_name="job2", queue_name="q2", project="p2", department="d2", nodepool="np1"}`, Values: "1x3"},
		},
		Tests: []Test{
			{Expr: `count(runai_workload_info_unique)`, EvalTime: "2m", Samples: []Sample{{Value: 2}}},
			{Expr: `count(runai_workload_info_unique{workload_id="wl-1"})`, EvalTime: "2m", Samples: []Sample{{Value: 1}}},
			{Expr: `count(runai_workload_info_unique{workload_id="wl-2"})`, EvalTime: "2m", Samples: []Sample{{Value: 1}}},
		},
	})
}
