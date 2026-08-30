package install_v2_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/installconfig"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/specrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helmOverridesConfig builds a connected install where the Prometheus prereq,
// the control-plane chart, and the cluster chart each carry a customHelmValues
// entry that collides with one of our built-in defaults, so the tests can
// assert the user's value wins (and is not emitted twice).
func helmOverridesConfig(t *testing.T, path string, dryRun bool) (*installconfig.Config, installconfig.RunOptions) {
	t.Helper()
	cfg, opts := connectedConfig(t, path, dryRun)
	cfg.Prerequisites.Prometheus.CustomHelmValues = map[string]any{"grafana.enabled": true}
	cfg.ControlPlane.CustomHelmValues = map[string]any{"global.domain": "cp-override.example.com"}
	cfg.Cluster.CustomHelmValues = map[string]any{"clusterConfig.global.subdomainSupport": true}
	return cfg, opts
}

func countWithPrefix(values []string, prefix string) int {
	n := 0
	for _, v := range values {
		if strings.HasPrefix(v, prefix) {
			n++
		}
	}
	return n
}

// TestHelmOverrides_UserValueWins verifies the customHelmValues mechanism for
// every installed component (prereqs, features, and the Run:ai charts). A value
// that collides with one of our built-in defaults replaces it (and is emitted
// once, not duplicated); a value with no matching default is appended through.
func TestHelmOverrides_UserValueWins(t *testing.T) {
	cases := []struct {
		name    string
		release string
		setup   func(*installconfig.Config)
		// want is applied; absent (a replaced default) must not remain; prefix
		// must appear exactly once. absent is empty for purely additive values.
		want   string
		absent string
		prefix string
	}{
		{
			name:    "prometheus (replaces default)",
			release: "kube-prometheus-stack",
			setup:   func(c *installconfig.Config) { c.Prerequisites.Prometheus.CustomHelmValues = map[string]any{"grafana.enabled": true} },
			want:    "grafana.enabled=true", absent: "grafana.enabled=false", prefix: "grafana.enabled=",
		},
		{
			name:    "gpu-operator real (additive)",
			release: "gpu-operator",
			setup:   func(c *installconfig.Config) { c.Prerequisites.GPUOperator.CustomHelmValues = map[string]any{"driver.enabled": false} },
			want:    "driver.enabled=false", prefix: "driver.enabled=",
		},
		{
			name:    "gpu-operator fake (replaces default)",
			release: "gpu-operator",
			setup: func(c *installconfig.Config) {
				c.Prerequisites.GPUOperator.Type = "fake"
				c.Prerequisites.GPUOperator.CustomHelmValues = map[string]any{"topology.nodePools.default.gpuCount": 4}
			},
			want: "topology.nodePools.default.gpuCount=4", absent: "topology.nodePools.default.gpuCount=8", prefix: "topology.nodePools.default.gpuCount=",
		},
		{
			name:    "network-operator (additive)",
			release: "network-operator",
			setup:   func(c *installconfig.Config) { c.Prerequisites.NetworkOperator.CustomHelmValues = map[string]any{"sriovNetworkOperator.enabled": true} },
			want:    "sriovNetworkOperator.enabled=true", prefix: "sriovNetworkOperator.enabled=",
		},
		{
			name:    "lws feature (additive)",
			release: "lws",
			setup: func(c *installconfig.Config) {
				c.Features = &installconfig.Features{LWS: &installconfig.Dependency{Mode: "install", CustomHelmValues: map[string]any{"image.manager.tag": "custom"}}}
			},
			want: "image.manager.tag=custom", prefix: "image.manager.tag=",
		},
		{
			name:    "control-plane (replaces default)",
			release: "runai-backend",
			setup:   func(c *installconfig.Config) { c.ControlPlane.CustomHelmValues = map[string]any{"global.domain": "cp-override.example.com"} },
			want:    "global.domain=cp-override.example.com", absent: "global.domain=test.runai.local", prefix: "global.domain=",
		},
		{
			name:    "cluster (replaces default)",
			release: "runai-cluster",
			setup:   func(c *installconfig.Config) { c.Cluster.CustomHelmValues = map[string]any{"clusterConfig.global.subdomainSupport": true} },
			want:    "clusterConfig.global.subdomainSupport=true", absent: "clusterConfig.global.subdomainSupport=false", prefix: "clusterConfig.global.subdomainSupport=",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			helm := harness.NewFakeHelmWithReleases()
			k8s := &harness.FakeK8sExecutor{}
			_, provider := harness.NewFakeRunAIClientProvider(helm)
			deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe(), NodeLabeler: &harness.FakeNodeLabeler{Labeled: 1}}

			eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
			cfg, opts := connectedConfig(t, eventsPath, false)
			tc.setup(cfg)
			term, _, _ := newTestTerminal()

			require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

			calls := helm.CallsByRelease(tc.release)
			require.Len(t, calls, 1, "%s installs once", tc.release)
			assert.Contains(t, calls[0].Values, tc.want, "user value must be applied")
			if tc.absent != "" {
				assert.NotContains(t, calls[0].Values, tc.absent, "replaced default must not remain")
			}
			assert.Equal(t, 1, countWithPrefix(calls[0].Values, tc.prefix), "value must not be duplicated")
		})
	}
}

// TestHelmOverrides_DryRunDescribe verifies that a dry run prints the intended
// helm command for prereqs and the Run:ai charts with each value override on
// its own continued line, showing the user's value and not the replaced default.
func TestHelmOverrides_DryRunDescribe(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := helmOverridesConfig(t, eventsPath, true)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	assert.Empty(t, helm.Calls, "dry-run performs no helm installs")
	evs := readEvents(t, eventsPath)

	promOut := dryRunOutput(t, evs, "install-prometheus")
	assert.Contains(t, promOut, "helm.install: helm upgrade --install kube-prometheus-stack")
	assert.Contains(t, promOut, " \\\n  --set grafana.enabled=true", "value override renders on its own continued line")
	assert.NotContains(t, promOut, "grafana.enabled=false", "replaced default must not appear")

	cpOut := dryRunOutput(t, evs, "install-control-plane-chart")
	assert.Contains(t, cpOut, " \\\n  --set global.domain=cp-override.example.com")
	assert.NotContains(t, cpOut, "global.domain=test.runai.local")

	clOut := dryRunOutput(t, evs, "install-cluster-chart")
	assert.Contains(t, clOut, " \\\n  --set clusterConfig.global.subdomainSupport=true")
}

// dryRunOutput returns the rendered command from a dry-run step_skipped event.
func dryRunOutput(t *testing.T, evs []map[string]any, step string) string {
	t.Helper()
	e := findEvent(evs, "step_skipped", step)
	require.NotNil(t, e, "expected a step_skipped event for %s", step)
	require.Equal(t, "dry_run", e["skipReason"], "%s must be a dry-run skip", step)
	out, _ := e["output"].(string)
	return out
}
