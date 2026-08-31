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

const systemLabel = "node-role.kubernetes.io/runai-system"

func worker(name string, labels map[string]string) orchestrator2.NodeInfo {
	return orchestrator2.NodeInfo{Name: name, Roles: []string{"worker"}, Labels: labels}
}

// runSystemNodesInstall runs a connected install whose SystemNodes block is set
// by mutate, against a cluster with the given nodes, returning the fake helm
// installer, the fake node labeler, and the decoded events.
func runSystemNodesInstall(t *testing.T, nodes []orchestrator2.NodeInfo, mutate func(*installconfig.Config)) (*harness.FakeHelmInstaller, *harness.FakeNodeLabeler, []map[string]any) {
	t.Helper()
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	labeler := &harness.FakeNodeLabeler{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	probe := &harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true, Nodes: nodes}
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: probe, NodeLabeler: labeler}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	mutate(cfg)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))
	return helm, labeler, readEvents(t, eventsPath)
}

// TestSystemNodes_Preferred verifies a preferred selection labels the chosen
// node and injects the preferred affinity into both the control-plane chart and
// the cluster chart (twice: global.affinity and the RunaiConfig path).
func TestSystemNodes_Preferred(t *testing.T) {
	helm, labeler, evs := runSystemNodesInstall(t, []orchestrator2.NodeInfo{worker("cpu-1", nil)}, func(c *installconfig.Config) {
		c.SystemNodes = &installconfig.SystemNodes{Enabled: true, Nodes: []string{"cpu-1"}}
	})

	cpKey := "global.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].preference.matchExpressions[0].key=" + systemLabel
	cp := helm.CallsByRelease("runai-backend")
	require.Len(t, cp, 1)
	assert.Contains(t, cp[0].Values, "global.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].weight=1")
	assert.Contains(t, cp[0].Values, cpKey)

	cl := helm.CallsByRelease("runai-cluster")
	require.Len(t, cl, 1)
	assert.Contains(t, cl[0].Values, cpKey, "operator/hook pods")
	assert.Contains(t, cl[0].Values, "runai-operator.config.global.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].preference.matchExpressions[0].key="+systemLabel, "operand pods via RunaiConfig")

	require.Len(t, labeler.NodesCalls, 1)
	assert.Equal(t, []string{"cpu-1"}, labeler.NodesCalls[0].Names)
	assert.Equal(t, systemLabel, labeler.NodesCalls[0].Key)
	assert.Empty(t, labeler.NodesCalls[0].Value, "node-role label uses an empty value")
	require.NotNil(t, findEvent(evs, "step_succeeded", "label-system-nodes"))
}

// TestSystemNodes_Required verifies required mode injects the hard affinity term.
func TestSystemNodes_Required(t *testing.T) {
	helm, _, _ := runSystemNodesInstall(t, []orchestrator2.NodeInfo{worker("cpu-1", nil)}, func(c *installconfig.Config) {
		c.SystemNodes = &installconfig.SystemNodes{Enabled: true, Mode: "required", Nodes: []string{"cpu-1"}}
	})

	reqKey := "global.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].key=" + systemLabel
	cp := helm.CallsByRelease("runai-backend")
	require.Len(t, cp, 1)
	assert.Contains(t, cp[0].Values, reqKey)
	for _, v := range cp[0].Values {
		assert.NotContains(t, v, "preferredDuringScheduling", "required mode must not emit a preferred term")
	}
	cl := helm.CallsByRelease("runai-cluster")
	require.Len(t, cl, 1)
	assert.Contains(t, cl[0].Values, "runai-operator.config.global.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].key="+systemLabel)
}

// TestSystemNodes_Selector verifies selector mode labels only matching, non-
// control-plane nodes.
func TestSystemNodes_Selector(t *testing.T) {
	nodes := []orchestrator2.NodeInfo{
		worker("cpu-1", map[string]string{"pool": "system"}),
		worker("gpu-1", map[string]string{"pool": "gpu"}),
		{Name: "cp-1", Roles: []string{"control-plane"}, Labels: map[string]string{"pool": "system"}},
	}
	_, labeler, _ := runSystemNodesInstall(t, nodes, func(c *installconfig.Config) {
		c.SystemNodes = &installconfig.SystemNodes{Enabled: true, Selector: map[string]string{"pool": "system"}}
	})

	require.Len(t, labeler.NodesCalls, 1)
	assert.Equal(t, []string{"cpu-1"}, labeler.NodesCalls[0].Names, "gpu node excluded by label, control-plane node excluded by role")
}

