/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package upgrade_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/clients/runai/cluster"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/upgrade"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/artifactlock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newBaseCfg() *config.UpgradeConfig {
	return &config.UpgradeConfig{
		ExecutionTimeout:    5 * time.Minute,
		StepTimeout:         2 * time.Minute,
		ControlPlaneVersion: "2.20.0",
		ClusterVersion:      "2.20.0",
		UpgradeControlPlane: true,
		UpgradeCluster:      true,
	}
}

func TestUpgrade_BothCPAndCluster_AutoDetectedParams(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	cfg := newBaseCfg()
	deps := upgrade.Deps{
		HelmInstaller:    fakeHelm,
		ClusterLookup:    harness.NewFakeClusterLookup(),
		HelmValuesReader: harness.HelmValues,
	}

	err := orchestrator.RunUpgrade(cfg, deps)
	require.NoError(t, err)

	assert.Equal(t, "test.run.ai", cfg.FQDN)
	assert.Equal(t, "uuid-abc", cfg.ClusterUUID)

	require.Len(t, fakeHelm.Calls, 2)

	cpCall := fakeHelm.Calls[0]
	assert.Equal(t, step.CPReleaseName, cpCall.ReleaseName)
	assert.Equal(t, "2.20.0", cpCall.Version)
	assert.True(t, harness.ContainsValue(cpCall.Values, "global.domain=test.run.ai"))

	clusterCall := fakeHelm.Calls[1]
	assert.Equal(t, step.ClusterReleaseName, clusterCall.ReleaseName)
	assert.Equal(t, "2.20.0", clusterCall.Version)
	assert.True(t, harness.ContainsValue(clusterCall.Values, "controlPlane.url=test.run.ai"))
	assert.True(t, harness.ContainsValue(clusterCall.Values, "controlPlane.clientSecret=s3cr3t"))
	assert.True(t, harness.ContainsValue(clusterCall.Values, "cluster.uid=uuid-abc"))
}

func TestUpgrade_CPOnly(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	cfg := newBaseCfg()
	cfg.UpgradeCluster = false
	cfg.FQDN = "explicit.run.ai"
	deps := upgrade.Deps{
		HelmInstaller:    fakeHelm,
		ClusterLookup:    harness.NewFakeClusterLookup(),
		HelmValuesReader: harness.HelmValues,
	}

	err := orchestrator.RunUpgrade(cfg, deps)
	require.NoError(t, err)

	require.Len(t, fakeHelm.Calls, 1)
	assert.Equal(t, step.CPReleaseName, fakeHelm.Calls[0].ReleaseName)
	assert.True(t, harness.ContainsValue(fakeHelm.Calls[0].Values, "global.domain=explicit.run.ai"))
}

func TestUpgrade_ClusterOnly(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	cfg := newBaseCfg()
	cfg.UpgradeControlPlane = false
	cfg.FQDN = "explicit.run.ai"
	cfg.ClusterUUID = "uuid-abc"
	deps := upgrade.Deps{
		HelmInstaller:    fakeHelm,
		ClusterLookup:    harness.NewFakeClusterLookup(),
		HelmValuesReader: harness.HelmValues,
	}

	err := orchestrator.RunUpgrade(cfg, deps)
	require.NoError(t, err)

	require.Len(t, fakeHelm.Calls, 1)
	assert.Equal(t, step.ClusterReleaseName, fakeHelm.Calls[0].ReleaseName)
	assert.True(t, harness.ContainsValue(fakeHelm.Calls[0].Values, "cluster.uid=uuid-abc"))
	assert.True(t, harness.ContainsValue(fakeHelm.Calls[0].Values, "controlPlane.clientSecret=s3cr3t"))
}

