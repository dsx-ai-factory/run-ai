/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"

	wlclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

const (
	bucketRunning   = "running"
	bucketPending   = "pending"
	bucketFailed    = "failed"
	bucketCompleted = "completed"
	bucketStopped   = "stopped"
	bucketOther     = "other"
)

var phaseBuckets = map[wlclient.Phase]string{
	wlclient.PHASE_RUNNING:      bucketRunning,
	wlclient.PHASE_PENDING:      bucketPending,
	wlclient.PHASE_INITIALIZING: bucketPending,
	wlclient.PHASE_CREATING:     bucketPending,
	wlclient.PHASE_RESUMING:     bucketPending,
	wlclient.PHASE_FAILED:       bucketFailed,
	wlclient.PHASE_DEGRADED:     bucketFailed,
	wlclient.PHASE_COMPLETED:    bucketCompleted,
	wlclient.PHASE_STOPPED:      bucketStopped,
	wlclient.PHASE_STOPPING:     bucketStopped,
	wlclient.PHASE_DELETING:     bucketOther,
	wlclient.PHASE_TERMINATING:  bucketOther,
	wlclient.PHASE_UPDATING:     bucketOther,
	wlclient.PHASE_UNKNOWN:      bucketOther,
}

var bucketByPhaseLower = func() map[string]string {
	m := make(map[string]string, len(phaseBuckets))
	for p, b := range phaseBuckets {
		m[strings.ToLower(string(p))] = b
	}
	return m
}()

type getWorkloadsSummaryInput struct {
	ClusterID    string `json:"clusterId,omitempty"    jsonschema:"Scope counts to one cluster (UUID). Omit to span all clusters."`
	OrgType      string `json:"orgType,omitempty"      jsonschema:"Org-unit kind to scope to: project or department. Must accompany orgName."`
	OrgName      string `json:"orgName,omitempty"      jsonschema:"Project or department name; resolved to its ID. Must accompany orgType."`
	NodePoolName string `json:"nodePoolName,omitempty" jsonschema:"Scope counts to one node pool."`
	GPUIdleOnly  bool   `json:"gpuIdleOnly,omitempty"  jsonschema:"Count only running workloads whose idleAllocatedGpus value is non-zero, using the precise workload count endpoint."`
}

func (in getWorkloadsSummaryInput) Validate() error {
	if in.ClusterID != "" {
		if err := common.ValidateUUID("clusterId", in.ClusterID); err != nil {
			return err
		}
	}
	if (in.OrgType == "") != (in.OrgName == "") {
		return fmt.Errorf("orgType and orgName must be provided together")
	}
	if in.OrgType != "" && in.OrgType != common.OrgTypeProject && in.OrgType != common.OrgTypeDepartment {
		return fmt.Errorf("orgType must be one of: %s, %s", common.OrgTypeProject, common.OrgTypeDepartment)
	}
	if err := common.ValidateFilterValue("nodePoolName", in.NodePoolName); err != nil {
		return err
	}

	return common.ValidateNoCommas("orgName", in.OrgName)
}

type workloadsSummaryScope struct {
	ClusterID   string `json:"clusterId,omitempty" jsonschema:"Cluster the counts are scoped to, when set."`
	OrgType     string `json:"orgType,omitempty"   jsonschema:"Org-unit kind the counts are scoped to (project | department), when set."`
	OrgName     string `json:"orgName,omitempty"   jsonschema:"Org-unit name the counts are scoped to, when set."`
	OrgUnitID   string `json:"orgUnitId,omitempty" jsonschema:"Resolved ID of the scoped org unit, when set."`
	NodePool    string `json:"nodePool,omitempty"  jsonschema:"Node pool the counts are scoped to, when set."`
	GPUIdleOnly bool   `json:"gpuIdleOnly,omitempty" jsonschema:"True when counts include only running workloads whose idleAllocatedGpus value is non-zero."`
}