// TestSystemNodes_ManageLabelsFalse verifies the affinity is still injected but
// no labeling happens when the installer is told not to manage labels.
func TestSystemNodes_ManageLabelsFalse(t *testing.T) {
	manage := false
	helm, labeler, evs := runSystemNodesInstall(t, []orchestrator2.NodeInfo{worker("cpu-1", nil)}, func(c *installconfig.Config) {
		c.SystemNodes = &installconfig.SystemNodes{Enabled: true, ManageLabels: &manage, Nodes: []string{"cpu-1"}}
	})

	assert.Empty(t, labeler.NodesCalls, "no labeling when manageLabels is false")
	assert.Nil(t, findEvent(evs, "step_succeeded", "label-system-nodes"), "no label step")

	cp := helm.CallsByRelease("runai-backend")
	require.Len(t, cp, 1)
	assert.Contains(t, cp[0].Values, "global.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].preference.matchExpressions[0].key="+systemLabel)
}

// TestSystemNodes_Advisory_WarnsOnGpuInstall verifies that a GPU install without
// a system-node selection emits the advisory warning (and still proceeds).
func TestSystemNodes_Advisory_WarnsOnGpuInstall(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false) // real GPU operator, no systemNodes
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "warning", preflightOutcome(t, evs))
	step := findEvent(evs, "step_succeeded", "preflight-system-nodes")
	require.NotNil(t, step)
	result := step["results"].([]any)[0].(map[string]any)
	assert.Equal(t, "warn", result["status"])
	assert.Equal(t, "no system nodes selected", result["reason"], "the per-check line stays short")

	// the full advisory is carried in the preflight summary, like resource issues
	summary := findPreflightCompleted(evs)["preflight"].(map[string]any)
	assert.Contains(t, summary["systemNodesAdvisory"], "may be scheduled onto GPU nodes")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"), "an advisory must not block the install")
}

// TestSystemNodes_Advisory_SilentWithoutRealGpu verifies the advisory does not
// fire on a fake-GPU install (nothing to protect), keeping preflight green.
func TestSystemNodes_Advisory_SilentWithoutRealGpu(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	labeler := &harness.FakeNodeLabeler{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe(), NodeLabeler: labeler}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Prerequisites.GPUOperator = &installconfig.Dependency{Mode: "install", Type: "fake"}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "passed", preflightOutcome(t, evs))
	assert.Nil(t, findEvent(evs, "step_succeeded", "preflight-system-nodes"), "no advisory on a fake-GPU install")
}

// TestSystemNodes_EmptySelectionWarns verifies that a selection matching no node
// only warns (even in required mode) and does not block: preflight warns, the
// label step no-ops, and the install proceeds.
func TestSystemNodes_EmptySelectionWarns(t *testing.T) {
	nodes := []orchestrator2.NodeInfo{worker("gpu-1", map[string]string{"pool": "gpu"})}
	helm, labeler, evs := runSystemNodesInstall(t, nodes, func(c *installconfig.Config) {
		c.SystemNodes = &installconfig.SystemNodes{Enabled: true, Mode: "required", Selector: map[string]string{"pool": "system"}}
	})

	assert.Equal(t, "warning", preflightOutcome(t, evs), "an empty selection warns, even in required mode")
	step := findEvent(evs, "step_succeeded", "preflight-system-nodes")
	require.NotNil(t, step)
	result := step["results"].([]any)[0].(map[string]any)
	assert.Equal(t, "warn", result["status"])
	assert.Contains(t, result["reason"], "no nodes match")

	assert.Empty(t, labeler.NodesCalls, "empty match labels nothing")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"), "a warning must not block the install")
	assert.NotEmpty(t, helm.Calls)
}

// TestSystemNodes_MissingNamedNodeWarns verifies a non-existent node name warns
// (listing it) but does not block, and the valid nodes are still labeled.
func TestSystemNodes_MissingNamedNodeWarns(t *testing.T) {
	helm, labeler, evs := runSystemNodesInstall(t, []orchestrator2.NodeInfo{worker("cpu-1", nil)}, func(c *installconfig.Config) {
		c.SystemNodes = &installconfig.SystemNodes{Enabled: true, Nodes: []string{"cpu-1", "cpu-typo"}}
	})

	assert.Equal(t, "warning", preflightOutcome(t, evs))
	step := findEvent(evs, "step_succeeded", "preflight-system-nodes")
	require.NotNil(t, step)
	result := step["results"].([]any)[0].(map[string]any)
	assert.Equal(t, "warn", result["status"])
	assert.Contains(t, result["reason"], "cpu-typo", "the missing node is named in the warning")

	require.Len(t, labeler.NodesCalls, 1)
	assert.Equal(t, []string{"cpu-1"}, labeler.NodesCalls[0].Names, "the valid node is still labeled")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
	assert.NotEmpty(t, helm.Calls)
}

