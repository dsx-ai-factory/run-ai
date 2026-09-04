/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package events

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/terminal"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	pterm.DisableColor()
	pterm.DisableStyling()
}

type stubSink struct {
	events []orchestrator2.Event
	err    error
}

func (s *stubSink) Handle(e orchestrator2.Event) error {
	s.events = append(s.events, e)
	return s.err
}

func (s *stubSink) Close() error { return nil }

func TestManager_FanOut_AllSinksReceiveEvent(t *testing.T) {
	a := &stubSink{}
	b := &stubSink{}
	m := &Manager{sinks: []sink{a, b}}

	require.NoError(t, m.HandleEvent(orchestrator2.Event{Kind: orchestrator2.EventKindPipelineStarted, Pipeline: "p"}))

	require.Len(t, a.events, 1)
	require.Len(t, b.events, 1)
}

func TestManager_HandleEvent_StopsOnFirstError(t *testing.T) {
	boom := errors.New("boom")
	a := &stubSink{err: boom}
	b := &stubSink{}
	m := &Manager{sinks: []sink{a, b}}

	err := m.HandleEvent(orchestrator2.Event{Kind: orchestrator2.EventKindPipelineStarted, Pipeline: "p"})

	require.ErrorIs(t, err, boom)
	assert.Len(t, a.events, 1)
	assert.Empty(t, b.events, "subsequent sinks must not receive the event")
}

func TestManager_GetEvents_ReturnsAllEmitted(t *testing.T) {
	m := &Manager{}
	require.NoError(t, m.HandleEvent(orchestrator2.Event{Kind: orchestrator2.EventKindPipelineStarted, Pipeline: "p"}))
	require.NoError(t, m.HandleEvent(orchestrator2.Event{Kind: orchestrator2.EventKindPipelineSucceeded, Pipeline: "p"}))

	got := m.GetEvents()
	require.Len(t, got, 2)
	assert.Equal(t, orchestrator2.EventKindPipelineStarted, got[0].Kind)
	assert.Equal(t, orchestrator2.EventKindPipelineSucceeded, got[1].Kind)
}

func TestManager_GetEvents_RecordsEvenWhenSinkErrors(t *testing.T) {
	boom := errors.New("boom")
	m := &Manager{sinks: []sink{&stubSink{err: boom}}}

	err := m.HandleEvent(orchestrator2.Event{Kind: orchestrator2.EventKindStepFailed, Step: "s"})
	require.ErrorIs(t, err, boom)

	got := m.GetEvents()
	require.Len(t, got, 1, "event should be recorded before fan-out, even if a sink errors")
	assert.Equal(t, orchestrator2.EventKindStepFailed, got[0].Kind)
}

func TestManager_GetEvents_ReturnsCopy(t *testing.T) {
	m := &Manager{}
	require.NoError(t, m.HandleEvent(orchestrator2.Event{Kind: orchestrator2.EventKindPipelineStarted, Pipeline: "p"}))

	got := m.GetEvents()
	got[0].Pipeline = "mutated"

	assert.Equal(t, "p", m.GetEvents()[0].Pipeline, "internal state must not be mutated through GetEvents return value")
}

func TestNDJSONSink_WritesOneLinePerEvent(t *testing.T) {
	var buf bytes.Buffer
	s := newNDJSONSink(&buf, nil)

	require.NoError(t, s.Handle(orchestrator2.Event{Kind: orchestrator2.EventKindPipelineStarted, Pipeline: "p"}))
	require.NoError(t, s.Handle(orchestrator2.Event{
		Kind:       orchestrator2.EventKindStepSkipped,
		Pipeline:   "p",
		Flow:       "f",
		Step:       "s",
		SkipReason: orchestrator2.SkipReasonDryRun,
		Output:     "helm.install: chart",
	}))

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	require.Len(t, lines, 2)

	var first map[string]any
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &first))
	assert.Equal(t, "pipeline_started", first["kind"])
	assert.Equal(t, "p", first["pipeline"])
	_, hasOutput := first["output"]
	assert.False(t, hasOutput, "empty fields must be omitted")

	var second map[string]any
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &second))
	assert.Equal(t, "step_skipped", second["kind"])
	assert.Equal(t, "dry_run", second["skipReason"])
	assert.Equal(t, "helm.install: chart", second["output"])
}