func TestUpgrade_ClusterWithBootstrapToken(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	cfg := newBaseCfg()
	cfg.UpgradeControlPlane = false
	cfg.FQDN = "explicit.run.ai"
	cfg.ClusterUUID = "uuid-abc"
	deps := upgrade.Deps{
		HelmInstaller: fakeHelm,
		ClusterLookup: &harness.FakeClusterLookup{
			InstallInfo: &cluster.ClusterInstallationInfoResponse{
				BootstrapToken: "bootstrap-token",
			},
		},
		HelmValuesReader: harness.HelmValues,
	}

	err := orchestrator.RunUpgrade(cfg, deps)
	require.NoError(t, err)

	require.Len(t, fakeHelm.Calls, 1)
	assert.True(t, harness.ContainsValue(fakeHelm.Calls[0].Values, "controlPlane.bootstrapToken=bootstrap-token"))
	assert.False(t, harness.ContainsValue(fakeHelm.Calls[0].Values, "controlPlane.clientSecret="))
}

func TestUpgrade_ClusterUUIDUnresolvable(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	cfg := newBaseCfg()
	cfg.UpgradeControlPlane = false
	cfg.FQDN = "test.run.ai"
	noUIDReader := func(release, ns string) (map[string]interface{}, error) {
		if release == step.ClusterReleaseName {
			return map[string]interface{}{
				"cluster": map[string]interface{}{},
			}, nil
		}
		return harness.HelmValues(release, ns)
	}
	deps := upgrade.Deps{
		HelmInstaller:    fakeHelm,
		ClusterLookup:    harness.NewFakeClusterLookup(),
		HelmValuesReader: noUIDReader,
	}

	err := orchestrator.RunUpgrade(cfg, deps)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "--cluster-uuid not provided"))
}

func TestUpgrade_CustomCAValuesCarriedForward(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	cfg := newBaseCfg()
	deps := upgrade.Deps{
		HelmInstaller:    fakeHelm,
		ClusterLookup:    harness.NewFakeClusterLookup(),
		HelmValuesReader: harness.HelmValuesWithCustomCA,
	}

	err := orchestrator.RunUpgrade(cfg, deps)
	require.NoError(t, err)

	require.Len(t, fakeHelm.Calls, 2)

	cpCall := fakeHelm.Calls[0]
	assert.True(t, harness.ContainsValue(cpCall.Values, "global.customCA.enabled=true"))
	assert.True(t, harness.ContainsValue(cpCall.Values, "global.customCA.secret.name=runai-ca-cert"))
	assert.True(t, harness.ContainsValue(cpCall.Values, "global.customCA.secret.key=runai-ca.pem"))

	clusterCall := fakeHelm.Calls[1]
	assert.True(t, harness.ContainsValue(clusterCall.Values, "global.customCA.enabled=true"))
	assert.True(t, harness.ContainsValue(clusterCall.Values, "global.customCA.secret.name=runai-ca-cert"))
	assert.True(t, harness.ContainsValue(clusterCall.Values, "global.customCA.secret.key=runai-ca.pem"))
}

func TestUpgrade_NoCustomCA_NoCarryForward(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	cfg := newBaseCfg()
	deps := upgrade.Deps{
		HelmInstaller:    fakeHelm,
		ClusterLookup:    harness.NewFakeClusterLookup(),
		HelmValuesReader: harness.HelmValues,
	}

	err := orchestrator.RunUpgrade(cfg, deps)
	require.NoError(t, err)

	require.Len(t, fakeHelm.Calls, 2)
	for _, call := range fakeHelm.Calls {
		for _, v := range call.Values {
			assert.False(t, strings.HasPrefix(v, "global.customCA."), "unexpected carry-forward value: %s", v)
		}
	}
}

func TestUpgrade_OpenShiftDistributionCarriedForward(t *testing.T) {
	harness.UseFastPollInterval(t)
	fakeHelm := &harness.FakeHelmInstaller{}
	cfg := newBaseCfg()
	deps := upgrade.Deps{
		HelmInstaller:    fakeHelm,
		ClusterLookup:    harness.NewFakeClusterLookup(),
		HelmValuesReader: harness.HelmValuesOpenShift,
	}

	err := orchestrator.RunUpgrade(cfg, deps)
	require.NoError(t, err)

	require.Len(t, fakeHelm.Calls, 2)

	cpCall := fakeHelm.Calls[0]
	assert.True(t, harness.ContainsValue(cpCall.Values, "global.config.kubernetesDistribution=openshift"))

	clusterCall := fakeHelm.Calls[1]
	assert.True(t, harness.ContainsValue(clusterCall.Values, "global.config.kubernetesDistribution=openshift"))
}

