/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	wlclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

const workloadSortFieldOptions = "type, name, clusterId, projectId, projectName, departmentId, departmentName, createdAt, deletedAt, submittedBy, phase, completedAt, nodepool, nodePoolName, distributedFramework, allocatedGPU, allocatedGpu, idleAllocatedGpus, phaseUpdatedAt, category, priority, totalPendingTimeSeconds, totalRunningTimeSeconds, priorityClassName, preemptibility, guaranteedRuntimeEndsAt, aiApplicationId, aiApplicationName" //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag

const workloadFilterFieldOptions = "type, name, clusterId, projectId, projectName, departmentId, departmentName, createdAt, deletedAt, submittedBy, phase, completedAt, nodepool, distributedFramework, allocatedGPU, idleAllocatedGpus, phaseUpdatedAt, category, priority, totalPendingTimeSeconds, totalRunningTimeSeconds, priorityClassName, preemptibility, guaranteedRuntimeEndsAt, aiApplicationId, aiApplicationName" //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag

var workloadSortFields = map[string]struct{}{
	"type": {}, "name": {}, "clusterId": {}, "projectId": {}, "projectName": {},
	"departmentId": {}, "departmentName": {}, "createdAt": {}, "deletedAt": {},
	"submittedBy": {}, "phase": {}, "completedAt": {}, "nodepool": {},
	"distributedFramework": {}, "allocatedGPU": {}, "idleAllocatedGpus": {},
	"phaseUpdatedAt": {}, "category": {}, "priority": {}, "totalPendingTimeSeconds": {},
	"totalRunningTimeSeconds": {}, "priorityClassName": {}, "preemptibility": {},
	"guaranteedRuntimeEndsAt": {},
	"aiApplicationId":         {}, "aiApplicationName": {},
	// MCP-facing aliases for API spellings that differ from returned/input field names.
	"nodePoolName": {}, "allocatedGpu": {},
}

var workloadFilterKinds = map[string]string{
	"type": "string", "name": "string", "clusterId": "string", "projectId": "string",
	"projectName": "string", "departmentId": "string", "departmentName": "string",
	"createdAt": "timestamp", "deletedAt": "timestamp", "submittedBy": "string",
	"phase": "string", "completedAt": "timestamp", "nodepool": "string",
	"distributedFramework": "string", "allocatedGPU": "number", "idleAllocatedGpus": "number",
	"phaseUpdatedAt": "timestamp", "category": "string", "priority": "integer",
	"totalPendingTimeSeconds": "integer", "totalRunningTimeSeconds": "integer",
	"priorityClassName": "string", "preemptibility": "string", "guaranteedRuntimeEndsAt": "timestamp",
	"aiApplicationId": "string", "aiApplicationName": "string",
}

type workloadFilterComparison struct {
	Operator string `json:"operator" jsonschema:"Comparison operator. String fields support ==, !=, <=, >=, =@, !@, =^, =$. Numeric and timestamp fields also support strict > and <, translated to exact server-side API filters."` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	Value    string `json:"value"    jsonschema:"Value to compare. Timestamps must be RFC3339; integer values must be non-negative."`
}

func (c workloadFilterComparison) validate(field, kind string) error {
	allowed := map[string]bool{"==": true, "!=": true, "<": true, "<=": true, ">": true, ">=": true}
	if kind == "string" {
		allowed = map[string]bool{"==": true, "!=": true, "<=": true, ">=": true, "=@": true, "!@": true, "=^": true, "=$": true}
	}
	if !allowed[c.Operator] {
		return fmt.Errorf("%s.operator is not valid for a %s filter", field, kind)
	}
	if strings.TrimSpace(c.Value) == "" {
		return fmt.Errorf("%s.value must not be blank", field)
	}
	if err := common.ValidateFilterValue(field, c.Value); err != nil {
		return err
	}
	switch kind {
	case "timestamp":
		if _, err := time.Parse(time.RFC3339, c.Value); err != nil {
			return fmt.Errorf("%s.value must be RFC3339 (got %q)", field, c.Value)
		}
	case "integer":
		value, err := strconv.ParseInt(c.Value, 10, 32)
		if err != nil || value < 0 {
			return fmt.Errorf("%s.value must be a non-negative 32-bit integer", field)
		}
	case "number":
		value, err := strconv.ParseFloat(c.Value, 64)
		if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			return fmt.Errorf("%s.value must be a finite non-negative number", field)
		}
	}
	return nil
}

