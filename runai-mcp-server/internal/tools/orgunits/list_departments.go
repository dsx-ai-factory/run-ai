/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orgunits

import (
	"context"
	"fmt"
	"sort"

	orgunit "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

type listDepartmentsInput struct {
	ClusterID string `json:"clusterId,omitempty" jsonschema:"Cluster ID to scope the listing to. If omitted, results span every cluster the caller can access."`
	Limit     int32  `json:"limit,omitempty"     jsonschema:"Maximum number of departments to return. Defaults to 50, capped at 200."`
	Offset    int32  `json:"offset,omitempty"    jsonschema:"Pagination cursor from a previous response's nextOffset. Defaults to 0."`
}

func (in listDepartmentsInput) Validate() error {
	if err := common.ValidatePagination(in.Limit, in.Offset); err != nil {
		return err
	}
	return common.ValidateNoCommas("clusterId", in.ClusterID)
}

type departmentAllocation struct {
	GPU       *float64 `json:"gpu,omitempty"       jsonschema:"GPUs allocated to the department now, in device count — the UI's 'Allocated GPUs'."`
	CPUCores  *float64 `json:"cpuCores,omitempty"  jsonschema:"CPU cores allocated to the department now (normalized from millicores)."`
	MemoryGiB *float64 `json:"memoryGib,omitempty" jsonschema:"System memory allocated to the department now, in GiB (normalized from MB)."`
}

type departmentEntity struct {
	ID           string                `json:"id"                  jsonschema:"Run:ai department ID."`
	Name         string                `json:"name"                jsonschema:"Run:ai department display name."`
	ClusterID    string                `json:"clusterId"           jsonschema:"Cluster the department belongs to."`
	ProjectCount *int32                `json:"projectCount"        jsonschema:"Number of child projects under this department. Null when not derivable; use list_projects with departmentId to count directly."`
	Allocated    *departmentAllocation `json:"allocated,omitempty" jsonschema:"Resources allocated to the department now (point-in-time); compare allocated.gpu to gpuQuota for over-quota. Null when no status is reported."`
	GPUQuota     *float64              `json:"gpuQuota,omitempty"  jsonschema:"Total GPU quota (deserved) for the department across node pools. -1 means unlimited. Null when not reported."`
}

type listDepartmentsOutput struct {
	Departments []departmentEntity `json:"departments"           jsonschema:"Department entities matching the filters, including departments with no resource quota configured. Always present; empty array when none match."`
	TotalCount  int32              `json:"totalCount"            jsonschema:"Total number of departments matching the filters across all pages — authoritative for counting questions."`
	HasMore     bool               `json:"hasMore"               jsonschema:"True when more departments exist beyond the returned page; pass nextOffset back as offset to fetch them."`
	NextOffset  int32              `json:"nextOffset,omitempty"  jsonschema:"Offset to pass on the next call to fetch additional departments. Only set when hasMore is true."`
}

func registerListDepartments(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[listDepartmentsInput, listDepartmentsOutput]{
		Name: "list_departments",
		Description: "Enumerate Run:ai department entities — \"which departments exist?\", \"how many departments?\", " +
			"\"which departments are allocated GPUs beyond their quota?\". Lists every department the caller can see, " +
			"including departments with no resource quota configured (which list_department_resources cannot see), and " +
			"returns an authoritative totalCount. Each entity carries allocated (current point-in-time gpu/cpu/memory) and " +
			"gpuQuota, so over-quota departments are found by comparing allocated.gpu to gpuQuota (gpuQuota -1 is unlimited — never over-quota). Optionally filter by cluster. For a " +
			"department's per-node-pool quota use list_department_resources; for allocation over time use get_org_unit_metrics.",
		ReadOnly: true,
		Handler: func(ctx context.Context, in listDepartmentsInput) (listDepartmentsOutput, error) {
			return listDepartmentsRun(ctx, client, in)
		},
	})
}

func listDepartmentsRun(ctx context.Context, client *runaiclient.Client, in listDepartmentsInput) (listDepartmentsOutput, error) {
	out := listDepartmentsOutput{Departments: []departmentEntity{}}

	var filters []string
	if in.ClusterID != "" {
		filters = append(filters, common.FilterEq("clusterId", in.ClusterID))
	}

	raw, err := fetchAllDepartments(ctx, client, filters)
	if err != nil {
		return out, err
	}

	// The list endpoint does not reliably populate children, so child counts are
	// derived from the projects listing — best-effort: the department listing must
	// not fail because the (unrelated) projects fetch does, e.g. on an oversized
	// tenant. When counts can't be derived they are null, not a misleading zero.
	var childCount map[string]int32
	if projects, err := fetchAllProjects(ctx, client, filters); err == nil {
		childCount = make(map[string]int32, len(raw))
		for _, p := range projects {
			childCount[p.ParentId]++
		}
	}

	entities := make([]departmentEntity, 0, len(raw))
	for _, d := range raw {
		var count *int32
		if childCount != nil {
			v := childCount[d.Id]
			count = &v
		}
		entities = append(entities, departmentEntity{
			ID:           d.Id,
			Name:         d.Name,
			ClusterID:    d.ClusterId,
			ProjectCount: count,
			Allocated:    departmentAllocatedResources(d),
			GPUQuota:     d.TotalResources.GpuQuota.Get(),
		})
	}
	sort.SliceStable(entities, func(i, j int) bool { return entities[i].Name < entities[j].Name })

	out.TotalCount = int32(len(entities))
	out.Departments, out.HasMore, out.NextOffset = common.Paginate(entities, in.Offset, common.EffectiveLimit(in.Limit))
	return out, nil
}

// departmentAllocatedResources normalizes the department's point-in-time allocation with the
// same common helpers list_project_resources uses. Null when the platform reports no allocation.
func departmentAllocatedResources(d orgunit.DepartmentForList) *departmentAllocation {
	status := d.Status.Get()
	if status == nil {
		return nil
	}
	gpu := common.AllocatedGPU(status.QuotaStatus)
	cpu := common.AllocatedCPUCores(status.QuotaStatus)
	mem := common.AllocatedMemoryGiB(status.QuotaStatus)
	if gpu == nil && cpu == nil && mem == nil {
		return nil
	}
	return &departmentAllocation{GPU: gpu, CPUCores: cpu, MemoryGiB: mem}
}

func fetchAllDepartments(ctx context.Context, client *runaiclient.Client, filters []string) ([]orgunit.DepartmentForList, error) {
	var all []orgunit.DepartmentForList
	offset := int32(0)
	for {
		// Verbose so the response carries status.quotaStatus (the point-in-time allocation);
		// the lean default omits it.
		req := client.OrgUnit.DepartmentsAPI.GetDepartments(ctx).
			Verbosity(orgunit.VERBOSITY_VERBOSE).Limit(common.UpstreamPageSize).Offset(offset)
		if len(filters) > 0 {
			req = req.FilterBy(filters)
		}
		resp, httpResp, err := req.Execute()
		if err != nil {
			return nil, fmt.Errorf("list departments: %w", runaimcp.MapHTTPError(httpResp, err))
		}
		if resp == nil {
			return nil, fmt.Errorf("nil response from upstream: %w", runaimcp.ErrUpstream)
		}
		all = append(all, resp.Departments...)
		if resp.Next == nil || *resp.Next <= offset {
			break
		}
		if int32(len(all)) >= common.MaxPaginatedResults {
			return nil, fmt.Errorf("result set exceeds the maximum supported size (%d departments); narrow the filter: %w", common.MaxPaginatedResults, runaimcp.ErrUpstream)
		}
		offset = *resp.Next
	}
	return all, nil
}
