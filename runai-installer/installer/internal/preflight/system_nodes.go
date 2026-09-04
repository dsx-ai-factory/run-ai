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

func init() { builders = append(builders, systemNodesBuilder) }

// systemNodesBuilder assembles the system-node check. When a selection is
// configured it validates it (strandedness + label permission). When it is not,
// it emits a soft advisory — but only for GPU installs, so CPU-only/fake-GPU
// installs stay silent (there is nothing to protect, and the preflight outcome
// remains "passed"). GPU intent is read from the spec (the GPU operator is not
// installed yet at preflight time, so the live cluster cannot be probed for it).
func systemNodesBuilder(s *spec.Spec, deps orchestrator2.Deps) orchestrator2.Step {
	if s.ControlPlane == nil && s.Cluster == nil {
		return nil
	}
	if s.SystemNodes == nil {
		if !gpuIntended(s) {
			return nil
		}
		return steps.NewActionStepTitled("preflight-system-nodes", "System node selection", actions.CheckSystemNodes{Configured: false})
	}
	if deps.ClusterProbe == nil {
		return nil
	}
	return steps.NewActionStepTitled("preflight-system-nodes", "System node selection", actions.CheckSystemNodes{
		Configured:   true,
		Names:        s.SystemNodes.Nodes,
		Selector:     s.SystemNodes.Selector,
		ManageLabels: s.SystemNodes.ManageLabels,
	})
}

// gpuIntended reports whether this install provisions real GPUs (the real GPU
// operator is being installed). Fake GPUs do not count — there are no real GPU
// nodes to dedicate.
func gpuIntended(s *spec.Spec) bool {
	p := s.Prerequisites
	return p != nil && !p.Skip && p.GPUOperator.IsInstall() && p.GPUType != spec.GPUTypeFake
}