type apiWorkloadFilter struct {
	Field    string `json:"field"    jsonschema:"API field to filter. One of: type, name, clusterId, projectId, projectName, departmentId, departmentName, createdAt, deletedAt, submittedBy, phase, completedAt, nodepool, distributedFramework, allocatedGPU, idleAllocatedGpus, phaseUpdatedAt, category, priority, totalPendingTimeSeconds, totalRunningTimeSeconds, priorityClassName, preemptibility, guaranteedRuntimeEndsAt, aiApplicationId, aiApplicationName."` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	Operator string `json:"operator" jsonschema:"Comparison operator. String fields support ==, !=, <=, >=, =@, !@, =^, and =$. Numeric and timestamp fields support ==, !=, <, <=, >, and >=."`
	Value    string `json:"value"    jsonschema:"Comparison value. Timestamp fields require RFC3339; numeric fields require finite non-negative values."`
}

func (f apiWorkloadFilter) validate(index int) error {
	kind, ok := workloadFilterKinds[f.Field]
	if !ok {
		return fmt.Errorf("filterBy[%d].field must be one of: %s", index, workloadFilterFieldOptions)
	}
	return (workloadFilterComparison{Operator: f.Operator, Value: f.Value}).validate(fmt.Sprintf("filterBy[%d]", index), kind)
}

type gpuComparison struct {
	Operator string  `json:"operator" jsonschema:"Comparison operator. One of: >, >=, <, <=, ==, !=. Preserve the user's wording exactly: more than is >, while at least is >=."`
	Value    float64 `json:"value"    jsonschema:"GPU device count to compare against. Must be a finite non-negative number; fractional values are supported."`
}

func (c gpuComparison) Validate(field string) error {
	switch c.Operator {
	case ">", ">=", "<", "<=", "==", "!=":
	default:
		return fmt.Errorf("%s.operator must be one of: >, >=, <, <=, ==, !=", field)
	}
	if math.IsNaN(c.Value) || math.IsInf(c.Value, 0) || c.Value < 0 {
		return fmt.Errorf("%s.value must be a finite non-negative number", field)
	}
	return nil
}

