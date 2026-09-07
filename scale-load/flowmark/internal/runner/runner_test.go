/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
)

type fakeSuite struct {
	name          string
	fail          bool
	failedSection bool
	calls         int
	got           core.Config
}

func (f *fakeSuite) Name() string { return f.name }

func (f *fakeSuite) Run(ctx context.Context, env core.Env, cfg core.Config) (core.Section, error) {
	f.calls++
	f.got = cfg
	if f.fail {
		return core.Section{}, fmt.Errorf("boom")
	}
	if f.failedSection {
		return core.Section{Failed: true}, nil
	}
	return core.Section{Category: core.CategoryScale, Metrics: map[string]any{"ok": true}}, nil
}

func writeScenario(t *testing.T, dir, filename, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644); err != nil {
		t.Fatalf("writing scenario file: %v", err)
	}
}

func TestRunTwoScenarios(t *testing.T) {
	dir := t.TempDir()
	writeScenario(t, dir, "alpha.yaml", "suite: fake-a\n")
	writeScenario(t, dir, "beta.yaml", "suite: fake-b\n")
	a := &fakeSuite{name: "fake-a"}
	b := &fakeSuite{name: "fake-b"}
	registry := map[string]core.Scenario{"fake-a": a, "fake-b": b}

	results, err := Run(context.Background(), core.Env{}, dir, registry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(results))
	}
	if _, ok := results["alpha"]; !ok {
		t.Errorf("missing section keyed by filename 'alpha'")
	}
	if a.calls != 1 || b.calls != 1 {
		t.Errorf("expected each suite called once, got a=%d b=%d", a.calls, b.calls)
	}
}

func TestUnknownSuiteRunsNothing(t *testing.T) {
	dir := t.TempDir()
	writeScenario(t, dir, "good.yaml", "suite: fake-a\n")
	writeScenario(t, dir, "bad.yaml", "suite: no-such-suite\n")
	a := &fakeSuite{name: "fake-a"}
	registry := map[string]core.Scenario{"fake-a": a}

	_, err := Run(context.Background(), core.Env{}, dir, registry)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "bad") || !strings.Contains(err.Error(), "no-such-suite") {
		t.Errorf("error should name the file and the suite, got: %v", err)
	}
	if a.calls != 0 {
		t.Errorf("nothing should run when validation fails, but fake-a ran %d times", a.calls)
	}
}

func TestFailingSuiteDoesNotStopOthers(t *testing.T) {
	dir := t.TempDir()
	writeScenario(t, dir, "one.yaml", "suite: failing\n")
	writeScenario(t, dir, "two.yaml", "suite: healthy\n")
	failing := &fakeSuite{name: "failing", fail: true}
	healthy := &fakeSuite{name: "healthy"}
	registry := map[string]core.Scenario{"failing": failing, "healthy": healthy}

	results, err := Run(context.Background(), core.Env{}, dir, registry)
	if err == nil {
		t.Fatal("aggregate result must be an error when any scenario failed")
	}
	if healthy.calls != 1 {
		t.Errorf("healthy suite should still run, ran %d times", healthy.calls)
	}
	if !results["one"].Failed || results["one"].Error == "" {
		t.Errorf("failed scenario should be recorded with its error, got %+v", results["one"])
	}
	if results["two"].Failed {
		t.Errorf("healthy scenario wrongly marked failed")
	}
}

func TestEmptyAndMissingDir(t *testing.T) {
	if _, err := Run(context.Background(), core.Env{}, t.TempDir(), nil); err == nil {
		t.Error("empty dir should error")
	}
	if _, err := Run(context.Background(), core.Env{}, "/no/such/dir", nil); err == nil {
		t.Error("missing dir should error")
	}
}

func TestKnobsReachSuite(t *testing.T) {
	dir := t.TempDir()
	writeScenario(t, dir, "knobs.yaml", "suite: fake-a\nworkloads: 5\ntimeout: 5m\n")
	a := &fakeSuite{name: "fake-a"}
	registry := map[string]core.Scenario{"fake-a": a}

	if _, err := Run(context.Background(), core.Env{}, dir, registry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.got.Knobs["workloads"] != 5 {
		t.Errorf("knob 'workloads' should arrive as 5, got %v", a.got.Knobs["workloads"])
	}
	if a.got.Knobs["timeout"] != "5m" {
		t.Errorf("knob 'timeout' should arrive as \"5m\", got %v", a.got.Knobs["timeout"])
	}
	if _, exists := a.got.Knobs["suite"]; exists {
		t.Errorf("'suite' key must not leak into knobs")
	}
}

func TestCancelledContextStopsRun(t *testing.T) {
	dir := t.TempDir()
	writeScenario(t, dir, "one.yaml", "suite: fake-a\n")
	a := &fakeSuite{name: "fake-a"}
	registry := map[string]core.Scenario{"fake-a": a}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Run(ctx, core.Env{}, dir, registry)
	if err == nil {
		t.Fatal("cancelled context should abort the run")
	}
	if a.calls != 0 {
		t.Errorf("no scenario should start after cancellation, ran %d times", a.calls)
	}
}

func TestFailedSectionWithNilErrorFailsRun(t *testing.T) {
	dir := t.TempDir()
	writeScenario(t, dir, "sneaky.yaml", "suite: fake-a\n")
	a := &fakeSuite{name: "fake-a", failedSection: true}
	registry := map[string]core.Scenario{"fake-a": a}

	_, err := Run(context.Background(), core.Env{}, dir, registry)
	if err == nil {
		t.Fatal("a section marked Failed must fail the run, even with a nil error")
	}
	if !strings.Contains(err.Error(), "sneaky") {
		t.Errorf("error should name the scenario, got: %v", err)
	}
}
