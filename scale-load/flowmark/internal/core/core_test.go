/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package core

import (
	"testing"
	"time"
)

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
