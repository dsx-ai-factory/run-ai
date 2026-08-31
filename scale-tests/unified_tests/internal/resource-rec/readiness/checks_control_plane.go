/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package readiness

import (
	"context"
	"fmt"
	"strings"

	node_pools "github.com/run-ai/runai/scale-tests/unified_tests/internal/node-pools"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Phase values as they appear on the wire. Compared as strings because FetchNodepools
// parses the raw body rather than using the generated client, which fails to unmarshal
// GetNodepools responses.
const (
	nodepoolPhaseReady = "Ready"
	projectPhaseReady  = "Ready"
	clusterStateReady  = "Connected"
)

func (c *Checker) checkNodepools(ctx context.Context) CheckResult {
	res := CheckResult{Check: "nodepools_ready", Status: StatusPass}

	all, err := node_pools.FetchNodepools(ctx, c.clients.ClusterServiceClient.NodePoolsAPI)
	if err != nil {
		return failf(res, "fetching nodepools: %v", err)
	}

	// Only the seeded pools count. Clusters also carry a "default" nodepool holding their
	// real nodes, and counting it here would let a missing seeded pool pass unnoticed —
	// which is exactly what the submitter must not submit into.
	var pools []node_pools.NodepoolInfo
	for _, p := range all {
		if strings.HasPrefix(p.Name, c.cfg.NodepoolPrefix) {
			pools = append(pools, p)
		}
	}

	for _, p := range pools {
		if p.Phase != nodepoolPhaseReady {
			res.Status = StatusFail
			res.Detail = append(res.Detail, fmt.Sprintf("nodepool %s phase=%q", p.Name, p.Phase))
		}
	}
	// Exact, not a floor. A leftover pool from an earlier run is scheduled against and
	// changes the result, so it has to be removed rather than tolerated.
	if len(pools) != c.cluster.NodepoolCount {
		names := make([]string, 0, len(pools))
		for _, p := range pools {
			names = append(names, p.Name)
		}
		res.Status = StatusFail
		res.Detail = append(res.Detail, fmt.Sprintf(
			"expected exactly %d nodepools with prefix %q, found %d: %v (all pools: %d)",
			c.cluster.NodepoolCount, c.cfg.NodepoolPrefix, len(pools), names, len(all)))
	}

	// Only KWOK nodes are counted. The cluster also carries real system nodes, whose
	// count is a provisioning input this binary is not given and must not guess at.
	wantNodes := c.cluster.NodepoolCount * c.cluster.NodesPerPool
	nodes := &corev1.NodeList{}
	if err := c.clients.KubeClient.List(ctx, nodes, client.MatchingLabels{"type": "kwok"}); err != nil {
		return failf(res, "listing kwok nodes: %v", err)
	}
	ready := 0
	for i := range nodes.Items {
		if nodeReady(&nodes.Items[i]) {
			ready++
		}
	}
	if ready < wantNodes {
		res.Status = StatusFail
		res.Detail = append(res.Detail, fmt.Sprintf(
			"expected %d ready kwok nodes (%d pools x %d), found %d ready of %d total",
			wantNodes, c.cluster.NodepoolCount, c.cluster.NodesPerPool, ready, len(nodes.Items)))
	}

	res.Summary = fmt.Sprintf("%d nodepools, %d/%d ready kwok nodes", len(pools), ready, wantNodes)
	return res
}

func (c *Checker) checkProjects(ctx context.Context) CheckResult {
	res := CheckResult{Check: "projects_ready", Status: StatusPass}

	if c.cluster.ClusterID == "" {
		return failf(res, "cluster ID not resolved")
	}

	projects, _, err := c.clients.OrgUnitServiceClient.ProjectsAPI.
		GetProjects(ctx).
		FilterBy([]string{
			fmt.Sprintf("clusterId==%s", c.cluster.ClusterID),
			fmt.Sprintf("name=@%s", c.cfg.ProjectPrefix),
		}).
		// Unlimited, so an extra prefixed project is visible. Limiting to the expected
		// count hides exactly the leftovers this check exists to catch.
		Execute()
	if err != nil {
		return failf(res, "fetching projects: %v", err)
	}

	found := len(projects.Projects)
	notReady := 0
	names := make([]string, 0, found)
	for _, p := range projects.Projects {
		names = append(names, p.Name)
		if phase := string(p.Status.GetPhase()); phase != projectPhaseReady {
			notReady++
			res.Status = StatusFail
			res.Detail = append(res.Detail, fmt.Sprintf("project %s phase=%q", p.Name, phase))
		}
	}
	// Exact, not a floor. The submitter round-robins over every prefixed project it finds,
	// so a leftover one takes real submissions and changes the result.
	if found != c.cfg.ProjectCount {
		res.Status = StatusFail
		res.Detail = append(res.Detail, fmt.Sprintf(
			"expected exactly %d projects with prefix %q, found %d: %v",
			c.cfg.ProjectCount, c.cfg.ProjectPrefix, found, names))
	}

	res.Summary = fmt.Sprintf("%d/%d projects with prefix %q ready",
		found-notReady, c.cfg.ProjectCount, c.cfg.ProjectPrefix)
	return res
}

func (c *Checker) checkClusterConnected(ctx context.Context) CheckResult {
	res := CheckResult{Check: "cluster_connected", Status: StatusPass}

	clusters, _, err := c.clients.ClusterServiceClient.ClustersAPI.GetClusters(ctx).Execute()
	if err != nil {
		return failf(res, "fetching clusters: %v", err)
	}
	for _, cl := range clusters {
		if c.cluster.ClusterID != "" && cl.Uuid != c.cluster.ClusterID {
			continue
		}
		state := "<unset>"
		if s := cl.Status.Get(); s != nil && s.State != nil {
			state = string(*s.State)
		}
		if state != clusterStateReady {
			res.Status = StatusFail
			res.Detail = append(res.Detail, fmt.Sprintf("cluster %s state=%q", cl.Name, state))
		}
		res.Summary = fmt.Sprintf("cluster %s state=%s", cl.Name, state)
		return res
	}

	return failf(res, "cluster %s not found in the control plane", c.cluster.ClusterID)
}

func nodeReady(n *corev1.Node) bool {
	for _, cond := range n.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}
