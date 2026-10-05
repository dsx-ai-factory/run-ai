/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package report

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
)

func exampleDocument() Document {
	return Build(map[string]core.Section{
		"deployment-burst": {Category: core.CategoryScale, Metrics: map[string]any{
			"flow": "deployment_to_external_workload", "p90_ms@100": int64(5754), "never_arrived_count@100": 0, "slo_p90_ms": 5000,
		}},
		"selfcheck": {Category: core.CategoryScale, Metrics: map[string]any{"selfcheck_duration_seconds": 0.1}, Failed: true, Error: "example failure text"},
	}, RunInfo{
		StartedAt:           time.Date(2026, 9, 28, 1, 0, 0, 0, time.UTC),
		FinishedAt:          time.Date(2026, 9, 28, 1, 12, 30, 0, time.UTC),
		Provider:            "ci",
		ClusterType:         "kind-fake-gpu",
		ControlPlaneVersion: "3.1.0",
		ClusterVersion:      "2.26.5",
	})
}

func TestDocumentMatchesGoldenShape(t *testing.T) {
	encoded, err := exampleDocument().Encode()
	if err != nil {
		t.Fatal(err)
	}
	golden, err := os.ReadFile("testdata/results.golden.json")
	if err != nil {
		t.Fatal(err)
	}

	var got, want any
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatalf("encoded document is not valid json: %v", err)
	}
	if err := json.Unmarshal(golden, &want); err != nil {
		t.Fatalf("golden file is not valid json: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("document differs from testdata/results.golden.json:\n%s", encoded)
	}
	if encoded[len(encoded)-1] != '\n' {
		t.Fatal("encoded document must end with a newline")
	}
}

func TestDocumentOmitsUnknownVersionsAndKeepsRegressionField(t *testing.T) {
	doc := Build(map[string]core.Section{}, RunInfo{StartedAt: time.Date(2026, 9, 28, 1, 0, 0, 0, time.UTC), Provider: "local", ClusterType: "local"})
	encoded, err := doc.Encode()
	if err != nil {
		t.Fatal(err)
	}
	text := string(encoded)
	for _, absent := range []string{"control_plane_version", "cluster_version"} {
		if strings.Contains(text, absent) {
			t.Fatalf("%s must be omitted when unknown:\n%s", absent, text)
		}
	}
	if !strings.Contains(text, `"regression_detected": false`) || !strings.Contains(text, `"sections": {}`) {
		t.Fatalf("regression_detected and an empty sections object must always be present:\n%s", text)
	}
}

func TestFileNameIsSortableUTCAndUnique(t *testing.T) {
	a := Build(nil, RunInfo{StartedAt: time.Date(2026, 9, 28, 1, 0, 0, 0, time.UTC)})
	b := Build(nil, RunInfo{StartedAt: time.Date(2026, 9, 28, 1, 0, 1, 0, time.UTC)})
	if a.FileName() != "results-20260928T010000Z.json" {
		t.Fatalf("unexpected file name %q", a.FileName())
	}
	if a.FileName() >= b.FileName() {
		t.Fatalf("names must sort by time: %q vs %q", a.FileName(), b.FileName())
	}
	local := Build(nil, RunInfo{StartedAt: time.Date(2026, 9, 28, 4, 0, 0, 0, time.FixedZone("IDT", 3*3600))})
	if local.FileName() != "results-20260928T010000Z.json" {
		t.Fatalf("file name must use UTC, got %q", local.FileName())
	}
}
