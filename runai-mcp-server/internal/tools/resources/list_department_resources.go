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

	orgunit "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

type listDepartmentResourcesInput struct {
	ClusterID      string `json:"clusterId,omitempty"      jsonschema:"Cluster ID to scope the lookup to. If omitted, results span every cluster the caller can access."`
	DepartmentID   string `json:"departmentId,omitempty"   jsonschema:"Department ID to filter cells to. Mutually exclusive with departmentName."`
	DepartmentName string `json:"departmentName,omitempty" jsonschema:"Department name to filter cells to. Mutually exclusive with departmentId; returns ambiguous_department if the name matches multiple departments."`
	NodePoolID     string `json:"nodePoolId,omitempty"     jsonschema:"Node pool ID to filter cells to. Mutually exclusive with nodePoolName."`
	NodePoolName   string `json:"nodePoolName,omitempty"   jsonschema:"Node pool name to filter cells to. Mutually exclusive with nodePoolId; returns ambiguous_node_pool if the name matches multiple pools."`
	Limit          int32  `json:"limit,omitempty"          jsonschema:"Maximum number of cells to return. Defaults to 50, capped at 200."`
	Offset         int32  `json:"offset,omitempty"         jsonschema:"Pagination cursor from a previous response's nextOffset. Defaults to 0."`
}

func (in listDepartmentResourcesInput) Validate() error {
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

// departmentQuota is the configured quota for one resource. Unlike the project
// resourceQuota it has no allocated field: the node-pool-resources endpoint this
// tool uses carries no quotaStatus, so point-in-time allocation comes from
// list_departments (allocated) and allocation over time from get_org_unit_metrics.
type departmentQuota struct {
	Deserved        *float64 `json:"deserved"        jsonschema:"Configured quota assigned to this department on this node pool. Null when the resource is not metered."`
	Limit           *float64 `json:"limit"           jsonschema:"Maximum the department can consume on this node pool, including over-quota. Null when unlimited or not set."`
	OverQuotaWeight *float64 `json:"overQuotaWeight" jsonschema:"Fair-share weight (1-20) when org-units compete for over-quota resources within the same rank. Null when not set."`
}

type departmentResourceCell struct {
	DepartmentID   string          `json:"departmentId"   jsonschema:"Run:ai department ID (the department's own org-unit) for this cell."`
	DepartmentName string          `json:"departmentName" jsonschema:"Run:ai department display name."`
	NodePoolID     string          `json:"nodePoolId"     jsonschema:"Node pool ID this assignment applies to."`
	NodePoolName   string          `json:"nodePoolName"   jsonschema:"Node pool name this assignment applies to."`
	Rank           string          `json:"rank"           jsonschema:"Scheduling rank for this department on this node pool. Common values: High, Medium, MediumLow (default), Low. Replaces the deprecated priority field."`
	GPU            departmentQuota `json:"gpu"            jsonschema:"Configured GPU quota in GPU device count."`
	CPUCores       departmentQuota `json:"cpuCores"       jsonschema:"Configured CPU quota in cores (normalized from millicores upstream)."`
	MemoryGiB      departmentQuota `json:"memoryGib"      jsonschema:"Configured memory quota in GiB (normalized from mixed units upstream)."`
}

type listDepartmentResourcesOutput struct {
	Resources  []departmentResourceCell `json:"resources"             jsonschema:"Flat list of (department × node-pool) quota cells matching the requested filters. Always present; empty array when no cells match."`
	HasMore    bool                     `json:"hasMore"               jsonschema:"True when more cells exist beyond the returned page; pass nextOffset back as offset to fetch them."`
	NextOffset int32                    `json:"nextOffset,omitempty"  jsonschema:"Offset to pass on the next call to fetch additional cells. Only set when hasMore is true."`
}

func registerListDepartmentResources(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[listDepartmentResourcesInput, listDepartmentResourcesOutput]{
		Name: "list_department_resources",
		Description: "List a department's OWN configured GPU/CPU/memory quota per node pool. " +
			"Use when the user asks about a department's own resource quota on a node pool — " +
			"\"what is department Engineering's GPU quota?\", \"which departments have quota on H100?\". " +
			"Returns (department × node-pool) cells with the department's configured quota (no live allocation); " +
			"call without filters for the tenant-wide view. A department with no configured quota returns zero " +
			"cells, so do NOT use this to enumerate or count departments — use list_departments for that. For a " +
			"department's CHILD-PROJECT quotas, use list_project_resources with a department filter instead. For a " +
			"department's live allocation or utilization, use get_org_unit_metrics.",
		ReadOnly: true,
		Handler: func(ctx context.Context, in listDepartmentResourcesInput) (listDepartmentResourcesOutput, error) {
			return listDepartmentResourcesRun(ctx, client, in)
		},
	})
}

func listDepartmentResourcesRun(ctx context.Context, client *runaiclient.Client, in listDepartmentResourcesInput) (listDepartmentResourcesOutput, error) {
	out := listDepartmentResourcesOutput{Resources: []departmentResourceCell{}}

	filters := buildDepartmentFilters(in)
	limit := common.EffectiveLimit(in.Limit)

	rawCells, err := fetchAllDepartmentNodePoolResources(ctx, client, filters)
	if err != nil {
		return out, err
	}

	if err := detectDepartmentAmbiguity(in, rawCells); err != nil {
		return out, err
	}

	filtered := make([]departmentResourceCell, 0, len(rawCells))
	for _, c := range rawCells {
		cell := mapDepartmentCell(c)
		if !departmentCellHasAssignedQuota(cell) {
			continue
		}
		filtered = append(filtered, cell)
	}
	sortDepartmentCells(filtered)

	out.Resources, out.HasMore, out.NextOffset = common.Paginate(filtered, in.Offset, limit)
	return out, nil
}

