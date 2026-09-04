/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install_v2_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/installconfig"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/specrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/resource"
)

// undersizedWorker is a worker node below every worker minimum, so it is flagged
// by the node-sizing checks (unless scoped out).
func undersizedWorker(name string, labels map[string]string) orchestrator2.NodeInfo {
	return orchestrator2.NodeInfo{
		Name:   name,
		Roles:  []string{"worker"},
		CPU:    resource.MustParse("2"),
		Memory: resource.MustParse("8Gi"),
		Disk:   resource.MustParse("20Gi"),
		Labels: labels,
	}
}

// flaggedResourceNodes returns the set of node names that appear in the preflight
// summary's resource issues.
func flaggedResourceNodes(t *testing.T, evs []map[string]any) map[string]bool {
	t.Helper()
	summary := findPreflightCompleted(evs)["preflight"].(map[string]any)
	out := map[string]bool{}
	if issues, ok := summary["resourceIssues"].([]any); ok {
		for _, raw := range issues {
			out[raw.(map[string]any)["node"].(string)] = true
		}
	}
	return out
}

func runNodeScopeInstall(t *testing.T, nodes []orchestrator2.NodeInfo, systemNodes *installconfig.SystemNodes) []map[string]any {
	t.Helper()
	probe := &harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true, Nodes: nodes}
	deps, _, _, _ := depsWithProbe(probe)

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath

	cfg := newInstallRunConfig()
	cfg.SystemNodes = systemNodes
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))
	return readEvents(t, eventsPath)
}

func boolPtr(b bool) *bool { return &b }

// TestPreflight_NodeResources_ScopedByName verifies that with systemNodes.nodes
// set, only the named nodes are sizing-checked; others are ignored.
func TestPreflight_NodeResources_ScopedByName(t *testing.T) {
	nodes := []orchestrator2.NodeInfo{
		undersizedWorker("sys-1", nil),
		undersizedWorker("gpu-1", nil),
	}
	evs := runNodeScopeInstall(t, nodes, &installconfig.SystemNodes{
		Enabled:      true,
		Nodes:        []string{"sys-1"},
		ManageLabels: boolPtr(false),
	})

	assert.Equal(t, map[string]bool{"sys-1": true}, flaggedResourceNodes(t, evs),
		"only the selected system node is sizing-checked")
}

// TestPreflight_NodeResources_ScopedBySelector verifies that with
// systemNodes.selector set, only nodes matching the selector are sizing-checked.
func TestPreflight_NodeResources_ScopedBySelector(t *testing.T) {
	nodes := []orchestrator2.NodeInfo{
		undersizedWorker("sys-1", map[string]string{"node-role": "runai-system"}),
		undersizedWorker("gpu-1", map[string]string{"node-role": "gpu"}),
	}
	evs := runNodeScopeInstall(t, nodes, &installconfig.SystemNodes{
		Enabled:      true,
		Selector:     map[string]string{"node-role": "runai-system"},
		ManageLabels: boolPtr(false),
	})

	assert.Equal(t, map[string]bool{"sys-1": true}, flaggedResourceNodes(t, evs),
		"only nodes matching the selector are sizing-checked")
}

// TestPreflight_NodeResources_AllNodesWhenNoSystemNodes verifies the unchanged
// default: with no systemNodes configured, every node is sizing-checked.
func TestPreflight_NodeResources_AllNodesWhenNoSystemNodes(t *testing.T) {
	nodes := []orchestrator2.NodeInfo{
		undersizedWorker("sys-1", nil),
		undersizedWorker("gpu-1", nil),
	}
	evs := runNodeScopeInstall(t, nodes, nil)

	assert.Equal(t, map[string]bool{"sys-1": true, "gpu-1": true}, flaggedResourceNodes(t, evs),
		"without systemNodes, all nodes are sizing-checked")
}
