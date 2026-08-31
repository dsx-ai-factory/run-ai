/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install_v2_test

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/installconfig"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/internal/specrun"
	"github.com/run-ai/runai/runai-installer/installer/internal/terminal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// interactiveTerminal builds a terminal that reports as a TTY and answers every
// confirm prompt with the given answer, so tests can drive the interactive
// proceed-past-failure paths without a real stdin.
func interactiveTerminal(answer bool) *terminal.Terminal {
	return terminal.New(
		terminal.WithWriter(io.Discard),
		terminal.WithErrWriter(io.Discard),
		terminal.WithTTY(true),
		terminal.WithConfirmFunc(func(string, bool) (bool, error) { return answer, nil }),
	)
}

// badClusterCertConfig returns a connected install whose cluster certificate
// does not match its domain, producing exactly one input-validation finding.
func badClusterCertConfig(t *testing.T, eventsPath string) (*installconfig.Config, installconfig.RunOptions) {
	t.Helper()
	cfg, opts := connectedConfig(t, eventsPath, false)
	opts.Yes = false // interactive
	cert, key := harness.GenCert(t, []string{"wrong.example.com"}, time.Now().Add(24*time.Hour))
	cfg.Cluster.TLS = &installconfig.TLS{CertPath: cert, KeyPath: key}
	return cfg, opts
}

// TestInputValidation_InteractiveProceed verifies that an interactive user can
// choose to continue past input-validation findings: the findings are still
// reported, but the install proceeds instead of hard-stopping.
func TestInputValidation_InteractiveProceed(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := badClusterCertConfig(t, eventsPath)

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, interactiveTerminal(true), execOpts()))

	evs := readEvents(t, eventsPath)
	require.NotNil(t, findValidationEvent(evs), "the findings must still be surfaced")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"), "the install must proceed after the user confirms")
	assert.NotEmpty(t, helm.Calls, "install must run after proceeding")
}

// TestInputValidation_InteractiveDecline verifies that declining the prompt
// aborts with the validation error and no side effects.
func TestInputValidation_InteractiveDecline(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := badClusterCertConfig(t, eventsPath)

	err := specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, interactiveTerminal(false), execOpts())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "input validation failed")

	evs := readEvents(t, eventsPath)
	require.NotNil(t, findValidationEvent(evs))
	assert.False(t, hasPipelineEvent(evs, "pipeline_started", "runai"), "declining must not start the install")
	assert.Empty(t, helm.Calls)
	assert.False(t, apiClient.ClusterCreated)
}

// TestInputValidation_OnFailureSkipProceeds verifies that an explicit
// --on-failure=skip overrides the interactive prompt: the run proceeds past the
// findings without ever consulting the user (the confirm func would decline).
func TestInputValidation_OnFailureSkipProceeds(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := badClusterCertConfig(t, eventsPath)
	opts.OnFailure = spec.FailurePolicySkip

	exec := execOpts()
	exec.OnFailureSet = true

	// interactiveTerminal(false) would decline if prompted; the explicit policy
	// must take precedence and proceed anyway.
	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, interactiveTerminal(false), exec))

	evs := readEvents(t, eventsPath)
	require.NotNil(t, findValidationEvent(evs))
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"), "an explicit skip policy proceeds without prompting")
}

// TestPreflight_InteractiveProceed verifies that an interactive user can
// continue past a failed preflight and the install runs.
func TestPreflight_InteractiveProceed(t *testing.T) {
	deps, helm, _, _ := depsWithProbe(&harness.FakeClusterProbe{VersionErr: errors.New("connection refused"), AllowAll: true})

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Yes = false
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, interactiveTerminal(true), execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "failed", preflightOutcome(t, evs))
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"), "confirming proceeds past a failed preflight")
	assert.NotEmpty(t, helm.Calls)
}

// TestPreflight_InteractiveDecline verifies that declining a failed preflight
// stops the run with ErrPreflightFailed and no side effects.
func TestPreflight_InteractiveDecline(t *testing.T) {
	deps, helm, k8s, _ := depsWithProbe(&harness.FakeClusterProbe{VersionErr: errors.New("connection refused"), AllowAll: true})

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Yes = false
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath

	err := specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, interactiveTerminal(false), execOpts())
	require.ErrorIs(t, err, specrun.ErrPreflightFailed)

	evs := readEvents(t, eventsPath)
	assert.False(t, hasPipelineEvent(evs, "pipeline_started", "runai"))
	assert.Empty(t, helm.Calls)
	assert.Empty(t, k8s.CallsByVerb("apply"))
}

// TestPreflight_DryRunInteractiveProceeds guards the dry-run fix: a dry run must
// not hard-stop on a failed preflight. The user is prompted, proceeds, and the
// install pipeline runs (in dry-run, so no side effects).
func TestPreflight_DryRunInteractiveProceeds(t *testing.T) {
	deps, helm, k8s, _ := depsWithProbe(&harness.FakeClusterProbe{VersionErr: errors.New("connection refused"), AllowAll: true})

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Yes = false
	opts.DryRun = true
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, interactiveTerminal(true), execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "failed", preflightOutcome(t, evs))
	assert.True(t, hasPipelineEvent(evs, "pipeline_started", "runai"), "a dry run must proceed past a failed preflight, not hard-stop")
	assert.Empty(t, helm.Calls, "dry-run performs no installs")
	assert.Empty(t, k8s.CallsByVerb("apply"), "dry-run performs no writes")
}
