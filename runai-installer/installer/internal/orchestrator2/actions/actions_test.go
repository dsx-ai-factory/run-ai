/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"testing"

	runaicluster "github.com/dsx-ai-factory/run-ai/runai-installer/clients/runai/cluster"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type fakeAPI struct {
	clusters     []runaicluster.DisplayedCluster
	getErr       error
	created      []runaicluster.ClusterCreationRequest
	createStatus int
	clientSecret string
}

func (f *fakeAPI) GetClusters(context.Context) ([]runaicluster.DisplayedCluster, int, error) {
	if f.getErr != nil {
		return nil, 0, f.getErr
	}
	return f.clusters, 200, nil
}

func (f *fakeAPI) CreateCluster(_ context.Context, req runaicluster.ClusterCreationRequest) (*runaicluster.DisplayedCluster, int, error) {
	f.created = append(f.created, req)
	status := f.createStatus
	if status == 0 {
		status = 200
	}
	dc := runaicluster.DisplayedCluster{Name: req.Name, Uuid: "new-uuid"}
	return &dc, status, nil
}

func (f *fakeAPI) GetClusterByUUID(context.Context, string) (*runaicluster.DisplayedCluster, int, error) {
	return nil, 200, nil
}

func (f *fakeAPI) GetClusterInstallInfoByUUID(context.Context, string, string, string) (*runaicluster.ClusterInstallationInfoResponse, int, error) {
	return runaicluster.NewClusterInstallationInfoResponse("", "", "", f.clientSecret), 200, nil
}

func providerFor(api step.RunAIAPIClient) step.RunAIAPIClientProvider {
	return func(_, _ string) (step.RunAIAPIClient, error) { return api, nil }
}

func execOK(t *testing.T, a orchestrator2.PerformAction, deps orchestrator2.Deps) {
	t.Helper()
	_, err := a.Execute(context.Background(), deps)
	require.NoError(t, err)
}

func TestK8sApply(t *testing.T) {
	fakeK8s := &harness.FakeK8sExecutor{}
	deps := orchestrator2.Deps{K8sExecutor: fakeK8s}
	ns := &corev1.Namespace{
		TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "runai"},
	}

	execOK(t, K8sApply{Object: ns}, deps)

	applies := fakeK8s.CallsByVerb("apply")
	require.Len(t, applies, 1)
	assert.Equal(t, "runai", applies[0].Name)
}

func TestHelmInstallLocal(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	deps := orchestrator2.Deps{HelmInstaller: fakeHelm}
	act := HelmInstallLocal{
		ChartPath:   "/pkg/charts/cp.tgz",
		ReleaseName: "runai-backend",
		Namespace:   "runai-backend",
		Values:      []string{"a=b"},
		ValueFiles:  []string{"/pkg/custom-env.yaml"},
	}

	execOK(t, act, deps)

	calls := fakeHelm.CallsByRelease("runai-backend")
	require.Len(t, calls, 1)
	assert.Equal(t, "/pkg/charts/cp.tgz", calls[0].ChartName)
	assert.Equal(t, []string{"/pkg/custom-env.yaml"}, calls[0].ValueFiles)
}

func TestHelmInstallLocal_ResolvesChartPathAtExecute(t *testing.T) {
	workingDir := t.TempDir()
	act := HelmInstallLocal{
		WorkingDir:  workingDir,
		ChartPrefix: "control-plane",
		ReleaseName: "runai-backend",
		Namespace:   "runai-backend",
	}

	chartsDir := filepath.Join(workingDir, "charts")
	require.NoError(t, os.MkdirAll(chartsDir, 0755))
	chartPath := filepath.Join(chartsDir, "control-plane-2.20.0.tgz")
	require.NoError(t, os.WriteFile(chartPath, []byte("chart"), 0644))

	fakeHelm := &harness.FakeHelmInstaller{}
	deps := orchestrator2.Deps{HelmInstaller: fakeHelm}
	execOK(t, act, deps)

	calls := fakeHelm.CallsByRelease("runai-backend")
	require.Len(t, calls, 1)
	assert.Equal(t, chartPath, calls[0].ChartName)
}

