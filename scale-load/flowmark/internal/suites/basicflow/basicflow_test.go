/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package basicflow

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/measure"
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

func timing(name string, outcome measure.Outcome, d time.Duration, err error) measure.Timing {
	created := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	t := measure.Timing{Name: name, Outcome: outcome, Created: created, Err: err}
	if outcome == measure.Arrived {
		t.Arrived = created.Add(d)
		t.Duration = d
	}
	return t
}

func TestBuildSectionAllArrived(t *testing.T) {
	section := buildSection([]measure.Timing{
		timing("w1", measure.Arrived, 10*time.Second, nil),
		timing("w2", measure.Arrived, 12*time.Second, nil),
	}, 2)

	if section.Failed || section.Error != "" {
		t.Fatalf("all arrived must not fail: %+v", section)
	}
	sync := section.Metrics["sync_to_cluster_ms"].(map[string]int64)
	if sync["w1"] != 10000 || sync["w2"] != 12000 {
		t.Fatalf("per-workload durations wrong: %v", sync)
	}
	if section.Metrics["p50_ms"] != int64(10000) || section.Metrics["max_ms"] != int64(12000) || section.Metrics["pending_count"] != 0 {
		t.Fatalf("stats wrong: %v", section.Metrics)
	}
}

func TestBuildSectionPendingIsNeverZero(t *testing.T) {
	section := buildSection([]measure.Timing{
		timing("w1", measure.Pending, 0, nil),
		timing("w2", measure.Pending, 0, nil),
	}, 2)

	if !section.Failed || !strings.Contains(section.Error, "2 of 2 workloads never reached the cluster") {
		t.Fatalf("pending workloads must fail the section: %+v", section)
	}
	if len(section.Metrics["sync_to_cluster_ms"].(map[string]int64)) != 0 {
		t.Fatal("pending workloads must not get a duration")
	}
	for _, key := range []string{"p50_ms", "max_ms", "throughput_per_s"} {
		if _, present := section.Metrics[key]; present {
			t.Fatalf("%s must be omitted when nothing arrived", key)
		}
	}
	if section.Metrics["pending_count"] != 2 {
		t.Fatalf("pending_count must be 2: %v", section.Metrics)
	}
}

func TestBuildSectionReportsWatchAndSubmissionErrors(t *testing.T) {
	section := buildSection([]measure.Timing{
		timing("w1", measure.Arrived, 9*time.Second, nil),
		timing("w2", measure.Pending, 0, errors.New("watch closed by the server")),
		timing("w3", measure.Failed, 0, errors.New("HTTP 500")),
	}, 3)

	if !section.Failed {
		t.Fatal("pending and failed items must fail the section")
	}
	for _, want := range []string{"1 of 3 workloads never reached the cluster", "watch closed by the server", "1 of 3 workload submissions failed", "w3: HTTP 500"} {
		if !strings.Contains(section.Error, want) {
			t.Fatalf("error text must contain %q, got %q", want, section.Error)
		}
	}
	if section.Metrics["sample_size"] != 1 || section.Metrics["failed_count"] != 1 {
		t.Fatalf("counts wrong: %v", section.Metrics)
	}
}

func TestRunIDIsShortUniqueHex(t *testing.T) {
	first, err := runID()
	if err != nil {
		t.Fatal(err)
	}
	second, err := runID()
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 8 || strings.Trim(first, "0123456789abcdef") != "" {
		t.Fatalf("run id must be 8 lowercase hex characters, got %q", first)
	}
	if first == second {
		t.Fatalf("two run ids must differ, both were %q", first)
	}
}
