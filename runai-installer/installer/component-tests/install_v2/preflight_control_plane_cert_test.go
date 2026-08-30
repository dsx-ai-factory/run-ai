package install_v2_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/installconfig"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/specrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runPreflightWithTLSProber runs an install with a fake TLS prober wired in, so
// the control-plane certificate check is assembled.
func runPreflightWithTLSProber(t *testing.T, cfg *installconfig.Config, tlsProber *harness.FakeTLSProber) (error, []map[string]any, *harness.FakeHelmInstaller) {
	t.Helper()
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe(), TLSProber: tlsProber}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()
	err := specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts())
	return err, readEvents(t, eventsPath), helm
}

func controlPlaneCertResult(t *testing.T, evs []map[string]any) map[string]any {
	t.Helper()
	step := findEvent(evs, "step_succeeded", "preflight-control-plane-certificate")
	require.NotNil(t, step, "the control-plane certificate check must run")
	results, ok := step["results"].([]any)
	require.True(t, ok && len(results) > 0, "the certificate check must report a finding")
	return results[0].(map[string]any)
}

// TestPreflight_ControlPlaneCert_IngressSkipped verifies the check does NOT run
// for an ingress front door: the installer creates the TLS secret from the
// user-provided (often self-signed) certificate, so there is nothing external
// to verify. The prober is never consulted and the install proceeds.
func TestPreflight_ControlPlaneCert_IngressSkipped(t *testing.T) {
	tlsProber := &harness.FakeTLSProber{Result: orchestrator2.TLSProbeResult{
		Reachable: true, Verified: false, Detail: "x509: certificate is valid for ingress.local",
	}}
	err, evs, helm := runPreflightWithTLSProber(t, newInstallRunConfig(), tlsProber)

	require.NoError(t, err)
	assert.Nil(t, findEvent(evs, "step_succeeded", "preflight-control-plane-certificate"), "the check must not run for an ingress front door")
	assert.Empty(t, tlsProber.Calls, "the prober must not be consulted for ingress")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
	assert.NotEmpty(t, helm.Calls)
}

// TestPreflight_ControlPlaneCert_GatewayTrusted verifies the check passes (no
// finding) and the install proceeds when the Gateway serves a trusted
// certificate.
func TestPreflight_ControlPlaneCert_GatewayTrusted(t *testing.T) {
	tlsProber := &harness.FakeTLSProber{Result: orchestrator2.TLSProbeResult{Reachable: true, Verified: true}}
	err, evs, helm := runPreflightWithTLSProber(t, gatewayConfig(), tlsProber)

	require.NoError(t, err)
	step := findEvent(evs, "step_succeeded", "preflight-control-plane-certificate")
	require.NotNil(t, step, "the control-plane certificate check must run for a Gateway")
	assert.Empty(t, step["results"], "no finding when the served certificate is trusted")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
	assert.NotEmpty(t, helm.Calls)

	require.Len(t, tlsProber.Calls, 1)
	assert.Equal(t, "test.runai.local", tlsProber.Calls[0].Host)
	assert.Equal(t, 443, tlsProber.Calls[0].Port)
	assert.Empty(t, tlsProber.Calls[0].CAPath)
}

func TestPreflight_ControlPlaneCert_GatewayCustomCA(t *testing.T) {
	rootCA := filepath.Join(t.TempDir(), "root-ca.crt")
	cpCA := filepath.Join(t.TempDir(), "cp-ca.crt")
	require.NoError(t, os.WriteFile(rootCA, []byte("root-ca"), 0o600))
	require.NoError(t, os.WriteFile(cpCA, []byte("cp-ca"), 0o600))
	cases := []struct {
		name   string
		mutate func(*installconfig.Config)
		wantCA string
	}{
		{
			name: "root CA",
			mutate: func(cfg *installconfig.Config) {
				cfg.TLS = &installconfig.TLS{CAChainPath: rootCA}
			},
			wantCA: rootCA,
		},
		{
			name: "control plane override CA",
			mutate: func(cfg *installconfig.Config) {
				cfg.TLS = &installconfig.TLS{CAChainPath: rootCA}
				cfg.ControlPlane.TLS = &installconfig.TLS{CAChainPath: cpCA}
			},
			wantCA: cpCA,
		},
		{
			name: "control plane override without CA",
			mutate: func(cfg *installconfig.Config) {
				cfg.TLS = &installconfig.TLS{CAChainPath: rootCA}
				cfg.ControlPlane.TLS = &installconfig.TLS{}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := gatewayConfig()
			tc.mutate(cfg)
			tlsProber := &harness.FakeTLSProber{Result: orchestrator2.TLSProbeResult{Reachable: true, Verified: true}}

			err, _, _ := runPreflightWithTLSProber(t, cfg, tlsProber)

			require.NoError(t, err)
			require.Len(t, tlsProber.Calls, 1)
			assert.Equal(t, tc.wantCA, tlsProber.Calls[0].CAPath)
		})
	}
}

// TestPreflight_ControlPlaneCert_GatewayUntrusted_Warns verifies that an
// untrusted certificate on a Gateway is a non-blocking warning: the install
// still proceeds (the user owns the Gateway's certificate).
func TestPreflight_ControlPlaneCert_GatewayUntrusted_Warns(t *testing.T) {
	tlsProber := &harness.FakeTLSProber{Result: orchestrator2.TLSProbeResult{
		Reachable: true, Verified: false, Detail: "x509: certificate signed by unknown authority",
	}}
	err, evs, helm := runPreflightWithTLSProber(t, gatewayConfig(), tlsProber)

	require.NoError(t, err)
	assert.Equal(t, "warning", preflightOutcome(t, evs))
	result := controlPlaneCertResult(t, evs)
	assert.Equal(t, "warn", result["status"])
	assert.Contains(t, result["reason"], "not trusted")
	assert.Contains(t, result["reason"], "unknown authority")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"), "a warning must not block the install")
	assert.NotEmpty(t, helm.Calls)
}

// TestPreflight_ControlPlaneCert_GatewayUnreachable_Warns verifies that an
// unreachable endpoint (DNS/front door not ready) is a non-blocking warning.
func TestPreflight_ControlPlaneCert_GatewayUnreachable_Warns(t *testing.T) {
	tlsProber := &harness.FakeTLSProber{Result: orchestrator2.TLSProbeResult{
		Reachable: false, Detail: "dial tcp: lookup test.runai.local: no such host",
	}}
	err, evs, helm := runPreflightWithTLSProber(t, gatewayConfig(), tlsProber)

	require.NoError(t, err)
	assert.Equal(t, "warning", preflightOutcome(t, evs))
	result := controlPlaneCertResult(t, evs)
	assert.Equal(t, "warn", result["status"])
	assert.Contains(t, result["reason"], "not reachable yet")
	assert.NotEmpty(t, helm.Calls, "a warning must not block the install")
}
