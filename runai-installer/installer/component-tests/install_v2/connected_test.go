/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install_v2_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/installconfig"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/specrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// connectedConfig builds a full connected install (control plane + cluster,
// both with TLS) and the matching prod-env run options writing events to path.
func connectedConfig(t *testing.T, path string, dryRun bool) (*installconfig.Config, installconfig.RunOptions) {
	t.Helper()
	certPath, keyPath := harness.GenCert(t, []string{"test.runai.local"}, time.Now().Add(24*time.Hour))

	cfg := newInstallRunConfig()
	cfg.ControlPlane.Version = "2.26.0"
	cfg.ControlPlane.TLS = &installconfig.TLS{CertPath: certPath, KeyPath: keyPath}
	cfg.Cluster.Version = "2.26.0"
	cfg.Cluster.TLS = &installconfig.TLS{CertPath: certPath, KeyPath: keyPath}
	cfg.Prerequisites = &installconfig.Prerequisites{
		Prometheus:      &installconfig.Dependency{Mode: "install"},
		GPUOperator:     &installconfig.Dependency{Mode: "install", Type: "real"},
		NetworkOperator: &installconfig.Dependency{Mode: "install"},
	}

	opts := baseRunOptions()
	opts.Env = spec.EnvProd
	opts.DryRun = dryRun
	opts.EventsFile = path
	return cfg, opts
}

func appliedNames(k8s *harness.FakeK8sExecutor) map[string]bool {
	out := map[string]bool{}
	for _, c := range k8s.CallsByVerb("apply") {
		out[c.Name] = true
	}
	return out
}

func TestConnectedInstall_FullFlow(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	// Prerequisites install first, with versions resolved from the 2.26 matrix.
	prom := helm.CallsByRelease("kube-prometheus-stack")
	require.Len(t, prom, 1)
	assert.Equal(t, "76.2.0", prom[0].Version)
	assert.Contains(t, prom[0].Values, "grafana.enabled=false")

	gpu := helm.CallsByRelease("gpu-operator")
	require.Len(t, gpu, 1)
	assert.Equal(t, "v25.3.0", gpu[0].Version)
	assert.Contains(t, gpu[0].RepoUrl, "helm.ngc.nvidia.com")

	net := helm.CallsByRelease("network-operator")
	require.Len(t, net, 1)
	assert.Equal(t, "v25.10.0", net[0].Version)
	assert.Empty(t, net[0].KubeVersion, "network operator must not pin a kube version; chart has no upper-bound capability gate")

	cpCalls := helm.CallsByRelease("runai-backend")
	require.Len(t, cpCalls, 1)
	assert.Equal(t, "2.26.0", cpCalls[0].Version)
	assert.Contains(t, cpCalls[0].RepoUrl, "cp-charts-prod")
	assert.Contains(t, cpCalls[0].Values, "global.domain=test.runai.local")

	clCalls := helm.CallsByRelease("runai-cluster")
	require.Len(t, clCalls, 1)
	assert.Equal(t, "2.26.0", clCalls[0].Version)
	assert.Contains(t, clCalls[0].RepoUrl, "run-ai-charts")
	assert.Subset(t, clCalls[0].Values, []string{
		"controlPlane.url=test.runai.local",
		"controlPlane.clientSecret=s3cr3t-airgap",
		"cluster.uid=uuid-airgap-test",
	})

	applied := appliedNames(k8s)
	assert.True(t, applied["runai-backend"], "CP namespace applied")
	assert.True(t, applied["runai"], "cluster namespace applied")
	assert.True(t, applied["runai-reg-creds"], "registry secret applied")
	assert.True(t, applied["runai-backend-tls"], "CP tls secret applied")
	assert.True(t, applied["runai-cluster-domain-tls-secret"], "cluster tls secret applied")

	assert.True(t, apiClient.ClusterCreated, "cluster registered via control-plane API")

	evs := readEvents(t, eventsPath)
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "prereq"))
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
	assert.False(t, hasPipelineEvent(evs, "pipeline_failed", ""))

	summary := prereqSummaryEvent(t, evs)
	assert.Equal(t, "installed", summary["outcome"])
	assert.EqualValues(t, 3, summary["installed"])
	assert.EqualValues(t, 0, summary["failed"])

	// The run ends with a run_completed rollup covering every pipeline that ran.
	runEv := findEventByKind(evs, "run_completed")
	require.NotNil(t, runEv, "a run_completed event must be emitted")
	run := runEv["run"].(map[string]any)
	assert.Equal(t, "succeeded", run["outcome"])
	pipelines := run["pipelines"].(map[string]any)
	assert.Contains(t, pipelines, "preflight")
	assert.Contains(t, pipelines, "prerequisites")
	assert.Contains(t, pipelines, "runai")
	prereqRollup := pipelines["prerequisites"].(map[string]any)
	assert.EqualValues(t, 3, prereqRollup["installed"], "the run rollup reuses the prerequisites summary")
}

