/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clusters

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"

	clusterservice "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

// Telemetry returns the group key as "Nodepool" (capitalized) — verified live.
const groupByNodepool = "Nodepool"

const outputDecimals = 3

type capacityMetric struct {
	telemetryType clusterservice.NodeTelemetryType
	setter        func(*physicalInventoryCell, float64)
}

var capacityMetrics = []capacityMetric{
	{telemetryType: clusterservice.NODETELEMETRYTYPE_TOTAL_GPUS, setter: func(c *physicalInventoryCell, v float64) { c.GPUsTotal = int(v) }},
	{telemetryType: clusterservice.NODETELEMETRYTYPE_READY_GPUS, setter: func(c *physicalInventoryCell, v float64) { c.GPUsAllocatable = int(v) }},
	{telemetryType: clusterservice.NODETELEMETRYTYPE_ALLOCATED_GPUS, setter: func(c *physicalInventoryCell, v float64) { c.GPUsAllocated = int(v) }},
	{telemetryType: clusterservice.NODETELEMETRYTYPE_FREE_GPUS, setter: func(c *physicalInventoryCell, v float64) { c.GPUsFree = int(v) }},
	{telemetryType: clusterservice.NODETELEMETRYTYPE_TOTAL_GPU_NODES, setter: func(c *physicalInventoryCell, v float64) { c.GPUNodes = int(v) }},
	{telemetryType: clusterservice.NODETELEMETRYTYPE_TOTAL_CPU_CORES, setter: func(c *physicalInventoryCell, v float64) { c.CPUCoresTotal = common.RoundTo(v, outputDecimals) }},
	{telemetryType: clusterservice.NODETELEMETRYTYPE_ALLOCATED_CPU_CORES, setter: func(c *physicalInventoryCell, v float64) { c.CPUCoresAllocated = common.RoundTo(v, outputDecimals) }},
	{telemetryType: clusterservice.NODETELEMETRYTYPE_TOTAL_CPU_MEMORY_BYTES, setter: func(c *physicalInventoryCell, v float64) {
		c.MemoryGiBTotal = common.RoundTo(v/common.BytesPerGiB, outputDecimals)
	}},
	{telemetryType: clusterservice.NODETELEMETRYTYPE_ALLOCATED_CPU_MEMORY_BYTES, setter: func(c *physicalInventoryCell, v float64) {
		c.MemoryGiBAllocated = common.RoundTo(v/common.BytesPerGiB, outputDecimals)
	}},
}

type getClusterPhysicalInventoryInput struct {
	ClusterID    string `json:"clusterId"              jsonschema:"Cluster ID (UUID) to inventory."`
	NodePoolName string `json:"nodePoolName,omitempty" jsonschema:"Scope the whole output to one node pool. If omitted, the full cluster is returned."`
}

func (in getClusterPhysicalInventoryInput) Validate() error {
	if err := common.ValidateUUID("clusterId", in.ClusterID); err != nil {
		return err
	}
	return common.ValidateNoCommas("nodePoolName", in.NodePoolName)
}

type physicalInventoryCell struct {
	GPUNodes           int     `json:"gpuNodes"           jsonschema:"Number of GPU-bearing nodes."`
	GPUsTotal          int     `json:"gpusTotal"          jsonschema:"Total GPU devices on the nodes."`
	GPUsAllocatable    int     `json:"gpusAllocatable"    jsonschema:"Schedulable GPU devices (READY_GPUS) — GPUs on Ready nodes available to be allocated."`
	GPUsAllocated      int     `json:"gpusAllocated"      jsonschema:"GPU devices with any allocation, whole or fractional."`
	GPUsFree           int     `json:"gpusFree"           jsonschema:"Fully idle GPU devices — no allocation, including fractional. With fractional workloads this can be less than gpusAllocatable - gpusAllocated."`
	CPUCoresTotal      float64 `json:"cpuCoresTotal"      jsonschema:"Total CPU cores."`
	CPUCoresAllocated  float64 `json:"cpuCoresAllocated"  jsonschema:"CPU cores allocated to workloads."`
	MemoryGiBTotal     float64 `json:"memoryGiBTotal"     jsonschema:"Total system memory in GiB."`
	MemoryGiBAllocated float64 `json:"memoryGiBAllocated" jsonschema:"System memory allocated to workloads, in GiB."`
}

type nodePoolInventory struct {
	NodePool string `json:"nodePool" jsonschema:"Node pool name."`
	physicalInventoryCell
}

