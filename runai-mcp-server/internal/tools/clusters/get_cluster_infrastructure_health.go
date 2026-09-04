/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clusters

import (
	"context"
	"fmt"
	"sort"
	"time"

	clusterservice "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

const unhealthyNodesPageSize = 200

type getClusterInfrastructureHealthInput struct {
	ClusterID    string `json:"clusterId"              jsonschema:"Cluster ID (UUID)."`
	NodePoolName string `json:"nodePoolName,omitempty" jsonschema:"Scope the whole output to one node pool. If omitted, the full cluster is reported."`
}

func (in getClusterInfrastructureHealthInput) Validate() error {
	if err := common.ValidateUUID("clusterId", in.ClusterID); err != nil {
		return err
	}
	// CSV-joined filterBy: a comma in the pool name would smuggle a predicate.
	return common.ValidateNoCommas("nodePoolName", in.NodePoolName)
}

type nodeGPUs struct {
	Model string `json:"model" jsonschema:"Raw GPU product string as reported by the node (e.g. \"Tesla-V100\", \"NVIDIA-A100-SXM4-80GB\")."`
	Count int    `json:"count" jsonschema:"GPU device count on this node."`
}

type nodeCondition struct {
	Type    string `json:"type"              jsonschema:"k8s-style condition type, verbatim. Common: Ready, DiskPressure, MemoryPressure, PIDPressure, NetworkUnavailable, UnschedulableCordoned. No enforced vocabulary."`
	Reason  string `json:"reason"            jsonschema:"Short reason for the condition's last transition."`
	Message string `json:"message,omitempty" jsonschema:"Human-readable detail; omitted when the upstream returns null."`
}

type nodeTaint struct {
	Key    string `json:"key"              jsonschema:"k8s taint key. May be intentional (e.g. node-role.kubernetes.io/control-plane) or Run:ai-applied (e.g. runai=drain from runai cordon)."`
	Value  string `json:"value,omitempty"  jsonschema:"k8s taint value; omitted when the upstream returns null."`
	Effect string `json:"effect"           jsonschema:"k8s taint effect — one of NoSchedule, PreferNoSchedule, NoExecute."`
}

type unhealthyNodeRow struct {
	Name       string          `json:"name"                 jsonschema:"Node name (k8s-side identifier)."`
	NodePool   string          `json:"nodePool"             jsonschema:"Node pool the node belongs to."`
	Status     string          `json:"status"               jsonschema:"Run:ai-calculated node status. In this list it is one of: NotReady (kubelet reports unhealthy) or Unknown (control plane lost contact with the node)."`
	GPUs       *nodeGPUs       `json:"gpus,omitempty"       jsonschema:"GPU info on this node; omitted when the node carries no GPUs."`
	Conditions []nodeCondition `json:"conditions,omitempty" jsonschema:"k8s node conditions reported on this node, verbatim. Empty when the upstream reports none."`
	Taints     []nodeTaint     `json:"taints,omitempty"     jsonschema:"k8s taints on this node. Includes intentional admin taints (e.g. control-plane:NoSchedule) — surfaced as context, not a verdict of brokenness."`
}

type healthTotals struct {
	NodesTotal        int `json:"nodesTotal"        jsonschema:"Total nodes in scope (cluster-wide, or single-pool when scoped). Denominator for the unhealthy count."`
	NodesUnhealthy    int `json:"nodesUnhealthy"    jsonschema:"Number of nodes whose status is not Ready. Equals len(unhealthyNodes)."`
	GPUsUnschedulable int `json:"gpusUnschedulable" jsonschema:"Sum of gpuCount over all unhealthy nodes that carry GPUs. GPUs on Ready nodes are NOT included (this tool does not see per-GPU health on Ready hosts)."`
}

type healthByNodePool struct {
	NodePool          string `json:"nodePool"          jsonschema:"Node pool name."`
	NodesUnhealthy    int    `json:"nodesUnhealthy"    jsonschema:"Unhealthy nodes in this pool."`
	GPUsUnschedulable int    `json:"gpusUnschedulable" jsonschema:"Sum of gpuCount over this pool's unhealthy GPU-bearing nodes."`
}

type getClusterInfrastructureHealthOutput struct {
	ClusterID      string             `json:"clusterId"      jsonschema:"The cluster queried."`
	NodePool       *string            `json:"nodePool"       jsonschema:"Node pool the output is scoped to, or null when cluster-wide."`
	AsOf           string             `json:"asOf"           jsonschema:"Tool invocation timestamp (RFC3339, UTC). The upstream endpoints carry no native timestamp."`
	Totals         healthTotals       `json:"totals"         jsonschema:"Cluster-wide (or single-pool) health totals."`
	ByNodePool     []healthByNodePool `json:"byNodePool"     jsonschema:"One row per node pool that has at least one unhealthy node. Healthy pools are omitted."`
	UnhealthyNodes []unhealthyNodeRow `json:"unhealthyNodes" jsonschema:"Per-node detail for every node whose status is not Ready. Sorted by nodePool, then name."`
}

func registerGetClusterInfrastructureHealth(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[getClusterInfrastructureHealthInput, getClusterInfrastructureHealthOutput]{
		Name: "get_cluster_infrastructure_health",
		Description: "Get a cluster's current node-level infrastructure health: nodes that are NotReady or Unknown (control plane lost contact), with their k8s conditions, taints, and GPU counts. " +
			"Use when the user asks \"is anything broken?\", \"any unhealthy nodes?\", \"why isn't this pool scheduling?\", or \"how many GPU nodes are offline?\". " +
			"Returns a point-in-time snapshot: counts (cluster + per-pool) plus a small list of unhealthy nodes. " +
			"This tool surfaces NODE-level state only — it will NOT detect a failing GPU on an otherwise-Ready node (use get_workload_status for per-workload symptoms). " +
			"It is point-in-time and CANNOT answer historical questions like \"GPU failures last hour\".",
		ReadOnly: true,
		Handler: func(ctx context.Context, in getClusterInfrastructureHealthInput) (getClusterInfrastructureHealthOutput, error) {
			return getClusterInfrastructureHealthRun(ctx, client, in)
		},
	})
}

