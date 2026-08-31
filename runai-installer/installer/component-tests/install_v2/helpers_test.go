/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install_v2_test

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/stretchr/testify/require"
)

// Every component test in this package drives the CLI entrypoint
// (installconfig.Config -> ToSpec -> specrun.Execute). No test constructs the
// orchestrator Runner or a Plan directly; assertions are made against the
// harness fakes and the NDJSON events file produced by the run.

func newTestDeps(fakeHelm *harness.FakeHelmInstaller) orchestrator2.Deps {
	return newTestDepsWith(fakeHelm, &harness.FakeK8sExecutor{})
}

func newTestDepsWith(fakeHelm *harness.FakeHelmInstaller, fakeK8s *harness.FakeK8sExecutor) orchestrator2.Deps {
	_, provider := harness.NewFakeRunAIClientProvider(fakeHelm)
	return orchestrator2.Deps{
		K8sExecutor:         fakeK8s,
		HelmInstaller:       fakeHelm,
		RunAIClientProvider: provider,
		ClusterProbe:        okProbe(),
	}
}

// okProbe is a ClusterProbe that passes every preflight check, for tests that
// aren't exercising preflight failures.
func okProbe() *harness.FakeClusterProbe {
	return &harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true}
}

// readEvents parses an NDJSON events file into decoded maps.
func readEvents(t *testing.T, path string) []map[string]any {
	t.Helper()
	f, err := os.Open(path)
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

// findEvent returns the first event whose kind and step match, or nil.
func findEvent(events []map[string]any, kind, step string) map[string]any {
	for _, e := range events {
		if e["kind"] == kind && e["step"] == step {
			return e
		}
	}
	return nil
}

// eventIndex returns the position of the first event matching kind and step, or
// -1. Used to assert relative ordering between steps.
func eventIndex(events []map[string]any, kind, step string) int {
	for i, e := range events {
		if e["kind"] == kind && e["step"] == step {
			return i
		}
	}
	return -1
}

// findEventByKind returns the first event of the given kind, or nil.
func findEventByKind(events []map[string]any, kind string) map[string]any {
	for _, e := range events {
		if e["kind"] == kind {
			return e
		}
	}
	return nil
}

// prereqSummaryEvent returns the prereq_completed event's summary map, failing
// the test when the event is missing.
func prereqSummaryEvent(t *testing.T, evs []map[string]any) map[string]any {
	t.Helper()
	pc := findEventByKind(evs, "prereq_completed")
	require.NotNil(t, pc, "a prereq_completed event must be emitted")
	summary, ok := pc["prereq"].(map[string]any)
	require.True(t, ok, "prereq_completed must carry a summary")
	return summary
}

// featuresSummaryEvent returns the features_completed event's summary map,
// failing the test when the event is missing.
func featuresSummaryEvent(t *testing.T, evs []map[string]any) map[string]any {
	t.Helper()
	fc := findEventByKind(evs, "features_completed")
	require.NotNil(t, fc, "a features_completed event must be emitted")
	summary, ok := fc["features"].(map[string]any)
	require.True(t, ok, "features_completed must carry a summary")
	return summary
}

// pipelineEventIndex returns the position of the first event matching the given
// kind and pipeline name, or -1. Used to assert ordering across pipelines.
func pipelineEventIndex(events []map[string]any, kind, pipeline string) int {
	for i, e := range events {
		if e["kind"] == kind && e["pipeline"] == pipeline {
			return i
		}
	}
	return -1
}

// hasPipelineEvent reports whether any event matches the given kind and (when
// non-empty) pipeline name.
func hasPipelineEvent(events []map[string]any, kind, pipeline string) bool {
	for _, e := range events {
		if e["kind"] == kind && (pipeline == "" || e["pipeline"] == pipeline) {
			return true
		}
	}
	return false
}

// skipReasonCount counts step_skipped events carrying the given skipReason.
func skipReasonCount(events []map[string]any, reason string) int {
	n := 0
	for _, e := range events {
		if e["kind"] == "step_skipped" && e["skipReason"] == reason {
			n++
		}
	}
	return n
}
