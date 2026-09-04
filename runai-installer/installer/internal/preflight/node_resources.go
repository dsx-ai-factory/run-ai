/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package preflight

import (
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/steps"
)

func init() {
	builders = append(builders,
		nodeResourceBuilder("preflight-node-cpu", "Node CPU capacity", actions.NodeResourceCPU),
		nodeResourceBuilder("preflight-node-memory", "Node memory capacity", actions.NodeResourceMemory),
		nodeResourceBuilder("preflight-node-disk", "Node disk capacity", actions.NodeResourceDisk),
	)
}

func nodeResourceBuilder(name, title string, kind actions.NodeResourceKind) builder {
	return func(s *spec.Spec, _ orchestrator2.Deps) orchestrator2.Step {
		act := actions.CheckNodeResources{Resource: kind}
		// Scope the sizing check to the configured system nodes; when none are
		// configured the action checks every node.
		if s.SystemNodes != nil {
			act.SystemNodeNames = s.SystemNodes.Nodes
			act.SystemNodeSelector = s.SystemNodes.Selector
		}
		return steps.NewActionStepTitled(name, title, act)
	}
}
