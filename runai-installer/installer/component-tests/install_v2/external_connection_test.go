/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install_v2_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/installconfig"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/specrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runExternalConnectionInstall runs a connected install with the external
// connection set by mutate, returning the fake helm installer for assertions.
func runExternalConnectionInstall(t *testing.T, mutate func(*installconfig.Config)) *harness.FakeHelmInstaller {
	t.Helper()
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	mutate(cfg)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))
	return helm
}

// TestExternalConnection_IngressClassToBothCharts verifies the ingress class is
// passed to the control-plane chart (global.ingress.ingressClass) and the
// cluster chart (clusterConfig.global.ingress.ingressClass).
func TestExternalConnection_IngressClassToBothCharts(t *testing.T) {
	helm := runExternalConnectionInstall(t, func(c *installconfig.Config) {
		c.ExternalConnection = &installconfig.ExternalConnection{
			Type:    "ingress",
			Ingress: &installconfig.IngressConnection{ClassName: "haproxy"},
		}
	})

	cp := helm.CallsByRelease("runai-backend")
	require.Len(t, cp, 1)
	assert.True(t, harness.ContainsValue(cp[0].Values, "global.ingress.ingressClass=haproxy"))

	cl := helm.CallsByRelease("runai-cluster")
	require.Len(t, cl, 1)
	assert.True(t, harness.ContainsValue(cl[0].Values, "clusterConfig.global.ingress.ingressClass=haproxy"))
}

// TestExternalConnection_GatewayAPIToBothCharts verifies Gateway API routing is
// passed to both charts with their respective shapes: the control-plane chart
// uses global.gatewayAPI.parentRefs[]; the cluster chart uses a single
// clusterConfig.global.gatewayAPI.{name,namespace}.
func TestExternalConnection_GatewayAPIToBothCharts(t *testing.T) {
	helm := runExternalConnectionInstall(t, func(c *installconfig.Config) {
		c.ExternalConnection = &installconfig.ExternalConnection{
			Type:       "gatewayApi",
			GatewayAPI: &installconfig.GatewayAPIConnection{Name: "my-gateway", Namespace: "gateway-system"},
		}
	})

	cp := helm.CallsByRelease("runai-backend")
	require.Len(t, cp, 1)
	assert.True(t, harness.ContainsValue(cp[0].Values, "global.gatewayAPI.enabled=true"))
	assert.True(t, harness.ContainsValue(cp[0].Values, "global.gatewayAPI.parentRefs[0].name=my-gateway"))
	assert.True(t, harness.ContainsValue(cp[0].Values, "global.gatewayAPI.parentRefs[0].namespace=gateway-system"))

	cl := helm.CallsByRelease("runai-cluster")
	require.Len(t, cl, 1)
	assert.True(t, harness.ContainsValue(cl[0].Values, "clusterConfig.global.gatewayAPI.enabled=true"))
	assert.True(t, harness.ContainsValue(cl[0].Values, "clusterConfig.global.gatewayAPI.name=my-gateway"))
	assert.True(t, harness.ContainsValue(cl[0].Values, "clusterConfig.global.gatewayAPI.namespace=gateway-system"))
}
