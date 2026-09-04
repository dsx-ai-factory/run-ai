/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package aiapplications

import (
	"context"
	"fmt"

	workloads "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

const maxAIApplicationsLimit int32 = 500

type listAIApplicationsInput struct {
	ClusterID      string `json:"clusterId,omitempty"      jsonschema:"Cluster UUID to scope the AI application list."`
	ProjectID      string `json:"projectId,omitempty"      jsonschema:"Project ID to scope the AI application list."`
	Name           string `json:"name,omitempty"           jsonschema:"Exact AI application name to match."`
	ProjectName    string `json:"projectName,omitempty"    jsonschema:"Exact project name to match."`
	DepartmentID   string `json:"departmentId,omitempty"   jsonschema:"Exact department ID to match."`
	DepartmentName string `json:"departmentName,omitempty" jsonschema:"Exact department name to match."`
	Source         string `json:"source,omitempty"         jsonschema:"Exact AI application source to match."`
	Type           string `json:"type,omitempty"           jsonschema:"Exact AI application type to match."`
	Search         string `json:"search,omitempty"         jsonschema:"Free-text search passed to the List AI Applications API."`
	SortBy         string `json:"sortBy,omitempty"         jsonschema:"API field used to sort results: clusterId, name, projectId, projectName, departmentId, departmentName, source, type, or createdAt."`
	SortOrder      string `json:"sortOrder,omitempty"      jsonschema:"API sort direction: asc or desc. Defaults to asc."`
	Limit          int32  `json:"limit,omitempty"          jsonschema:"Maximum number of AI applications to return. Defaults to 50, capped at 500 by the API."`
	Offset         int32  `json:"offset,omitempty"         jsonschema:"Pagination cursor from a previous response's nextOffset. Defaults to 0."`
}

func (in listAIApplicationsInput) Validate() error {
	if in.Limit < 0 || in.Limit > maxAIApplicationsLimit {
		return fmt.Errorf("limit must be between 0 and %d", maxAIApplicationsLimit)
	}
	if in.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}
	for _, field := range []struct{ name, value string }{
		{"clusterId", in.ClusterID}, {"projectId", in.ProjectID}, {"name", in.Name}, {"projectName", in.ProjectName},
		{"departmentId", in.DepartmentID}, {"departmentName", in.DepartmentName},
		{"source", in.Source}, {"type", in.Type},
	} {
		if err := common.ValidateFilterValue(field.name, field.value); err != nil {
			return err
		}
	}
	return nil
}

type aiApplicationListItem struct {
	ID             string `json:"id"                       jsonschema:"AI application ID."`
	Name           string `json:"name"                     jsonschema:"AI application name."`
	Source         string `json:"source,omitempty"         jsonschema:"AI application source."`
	Type           string `json:"type,omitempty"           jsonschema:"AI application type."`
	DesiredPhase   string `json:"desiredPhase,omitempty"   jsonschema:"Desired lifecycle phase."`
	Phase          string `json:"phase,omitempty"          jsonschema:"Current Kubernetes or Helm release phase."`
	ActionStatus   string `json:"actionStatus,omitempty"   jsonschema:"Current AI application action status, such as creating, deployed, or failed."`
	ActionMessage  string `json:"actionMessage,omitempty"  jsonschema:"Message associated with the current AI application action status."`
	ProjectID      string `json:"projectId,omitempty"      jsonschema:"Owning project ID."`
	ProjectName    string `json:"projectName,omitempty"    jsonschema:"Owning project name."`
	DepartmentID   string `json:"departmentId,omitempty"   jsonschema:"Owning department ID."`
	DepartmentName string `json:"departmentName,omitempty" jsonschema:"Owning department name."`
	CreatedAt      string `json:"createdAt,omitempty"      jsonschema:"Creation timestamp in RFC3339 UTC."`
	CreatedBy      string `json:"createdBy,omitempty"      jsonschema:"User who created the AI application."`
	UpdatedAt      string `json:"updatedAt,omitempty"      jsonschema:"Last update timestamp in RFC3339 UTC."`
	UpdatedBy      string `json:"updatedBy,omitempty"      jsonschema:"User who last updated the AI application."`
}

