/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package report

import (
	"slices"
	"testing"
)

func TestManifestAppend(t *testing.T) {
	cases := []struct {
		name  string
		start []string
		add   string
		want  []string
	}{
		{name: "empty manifest", start: nil, add: "results-1.json", want: []string{"results-1.json"}},
		{name: "existing entries keep their order", start: []string{"results-1.json", "results-2.json"}, add: "results-3.json", want: []string{"results-1.json", "results-2.json", "results-3.json"}},
		{name: "re-run of the same file is a no-op", start: []string{"results-1.json", "results-2.json"}, add: "results-2.json", want: []string{"results-1.json", "results-2.json"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			start := slices.Clone(tc.start)
			got := Manifest{Runs: tc.start}.with(tc.add)
			if !slices.Equal(got.Runs, tc.want) {
				t.Fatalf("got %v want %v", got.Runs, tc.want)
			}
			if !slices.Equal(tc.start, start) {
				t.Fatal("with must not mutate the manifest it was called on")
			}
		})
	}
}

func TestManifestRoundTrip(t *testing.T) {
	encoded, err := Manifest{Runs: []string{"results-1.json"}}.encode()
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != "{\n  \"runs\": [\n    \"results-1.json\"\n  ]\n}\n" {
		t.Fatalf("unexpected manifest layout:\n%s", encoded)
	}
	parsed, err := parseManifest(encoded)
	if err != nil || !parsed.contains("results-1.json") {
		t.Fatalf("round trip failed: %v %v", parsed, err)
	}
	if _, err := parseManifest([]byte("not json")); err == nil {
		t.Fatal("a corrupt manifest must be an error, not an empty manifest")
	}
}