type workloadCounts struct {
	Running   int `json:"running"   jsonschema:"Workloads actively executing (phase Running)."`
	Pending   int `json:"pending"   jsonschema:"Workloads waiting to run or spinning up (Pending, Initializing, Creating, Resuming)."`
	Failed    int `json:"failed"    jsonschema:"Workloads needing attention (Failed, Degraded)."`
	Completed int `json:"completed" jsonschema:"Workloads that finished successfully (Completed)."`
	Stopped   int `json:"stopped"   jsonschema:"Workloads intentionally suspended (Stopped, Stopping)."`
	Other     int `json:"other"     jsonschema:"Transitional or unrecognized phases (Deleting, Terminating, Updating, Unknown)."`
}

func (c *workloadCounts) add(bucket string, n int) {
	switch bucket {
	case bucketRunning:
		c.Running += n
	case bucketPending:
		c.Pending += n
	case bucketFailed:
		c.Failed += n
	case bucketCompleted:
		c.Completed += n
	case bucketStopped:
		c.Stopped += n
	default:
		c.Other += n
	}
}

type getWorkloadsSummaryOutput struct {
	Scope   workloadsSummaryScope `json:"scope"   jsonschema:"The scope the counts cover."`
	AsOf    string                `json:"asOf"    jsonschema:"Snapshot timestamp (RFC3339, UTC). Counts are current state, not a time range."`
	Total   int                   `json:"total"   jsonschema:"Total workloads across all phases in scope."`
	Counts  workloadCounts        `json:"counts"  jsonschema:"Counts folded into running/pending/failed/completed/stopped/other; reconciles to total."`
	ByPhase map[string]int        `json:"byPhase" jsonschema:"Raw count per workload phase (zero phases omitted); the buckets in counts are derived from this."`
}

func registerGetWorkloadsSummary(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[getWorkloadsSummaryInput, getWorkloadsSummaryOutput]{
		Name: "get_workloads_summary",
		Description: "Counts of workloads bucketed by phase for a given scope — a cluster, a project/department, a node pool, or the whole platform. " +
			"Running, pending (backlog), failed, and more. " +
			"Use gpuIdleOnly for how many workloads are GPU-idle: this counts Running workloads whose idleAllocatedGpus value is non-zero, using the precise Count Workloads API without listing workload pages. " +
			"Use when the user asks HOW MANY workloads are in some state, e.g. \"how many jobs are running cluster-wide?\", " +
			"\"is there a scheduling backlog?\", \"how many workloads does project alpha have running?\", \"any failed jobs in the research department?\". " +
			"Scope with clusterId, with an org unit (orgType+orgName), with nodePoolName, or any combination; omit all for the whole platform. " +
			"Returns a point-in-time snapshot. " +
			"Do NOT call when the user asks WHICH workloads match or needs workload names — use list_workloads. " +
			"Do NOT use this tool for allocatedGpu or requestedGpu comparisons; use list_workloads and follow all pages before counting. For actual GPU utilization use get_workload_metrics, get_cluster_metrics, or get_org_unit_metrics according to scope. " + //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
			"Do NOT call for why a workload is pending (use get_workload_status), for GPU/CPU metrics (use get_cluster_metrics / get_org_unit_metrics), or for quota (use list_project_resources / list_department_resources).",
		ReadOnly: true,
		Handler: func(ctx context.Context, in getWorkloadsSummaryInput) (getWorkloadsSummaryOutput, error) {
			return getWorkloadsSummaryRun(ctx, client, in)
		},
	})
}