func TestNDJSONSink_SerializesErrorAsString(t *testing.T) {
	var buf bytes.Buffer
	s := newNDJSONSink(&buf, nil)

	require.NoError(t, s.Handle(orchestrator2.Event{
		Kind:  orchestrator2.EventKindStepFailed,
		Step:  "x",
		Error: errors.New("nope"),
	}))

	var got map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, "nope", got["error"])
}

type recordingProgressSink struct{ got [][]string }

func (r *recordingProgressSink) Handle(orchestrator2.Event) error { return nil }
func (r *recordingProgressSink) Close() error                     { return nil }
func (r *recordingProgressSink) Progress(lines []string)          { r.got = append(r.got, lines) }

type plainSink struct{}

func (plainSink) Handle(orchestrator2.Event) error { return nil }
func (plainSink) Close() error                     { return nil }

// TestManager_ProgressForwardsToProgressSinks verifies Progress reaches sinks
// that implement it (skipping those that don't) and records no events.
func TestManager_ProgressForwardsToProgressSinks(t *testing.T) {
	pr := &recordingProgressSink{}
	m := &Manager{sinks: []sink{plainSink{}, pr}}

	m.Progress([]string{"a", "b"})

	require.Equal(t, [][]string{{"a", "b"}}, pr.got)
	assert.Empty(t, m.GetEvents(), "progress must not be recorded as events")
}

// TestTextSink_ProgressNoActiveStep verifies Progress is a safe no-op when no
// step is in flight (e.g. between steps or on a non-TTY).
func TestTextSink_ProgressNoActiveStep(t *testing.T) {
	s := newTextSink(terminal.New(terminal.WithWriter(&bytes.Buffer{})))
	assert.NotPanics(t, func() { s.Progress([]string{"anything"}) })
}

// TestTruncateLine verifies lines are cut to the terminal width (with an
// ellipsis) so the live area never wraps — a wrapped line desyncs the redraw.
func TestTruncateLine(t *testing.T) {
	assert.Equal(t, "abc", truncateLine("abc", 5), "shorter than max is unchanged")
	assert.Equal(t, "abc", truncateLine("abc", 3), "exactly max is unchanged")
	assert.Equal(t, "ab…", truncateLine("abcdef", 3), "longer than max is cut with an ellipsis")
	assert.Equal(t, "abcdef", truncateLine("abcdef", 0), "max <= 0 disables truncation")
}

func TestTextSink_RendersStepResults(t *testing.T) {
	var out, errOut bytes.Buffer
	term := terminal.New(terminal.WithWriter(&out), terminal.WithErrWriter(&errOut))
	s := newTextSink(term)

	require.NoError(t, s.Handle(orchestrator2.Event{
		Kind:      orchestrator2.EventKindStepSucceeded,
		Step:      "check-x",
		StepTitle: "Check X",
		Results: []orchestrator2.ActionResult{
			{Status: orchestrator2.StatusWarn, Reason: "minor thing", Remediation: "do this"},
			{Status: orchestrator2.StatusFail, Reason: "bad thing", Remediation: "fix that"},
		},
	}))

	assert.Contains(t, out.String(), "Check X: minor thing (do this)", "warnings render on stdout, prefixed with the step name")
	assert.Contains(t, errOut.String(), "Check X: bad thing (fix that)", "failures render on stderr, prefixed with the step name")
}

