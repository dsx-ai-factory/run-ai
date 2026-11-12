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

// Plugin captures a snapshot of each opted-in workload. dromos calls it on every
// reconcile of the runner, in this order:
//
//  1. HandleAdditionalResources (capture.go): once a replica is Ready, creates
//     the SnapshotJob that captures its key. It runs before this reconcile's
//     mutate and apply; the replicas it captures come from an earlier one.
//  2. MutateResource (mutate.go): stamps the opt-in and runner labels on the pod
//     template, which the restore webhook and the capture find replicas by.
//  3. dromos applies the workload.
//  4. Status (status.go): Ready when the workload can be captured; otherwise
//     Disabled, with a warning when its opt-in can't be honoured.
//
// A replica turning Ready, or a capture changing, reconciles the runner again
// (watches.go). Once the runner is gone, NewCleanupSetup deletes its captures
// (cleanup.go).
type Plugin struct {
	kartas            kartaGetter
	parser            *templateutils.Parser
	snapshotInstalled bool
	agentNodeSelector map[string]string
}

// NewBuilder returns the framework.PluginBuilder. main registers the plugin
// whether or not Snapshot is installed, so the labels it stamps never change
// when Snapshot is installed or removed.
func NewBuilder(snapshotInstalled bool, agentNodeSelector map[string]string) framework.PluginBuilder {
	return func(_ context.Context, deps framework.PluginDeps) (framework.Plugin, error) {
		plugin := newPlugin(kartautils.NewFactory(deps.Client), snapshotInstalled)
		plugin.agentNodeSelector = agentNodeSelector
		return plugin, nil
	}
}

func newPlugin(kartas kartaGetter, snapshotInstalled bool) *Plugin {
	return &Plugin{kartas: kartas, parser: templateutils.NewParser(), snapshotInstalled: snapshotInstalled}
}

func (p *Plugin) Name() string {
	return PluginName
}

var (
	_ framework.StatusPlugin             = &Plugin{}
	_ framework.ResourceMutatePlugin     = &Plugin{}
	_ framework.AdditionalResourcePlugin = &Plugin{}
	_ framework.WatcherPlugin            = &Plugin{}
)