type listWorkloadsInput struct {
	Name                    string                    `json:"name,omitempty"                jsonschema:"Exact workload name to match. Combine with projectId when resolving a workload ID."`
	ProjectID               string                    `json:"projectId,omitempty"           jsonschema:"Project ID to scope the workload list or disambiguate a workload name."`
	ClusterID               string                    `json:"clusterId,omitempty"           jsonschema:"Cluster UUID to scope the workload list."`
	Phase                   string                    `json:"phase,omitempty"               jsonschema:"Exact workload phase to match. One of: Creating, Initializing, Resuming, Pending, Deleting, Running, Updating, Stopped, Stopping, Degraded, Failed, Completed, Terminating, Unknown."` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	Type                    string                    `json:"type,omitempty"                jsonschema:"Exact workload type to match, such as Training, Workspace, Distributed, or Inference."`
	NodePoolName            string                    `json:"nodePoolName,omitempty"        jsonschema:"Current node pool name to match. Pending workloads without current placement do not match this filter."`
	CreatedAfter            string                    `json:"createdAfter,omitempty"        jsonschema:"Return workloads created at or after this RFC3339 timestamp."`
	PhaseUpdatedAfter       string                    `json:"phaseUpdatedAfter,omitempty"   jsonschema:"Return workloads whose phase last changed at or after this RFC3339 timestamp."`
	DeletedOnly             bool                      `json:"deletedOnly,omitempty"         jsonschema:"Return only deleted workloads. Use for requests to list deleted workloads."`
	DeletedAfter            string                    `json:"deletedAfter,omitempty"        jsonschema:"Return workloads deleted at or after this RFC3339 timestamp. For 'deleted in the last X hours', compute the current time minus X hours and pass that cutoff. This automatically selects only deleted workloads; deletedOnly does not also need to be set."` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	Search                  string                    `json:"search,omitempty"              jsonschema:"Free-text workload search passed to the List Workloads API."`
	SortBy                  string                    `json:"sortBy,omitempty"              jsonschema:"Field used to sort results. Defaults to phaseUpdatedAt. One of: type, name, clusterId, projectId, projectName, departmentId, departmentName, createdAt, deletedAt, submittedBy, phase, completedAt, nodepool, nodePoolName, distributedFramework, allocatedGPU, allocatedGpu, idleAllocatedGpus, phaseUpdatedAt, category, priority, totalPendingTimeSeconds, totalRunningTimeSeconds, priorityClassName, preemptibility, guaranteedRuntimeEndsAt, aiApplicationId, aiApplicationName."` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	SortOrder               string                    `json:"sortOrder,omitempty"           jsonschema:"Sort direction: asc or desc. Defaults to desc."`
	FilterBy                []apiWorkloadFilter       `json:"filterBy,omitempty"            jsonschema:"Additional typed List Workloads API filters. Use this for operators or fields not covered by the convenience inputs. All filters are ANDed. A deletedAt filter automatically selects only deleted workloads."` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	ProjectName             *workloadFilterComparison `json:"projectName,omitempty" jsonschema:"Filter by project name with a string comparison."`
	DepartmentID            *workloadFilterComparison `json:"departmentId,omitempty" jsonschema:"Filter by department ID with a string comparison."`
	DepartmentName          *workloadFilterComparison `json:"departmentName,omitempty" jsonschema:"Filter by department name with a string comparison."`
	SubmittedBy             *workloadFilterComparison `json:"submittedBy,omitempty" jsonschema:"Filter by submitter with a string comparison."`
	CompletedAt             *workloadFilterComparison `json:"completedAt,omitempty" jsonschema:"Filter by completion time with an RFC3339 timestamp comparison."`
	DistributedFramework    *workloadFilterComparison `json:"distributedFramework,omitempty" jsonschema:"Filter by distributed framework with a string comparison."`
	Category                *workloadFilterComparison `json:"category,omitempty" jsonschema:"Filter by workload category with a string comparison."`
	TotalPendingTimeSeconds *workloadFilterComparison `json:"totalPendingTimeSeconds,omitempty" jsonschema:"Filter by total pending seconds with an integer comparison."`
	TotalRunningTimeSeconds *workloadFilterComparison `json:"totalRunningTimeSeconds,omitempty" jsonschema:"Filter by total running seconds with an integer comparison."`
	Priority                *workloadFilterComparison `json:"priority,omitempty" jsonschema:"Filter by numeric workload priority with an integer comparison."`
	PriorityClassName       *workloadFilterComparison `json:"priorityClassName,omitempty" jsonschema:"Filter by priority class name with a string comparison."`
	GuaranteedRuntimeEndsAt *workloadFilterComparison `json:"guaranteedRuntimeEndsAt,omitempty" jsonschema:"Filter by guaranteed-runtime end with an RFC3339 timestamp comparison."`
	AIApplicationID         *workloadFilterComparison `json:"aiApplicationId,omitempty" jsonschema:"Filter by AI application ID with a string comparison."`
	AIApplicationName       *workloadFilterComparison `json:"aiApplicationName,omitempty" jsonschema:"Filter by AI application name with a string comparison."`
	GPUIdleOnly             bool                      `json:"gpuIdleOnly,omitempty"         jsonschema:"Return only running workloads whose idleAllocatedGpus value is non-zero."`
	AllocatedGPU            *gpuComparison            `json:"allocatedGpu,omitempty" jsonschema:"Filter by GPUs currently assigned to each workload. Use for a per-workload predicate such as 'workloads using more than 2 GPUs', unless the user explicitly says requested GPUs or GPU utilization. Applied server-side. Workloads with a null allocated GPU value are excluded from every comparison, including == 0 and != N."`                          //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	RequestedGPU            *gpuComparison            `json:"requestedGpu,omitempty" jsonschema:"Filter multiple workloads by GPUs requested in their specification. Requires projectId or clusterId; ask for one before calling if absent. This cannot be filtered server-side, so at most 200 scoped workloads are fetched and filtered locally. Workloads with a null requested GPU value are excluded from every comparison, including == 0 and != N."` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	Limit                   int32                     `json:"limit,omitempty"               jsonschema:"Maximum number of workloads to return. Defaults to 50, capped at 200."`
	Offset                  int32                     `json:"offset,omitempty"              jsonschema:"Pagination cursor from a previous response's nextOffset. Defaults to 0."`
}

