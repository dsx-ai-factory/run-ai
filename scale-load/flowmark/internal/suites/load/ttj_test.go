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
			return o.users == 10 && o.rate == 2 && o.duration == 3*time.Minute && o.timeout == 10*time.Minute && o.workloadType == "Workspace" && o.slo == nil
		}},
		{name: "all knobs", knobs: map[string]any{"users": 3, "rate": 5, "duration": "1m", "image": "busybox", "timeout": "2m", "setupTimeout": "1m", "kubeRequestsPerSecond": 20, "controlPlaneRequestsPerSecond": 10, "cluster": "c1", "slo": map[string]any{"api_accepted_p90_ms": 500, "ttj_p90_ms": 10000, "cluster_synced_p99_ms": 12000}},
			check: func(o options) bool {
				return o.users == 3 && o.rate == 5 && o.duration == time.Minute && o.image == "busybox" && o.cluster == "c1" && o.slo["ttj_p90_ms"] == 10000 && o.slo["cluster_synced_p99_ms"] == 12000 && o.limits.ControlPlaneRequestsPerSecond == 10
			}},
		{name: "training workloads", knobs: map[string]any{"workloadType": "Training"}, check: func(o options) bool { return o.workloadType == "Training" }},
		{name: "spawn rate", knobs: map[string]any{"spawnRate": 5, "slo": map[string]any{"failed_count": 2}}, check: func(o options) bool { return o.spawnRate == 5 && o.slo["failed_count"] == 2 }},
		{name: "spawn rate above the users cap", knobs: map[string]any{"spawnRate": 1000}, wantErr: `"spawnRate" must be between 0 and 512`},
		{name: "unknown workload type", knobs: map[string]any{"workloadType": "StatefulSet"}, wantErr: `"workloadType" must be Workspace or Training`},
		{name: "misspelled knob", knobs: map[string]any{"rat": 2}, wantErr: `unknown knob "rat"`},
		{name: "rate above the control plane limit", knobs: map[string]any{"rate": 30, "controlPlaneRequestsPerSecond": 25}, wantErr: "cannot exceed controlPlaneRequestsPerSecond"},
		{name: "zero users", knobs: map[string]any{"users": 0}, wantErr: `"users" must be between 1 and 512`},
		{name: "submitted floor off", knobs: map[string]any{"minSubmittedPercent": 0}, check: func(o options) bool { return o.minSubmitted == 0 }},
		{name: "submitted floor default", knobs: map[string]any{}, check: func(o options) bool { return o.minSubmitted == 90 }},
		{name: "submitted floor above 100", knobs: map[string]any{"minSubmittedPercent": 101}, wantErr: `"minSubmittedPercent" must be between 0 and 100`},
		{name: "slo not a number", knobs: map[string]any{"slo": map[string]any{"ttj_p90_ms": "10s"}}, wantErr: "must be a whole number"},
		{name: "slo of another suite", knobs: map[string]any{"slo": map[string]any{"p90_ms": 10000}}, wantErr: `slo key "p90_ms" is not a metric this suite reports; use failed_count or <span>_<stat> with span api_accepted, cluster_synced, pod_created, ttj and stat p50_ms, p90_ms, p99_ms`},
		{name: "slo on a rate", knobs: map[string]any{"slo": map[string]any{"success_rate": 1}}, wantErr: `slo key "success_rate" is not a metric`},
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
	return options{users: 10, rate: 2, duration: time.Minute, workloadType: submit.WorkspaceType, image: defaultImage, timeout: 10 * time.Minute, minSubmitted: minSubmittedDefault, slo: map[string]int{"ttj_p90_ms": 10000, "api_accepted_p90_ms": 500}}
}

func TestBuildSectionReportsEverySpanAndTheRates(t *testing.T) {
	submissions := []measure.Submission{
		{Name: "w1", RequestedAt: at(0), ApiAcceptedAt: at(0.3), ClusterSyncedAt: at(2.3), PodCreatedAt: at(5.3)},
		{Name: "w2", RequestedAt: at(0.5), ApiAcceptedAt: at(0.9), ClusterSyncedAt: at(2.0), PodCreatedAt: at(4.0)},
		{Name: "w3", RequestedAt: at(1.0), Err: errors.New("503")},
	}
	pacing := submit.Pacing{Submitted: 3, FirstDispatch: at(0), LastDispatch: at(1.0)}

	section := buildSection(measure.CollectSpans(submissions), pacing, submissions, testOptions(), 0, nil, nil)

	if section.Failed {
		t.Fatalf("a rejected submission must not fail the section: %+v", section)
	}
	m := section.Metrics
	for _, key := range []string{"api_accepted_p90_ms", "cluster_synced_p90_ms", "pod_created_p90_ms", "ttj_p90_ms", "ttj_p50_ms", "ttj_p99_ms", "failed_count"} {
		if _, ok := m[key]; !ok {
			t.Errorf("metric %s is missing", key)
		}
	}
	if m["failed_count"] != 1 {
		t.Fatalf("the rejected submission must be counted: %v", m)
	}
	if _, present := m["never_arrived_count"]; present {
		t.Fatal("never_arrived_count is explained in the error text, not reported as a metric")
	}
	for _, key := range []string{"users", "duration_ms", "image", "slo", "success_rate", "flow", "submitted", "target_rate_per_sec", "actual_rate_per_sec"} {
		if _, present := m[key]; present {
			t.Fatalf("%s is a scenario knob or derivable, it must stay out of the report", key)
		}
	}
	if m["ttj_p90_ms"] != int64(5000) || section.SLO["ttj_p90_ms"] != 10000 {
		t.Fatalf("ttj p90 of [3100, 5000] is 5000 and its SLO sits in the slo block: %v / %v", m["ttj_p90_ms"], section.SLO)
	}
}