func TestTextSink_RendersAllKnownKinds(t *testing.T) {
	var out, errOut bytes.Buffer
	term := terminal.New(terminal.WithWriter(&out), terminal.WithErrWriter(&errOut))
	s := newTextSink(term)

	// Events carry the human titles the runner stamps; the sink renders those
	// (not the machine names). The machine names stay on the events for JSON.
	events := []orchestrator2.Event{
		{Kind: orchestrator2.EventKindPipelineStarted, Pipeline: "p", PipelineTitle: "Setup"},
		{Kind: orchestrator2.EventKindFlowStarted, Pipeline: "p", Flow: "f", FlowTitle: "Flow one"},

		{Kind: orchestrator2.EventKindStepStarted, Pipeline: "p", Flow: "f", Step: "s", StepTitle: "Step one"},
		{Kind: orchestrator2.EventKindStepSucceeded, Pipeline: "p", Flow: "f", Step: "s", StepTitle: "Step one", Output: "ok"},

		{Kind: orchestrator2.EventKindStepStarted, Pipeline: "p", Flow: "f", Step: "s2", StepTitle: "Step two"},
		{Kind: orchestrator2.EventKindStepSkipped, Pipeline: "p", Flow: "f", Step: "s2", StepTitle: "Step two", SkipReason: orchestrator2.SkipReasonDryRun, Output: "helm.install: chart"},

		{Kind: orchestrator2.EventKindStepStarted, Pipeline: "p", Flow: "f", Step: "s3", StepTitle: "Step three"},
		{Kind: orchestrator2.EventKindStepFailed, Pipeline: "p", Flow: "f", Step: "s3", StepTitle: "Step three", Output: "execute_helm", Error: errors.New("kaboom")},

		{Kind: orchestrator2.EventKindInputValidationFailed, Findings: []orchestrator2.ValidationFinding{
			{Check: "cert", Field: "controlPlane.tls.certPath", Message: "Certificate expired"},
		}},

		{Kind: orchestrator2.EventKindPreflightCompleted, Preflight: &orchestrator2.PreflightSummary{
			Outcome: "warning", Passed: 2, Warned: 1, Failed: 0,
		}},
		{Kind: orchestrator2.EventKindSelfSignedTLS, SelfSignedTLS: &orchestrator2.SelfSignedTLSWarning{
			OutputDir: "/tmp/runai_certs",
			Message:   "self-signed warning",
		}},

		{Kind: orchestrator2.EventKindFlowSucceeded, Pipeline: "p", Flow: "f", FlowTitle: "Flow one"},
		{Kind: orchestrator2.EventKindPipelineSucceeded, Pipeline: "p", PipelineTitle: "Setup"},
	}
	for _, e := range events {
		require.NoError(t, s.Handle(e))
	}

	stdout := out.String()
	stderr := errOut.String()

	assert.Contains(t, stdout, "Setup", "pipeline heading (title) goes to stdout")
	assert.Contains(t, stdout, "Flow one")
	assert.Contains(t, stdout, "SUCCESS")
	assert.Contains(t, stdout, "ok")
	assert.Contains(t, stdout, "SKIPPED")
	assert.Contains(t, stdout, "Step two skipped (dry_run)")
	assert.Contains(t, stdout, "helm.install: chart")
	assert.Contains(t, stdout, "Setup complete")
	assert.Contains(t, stdout, "Input validation failed", "validation heading goes to stdout")
	assert.Contains(t, stdout, "Preflight checks: 2 passed, 1 warning(s), 0 failed", "preflight warning summary renders to stdout")
	assert.Contains(t, stdout, "self-signed warning")

	assert.Contains(t, stderr, "ERROR")
	assert.Contains(t, stderr, "Step three failed (execute_helm): kaboom")
	assert.Contains(t, stderr, "controlPlane.tls.certPath: Certificate expired")
}