func TestUpgrade_DefaultDistribution_NotCarriedForward(t *testing.T) {
	harness.UseFastPollInterval(t)
	fakeHelm := &harness.FakeHelmInstaller{}
	cfg := newBaseCfg()
	deps := upgrade.Deps{
		HelmInstaller:    fakeHelm,
		ClusterLookup:    harness.NewFakeClusterLookup(),
		HelmValuesReader: harness.HelmValues,
	}

	err := orchestrator.RunUpgrade(cfg, deps)
	require.NoError(t, err)

	require.Len(t, fakeHelm.Calls, 2)
	for _, call := range fakeHelm.Calls {
		for _, v := range call.Values {
			assert.False(t, strings.HasPrefix(v, "global.config.kubernetesDistribution"), "unexpected carry-forward value: %s", v)
		}
	}
}

func TestUpgrade_FQDNUnresolvable(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	cfg := newBaseCfg()
	failReader := func(release, ns string) (map[string]interface{}, error) {
		return nil, fmt.Errorf("no release found")
	}
	deps := upgrade.Deps{
		HelmInstaller:    fakeHelm,
		ClusterLookup:    harness.NewFakeClusterLookup(),
		HelmValuesReader: failReader,
	}

	err := orchestrator.RunUpgrade(cfg, deps)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "--fqdn not provided"))
}

func newAirgapUpgradeCfg(tarballPath string) *config.UpgradeConfig {
	return &config.UpgradeConfig{
		ExecutionTimeout:    5 * time.Minute,
		StepTimeout:         2 * time.Minute,
		ControlPlaneVersion: "2.21.0",
		ClusterVersion:      "2.21.0",
		UpgradeControlPlane: true,
		UpgradeCluster:      true,
		AirgapPackagePath:   tarballPath,
		RegistryURL:         "registry.local:5000",
		RegistryUsername:    "admin",
		RegistryPassword:    "pass",
	}
}

// bothChartsLock is an artifact lock covering the two charts an upgrade
// installs, as a bundle built for a full environment does.
func bothChartsLock() *artifactlock.ArtifactLock {
	lock := harness.TestLock()
	lock.Metadata.Version = "3.0.0"
	lock.Spec.Charts = []artifactlock.Chart{
		{
			Name:       step.CPChartName,
			Version:    "3.0.0",
			BundlePath: "charts/control-plane-3.0.0.tgz",
			Digest:     harness.ChartDigest(step.CPChartName),
		},
		{
			Name:       step.ClusterChartName,
			Version:    "3.0.0",
			BundlePath: "charts/runai-cluster-3.0.0.tgz",
			Digest:     harness.ChartDigest(step.ClusterChartName),
		},
	}
	return lock
}

func newLoadedBundleUpgradeCfg(workingDir string) *config.UpgradeConfig {
	cfg := newAirgapUpgradeCfg("")
	cfg.AirgapPackagePath = ""
	cfg.AirgapWorkingDir = workingDir
	cfg.ControlPlaneVersion = "3.0.0"
	cfg.ClusterVersion = "3.0.0"
	return cfg
}

