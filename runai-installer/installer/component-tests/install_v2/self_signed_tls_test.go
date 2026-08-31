/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install_v2_test

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/installconfig"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/internal/specrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectedInstall_GeneratesSelfSignedTLS(t *testing.T) {
	workingDir := t.TempDir()
	t.Chdir(workingDir)

	helm := harness.NewFakeHelmWithReleases()
	manifests := &harness.FakeManifestProvider{}
	k8s := &harness.FakeK8sExecutor{SeededObjects: seedKnativeReady()}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{
		K8sExecutor:         k8s,
		HelmInstaller:       helm,
		RunAIClientProvider: provider,
		ClusterProbe:        okProbe(),
		ManifestProvider:    manifests,
	}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.TLS = &installconfig.TLS{GenerateSelfSigned: true}
	cfg.ControlPlane.TLS = nil
	cfg.Cluster.TLS = nil
	cfg.Cluster.Inference = &installconfig.Inference{WildcardDomain: "*.inference.test.runai.local"}
	cfg.Features = &installconfig.Features{Knative: &installconfig.Dependency{Mode: "install"}}
	opts.Output = spec.OutputFormatText
	term, out, _ := newTestTerminal()
	s := cfg.ToSpec(opts)

	require.NoError(t, specrun.Execute(context.Background(), s, deps, term, execOpts()))

	outputDir := filepath.Join(workingDir, "runai_certs")
	expectedWarning := orchestrator2.NewSelfSignedTLSWarning(outputDir).Message
	assert.Contains(t, out.String(), expectedWarning)
	assert.Equal(t, outputDir, s.SelfSignedTLS.OutputDir)
	require.NotNil(t, s.TLS)

	certData, err := os.ReadFile(s.TLS.CertPath)
	require.NoError(t, err)
	block, _ := pem.Decode(certData)
	require.NotNil(t, block)
	leaf, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{
		"test.runai.local",
		"*.test.runai.local",
		"*.inference.test.runai.local",
	}, leaf.DNSNames)

	resolved := &installconfig.TLS{
		CertPath:    s.TLS.CertPath,
		KeyPath:     s.TLS.KeyPath,
		CAChainPath: s.TLS.CAChainPath,
	}
	requireTLSSecret(t, k8s, "runai-backend", "runai-backend-tls", resolved)
	requireTLSSecret(t, k8s, "runai", "runai-cluster-domain-tls-secret", resolved)
	requireTLSSecret(t, k8s, "knative-serving", "runai-cluster-inference-tls", resolved)
	requireCASecret(t, k8s, "runai-backend", resolved.CAChainPath)
	requireCASecret(t, k8s, "runai", resolved.CAChainPath)

	evs := readEvents(t, eventsPath)
	warningIndex := -1
	pipelineIndex := -1
	for i, event := range evs {
		if event["kind"] == "self_signed_tls" {
			warningIndex = i
			payload := event["selfSignedTLS"].(map[string]any)
			assert.Equal(t, outputDir, payload["outputDir"])
			assert.Equal(t, expectedWarning, payload["message"])
		}
		if pipelineIndex == -1 && event["kind"] == "pipeline_started" {
			pipelineIndex = i
		}
	}
	require.NotEqual(t, -1, warningIndex)
	require.NotEqual(t, -1, pipelineIndex)
	assert.Less(t, warningIndex, pipelineIndex)
}

func TestConnectedInstall_SelfSignedTLSDryRunDoesNotWrite(t *testing.T) {
	workingDir := t.TempDir()
	t.Chdir(workingDir)
	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, true)
	cfg.TLS = &installconfig.TLS{GenerateSelfSigned: true}
	cfg.ControlPlane.TLS = nil
	cfg.Cluster.TLS = nil
	opts.Output = spec.OutputFormatText
	helm := harness.NewFakeHelmWithReleases()
	term, _, _ := newTestTerminal()
	s := cfg.ToSpec(opts)
	initialTLS := s.TLS

	require.NoError(t, specrun.Execute(context.Background(), s, newTestDeps(helm), term, execOpts()))

	_, err := os.Stat(filepath.Join(workingDir, "runai_certs"))
	require.ErrorIs(t, err, os.ErrNotExist)
	assert.Same(t, initialTLS, s.TLS)
	assert.Empty(t, s.TLS.CertPath)
	assert.Empty(t, s.TLS.KeyPath)
	assert.Empty(t, s.TLS.CAChainPath)
	assert.Empty(t, s.SelfSignedTLS.OutputDir)
	assert.Nil(t, findEventByKind(readEvents(t, eventsPath), "self_signed_tls"))
}