func (in listWorkloadsInput) Validate() error {
	if err := common.ValidatePagination(in.Limit, in.Offset); err != nil {
		return err
	}
	if in.ClusterID != "" {
		if err := common.ValidateUUID("clusterId", in.ClusterID); err != nil {
			return err
		}
	}
	if in.Phase != "" && !wlclient.Phase(in.Phase).IsValid() {
		return fmt.Errorf("phase must be one of: %s", phaseValues())
	}
	if in.GPUIdleOnly && in.Phase != "" && in.Phase != string(wlclient.PHASE_RUNNING) {
		return fmt.Errorf("phase must be Running when gpuIdleOnly is true")
	}
	if in.GPUIdleOnly && deletedWorkloadsRequested(in) {
		return fmt.Errorf("gpuIdleOnly cannot be combined with deletedOnly or deletedAfter because GPU-idle workloads must be currently Running")
	}
	if in.AllocatedGPU != nil {
		if err := in.AllocatedGPU.Validate("allocatedGpu"); err != nil {
			return err
		}
	}
	if in.RequestedGPU != nil {
		if err := in.RequestedGPU.Validate("requestedGpu"); err != nil {
			return err
		}
		if in.ProjectID == "" && in.ClusterID == "" {
			return fmt.Errorf("requestedGpu requires projectId or clusterId; ask the user to narrow the scope first")
		}
	}
	if _, _, err := common.ParseRFC3339IfSet("createdAfter", in.CreatedAfter); err != nil {
		return err
	}
	if _, _, err := common.ParseRFC3339IfSet("phaseUpdatedAfter", in.PhaseUpdatedAfter); err != nil {
		return err
	}
	if _, _, err := common.ParseRFC3339IfSet("deletedAfter", in.DeletedAfter); err != nil {
		return err
	}
	if in.Search != "" && strings.TrimSpace(in.Search) == "" {
		return fmt.Errorf("search must not be blank")
	}
	if in.SortBy != "" {
		if _, ok := workloadSortFields[in.SortBy]; !ok {
			return fmt.Errorf("sortBy must be one of: %s", workloadSortFieldOptions)
		}
	}
	if in.SortOrder != "" && in.SortOrder != "asc" && in.SortOrder != "desc" {
		return fmt.Errorf("sortOrder must be one of: asc, desc")
	}
	for i, filter := range in.FilterBy {
		if err := filter.validate(i); err != nil {
			return err
		}
	}
	for _, comparison := range workloadFilterComparisons(in) {
		if comparison.value != nil {
			if err := comparison.value.validate(comparison.field, comparison.kind); err != nil {
				return err
			}
		}
	}
	for _, fv := range []struct{ field, value string }{
		{"name", in.Name},
		{"projectId", in.ProjectID},
		{"type", in.Type},
		{"nodePoolName", in.NodePoolName},
	} {
		if fv.value != "" && strings.TrimSpace(fv.value) == "" {
			return fmt.Errorf("%s must not be blank", fv.field)
		}
		if err := common.ValidateFilterValue(fv.field, fv.value); err != nil {
			return err
		}
	}
	if len(listWorkloadsFilters(in)) > common.MaxAPIFilters {
		return fmt.Errorf("combined filters exceed the Workloads API maximum of %d; narrow the query", common.MaxAPIFilters)
	}
	return nil
}

func phaseValues() string {
	values := make([]string, 0, len(wlclient.AllowedPhaseEnumValues))
	for _, phase := range wlclient.AllowedPhaseEnumValues {
		values = append(values, string(phase))
	}
	return strings.Join(values, ", ")
}