func TestTLSSecret(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "tls.crt")
	keyPath := filepath.Join(dir, "tls.key")
	require.NoError(t, os.WriteFile(certPath, []byte("CERT"), 0600))
	require.NoError(t, os.WriteFile(keyPath, []byte("KEY"), 0600))

	fakeK8s := &harness.FakeK8sExecutor{}
	deps := orchestrator2.Deps{K8sExecutor: fakeK8s}
	act := TLSSecret{SecretName: "runai-backend-tls", Namespace: "runai-backend", CertPath: certPath, KeyPath: keyPath}

	execOK(t, act, deps)

	applies := fakeK8s.CallsByVerb("apply")
	require.Len(t, applies, 1)
	sec, ok := applies[0].Object.(*corev1.Secret)
	require.True(t, ok)
	assert.Equal(t, corev1.SecretTypeTLS, sec.Type)
	assert.Equal(t, []byte("CERT"), sec.Data["tls.crt"])
	assert.Equal(t, []byte("KEY"), sec.Data["tls.key"])
}

func TestEnsureCluster_CreatesWhenMissing(t *testing.T) {
	api := &fakeAPI{}
	deps := orchestrator2.Deps{RunAIClientProvider: providerFor(api)}

	execOK(t, EnsureCluster{ClusterName: "c1", Domain: "runai.example.com"}, deps)

	require.Len(t, api.created, 1)
	assert.Equal(t, "c1", api.created[0].Name)
	require.NotNil(t, api.created[0].Domain)
	assert.Equal(t, "https://runai.example.com", *api.created[0].Domain)
}

func TestEnsureCluster_SkipsWhenExists(t *testing.T) {
	api := &fakeAPI{clusters: []runaicluster.DisplayedCluster{{Name: "c1", Uuid: "u1"}}}
	deps := orchestrator2.Deps{RunAIClientProvider: providerFor(api)}

	execOK(t, EnsureCluster{ClusterName: "c1", Domain: "d"}, deps)

	assert.Empty(t, api.created)
}

func TestInstallClusterChart_BuildsResolvedValues(t *testing.T) {
	api := &fakeAPI{clusters: []runaicluster.DisplayedCluster{{Name: "c1", Uuid: "u1"}}, clientSecret: "sekret"}
	fakeHelm := &harness.FakeHelmInstaller{}
	deps := orchestrator2.Deps{RunAIClientProvider: providerFor(api), HelmInstaller: fakeHelm}
	act := InstallClusterChart{
		ClusterName: "c1",
		Domain:      "runai.example.com",
		Repo:        "https://repo",
		Version:     "2.26.0",
		ReleaseName: "runai-cluster",
		ChartName:   "runai-cluster",
		Namespace:   "runai",
		HostRouting: true,
	}

	execOK(t, act, deps)

	calls := fakeHelm.CallsByRelease("runai-cluster")
	require.Len(t, calls, 1)
	assert.Equal(t, "2.26.0", calls[0].Version)
	assert.Equal(t, "https://repo", calls[0].RepoUrl)
	assert.Subset(t, calls[0].Values, []string{
		"controlPlane.url=runai.example.com",
		"controlPlane.clientSecret=sekret",
		"cluster.uid=u1",
		"cluster.url=runai.example.com",
		"clusterConfig.global.subdomainSupport=true",
	})
}

