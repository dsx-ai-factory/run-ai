/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install_v2_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/specrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// retryHandler retries a failed step, counting the calls. It aborts once calls
// exceeds max so a permanently-failing step can't loop; max <= 0 retries always.
type retryHandler struct {
	calls int
	max   int
}

func (h *retryHandler) OnFailure(context.Context, orchestrator2.FailureInfo) orchestrator2.FailureDecision {
	h.calls++
	if h.max > 0 && h.calls > h.max {
		return orchestrator2.FailureAbort
	}
	return orchestrator2.FailureRetry
}

// TestFailure_StepFailureAborts runs the real install plan with the
// control-plane chart install returning an error. With the default abort
// policy the run fails, the cluster flow never starts, and the failure is
// surfaced as step_failed + pipeline_failed events.
func TestFailure_StepFailureAborts(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{InstallErrors: map[string]error{"runai-backend": errors.New("helm error")}}
	term, _, _ := newTestTerminal()

	eventsPath := t.TempDir() + "/events.ndjson"
	opts := baseRunOptions()
	opts.EventsFile = eventsPath
	s := newInstallRunConfig().ToSpec(opts)

	err := specrun.Execute(context.Background(), s, newTestDeps(fakeHelm), term, execOpts())
	require.Error(t, err)

	require.Len(t, fakeHelm.CallsByRelease("runai-backend"), 1, "control-plane install must be attempted")
	assert.Empty(t, fakeHelm.CallsByRelease("runai-cluster"), "cluster must not be attempted after abort")

	evs := readEvents(t, eventsPath)
	failed := findEvent(evs, "step_failed", "install-control-plane-chart")
	require.NotNil(t, failed)
	assert.Contains(t, failed["error"], "helm error")
	assert.True(t, hasPipelineEvent(evs, "pipeline_failed", "runai"))
}

// TestFailure_BrokenReleaseAborts verifies that a Helm release detected in a
// broken state surfaces as a step error (at the detect phase) that the failure
// handler can act on. With the default abort policy the run fails before the
// control-plane chart is installed.
func TestFailure_BrokenReleaseAborts(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{BrokenReleases: map[string]bool{"runai-backend": true}}
	term, _, _ := newTestTerminal()

	eventsPath := t.TempDir() + "/events.ndjson"
	opts := baseRunOptions()
	opts.EventsFile = eventsPath
	s := newInstallRunConfig().ToSpec(opts)

	err := specrun.Execute(context.Background(), s, newTestDeps(fakeHelm), term, execOpts())
	require.Error(t, err)

	assert.Empty(t, fakeHelm.CallsByRelease("runai-backend"), "broken release must not be re-driven")

	evs := readEvents(t, eventsPath)
	failed := findEvent(evs, "step_failed", "install-control-plane-chart")
	require.NotNil(t, failed)
	assert.Equal(t, "detect", failed["output"])
	assert.True(t, hasPipelineEvent(evs, "pipeline_failed", "runai"))
}

// TestTimeout_StepTimeoutAborts exercises the per-step deadline through the real
// install plan: the control-plane chart install is made to block until its
// context is cancelled, so a short StepTimeout aborts the step.
func TestTimeout_StepTimeoutAborts(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{Block: map[string]bool{"runai-backend": true}}
	term, _, _ := newTestTerminal()

	eventsPath := t.TempDir() + "/events.ndjson"
	opts := baseRunOptions()
	opts.EventsFile = eventsPath
	s := newInstallRunConfig().ToSpec(opts)

	exec := execOpts()
	exec.StepTimeout = 100 * time.Millisecond

	err := specrun.Execute(context.Background(), s, newTestDeps(fakeHelm), term, exec)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	evs := readEvents(t, eventsPath)
	require.NotNil(t, findEvent(evs, "step_failed", "install-control-plane-chart"))
}

// TestTimeout_RetryGetsFreshTimeoutAndSucceeds verifies that a step which times
// out on its first attempt is retried with a brand-new step-timeout window
// rather than the already-expired deadline of the first attempt. The
// control-plane install blocks (and times out) once, the handler asks to retry,
// and the second attempt completes within a fresh timeout.
func TestTimeout_RetryGetsFreshTimeoutAndSucceeds(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{BlockAttempts: map[string]int{"runai-backend": 1}}
	term, _, _ := newTestTerminal()

	eventsPath := t.TempDir() + "/events.ndjson"
	opts := baseRunOptions()
	opts.EventsFile = eventsPath
	s := newInstallRunConfig().ToSpec(opts)

	handler := &retryHandler{}
	exec := execOpts()
	exec.StepTimeout = 150 * time.Millisecond
	exec.FailureHandler = handler

	err := specrun.Execute(context.Background(), s, newTestDeps(fakeHelm), term, exec)
	require.NoError(t, err)

	assert.Equal(t, 1, handler.calls, "first attempt should time out and trigger exactly one retry")
	require.Len(t, fakeHelm.CallsByRelease("runai-backend"), 2, "control-plane install must be attempted twice")

	evs := readEvents(t, eventsPath)
	require.NotNil(t, findEvent(evs, "step_succeeded", "install-control-plane-chart"))
	require.NotNil(t, findEvent(evs, "step_succeeded", "install-cluster-chart"))
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
}

// TestRetry_RetriesMoreThanOnceUntilSuccess verifies the runner honors repeated
// retry decisions rather than capping at one. The control-plane install fails
// (times out) three times; the handler retries each time and the fourth attempt
// succeeds. Before the cap was removed this aborted on the second retry.
func TestRetry_RetriesMoreThanOnceUntilSuccess(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{BlockAttempts: map[string]int{"runai-backend": 3}}
	term, _, _ := newTestTerminal()

	eventsPath := t.TempDir() + "/events.ndjson"
	opts := baseRunOptions()
	opts.EventsFile = eventsPath
	s := newInstallRunConfig().ToSpec(opts)

	handler := &retryHandler{max: 10}
	exec := execOpts()
	exec.StepTimeout = 150 * time.Millisecond
	exec.FailureHandler = handler

	err := specrun.Execute(context.Background(), s, newTestDeps(fakeHelm), term, exec)
	require.NoError(t, err)

	assert.Equal(t, 3, handler.calls, "the step must be retried three times before it succeeds")
	require.Len(t, fakeHelm.CallsByRelease("runai-backend"), 4, "control-plane install must be attempted four times (3 failures + 1 success)")

	evs := readEvents(t, eventsPath)
	require.NotNil(t, findEvent(evs, "step_succeeded", "install-control-plane-chart"))
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
}
