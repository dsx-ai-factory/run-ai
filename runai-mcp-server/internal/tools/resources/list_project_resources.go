/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"
	"sort"
	"strings"

	orgunit "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

type listProjectResourcesInput struct {
	ClusterID      string `json:"clusterId,omitempty"      jsonschema:"Cluster ID to scope the lookup to. If omitted, results span every cluster the caller can access."`
	ProjectID      string `json:"projectId,omitempty"      jsonschema:"Project ID to filter cells to. Mutually exclusive with projectName."`
	ProjectName    string `json:"projectName,omitempty"    jsonschema:"Project name to filter cells to. Mutually exclusive with projectId; returns ambiguous_project if the name matches multiple projects."`
	DepartmentID   string `json:"departmentId,omitempty"   jsonschema:"Department ID to filter cells to as a parent (parentId). Mutually exclusive with departmentName. Returns all child projects' cells."`
	DepartmentName string `json:"departmentName,omitempty" jsonschema:"Department name to filter by (the projects' parent). Mutually exclusive with departmentId. Returns ambiguous_department on a name collision."`
	NodePoolID     string `json:"nodePoolId,omitempty"     jsonschema:"Node pool ID to filter cells to. Mutually exclusive with nodePoolName."`
	NodePoolName   string `json:"nodePoolName,omitempty"   jsonschema:"Node pool name to filter cells to. Mutually exclusive with nodePoolId; returns ambiguous_node_pool if the name matches multiple pools."`
	Limit          int32  `json:"limit,omitempty"          jsonschema:"Maximum number of cells to return. Defaults to 50, capped at 200."`
	Offset         int32  `json:"offset,omitempty"         jsonschema:"Pagination cursor from a previous response's nextOffset. Defaults to 0."`
}

func (in listProjectResourcesInput) Validate() error {
	if in.ProjectID != "" && in.ProjectName != "" {
		return fmt.Errorf("projectId and projectName are mutually exclusive")
	}
	if in.DepartmentID != "" && in.DepartmentName != "" {
		return fmt.Errorf("departmentId and departmentName are mutually exclusive")
	}
	if in.NodePoolID != "" && in.NodePoolName != "" {
		return fmt.Errorf("nodePoolId and nodePoolName are mutually exclusive")
	}
	if err := common.ValidatePagination(in.Limit, in.Offset); err != nil {
		return err
	}
	for _, fv := range []struct{ field, value string }{
		{"clusterId", in.ClusterID},
		{"projectId", in.ProjectID},
		{"projectName", in.ProjectName},
		{"departmentId", in.DepartmentID},
		{"departmentName", in.DepartmentName},
		{"nodePoolId", in.NodePoolID},
		{"nodePoolName", in.NodePoolName},
	} {
		if err := common.ValidateNoCommas(fv.field, fv.value); err != nil {
			return err
		}
	}
	return nil
}

type resourceCell struct {
	ProjectID    string        `json:"projectId"        jsonschema:"Run:ai project ID for this cell."`
	ProjectName  string        `json:"projectName"      jsonschema:"Run:ai project display name."`
	DepartmentID *string       `json:"departmentId"     jsonschema:"Parent department ID for this project. Null when the project has no department parent."`
	NodePoolID   string        `json:"nodePoolId"       jsonschema:"Node pool ID this assignment applies to."`
	NodePoolName string        `json:"nodePoolName"     jsonschema:"Node pool name this assignment applies to."`
	Rank         string        `json:"rank"             jsonschema:"Scheduling rank for this project on this node pool. Common values: High, Medium, MediumLow (default), Low. Replaces the deprecated priority field."`
	GPU          resourceQuota `json:"gpu"              jsonschema:"GPU quota and current allocation in GPU device count."`
	CPUCores     resourceQuota `json:"cpuCores"         jsonschema:"CPU quota and current allocation in cores. The API ships millicores upstream; values are normalized to cores at the tool boundary."`
	MemoryGiB    resourceQuota `json:"memoryGib"        jsonschema:"Memory quota and current allocation in GiB. The API ships values in mixed units upstream; values are normalized to GiB at the tool boundary."`
}

