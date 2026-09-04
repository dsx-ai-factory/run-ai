/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/rs/zerolog"

	clusterservice "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

const defaultClusterWindow = "last_1h"

const (
	cmAllocatedGpu         = "allocatedGpu"
	cmUnallocatedGpu       = "unallocatedGpu"
	cmTotalGpu             = "totalGpu"
	cmGPUUtilization       = "gpuUtilization"
	cmGPUMemoryUtilization = "gpuMemoryUtilization"
	cmCPUUtilization       = "cpuUtilization"
	cmCPUMemoryUtilization = "cpuMemoryUtilization"
	cmAvgWorkloadWaitTime  = "avgWorkloadWaitTime"
	cmWorkloadsCount       = "workloadsCount"
)

var clusterMetricCatalog = map[string]clusterservice.MetricsType{
	cmAllocatedGpu:         clusterservice.METRICSTYPE_ALLOCATED_GPU,
	cmUnallocatedGpu:       clusterservice.METRICSTYPE_UNALLOCATED_GPU,
	cmTotalGpu:             clusterservice.METRICSTYPE_TOTAL_GPU,
	cmGPUUtilization:       clusterservice.METRICSTYPE_GPU_UTILIZATION,
	cmGPUMemoryUtilization: clusterservice.METRICSTYPE_GPU_MEMORY_UTILIZATION,
	cmCPUUtilization:       clusterservice.METRICSTYPE_CPU_UTILIZATION,
	cmCPUMemoryUtilization: clusterservice.METRICSTYPE_CPU_MEMORY_UTILIZATION,
	cmAvgWorkloadWaitTime:  clusterservice.METRICSTYPE_AVG_WORKLOAD_WAIT_TIME,
	cmWorkloadsCount:       clusterservice.METRICSTYPE_WORKLOADS_COUNT,
}

var defaultClusterMetrics = []string{cmAllocatedGpu, cmUnallocatedGpu, cmGPUUtilization}

type getClusterMetricsInput struct {
	ClusterID       string   `json:"clusterId"                 jsonschema:"Cluster ID to query."`
	NodePoolName    string   `json:"nodePoolName,omitempty"    jsonschema:"Scope the series to a single node pool. If omitted, spans the whole cluster."`
	MetricTypes     []string `json:"metricTypes,omitempty"     jsonschema:"Capacity (allocatedGpu, unallocatedGpu, totalGpu), utilization (gpu/gpuMemory/cpu/cpuMemory), scheduling (avgWorkloadWaitTime, workloadsCount)."`
	Window          string   `json:"window,omitempty"          jsonschema:"Convenience time range ending now. One of: last_1h, last_24h, last_7d, last_30d. Defaults to last_1h. Ignored when start and end are both set."`
	Start           string   `json:"start,omitempty"           jsonschema:"Explicit range start (RFC3339). Overrides window; must be paired with end."`
	End             string   `json:"end,omitempty"             jsonschema:"Explicit range end (RFC3339). Overrides window; must be paired with start."`
	NumberOfSamples int32    `json:"numberOfSamples,omitempty" jsonschema:"Number of evenly spaced points across the range. Defaults to 20, capped at 1000."`
}

func (in getClusterMetricsInput) Validate() error {
	if err := common.ValidateUUID("clusterId", in.ClusterID); err != nil {
		return err
	}
	for _, m := range in.MetricTypes {
		if _, ok := clusterMetricCatalog[m]; !ok {
			return fmt.Errorf("unknown metric type %q", m)
		}
	}
	if in.Window != "" {
		if _, ok := windowDurations[in.Window]; !ok {
			return fmt.Errorf("window must be one of: last_1h, last_24h, last_7d, last_30d")
		}
	}
	if (in.Start == "") != (in.End == "") {
		return fmt.Errorf("start and end must be provided together")
	}
	if err := common.ValidateTimeWindow(in.Start, in.End); err != nil {
		return err
	}
	if in.NumberOfSamples < 0 || in.NumberOfSamples > maxSamples {
		return fmt.Errorf("numberOfSamples must be between 0 and %d", maxSamples)
	}
	return nil
}

type clusterSeries struct {
	AllocatedGpu         []point `json:"allocatedGpu,omitempty"         jsonschema:"GPUs allocated cluster-wide over time, in device count."`
	UnallocatedGpu       []point `json:"unallocatedGpu,omitempty"       jsonschema:"Free (unallocated) GPUs over time, in device count."`
	TotalGpu             []point `json:"totalGpu,omitempty"             jsonschema:"Total GPUs over time, in device count."`
	GPUUtilization       []point `json:"gpuUtilization,omitempty"       jsonschema:"GPU utilization over time, in percent (0-100)."`
	GPUMemoryUtilization []point `json:"gpuMemoryUtilization,omitempty" jsonschema:"GPU memory utilization over time, in percent (0-100)."`
	CPUUtilization       []point `json:"cpuUtilization,omitempty"       jsonschema:"CPU utilization over time, in percent (0-100)."`
	CPUMemoryUtilization []point `json:"cpuMemoryUtilization,omitempty" jsonschema:"System memory utilization over time, in percent (0-100)."`
	AvgWorkloadWaitTime  []point `json:"avgWorkloadWaitTime,omitempty"  jsonschema:"Average workload scheduling wait time over time, in seconds."`
	WorkloadsCount       []point `json:"workloadsCount,omitempty"       jsonschema:"Number of workloads over time."`
}

