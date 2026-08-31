/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install_v2_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/installconfig"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/specrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// externalConnResult returns the (single) result the external-connection step
// reported, failing the test when the step or its result is missing.
func externalConnResult(t *testing.T, evs []map[string]any) map[string]any {
	t.Helper()
	step := findEvent(evs, "step_succeeded", "preflight-external-connection")
	require.NotNil(t, step, "the external-connection check must run")
	results, ok := step["results"].([]any)
	require.True(t, ok && len(results) > 0, "the external-connection check must report a finding")
	return results[0].(map[string]any)
}

func runPreflightWith(t *testing.T, cfg *installconfig.Config, probe *harness.FakeClusterProbe) (error, []map[string]any, *harness.FakeHelmInstaller) {
	t.Helper()
	deps, helm, _, _ := depsWithProbe(probe)
	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()
	err := specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts())
	return err, readEvents(t, eventsPath), helm
}

// TestPreflight_ExternalConnection_IngressClassMissing verifies that a missing
// IngressClass fails preflight and blocks the install.
func TestPreflight_ExternalConnection_IngressClassMissing(t *testing.T) {
	// AllowAll is unset and no IngressClasses are seeded, so the nginx class the
	// config selects is reported absent.
	probe := &harness.FakeClusterProbe{Version: "v1.30.0"}
	err, evs, helm := runPreflightWith(t, newInstallRunConfig(), probe)

	require.ErrorIs(t, err, specrun.ErrPreflightFailed)
	assert.Equal(t, "failed", preflightOutcome(t, evs))

	result := externalConnResult(t, evs)
	assert.Equal(t, "fail", result["status"])
	assert.Contains(t, result["reason"], `ingress class "nginx" was not found`)

	assert.False(t, hasPipelineEvent(evs, "pipeline_started", "runai"), "install must not start after a failed preflight")
	assert.Empty(t, helm.Calls)
}

// TestPreflight_ExternalConnection_IngressClassPresent verifies the check passes
// (no finding) and the install proceeds when the configured class exists.
func TestPreflight_ExternalConnection_IngressClassPresent(t *testing.T) {
	probe := &harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true, IngressClasses: map[string]bool{"nginx": true}}
	err, evs, helm := runPreflightWith(t, newInstallRunConfig(), probe)

	require.NoError(t, err)
	step := findEvent(evs, "step_succeeded", "preflight-external-connection")
	require.NotNil(t, step, "the external-connection check must run")
	assert.Empty(t, step["results"], "no finding when the ingress class exists")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
	assert.NotEmpty(t, helm.Calls)
}

// TestPreflight_ExternalConnection_IngressClassProbeError verifies a probe error
// is a non-blocking warning (the install still proceeds).
func TestPreflight_ExternalConnection_IngressClassProbeError(t *testing.T) {
	probe := &harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true, IngressClassErr: errors.New("forbidden")}
	err, evs, helm := runPreflightWith(t, newInstallRunConfig(), probe)

	require.NoError(t, err)
	assert.Equal(t, "warning", preflightOutcome(t, evs))
	result := externalConnResult(t, evs)
	assert.Equal(t, "warn", result["status"])
	assert.Contains(t, result["reason"], "could not verify ingress class")
	assert.NotEmpty(t, helm.Calls, "a warning must not block the install")
}

func gatewayConfig() *installconfig.Config {
	cfg := newInstallRunConfig()
	cfg.ExternalConnection = &installconfig.ExternalConnection{
		Type:       "gatewayApi",
		GatewayAPI: &installconfig.GatewayAPIConnection{Name: "my-gw", Namespace: "gw-system"},
	}
	return cfg
}

// TestPreflight_ExternalConnection_GatewayAPINotInstalled verifies preflight
// fails when the Gateway API is not installed on the cluster.
func TestPreflight_ExternalConnection_GatewayAPINotInstalled(t *testing.T) {
	probe := &harness.FakeClusterProbe{Version: "v1.30.0", GatewayAPIAbsent: true}
	err, evs, helm := runPreflightWith(t, gatewayConfig(), probe)

	require.ErrorIs(t, err, specrun.ErrPreflightFailed)
	result := externalConnResult(t, evs)
	assert.Equal(t, "fail", result["status"])
	assert.Contains(t, result["reason"], "Gateway API is not installed")
	assert.Empty(t, helm.Calls)
}

// TestPreflight_ExternalConnection_GatewayNotFound verifies preflight fails when
// the referenced Gateway does not exist (Gateway API is installed).
func TestPreflight_ExternalConnection_GatewayNotFound(t *testing.T) {
	// GatewayAPIAbsent is false so the API is installed; no gateways seeded so
	// the referenced one is not found.
	probe := &harness.FakeClusterProbe{Version: "v1.30.0"}
	err, evs, helm := runPreflightWith(t, gatewayConfig(), probe)

	require.ErrorIs(t, err, specrun.ErrPreflightFailed)
	result := externalConnResult(t, evs)
	assert.Equal(t, "fail", result["status"])
	assert.Contains(t, result["reason"], `gateway gw-system/my-gw was not found`)
	assert.Empty(t, helm.Calls)
}

// TestPreflight_ExternalConnection_GatewayNotProgrammed verifies preflight fails
// when the Gateway exists but the controller has not programmed it.
func TestPreflight_ExternalConnection_GatewayNotProgrammed(t *testing.T) {
	probe := &harness.FakeClusterProbe{
		Version:  "v1.30.0",
		Gateways: map[string]orchestrator2.GatewayStatus{"gw-system/my-gw": {Found: true, Accepted: true, Programmed: false}},
	}
	err, evs, helm := runPreflightWith(t, gatewayConfig(), probe)

	require.ErrorIs(t, err, specrun.ErrPreflightFailed)
	result := externalConnResult(t, evs)
	assert.Equal(t, "fail", result["status"])
	assert.Contains(t, result["reason"], "is not ready")
	assert.Contains(t, result["reason"], "programmed=false")
	assert.Empty(t, helm.Calls)
}

// TestPreflight_ExternalConnection_GatewayReady verifies the check passes (no
// finding) and the install proceeds when the Gateway is programmed.
func TestPreflight_ExternalConnection_GatewayReady(t *testing.T) {
	probe := &harness.FakeClusterProbe{
		Version:  "v1.30.0",
		AllowAll: true,
		Gateways: map[string]orchestrator2.GatewayStatus{"gw-system/my-gw": {Found: true, Accepted: true, Programmed: true}},
	}
	err, evs, helm := runPreflightWith(t, gatewayConfig(), probe)

	require.NoError(t, err)
	step := findEvent(evs, "step_succeeded", "preflight-external-connection")
	require.NotNil(t, step, "the external-connection check must run")
	assert.Empty(t, step["results"], "no finding when the gateway is programmed")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
	assert.NotEmpty(t, helm.Calls)
}
