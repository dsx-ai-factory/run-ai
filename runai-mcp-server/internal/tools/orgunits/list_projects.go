/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orgunits

import (
	"context"
	"fmt"
	"sort"

	orgunit "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

type listProjectsInput struct {
	ClusterID      string `json:"clusterId,omitempty"      jsonschema:"Cluster ID to scope the listing to. If omitted, results span every cluster the caller can access."`
	DepartmentID   string `json:"departmentId,omitempty"   jsonschema:"Department ID to filter to that department's child projects. Mutually exclusive with departmentName."`
	DepartmentName string `json:"departmentName,omitempty" jsonschema:"Department name to filter to that department's child projects. Mutually exclusive with departmentId."`
	Limit          int32  `json:"limit,omitempty"          jsonschema:"Maximum number of projects to return. Defaults to 50, capped at 200."`
	Offset         int32  `json:"offset,omitempty"         jsonschema:"Pagination cursor from a previous response's nextOffset. Defaults to 0."`
}

func (in listProjectsInput) Validate() error {
	if in.DepartmentID != "" && in.DepartmentName != "" {
		return fmt.Errorf("departmentId and departmentName are mutually exclusive")
	}
	if err := common.ValidatePagination(in.Limit, in.Offset); err != nil {
		return err
	}
	for _, fv := range []struct{ field, value string }{
		{"clusterId", in.ClusterID},
		{"departmentId", in.DepartmentID},
		{"departmentName", in.DepartmentName},
	} {
		if err := common.ValidateNoCommas(fv.field, fv.value); err != nil {
			return err
		}
	}
	return nil
}

type projectEntity struct {
	ID           string  `json:"id"           jsonschema:"Run:ai project ID."`
	Name         string  `json:"name"         jsonschema:"Run:ai project display name."`
	ClusterID    string  `json:"clusterId"    jsonschema:"Cluster the project belongs to."`
	DepartmentID *string `json:"departmentId" jsonschema:"Parent department ID. Null when the project has no department parent."`
	Phase        *string `json:"phase"        jsonschema:"Project lifecycle phase (e.g. Ready, Creating, Deleting, NotReady). Null when not reported upstream."`
}

type listProjectsOutput struct {
	Projects   []projectEntity `json:"projects"              jsonschema:"Project entities matching the filters, including projects with no resource quota configured. Always present; empty array when none match."`
	TotalCount int32           `json:"totalCount"            jsonschema:"Total number of projects matching the filters across all pages — authoritative for counting questions."`
	HasMore    bool            `json:"hasMore"               jsonschema:"True when more projects exist beyond the returned page; pass nextOffset back as offset to fetch them."`
	NextOffset int32           `json:"nextOffset,omitempty"  jsonschema:"Offset to pass on the next call to fetch additional projects. Only set when hasMore is true."`
}

func registerListProjects(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[listProjectsInput, listProjectsOutput]{
		Name: "list_projects",
		Description: "Enumerate Run:ai project entities — \"which projects do I have?\", \"how many projects are there?\", " +
			"\"what projects are in department X?\". Lists every project the caller can see, including projects with no " +
			"resource quota configured on any node pool (which list_project_resources cannot see), and returns an " +
			"authoritative totalCount. Optionally filter by cluster or parent department. For quota and allocation " +
			"numbers, follow up with list_project_resources.",
		ReadOnly: true,
		Handler: func(ctx context.Context, in listProjectsInput) (listProjectsOutput, error) {
			return listProjectsRun(ctx, client, in)
		},
	})
}

func listProjectsRun(ctx context.Context, client *runaiclient.Client, in listProjectsInput) (listProjectsOutput, error) {
	out := listProjectsOutput{Projects: []projectEntity{}}

	parentID := in.DepartmentID
	if in.DepartmentName != "" {
		id, err := common.ResolveOrgUnitID(ctx, client, common.OrgTypeDepartment, in.DepartmentName, in.ClusterID)
		if err != nil {
			return out, err
		}
		parentID = id
	}

	var filters []string
	if in.ClusterID != "" {
		filters = append(filters, common.FilterEq("clusterId", in.ClusterID))
	}

	raw, err := fetchAllProjects(ctx, client, filters)
	if err != nil {
		return out, err
	}

	entities := make([]projectEntity, 0, len(raw))
	for _, p := range raw {
		// The parent filter runs in memory: parentId is not an established filterBy
		// key on the entity endpoint, and the full set is already fetched for local
		// pagination anyway.
		if parentID != "" && p.ParentId != parentID {
			continue
		}
		entities = append(entities, mapProjectEntity(p))
	}
	sort.SliceStable(entities, func(i, j int) bool { return entities[i].Name < entities[j].Name })

	out.TotalCount = int32(len(entities))
	out.Projects, out.HasMore, out.NextOffset = common.Paginate(entities, in.Offset, common.EffectiveLimit(in.Limit))
	return out, nil
}

func fetchAllProjects(ctx context.Context, client *runaiclient.Client, filters []string) ([]orgunit.Project, error) {
	var all []orgunit.Project
	offset := int32(0)
	for {
		req := client.OrgUnit.ProjectsAPI.GetProjects(ctx).Limit(common.UpstreamPageSize).Offset(offset)
		if len(filters) > 0 {
			req = req.FilterBy(filters)
		}
		resp, httpResp, err := req.Execute()
		if err != nil {
			return nil, fmt.Errorf("list projects: %w", runaimcp.MapHTTPError(httpResp, err))
		}
		if resp == nil {
			return nil, fmt.Errorf("nil response from upstream: %w", runaimcp.ErrUpstream)
		}
		all = append(all, resp.Projects...)
		if resp.Next == nil || *resp.Next <= offset {
			break
		}
		if int32(len(all)) >= common.MaxPaginatedResults {
			return nil, fmt.Errorf("result set exceeds the maximum supported size (%d projects); narrow the filter: %w", common.MaxPaginatedResults, runaimcp.ErrUpstream)
		}
		offset = *resp.Next
	}
	return all, nil
}

func mapProjectEntity(p orgunit.Project) projectEntity {
	var deptID *string
	if p.ParentId != "" {
		v := p.ParentId
		deptID = &v
	}
	var phase *string
	if p.Status.Phase != nil {
		v := string(*p.Status.Phase)
		phase = &v
	}
	return projectEntity{
		ID:           p.Id,
		Name:         p.Name,
		ClusterID:    p.ClusterId,
		DepartmentID: deptID,
		Phase:        phase,
	}
}