type workloadListItem struct {
	ID                 string   `json:"id"                 jsonschema:"Workload ID. Pass this as workloadId to workload-specific tools."`
	Name               string   `json:"name"               jsonschema:"Workload name."`
	Type               string   `json:"type"               jsonschema:"Workload type, such as Training, Workspace, Distributed, or Inference."`
	Phase              string   `json:"phase"              jsonschema:"Current workload phase."`
	ProjectID          string   `json:"projectId"          jsonschema:"Owning project ID."`
	ProjectName        string   `json:"projectName"        jsonschema:"Owning project name."`
	ClusterID          string   `json:"clusterId"          jsonschema:"Cluster UUID containing the workload."`
	CreatedAt          string   `json:"createdAt"          jsonschema:"Workload creation timestamp in RFC3339 UTC."`
	PhaseUpdatedAt     string   `json:"phaseUpdatedAt"     jsonschema:"Timestamp of the last phase change in RFC3339 UTC."`
	DeletedAt          string   `json:"deletedAt,omitempty" jsonschema:"Workload deletion timestamp in RFC3339 UTC. Present for deleted workloads."`
	RequestedNodePools []string `json:"requestedNodePools" jsonschema:"Node pools requested by the workload. Always present; empty when none were requested."`
	CurrentNodePools   []string `json:"currentNodePools"   jsonschema:"Node pools currently running the workload. Always present; empty when the workload has no placement."`
	Priority           *int32   `json:"priority"           jsonschema:"Numeric workload priority. Null when unavailable."`
	PriorityClassName  string   `json:"priorityClassName"  jsonschema:"Workload priority class name."`
	Preemptible        *bool    `json:"preemptible"        jsonschema:"Whether the workload can be preempted. Null when unavailable."`
	IdleAllocatedGpus  *float32 `json:"idleAllocatedGpus"  jsonschema:"Number of currently allocated GPUs that are idle. Null when unavailable."`
	AllocatedGPU       *float64 `json:"allocatedGpu"       jsonschema:"GPUs currently assigned to the workload. Null when the API has no allocation value."`
	RequestedGPU       *float64 `json:"requestedGpu"       jsonschema:"GPUs requested by the workload specification. Null when no GPU-device request value is available (including GPU-memory-only requests)."`
}

type listWorkloadsOutput struct {
	Workloads  []workloadListItem `json:"workloads"            jsonschema:"Workloads matching the requested filters. Always present; empty array when none match."`
	HasMore    bool               `json:"hasMore"              jsonschema:"True when more workloads exist beyond the returned page; pass nextOffset back as offset."`
	NextOffset int32              `json:"nextOffset,omitempty" jsonschema:"Offset to pass on the next call to fetch more workloads. Only set when hasMore is true."`
}

func registerListWorkloads(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[listWorkloadsInput, listWorkloadsOutput]{
		Name: "list_workloads",
		Description: "List workloads with their ID, name, type, phase, project, cluster, timestamps, node pools, priority, and preemptibility. " +
			"Optionally filter by name, project, cluster, phase, type, node pool, or recent activity. " +
			"Supports every List Workloads API query parameter and filter field. Use filterBy for typed field/operator/value comparisons not covered by the convenience inputs; all filters are ANDed and strict numeric/timestamp comparisons are preserved. " + //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
			"For workloads deleted in the last X hours, pass an RFC3339 cutoff (now minus X hours) as deletedAfter; it selects deleted workloads and returns deletedAt. Use deletedOnly only to list all deleted workloads. " +
			"Use gpuIdleOnly for which workloads are GPU-idle: this returns Running workloads whose idleAllocatedGpus value is non-zero, and includes that value in each result. " +
			"When resolving an ID for another tool, filter by workload name and projectId when known; " +
			"use the ID if one workload matches, or show the matching projects and phases and ask the user to select one if multiple match. " +
			"Do NOT call when the user asks only HOW MANY workloads match, including how many are GPU-idle — use get_workloads_summary. " +
			"Returns allocatedGpu and requestedGpu. For multiple workloads, ambiguous 'using N GPUs' means allocatedGpu; preserve >, >=, <, <=, ==, or != exactly and use the server-side filter. " +
			"RequestedGpu filtering requires projectId or clusterId, scans at most 200 scoped workloads, and refuses partial results; ask for scope first when missing. For one workload's allocated or requested GPU count, use get_workload_spec. " + //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
			"GPU utilization is metrics: one workload uses get_workload_metrics, cluster/node pool uses get_cluster_metrics, and project/department uses get_org_unit_metrics. To compare utilization across workloads, first offer aggregate metrics or narrower scope; only after the user insists, clarify the window and statistic, enumerate every candidate, and continue only for a complete set of 10 or fewer. " + //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
			"For counts matching allocatedGpu or requestedGpu, use this tool and follow all pages; get_workloads_summary cannot apply those filters. " +
			"Do NOT call for detailed status, events, specifications, pods, metrics, or history — use the relevant workload-specific tool.",
		ReadOnly: true,
		Handler: func(ctx context.Context, in listWorkloadsInput) (listWorkloadsOutput, error) {
			return listWorkloadsRun(ctx, client, in)
		},
	})
}