// Bounded by common.MaxPaginatedResults: the result is one row per
// (department × node-pool) for the filter, which is small in practice, so
// fetching it all to paginate locally is fine.
func fetchAllDepartmentNodePoolResources(ctx context.Context, client *runaiclient.Client, filters []string) ([]orgunit.DepartmentNodePoolResources, error) {
	var all []orgunit.DepartmentNodePoolResources
	offset := int32(0)
	for {
		resp, httpResp, err := client.OrgUnit.DepartmentsAPI.GetDepartmentNodePoolResources(ctx).
			FilterBy(filters).
			Limit(common.UpstreamPageSize).
			Offset(offset).
			Execute()
		if err != nil {
			return nil, fmt.Errorf("list department node-pool resources: %w", runaimcp.MapHTTPError(httpResp, err))
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

func buildDepartmentFilters(in listDepartmentResourcesInput) []string {
	filters := make([]string, 0, 4)
	if in.ClusterID != "" {
		filters = append(filters, common.FilterEq("clusterId", in.ClusterID))
	}
	if in.DepartmentID != "" {
		filters = append(filters, common.FilterEq("orgUnitId", in.DepartmentID))
	}
	if in.DepartmentName != "" {
		filters = append(filters, common.FilterEq("orgUnitName", in.DepartmentName))
	}
	if in.NodePoolID != "" {
		filters = append(filters, common.FilterEq("nodePoolId", in.NodePoolID))
	}
	if in.NodePoolName != "" {
		filters = append(filters, common.FilterEq("nodePoolName", in.NodePoolName))
	}
	return filters
}

func mapDepartmentCell(c orgunit.DepartmentNodePoolResources) departmentResourceCell {
	return departmentResourceCell{
		DepartmentID:   c.OrgUnitId,
		DepartmentName: c.OrgUnitName,
		NodePoolID:     c.NodePoolId,
		NodePoolName:   c.NodePoolName,
		Rank:           c.Rank,
		GPU: configuredQuota(resourceQuota{
			Deserved:        gpuDeserved(c.Resources),
			Limit:           gpuLimit(c.Resources),
			OverQuotaWeight: gpuOverQuotaWeight(c.Resources),
		}),
		CPUCores: configuredQuota(resourceQuota{
			Deserved:        cpuCoresDeserved(c.Resources),
			Limit:           cpuCoresLimit(c.Resources),
			OverQuotaWeight: cpuOverQuotaWeight(c.Resources),
		}),
		MemoryGiB: configuredQuota(resourceQuota{
			Deserved:        memoryGiBDeserved(c.Resources),
			Limit:           memoryGiBLimit(c.Resources),
			OverQuotaWeight: memoryOverQuotaWeight(c.Resources),
		}),
	}
}

// Departments expose no live allocation; drop the allocated field.
func configuredQuota(q resourceQuota) departmentQuota {
	n := normalizeQuota(q)
	return departmentQuota{Deserved: n.Deserved, Limit: n.Limit, OverQuotaWeight: n.OverQuotaWeight}
}

// departmentCellHasAssignedQuota keeps a cell when any resource has configured
// quota — not just GPU: a CPU/memory-only quota is a real assignment. Department
// cells carry no live allocation upstream, so there is no allocation to consider.
func departmentCellHasAssignedQuota(cell departmentResourceCell) bool {
	for _, q := range []departmentQuota{cell.GPU, cell.CPUCores, cell.MemoryGiB} {
		if q.Deserved != nil && *q.Deserved > 0 {
			return true
		}
		if q.Limit != nil && *q.Limit > 0 {
			return true
		}
	}
	return false
}

func sortDepartmentCells(cells []departmentResourceCell) {
	sort.SliceStable(cells, func(i, j int) bool {
		if cells[i].DepartmentName != cells[j].DepartmentName {
			return cells[i].DepartmentName < cells[j].DepartmentName
		}
		return cells[i].NodePoolName < cells[j].NodePoolName
	})
}

func detectDepartmentAmbiguity(in listDepartmentResourcesInput, cells []orgunit.DepartmentNodePoolResources) error {
	if in.DepartmentName != "" {
		ids := distinct(cells, func(c orgunit.DepartmentNodePoolResources) string { return c.OrgUnitId })
		if len(ids) > 1 {
			return fmt.Errorf("ambiguous_department: department name %q matches multiple departments (%s): %w",
				in.DepartmentName, strings.Join(ids, ", "), runaimcp.ErrValidation)
		}
	}
	if in.NodePoolName != "" {
		ids := distinct(cells, func(c orgunit.DepartmentNodePoolResources) string { return c.NodePoolId })
		if len(ids) > 1 {
			return fmt.Errorf("ambiguous_node_pool: node pool name %q matches multiple pools (%s): %w",
				in.NodePoolName, strings.Join(ids, ", "), runaimcp.ErrValidation)
		}
	}
	return nil
}
