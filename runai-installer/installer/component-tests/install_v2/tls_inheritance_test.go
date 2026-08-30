package install_v2_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/installconfig"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/specrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

type tlsFixtures struct {
	root         *installconfig.TLS
	controlPlane *installconfig.TLS
	cluster      *installconfig.TLS
	inference    *installconfig.TLS
	byName       map[string]*installconfig.TLS
}

type tlsSelection struct {
	controlPlane string
	cluster      string
	inference    string
}

// TestConnectedInstall_TLSInheritance verifies that root TLS defaults and
// component overrides produce the expected control-plane, cluster, inference,
// and CA secrets during a full installation.
func TestConnectedInstall_TLSInheritance(t *testing.T) {
	cases := []struct {
		name      string
		configure func(*installconfig.Config, tlsFixtures)
		want      tlsSelection
	}{
		{
			name: "root defaults",
			configure: func(cfg *installconfig.Config, tls tlsFixtures) {
				cfg.TLS = tls.root
				cfg.ControlPlane.TLS = nil
				cfg.Cluster.TLS = nil
				cfg.Cluster.Inference.WildcardTLS = nil
			},
			want: tlsSelection{
				controlPlane: "root",
				cluster:      "root",
				inference:    "root",
			},
		},
		{
			name: "control plane override",
			configure: func(cfg *installconfig.Config, tls tlsFixtures) {
				cfg.TLS = tls.root
				cfg.ControlPlane.TLS = tls.controlPlane
				cfg.Cluster.TLS = nil
				cfg.Cluster.Inference.WildcardTLS = nil
			},
			want: tlsSelection{
				controlPlane: "control-plane",
				cluster:      "root",
				inference:    "root",
			},
		},
		{
			name: "cluster override",
			configure: func(cfg *installconfig.Config, tls tlsFixtures) {
				cfg.TLS = tls.root
				cfg.ControlPlane.TLS = nil
				cfg.Cluster.TLS = tls.cluster
				cfg.Cluster.Inference.WildcardTLS = nil
			},
			want: tlsSelection{
				controlPlane: "root",
				cluster:      "cluster",
				inference:    "cluster",
			},
		},
		{
			name: "inference override",
			configure: func(cfg *installconfig.Config, tls tlsFixtures) {
				cfg.TLS = tls.root
				cfg.ControlPlane.TLS = nil
				cfg.Cluster.TLS = nil
				cfg.Cluster.Inference.WildcardTLS = tls.inference
			},
			want: tlsSelection{
				controlPlane: "root",
				cluster:      "root",
				inference:    "inference",
			},
		},
		{
			name: "all overrides",
			configure: func(cfg *installconfig.Config, tls tlsFixtures) {
				cfg.TLS = tls.root
				cfg.ControlPlane.TLS = tls.controlPlane
				cfg.Cluster.TLS = tls.cluster
				cfg.Cluster.Inference.WildcardTLS = tls.inference
			},
			want: tlsSelection{
				controlPlane: "control-plane",
				cluster:      "cluster",
				inference:    "inference",
			},
		},
		{
			name: "component overrides without root",
			configure: func(cfg *installconfig.Config, tls tlsFixtures) {
				cfg.TLS = nil
				cfg.ControlPlane.TLS = tls.controlPlane
				cfg.Cluster.TLS = tls.cluster
				cfg.Cluster.Inference.WildcardTLS = tls.inference
			},
			want: tlsSelection{
				controlPlane: "control-plane",
				cluster:      "cluster",
				inference:    "inference",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rootCert, rootKey := harness.GenCert(t, []string{"test.runai.local", "*.inference.test.runai.local"}, time.Now().Add(24*time.Hour))
			cpCert, cpKey := harness.GenCert(t, []string{"test.runai.local"}, time.Now().Add(24*time.Hour))
			clusterCert, clusterKey := harness.GenCert(t, []string{"test.runai.local", "*.inference.test.runai.local"}, time.Now().Add(24*time.Hour))
			inferenceCert, inferenceKey := harness.GenCert(t, []string{"*.inference.test.runai.local"}, time.Now().Add(24*time.Hour))

			tls := tlsFixtures{
				root:         &installconfig.TLS{CertPath: rootCert, KeyPath: rootKey, CAChainPath: rootCert},
				controlPlane: &installconfig.TLS{CertPath: cpCert, KeyPath: cpKey, CAChainPath: cpCert},
				cluster:      &installconfig.TLS{CertPath: clusterCert, KeyPath: clusterKey, CAChainPath: clusterCert},
				inference:    &installconfig.TLS{CertPath: inferenceCert, KeyPath: inferenceKey, CAChainPath: inferenceCert},
			}
			tls.byName = map[string]*installconfig.TLS{
				"root":          {CertPath: tls.root.CertPath, KeyPath: tls.root.KeyPath, CAChainPath: tls.root.CAChainPath},
				"control-plane": tls.controlPlane,
				"cluster":       tls.cluster,
				"inference":     tls.inference,
			}

			helm := harness.NewFakeHelmWithReleases()
			manifests := &harness.FakeManifestProvider{}
			k8s := &harness.FakeK8sExecutor{SeededObjects: seedKnativeReady()}
			apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
			deps := orchestrator2.Deps{
				K8sExecutor:         k8s,
				HelmInstaller:       helm,
				RunAIClientProvider: provider,
				ClusterProbe:        okProbe(),
				ManifestProvider:    manifests,
			}

			eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
			cfg, opts := connectedConfig(t, eventsPath, false)
			cfg.Cluster.Inference = &installconfig.Inference{WildcardDomain: "*.inference.test.runai.local"}
			cfg.Features = &installconfig.Features{Knative: &installconfig.Dependency{Mode: "install"}}
			tc.configure(cfg, tls)

			term, _, _ := newTestTerminal()
			require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

			requireTLSSecret(t, k8s, "runai-backend", "runai-backend-tls", tls.byName[tc.want.controlPlane])
			requireTLSSecret(t, k8s, "runai", "runai-cluster-domain-tls-secret", tls.byName[tc.want.cluster])
			requireTLSSecret(t, k8s, "knative-serving", "runai-cluster-inference-tls", tls.byName[tc.want.inference])
			requireCASecret(t, k8s, "runai-backend", tls.byName[tc.want.controlPlane].CAChainPath)
			requireCASecret(t, k8s, "runai", tls.byName[tc.want.cluster].CAChainPath)

			cpCalls := helm.CallsByRelease("runai-backend")
			require.Len(t, cpCalls, 1)
			require.Contains(t, cpCalls[0].Values, "global.customCA.enabled=true")
			clusterCalls := helm.CallsByRelease("runai-cluster")
			require.Len(t, clusterCalls, 1)
			require.Contains(t, clusterCalls[0].Values, "global.customCA.enabled=true")

			require.NotEmpty(t, apiClient.ProviderCalls)
			assert.Equal(t, tls.byName[tc.want.controlPlane].CAChainPath, apiClient.ProviderCalls[0].CACertPath)
			for _, call := range apiClient.ProviderCalls[1:] {
				assert.Equal(t, tls.byName[tc.want.cluster].CAChainPath, call.CACertPath)
			}
		})
	}
}

