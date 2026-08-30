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

// TestInstallMode_ControlPlaneOnly_InstallsControlPlaneOnly verifies that with
// install.mode=runai_control_plane_only (and no cluster configured) the control
// plane installs and the cluster flow never runs.
func TestInstallMode_ControlPlaneOnly_InstallsControlPlaneOnly(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Install.Mode = "runai_control_plane_only"
	cfg.Cluster.Enabled = false
	cfg.Prerequisites = &installconfig.Prerequisites{Skip: true}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	assert.NotEmpty(t, helm.CallsByRelease("runai-backend"), "control plane must install")
	assert.Empty(t, helm.CallsByRelease("runai-cluster"), "cluster must not install in control-plane-only mode")
	requireTLSSecret(t, k8s, "runai-backend", "runai-backend-tls", cfg.ControlPlane.TLS)
	assert.False(t, appliedObject(k8s, "runai", "runai-cluster-domain-tls-secret"))
	assert.False(t, appliedObject(k8s, "runai-backend", "runai-ca-cert"))
	require.NotEmpty(t, apiClient.ProviderCalls)
	for _, call := range apiClient.ProviderCalls {
		assert.Empty(t, call.CACertPath)
	}

	evs := readEvents(t, eventsPath)
	require.NotNil(t, findEvent(evs, "step_succeeded", "install-control-plane-chart"))
	assert.Nil(t, findEvent(evs, "step_started", "install-cluster-chart"), "no cluster flow in control-plane-only mode")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
}

// TestInstallMode_ClusterOnly_InstallsClusterOnly verifies that with
// install.mode=runai_cluster_only (and no control plane configured) the cluster
// installs against an already-present control plane and the CP flow never runs.
func TestInstallMode_ClusterOnly_InstallsClusterOnly(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	apiClient.CPReady = true // cluster-only installs against an existing control plane
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Install.Mode = "runai_cluster_only"
	cfg.ControlPlane.Enabled = false
	cfg.Cluster.Domain = "test.runai.local" // no CP block to fall back to
	cfg.Prerequisites = &installconfig.Prerequisites{Skip: true}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	assert.NotEmpty(t, helm.CallsByRelease("runai-cluster"), "cluster must install")
	assert.Empty(t, helm.CallsByRelease("runai-backend"), "control plane must not install in cluster-only mode")
	requireTLSSecret(t, k8s, "runai", "runai-cluster-domain-tls-secret", cfg.Cluster.TLS)
	assert.False(t, appliedObject(k8s, "runai-backend", "runai-backend-tls"))
	assert.False(t, appliedObject(k8s, "runai", "runai-ca-cert"))
	require.NotEmpty(t, apiClient.ProviderCalls)
	for _, call := range apiClient.ProviderCalls {
		assert.Empty(t, call.CACertPath)
	}

	evs := readEvents(t, eventsPath)
	require.NotNil(t, findEvent(evs, "step_succeeded", "install-cluster-chart"))
	assert.Nil(t, findEvent(evs, "step_started", "install-control-plane-chart"), "no control-plane flow in cluster-only mode")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
}
