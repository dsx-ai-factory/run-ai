/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install_v2_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/installconfig"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/specrun"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ngcChartsRepo = "https://helm.ngc.nvidia.com/nvidia/runai"

func ngcConfig(t *testing.T, path, token string) (*installconfig.Config, installconfig.RunOptions) {
	t.Helper()
	cfg, opts := connectedConfig(t, path, false)
	opts.Registry = &spec.Registry{Type: spec.RegistryTypeNGC, Token: token}
	return cfg, opts
}

func TestNGCInstall_WithToken(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := ngcConfig(t, eventsPath, "ngc-key")
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	cp := helm.CallsByRelease("runai-backend")
	require.Len(t, cp, 1)
	assert.Equal(t, ngcChartsRepo, cp[0].RepoUrl)
	assert.Equal(t, "control-plane", cp[0].ChartName)
	assert.False(t, cp[0].IsOci, "NGC serves the charts from a classic Helm repo")
	assert.Equal(t, "$oauthtoken", cp[0].Auth.Username)
	assert.Equal(t, "ngc-key", cp[0].Auth.Password)

	cl := helm.CallsByRelease("runai-cluster")
	require.Len(t, cl, 1)
	assert.Equal(t, ngcChartsRepo, cl[0].RepoUrl)
	assert.Equal(t, "runai-cluster", cl[0].ChartName)
	assert.False(t, cl[0].IsOci)
	assert.Equal(t, "$oauthtoken", cl[0].Auth.Username)
	assert.Equal(t, "ngc-key", cl[0].Auth.Password)

	gpu := helm.CallsByRelease("gpu-operator")
	require.Len(t, gpu, 1)
	assert.True(t, gpu[0].Auth.IsZero(), "prerequisite charts pull anonymously")

	assert.True(t, appliedNames(k8s)["runai-reg-creds"], "image pull secret applied for authenticated NGC")

	evs := readEvents(t, eventsPath)
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
	assert.False(t, hasPipelineEvent(evs, "pipeline_failed", ""))
}

func TestNGCInstall_TokenFree(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := ngcConfig(t, eventsPath, "")
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	cp := helm.CallsByRelease("runai-backend")
	require.Len(t, cp, 1)
	assert.Equal(t, ngcChartsRepo, cp[0].RepoUrl)
	assert.False(t, cp[0].IsOci)
	assert.True(t, cp[0].Auth.IsZero(), "token-free NGC pulls the control-plane chart anonymously")

	cl := helm.CallsByRelease("runai-cluster")
	require.Len(t, cl, 1)
	assert.Equal(t, ngcChartsRepo, cl[0].RepoUrl)
	assert.False(t, cl[0].IsOci)
	assert.True(t, cl[0].Auth.IsZero(), "token-free NGC pulls the cluster chart anonymously")

	assert.False(t, appliedNames(k8s)["runai-reg-creds"], "no image pull secret without credentials")

	evs := readEvents(t, eventsPath)
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
	assert.False(t, hasPipelineEvent(evs, "pipeline_failed", ""))
}

func TestNGCInstall_ChartRepositoryOverride(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := ngcConfig(t, eventsPath, "ngc-key")
	opts.Registry.ChartRepository = "https://helm.ngc.nvidia.com/myorg/myteam"
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	cp := helm.CallsByRelease("runai-backend")
	require.Len(t, cp, 1)
	assert.Equal(t, "https://helm.ngc.nvidia.com/myorg/myteam", cp[0].RepoUrl)
	assert.Equal(t, "control-plane", cp[0].ChartName)
	assert.False(t, cp[0].IsOci)

	cl := helm.CallsByRelease("runai-cluster")
	require.Len(t, cl, 1)
	assert.Equal(t, "https://helm.ngc.nvidia.com/myorg/myteam", cl[0].RepoUrl)
	assert.Equal(t, "runai-cluster", cl[0].ChartName)
}

func TestNGCInstall_NonProdChartRepoFromEnv(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := ngcConfig(t, eventsPath, "ngc-key")
	opts.Env = spec.EnvStaging
	t.Setenv(step.NGCNonProdChartRepoEnv, "https://helm.ngc.nvidia.com/myorg/myteam/")
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	cp := helm.CallsByRelease("runai-backend")
	require.Len(t, cp, 1)
	assert.Equal(t, "https://helm.ngc.nvidia.com/myorg/myteam", cp[0].RepoUrl)
	assert.Equal(t, "$oauthtoken", cp[0].Auth.Username)

	cl := helm.CallsByRelease("runai-cluster")
	require.Len(t, cl, 1)
	assert.Equal(t, "https://helm.ngc.nvidia.com/myorg/myteam", cl[0].RepoUrl)
}

func TestDockerConfigInstall_ChartsAnonymous(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	cp := helm.CallsByRelease("runai-backend")
	require.Len(t, cp, 1)
	assert.Contains(t, cp[0].RepoUrl, "cp-charts-prod")
	assert.False(t, cp[0].IsOci, "chart repo base charts are classic Helm repos, not OCI")
	assert.True(t, cp[0].Auth.IsZero(), "chart repo base Helm repos are anonymous")

	cl := helm.CallsByRelease("runai-cluster")
	require.Len(t, cl, 1)
	assert.Contains(t, cl[0].RepoUrl, "run-ai-charts")
	assert.False(t, cl[0].IsOci)
	assert.True(t, cl[0].Auth.IsZero())

	assert.True(t, appliedNames(k8s)["runai-reg-creds"], "docker config install applies the pull secret from its token")
}

func TestDockerConfigInstall_ChartRepositoryOverride(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	// A configured chart repository wins over the environment base, so a
	// private mirror can serve both charts without any environment setup.
	t.Setenv(step.ChartRepoBaseEnv, "")
	opts.Registry.ChartRepository = "https://charts.example.com/mirror/"
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	cp := helm.CallsByRelease("runai-backend")
	require.Len(t, cp, 1)
	assert.Equal(t, "https://charts.example.com/mirror", cp[0].RepoUrl)
	assert.True(t, cp[0].Auth.IsZero())

	cl := helm.CallsByRelease("runai-cluster")
	require.Len(t, cl, 1)
	assert.Equal(t, "https://charts.example.com/mirror", cl[0].RepoUrl)
}