type listProjectResourcesOutput struct {
	Resources  []resourceCell `json:"resources"             jsonschema:"Flat list of (project × node-pool) quota cells matching the requested filters. Always present; empty array when no cells match."`
	HasMore    bool           `json:"hasMore"               jsonschema:"True when more cells exist beyond the returned page; pass nextOffset back as offset to fetch them."`
	NextOffset int32          `json:"nextOffset,omitempty"  jsonschema:"Offset to pass on the next call to fetch additional cells. Only set when hasMore is true."`
}

func registerListProjectResources(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[listProjectResourcesInput, listProjectResourcesOutput]{
		Name: "list_project_resources",
		Description: "List configured GPU/CPU/memory quotas and current allocation per project per node pool. " +
			"Use when the user asks about a project's resource quota or current allocation on a node pool — " +
			"\"what's project X's quota on H100?\", \"which projects in department Y are over-allocated?\". " +
			"Returns (project × node-pool) cells with configured quota and live allocation for GPU, CPU, and memory. " +
			"Call without filters for the tenant-wide view (\"who's using my GPUs?\"); a project appears only where " +
			"it has quota or allocation, so do NOT use this to enumerate or count projects — use list_projects for " +
			"that. Do NOT call for running workloads (use a workload tool) or time-series usage (use " +
			"get_org_unit_metrics or get_cluster_metrics).",
		ReadOnly: true,
		Handler: func(ctx context.Context, in listProjectResourcesInput) (listProjectResourcesOutput, error) {
			return listProjectResourcesRun(ctx, client, in)
		},
	})
}

func listProjectResourcesRun(ctx context.Context, client *runaiclient.Client, in listProjectResourcesInput) (listProjectResourcesOutput, error) {
	out := listProjectResourcesOutput{Resources: []resourceCell{}}

	resolvedDepartmentID := in.DepartmentID
	if in.DepartmentName != "" {
		id, err := common.ResolveOrgUnitID(ctx, client, common.OrgTypeDepartment, in.DepartmentName, in.ClusterID)
		if err != nil {
			return out, err
		}
		resolvedDepartmentID = id
	}

	filters := buildFilters(in, resolvedDepartmentID)
	limit := common.EffectiveLimit(in.Limit)

	// Paginate locally over the scope-filtered set: upstream pages include the
	// unassigned cells we drop, so upstream paging would yield short pages with
	// hasMore=true.
	rawCells, err := fetchAllProjectNodePoolResources(ctx, client, filters)
	if err != nil {
		return out, err
	}

	if err := detectAmbiguity(in, rawCells); err != nil {
		return out, err
	}

	filtered := make([]resourceCell, 0, len(rawCells))
	for _, c := range rawCells {
		cell := mapCell(c)
		if !cellHasQuotaOrAllocation(cell) {
			continue
		}
		filtered = append(filtered, cell)
	}
	sortCells(filtered)

	out.Resources, out.HasMore, out.NextOffset = common.Paginate(filtered, in.Offset, limit)
	return out, nil
}

func fetchAllProjectNodePoolResources(ctx context.Context, client *runaiclient.Client, filters []string) ([]orgunit.ProjectNodePoolResources, error) {
	var all []orgunit.ProjectNodePoolResources
	offset := int32(0)
	for {
		resp, httpResp, err := client.OrgUnit.ProjectsAPI.GetProjectNodePoolResources(ctx).
			FilterBy(filters).
			Limit(common.UpstreamPageSize).
			Offset(offset).
			Execute()
		if err != nil {
			return nil, fmt.Errorf("list project node-pool resources: %w", runaimcp.MapHTTPError(httpResp, err))
		}
		if resp == nil {
			return nil, fmt.Errorf("nil response from upstream: %w", runaimcp.ErrUpstream)
		}
		all = append(all, resp.NodePoolResources...)
		if resp.Next == nil || *resp.Next <= offset {
			break
		}
		if int32(len(all)) >= common.MaxPaginatedResults {
			return nil, fmt.Errorf("result set exceeds the maximum supported size (%d cells); narrow the filter: %w", common.MaxPaginatedResults, runaimcp.ErrUpstream)
		}
		offset = *resp.Next
	}
	return all, nil
}