func listWorkloadsRun(ctx context.Context, client *runaiclient.Client, in listWorkloadsInput) (listWorkloadsOutput, error) {
	if in.RequestedGPU != nil {
		return listWorkloadsByRequestedGPU(ctx, client, in)
	}
	return listWorkloadsPage(ctx, client, in)
}

func listWorkloadsPage(ctx context.Context, client *runaiclient.Client, in listWorkloadsInput) (listWorkloadsOutput, error) {
	// Step 1 — build the filtered, paginated request.
	out := listWorkloadsOutput{Workloads: []workloadListItem{}}
	limit := common.EffectiveLimit(in.Limit)
	filters := listWorkloadsFilters(in)

	sortBy, sortOrder := workloadSort(in)
	req := client.Workloads.WorkloadsAPI.GetWorkloads(ctx).
		Limit(limit).
		Offset(in.Offset).
		SortBy(sortBy).
		SortOrder(sortOrder)
	if deletedWorkloadsRequested(in) {
		req = req.Deleted(true)
	}
	if len(filters) > 0 {
		req = req.FilterBy(filters)
	}
	if in.Search != "" {
		req = req.Search(in.Search)
	}

	// Step 2 — fetch the workload page.
	resp, httpResp, err := req.Execute()
	if err != nil {
		return out, fmt.Errorf("list workloads: %w", runaimcp.MapHTTPError(httpResp, err))
	}
	if resp == nil {
		return out, fmt.Errorf("nil response from upstream: %w", runaimcp.ErrUpstream)
	}

	// Step 3 — map the compact workload fields returned by this tool.
	out.Workloads = make([]workloadListItem, 0, len(resp.Workloads))
	for _, workload := range resp.Workloads {
		out.Workloads = append(out.Workloads, mapWorkloadListItem(workload))
	}

	// Step 4 — return the upstream next offset if it advances pagination.
	if resp.Next != nil && *resp.Next > in.Offset {
		out.HasMore = true
		out.NextOffset = *resp.Next
	}
	return out, nil
}

func listWorkloadsByRequestedGPU(ctx context.Context, client *runaiclient.Client, in listWorkloadsInput) (listWorkloadsOutput, error) {
	out := listWorkloadsOutput{Workloads: []workloadListItem{}}
	filters := listWorkloadsFilters(in)
	count, err := countWorkloads(ctx, client, filters, in.Search, deletedWorkloadsRequested(in))
	if err != nil {
		return out, fmt.Errorf("count workloads for requested-GPU filtering: %w", err)
	}
	if count > int64(common.UpstreamPageSize) {
		return out, requestedGPUScopeTooLargeError(count)
	}
	sortBy, sortOrder := workloadSort(in)
	req := client.Workloads.WorkloadsAPI.GetWorkloads(ctx).
		Limit(common.UpstreamPageSize).
		Offset(0).
		SortBy(sortBy).
		SortOrder(sortOrder)
	if deletedWorkloadsRequested(in) {
		req = req.Deleted(true)
	}
	if len(filters) > 0 {
		req = req.FilterBy(filters)
	}
	if in.Search != "" {
		req = req.Search(in.Search)
	}

	resp, httpResp, err := req.Execute()
	if err != nil {
		return out, fmt.Errorf("list workloads for requested-GPU filtering: %w", runaimcp.MapHTTPError(httpResp, err))
	}
	if resp == nil {
		return out, fmt.Errorf("nil response from upstream: %w", runaimcp.ErrUpstream)
	}
	if resp.Next != nil && *resp.Next > 0 && count < int64(common.UpstreamPageSize) {
		return out, fmt.Errorf("the requested-GPU scope grew while it was being read; ask the user to narrow the project or cluster and retry: %w", runaimcp.ErrValidation)
	}
	if resp.Next != nil && *resp.Next > 0 && count == int64(common.UpstreamPageSize) {
		latestCount, err := countWorkloads(ctx, client, filters, in.Search, deletedWorkloadsRequested(in))
		if err != nil {
			return out, fmt.Errorf("recount workloads for requested-GPU filtering: %w", err)
		}
		if latestCount > int64(common.UpstreamPageSize) {
			return out, requestedGPUScopeTooLargeError(latestCount)
		}
	}

	matched := make([]wlclient.Workload, 0, len(resp.Workloads))
	for _, workload := range resp.Workloads {
		if compareGPU(requestedGPU(workload), *in.RequestedGPU) {
			matched = append(matched, workload)
		}
	}
	// Requested-GPU filtering re-reads the complete scoped set on every MCP page.
	// As with server-side offset pagination, concurrent workload changes can move
	// rows between calls and may cause a later page to skip or repeat a workload.
	page, hasMore, next := common.Paginate(matched, in.Offset, common.EffectiveLimit(in.Limit))
	out.Workloads = make([]workloadListItem, 0, len(page))
	for _, workload := range page {
		out.Workloads = append(out.Workloads, mapWorkloadListItem(workload))
	}
	out.HasMore = hasMore
	out.NextOffset = next
	return out, nil
}

