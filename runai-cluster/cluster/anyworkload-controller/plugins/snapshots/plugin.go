/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshots

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	kartautils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/utils/karta"
	templateutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/utils/template"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// PluginName identifies the plugin in the registry and in status.plugins.
const PluginName = "snapshot-plugin"

// kartaGetter is the Karta lookup the plugin needs, narrowed so tests can stub it.
type kartaGetter interface {
	GetKartaForGVK(ctx context.Context, gvk schema.GroupVersionKind) (*kartav1alpha1.Karta, error)
}

type Plugin struct {
	kartas            kartaGetter
	parser            *templateutils.Parser
	snapshotInstalled bool
}

// NewBuilder returns the framework.PluginBuilder. main registers the plugin
// whether or not Snapshot is installed, so the labels it stamps never change
// when Snapshot is installed or removed.
func NewBuilder(snapshotInstalled bool) framework.PluginBuilder {
	return func(_ context.Context, deps framework.PluginDeps) (framework.Plugin, error) {
		return newPlugin(kartautils.NewFactory(deps.Client), snapshotInstalled), nil
	}
}

func newPlugin(kartas kartaGetter, snapshotInstalled bool) *Plugin {
	return &Plugin{kartas: kartas, parser: templateutils.NewParser(), snapshotInstalled: snapshotInstalled}
}

func (p *Plugin) Name() string {
	return PluginName
}

var (
	_ framework.StatusPlugin         = &Plugin{}
	_ framework.ResourceMutatePlugin = &Plugin{}
)