// A bundle prepared by `airgap load` is upgraded from the charts its lock pins
// and the values files load generated, with nothing to extract or push.
func TestAirgapUpgrade_FromLoadedBundle(t *testing.T) {
	harness.UseFastPollInterval(t)
	fakeHelm := &harness.FakeHelmInstaller{}
	fakeK8s := &harness.FakeK8sExecutor{}
	workingDir := harness.CreateLoadedBundle(t, bothChartsLock())
	cfg := newLoadedBundleUpgradeCfg(workingDir)
	deps := upgrade.Deps{
		HelmInstaller:    fakeHelm,
		ClusterLookup:    harness.NewFakeClusterLookup(),
		HelmValuesReader: harness.HelmValues,
		K8sExecutor:      fakeK8s,
	}

	err := orchestrator.RunUpgrade(cfg, deps)
	require.NoError(t, err)

	assert.Empty(t, cfg.AirgapExtractedDir, "a loaded bundle is not extracted again")
	require.Len(t, fakeHelm.Calls, 2)

	cpCall := fakeHelm.Calls[0]
	assert.Equal(t, step.CPReleaseName, cpCall.ReleaseName)
	assert.Equal(t, filepath.Join(workingDir, "bundle", "charts", "control-plane-3.0.0.tgz"), cpCall.ChartName)
	assert.Equal(t, []string{filepath.Join(workingDir, "control-plane-airgap-values.yaml")}, cpCall.ValueFiles)
	assert.True(t, harness.ContainsValue(cpCall.Values, "global.domain=test.run.ai"))

	clusterCall := fakeHelm.Calls[1]
	assert.Equal(t, step.ClusterReleaseName, clusterCall.ReleaseName)
	assert.Equal(t, filepath.Join(workingDir, "bundle", "charts", "runai-cluster-3.0.0.tgz"), clusterCall.ChartName)
	assert.Equal(t, []string{filepath.Join(workingDir, "cluster-airgap-values.yaml")}, clusterCall.ValueFiles)
	assert.True(t, harness.ContainsValue(clusterCall.Values, "controlPlane.url=test.run.ai"))
	assert.True(t, harness.ContainsValue(clusterCall.Values, "controlPlane.clientSecret=s3cr3t"))
	assert.True(t, harness.ContainsValue(clusterCall.Values, "cluster.uid=uuid-abc"))
	assert.True(t, harness.ContainsValue(clusterCall.Values, "cluster.url=test.run.ai"))
	assert.True(t, harness.ContainsValue(clusterCall.Values, "global.additionalImagePullSecrets[0]=runai-reg-creds"))

	secretCalls := fakeK8s.CallsByVerb("create")
	require.Len(t, secretCalls, 2)
	assert.Equal(t, step.CPNamespace, secretCalls[0].Namespace)
	assert.Equal(t, "runai-reg-creds", secretCalls[0].Name)
	assert.Equal(t, step.ClusterNamespace, secretCalls[1].Namespace)
	assert.Equal(t, "runai-reg-creds", secretCalls[1].Name)
}

func TestAirgapUpgrade_FromLoadedBundleUpdatesExistingRegistrySecrets(t *testing.T) {
	harness.UseFastPollInterval(t)
	workingDir := harness.CreateLoadedBundle(t, bothChartsLock())
	fakeK8s := &harness.FakeK8sExecutor{
		Existing: map[client.ObjectKey]bool{
			{Namespace: step.CPNamespace, Name: "runai-reg-creds"}:      true,
			{Namespace: step.ClusterNamespace, Name: "runai-reg-creds"}: true,
		},
	}
	cfg := newLoadedBundleUpgradeCfg(workingDir)
	deps := upgrade.Deps{
		HelmInstaller:    &harness.FakeHelmInstaller{},
		ClusterLookup:    harness.NewFakeClusterLookup(),
		HelmValuesReader: harness.HelmValues,
		K8sExecutor:      fakeK8s,
	}

	require.NoError(t, orchestrator.RunUpgrade(cfg, deps))

	secretCalls := fakeK8s.CallsByVerb("patch")
	require.Len(t, secretCalls, 2)
	assert.Equal(t, step.CPNamespace, secretCalls[0].Namespace)
	assert.Equal(t, step.ClusterNamespace, secretCalls[1].Namespace)
}