// cellHasQuotaOrAllocation keeps a cell when any resource has configured quota OR live
// allocation. Upstream emits a cell for every (project × node-pool) combination,
// so fully-empty combinations are dropped as noise — but a cell must NOT be
// dropped just because GPU quota is unset: CPU/memory-only quotas are real, and
// a project can hold live allocation with no configured quota at all (dropping
// those made "who's using my GPUs?" return zero rows).
func cellHasQuotaOrAllocation(cell resourceCell) bool {
	for _, q := range []resourceQuota{cell.GPU, cell.CPUCores, cell.MemoryGiB} {
		if q.Deserved != nil && *q.Deserved > 0 {
			return true
		}
		if q.Limit != nil && *q.Limit > 0 {
			return true
		}
		if q.Allocated != nil && *q.Allocated > 0 {
			return true
		}
	}
	return false
}

func sortCells(cells []resourceCell) {
	sort.SliceStable(cells, func(i, j int) bool {
		if cells[i].ProjectName != cells[j].ProjectName {
			return cells[i].ProjectName < cells[j].ProjectName
		}
		return cells[i].NodePoolName < cells[j].NodePoolName
	})
}

func mapCell(c orgunit.ProjectNodePoolResources) resourceCell {
	var deptID *string
	if c.ParentId != "" {
		v := c.ParentId
		deptID = &v
	}
	return resourceCell{
		ProjectID:    c.OrgUnitId,
		ProjectName:  c.OrgUnitName,
		DepartmentID: deptID,
		NodePoolID:   c.NodePoolId,
		NodePoolName: c.NodePoolName,
		Rank:         c.Rank,
		GPU: normalizeQuota(resourceQuota{
			Deserved:        gpuDeserved(c.Resources),
			Limit:           gpuLimit(c.Resources),
			OverQuotaWeight: gpuOverQuotaWeight(c.Resources),
			Allocated:       common.AllocatedGPU(c.QuotaStatus),
		}),
		CPUCores: normalizeQuota(resourceQuota{
			Deserved:        cpuCoresDeserved(c.Resources),
			Limit:           cpuCoresLimit(c.Resources),
			OverQuotaWeight: cpuOverQuotaWeight(c.Resources),
			Allocated:       common.AllocatedCPUCores(c.QuotaStatus),
		}),
		MemoryGiB: normalizeQuota(resourceQuota{
			Deserved:        memoryGiBDeserved(c.Resources),
			Limit:           memoryGiBLimit(c.Resources),
			OverQuotaWeight: memoryOverQuotaWeight(c.Resources),
			Allocated:       common.AllocatedMemoryGiB(c.QuotaStatus),
		}),
	}
}

func buildFilters(in listProjectResourcesInput, resolvedDepartmentID string) []string {
	filters := make([]string, 0, 5)
	if in.ClusterID != "" {
		filters = append(filters, common.FilterEq("clusterId", in.ClusterID))
	}
	if in.ProjectID != "" {
		filters = append(filters, common.FilterEq("orgUnitId", in.ProjectID))
	}
	if in.ProjectName != "" {
		filters = append(filters, common.FilterEq("orgUnitName", in.ProjectName))
	}
	if resolvedDepartmentID != "" {
		filters = append(filters, common.FilterEq("parentId", resolvedDepartmentID))
	}
	if in.NodePoolID != "" {
		filters = append(filters, common.FilterEq("nodePoolId", in.NodePoolID))
	}
	if in.NodePoolName != "" {
		filters = append(filters, common.FilterEq("nodePoolName", in.NodePoolName))
	}
	return filters
}

func detectAmbiguity(in listProjectResourcesInput, cells []orgunit.ProjectNodePoolResources) error {
	if in.ProjectName != "" {
		ids := distinct(cells, func(c orgunit.ProjectNodePoolResources) string { return c.OrgUnitId })
		if len(ids) > 1 {
			return fmt.Errorf("ambiguous_project: project name %q matches multiple projects (%s): %w",
				in.ProjectName, strings.Join(ids, ", "), runaimcp.ErrValidation)
		}
	}
	if in.NodePoolName != "" {
		ids := distinct(cells, func(c orgunit.ProjectNodePoolResources) string { return c.NodePoolId })
		if len(ids) > 1 {
			return fmt.Errorf("ambiguous_node_pool: node pool name %q matches multiple pools (%s): %w",
				in.NodePoolName, strings.Join(ids, ", "), runaimcp.ErrValidation)
		}
	}
	return nil
}

func distinct[T any](items []T, key func(T) string) []string {
	seen := make(map[string]struct{}, len(items))
	keys := make([]string, 0, len(items))
	for _, item := range items {
		k := key(item)
		if k == "" {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		keys = append(keys, k)
	}
	return keys
}