func TestConnectedInstall_FullFlowWithBootstrapToken(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	apiClient.BootstrapToken = "bootstrap-token"
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	require.Len(t, helm.CallsByRelease("runai-backend"), 1)
	clusterCalls := helm.CallsByRelease("runai-cluster")
	require.Len(t, clusterCalls, 1)
	assert.Contains(t, clusterCalls[0].Values, "controlPlane.bootstrapToken=bootstrap-token")
	assert.NotContains(t, clusterCalls[0].Values, "controlPlane.clientSecret=s3cr3t-airgap")
	assert.True(t, apiClient.ClusterCreated)

	events := readEvents(t, eventsPath)
	assert.True(t, hasPipelineEvent(events, "pipeline_succeeded", "runai"))
	assert.False(t, hasPipelineEvent(events, "pipeline_failed", ""))
}

// TestConnectedInstall_CompatFallbackWarns verifies that, in dev mode, an
// unrecognized Run:ai version emits a compat_fallback warning (listing the
// versions that will be installed from the latest matrix) without failing the
// install. Without dev mode the same version is rejected at input validation
// (see TestInputValidation_Failures).
func TestConnectedInstall_CompatFallbackWarns(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.ControlPlane.Version = "2.23.0" // not in the matrix -> fall back to latest
	cfg.Cluster.Version = "2.23.0"
	opts.DevMode = true // an out-of-matrix version is only allowed in dev mode
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)

	fb := findEventByKind(evs, "compat_fallback")
	require.NotNil(t, fb, "an unrecognized Run:ai version must emit a compat_fallback warning")
	compat := fb["compat"].(map[string]any)
	assert.Equal(t, "2.23.0", compat["requestedVersion"])
	assert.Equal(t, "2.26", compat["usingVersion"])

	components := compat["components"].([]any)
	versions := map[string]string{}
	for _, c := range components {
		m := c.(map[string]any)
		versions[m["name"].(string)] = m["version"].(string)
	}
	assert.Equal(t, "76.2.0", versions["prometheus"], "fallback reports the latest matrix versions")
	assert.Equal(t, "v25.3.0", versions["gpu-operator"])

	// The warning does not fail the install: prereqs still install with the
	// fallback versions and the run succeeds.
	assert.Equal(t, "76.2.0", helm.CallsByRelease("kube-prometheus-stack")[0].Version)
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
	assert.False(t, hasPipelineEvent(evs, "pipeline_failed", ""))
}

// TestConnectedInstall_DefaultGPUType_Real verifies that omitting the GPU
// operator type defaults to the real operator (not the fake OCI chart).
func TestConnectedInstall_DefaultGPUType_Real(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Prerequisites.GPUOperator.Type = "" // unset -> defaults to real
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	gpu := helm.CallsByRelease("gpu-operator")
	require.Len(t, gpu, 1)
	assert.Contains(t, gpu[0].RepoUrl, "helm.ngc.nvidia.com", "unset gpu type defaults to the real operator")
	assert.False(t, gpu[0].IsOci, "real operator is not the fake OCI chart")
}

func TestConnectedInstall_DryRun(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, true)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	assert.Empty(t, helm.Calls, "dry-run performs no helm installs")
	assert.Empty(t, k8s.CallsByVerb("apply"), "dry-run performs no k8s writes")
	assert.False(t, apiClient.ClusterCreated, "dry-run does not create cluster")

	evs := readEvents(t, eventsPath)
	assert.Positive(t, skipReasonCount(evs, "dry_run"), "dry-run emits dry_run step_skipped events")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))

	// Under dry-run the prerequisites are reported as skipped, not installed.
	summary := prereqSummaryEvent(t, evs)
	assert.Equal(t, "completed", summary["outcome"], "nothing installed must not report outcome installed")
	assert.EqualValues(t, 0, summary["installed"], "dry-run installs no prerequisites")
	assert.EqualValues(t, 3, summary["skipped"], "all three prerequisites reported as skipped")
}

