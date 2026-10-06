/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package load

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
)

func TestCheckedInLoadScenariosParse(t *testing.T) {
	files, err := filepath.Glob("../../../scenarios/load/*.yaml")
	if err != nil || len(files) == 0 {
		t.Fatalf("no load scenarios found: %v", err)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var raw map[string]any
		if err := yaml.Unmarshal(data, &raw); err != nil {
			t.Fatalf("%s: %v", file, err)
		}
		if raw["suite"] != suiteName {
			t.Fatalf("%s names suite %v, the folder belongs to %s", file, raw["suite"], suiteName)
		}
		delete(raw, "suite")
		opts, err := parseOptions(core.Scenario{Suite: suiteName, Knobs: raw})
		if err != nil {
			t.Fatalf("%s must parse with the current knob rules: %v", file, err)
		}
		if len(opts.slo) != len(sloKeys) {
			t.Fatalf("%s should set every supported SLO (%d), got %d", file, len(sloKeys), len(opts.slo))
		}
	}
}
