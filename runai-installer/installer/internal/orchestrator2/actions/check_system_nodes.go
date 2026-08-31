/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
)

// CheckSystemNodes is a read-only preflight check for the system-node placement.
// When Configured is false it is the advisory: it warns that no system nodes are
// selected (the builder only assembles it for GPU installs). When Configured is
// true it validates the selection — always advisory (warn, never fail): it
// warns when named nodes do not exist or when the selection matches nothing.
// The only hard failure is a missing node-patch permission when the installer
// must label nodes (otherwise labeling cannot proceed).
type CheckSystemNodes struct {
	Configured   bool
	Names        []string
	Selector     map[string]string
	ManageLabels bool
}

func (a CheckSystemNodes) Kind() string { return "preflight.system_nodes" }

func (a CheckSystemNodes) Args() map[string]any {
	return map[string]any{"configured": a.Configured}
}

func (a CheckSystemNodes) Describe() string {
	if !a.Configured {
		return "advise selecting system nodes so Run:ai system pods avoid GPU nodes"
	}
	return "verify the selected system nodes can host the Run:ai system pods"
}

func (a CheckSystemNodes) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	if !a.Configured {
		return orchestrator2.ActionResult{
			Status: orchestrator2.StatusWarn,
			Reason: "no system nodes selected",
			Data: map[string]any{
				"systemNodesAdvisory": "No system nodes selected — Run:ai system pods may be scheduled onto GPU nodes. Pin the control plane and cluster services to non-GPU nodes via 'systemNodes' (recommended for production).",
			},
		}, nil
	}

	nodes, err := deps.ClusterProbe.ListNodes(ctx)
	if err != nil {
		return orchestrator2.ActionResult{
			Status: orchestrator2.StatusWarn,
			Reason: fmt.Sprintf("could not list nodes to validate system node selection: %v", err),
		}, nil
	}

	// A named node that does not exist is almost always a typo; warn (listing
	// the names) but never block — the install proceeds and any valid names
	// still take effect.
	if missing := missingNamedNodes(nodes, a.Names); len(missing) > 0 {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusWarn,
			Reason:      "system node(s) not found: " + strings.Join(missing, ", "),
			Remediation: "check the node names in systemNodes.nodes",
		}, nil
	}

	schedulable := 0
	for _, name := range selectSystemNodeNames(nodes, a.Names, a.Selector) {
		if !nodeUnschedulable(nodes, name) {
			schedulable++
		}
	}
	if schedulable == 0 {
		return orchestrator2.ActionResult{
			Status: orchestrator2.StatusWarn,
			Reason: "no nodes match the system node selection; the affinity will have no effect",
		}, nil
	}

	if a.ManageLabels {
		allowed, _, err := deps.ClusterProbe.CanI(ctx, orchestrator2.ResourceAttributes{Verb: "patch", Resource: "nodes"})
		if err != nil {
			return orchestrator2.ActionResult{
				Status: orchestrator2.StatusWarn,
				Reason: fmt.Sprintf("could not evaluate node-label permission: %v", err),
			}, nil
		}
		if !allowed {
			return orchestrator2.ActionResult{
				Status:      orchestrator2.StatusFail,
				Reason:      "missing permission to label nodes (patch nodes), required to mark the selected system nodes",
				Remediation: "grant the installer permission to patch nodes, or pre-label the nodes and set systemNodes.manageLabels=false",
			}, nil
		}
	}
	return orchestrator2.ActionResult{}, nil
}

func nodeUnschedulable(nodes []orchestrator2.NodeInfo, name string) bool {
	for _, n := range nodes {
		if n.Name == name {
			return n.Unschedulable
		}
	}
	return false
}