func requestedGPUScopeTooLargeError(count int64) error {
	return fmt.Errorf("%d workloads match the scope, exceeding the maximum of %d; ask the user to narrow the project or cluster before filtering requested GPUs: %w", count, common.UpstreamPageSize, runaimcp.ErrValidation)
}

type namedWorkloadFilterComparison struct {
	field string
	kind  string
	value *workloadFilterComparison
}

func workloadFilterComparisons(in listWorkloadsInput) []namedWorkloadFilterComparison {
	return []namedWorkloadFilterComparison{
		{field: "projectName", kind: "string", value: in.ProjectName},
		{field: "departmentId", kind: "string", value: in.DepartmentID},
		{field: "departmentName", kind: "string", value: in.DepartmentName},
		{field: "submittedBy", kind: "string", value: in.SubmittedBy},
		{field: "completedAt", kind: "timestamp", value: in.CompletedAt},
		{field: "distributedFramework", kind: "string", value: in.DistributedFramework},
		{field: "category", kind: "string", value: in.Category},
		{field: "totalPendingTimeSeconds", kind: "integer", value: in.TotalPendingTimeSeconds},
		{field: "totalRunningTimeSeconds", kind: "integer", value: in.TotalRunningTimeSeconds},
		{field: "priority", kind: "integer", value: in.Priority},
		{field: "priorityClassName", kind: "string", value: in.PriorityClassName},
		{field: "guaranteedRuntimeEndsAt", kind: "timestamp", value: in.GuaranteedRuntimeEndsAt},
		{field: "aiApplicationId", kind: "string", value: in.AIApplicationID},
		{field: "aiApplicationName", kind: "string", value: in.AIApplicationName},
	}
}

func workloadSort(in listWorkloadsInput) (string, string) {
	sortBy := in.SortBy
	if sortBy == "" {
		sortBy = "phaseUpdatedAt"
	}
	switch sortBy {
	case "nodePoolName":
		sortBy = "nodepool"
	case "allocatedGpu":
		sortBy = "allocatedGPU"
	}
	sortOrder := in.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	return sortBy, sortOrder
}

func listWorkloadsFilters(in listWorkloadsInput) []string {
	filters := make([]string, 0, 10)
	for _, fv := range []struct{ field, value string }{
		{"name", in.Name},
		{"projectId", in.ProjectID},
		{"clusterId", in.ClusterID},
		{"phase", in.Phase},
		{"type", in.Type},
		{"nodepool", in.NodePoolName},
	} {
		if fv.value != "" {
			filters = append(filters, common.FilterEq(fv.field, fv.value))
		}
	}
	if in.CreatedAfter != "" {
		filters = append(filters, "createdAt>="+in.CreatedAfter)
	}
	if in.PhaseUpdatedAfter != "" {
		filters = append(filters, "phaseUpdatedAt>="+in.PhaseUpdatedAfter)
	}
	if in.DeletedAfter != "" {
		filters = append(filters, "deletedAt>="+in.DeletedAfter)
	}
	for _, filter := range in.FilterBy {
		filters = append(filters, serverWorkloadComparisonFilters(filter.Field, workloadFilterComparison{
			Operator: filter.Operator,
			Value:    filter.Value,
		})...)
	}
	for _, comparison := range workloadFilterComparisons(in) {
		if comparison.value != nil {
			filters = append(filters, serverWorkloadComparisonFilters(comparison.field, *comparison.value)...)
		}
	}
	if in.GPUIdleOnly {
		if in.Phase == "" {
			filters = append(filters, common.FilterEq("phase", string(wlclient.PHASE_RUNNING)))
		}
		filters = append(filters, "idleAllocatedGpus!=0")
	}
	if in.AllocatedGPU != nil {
		filters = append(filters, serverGPUFilters("allocatedGPU", *in.AllocatedGPU)...)
	}
	return filters
}

