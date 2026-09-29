/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestRunOptions(t *testing.T) {
	cases := []struct {
		name      string
		only      string
		workloads string
		wantOnly  []string
		wantSizes []any
		wantErr   string
	}{
		{name: "defaults run everything unchanged"},
		{name: "one scenario", only: "training-burst", wantOnly: []string{"training-burst"}},
		{name: "several scenarios with spaces", only: " selfcheck, training-burst ", wantOnly: []string{"selfcheck", "training-burst"}},
		{name: "one round", workloads: "100", wantSizes: []any{100}},
		{name: "two rounds", workloads: "100,250", wantSizes: []any{100, 250}},
		{name: "not a number", workloads: "100,many", wantErr: `"many" is not a positive whole number`},
		{name: "zero", workloads: "0", wantErr: "not a positive whole number"},
		{name: "only commas", workloads: ",", wantErr: "expected one or more whole numbers"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := runOptions(tc.only, tc.workloads)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got.Only, tc.wantOnly) {
				t.Errorf("Only = %v, want %v", got.Only, tc.wantOnly)
			}
			if tc.wantSizes == nil {
				if got.Overrides != nil {
					t.Errorf("expected no overrides, got %v", got.Overrides)
				}
				return
			}
			if !reflect.DeepEqual(got.Overrides["workloads"], tc.wantSizes) {
				t.Errorf("workloads override = %v, want %v", got.Overrides["workloads"], tc.wantSizes)
			}
		})
	}
}
