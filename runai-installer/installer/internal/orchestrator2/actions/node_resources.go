/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import "k8s.io/apimachinery/pkg/api/resource"

// This file holds the static, recommended minimum node resources the installer
// checks against in preflight, analogous to the version compatibility matrix.
// The numbers are placeholders meant to be tuned later; they are intentionally
// expressed per node role so control-plane and worker nodes can differ.

// NodeResourceKind identifies a node resource the installer sets a recommended
// minimum on.
type NodeResourceKind string

const (
	NodeResourceCPU    NodeResourceKind = "cpu"
	NodeResourceMemory NodeResourceKind = "memory"
	NodeResourceDisk   NodeResourceKind = "ephemeral-storage"
)

const (
	nodeRoleControlPlane = "control-plane"
	nodeRoleWorker       = "worker"
)

// nodeMinimums is the recommended minimum total (capacity) for a node role.
type nodeMinimums struct {
	CPU    string
	Memory string
	Disk   string
}

// recommendedNodeMinimums maps a node role to its recommended minimum total
// resources. Placeholder values — tune later.
var recommendedNodeMinimums = map[string]nodeMinimums{
	nodeRoleControlPlane: {CPU: "4", Memory: "16Gi", Disk: "50Gi"},
	nodeRoleWorker:       {CPU: "8", Memory: "24Gi", Disk: "80Gi"},
}

// minimumFor returns the recommended minimum for a role/resource. Roles without
// an explicit entry fall back to the worker minimums.
func minimumFor(role string, kind NodeResourceKind) resource.Quantity {
	m, ok := recommendedNodeMinimums[role]
	if !ok {
		m = recommendedNodeMinimums[nodeRoleWorker]
	}
	switch kind {
	case NodeResourceCPU:
		return resource.MustParse(m.CPU)
	case NodeResourceMemory:
		return resource.MustParse(m.Memory)
	case NodeResourceDisk:
		return resource.MustParse(m.Disk)
	}
	return resource.Quantity{}
}

// effectiveRole reduces a node's roles to the one whose minimums apply: a node
// that runs workloads (worker, combined control-plane+worker, or unlabeled) is
// held to the worker minimums; a node that is control-plane only is held to the
// control-plane minimums.
func effectiveRole(roles []string) string {
	hasControlPlane := false
	for _, r := range roles {
		switch r {
		case nodeRoleControlPlane, "master":
			hasControlPlane = true
		case nodeRoleWorker:
			return nodeRoleWorker
		}
	}
	if hasControlPlane {
		return nodeRoleControlPlane
	}
	return nodeRoleWorker
}