// sysNodeDeps builds orchestrator deps for a system-node test: the given probe,
// a fake helm installer, a fake node labeler, and the fake Run:ai API client
// (so a test can mark the control plane ready for cluster-only installs).
func sysNodeDeps(probe *harness.FakeClusterProbe) (orchestrator2.Deps, *harness.FakeHelmInstaller, *harness.FakeNodeLabeler, *harness.FakeRunAIAPIClient) {
	helm := harness.NewFakeHelmWithReleases()
	labeler := &harness.FakeNodeLabeler{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{
		K8sExecutor:         &harness.FakeK8sExecutor{},
		HelmInstaller:       helm,
		RunAIClientProvider: provider,
		ClusterProbe:        probe,
		NodeLabeler:         labeler,
	}
	return deps, helm, labeler, apiClient
}

// TestSystemNodes_Placement_ControlPlaneOnly verifies that on a control-plane-
// only install the labeling step runs in the control-plane flow.
func TestSystemNodes_Placement_ControlPlaneOnly(t *testing.T) {
	deps, _, labeler, _ := sysNodeDeps(&harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true, Nodes: []orchestrator2.NodeInfo{worker("cpu-1", nil)}})

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Prerequisites = &installconfig.Prerequisites{Skip: true}
	cfg.Cluster.Enabled = false
	cfg.SystemNodes = &installconfig.SystemNodes{Enabled: true, Nodes: []string{"cpu-1"}}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	step := findEvent(evs, "step_succeeded", "label-system-nodes")
	require.NotNil(t, step)
	assert.Equal(t, "control-plane", step["flow"], "labeling runs in the CP flow when only the control plane installs")
	require.Len(t, labeler.NodesCalls, 1)
	assert.Nil(t, findEvent(evs, "step_succeeded", "install-cluster-chart"), "no cluster flow on a CP-only install")
}

// TestSystemNodes_Placement_ClusterOnly verifies that on a cluster-only install
// the labeling step runs in the cluster flow (there is no CP flow to host it).
// The install itself does not complete in the fake harness (register-cluster
// reports the control plane as not ready, since no CP was installed this run),
// which is irrelevant here — the assertion is purely about label-step placement.
func TestSystemNodes_Placement_ClusterOnly(t *testing.T) {
	deps, helm, labeler, apiClient := sysNodeDeps(&harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true, Nodes: []orchestrator2.NodeInfo{worker("cpu-1", nil)}})
	apiClient.CPReady = true // cluster-only installs against an already-present control plane

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Prerequisites = &installconfig.Prerequisites{Skip: true}
	cfg.ControlPlane.Enabled = false
	cfg.Cluster.Domain = "test.runai.local" // no CP to fall back to
	cfg.SystemNodes = &installconfig.SystemNodes{Enabled: true, Nodes: []string{"cpu-1"}}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	step := findEvent(evs, "step_succeeded", "label-system-nodes")
	require.NotNil(t, step, "the labeling step must run on a cluster-only install")
	assert.Equal(t, "cluster", step["flow"], "labeling runs in the cluster flow when only the cluster installs")
	require.Len(t, labeler.NodesCalls, 1)
	assert.Nil(t, findEvent(evs, "step_started", "install-control-plane-chart"), "no CP flow on a cluster-only install")
	assert.NotEmpty(t, helm.CallsByRelease("runai-cluster"), "cluster chart installs against the ready CP")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
}

