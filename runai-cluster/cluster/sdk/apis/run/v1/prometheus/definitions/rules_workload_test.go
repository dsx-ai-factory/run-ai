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

// --- Workload Count ---

func TestWorkloadCountPerTypePerPhasePerQueue(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_workload_info_enriched{workload_id="wl-1", workload_type="train", detailed_status="Running", is_gpu="true", queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "1x3"},
			{Series: `runai_workload_info_enriched{workload_id="wl-2", workload_type="train", detailed_status="Running", is_gpu="true", queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "1x3"},
			{Series: `runai_workload_info_enriched{workload_id="wl-3", workload_type="inference", detailed_status="Pending", is_gpu="false", queue_name="q1", project="p1", department="d1", nodepool="np1"}`, Values: "1x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_workload_count_per_type_per_phase_per_queue{queue_name="q1", workload_type="train", detailed_status="Running"})`, EvalTime: "2m", Samples: []Sample{{Value: 2}}},
		},
	})
}

func TestWorkloadCountPerCategory(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_workload_info_enriched{workload_id="wl-1", workload_category="training"}`, Values: "1x3"},
			{Series: `runai_workload_info_enriched{workload_id="wl-2", workload_category="training"}`, Values: "1x3"},
			{Series: `runai_workload_info_enriched{workload_id="wl-3", workload_category="inference"}`, Values: "1x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_workload_count_per_category{workload_category="training"})`, EvalTime: "2m", Samples: []Sample{{Value: 2}}},
			{Expr: `sum(runai_workload_count_per_category{workload_category="inference"})`, EvalTime: "2m", Samples: []Sample{{Value: 1}}},
		},
	})
}

func TestWorkloadCountPerProject(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_workload_info_enriched{workload_id="wl-1", project="p1", department="d1"}`, Values: "1x3"},
			{Series: `runai_workload_info_enriched{workload_id="wl-2", project="p1", department="d1"}`, Values: "1x3"},
			{Series: `runai_workload_info_enriched{workload_id="wl-3", project="p2", department="d1"}`, Values: "1x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_workload_count_per_project{project="p1"})`, EvalTime: "2m", Samples: []Sample{{Value: 2}}},
		},
	})
}

func TestWorkloadCountPerDepartment(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_workload_count_per_project{project="p1", department="d1"}`, Values: "3x3"},
			{Series: `runai_workload_count_per_project{project="p2", department="d1"}`, Values: "2x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_workload_count_per_department{department="d1"})`, EvalTime: "2m", Samples: []Sample{{Value: 5}}},
		},
	})
}

func TestWorkloadCountPerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_workload_info_enriched{workload_id="wl-1", nodepool="np1"}`, Values: "1x3"},
			{Series: `runai_workload_info_enriched{workload_id="wl-2", nodepool="np1"}`, Values: "1x3"},
			{Series: `runai_workload_info_enriched{workload_id="wl-3", nodepool="np2"}`, Values: "1x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_workload_count_per_nodepool{nodepool="np1"})`, EvalTime: "2m", Samples: []Sample{{Value: 2}}},
		},
	})
}

func TestWorkloadCountPerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_workload_count_per_nodepool{nodepool="np1"}`, Values: "3x3"},
			{Series: `runai_workload_count_per_nodepool{nodepool="np2"}`, Values: "2x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_workload_count_per_cluster)`, EvalTime: "2m", Samples: []Sample{{Value: 5}}},
		},
	})
}

func TestWorkloadCountPerTypePerPhasePerNodepool(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_workload_count_per_type_per_phase_per_queue{queue_name="q1", project="p1", department="d1", nodepool="np1", detailed_status="Running", is_gpu="true", workload_type="train"}`, Values: "2x3"},
			{Series: `runai_workload_count_per_type_per_phase_per_queue{queue_name="q2", project="p2", department="d1", nodepool="np1", detailed_status="Running", is_gpu="true", workload_type="train"}`, Values: "1x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_workload_count_per_type_per_phase_per_nodepool{nodepool="np1", workload_type="train"})`, EvalTime: "2m", Samples: []Sample{{Value: 3}}},
		},
	})
}

func TestWorkloadCountPerTypePerPhasePerCluster(t *testing.T) {
	newTester(t).Test(TestCase{
		Interval: "1m",
		Inputs: []Input{
			{Series: `runai_workload_count_per_type_per_phase_per_nodepool{nodepool="np1", detailed_status="Running", is_gpu="true", workload_type="train"}`, Values: "3x3"},
			{Series: `runai_workload_count_per_type_per_phase_per_nodepool{nodepool="np2", detailed_status="Running", is_gpu="true", workload_type="train"}`, Values: "2x3"},
		},
		Tests: []Test{
			{Expr: `sum(runai_workload_count_per_type_per_phase_per_cluster{workload_type="train"})`, EvalTime: "2m", Samples: []Sample{{Value: 5}}},
		},
	})
}