type gpuModelInventory struct {
	GPUModel  string `json:"gpuModel"  jsonschema:"Raw GPU product string as reported by the node (e.g. \"Tesla-V100\", \"NVIDIA-A100-SXM4-80GB\")."`
	Nodes     int    `json:"nodes"     jsonschema:"Number of nodes carrying this GPU model."`
	GPUsTotal int    `json:"gpusTotal" jsonschema:"Total GPU devices of this model. Allocation/free are not available per-model."`
}

type getClusterPhysicalInventoryOutput struct {
	ClusterID  string                `json:"clusterId"            jsonschema:"The cluster queried."`
	NodePool   *string               `json:"nodePool"             jsonschema:"Node pool the output is scoped to, or null when cluster-wide."`
	AsOf       string                `json:"asOf"                 jsonschema:"Snapshot timestamp (RFC3339, UTC). Capacity is a point-in-time read."`
	Totals     physicalInventoryCell `json:"totals"               jsonschema:"Cluster-wide capacity totals (or single-pool totals when scoped)."`
	ByNodePool []nodePoolInventory   `json:"byNodePool"           jsonschema:"Capacity per node pool, from telemetry grouped by Nodepool."`
	ByGPUModel []gpuModelInventory   `json:"byGpuModel,omitempty" jsonschema:"GPU device totals grouped by raw model string, from the node list. Omitted when the node-list call fails (telemetry data is still returned)."`
}

func registerGetClusterPhysicalInventory(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[getClusterPhysicalInventoryInput, getClusterPhysicalInventoryOutput]{
		Name: "get_cluster_physical_inventory",
		Description: "Get a cluster's current hardware capacity: GPU nodes, GPU models, and GPUs / CPU cores / memory — total, allocatable, allocated, and free — " +
			"aggregated per node pool and per GPU model. " +
			"Use when the user asks \"what GPUs do we have?\", \"how many A100s are free?\", \"GPU capacity per node pool\", or \"how many GPU nodes?\". " +
			"Scope with clusterId, optionally narrow with nodePoolName. " +
			"Do NOT call for configured quota (use list_project_resources / list_department_resources), " +
			"for utilization over time (use get_cluster_metrics), or for node health/conditions.",
		ReadOnly: true,
		Handler: func(ctx context.Context, in getClusterPhysicalInventoryInput) (getClusterPhysicalInventoryOutput, error) {
			return getClusterPhysicalInventoryRun(ctx, client, in)
		},
	})
}

func getClusterPhysicalInventoryRun(ctx context.Context, client *runaiclient.Client, in getClusterPhysicalInventoryInput) (getClusterPhysicalInventoryOutput, error) {
	out := getClusterPhysicalInventoryOutput{ClusterID: in.ClusterID}
	if in.NodePoolName != "" {
		out.NodePool = &in.NodePoolName
	}

	cells, asOf, err := fetchCapacity(ctx, client, in)
	if err != nil {
		return getClusterPhysicalInventoryOutput{}, err
	}
	out.AsOf = asOf
	out.ByNodePool = sortedCells(cells)
	out.Totals = sumCells(out.ByNodePool)

	if models, err := fetchGPUModels(ctx, client, in); err == nil {
		out.ByGPUModel = models
	} else {
		zerolog.Ctx(ctx).Warn().Err(err).Str("cluster_id", in.ClusterID).
			Msg("get_cluster_physical_inventory: node-list fetch failed; returning without byGpuModel")
	}

	return out, nil
}

func fetchCapacity(ctx context.Context, client *runaiclient.Client, in getClusterPhysicalInventoryInput) (map[string]*physicalInventoryCell, string, error) {
	type result struct {
		metric capacityMetric
		resp   *clusterservice.TelemetryResponse
		err    error
	}

	results := make([]result, len(capacityMetrics))
	var wg sync.WaitGroup
	for i, m := range capacityMetrics {
		wg.Add(1)
		go func(i int, m capacityMetric) {
			defer wg.Done()
			req := client.Cluster.NodesAPI.GetNodeTelemetry(ctx).
				TelemetryType(m.telemetryType).
				ClusterId(in.ClusterID).
				GroupBy([]string{groupByNodepool})
			if in.NodePoolName != "" {
				req = req.NodepoolName(in.NodePoolName)
			}
			resp, httpResp, err := req.Execute()
			if err != nil {
				err = fmt.Errorf("get node telemetry %s: %w", m.telemetryType, runaimcp.MapHTTPError(httpResp, err))
			}
			results[i] = result{metric: m, resp: resp, err: err}
		}(i, m)
	}
	wg.Wait()

	for _, r := range results {
		if r.err != nil {
			return nil, "", r.err
		}
	}

	cells := map[string]*physicalInventoryCell{}
	var latest time.Time
	for _, r := range results {
		if r.resp == nil {
			continue
		}
		if r.resp.Timestamp.After(latest) {
			latest = r.resp.Timestamp
		}
		for _, v := range r.resp.Values {
			pool := groupValue(v.Groups, groupByNodepool)
			f, ok := common.ParseFiniteFloat(v.Value)
			if !ok || pool == "" {
				zerolog.Ctx(ctx).Debug().
					Str("metric", string(r.metric.telemetryType)).
					Str("raw_value", v.Value).
					Str("pool", pool).
					Msg("get_cluster_physical_inventory: skipped unparseable telemetry value")
				continue
			}
			cell, ok := cells[pool]
			if !ok {
				cell = &physicalInventoryCell{}
				cells[pool] = cell
			}
			r.metric.setter(cell, f)
		}
	}
	return cells, common.TimestampString(common.TimePtrIfNonZero(latest)), nil
}