func getWorkloadsSummaryRun(ctx context.Context, client *runaiclient.Client, in getWorkloadsSummaryInput) (getWorkloadsSummaryOutput, error) {
	scope := workloadsSummaryScope{ClusterID: in.ClusterID, NodePool: in.NodePoolName, GPUIdleOnly: in.GPUIdleOnly}
	orgUnitID := ""
	if in.OrgType != "" {
		id, err := common.ResolveOrgUnitID(ctx, client, in.OrgType, in.OrgName, in.ClusterID)
		if err != nil {
			return getWorkloadsSummaryOutput{}, err
		}
		orgUnitID = id
		scope.OrgType, scope.OrgName, scope.OrgUnitID = in.OrgType, in.OrgName, id
	}

	if in.GPUIdleOnly {
		return getGPUIdleWorkloadsSummary(ctx, client, in, scope, orgUnitID)
	}

	req := client.Workloads.WorkloadsAPI.GetWorkloadsTelemetry(ctx).
		TelemetryType(wlclient.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT).
		GroupBy([]string{"Phase"}) // group name is capitalized server-side; lowercase 400s
	if in.ClusterID != "" {
		req = req.ClusterId(in.ClusterID)
	}
	if in.NodePoolName != "" {
		req = req.NodepoolName(in.NodePoolName)
	}
	if orgUnitID != "" {
		if in.OrgType == common.OrgTypeProject {
			req = req.ProjectId(orgUnitID)
		} else {
			req = req.DepartmentId(orgUnitID)
		}
	}

	resp, httpResp, err := req.Execute()
	if err != nil {
		return getWorkloadsSummaryOutput{}, fmt.Errorf("get workloads summary: %w", runaimcp.MapHTTPError(httpResp, err))
	}

	out := getWorkloadsSummaryOutput{Scope: scope, ByPhase: map[string]int{}}
	if resp != nil {
		out.AsOf = resp.Timestamp.UTC().Format(time.RFC3339)
	}
	out.aggregate(ctx, resp)
	return out, nil
}

func getGPUIdleWorkloadsSummary(
	ctx context.Context,
	client *runaiclient.Client,
	in getWorkloadsSummaryInput,
	scope workloadsSummaryScope,
	orgUnitID string,
) (getWorkloadsSummaryOutput, error) {
	filters := []string{
		common.FilterEq("phase", string(wlclient.PHASE_RUNNING)),
		"idleAllocatedGpus!=0",
	}
	for _, fv := range []struct{ field, value string }{
		{"clusterId", in.ClusterID},
		{"nodepool", in.NodePoolName},
	} {
		if fv.value != "" {
			filters = append(filters, common.FilterEq(fv.field, fv.value))
		}
	}
	if orgUnitID != "" {
		field := "departmentId"
		if in.OrgType == common.OrgTypeProject {
			field = "projectId"
		}
		filters = append(filters, common.FilterEq(field, orgUnitID))
	}

	countValue, err := countWorkloads(ctx, client, filters, "", false)
	if err != nil {
		return getWorkloadsSummaryOutput{}, fmt.Errorf("count GPU-idle workloads: %w", err)
	}

	count := int(countValue)
	byPhase := map[string]int{}
	if count > 0 {
		byPhase[string(wlclient.PHASE_RUNNING)] = count
	}
	return getWorkloadsSummaryOutput{
		Scope:   scope,
		AsOf:    time.Now().UTC().Format(time.RFC3339),
		Total:   count,
		Counts:  workloadCounts{Running: count},
		ByPhase: byPhase,
	}, nil
}

func (out *getWorkloadsSummaryOutput) aggregate(ctx context.Context, resp *wlclient.TelemetryResponse) {
	if resp == nil {
		return
	}
	dropped := 0
	for _, v := range resp.Values {
		phase := phaseFromGroups(v.Groups)
		n, err := strconv.Atoi(strings.TrimSpace(v.Value))
		if phase == "" || err != nil || n < 0 {
			dropped++
			continue
		}
		out.Total += n
		out.ByPhase[phase] += n
		// Unknown / missing phase → empty bucket → falls through to Other in add().
		out.Counts.add(bucketByPhaseLower[strings.ToLower(phase)], n)
	}
	if dropped > 0 {
		zerolog.Ctx(ctx).Debug().Int("dropped", dropped).
			Msg("get_workloads_summary: skipped telemetry values with no phase or an unparseable count")
	}
}

func phaseFromGroups(groups []wlclient.TelemetryResponseValuesInnerGroupsInner) string {
	for _, g := range groups {
		if strings.EqualFold(g.Key, "phase") {
			return g.Value
		}
	}
	return ""
}