func TestTextSink_RendersPrereqSummary(t *testing.T) {
	tests := []struct {
		name       string
		summary    orchestrator2.PrereqSummary
		wantText   string
		wantStdout string
		wantStderr string
	}{
		{
			name:       "completed when nothing installed",
			summary:    orchestrator2.PrereqSummary{Outcome: "completed", Installed: 0, AlreadyPresent: 3},
			wantText:   "Prerequisites: 0 installed, 3 already present, 0 skipped, 0 failed",
			wantStdout: "SUCCESS",
		},
		{
			name:       "installed",
			summary:    orchestrator2.PrereqSummary{Outcome: "installed", Installed: 3},
			wantText:   "Prerequisites: 3 installed, 0 already present, 0 skipped, 0 failed",
			wantStdout: "SUCCESS",
		},
		{
			name:       "partial",
			summary:    orchestrator2.PrereqSummary{Outcome: "partial", Installed: 2, Failed: 1},
			wantText:   "Prerequisites: 2 installed, 0 already present, 0 skipped, 1 failed",
			wantStdout: "WARNING",
		},
		{
			name:       "failed",
			summary:    orchestrator2.PrereqSummary{Outcome: "failed", Failed: 2},
			wantText:   "Prerequisites: 0 installed, 0 already present, 0 skipped, 2 failed",
			wantStderr: "ERROR",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			term := terminal.New(terminal.WithWriter(&out), terminal.WithErrWriter(&errOut))
			s := newTextSink(term)

			require.NoError(t, s.Handle(orchestrator2.Event{
				Kind:   orchestrator2.EventKindPrereqCompleted,
				Prereq: &tc.summary,
			}))

			stream := out.String()
			if tc.wantStderr != "" {
				stream = errOut.String()
				assert.Contains(t, errOut.String(), tc.wantStderr)
			} else {
				assert.Contains(t, out.String(), tc.wantStdout)
			}
			assert.Contains(t, stream, tc.wantText)
		})
	}
}

func TestTextSink_InputValidationFinding_WithoutField(t *testing.T) {
	var out, errOut bytes.Buffer
	term := terminal.New(terminal.WithWriter(&out), terminal.WithErrWriter(&errOut))
	s := newTextSink(term)

	require.NoError(t, s.Handle(orchestrator2.Event{
		Kind: orchestrator2.EventKindInputValidationFailed,
		Findings: []orchestrator2.ValidationFinding{
			{Check: "bundle", Message: "air-gap bundle is required"},
		},
	}))

	stderr := errOut.String()
	assert.Contains(t, stderr, "ERROR: air-gap bundle is required")
	assert.NotContains(t, stderr, "ERROR: :", "a finding without a field must render the message alone, with no empty field prefix")
}

func TestNew_FormatJSON_WritesNDJSONToStdout(t *testing.T) {
	m, err := New(Config{Format: spec.OutputFormatJSON})
	require.NoError(t, err)
	require.Len(t, m.sinks, 1)
	_, ok := m.sinks[0].(*ndjsonSink)
	assert.True(t, ok)
}

func TestNew_FormatText_AddsTextSink(t *testing.T) {
	m, err := New(Config{Format: spec.OutputFormatText})
	require.NoError(t, err)
	require.Len(t, m.sinks, 1)
	_, ok := m.sinks[0].(*textSink)
	assert.True(t, ok)
}

func TestNew_FormatNone_NoStdoutSink(t *testing.T) {
	m, err := New(Config{Format: spec.OutputFormatNone})
	require.NoError(t, err)
	assert.Empty(t, m.sinks)
}

func TestNew_EventsFile_WritesToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.ndjson")

	m, err := New(Config{Format: spec.OutputFormatNone, EventsFile: path})
	require.NoError(t, err)
	require.Len(t, m.sinks, 1)

	require.NoError(t, m.HandleEvent(orchestrator2.Event{Kind: orchestrator2.EventKindPipelineStarted, Pipeline: "p"}))
	require.NoError(t, m.Close())

	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimRight(contents, "\n"), &got))
	assert.Equal(t, "pipeline_started", got["kind"])
}

func TestNew_InvalidFormat_ReturnsError(t *testing.T) {
	_, err := New(Config{Format: spec.OutputFormat("zzz")})
	require.Error(t, err)
}