func requireTLSSecret(t *testing.T, k8s *harness.FakeK8sExecutor, namespace, name string, tls *installconfig.TLS) {
	t.Helper()
	wantCert, err := os.ReadFile(tls.CertPath)
	require.NoError(t, err)
	wantKey, err := os.ReadFile(tls.KeyPath)
	require.NoError(t, err)

	for _, call := range k8s.CallsByVerb("apply") {
		if call.Namespace != namespace || call.Name != name {
			continue
		}
		secret, ok := call.Object.(*corev1.Secret)
		require.True(t, ok)
		require.Equal(t, wantCert, secret.Data[corev1.TLSCertKey])
		require.Equal(t, wantKey, secret.Data[corev1.TLSPrivateKeyKey])
		return
	}

	require.Failf(t, "TLS secret not applied", "%s/%s", namespace, name)
}

func requireCASecret(t *testing.T, k8s *harness.FakeK8sExecutor, namespace, caPath string) {
	t.Helper()
	want, err := os.ReadFile(caPath)
	require.NoError(t, err)

	for _, call := range k8s.CallsByVerb("apply") {
		if call.Namespace != namespace || call.Name != "runai-ca-cert" {
			continue
		}
		secret, ok := call.Object.(*corev1.Secret)
		require.True(t, ok)
		require.Equal(t, want, secret.Data["runai-ca.pem"])
		return
	}

	require.Failf(t, "CA secret not applied", "%s/runai-ca-cert", namespace)
}