type getClusterMetricsOutput struct {
	ClusterID       string         `json:"clusterId"         jsonschema:"The cluster queried."`
	NodePool        *string        `json:"nodePool"          jsonschema:"Node pool the series is scoped to, or null when cluster-wide."`
	TimeRange       timeRange      `json:"timeRange"         jsonschema:"The time range and sample count the series cover."`
	Series          *clusterSeries `json:"series,omitempty"  jsonschema:"Cluster-wide time-series per requested metric. Omitted when seriesTruncated."`
	SeriesTruncated bool           `json:"seriesTruncated"   jsonschema:"True when the request was too large and series were dropped; retry with a smaller numberOfSamples or fewer metricTypes."`
	PartialData     bool           `json:"partialData"       jsonschema:"True when the upstream returned partial results (some metrics failed); failed metrics are absent."`
}

func registerGetClusterMetrics(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[getClusterMetricsInput, getClusterMetricsOutput]{
		Name: "get_cluster_metrics",
		Description: "Get a cluster's resource capacity and consumption over time — free vs allocated GPUs, GPU/CPU utilization, and scheduling pressure. " +
			"Use when the user asks about cluster-wide capacity or saturation (e.g. \"how many GPUs are free this week?\", \"is the cluster GPU-saturated?\"). " +
			"Scope to one node pool with nodePoolName. " +
			"Use get_workload_metrics for one workload and list_workloads for per-workload allocated or requested GPU-count filters. This aggregate tool cannot identify individual workloads above a utilization threshold. " +
			"Do NOT call for a single project's/department's consumption (use get_org_unit_metrics), for configured quota (use list_*_resources), or for per-workload metrics.",
		ReadOnly: true,
		Handler: func(ctx context.Context, in getClusterMetricsInput) (getClusterMetricsOutput, error) {
			return getClusterMetricsRun(ctx, client, in)
		},
	})
}

func getClusterMetricsRun(ctx context.Context, client *runaiclient.Client, in getClusterMetricsInput) (getClusterMetricsOutput, error) {
	window := in.Window
	if window == "" {
		window = defaultClusterWindow
	}
	from, to := resolveTimeRange(window, in.Start, in.End)
	samples := in.NumberOfSamples
	if samples == 0 {
		samples = defaultSamples
	}
	metricNames := dedupe(in.MetricTypes)
	if len(metricNames) == 0 {
		metricNames = defaultClusterMetrics
	}

	out := getClusterMetricsOutput{
		ClusterID: in.ClusterID,
		NodePool:  common.StringPtrIfNonEmpty(in.NodePoolName),
		TimeRange: timeRange{From: common.TimestampString(&from), To: common.TimestampString(&to), NumberOfSamples: samples},
	}

	if int(samples)*len(metricNames) > maxResultPoints {
		out.SeriesTruncated = true
		return out, nil
	}

	apiTypes := make([]clusterservice.MetricsType, 0, len(metricNames))
	for _, m := range metricNames {
		apiTypes = append(apiTypes, clusterMetricCatalog[m])
	}

	req := client.Cluster.ClustersAPI.GetClusterMetrics(ctx, in.ClusterID).
		MetricType(apiTypes).Start(from).End(to).NumberOfSamples(samples)
	if in.NodePoolName != "" {
		req = req.NodepoolName(in.NodePoolName)
	}
	resp, httpResp, err := req.Execute()
	if err != nil {
		return getClusterMetricsOutput{}, fmt.Errorf("get cluster metrics %q: %w", in.ClusterID, runaimcp.MapHTTPError(httpResp, err))
	}
	if httpResp != nil && httpResp.StatusCode == 207 {
		out.PartialData = true
	}

	out.Series = buildClusterSeries(ctx, resp, metricNames)
	return out, nil
}

func buildClusterSeries(ctx context.Context, resp *clusterservice.GroupedMetricsResponse, requested []string) *clusterSeries {
	s := &clusterSeries{}
	if resp == nil {
		return s
	}
	byType := map[string][]clusterservice.GroupedMeasurementResponseValuesInner{}
	for _, m := range resp.Measurements {
		byType[m.Type] = append(byType[m.Type], m.Values...)
	}
	for _, name := range requested {
		pts := clusterPoints(ctx, byType[string(clusterMetricCatalog[name])], name)
		if len(pts) > 0 {
			assignClusterSeries(s, name, pts)
		}
	}
	return s
}

func clusterPoints(ctx context.Context, values []clusterservice.GroupedMeasurementResponseValuesInner, metric string) []point {
	pts := make([]point, 0, len(values))
	dropped := 0
	for _, v := range values {
		ts := v.Timestamp.Get()
		if ts == nil {
			dropped++
			continue
		}
		f, err := strconv.ParseFloat(v.Value, 64)
		if err != nil || !common.IsFinite(f) {
			dropped++
			continue
		}
		pts = append(pts, point{T: ts.UTC().Format(time.RFC3339), V: common.RoundTo(f, 3)})
	}
	if dropped > 0 {
		zerolog.Ctx(ctx).Debug().Int("dropped", dropped).Str("metric", metric).
			Msg("get_cluster_metrics: skipped unparseable measurement points")
	}
	sort.Slice(pts, func(i, j int) bool { return pts[i].T < pts[j].T })
	return pts
}

func assignClusterSeries(s *clusterSeries, name string, pts []point) {
	switch name {
	case cmAllocatedGpu:
		s.AllocatedGpu = pts
	case cmUnallocatedGpu:
		s.UnallocatedGpu = pts
	case cmTotalGpu:
		s.TotalGpu = pts
	case cmGPUUtilization:
		s.GPUUtilization = pts
	case cmGPUMemoryUtilization:
		s.GPUMemoryUtilization = pts
	case cmCPUUtilization:
		s.CPUUtilization = pts
	case cmCPUMemoryUtilization:
		s.CPUMemoryUtilization = pts
	case cmAvgWorkloadWaitTime:
		s.AvgWorkloadWaitTime = pts
	case cmWorkloadsCount:
		s.WorkloadsCount = pts
	}
}