// A values directory kept apart from the working directory is honored, which is
// how `airgap load --config` can be pointed at a separate values dir.
func TestAirgapUpgrade_FromLoadedBundleWithSeparateValuesDir(t *testing.T) {
	harness.UseFastPollInterval(t)
	fakeHelm := &harness.FakeHelmInstaller{}
	workingDir := harness.CreateLoadedBundle(t, bothChartsLock())
	valuesDir := t.TempDir()
	for _, name := range []string{"control-plane-airgap-values.yaml", "cluster-airgap-values.yaml"} {
		require.NoError(t, os.Rename(filepath.Join(workingDir, name), filepath.Join(valuesDir, name)))
	}
	cfg := newLoadedBundleUpgradeCfg(workingDir)
	cfg.AirgapValuesDir = valuesDir
	deps := upgrade.Deps{
		HelmInstaller:    fakeHelm,
		ClusterLookup:    harness.NewFakeClusterLookup(),
		HelmValuesReader: harness.HelmValues,
		K8sExecutor:      &harness.FakeK8sExecutor{},
	}

	require.NoError(t, orchestrator.RunUpgrade(cfg, deps))

	require.Len(t, fakeHelm.Calls, 2)
	assert.Equal(t, []string{filepath.Join(valuesDir, "control-plane-airgap-values.yaml")}, fakeHelm.Calls[0].ValueFiles)
	assert.Equal(t, []string{filepath.Join(valuesDir, "cluster-airgap-values.yaml")}, fakeHelm.Calls[1].ValueFiles)
}

// A chart the bundle's lock does not pin fails the upgrade rather than falling
// back to some other path.
func TestAirgapUpgrade_FromLoadedBundleMissingChart(t *testing.T) {
	harness.UseFastPollInterval(t)
	lock := bothChartsLock()
	lock.Spec.Charts = lock.Spec.Charts[:1]
	workingDir := harness.CreateLoadedBundle(t, lock)
	cfg := newLoadedBundleUpgradeCfg(workingDir)
	cfg.UpgradeControlPlane = false
	fakeHelm := &harness.FakeHelmInstaller{}
	deps := upgrade.Deps{
		HelmInstaller:    fakeHelm,
		ClusterLookup:    harness.NewFakeClusterLookup(),
		HelmValuesReader: harness.HelmValues,
		K8sExecutor:      &harness.FakeK8sExecutor{},
	}

	err := orchestrator.RunUpgrade(cfg, deps)

	require.Error(t, err)
	assert.Empty(t, fakeHelm.Calls, "nothing is installed when the lock does not pin the chart")
}

