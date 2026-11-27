/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshots

import (
	"context"
	"errors"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	kartautils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/utils/karta"
	templateutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/utils/template"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
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
//  1. HandleAdditionalResources (capture.go): deletes the snapshots no longer
//     needed, then, once a replica is Ready, creates the SnapshotJob that
//     captures its key, and emits each capture event on the runner once. It
//     runs before this reconcile's mutate and apply; the replicas it captures
//     come from an earlier one.
//  2. MutateResource (mutate.go): stamps the opt-in and runner labels on the pod
//     template, which the restore webhook and the capture find replicas by.
//  3. dromos applies the workload.
//  4. Status (status.go): Ready when the workload can be captured, Failed once a
//     capture of a key in use failed; otherwise Disabled, with a warning when its
//     opt-in can't be honoured. It records the newest key, which step 1 keeps
//     through a scale to zero.
//
// A replica turning Ready, or a capture changing, reconciles the runner again
// (watches.go). The runner owns its SnapshotJobs and PodSnapshots, so they are
// deleted with it (capture.go).
type Plugin struct {
	kartas            kartaGetter
	parser            *templateutils.Parser
	snapshotInstalled bool
	agentNodeSelector map[string]string
	events            record.EventRecorder
}

// eventSource matches the controller's own, so the plugin's events on the runner
// read like the controller's.
const eventSource = "anyworkload-controller"

// Integration ties the plugin to the manager setup that gives it its event
// recorder: a plugin builder gets no manager.
type Integration struct {
	snapshotInstalled bool
	agentNodeSelector map[string]string
	plugin            *Plugin
}

func NewIntegration(snapshotInstalled bool, agentNodeSelector map[string]string) *Integration {
	return &Integration{snapshotInstalled: snapshotInstalled, agentNodeSelector: agentNodeSelector}
}

// Build is the framework.PluginBuilder. main registers the plugin whether or
// not Snapshot is installed, so the labels it stamps never change when Snapshot
// is installed or removed.
func (i *Integration) Build(_ context.Context, deps framework.PluginDeps) (framework.Plugin, error) {
	i.plugin = newPlugin(kartautils.NewFactory(deps.Client), i.snapshotInstalled)
	i.plugin.agentNodeSelector = i.agentNodeSelector
	return i.plugin, nil
}

// NewSetup returns the manager setup that gives the plugin its event recorder.
// main registers it with Snapshot installed, the only case the plugin emits in.
func (i *Integration) NewSetup() func(context.Context, ctrl.Manager) error {
	return func(_ context.Context, mgr ctrl.Manager) error {
		//nolint:staticcheck // SA1019: repo-wide convention is record.EventRecorder; migrate all controllers together
		return i.Setup(mgr.GetEventRecorderFor(eventSource))
	}
}

// Setup runs after Build and before the manager starts, so no reconcile sees
// the plugin without its recorder.
func (i *Integration) Setup(events record.EventRecorder) error {
	if i.plugin == nil {
		return errors.New("snapshot integration must be built before setup")
	}
	i.plugin.events = events
	return nil
}

func newPlugin(kartas kartaGetter, snapshotInstalled bool) *Plugin {
	return &Plugin{
		kartas:            kartas,
		parser:            templateutils.NewParser(),
		snapshotInstalled: snapshotInstalled,
	}
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
