/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package load

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/measure"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/submit"
)

func TestParseOptions(t *testing.T) {
	cases := []struct {
		name    string
		knobs   map[string]any
		check   func(options) bool
		wantErr string
	}{
		{name: "defaults", knobs: nil, check: func(o options) bool {
			return o.users == 10 && o.rate == 2 && o.duration == 5*time.Minute && o.timeout == 10*time.Minute && o.workloadType == "Workspace" && o.slo == nil
		}},
		{name: "all knobs", knobs: map[string]any{"users": 3, "rate": 5, "duration": "1m", "image": "busybox", "timeout": "2m", "setupTimeout": "1m", "kubeRequestsPerSecond": 20, "controlPlaneRequestsPerSecond": 10, "cluster": "c1", "slo": map[string]any{"api_accepted_p90_ms": 500, "ttj_p90_ms": 10000}},
			check: func(o options) bool {
				return o.users == 3 && o.rate == 5 && o.duration == time.Minute && o.image == "busybox" && o.cluster == "c1" && o.slo["ttj_p90_ms"] == 10000 && o.limits.ControlPlaneRequestsPerSecond == 10
			}},
		{name: "training workloads", knobs: map[string]any{"workloadType": "Training"}, check: func(o options) bool { return o.workloadType == "Training" }},
		{name: "unknown workload type", knobs: map[string]any{"workloadType": "Deployment"}, wantErr: `"workloadType" must be Workspace or Training`},
		{name: "misspelled knob", knobs: map[string]any{"rat": 2}, wantErr: `unknown knob "rat"`},
		{name: "rate above the control plane limit", knobs: map[string]any{"rate": 30, "controlPlaneRequestsPerSecond": 25}, wantErr: "cannot exceed controlPlaneRequestsPerSecond"},
		{name: "zero users", knobs: map[string]any{"users": 0}, wantErr: `"users" must be between 1 and 64`},
		{name: "slo not a number", knobs: map[string]any{"slo": map[string]any{"ttj_p90_ms": "10s"}}, wantErr: "must be a whole number"},
		{name: "duration not a duration", knobs: map[string]any{"duration": 5}, wantErr: "must be a duration string"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseOptions(core.Scenario{Suite: suiteName, Knobs: tc.knobs})
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.check(got) {
				t.Fatalf("unexpected options: %+v", got)
			}
		})
	}
}

var start = time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)

func at(seconds float64) time.Time {
	return start.Add(time.Duration(seconds * float64(time.Second)))
}

func testOptions() options {
	return options{users: 10, rate: 2, duration: time.Minute, workloadType: workspaceType, image: defaultImage, timeout: 10 * time.Minute, slo: map[string]int{"ttj_p90_ms": 10000, "api_accepted_p90_ms": 500}}
}

func TestBuildSectionReportsEverySpanAndTheRates(t *testing.T) {
	submissions := []measure.Submission{
		{Name: "w1", RequestedAt: at(0), ApiAcceptedAt: at(0.3), ClusterSyncedAt: at(2.3), PodCreatedAt: at(5.3)},
		{Name: "w2", RequestedAt: at(0.5), ApiAcceptedAt: at(0.9), ClusterSyncedAt: at(2.0), PodCreatedAt: at(4.0)},
		{Name: "w3", RequestedAt: at(1.0), Err: errors.New("503")},
	}
	pacing := submit.Pacing{Submitted: 3, FirstDispatch: at(0), LastDispatch: at(1.0)}

	section := buildSection(measure.CollectSpans(submissions), pacing, submissions, testOptions(), 0, nil)

	if section.Category != core.CategoryLoad || section.Failed {
		t.Fatalf("a rejected submission must not fail the section: %+v", section)
	}
	m := section.Metrics
	for _, key := range []string{"api_accepted_p90_ms", "cluster_synced_p90_ms", "pod_created_p90_ms", "ttj_p90_ms", "ttj_p50_ms", "ttj_p99_ms", "actual_rate_per_sec", "target_rate_per_sec", "success_rate", "failed_count", "never_arrived_count", "slo_ttj_p90_ms"} {
		if _, ok := m[key]; !ok {
			t.Errorf("metric %s is missing", key)
		}
	}
	if m["submitted"] != 3 || m["failed_count"] != 1 || m["never_arrived_count"] != 0 {
		t.Fatalf("counts wrong: %v", m)
	}
	if rate, _ := m["success_rate"].(float64); rate < 0.66 || rate > 0.67 {
		t.Fatalf("2 of 3 accepted is 0.67, got %v", m["success_rate"])
	}
	if m["ttj_p90_ms"] != int64(5000) || m["slo_ttj_p90_ms"] != 10000 {
		t.Fatalf("ttj p90 of [3100, 5000] is 5000 next to its SLO: %v / %v", m["ttj_p90_ms"], m["slo_ttj_p90_ms"])
	}
}

func TestBuildSectionFailsOnNeverArrivedWorkloadsAndTimestampsOutOfOrder(t *testing.T) {
	neverArrived := []measure.Submission{{Name: "w1", RequestedAt: at(0), ApiAcceptedAt: at(0.3), ClusterSyncedAt: at(2.3)}}
	section := buildSection(measure.CollectSpans(neverArrived), submit.Pacing{Submitted: 1}, neverArrived, testOptions(), 0, nil)
	if !section.Failed || !strings.Contains(section.Error, "never reached the cluster") {
		t.Fatalf("a workload that never arrived must fail the section: %+v", section)
	}
	if _, ok := section.Metrics["ttj_p90_ms"]; ok {
		t.Fatal("no ttj latency may be reported when nothing arrived")
	}
	section = buildSection(measure.CollectSpans(neverArrived), submit.Pacing{Submitted: 1}, neverArrived, testOptions(), 0, context.Canceled)
	if !section.Failed || !strings.Contains(section.Error, "when the run was interrupted") || strings.Contains(section.Error, "within") {
		t.Fatalf("an interrupted run must say so instead of blaming the timeout: %+v", section)
	}

	podBeforeAccept := []measure.Submission{{Name: "w1", RequestedAt: at(0), ApiAcceptedAt: at(6.0), ClusterSyncedAt: at(2.0), PodCreatedAt: at(2.5)}}
	section = buildSection(measure.CollectSpans(podBeforeAccept), submit.Pacing{Submitted: 1}, podBeforeAccept, testOptions(), 0, nil)
	if !section.Failed || !strings.Contains(section.Error, "timestamps out of order") {
		t.Fatalf("timestamps out of order must fail the section loudly: %+v", section)
	}
	for key := range section.Metrics {
		if strings.HasPrefix(key, "measurement_") {
			t.Fatalf("measurement health belongs in the log and the error text, not in the metrics: %s", key)
		}
	}
	interruptedSection := buildSection(measure.CollectSpans(neverArrived), submit.Pacing{Submitted: 1}, neverArrived, testOptions(), 2, nil)
	if !strings.Contains(interruptedSection.Error, "interrupted 2 times") {
		t.Fatalf("watch interruptions must be explained in the error text: %+v", interruptedSection)
	}
}
