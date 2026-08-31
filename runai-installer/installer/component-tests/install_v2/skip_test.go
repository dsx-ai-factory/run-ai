/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install_v2_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/installconfig"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/specrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runSkipInstall executes a connected install with the given config mutation and
// returns the fake helm installer and decoded events for assertions.
func runSkipInstall(t *testing.T, mutate func(*installconfig.Config)) (*harness.FakeHelmInstaller, []map[string]any) {
	t.Helper()
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	mutate(cfg)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))
	return helm, readEvents(t, eventsPath)
}

// TestConnectedInstall_PrerequisitesSkip verifies the global prerequisites.skip
// flag removes every prerequisite step from the plan while the Run:ai install
// still runs. The skip-off control proves the prerequisites would otherwise
// install, so the skip case is a real assertion.
func TestConnectedInstall_PrerequisitesSkip(t *testing.T) {
	prereqReleases := []string{"kube-prometheus-stack", "gpu-operator", "network-operator"}

	t.Run("control: prerequisites install when not skipped", func(t *testing.T) {
		helm, evs := runSkipInstall(t, func(c *installconfig.Config) { c.Prerequisites.Skip = false })
		for _, r := range prereqReleases {
			assert.Len(t, helm.CallsByRelease(r), 1, "%s installs when not skipped", r)
		}
		assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
	})

	t.Run("prerequisites skipped", func(t *testing.T) {
		helm, evs := runSkipInstall(t, func(c *installconfig.Config) { c.Prerequisites.Skip = true })
		for _, r := range prereqReleases {
			assert.Empty(t, helm.CallsByRelease(r), "%s must not install when prerequisites are skipped", r)
		}
		assert.Nil(t, findEvent(evs, "step_started", "install-prometheus"), "no prerequisite step is planned")
		require.Len(t, helm.CallsByRelease("runai-backend"), 1, "Run:ai control plane still installs")
		assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
	})
}

// TestConnectedInstall_SinglePrerequisiteSkip verifies that setting one
// prerequisite to mode: skip omits only that component while its siblings still
// install, and the skipped one is absent from the plan and the summary.
func TestConnectedInstall_SinglePrerequisiteSkip(t *testing.T) {
	helm, evs := runSkipInstall(t, func(c *installconfig.Config) {
		c.Prerequisites.Prometheus.Mode = "skip"
	})

	assert.Empty(t, helm.CallsByRelease("kube-prometheus-stack"), "skipped prerequisite must not install")
	assert.Nil(t, findEvent(evs, "step_started", "install-prometheus"), "skipped prerequisite is not planned")
	assert.Len(t, helm.CallsByRelease("gpu-operator"), 1, "other prerequisites still install")
	assert.Len(t, helm.CallsByRelease("network-operator"), 1)

	summary := prereqSummaryEvent(t, evs)
	assert.EqualValues(t, 2, summary["installed"], "gpu + network install; prometheus omitted")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
}

// TestConnectedInstall_FeaturesSkip verifies the global features.skip flag
// removes every feature step from the plan while the Run:ai install still runs.
func TestConnectedInstall_FeaturesSkip(t *testing.T) {
	withLWS := func(skip bool) func(*installconfig.Config) {
		return func(c *installconfig.Config) {
			c.Features = &installconfig.Features{Skip: skip, LWS: &installconfig.Dependency{Mode: "install"}}
		}
	}

	t.Run("control: feature installs when not skipped", func(t *testing.T) {
		helm, _ := runSkipInstall(t, withLWS(false))
		assert.Len(t, helm.CallsByRelease("lws"), 1, "lws installs when not skipped")
	})

	t.Run("features skipped", func(t *testing.T) {
		helm, evs := runSkipInstall(t, withLWS(true))
		assert.Empty(t, helm.CallsByRelease("lws"), "lws must not install when features are skipped")
		assert.Nil(t, findEvent(evs, "step_started", "install-lws"), "no feature step is planned")
		require.Len(t, helm.CallsByRelease("runai-backend"), 1, "Run:ai control plane still installs")
		assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
	})
}