func serverWorkloadComparisonFilters(field string, comparison workloadFilterComparison) []string {
	switch comparison.Operator {
	case ">":
		return []string{field + ">=" + comparison.Value, field + "!=" + comparison.Value}
	case "<":
		return []string{field + "<=" + comparison.Value, field + "!=" + comparison.Value}
	default:
		return []string{field + comparison.Operator + comparison.Value}
	}
}

func deletedWorkloadsRequested(in listWorkloadsInput) bool {
	if in.DeletedOnly || in.DeletedAfter != "" {
		return true
	}
	for _, filter := range in.FilterBy {
		if filter.Field == "deletedAt" {
			return true
		}
	}
	return false
}

func serverGPUFilters(field string, comparison gpuComparison) []string {
	value := strconv.FormatFloat(comparison.Value, 'f', -1, 64)
	switch comparison.Operator {
	case ">":
		return []string{field + ">=" + value, field + "!=" + value}
	case "<":
		return []string{field + "<=" + value, field + "!=" + value}
	case "!=":
		// The API's != filter includes null allocations. Constrain the value to
		// the numeric domain so null remains "unavailable" for every comparison.
		return []string{field + ">=0", field + "!=" + value}
	default:
		return []string{field + comparison.Operator + value}
	}
}

func compareGPU(value *float64, comparison gpuComparison) bool {
	if value == nil {
		return false
	}
	switch comparison.Operator {
	case ">":
		return *value > comparison.Value
	case ">=":
		return *value >= comparison.Value
	case "<":
		return *value < comparison.Value
	case "<=":
		return *value <= comparison.Value
	case "==":
		return *value == comparison.Value
	case "!=":
		return *value != comparison.Value
	default:
		return false
	}
}

func allocatedGPU(workload wlclient.Workload) *float64 {
	resources, ok := workload.GetAllocatedResourcesOk()
	if !ok || resources == nil {
		return nil
	}
	value, ok := resources.GetGpuOk()
	if !ok || value == nil {
		return nil
	}
	return value
}

func requestedGPU(workload wlclient.Workload) *float64 {
	for _, resources := range []*wlclient.WorkloadRequestResources{
		workload.WorkloadRequestedResources.Get(),
		workload.PodsRequestedResources.Get(),
	} {
		if resources == nil {
			continue
		}
		gpu, ok := resources.GetGpuOk()
		if !ok || gpu == nil {
			continue
		}
		value, ok := gpu.GetRequestOk()
		if ok && value != nil {
			return value
		}
	}
	return nil
}

func mapWorkloadListItem(workload wlclient.Workload) workloadListItem {
	return workloadListItem{
		ID:                 workload.Id,
		Name:               workload.Name,
		Type:               workload.Type,
		Phase:              string(workload.Phase),
		ProjectID:          workload.ProjectId,
		ProjectName:        workload.ProjectName,
		ClusterID:          workload.ClusterId,
		CreatedAt:          common.TimestampString(common.TimePtrIfNonZero(workload.CreatedAt)),
		PhaseUpdatedAt:     common.TimestampString(common.TimePtrIfNonZero(workload.PhaseUpdatedAt)),
		DeletedAt:          common.TimestampString(workload.DeletedAt.Get()),
		RequestedNodePools: append([]string{}, workload.RequestedNodePools...),
		CurrentNodePools:   append([]string{}, workload.CurrentNodePools...),
		Priority:           workload.Priority.Get(),
		PriorityClassName:  workload.PriorityClassName,
		Preemptible:        workload.Preemptible.Get(),
		IdleAllocatedGpus:  workload.IdleAllocatedGpus.Get(),
		AllocatedGPU:       allocatedGPU(workload),
		RequestedGPU:       requestedGPU(workload),
	}
}
