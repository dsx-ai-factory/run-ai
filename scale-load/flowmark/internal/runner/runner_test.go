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
	category      core.Category
	fail          bool
	failedSection bool
	calls         int
	got           core.Scenario
}

func (f *fakeSuite) Name() string { return f.name }

func (f *fakeSuite) Category() core.Category {
	if f.category == "" {
		return core.CategoryScale
	}
	return f.category
}

func (f *fakeSuite) Run(ctx context.Context, target core.Target, scenario core.Scenario) (core.Section, error) {
	f.calls++
	f.got = scenario
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
	registry := map[string]core.Suite{"fake-a": a, "fake-b": b}

	results, err := Run(context.Background(), core.Target{}, dir, registry, Options{})
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
	registry := map[string]core.Suite{"fake-a": a}

	_, err := Run(context.Background(), core.Target{}, dir, registry, Options{})
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
	registry := map[string]core.Suite{"failing": failing, "healthy": healthy}

	results, err := Run(context.Background(), core.Target{}, dir, registry, Options{})
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
	if _, err := Run(context.Background(), core.Target{}, t.TempDir(), nil, Options{}); err == nil {
		t.Error("empty dir should error")
	}
	if _, err := Run(context.Background(), core.Target{}, "/no/such/dir", nil, Options{}); err == nil {
		t.Error("missing dir should error")
	}
}

