/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package plugins

import (
	"context"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	InjectableStatusPluginName = "injectable-status-plugin"
)

// InjectableStatusPlugin is a StatusPlugin for tests. Keep a reference to it and call SetPhase
// during arrange to control which phase it reports.
type InjectableStatusPlugin struct {
	phase dromosv1alpha1.PluginPhase
}

// NewInjectableStatusPlugin returns a status plugin with the given name. Phase defaults to Unknown;
// call SetPhase in the test's arrange to set the reported phase.
func NewInjectableStatusPlugin() *InjectableStatusPlugin {
	return &InjectableStatusPlugin{
		phase: dromosv1alpha1.PluginPhaseDisabled,
	}
}

// SetPhase sets the phase reported by Status. Call this during arrange.
func (p *InjectableStatusPlugin) SetPhase(phase dromosv1alpha1.PluginPhase) {
	p.phase = phase
}

func (p *InjectableStatusPlugin) Name() string {
	return InjectableStatusPluginName
}

func (p *InjectableStatusPlugin) Status(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner, k8sClient client.Client, status framework.PluginStatusSetter, state *reconcile.State) error {
	status.SetPhase(p.phase)
	return nil
}

var _ framework.StatusPlugin = &InjectableStatusPlugin{}