func fetchGPUModels(ctx context.Context, client *runaiclient.Client, in getClusterPhysicalInventoryInput) ([]gpuModelInventory, error) {
	type bucket struct {
		nodes int
		gpus  int
	}
	agg := map[string]*bucket{}
	var scanned int32

	for offset := int32(0); ; {
		req := client.Cluster.NodesAPI.GetTenantNodes(ctx).
			Limit(common.UpstreamPageSize).
			Offset(offset)
		filters := []string{common.FilterEq("clusterUuid", in.ClusterID)}
		if in.NodePoolName != "" {
			filters = append(filters, common.FilterEq("nodePool", in.NodePoolName))
		}
		req = req.FilterBy(filters)
		page, httpResp, err := req.Execute()
		if err != nil {
			return nil, fmt.Errorf("list nodes for cluster %q: %w", in.ClusterID, runaimcp.MapHTTPError(httpResp, err))
		}
		if page == nil {
			break
		}
		for _, n := range page.Nodes {
			gi := n.GpuInfo.Get()
			if gi == nil {
				continue
			}
			b, ok := agg[gi.GpuType]
			if !ok {
				b = &bucket{}
				agg[gi.GpuType] = b
			}
			b.nodes++
			b.gpus += int(gi.GpuCount)
		}
		scanned += int32(len(page.Nodes))
		if int32(len(page.Nodes)) < common.UpstreamPageSize {
			break
		}
		offset += common.UpstreamPageSize
	}

	if scanned > 1000 {
		zerolog.Ctx(ctx).Info().
			Int32("nodes_scanned", scanned).
			Str("cluster_id", in.ClusterID).
			Msg("get_cluster_physical_inventory: large node-list scan; consider future paging optimization")
	}

	models := make([]gpuModelInventory, 0, len(agg))
	for model, b := range agg {
		models = append(models, gpuModelInventory{GPUModel: model, Nodes: b.nodes, GPUsTotal: b.gpus})
	}
	sort.Slice(models, func(i, j int) bool { return models[i].GPUModel < models[j].GPUModel })
	return models, nil
}

func sortedCells(cells map[string]*physicalInventoryCell) []nodePoolInventory {
	out := make([]nodePoolInventory, 0, len(cells))
	for name, c := range cells {
		out = append(out, nodePoolInventory{NodePool: name, physicalInventoryCell: *c})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].NodePool < out[j].NodePool })
	return out
}

func sumCells(cells []nodePoolInventory) physicalInventoryCell {
	var s physicalInventoryCell
	for _, c := range cells {
		s.GPUNodes += c.GPUNodes
		s.GPUsTotal += c.GPUsTotal
		s.GPUsAllocatable += c.GPUsAllocatable
		s.GPUsAllocated += c.GPUsAllocated
		s.GPUsFree += c.GPUsFree
		s.CPUCoresTotal += c.CPUCoresTotal
		s.CPUCoresAllocated += c.CPUCoresAllocated
		s.MemoryGiBTotal += c.MemoryGiBTotal
		s.MemoryGiBAllocated += c.MemoryGiBAllocated
	}

	s.CPUCoresTotal = common.RoundTo(s.CPUCoresTotal, outputDecimals)
	s.CPUCoresAllocated = common.RoundTo(s.CPUCoresAllocated, outputDecimals)
	s.MemoryGiBTotal = common.RoundTo(s.MemoryGiBTotal, outputDecimals)
	s.MemoryGiBAllocated = common.RoundTo(s.MemoryGiBAllocated, outputDecimals)
	return s
}

func groupValue(groups []clusterservice.TelemetryResponseValuesInnerGroupsInner, key string) string {
	for _, g := range groups {
		if strings.EqualFold(g.Key, key) {
			return g.Value
		}
	}
	return ""
}