func TestBuildSectionFailsWhenThePacerFellBehind(t *testing.T) {
	arrived := []measure.Submission{
		{Name: "w1", RequestedAt: at(0), ApiAcceptedAt: at(0.3), ClusterSyncedAt: at(2.3), PodCreatedAt: at(5.3)},
		{Name: "w2", RequestedAt: at(30), ApiAcceptedAt: at(30.3), ClusterSyncedAt: at(32.3), PodCreatedAt: at(35.3)},
	}
	kept := submit.Pacing{Planned: 120, Submitted: 121, FirstDispatch: at(0), LastDispatch: at(60)}
	if section := buildSection(measure.CollectSpans(arrived), kept, arrived, testOptions(), 0, nil, nil); section.Failed {
		t.Fatalf("2.00/s against a 2/s target is on pace: %+v", section)
	}

	behind := submit.Pacing{Planned: 120, Submitted: 61, FirstDispatch: at(0), LastDispatch: at(60)}
	section := buildSection(measure.CollectSpans(arrived), behind, arrived, testOptions(), 0, nil, nil)

	if !section.Failed || !strings.Contains(section.Error, "kept only 61 of the 120 planned submissions") {
		t.Fatalf("a pacer more than 10%% under its target must fail the section and say so: %+v", section)
	}

	tolerant := testOptions()
	tolerant.minSubmitted = 0
	section = buildSection(measure.CollectSpans(arrived), behind, arrived, tolerant, 0, nil, nil)
	if section.Failed {
		t.Fatalf("minSubmittedPercent 0 turns the shortfall into a measurement, not a failure: %+v", section)
	}
	if section.Metrics["planned_count"] != 120 || section.Metrics["submitted_count"] != 61 {
		t.Fatalf("the shortfall must stay visible in the report when it is not a failure: %v", section.Metrics)
	}
}

func TestBuildSectionFailsOnNeverArrivedWorkloadsAndTimestampsOutOfOrder(t *testing.T) {
	neverArrived := []measure.Submission{{Name: "w1", RequestedAt: at(0), ApiAcceptedAt: at(0.3), ClusterSyncedAt: at(2.3)}}
	section := buildSection(measure.CollectSpans(neverArrived), submit.Pacing{Submitted: 1}, neverArrived, testOptions(), 0, nil, nil)
	if !section.Failed || !strings.Contains(section.Error, "never reached the cluster") {
		t.Fatalf("a workload that never arrived must fail the section: %+v", section)
	}
	if _, ok := section.Metrics["ttj_p90_ms"]; ok {
		t.Fatal("no ttj latency may be reported when nothing arrived")
	}
	section = buildSection(measure.CollectSpans(neverArrived), submit.Pacing{Submitted: 1}, neverArrived, testOptions(), 0, context.Canceled, nil)
	if !section.Failed || !strings.Contains(section.Error, "the run was interrupted after 1 submissions") || !strings.Contains(section.Error, "1 workloads had not reached the cluster") || strings.Contains(section.Error, "within") {
		t.Fatalf("an interrupted run must say so instead of blaming the timeout: %+v", section)
	}
	complete := []measure.Submission{{Name: "w1", RequestedAt: at(0), ApiAcceptedAt: at(0.3), ClusterSyncedAt: at(2.3), PodCreatedAt: at(5.3)}}
	if section := buildSection(measure.CollectSpans(complete), submit.Pacing{Submitted: 1}, complete, testOptions(), 0, context.Canceled, nil); !section.Failed || !strings.Contains(section.Error, "interrupted") {
		t.Fatalf("an interrupted run fails even when everything submitted so far arrived: %+v", section)
	}

	podBeforeAccept := []measure.Submission{{Name: "w1", RequestedAt: at(0), ApiAcceptedAt: at(6.0), ClusterSyncedAt: at(2.0), PodCreatedAt: at(2.5)}}
	section = buildSection(measure.CollectSpans(podBeforeAccept), submit.Pacing{Submitted: 1}, podBeforeAccept, testOptions(), 0, nil, nil)
	if !section.Failed || !strings.Contains(section.Error, "timestamps out of order") {
		t.Fatalf("timestamps out of order must fail the section loudly: %+v", section)
	}
	for key := range section.Metrics {
		if strings.HasPrefix(key, "measurement_") {
			t.Fatalf("measurement health belongs in the log and the error text, not in the metrics: %s", key)
		}
	}
	interruptedSection := buildSection(measure.CollectSpans(neverArrived), submit.Pacing{Submitted: 1}, neverArrived, testOptions(), 2, nil, nil)
	if !strings.Contains(interruptedSection.Error, "interrupted 2 times") {
		t.Fatalf("watch interruptions must be explained in the error text: %+v", interruptedSection)
	}
}

func TestBuildSectionBlamesADeadWatchNotTheCluster(t *testing.T) {
	unresolved := []measure.Submission{{Name: "w1", RequestedAt: at(0), ApiAcceptedAt: at(0.3)}}
	section := buildSection(measure.CollectSpans(unresolved), submit.Pacing{Submitted: 1}, unresolved, testOptions(), 0, nil, measure.ErrWatchClosed)

	if !section.Failed || !strings.Contains(section.Error, "the cluster watch died, 1 workloads could not be followed") || strings.Contains(section.Error, "never reached the cluster") {
		t.Fatalf("a dead watch is a measurement failure, not a slow cluster: %+v", section)
	}
}
