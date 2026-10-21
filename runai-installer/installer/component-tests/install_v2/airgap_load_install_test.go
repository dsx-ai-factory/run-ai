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

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/airgapvalues"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/installconfig"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/specrun"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/artifactlock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const airgapLoadDestination = "registry.local:5000/runai"

func TestAirgapInstall_LoadsArtifactBundleAndInstalls(t *testing.T) {
	lock := artifactInstallLock()
	bundle := harness.CreateLoadBundle(t, lock)
	workingDir := t.TempDir()
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	apiClient.BootstrapToken = "bootstrap-token"
	pusher := &harness.FakeImagePusher{}
	deps := orchestrator2.Deps{
		K8sExecutor:         k8s,
		HelmInstaller:       helm,
		RunAIClientProvider: provider,
		ClusterProbe:        okProbe(),
		ImagePusher:         pusher,
	}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := artifactInstallConfig(bundle, workingDir, eventsPath)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	assert.ElementsMatch(t, []string{
		airgapLoadDestination + "/runai-control-plane/assets-service:2.26.13",
		airgapLoadDestination + "/runai-control-plane/postgresql:16",
	}, pusher.Pushes)
	cpCalls := helm.CallsByRelease("runai-backend")
	require.Len(t, cpCalls, 1)
	assert.Equal(t, filepath.Join(workingDir, "bundle", "charts", "control-plane-3.0.0.tgz"), cpCalls[0].ChartName)
	assert.Equal(t, []string{filepath.Join(workingDir, airgapvalues.ControlPlaneValuesFile)}, cpCalls[0].ValueFiles)
	assert.Contains(t, cpCalls[0].Values, "global.imagePullSecrets[0].name=runai-reg-creds")

	clusterCalls := helm.CallsByRelease("runai-cluster")
	require.Len(t, clusterCalls, 1)
	assert.Equal(t, filepath.Join(workingDir, "bundle", "charts", "runai-cluster-3.0.0.tgz"), clusterCalls[0].ChartName)
	assert.Equal(t, []string{filepath.Join(workingDir, airgapvalues.ClusterValuesFile)}, clusterCalls[0].ValueFiles)
	assert.Contains(t, clusterCalls[0].Values, "global.additionalImagePullSecrets[0]=runai-reg-creds")
	assert.Contains(t, clusterCalls[0].Values, "controlPlane.bootstrapToken=bootstrap-token")
	assert.NotContains(t, clusterCalls[0].Values, "controlPlane.clientSecret=s3cr3t-airgap")

	applied := appliedNames(k8s)
	assert.True(t, applied["runai-reg-creds"])
	events := readEvents(t, eventsPath)
	for _, stepName := range []string{"extract-bundle", "push-images", "generate-values", "install-control-plane-chart", "install-cluster-chart"} {
		require.NotNil(t, findEvent(events, "step_succeeded", stepName))
	}
	assert.Nil(t, findEvent(events, "step_started", "extract-airgap-package"))
	assert.True(t, hasPipelineEvent(events, "pipeline_succeeded", "runai"))
}

func TestAirgapInstall_UsesLoadedArtifactBundle(t *testing.T) {
	lock := artifactInstallLock()
	bundle := harness.CreateLoadBundle(t, lock)
	workingDir := t.TempDir()
	firstHelm := harness.NewFakeHelmWithReleases()
	firstK8s := &harness.FakeK8sExecutor{}
	_, firstProvider := harness.NewFakeRunAIClientProvider(firstHelm)
	eventsPath := filepath.Join(t.TempDir(), "first-events.ndjson")
	cfg, opts := artifactInstallConfig(bundle, workingDir, eventsPath)
	firstDeps := orchestrator2.Deps{
		K8sExecutor:         firstK8s,
		HelmInstaller:       firstHelm,
		RunAIClientProvider: firstProvider,
		ClusterProbe:        okProbe(),
		ImagePusher:         &harness.FakeImagePusher{},
	}
	term, _, _ := newTestTerminal()
	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), firstDeps, term, execOpts()))

	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	cfg.Airgap.BundlePath = ""
	cfg.Airgap.SkipPrepare = true
	opts.EventsFile = filepath.Join(t.TempDir(), "loaded-events.ndjson")
	deps := orchestrator2.Deps{
		K8sExecutor:         k8s,
		HelmInstaller:       helm,
		RunAIClientProvider: provider,
		ClusterProbe:        okProbe(),
	}
	term, _, _ = newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	require.Len(t, helm.CallsByRelease("runai-backend"), 1)
	require.Len(t, helm.CallsByRelease("runai-cluster"), 1)
	events := readEvents(t, opts.EventsFile)
	assert.Nil(t, findEvent(events, "step_started", "extract-bundle"))
	assert.Nil(t, findEvent(events, "step_started", "push-images"))
	assert.Nil(t, findEvent(events, "step_started", "generate-values"))
	assert.True(t, hasPipelineEvent(events, "pipeline_succeeded", "runai"))
}

func TestAirgapInstall_LoadFailureStopsInstallation(t *testing.T) {
	lock := artifactInstallLock()
	bundle := harness.CreateLoadBundle(t, lock)
	workingDir := t.TempDir()
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	pusher := &harness.FakeImagePusher{PushErrors: map[string]error{
		airgapLoadDestination + "/runai-control-plane/assets-service:2.26.13": errors.New("push failed"),
	}}
	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := artifactInstallConfig(bundle, workingDir, eventsPath)
	deps := orchestrator2.Deps{
		K8sExecutor:         k8s,
		HelmInstaller:       helm,
		RunAIClientProvider: provider,
		ClusterProbe:        okProbe(),
		ImagePusher:         pusher,
	}
	term, _, _ := newTestTerminal()

	require.Error(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))
	assert.Empty(t, helm.Calls)
	events := readEvents(t, eventsPath)
	require.NotNil(t, findEvent(events, "step_failed", "push-images"))
	assert.Nil(t, findEvent(events, "step_started", "install-control-plane-chart"))
}

func artifactInstallConfig(bundle, workingDir, eventsPath string) (*installconfig.Config, installconfig.RunOptions) {
	cfg := newInstallRunConfig()
	cfg.Install.Connectivity = string(spec.ConnectivityAirGapped)
	cfg.Airgap = &installconfig.Airgap{BundlePath: bundle, WorkingDir: workingDir}
	cfg.ControlPlane.Version = "3.0.0"
	cfg.Cluster.Version = "3.0.0"
	cfg.Prerequisites = &installconfig.Prerequisites{Skip: true}
	cfg.Features = &installconfig.Features{Skip: true}
	opts := baseRunOptions()
	opts.EventsFile = eventsPath
	opts.Registry = &spec.Registry{Host: airgapLoadDestination, Username: "admin", Password: "Harbor12345"}
	return cfg, opts
}

func artifactInstallLock() *artifactlock.ArtifactLock {
	lock := harness.TestLock()
	lock.Metadata.Version = "3.0.0"
	lock.Spec.ImageLocks[0].Version = "3.0.0"
	lock.Spec.Charts[0].Version = "3.0.0"
	lock.Spec.Charts[0].BundlePath = "charts/control-plane-3.0.0.tgz"
	lock.Spec.Charts[1].Version = "3.0.0"
	lock.Spec.Charts[1].Source = "https://helm.ngc.nvidia.com/nvidia/runai-oss/charts/runai-cluster-3.0.0.tgz"
	lock.Spec.Charts[1].BundlePath = "charts/runai-cluster-3.0.0.tgz"
	return lock
}