func TestKnobsReachSuite(t *testing.T) {
	dir := t.TempDir()
	writeScenario(t, dir, "knobs.yaml", "suite: fake-a\nworkloads: 5\ntimeout: 5m\n")
	a := &fakeSuite{name: "fake-a"}
	registry := map[string]core.Suite{"fake-a": a}

	if _, err := Run(context.Background(), core.Target{}, dir, registry, Options{}); err != nil {
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
	registry := map[string]core.Suite{"fake-a": a}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Run(ctx, core.Target{}, dir, registry, Options{})
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
	registry := map[string]core.Suite{"fake-a": a}

	_, err := Run(context.Background(), core.Target{}, dir, registry, Options{})
	if err == nil {
		t.Fatal("a section marked Failed must fail the run, even with a nil error")
	}
	if !strings.Contains(err.Error(), "sneaky") {
		t.Errorf("error should name the scenario, got: %v", err)
	}
}

func TestRunOnlyKeepsTheNamedScenariosInTheGivenOrder(t *testing.T) {
	dir := t.TempDir()
	writeScenario(t, dir, "alpha.yaml", "suite: fake-a\n")
	writeScenario(t, dir, "beta.yaml", "suite: fake-b\n")
	writeScenario(t, dir, "gamma.yaml", "suite: fake-a\n")
	a := &fakeSuite{name: "fake-a"}
	b := &fakeSuite{name: "fake-b"}
	registry := map[string]core.Suite{"fake-a": a, "fake-b": b}

	results, err := Run(context.Background(), core.Target{}, dir, registry, Options{Only: []string{"beta"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected only beta to run, got %d sections", len(results))
	}
	if _, ok := results["beta"]; !ok {
		t.Errorf("missing section for beta, got %v", results)
	}
	if a.calls != 0 || b.calls != 1 {
		t.Errorf("expected only fake-b to be called once, got a=%d b=%d", a.calls, b.calls)
	}
}

func TestRunOnlyUnknownNameFailsBeforeAnythingRuns(t *testing.T) {
	dir := t.TempDir()
	writeScenario(t, dir, "alpha.yaml", "suite: fake-a\n")
	a := &fakeSuite{name: "fake-a"}
	registry := map[string]core.Suite{"fake-a": a}

	_, err := Run(context.Background(), core.Target{}, dir, registry, Options{Only: []string{"alpha", "nosuch"}})
	if err == nil {
		t.Fatal("expected an error for an unknown scenario name")
	}
	for _, want := range []string{`"nosuch"`, "available: alpha"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should contain %q", err, want)
		}
	}
	if a.calls != 0 {
		t.Errorf("nothing should run when a name is unknown, fake-a ran %d times", a.calls)
	}
}

func TestRunOverridesReplaceKnobsWithoutTouchingOthers(t *testing.T) {
	dir := t.TempDir()
	writeScenario(t, dir, "alpha.yaml", "suite: fake-a\nworkloads: [100, 250, 500]\nworkloadType: Training\n")
	a := &fakeSuite{name: "fake-a"}
	registry := map[string]core.Suite{"fake-a": a}

	overrides := map[string]any{"workloads": []any{100}}
	if _, err := Run(context.Background(), core.Target{}, dir, registry, Options{Overrides: overrides}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := a.got.Ints("workloads", nil)
	if err != nil || len(got) != 1 || got[0] != 100 {
		t.Errorf("workloads knob should be overridden to [100], got %v (err %v)", got, err)
	}
	if kind, _ := a.got.String("workloadType", ""); kind != "Training" {
		t.Errorf("untouched knob workloadType changed to %q", kind)
	}
}

func TestRunOnlyRejectsARepeatedName(t *testing.T) {
	dir := t.TempDir()
	writeScenario(t, dir, "alpha.yaml", "suite: fake-a\n")
	a := &fakeSuite{name: "fake-a"}

	_, err := Run(context.Background(), core.Target{}, dir, map[string]core.Suite{"fake-a": a}, Options{Only: []string{"alpha", "alpha"}})
	if err == nil || !strings.Contains(err.Error(), "listed twice") {
		t.Fatalf("a repeated name must be rejected, got %v", err)
	}
	if a.calls != 0 {
		t.Errorf("nothing may run when --only is invalid, fake-a ran %d times", a.calls)
	}
}

func TestRunOnlyIgnoresUnselectedFilesThatDoNotParse(t *testing.T) {
	dir := t.TempDir()
	writeScenario(t, dir, "alpha.yaml", "suite: fake-a\n")
	writeScenario(t, dir, "wip.yaml", "- not a mapping\n")
	a := &fakeSuite{name: "fake-a"}

	if _, err := Run(context.Background(), core.Target{}, dir, map[string]core.Suite{"fake-a": a}, Options{Only: []string{"alpha"}}); err != nil {
		t.Fatalf("a broken file that was not selected must not block the run: %v", err)
	}
	if _, err := Run(context.Background(), core.Target{}, dir, map[string]core.Suite{"fake-a": a}, Options{}); err == nil {
		t.Fatal("running everything must still report the broken file")
	}
}

func writeScenarioIn(t *testing.T, dir, folder, filename, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, folder), 0o755); err != nil {
		t.Fatal(err)
	}
	writeScenario(t, filepath.Join(dir, folder), filename, content)
}

func TestRunFindsScenariosInCategoryFolders(t *testing.T) {
	dir := t.TempDir()
	writeScenario(t, dir, "selfcheck.yaml", "suite: fake-a\n")
	writeScenarioIn(t, dir, "scale", "deployment-burst.yaml", "suite: fake-a\n")
	writeScenarioIn(t, dir, "load", "training-ramp.yaml", "suite: fake-load\n")
	writeScenarioIn(t, dir, "load", "README.md", "load scenarios arrive later\n")
	a := &fakeSuite{name: "fake-a"}
	ramp := &fakeSuite{name: "fake-load", category: core.CategoryLoad}

	results, err := Run(context.Background(), core.Target{}, dir, map[string]core.Suite{"fake-a": a, "fake-load": ramp}, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"selfcheck", "deployment-burst", "training-ramp"} {
		if _, ok := results[name]; !ok {
			t.Errorf("scenario %q from its folder must run, got %v", name, results)
		}
	}
	if a.calls != 2 || ramp.calls != 1 {
		t.Errorf("expected fake-a twice and fake-load once, got %d and %d", a.calls, ramp.calls)
	}
}

func TestRunRejectsAScenarioInTheWrongCategoryFolder(t *testing.T) {
	dir := t.TempDir()
	writeScenarioIn(t, dir, "load", "deployment-burst.yaml", "suite: fake-a\n")
	a := &fakeSuite{name: "fake-a"}

	_, err := Run(context.Background(), core.Target{}, dir, map[string]core.Suite{"fake-a": a}, Options{})
	if err == nil || !strings.Contains(err.Error(), `folder "load"`) || !strings.Contains(err.Error(), "scale suite") {
		t.Fatalf("a scale scenario under load/ must be rejected, got %v", err)
	}
	if a.calls != 0 {
		t.Error("nothing may run when a scenario sits in the wrong folder")
	}
}

func TestRunRejectsTheSameScenarioNameInTwoFolders(t *testing.T) {
	dir := t.TempDir()
	writeScenarioIn(t, dir, "scale", "burst.yaml", "suite: fake-a\n")
	writeScenarioIn(t, dir, "load", "burst.yaml", "suite: fake-a\n")

	_, err := Run(context.Background(), core.Target{}, dir, map[string]core.Suite{"fake-a": &fakeSuite{name: "fake-a"}}, Options{})
	if err == nil || !strings.Contains(err.Error(), "exists twice") {
		t.Fatalf("duplicate names across folders must be rejected, got %v", err)
	}
}

func TestFailedSuiteSectionStillCarriesTheSuiteCategory(t *testing.T) {
	dir := t.TempDir()
	writeScenario(t, dir, "one.yaml", "suite: failing\n")
	failing := &fakeSuite{name: "failing", fail: true, category: core.CategoryLoad}

	results, err := Run(context.Background(), core.Target{}, dir, map[string]core.Suite{"failing": failing}, Options{})
	if err == nil {
		t.Fatal("a failing suite must fail the run")
	}
	if got := results["one"].Category; got != core.CategoryLoad {
		t.Fatalf("a failed section must keep its suite's category for the dashboard, got %q", got)
	}
}

func TestRunRejectsScenariosNestedDeeperThanOneFolder(t *testing.T) {
	dir := t.TempDir()
	writeScenarioIn(t, dir, "scale", "burst.yaml", "suite: fake-a\n")
	writeScenarioIn(t, dir, filepath.Join("scale", "wip"), "deep.yaml", "suite: fake-a\n")
	a := &fakeSuite{name: "fake-a"}

	_, err := Run(context.Background(), core.Target{}, dir, map[string]core.Suite{"fake-a": a}, Options{})
	if err == nil || !strings.Contains(err.Error(), "must not contain sub-folders") {
		t.Fatalf("a nested folder must abort the run instead of being skipped, got %v", err)
	}
	if a.calls != 0 {
		t.Error("nothing may run when the layout is wrong")
	}
}