func getClusterInfrastructureHealthRun(ctx context.Context, client *runaiclient.Client, in getClusterInfrastructureHealthInput) (getClusterInfrastructureHealthOutput, error) {
	out := getClusterInfrastructureHealthOutput{
		ClusterID:      in.ClusterID,
		AsOf:           time.Now().UTC().Format(time.RFC3339),
		ByNodePool:     []healthByNodePool{},
		UnhealthyNodes: []unhealthyNodeRow{},
	}
	if in.NodePoolName != "" {
		out.NodePool = &in.NodePoolName
	}

	total, err := fetchNodesTotal(ctx, client, in)
	if err != nil {
		return getClusterInfrastructureHealthOutput{}, err
	}
	out.Totals.NodesTotal = total

	unhealthy, err := fetchUnhealthyNodes(ctx, client, in)
	if err != nil {
		return getClusterInfrastructureHealthOutput{}, err
	}
	out.UnhealthyNodes = unhealthy
	out.Totals.NodesUnhealthy = len(unhealthy)
	for _, n := range unhealthy {
		if n.GPUs != nil {
			out.Totals.GPUsUnschedulable += n.GPUs.Count
		}
	}
	out.ByNodePool = rollupByPool(unhealthy)

	return out, nil
}

func fetchNodesTotal(ctx context.Context, client *runaiclient.Client, in getClusterInfrastructureHealthInput) (int, error) {
	req := client.Cluster.NodesAPI.CountNodes(ctx, in.ClusterID)
	if in.NodePoolName != "" {
		req = req.FilterBy([]string{common.FilterEq("nodePool", in.NodePoolName)})
	}
	resp, httpResp, err := req.Execute()
	if err != nil {
		return 0, fmt.Errorf("count nodes for cluster %q: %w", in.ClusterID, runaimcp.MapHTTPError(httpResp, err))
	}
	if resp == nil {
		return 0, nil
	}
	return int(resp.Count), nil
}

func fetchUnhealthyNodes(ctx context.Context, client *runaiclient.Client, in getClusterInfrastructureHealthInput) ([]unhealthyNodeRow, error) {
	filters := []string{"status!=Ready", common.FilterEq("clusterUuid", in.ClusterID)}
	if in.NodePoolName != "" {
		filters = append(filters, common.FilterEq("nodePool", in.NodePoolName))
	}

	rows := []unhealthyNodeRow{}
	for offset := int32(0); ; {
		req := client.Cluster.NodesAPI.GetTenantNodes(ctx).
			FilterBy(filters).
			Limit(unhealthyNodesPageSize).
			Offset(offset)
		page, httpResp, err := req.Execute()
		if err != nil {
			return nil, fmt.Errorf("list unhealthy nodes for cluster %q: %w", in.ClusterID, runaimcp.MapHTTPError(httpResp, err))
		}
		if page == nil {
			break
		}
		for _, n := range page.Nodes {
			rows = append(rows, mapUnhealthyNode(n))
		}
		if int32(len(page.Nodes)) < unhealthyNodesPageSize {
			break
		}
		offset += unhealthyNodesPageSize
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].NodePool != rows[j].NodePool {
			return rows[i].NodePool < rows[j].NodePool
		}
		return rows[i].Name < rows[j].Name
	})
	return rows, nil
}

func mapUnhealthyNode(n clusterservice.Node) unhealthyNodeRow {
	row := unhealthyNodeRow{
		Name:     n.Name,
		NodePool: n.NodePool,
		Status:   n.Status,
	}
	if gi := n.GpuInfo.Get(); gi != nil && gi.GpuCount > 0 {
		row.GPUs = &nodeGPUs{Model: gi.GpuType, Count: int(gi.GpuCount)}
	}
	for _, c := range n.Conditions {
		nc := nodeCondition{Type: c.Type, Reason: c.Reason}
		if c.Message != nil {
			nc.Message = *c.Message
		}
		row.Conditions = append(row.Conditions, nc)
	}
	for _, t := range n.Taints {
		nt := nodeTaint{Key: t.Key, Effect: t.Effect}
		if t.Value != nil {
			nt.Value = *t.Value
		}
		row.Taints = append(row.Taints, nt)
	}
	return row
}

func rollupByPool(rows []unhealthyNodeRow) []healthByNodePool {
	idx := map[string]*healthByNodePool{}
	var order []string
	for _, r := range rows {
		p, ok := idx[r.NodePool]
		if !ok {
			p = &healthByNodePool{NodePool: r.NodePool}
			idx[r.NodePool] = p
			order = append(order, r.NodePool)
		}
		p.NodesUnhealthy++
		if r.GPUs != nil {
			p.GPUsUnschedulable += r.GPUs.Count
		}
	}
	sort.Strings(order)
	out := make([]healthByNodePool, 0, len(order))
	for _, name := range order {
		out = append(out, *idx[name])
	}
	return out
}
