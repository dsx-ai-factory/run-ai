/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install_v2_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/installconfig"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/specrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectedInstall_TLSWithoutCustomCA(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}
	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	assert.True(t, appliedObject(k8s, "runai-backend", "runai-backend-tls"))
	assert.True(t, appliedObject(k8s, "runai", "runai-cluster-domain-tls-secret"))
	assert.False(t, appliedObject(k8s, "runai-backend", "runai-ca-cert"))
	assert.False(t, appliedObject(k8s, "runai", "runai-ca-cert"))

	cpCalls := helm.CallsByRelease("runai-backend")
	require.Len(t, cpCalls, 1)
	assert.NotContains(t, cpCalls[0].Values, "global.customCA.enabled=true")
	clusterCalls := helm.CallsByRelease("runai-cluster")
	require.Len(t, clusterCalls, 1)
	assert.NotContains(t, clusterCalls[0].Values, "global.customCA.enabled=true")

	require.NotEmpty(t, apiClient.ProviderCalls)
	for _, call := range apiClient.ProviderCalls {
		assert.Empty(t, call.CACertPath)
	}
}

func TestConnectedInstall_CompleteOverrideDoesNotInheritRootCA(t *testing.T) {
	rootCert, rootKey := harness.GenCert(t, []string{"test.runai.local", "*.test.runai.local"}, time.Now().Add(24*time.Hour))
	cpCert, cpKey := harness.GenCert(t, []string{"test.runai.local"}, time.Now().Add(24*time.Hour))
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}
	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.TLS = &installconfig.TLS{CertPath: rootCert, KeyPath: rootKey, CAChainPath: rootCert}
	cfg.ControlPlane.TLS = &installconfig.TLS{CertPath: cpCert, KeyPath: cpKey}
	cfg.Cluster.TLS = nil
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	requireTLSSecret(t, k8s, "runai-backend", "runai-backend-tls", cfg.ControlPlane.TLS)
	requireTLSSecret(t, k8s, "runai", "runai-cluster-domain-tls-secret", cfg.TLS)
	assert.False(t, appliedObject(k8s, "runai-backend", "runai-ca-cert"))
	requireCASecret(t, k8s, "runai", rootCert)

	cpCalls := helm.CallsByRelease("runai-backend")
	require.Len(t, cpCalls, 1)
	assert.NotContains(t, cpCalls[0].Values, "global.customCA.enabled=true")
	clusterCalls := helm.CallsByRelease("runai-cluster")
	require.Len(t, clusterCalls, 1)
	assert.Contains(t, clusterCalls[0].Values, "global.customCA.enabled=true")

	require.NotEmpty(t, apiClient.ProviderCalls)
	assert.Empty(t, apiClient.ProviderCalls[0].CACertPath)
	for _, call := range apiClient.ProviderCalls[1:] {
		assert.Equal(t, rootCert, call.CACertPath)
	}
}

func TestConnectedInstall_TrustOnlyTLSUsesExistingServingCertificate(t *testing.T) {
	caPath := filepath.Join(t.TempDir(), "ca.crt")
	require.NoError(t, os.WriteFile(caPath, []byte("test-ca"), 0o600))
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}
	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg := newInstallRunConfig()
	cfg.TLS = &installconfig.TLS{CAChainPath: caPath}
	opts := baseRunOptions()
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	assert.False(t, appliedObject(k8s, "runai-backend", "runai-backend-tls"))
	assert.False(t, appliedObject(k8s, "runai", "runai-cluster-domain-tls-secret"))
	requireCASecret(t, k8s, "runai-backend", caPath)
	requireCASecret(t, k8s, "runai", caPath)
	require.NotEmpty(t, apiClient.ProviderCalls)
	for _, call := range apiClient.ProviderCalls {
		assert.Equal(t, caPath, call.CACertPath)
	}
}

func TestInstallMode_RootTLSWithAndWithoutCustomCA(t *testing.T) {
	cases := []struct {
		name        string
		mode        string
		withCA      bool
		wantCP      bool
		wantCluster bool
	}{
		{name: "control plane only with CA", mode: "runai_control_plane_only", withCA: true, wantCP: true},
		{name: "control plane only without CA", mode: "runai_control_plane_only", wantCP: true},
		{name: "cluster only with CA", mode: "runai_cluster_only", withCA: true, wantCluster: true},
		{name: "cluster only without CA", mode: "runai_cluster_only", wantCluster: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hosts := []string{"test.runai.local"}
			if tc.wantCluster {
				hosts = append(hosts, "*.test.runai.local")
			}
			certPath, keyPath := harness.GenCert(t, hosts, time.Now().Add(24*time.Hour))
			cfg := newInstallRunConfig()
			cfg.Install.Mode = tc.mode
			cfg.TLS = &installconfig.TLS{CertPath: certPath, KeyPath: keyPath}
			if tc.withCA {
				cfg.TLS.CAChainPath = certPath
			}
			if tc.wantCP {
				cfg.Cluster = nil
			} else {
				cfg.ControlPlane = nil
				cfg.Cluster.Domain = "test.runai.local"
			}
			opts := baseRunOptions()
			opts.EventsFile = filepath.Join(t.TempDir(), "events.ndjson")
			helm := harness.NewFakeHelmWithReleases()
			k8s := &harness.FakeK8sExecutor{}
			apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
			if tc.wantCluster {
				apiClient.CPReady = true
			}
			deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}
			term, _, _ := newTestTerminal()

			require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

			assert.Equal(t, tc.wantCP, appliedObject(k8s, "runai-backend", "runai-backend-tls"))
			assert.Equal(t, tc.wantCluster, appliedObject(k8s, "runai", "runai-cluster-domain-tls-secret"))
			namespace := "runai"
			release := "runai-cluster"
			if tc.wantCP {
				namespace = "runai-backend"
				release = "runai-backend"
			}
			assert.Equal(t, tc.withCA, appliedObject(k8s, namespace, "runai-ca-cert"))
			calls := helm.CallsByRelease(release)
			require.Len(t, calls, 1)
			assert.Equal(t, tc.withCA, harness.ContainsValue(calls[0].Values, "global.customCA.enabled=true"))
			require.NotEmpty(t, apiClient.ProviderCalls)
			for _, call := range apiClient.ProviderCalls {
				assert.Equal(t, cfg.TLS.CAChainPath, call.CACertPath)
			}
		})
	}
}

func appliedObject(k8s *harness.FakeK8sExecutor, namespace, name string) bool {
	for _, call := range k8s.CallsByVerb("apply") {
		if call.Namespace == namespace && call.Name == name {
			return true
		}
	}
	return false
}
