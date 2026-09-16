/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package basicflow

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/submit"
)

func TestParseSettings(t *testing.T) {
	cases := []struct {
		name    string
		knobs   map[string]any
		want    settings
		wantErr string
	}{
		{name: "defaults", knobs: nil, want: settings{projects: 1, workloads: 2, workloadType: "Training", timeout: 5 * time.Minute}},
		{name: "yaml knobs", knobs: map[string]any{"projects": 2, "workloads": 5, "workloadType": "Training", "timeout": "90s", "cluster": "kind"},
			want: settings{projects: 2, workloads: 5, workloadType: "Training", timeout: 90 * time.Second, cluster: "kind"}},
		{name: "unsupported workload type", knobs: map[string]any{"workloadType": "Workspace"}, wantErr: "not supported yet"},
		{name: "zero workloads", knobs: map[string]any{"workloads": 0}, wantErr: "at least 1"},
		{name: "wrong knob type", knobs: map[string]any{"workloads": "two"}, wantErr: "whole number"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseSettings(core.Config{Knobs: tc.knobs})
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("got %+v (%v), want %+v", got, err, tc.want)
			}
		})
	}
}

func accepted(names ...string) []acceptedWorkload {
	workloads := make([]acceptedWorkload, 0, len(names))
	for _, name := range names {
		workloads = append(workloads, acceptedWorkload{ref: submit.WorkloadRef{ID: name + "-id", Name: name}, namespace: "runai-p", acceptedAt: time.Now()})
	}
	return workloads
}

func TestMeasureNeverArrivedIsPendingNotZero(t *testing.T) {
	never := func(context.Context, acceptedWorkload) (bool, error) { return false, nil }

	result := measure(context.Background(), accepted("w1", "w2"), never, 5*time.Millisecond, 40*time.Millisecond)
	section := result.section(2)

	if len(result.syncMillis) != 0 {
		t.Fatalf("pending workloads must not get a duration: %v", result.syncMillis)
	}
	if len(result.pending) != 2 || !section.Failed || section.Metrics["pending"] != 2 {
		t.Fatalf("expected 2 pending and a failed section, got pending=%v failed=%v metrics=%v", result.pending, section.Failed, section.Metrics)
	}
	if _, present := section.Metrics["sync_to_cluster_max_ms"]; present {
		t.Fatal("max must be omitted when nothing arrived")
	}
	if !strings.Contains(section.Error, "2 of 2 workloads never reached the cluster") {
		t.Fatalf("unexpected error text: %s", section.Error)
	}
}

func TestMeasureRecordsArrivalPerWorkload(t *testing.T) {
	polls := map[string]int{}
	afterTwoPolls := func(_ context.Context, workload acceptedWorkload) (bool, error) {
		polls[workload.ref.Name]++
		return polls[workload.ref.Name] >= 2, nil
	}

	result := measure(context.Background(), accepted("w1", "w2"), afterTwoPolls, 5*time.Millisecond, time.Second)
	section := result.section(2)

	if len(result.pending) != 0 || section.Failed {
		t.Fatalf("nothing should be pending: %v (%s)", result.pending, section.Error)
	}
	for _, name := range []string{"w1", "w2"} {
		if result.syncMillis[name] <= 0 {
			t.Fatalf("workload %s must have a positive duration, got %d", name, result.syncMillis[name])
		}
	}
	maxMillis, ok := section.Metrics["sync_to_cluster_max_ms"].(int64)
	if !ok || maxMillis < result.syncMillis["w1"] || maxMillis < result.syncMillis["w2"] {
		t.Fatalf("max %v must be the largest duration of %v", section.Metrics["sync_to_cluster_max_ms"], result.syncMillis)
	}
}

func TestMeasureClusterErrorsEndAsPending(t *testing.T) {
	broken := func(context.Context, acceptedWorkload) (bool, error) {
		return false, errors.New("dial tcp: no route to host")
	}

	result := measure(context.Background(), accepted("w1"), broken, 5*time.Millisecond, 30*time.Millisecond)
	section := result.section(1)

	if len(result.pending) != 1 || !section.Failed {
		t.Fatalf("broken cluster must leave the workload pending: %v", result)
	}
	if !strings.Contains(section.Error, "no route to host") {
		t.Fatalf("last cluster error must be reported: %s", section.Error)
	}
}

func TestMeasureStopsWhenContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	never := func(context.Context, acceptedWorkload) (bool, error) { return false, nil }
	go func() {
		time.Sleep(15 * time.Millisecond)
		cancel()
	}()

	result := measure(ctx, accepted("w1", "w2"), never, 5*time.Millisecond, time.Minute)

	if len(result.pending) != 2 || !errors.Is(result.lastErr, context.Canceled) {
		t.Fatalf("cancelled run must report remaining workloads as pending: %v (%v)", result.pending, result.lastErr)
	}
}
