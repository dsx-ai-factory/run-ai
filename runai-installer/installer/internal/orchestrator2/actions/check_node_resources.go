/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"fmt"
	"sort"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"k8s.io/apimachinery/pkg/api/resource"
)

// CheckNodeResources is a read-only preflight check that warns when nodes have
// less than the recommended total (capacity) of a single resource. It never
// fails the install: insufficient node sizing is surfaced as a warning only.
//
// When SystemNodeNames or SystemNodeSelector is set (the user configured
// systemNodes), only those nodes are checked; otherwise every node is.
type CheckNodeResources struct {
	Resource           NodeResourceKind
	SystemNodeNames    []string
	SystemNodeSelector map[string]string
}

func (a CheckNodeResources) Kind() string { return "preflight.node-" + string(a.Resource) }

func (a CheckNodeResources) Args() map[string]any {
	return map[string]any{"resource": string(a.Resource)}
}

func (a CheckNodeResources) Describe() string {
	return "verify each node meets the recommended " + a.label() + " (warning only)"
}

func (a CheckNodeResources) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	nodes, err := deps.ClusterProbe.ListNodes(ctx)
	if err != nil {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusWarn,
			Reason:      fmt.Sprintf("could not list nodes to check %s: %v", a.label(), err),
			Remediation: "grant the installer permission to list nodes so node sizing can be validated",
		}, nil
	}

	// When system nodes are configured, only those nodes host the Run:ai system
	// components, so scope the sizing check to them; otherwise check all nodes.
	if len(a.SystemNodeNames) > 0 || len(a.SystemNodeSelector) > 0 {
		nodes = scopeToSystemNodes(nodes, a.SystemNodeNames, a.SystemNodeSelector)
	}

	offenders := make([]map[string]any, 0)
	for _, n := range nodes {
		have := a.capacity(n)
		// A node that does not report this resource (e.g. ephemeral-storage is
		// not always present in a node's capacity) reads as zero; skip it rather
		// than flag a false "0 below recommended" shortfall.
		if have.IsZero() {
			continue
		}
		role := effectiveRole(n.Roles)
		minimum := minimumFor(role, a.Resource)
		if have.Cmp(minimum) < 0 {
			offenders = append(offenders, map[string]any{
				"node":        n.Name,
				"role":        role,
				"have":        a.format(have),
				"recommended": a.format(minimum),
			})
		}
	}
	if len(offenders) == 0 {
		return orchestrator2.ActionResult{}, nil
	}
	sort.Slice(offenders, func(i, j int) bool {
		return offenders[i]["node"].(string) < offenders[j]["node"].(string)
	})

	// The per-check warning is a single line; the full per-node breakdown is
	// carried in Data and aggregated into the preflight summary for pretty,
	// one-time rendering at the end.
	return orchestrator2.ActionResult{
		Status: orchestrator2.StatusWarn,
		Reason: fmt.Sprintf("%d node(s) below recommended %s", len(offenders), a.label()),
		Data:   map[string]any{"resource": a.label(), "nodes": offenders},
	}, nil
}

func (a CheckNodeResources) capacity(n orchestrator2.NodeInfo) resource.Quantity {
	switch a.Resource {
	case NodeResourceCPU:
		return n.CPU
	case NodeResourceMemory:
		return n.Memory
	case NodeResourceDisk:
		return n.Disk
	}
	return resource.Quantity{}
}

func (a CheckNodeResources) label() string {
	switch a.Resource {
	case NodeResourceCPU:
		return "CPU"
	case NodeResourceMemory:
		return "memory"
	case NodeResourceDisk:
		return "disk (ephemeral storage)"
	}
	return string(a.Resource)
}

// format renders a quantity for human messages: CPU as cores, memory/disk as
// whole GiB.
func (a CheckNodeResources) format(q resource.Quantity) string {
	if a.Resource == NodeResourceCPU {
		return q.String() + " CPU"
	}
	return fmt.Sprintf("%dGi", q.Value()/(1<<30))
}
