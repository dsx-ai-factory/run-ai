/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package load_test drives the airgap load command end to end
// (loadconfig.Load -> ToSpec -> specrun.Execute). No test reaches into the
// internals; assertions are made against the run's error, the terminal output,
// the NDJSON events file, and the files on disk.
package load_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/loadconfig"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/plan"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/specrun"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/terminal"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/require"
)

func init() {
	pterm.DisableColor()
	pterm.DisableStyling()
}

const destination = "registry.local:5000/runai"

// result is everything one load run produced.
type result struct {
	err        error
	stdout     string
	events     []map[string]any
	workingDir string
	valuesDir  string
	pusher     *harness.FakeImagePusher
	resolved   *spec.Spec
}

// fakes are the engine seams one run executes against, for a test to break in
// one specific way.
type fakes struct {
	pusher *harness.FakeImagePusher
}

func newFakes() *fakes {
	return &fakes{pusher: &harness.FakeImagePusher{}}
}

// run executes the load command with the given config YAML.
func run(t *testing.T, configYAML string, dryRun bool) result {
	t.Helper()
	return runWith(t, configYAML, dryRun, newFakes())
}

// runFakes is run against fakes a test has configured to fail.
func runFakes(t *testing.T, configYAML string, f *fakes) result {
	t.Helper()
	return runWith(t, configYAML, false, f)
}

// runStaged is run with the preflight staging a test wants to exercise.
func runStaged(t *testing.T, configYAML string, f *fakes, adjust func(*specrun.ExecOptions)) result {
	t.Helper()
	return runWith(t, configYAML, false, f, adjust)
}

func runWith(t *testing.T, configYAML string, dryRun bool, f *fakes, adjust ...func(*specrun.ExecOptions)) result {
	t.Helper()

	res := runWithDeps(t, configYAML, dryRun, orchestrator2.Deps{
		ImagePusher: f.pusher,
	}, adjust...)
	res.pusher = f.pusher
	return res
}

// runWithDeps is run against the engine a test supplies, so the real one can be
// exercised too.
func runWithDeps(t *testing.T, configYAML string, dryRun bool, deps orchestrator2.Deps, adjust ...func(*specrun.ExecOptions)) result {
	t.Helper()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "airgap-load.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(configYAML), 0o600))
	eventsPath := filepath.Join(dir, "events.ndjson")

	cfg, err := loadconfig.Load(configPath)
	if err != nil {
		return result{err: err}
	}

	s, err := cfg.ToSpec(loadconfig.RunOptions{
		DryRun:     dryRun,
		Yes:        true,
		Output:     spec.OutputFormatText,
		EventsFile: eventsPath,
	})
	if err != nil {
		return result{err: err}
	}

	var out bytes.Buffer
	term := terminal.New(
		terminal.WithWriter(&out),
		terminal.WithErrWriter(&out),
		terminal.WithYes(true),
	)

	execOpts := specrun.ExecOptions{PlanBuilder: plan.Resolve}
	for _, a := range adjust {
		a(&execOpts)
	}
	runErr := specrun.Execute(context.Background(), s, deps, term, execOpts)

	return result{
		err:        runErr,
		stdout:     out.String(),
		events:     readEvents(t, eventsPath),
		workingDir: s.AirgapLoad.WorkingDir,
		valuesDir:  s.AirgapLoad.ValuesDir,
		resolved:   s,
	}
}

// loadConfig is a valid config for a bundle, for tests to break in one specific
// way.
func loadConfig(t *testing.T, bundlePath string) string {
	t.Helper()
	return fmt.Sprintf(`settings:
  bundlePath: %s
  workingDir: %s
registry:
  url: %s
  username: admin
  password: Harbor12345
`, bundlePath, filepath.Join(t.TempDir(), "airgap"), destination)
}

func readEvents(t *testing.T, path string) []map[string]any {
	t.Helper()
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })

	var out []map[string]any
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var got map[string]any
		require.NoError(t, json.Unmarshal(scanner.Bytes(), &got), "each line must be valid JSON")
		out = append(out, got)
	}
	require.NoError(t, scanner.Err())
	return out
}

func findEvent(events []map[string]any, kind, step string) map[string]any {
	for _, e := range events {
		if e["kind"] == kind && e["step"] == step {
			return e
		}
	}
	return nil
}

func findEventByKind(events []map[string]any, kind string) map[string]any {
	for _, e := range events {
		if e["kind"] == kind {
			return e
		}
	}
	return nil
}

// stepNames lists the steps that reached a given state, in order.
func stepNames(events []map[string]any, kind string) []string {
	var out []string
	for _, e := range events {
		if e["kind"] == kind {
			name, _ := e["step"].(string)
			out = append(out, name)
		}
	}
	return out
}

// findings returns the messages of the input-validation event, keyed by field.
func findings(t *testing.T, events []map[string]any) map[string]string {
	t.Helper()
	ev := findEventByKind(events, "input_validation_failed")
	require.NotNil(t, ev, "an input_validation_failed event must be emitted")

	out := map[string]string{}
	for _, raw := range ev["findings"].([]any) {
		f := raw.(map[string]any)
		out[f["field"].(string)], _ = f["message"].(string)
	}
	return out
}
