/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package core

import (
	"slices"
	"strings"
	"testing"
	"time"
)

func TestConfigInt(t *testing.T) {
	cases := []struct {
		name    string
		knobs   map[string]any
		want    int
		wantErr bool
	}{
		{name: "missing knob uses fallback", knobs: nil, want: 2},
		{name: "explicit null uses fallback", knobs: map[string]any{"workloads": nil}, want: 2},
		{name: "yaml integer", knobs: map[string]any{"workloads": 5}, want: 5},
		{name: "string rejected", knobs: map[string]any{"workloads": "5"}, wantErr: true},
		{name: "float rejected", knobs: map[string]any{"workloads": 2.5}, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Config{Knobs: tc.knobs}.Int("workloads", 2)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %d", got)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("got %d (%v), want %d", got, err, tc.want)
			}
		})
	}
}

func TestConfigString(t *testing.T) {
	cases := []struct {
		name    string
		knobs   map[string]any
		want    string
		wantErr bool
	}{
		{name: "missing knob uses fallback", knobs: nil, want: "Training"},
		{name: "explicit null uses fallback", knobs: map[string]any{"workloadType": nil}, want: "Training"},
		{name: "string", knobs: map[string]any{"workloadType": "Workspace"}, want: "Workspace"},
		{name: "number rejected", knobs: map[string]any{"workloadType": 3}, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Config{Knobs: tc.knobs}.String("workloadType", "Training")
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("got %q (%v), want %q", got, err, tc.want)
			}
		})
	}
}

func TestConfigDuration(t *testing.T) {
	cases := []struct {
		name    string
		knobs   map[string]any
		want    time.Duration
		wantErr bool
	}{
		{name: "missing knob uses fallback", knobs: nil, want: 2 * time.Minute},
		{name: "duration string parsed", knobs: map[string]any{"timeout": "90s"}, want: 90 * time.Second},
		{name: "explicit null uses fallback", knobs: map[string]any{"timeout": nil}, want: 2 * time.Minute},
		{name: "non-string rejected", knobs: map[string]any{"timeout": 90}, wantErr: true},
		{name: "zero rejected", knobs: map[string]any{"timeout": "0s"}, wantErr: true},
		{name: "negative rejected", knobs: map[string]any{"timeout": "-5m"}, wantErr: true},
		{name: "unparsable rejected", knobs: map[string]any{"timeout": "soon"}, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Config{Knobs: tc.knobs}.Duration("timeout", 2*time.Minute)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %s", got)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("got %s (%v), want %s", got, err, tc.want)
			}
		})
	}
}

func TestConfigInts(t *testing.T) {
	cases := []struct {
		name    string
		knobs   map[string]any
		want    []int
		wantErr string
	}{
		{name: "absent uses fallback", knobs: nil, want: []int{1, 2}},
		{name: "yaml list", knobs: map[string]any{"workloads": []any{100, 250, 500}}, want: []int{100, 250, 500}},
		{name: "not a list", knobs: map[string]any{"workloads": 100}, wantErr: "must be a list"},
		{name: "item not a number", knobs: map[string]any{"workloads": []any{100, "many"}}, wantErr: "item 2 must be a whole number"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := (Config{Knobs: tc.knobs}).Ints("workloads", []int{1, 2})
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil || !slices.Equal(got, tc.want) {
				t.Fatalf("got %v (%v), want %v", got, err, tc.want)
			}
		})
	}
}

func TestConfigMap(t *testing.T) {
	got, err := (Config{Knobs: map[string]any{"slo": map[string]any{"create_to_ew_p90_ms": 5000}}}).Map("slo")
	if err != nil || got["create_to_ew_p90_ms"] != 5000 {
		t.Fatalf("got %v (%v)", got, err)
	}
	if got, err := (Config{}).Map("slo"); err != nil || got != nil {
		t.Fatalf("absent map must be nil without error, got %v (%v)", got, err)
	}
	if _, err := (Config{Knobs: map[string]any{"slo": "fast"}}).Map("slo"); err == nil || !strings.Contains(err.Error(), "must be a mapping") {
		t.Fatalf("expected mapping error, got %v", err)
	}
}

func TestRunIDIsShortUniqueHex(t *testing.T) {
	first, err := RunID()
	if err != nil {
		t.Fatal(err)
	}
	second, err := RunID()
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

func TestConfigUnknown(t *testing.T) {
	cfg := Config{Knobs: map[string]any{"workloads": []any{1}, "workloadTipe": "Training", "slos": nil}}

	got := cfg.Unknown("workloads", "workloadType", "slo")

	if !slices.Equal(got, []string{"slos", "workloadTipe"}) {
		t.Fatalf("unknown knobs must be reported sorted, got %v", got)
	}
	if got := (Config{}).Unknown("workloads"); len(got) != 0 {
		t.Fatalf("no knobs means nothing unknown, got %v", got)
	}
}
