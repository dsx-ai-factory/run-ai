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

// NodeLabel labels every worker node with Key=Value. The fake-GPU prerequisite
// uses it to mark nodes as simulated GPU nodes before installing the fake chart.
type NodeLabel struct {
	Key   string
	Value string
}

func (a NodeLabel) Kind() string { return "node.label" }

func (a NodeLabel) Args() map[string]any {
	return map[string]any{"key": a.Key, "value": a.Value}
}

func (a NodeLabel) Describe() string {
	return fmt.Sprintf("kubectl label nodes (workers) %s=%s", a.Key, a.Value)
}

func (a NodeLabel) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	if deps.NodeLabeler == nil {
		return orchestrator2.ActionResult{}, fmt.Errorf("node labeling is not supported in this context")
	}
	_, err := deps.NodeLabeler.LabelWorkerNodes(ctx, a.Key, a.Value)
	return orchestrator2.ActionResult{}, err
}
