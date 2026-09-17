/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator

import (
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/flow"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/uninstall"
	"github.com/stretchr/testify/require"
)

// The uninstall flow groups run strictly in order, so the order of this list is
// the order in which resources are removed from the cluster.
func TestBuildUninstallFlowGroups_Order(t *testing.T) {
	cfg := &config.UninstallConfig{StepTimeout: time.Minute}
	groups := buildUninstallFlowGroups(cfg, uninstall.Deps{}, make(chan *flow.FlowStatusReportingMsg, 100))

	var got [][]string
	for _, g := range groups {
		var names []string
		for _, f := range g.Flows {
			names = append(names, f.GetName())
		}
		got = append(got, names)
	}

	require.Equal(t, [][]string{
		{
			"Uninstall Cluster",
			"Uninstall Control Plane",
			"Delete namespace runai-reservation",
			"Delete namespace runai-scale-adjust",
		},
		// Webhook configurations go before the finalizer strip: once the namespaces
		// are gone, a still-registered validating webhook with failurePolicy=Fail
		// rejects the finalizer patch and leaves the CRD stuck in Terminating.
		{"Delete RunAI MutatingWebhookConfigurations", "Delete RunAI ValidatingWebhookConfigurations"},
		{"Strip RunAI CR finalizers"},
		{"Delete RunAI CRDs"},
		{"Delete RunAI ResourceInterfaces"},
		{"Delete RunAI ClusterRoles"},
		{"Delete RunAI PriorityClasses"},
	}, got)
}
