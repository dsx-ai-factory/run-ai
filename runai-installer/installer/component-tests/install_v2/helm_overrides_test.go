/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install_v2_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/installconfig"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/specrun"
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
	cfg.Prerequisites.Prometheus.CustomHelmValues = map[string]any{"grafana": map[string]any{"enabled": true}}
	cfg.ControlPlane.CustomHelmValues = map[string]any{"global": map[string]any{"domain": "cp-override.example.com"}}
	cfg.Cluster.CustomHelmValues = map[string]any{"clusterConfig": map[string]any{"global": map[string]any{"subdomainSupport": true}}}
	return cfg, opts
}

// TestHelmOverrides_UserValueWins verifies the customHelmValues mechanism for
// every installed component (prereqs, features, and the Run:ai charts). A value
// that collides with an installer default wins in the merged dictionary;
// unrelated defaults remain intact.
func TestHelmOverrides_UserValueWins(t *testing.T) {
	cases := []struct {
		name    string
		release string
		setup   func(*installconfig.Config)
		want    string
	}{
		{
			name:    "prometheus (replaces default)",
			release: "kube-prometheus-stack",
			setup: func(c *installconfig.Config) {
				c.Prerequisites.Prometheus.CustomHelmValues = map[string]any{"grafana": map[string]any{"enabled": true}}
			},
			want: "grafana.enabled=true",
		},
		{
			name:    "gpu-operator real (additive)",
			release: "gpu-operator",
			setup: func(c *installconfig.Config) {
				c.Prerequisites.GPUOperator.CustomHelmValues = map[string]any{"driver": map[string]any{"enabled": false}}
			},
			want: "driver.enabled=false",
		},
		{
			name:    "gpu-operator fake (replaces default)",
			release: "gpu-operator",
			setup: func(c *installconfig.Config) {
				c.Prerequisites.GPUOperator.Type = "fake"
				c.Prerequisites.GPUOperator.CustomHelmValues = map[string]any{"topology": map[string]any{"nodePools": map[string]any{"default": map[string]any{"gpuCount": 4}}}}
			},
			want: "topology.nodePools.default.gpuCount=4",
		},
		{
			name:    "network-operator (additive)",
			release: "network-operator",
			setup: func(c *installconfig.Config) {
				c.Prerequisites.NetworkOperator.CustomHelmValues = map[string]any{"sriovNetworkOperator": map[string]any{"enabled": true}}
			},
			want: "sriovNetworkOperator.enabled=true",
		},
		{
			name:    "lws feature (additive)",
			release: "lws",
			setup: func(c *installconfig.Config) {
				c.Features = &installconfig.Features{LWS: &installconfig.Dependency{Mode: "install", CustomHelmValues: map[string]any{"image": map[string]any{"manager": map[string]any{"tag": "custom"}}}}}
			},
			want: "image.manager.tag=custom",
		},
		{
			name:    "control-plane (replaces default)",
			release: "runai-backend",
			setup: func(c *installconfig.Config) {
				c.ControlPlane.CustomHelmValues = map[string]any{"global": map[string]any{"domain": "cp-override.example.com"}}
			},
			want: "global.domain=cp-override.example.com",
		},
		{
			name:    "cluster (replaces default)",
			release: "runai-cluster",
			setup: func(c *installconfig.Config) {
				c.Cluster.CustomHelmValues = map[string]any{"clusterConfig": map[string]any{"global": map[string]any{"subdomainSupport": true}}}
			},
			want: "clusterConfig.global.subdomainSupport=true",
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
			resolved, err := calls[0].ResolvedValues()
			require.NoError(t, err)
			key, want, _ := strings.Cut(tc.want, "=")
			value, present := spec.HelmValue(resolved, strings.Split(key, ".")...)
			require.True(t, present)
			assert.Equal(t, want, fmt.Sprint(value), "native custom value must override the installer default")
		})
	}
}

// TestHelmOverrides_DryRunDescribe verifies that a dry run prints the intended
// Helm command and the native custom values as YAML for each chart.
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
	assert.Contains(t, promOut, "Custom Helm values (YAML):")
	assert.Contains(t, promOut, "enabled: true")

	cpOut := dryRunOutput(t, evs, "install-control-plane-chart")
	assert.Contains(t, cpOut, "Custom Helm values (YAML):")
	assert.Contains(t, cpOut, "domain: cp-override.example.com")

	clOut := dryRunOutput(t, evs, "install-cluster-chart")
	assert.Contains(t, clOut, "Custom Helm values (YAML):")
	assert.Contains(t, clOut, "subdomainSupport: true")
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