func TestInstallMode_GeneratesSelfSignedTLS(t *testing.T) {
	cases := []struct {
		name        string
		mutate      func(*installconfig.Config)
		wantCP      bool
		wantCluster bool
	}{
		{
			name: "control plane only",
			mutate: func(cfg *installconfig.Config) {
				cfg.Install.Mode = "runai_control_plane_only"
				cfg.Cluster = nil
			},
			wantCP: true,
		},
		{
			name: "cluster only",
			mutate: func(cfg *installconfig.Config) {
				cfg.Install.Mode = "runai_cluster_only"
				cfg.ControlPlane = nil
				cfg.Cluster.Domain = "cluster.example.com"
			},
			wantCluster: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			cfg := newInstallRunConfig()
			cfg.TLS = &installconfig.TLS{GenerateSelfSigned: true}
			tc.mutate(cfg)
			opts := baseRunOptions()
			opts.Output = spec.OutputFormatText
			opts.EventsFile = filepath.Join(t.TempDir(), "events.ndjson")
			helm := harness.NewFakeHelmWithReleases()
			k8s := &harness.FakeK8sExecutor{}
			apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
			if tc.wantCluster {
				apiClient.CPReady = true
			}
			deps := orchestrator2.Deps{
				K8sExecutor:         k8s,
				HelmInstaller:       helm,
				RunAIClientProvider: provider,
				ClusterProbe:        okProbe(),
			}
			term, _, _ := newTestTerminal()

			require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

			applied := appliedNames(k8s)
			assert.Equal(t, tc.wantCP, applied["runai-backend-tls"])
			assert.Equal(t, tc.wantCluster, applied["runai-cluster-domain-tls-secret"])
		})
	}
}

func TestInputValidation_SelfSignedTLS(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*installconfig.Config)
		wantSub string
	}{
		{
			name: "root certificate paths",
			mutate: func(cfg *installconfig.Config) {
				cfg.TLS = &installconfig.TLS{
					GenerateSelfSigned: true,
					CertPath:           "/tls.crt",
					KeyPath:            "/tls.key",
					CAChainPath:        "/ca.crt",
				}
			},
			wantSub: "remove this certificate configuration",
		},
		{
			name: "control plane override",
			mutate: func(cfg *installconfig.Config) {
				cfg.TLS = &installconfig.TLS{GenerateSelfSigned: true}
				cfg.ControlPlane.TLS = &installconfig.TLS{CertPath: "/cp.crt"}
			},
			wantSub: "remove this certificate configuration",
		},
		{
			name: "cluster override",
			mutate: func(cfg *installconfig.Config) {
				cfg.TLS = &installconfig.TLS{GenerateSelfSigned: true}
				cfg.Cluster.TLS = &installconfig.TLS{CertPath: "/cluster.crt"}
			},
			wantSub: "remove this certificate configuration",
		},
		{
			name: "inference override",
			mutate: func(cfg *installconfig.Config) {
				cfg.TLS = &installconfig.TLS{GenerateSelfSigned: true}
				cfg.Cluster.Inference = &installconfig.Inference{
					WildcardDomain: "*.inference.test.runai.local",
					WildcardTLS:    &installconfig.TLS{CertPath: "/inference.crt"},
				}
			},
			wantSub: "remove this certificate configuration",
		},
		{
			name: "Gateway API",
			mutate: func(cfg *installconfig.Config) {
				cfg.TLS = &installconfig.TLS{GenerateSelfSigned: true}
				cfg.ExternalConnection = &installconfig.ExternalConnection{
					Type:       "gatewayApi",
					GatewayAPI: &installconfig.GatewayAPIConnection{Name: "gateway", Namespace: "gateway-system"},
				}
			},
			wantSub: "existing Gateway API resource",
		},
		{
			name: "OpenShift",
			mutate: func(cfg *installconfig.Config) {
				cfg.TLS = &installconfig.TLS{GenerateSelfSigned: true}
				cfg.Install.KubernetesDistribution = "openshift"
				cfg.ExternalConnection = nil
			},
			wantSub: "not supported on OpenShift",
		},
		{
			name: "missing domain",
			mutate: func(cfg *installconfig.Config) {
				cfg.TLS = &installconfig.TLS{GenerateSelfSigned: true}
				cfg.ControlPlane = nil
				cfg.Cluster.Domain = ""
			},
			wantSub: "domain is required",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := newInstallRunConfig()
			opts := baseRunOptions()
			opts.Output = spec.OutputFormatText
			opts.EventsFile = filepath.Join(t.TempDir(), "events.ndjson")
			tc.mutate(cfg)
			assertValidationFails(t, cfg, opts, tc.wantSub)
		})
	}
}

func TestSelfSignedTLSConflictCannotBeSkipped(t *testing.T) {
	cfg := newInstallRunConfig()
	cfg.TLS = &installconfig.TLS{GenerateSelfSigned: true}
	cfg.ControlPlane.TLS = &installconfig.TLS{CertPath: "/cp.crt"}
	opts := baseRunOptions()
	opts.OnFailure = spec.FailurePolicySkip
	opts.Output = spec.OutputFormatText
	opts.EventsFile = filepath.Join(t.TempDir(), "events.ndjson")
	helm := &harness.FakeHelmInstaller{}
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	term, _, _ := newTestTerminal()

	err := specrun.Execute(
		context.Background(),
		cfg.ToSpec(opts),
		orchestrator2.Deps{
			K8sExecutor:         k8s,
			HelmInstaller:       helm,
			RunAIClientProvider: provider,
			ClusterProbe:        okProbe(),
		},
		term,
		specrun.ExecOptions{PlanBuilder: execOpts().PlanBuilder, OnFailureSet: true},
	)

	require.ErrorContains(t, err, "cannot be combined")
	assert.Empty(t, helm.Calls)
	assert.Empty(t, k8s.CallsByVerb("apply"))
}