func TestInstallClusterChart_ClusterNotFound(t *testing.T) {
	api := &fakeAPI{}
	deps := orchestrator2.Deps{RunAIClientProvider: providerFor(api), HelmInstaller: &harness.FakeHelmInstaller{}}

	_, err := InstallClusterChart{ClusterName: "missing", Domain: "d", ReleaseName: "runai-cluster"}.Execute(context.Background(), deps)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCASecret(t *testing.T) {
	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.crt")
	require.NoError(t, os.WriteFile(caPath, []byte("CADATA"), 0600))

	fakeK8s := &harness.FakeK8sExecutor{}
	deps := orchestrator2.Deps{K8sExecutor: fakeK8s}
	act := CASecret{SecretName: "runai-ca-cert", Namespace: "runai", HelmRelease: "runai-cluster", CACertPath: caPath, Labels: map[string]string{"run.ai/cluster-wide": "true"}}

	execOK(t, act, deps)

	applies := fakeK8s.CallsByVerb("apply")
	require.Len(t, applies, 1)
	sec, ok := applies[0].Object.(*corev1.Secret)
	require.True(t, ok)
	assert.Equal(t, corev1.SecretTypeOpaque, sec.Type)
	assert.Equal(t, []byte("CADATA"), sec.Data["runai-ca.pem"])
	assert.Equal(t, "Helm", sec.Labels["app.kubernetes.io/managed-by"])
	assert.Equal(t, "true", sec.Labels["run.ai/cluster-wide"])
}

func TestRegistrySecret(t *testing.T) {
	token := base64.StdEncoding.EncodeToString([]byte(`{"auths":{}}`))
	fakeK8s := &harness.FakeK8sExecutor{}
	deps := orchestrator2.Deps{K8sExecutor: fakeK8s}

	execOK(t, RegistrySecret{SecretName: "runai-reg-creds", Namespace: "runai-backend", Token: token}, deps)

	applies := fakeK8s.CallsByVerb("apply")
	require.Len(t, applies, 1)
	sec, ok := applies[0].Object.(*corev1.Secret)
	require.True(t, ok)
	assert.Equal(t, corev1.SecretTypeDockerConfigJson, sec.Type)
	assert.Equal(t, []byte(`{"auths":{}}`), sec.Data[".dockerconfigjson"])
}

func TestExtractAirgap(t *testing.T) {
	dir := t.TempDir()
	tarball := harness.CreateTestTarball(t)
	workingDir := filepath.Join(dir, "extracted")

	execOK(t, ExtractAirgap{BundlePath: tarball, WorkingDir: workingDir}, orchestrator2.Deps{})

	_, err := os.Stat(filepath.Join(workingDir, "custom-env.yaml"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(workingDir, "charts", "control-plane-2.20.0.tgz"))
	require.NoError(t, err)
}

func TestPushAirgapImages(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "pushed")
	script := "#!/bin/sh\necho \"$REGISTRY_URL $REGISTRY_USERNAME $REGISTRY_PASSWORD\" > " + marker + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "setup.sh"), []byte(script), 0755))

	execOK(t, PushAirgapImages{
		WorkingDir:  dir,
		RegistryURL: "mirror.local:5000",
		Username:    "admin",
		Password:    "Harbor12345",
	}, orchestrator2.Deps{})

	got, err := os.ReadFile(marker)
	require.NoError(t, err)
	assert.Equal(t, "mirror.local:5000 admin Harbor12345\n", string(got), "registry URL and credentials are passed to setup.sh via the environment")
}

func TestInstallClusterChartLocal_BuildsResolvedValues(t *testing.T) {
	api := &fakeAPI{clusters: []runaicluster.DisplayedCluster{{Name: "c1", Uuid: "u1"}}, clientSecret: "sekret"}
	fakeHelm := &harness.FakeHelmInstaller{}
	deps := orchestrator2.Deps{RunAIClientProvider: providerFor(api), HelmInstaller: fakeHelm}
	act := InstallClusterChartLocal{
		ClusterName: "c1",
		Domain:      "runai.example.com",
		CACertPath:  "/ca.crt",
		ChartPath:   "/pkg/charts/runai-cluster-2.20.0.tgz",
		ReleaseName: "runai-cluster",
		Namespace:   "runai",
		HostRouting: true,
		ValueFiles:  []string{"/pkg/custom-env.yaml"},
		ExtraValues: []string{"foo.bar=baz"},
	}

	execOK(t, act, deps)

	calls := fakeHelm.CallsByRelease("runai-cluster")
	require.Len(t, calls, 1)
	assert.Equal(t, "/pkg/charts/runai-cluster-2.20.0.tgz", calls[0].ChartName)
	assert.Equal(t, []string{"/pkg/custom-env.yaml"}, calls[0].ValueFiles)
	assert.Subset(t, calls[0].Values, []string{
		"controlPlane.url=runai.example.com",
		"controlPlane.clientSecret=sekret",
		"cluster.uid=u1",
		"cluster.url=runai.example.com",
		"global.customCA.enabled=true",
		"clusterConfig.global.subdomainSupport=true",
		"foo.bar=baz",
	})
}

