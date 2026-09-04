/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install_test

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	k8smocks "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newAirgapRunaiData(tarballPath string) *model.RunAIInstallData {
	return &model.RunAIInstallData{
		InstallValues: model.InstallValues{
			Inputs: map[configinputs.InstallInputType]string{
				configinputs.InstallInputFQDN:              "test.runai.local",
				configinputs.InstallInputClusterName:       "test-cluster",
				configinputs.InstallInputAirgapPackagePath: tarballPath,
				configinputs.InstallInputRegistryURL:       "registry.local:5000",
			},
		},
		Config: &config.InstallConfig{
			StepTimeout:        2 * time.Minute,
			ExecutionTimeout:   5 * time.Minute,
			RunaiInstallFormat: config.Airgap,
			DevMode:            true,
		},
	}
}

func newAirgapTestDeps(t *testing.T, fakeHelm *harness.FakeHelmInstaller) (*harness.FakeRunAIAPIClient, orchestrator.InstallDeps) {
	t.Helper()
	harness.UseFastPollInterval(t)
	fakeClient := fake.NewClientBuilder().Build()
	mockK8s := k8smocks.NewMockK8sClients(t)
	mockK8s.On("GetClient").Return(fakeClient).Maybe()
	mockK8s.On("GetConfig").Return(nil).Maybe()
	mockK8s.On("GetContextName").Return("test-context").Maybe()
	apiClient, provider := harness.NewFakeRunAIClientProvider(fakeHelm)
	return apiClient, orchestrator.InstallDeps{
		K8sClient:           mockK8s,
		HelmInstaller:       fakeHelm,
		RunAIClientProvider: provider,
	}
}

func TestAirgapInstall_FullFlow(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	tarball := harness.CreateTestTarball(t)

	runaiData := newAirgapRunaiData(tarball)
	apiClient, deps := newAirgapTestDeps(t, fakeHelm)

	err := orchestrator.RunInstall(runaiData, deps)
	require.NoError(t, err)

	assert.True(t, apiClient.ClusterCreated, "CreateCluster should have been called")
	require.Len(t, fakeHelm.Calls, 2)

	cpCall := fakeHelm.Calls[0]
	assert.Equal(t, step.CPReleaseName, cpCall.ReleaseName)
	assert.Equal(t, step.CPNamespace, cpCall.Namespace)
	assert.True(t, harness.ContainsValue(cpCall.Values, "global.domain=test.runai.local"))
	assert.True(t, harness.ContainsValue(cpCall.Values, "global.customCA.enabled=true"))
	assert.True(t, strings.HasSuffix(cpCall.ChartName, "control-plane-2.20.0.tgz"))

	clusterCall := fakeHelm.Calls[1]
	assert.Equal(t, step.ClusterReleaseName, clusterCall.ReleaseName)
	assert.Equal(t, step.ClusterNamespace, clusterCall.Namespace)
	assert.True(t, harness.ContainsValue(clusterCall.Values, "controlPlane.url=test.runai.local"))
	assert.True(t, harness.ContainsValue(clusterCall.Values, "controlPlane.clientSecret=s3cr3t-airgap"))
	assert.True(t, harness.ContainsValue(clusterCall.Values, "cluster.uid=uuid-airgap-test"))
	assert.True(t, harness.ContainsValue(clusterCall.Values, "cluster.url=test.runai.local"))
	assert.True(t, harness.ContainsValue(clusterCall.Values, "global.customCA.enabled=true"))
	assert.True(t, strings.HasSuffix(clusterCall.ChartName, "runai-cluster-2.20.0.tgz"))
}

func TestAirgapInstall_UsesBootstrapToken(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	tarball := harness.CreateTestTarball(t)
	runaiData := newAirgapRunaiData(tarball)
	apiClient, deps := newAirgapTestDeps(t, fakeHelm)
	apiClient.BootstrapToken = "bootstrap-token"

	err := orchestrator.RunInstall(runaiData, deps)
	require.NoError(t, err)

	require.Len(t, fakeHelm.Calls, 2)
	clusterCall := fakeHelm.Calls[1]
	assert.True(t, harness.ContainsValue(clusterCall.Values, "controlPlane.bootstrapToken=bootstrap-token"))
	assert.False(t, harness.ContainsValue(clusterCall.Values, "controlPlane.clientSecret=s3cr3t-airgap"))
}

func TestAirgapInstall_ValuesIncludeSubdomainSupport(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	tarball := harness.CreateTestTarball(t)

	runaiData := newAirgapRunaiData(tarball)
	_, deps := newAirgapTestDeps(t, fakeHelm)

	err := orchestrator.RunInstall(runaiData, deps)
	require.NoError(t, err)

	require.Len(t, fakeHelm.Calls, 2)
	clusterCall := fakeHelm.Calls[1]
	assert.True(t, harness.ContainsValue(clusterCall.Values, "clusterConfig.global.subdomainSupport=false"))
}

func TestAirgapInstall_BadTarball_Fails(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	badPath := filepath.Join(t.TempDir(), "nonexistent.tar.gz")

	runaiData := newAirgapRunaiData(badPath)
	_, deps := newAirgapTestDeps(t, fakeHelm)

	err := orchestrator.RunInstall(runaiData, deps)
	require.Error(t, err)
	assert.Len(t, fakeHelm.Calls, 0)
}