// TestSystemNodes_PermissionDeniedFails verifies the one hard-failure path: when
// the installer must label nodes but lacks the patch-nodes permission, preflight
// fails and no install runs. RBAC's own required permissions are granted so the
// failure is isolated to the node-label permission.
func TestSystemNodes_PermissionDeniedFails(t *testing.T) {
	probe := &harness.FakeClusterProbe{
		Version: "v1.30.0",
		Nodes:   []orchestrator2.NodeInfo{worker("cpu-1", nil)},
		Allowed: map[string]bool{
			"create /namespaces": true,
			"create /secrets":    true,
			"create apiextensions.k8s.io/customresourcedefinitions": true,
			"create rbac.authorization.k8s.io/clusterroles":         true,
			"list storage.k8s.io/storageclasses":                    true,
			// "patch /nodes" intentionally absent -> denied
		},
	}
	deps, helm, labeler, _ := sysNodeDeps(probe)

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Prerequisites = &installconfig.Prerequisites{Skip: true}
	cfg.SystemNodes = &installconfig.SystemNodes{Enabled: true, Nodes: []string{"cpu-1"}}
	term, _, _ := newTestTerminal()

	err := specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts())
	require.ErrorIs(t, err, specrun.ErrPreflightFailed)

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "failed", preflightOutcome(t, evs))
	step := findEvent(evs, "step_succeeded", "preflight-system-nodes")
	require.NotNil(t, step)
	result := step["results"].([]any)[0].(map[string]any)
	assert.Equal(t, "fail", result["status"])
	assert.Contains(t, result["reason"], "patch nodes")
	assert.False(t, hasPipelineEvent(evs, "pipeline_started", "runai"), "install must not start")
	assert.Empty(t, helm.Calls)
	assert.Empty(t, labeler.NodesCalls)
}

// TestSystemNodes_UnschedulableNodeWarns verifies a selection that matches only
// cordoned (unschedulable) nodes warns (nothing schedulable) but still proceeds.
func TestSystemNodes_UnschedulableNodeWarns(t *testing.T) {
	cordoned := worker("cpu-1", nil)
	cordoned.Unschedulable = true
	_, _, evs := runSystemNodesInstall(t, []orchestrator2.NodeInfo{cordoned}, func(c *installconfig.Config) {
		c.SystemNodes = &installconfig.SystemNodes{Enabled: true, Nodes: []string{"cpu-1"}}
	})

	assert.Equal(t, "warning", preflightOutcome(t, evs))
	step := findEvent(evs, "step_succeeded", "preflight-system-nodes")
	require.NotNil(t, step)
	result := step["results"].([]any)[0].(map[string]any)
	assert.Equal(t, "warn", result["status"])
	assert.Contains(t, result["reason"], "no nodes match", "a cordoned node is not counted as schedulable")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
}

// TestSystemNodes_CustomLabel verifies a custom label is used both for labeling
// the nodes and as the affinity target.
func TestSystemNodes_CustomLabel(t *testing.T) {
	const customLabel = "acme.io/runai-system"
	helm, labeler, _ := runSystemNodesInstall(t, []orchestrator2.NodeInfo{worker("cpu-1", nil)}, func(c *installconfig.Config) {
		c.SystemNodes = &installconfig.SystemNodes{Enabled: true, Nodes: []string{"cpu-1"}, Label: customLabel}
	})

	cp := helm.CallsByRelease("runai-backend")
	require.Len(t, cp, 1)
	assert.Contains(t, cp[0].Values, "global.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].preference.matchExpressions[0].key="+customLabel)

	require.Len(t, labeler.NodesCalls, 1)
	assert.Equal(t, customLabel, labeler.NodesCalls[0].Key)
}

// TestSystemNodes_DryRun verifies a dry-run does not label any node and renders
// the labeling step and the affinity values in the planned (skipped) output.
func TestSystemNodes_DryRun(t *testing.T) {
	deps, _, labeler, _ := sysNodeDeps(&harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true, Nodes: []orchestrator2.NodeInfo{worker("cpu-1", nil)}})

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, true) // dry-run
	cfg.Prerequisites = &installconfig.Prerequisites{Skip: true}
	cfg.SystemNodes = &installconfig.SystemNodes{Enabled: true, Nodes: []string{"cpu-1"}}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	labelStep := findEvent(evs, "step_skipped", "label-system-nodes")
	require.NotNil(t, labelStep)
	assert.Equal(t, "dry_run", labelStep["skipReason"])
	assert.Contains(t, labelStep["output"], "label the selected system nodes")
	assert.Empty(t, labeler.NodesCalls, "dry-run must not mutate node labels")

	cpStep := findEvent(evs, "step_skipped", "install-control-plane-chart")
	require.NotNil(t, cpStep)
	assert.Contains(t, cpStep["output"], systemLabel, "the planned CP chart shows the system-node affinity")
}
