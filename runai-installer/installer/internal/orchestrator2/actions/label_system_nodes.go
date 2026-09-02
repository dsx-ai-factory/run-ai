/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
)

// LabelSystemNodes labels the selected nodes (by name or label selector) with
// Label so the system-node affinity has targets. Control-plane nodes are
// excluded. The node-role label is applied with an empty value (the affinity
// matches with operator Exists).
type LabelSystemNodes struct {
	Names    []string
	Selector map[string]string
	Label    string
}

func (a LabelSystemNodes) Kind() string { return "node.label_system" }

func (a LabelSystemNodes) Args() map[string]any {
	return map[string]any{"label": a.Label, "names": a.Names, "selector": a.Selector}
}

func (a LabelSystemNodes) Describe() string {
	return "label the selected system nodes with " + a.Label
}

func (a LabelSystemNodes) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	if deps.NodeLabeler == nil {
		return orchestrator2.ActionResult{}, fmt.Errorf("node labeling is not supported in this context")
	}
	nodes, err := deps.ClusterProbe.ListNodes(ctx)
	if err != nil {
		return orchestrator2.ActionResult{}, fmt.Errorf("list nodes for system-node labeling: %w", err)
	}
	names := selectSystemNodeNames(nodes, a.Names, a.Selector)
	if len(names) == 0 {
		// Nothing to label is not fatal: preflight already warned, and the
		// affinity simply has no targets (a no-op). Surface a warning rather
		// than aborting the install.
		return orchestrator2.ActionResult{
			Status: orchestrator2.StatusWarn,
			Reason: "no nodes matched the system node selection; nothing was labeled",
		}, nil
	}
	return orchestrator2.ActionResult{}, deps.NodeLabeler.LabelNodes(ctx, names, a.Label, "")
}