type listAIApplicationsOutput struct {
	Applications []aiApplicationListItem `json:"applications"         jsonschema:"AI applications matching the filters. Always present; empty array when none match."`
	HasMore      bool                    `json:"hasMore"              jsonschema:"True when more AI applications exist beyond the returned page."`
	NextOffset   int32                   `json:"nextOffset,omitempty" jsonschema:"Offset to pass on the next call. Only set when hasMore is true."`
}

func registerListAIApplications(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[listAIApplicationsInput, listAIApplicationsOutput]{
		Name: "list_ai_applications",
		Description: "Call when the user asks which AI applications exist, their current status, or who owns them. " +
			"Lists the AI applications the caller can access and returns application IDs, desired/action/release phases, " +
			"ownership, source, type, and timestamps with offset pagination. Optionally filter by cluster, project, " +
			"department, source, type, or exact application name, or use free-text search. " +
			"Use submit_ai_application to create an application. Do not use this tool to list the workloads belonging " +
			"to an application; use list_workloads with its aiApplicationId or aiApplicationName filter.",
		ReadOnly: true,
		Handler: func(ctx context.Context, in listAIApplicationsInput) (listAIApplicationsOutput, error) {
			return listAIApplicationsRun(ctx, client, in)
		},
	})
}

func listAIApplicationsRun(ctx context.Context, client *runaiclient.Client, in listAIApplicationsInput) (listAIApplicationsOutput, error) {
	out := listAIApplicationsOutput{Applications: []aiApplicationListItem{}}
	req := client.Workloads.AIApplicationsAPI.ListApplications(ctx).
		Limit(common.EffectiveLimit(in.Limit)).
		Offset(in.Offset)
	if filters := aiApplicationFilters(in); len(filters) > 0 {
		req = req.FilterBy(filters)
	}
	if in.Search != "" {
		req = req.Search(in.Search)
	}
	if in.SortBy != "" {
		req = req.SortBy(workloads.AIApplicationsSortFilterFields(in.SortBy))
	}
	if in.SortOrder != "" {
		req = req.SortOrder(in.SortOrder)
	}
	response, httpResp, err := req.Execute()
	if err != nil {
		return out, fmt.Errorf("list AI applications: %w", runaimcp.MapHTTPError(httpResp, err))
	}
	if response == nil {
		return out, fmt.Errorf("list AI applications: nil response: %w", runaimcp.ErrUpstream)
	}
	for i := range response.Applications {
		out.Applications = append(out.Applications, mapAIApplication(response.Applications[i]))
	}
	if response.Next != nil && *response.Next > in.Offset {
		out.HasMore = true
		out.NextOffset = *response.Next
	}
	return out, nil
}

func aiApplicationFilters(in listAIApplicationsInput) []string {
	fields := []struct{ name, value string }{
		{"clusterId", in.ClusterID}, {"projectId", in.ProjectID}, {"name", in.Name},
		{"projectName", in.ProjectName}, {"departmentId", in.DepartmentID},
		{"departmentName", in.DepartmentName}, {"source", in.Source}, {"type", in.Type},
	}
	filters := make([]string, 0, len(fields))
	for _, field := range fields {
		if field.value != "" {
			filters = append(filters, common.FilterEq(field.name, field.value))
		}
	}
	return filters
}

func mapAIApplication(application workloads.AIApplication) aiApplicationListItem {
	return aiApplicationListItem{
		ID: application.GetId(), Name: application.GetName(), Source: application.GetSource(), Type: application.GetType(),
		DesiredPhase: application.GetDesiredPhase(), Phase: application.GetPhase(), ProjectID: application.GetProjectId(),
		ActionStatus: application.GetActionStatus(), ActionMessage: application.GetActionMessage(),
		ProjectName: application.GetProjectName(), DepartmentID: application.GetDepartmentId(), DepartmentName: application.GetDepartmentName(),
		CreatedAt: common.TimestampString(application.CreatedAt), CreatedBy: application.GetCreatedBy(),
		UpdatedAt: common.TimestampString(application.UpdatedAt), UpdatedBy: application.GetUpdatedBy(),
	}
}