func TestInstallClusterChartLocal_NoHostRouting(t *testing.T) {
	api := &fakeAPI{clusters: []runaicluster.DisplayedCluster{{Name: "c1", Uuid: "u1"}}, clientSecret: "sekret"}
	fakeHelm := &harness.FakeHelmInstaller{}
	deps := orchestrator2.Deps{RunAIClientProvider: providerFor(api), HelmInstaller: fakeHelm}
	act := InstallClusterChartLocal{
		ClusterName: "c1",
		Domain:      "runai.example.com",
		ChartPath:   "/pkg/charts/runai-cluster-2.20.0.tgz",
		ReleaseName: "runai-cluster",
		Namespace:   "runai",
		HostRouting: false,
	}

	execOK(t, act, deps)

	calls := fakeHelm.CallsByRelease("runai-cluster")
	require.Len(t, calls, 1)
	assert.Contains(t, calls[0].Values, "clusterConfig.global.subdomainSupport=false")
	assert.NotContains(t, calls[0].Values, "clusterConfig.global.subdomainSupport=true")
	assert.NotContains(t, calls[0].Values, "global.customCA.enabled=true")
}

func TestInstallClusterChartLocal_ClusterNotFound(t *testing.T) {
	api := &fakeAPI{}
	deps := orchestrator2.Deps{RunAIClientProvider: providerFor(api), HelmInstaller: &harness.FakeHelmInstaller{}}

	_, err := InstallClusterChartLocal{ClusterName: "missing", Domain: "d", ReleaseName: "runai-cluster"}.Execute(context.Background(), deps)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestInstallClusterChartLocal_APIError covers a control-plane API failure that
// is not "cluster not found" (e.g. the cluster listing itself errors), which
// must abort the install rather than proceed with empty resolved values.
func TestInstallClusterChartLocal_APIError(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	api := &fakeAPI{getErr: errors.New("control plane unreachable")}
	deps := orchestrator2.Deps{RunAIClientProvider: providerFor(api), HelmInstaller: fakeHelm}

	_, err := InstallClusterChartLocal{ClusterName: "c1", Domain: "d", ReleaseName: "runai-cluster"}.Execute(context.Background(), deps)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "control plane unreachable")
	assert.Empty(t, fakeHelm.Calls, "no chart is installed when the control-plane API fails")
}

func TestActionsDescribe(t *testing.T) {
	descs := []string{
		K8sApply{Object: &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "runai"}}}.Describe(),
		HelmInstallLocal{ReleaseName: "r", ChartPath: "/c", Namespace: "n"}.Describe(),
		TLSSecret{SecretName: "s", Namespace: "n", CertPath: "/c"}.Describe(),
		CASecret{SecretName: "s", Namespace: "n", CACertPath: "/c"}.Describe(),
		RegistrySecret{SecretName: "s", Namespace: "n"}.Describe(),
		EnsureCluster{ClusterName: "c1", Domain: "d"}.Describe(),
		InstallClusterChart{ClusterName: "c1", ReleaseName: "r", ChartName: "ch", Version: "v", Namespace: "n"}.Describe(),
		ExtractAirgap{BundlePath: "/b.tgz", WorkingDir: "/w"}.Describe(),
		PushAirgapImages{WorkingDir: "/w", RegistryURL: "mirror.local:5000"}.Describe(),
		InstallClusterChartLocal{ClusterName: "c1", ReleaseName: "r", ChartPath: "/c", Namespace: "n"}.Describe(),
	}
	for _, d := range descs {
		assert.NotEmpty(t, d)
	}
}