func TestAirgapUpgrade_BothCPAndCluster(t *testing.T) {
	harness.UseFastPollInterval(t)
	fakeHelm := &harness.FakeHelmInstaller{}
	fakeK8s := &harness.FakeK8sExecutor{}
	tarball := harness.CreateTestTarball(t)
	cfg := newAirgapUpgradeCfg(tarball)
	deps := upgrade.Deps{
		HelmInstaller:    fakeHelm,
		ClusterLookup:    harness.NewFakeClusterLookup(),
		HelmValuesReader: harness.HelmValues,
		K8sExecutor:      fakeK8s,
	}

	err := orchestrator.RunUpgrade(cfg, deps)
	require.NoError(t, err)

	assert.Equal(t, "test.run.ai", cfg.FQDN)
	assert.Equal(t, "uuid-abc", cfg.ClusterUUID)
	assert.NotEmpty(t, cfg.AirgapExtractedDir)

	require.Len(t, fakeHelm.Calls, 2)

	cpCall := fakeHelm.Calls[0]
	assert.Equal(t, step.CPReleaseName, cpCall.ReleaseName)
	assert.Equal(t, step.CPNamespace, cpCall.Namespace)
	assert.True(t, harness.ContainsValue(cpCall.Values, "global.domain=test.run.ai"))
	assert.True(t, strings.HasSuffix(cpCall.ChartName, "control-plane-2.20.0.tgz"))
	require.Len(t, cpCall.ValueFiles, 1)
	assert.True(t, strings.HasSuffix(cpCall.ValueFiles[0], "custom-env.yaml"))

	clusterCall := fakeHelm.Calls[1]
	assert.Equal(t, step.ClusterReleaseName, clusterCall.ReleaseName)
	assert.Equal(t, step.ClusterNamespace, clusterCall.Namespace)
	assert.True(t, harness.ContainsValue(clusterCall.Values, "controlPlane.url=test.run.ai"))
	assert.True(t, harness.ContainsValue(clusterCall.Values, "controlPlane.clientSecret=s3cr3t"))
	assert.True(t, harness.ContainsValue(clusterCall.Values, "cluster.uid=uuid-abc"))
	assert.True(t, harness.ContainsValue(clusterCall.Values, "cluster.url=test.run.ai"))
	assert.True(t, harness.ContainsValue(clusterCall.Values, "global.additionalImagePullSecrets[0]=runai-reg-creds"))
	assert.True(t, strings.HasSuffix(clusterCall.ChartName, "runai-cluster-2.20.0.tgz"))
	require.Len(t, clusterCall.ValueFiles, 1)
	assert.True(t, strings.HasSuffix(clusterCall.ValueFiles[0], "custom-env.yaml"))

	secretCalls := fakeK8s.CallsByVerb("create")
	require.Len(t, secretCalls, 2)
	assert.Equal(t, step.CPNamespace, secretCalls[0].Namespace)
	assert.Equal(t, step.ClusterNamespace, secretCalls[1].Namespace)
}

func TestAirgapUpgrade_CPOnly(t *testing.T) {
	harness.UseFastPollInterval(t)
	fakeHelm := &harness.FakeHelmInstaller{}
	fakeK8s := &harness.FakeK8sExecutor{}
	tarball := harness.CreateTestTarball(t)
	cfg := newAirgapUpgradeCfg(tarball)
	cfg.UpgradeCluster = false
	cfg.FQDN = "explicit.run.ai"
	deps := upgrade.Deps{
		HelmInstaller:    fakeHelm,
		ClusterLookup:    harness.NewFakeClusterLookup(),
		HelmValuesReader: harness.HelmValues,
		K8sExecutor:      fakeK8s,
	}

	err := orchestrator.RunUpgrade(cfg, deps)
	require.NoError(t, err)

	require.Len(t, fakeHelm.Calls, 1)
	cpCall := fakeHelm.Calls[0]
	assert.Equal(t, step.CPReleaseName, cpCall.ReleaseName)
	assert.True(t, harness.ContainsValue(cpCall.Values, "global.domain=explicit.run.ai"))
	assert.True(t, strings.HasSuffix(cpCall.ChartName, "control-plane-2.20.0.tgz"))
	require.Len(t, cpCall.ValueFiles, 1)
}

func TestAirgapUpgrade_ClusterOnly(t *testing.T) {
	harness.UseFastPollInterval(t)
	fakeHelm := &harness.FakeHelmInstaller{}
	fakeK8s := &harness.FakeK8sExecutor{}
	tarball := harness.CreateTestTarball(t)
	cfg := newAirgapUpgradeCfg(tarball)
	cfg.UpgradeControlPlane = false
	cfg.FQDN = "explicit.run.ai"
	cfg.ClusterUUID = "uuid-abc"
	deps := upgrade.Deps{
		HelmInstaller:    fakeHelm,
		ClusterLookup:    harness.NewFakeClusterLookup(),
		HelmValuesReader: harness.HelmValues,
		K8sExecutor:      fakeK8s,
	}

	err := orchestrator.RunUpgrade(cfg, deps)
	require.NoError(t, err)

	require.Len(t, fakeHelm.Calls, 1)
	clusterCall := fakeHelm.Calls[0]
	assert.Equal(t, step.ClusterReleaseName, clusterCall.ReleaseName)
	assert.True(t, harness.ContainsValue(clusterCall.Values, "cluster.uid=uuid-abc"))
	assert.True(t, harness.ContainsValue(clusterCall.Values, "controlPlane.clientSecret=s3cr3t"))
	assert.True(t, harness.ContainsValue(clusterCall.Values, "global.additionalImagePullSecrets[0]=runai-reg-creds"))
	assert.True(t, strings.HasSuffix(clusterCall.ChartName, "runai-cluster-2.20.0.tgz"))
}