// TestConnectedInstall_DryRun_DistinguishesInstalled verifies that a dry-run
// runs Detect and reports an already-installed component as skipped for that
// reason, while not-yet-installed steps are reported as would-run (dry_run).
func TestConnectedInstall_DryRun_DistinguishesInstalled(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases(harness.ReleaseEntry{Name: "runai-backend", Version: "2.26.0"})
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, true)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	assert.Empty(t, helm.Calls, "dry-run performs no helm installs")

	cpChart := findEvent(readEvents(t, eventsPath), "step_skipped", "install-control-plane-chart")
	require.NotNil(t, cpChart)
	assert.Equal(t, "already_installed", cpChart["skipReason"], "installed CP chart is reported as already installed")

	cpNamespace := findEvent(readEvents(t, eventsPath), "step_skipped", "create-control-plane-namespace")
	require.NotNil(t, cpNamespace)
	assert.Equal(t, "dry_run", cpNamespace["skipReason"], "not-installed step is reported as would-run")
}

// TestConnectedInstall_FakeGPU verifies the fake-GPU branch: worker nodes are
// labeled and the fake chart is installed from its OCI registry. The network
// operator is independent of GPU flavor, so it still installs when requested.
func TestConnectedInstall_FakeGPU(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	labeler := &harness.FakeNodeLabeler{Labeled: 2}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe(), NodeLabeler: labeler}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Prerequisites.GPUOperator.Type = "fake"
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	require.Len(t, labeler.Calls, 1)
	assert.Equal(t, "run.ai/simulated-gpu-node-pool", labeler.Calls[0].Key)
	assert.Equal(t, "default", labeler.Calls[0].Value)

	gpu := helm.CallsByRelease("gpu-operator")
	require.Len(t, gpu, 1)
	assert.True(t, gpu[0].IsOci)
	assert.Contains(t, gpu[0].RepoUrl, "fake-gpu-operator")
	assert.Equal(t, "0.0.62", gpu[0].Version)
	assert.Contains(t, gpu[0].Values, "topology.nodePools.default.gpuCount=8")

	net := helm.CallsByRelease("network-operator")
	require.Len(t, net, 1, "network operator installs regardless of GPU flavor")
	assert.Equal(t, "v25.10.0", net[0].Version)

	summary := prereqSummaryEvent(t, readEvents(t, eventsPath))
	assert.Equal(t, "installed", summary["outcome"])
	assert.EqualValues(t, 3, summary["installed"])
}

// TestConnectedInstall_NetworkOperator_AlreadyPresent guards the network
// operator sentinel: its controller Deployment is named "network-operator", so
// an existing install must be detected and skipped rather than reinstalled.
func TestConnectedInstall_NetworkOperator_AlreadyPresent(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{Existing: map[client.ObjectKey]bool{
		{Namespace: "nvidia-network-operator", Name: "network-operator"}: true,
	}}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	assert.Empty(t, helm.CallsByRelease("network-operator"), "present network operator is not reinstalled")

	skipped := findEvent(readEvents(t, eventsPath), "step_skipped", "install-network-operator")
	require.NotNil(t, skipped)
	assert.Equal(t, "already_installed", skipped["skipReason"])

	summary := prereqSummaryEvent(t, readEvents(t, eventsPath))
	assert.EqualValues(t, 2, summary["installed"], "prometheus + gpu freshly installed")
	assert.EqualValues(t, 1, summary["alreadyPresent"], "present network operator reported separately")
}

// TestConnectedInstall_Prometheus_AlreadyPresent verifies that a prerequisite
// already running in the cluster is detected and skipped rather than
// reinstalled, and is reported as already_present in the summary.
func TestConnectedInstall_Prometheus_AlreadyPresent(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{Existing: map[client.ObjectKey]bool{
		{Namespace: "monitoring", Name: "kube-prometheus-stack-operator"}: true,
	}}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	assert.Empty(t, helm.CallsByRelease("kube-prometheus-stack"), "present prometheus is not reinstalled")

	skipped := findEvent(readEvents(t, eventsPath), "step_skipped", "install-prometheus")
	require.NotNil(t, skipped)
	assert.Equal(t, "already_installed", skipped["skipReason"])

	summary := prereqSummaryEvent(t, readEvents(t, eventsPath))
	assert.Equal(t, "installed", summary["outcome"])
	assert.EqualValues(t, 2, summary["installed"], "gpu + network freshly installed")
	assert.EqualValues(t, 1, summary["alreadyPresent"], "present prometheus reported separately")
}