func TestAirgapUpgrade_CustomCACarriedForward(t *testing.T) {
	harness.UseFastPollInterval(t)
	fakeHelm := &harness.FakeHelmInstaller{}
	fakeK8s := &harness.FakeK8sExecutor{}
	tarball := harness.CreateTestTarball(t)
	cfg := newAirgapUpgradeCfg(tarball)
	deps := upgrade.Deps{
		HelmInstaller:    fakeHelm,
		ClusterLookup:    harness.NewFakeClusterLookup(),
		HelmValuesReader: harness.HelmValuesWithCustomCA,
		K8sExecutor:      fakeK8s,
	}

	err := orchestrator.RunUpgrade(cfg, deps)
	require.NoError(t, err)

	require.Len(t, fakeHelm.Calls, 2)

	cpCall := fakeHelm.Calls[0]
	assert.True(t, harness.ContainsValue(cpCall.Values, "global.customCA.enabled=true"))
	assert.True(t, harness.ContainsValue(cpCall.Values, "global.customCA.secret.name=runai-ca-cert"))
	assert.True(t, harness.ContainsValue(cpCall.Values, "global.customCA.secret.key=runai-ca.pem"))

	clusterCall := fakeHelm.Calls[1]
	assert.True(t, harness.ContainsValue(clusterCall.Values, "global.customCA.enabled=true"))
	assert.True(t, harness.ContainsValue(clusterCall.Values, "global.customCA.secret.name=runai-ca-cert"))
}

func TestAirgapUpgrade_BadTarball_Fails(t *testing.T) {
	harness.UseFastPollInterval(t)
	fakeHelm := &harness.FakeHelmInstaller{}
	cfg := newAirgapUpgradeCfg(filepath.Join(t.TempDir(), "nonexistent.tar.gz"))
	cfg.FQDN = "test.run.ai"
	cfg.ClusterUUID = "uuid-abc"
	deps := upgrade.Deps{
		HelmInstaller:    fakeHelm,
		ClusterLookup:    harness.NewFakeClusterLookup(),
		HelmValuesReader: harness.HelmValues,
	}

	err := orchestrator.RunUpgrade(cfg, deps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Airgap Prepare")
	assert.Len(t, fakeHelm.Calls, 0)
}

func TestAirgapUpgrade_UnreadableTarball_Fails(t *testing.T) {
	harness.UseFastPollInterval(t)
	fakeHelm := &harness.FakeHelmInstaller{}

	dir := t.TempDir()
	tarball := filepath.Join(dir, "locked.tar.gz")
	require.NoError(t, os.WriteFile(tarball, []byte("data"), 0o644))
	require.NoError(t, os.Chmod(dir, 0o000))
	t.Cleanup(func() { os.Chmod(dir, 0o755) })

	cfg := newAirgapUpgradeCfg(tarball)
	cfg.FQDN = "test.run.ai"
	cfg.ClusterUUID = "uuid-abc"
	deps := upgrade.Deps{
		HelmInstaller:    fakeHelm,
		ClusterLookup:    harness.NewFakeClusterLookup(),
		HelmValuesReader: harness.HelmValues,
	}

	err := orchestrator.RunUpgrade(cfg, deps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Airgap Prepare")
	assert.Len(t, fakeHelm.Calls, 0)
}

func TestAirgapUpgrade_IsAirgapToggle(t *testing.T) {
	cfg := newAirgapUpgradeCfg("/some/path.tar.gz")
	assert.True(t, cfg.IsAirgap())

	cfg.AirgapPackagePath = ""
	assert.False(t, cfg.IsAirgap())
}
