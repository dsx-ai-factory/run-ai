/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package templates

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	workloads "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

var templateSortFields = map[string]bool{
	"name": true, "workloadType": true, "runnable": true, "scopeType": true,
	"clusterId": true, "projectId": true, "departmentId": true, "createdAt": true,
	"updatedAt": true, "createdBy": true, "updatedBy": true,
}

type listTemplatesInput struct {
	Name             string `json:"name,omitempty"             jsonschema:"Filter by exact template name."`
	WorkloadType     string `json:"workloadType,omitempty"     jsonschema:"Filter by native workload type. One of: Training, Workspace, Distributed, Inference, DistributedInference."`
	ScopeType        string `json:"scopeType,omitempty"        jsonschema:"Filter by scope type. One of: system, tenant, cluster, department, project."`
	ClusterID        string `json:"clusterId,omitempty"        jsonschema:"Filter by exact cluster UUID."`
	DepartmentID     string `json:"departmentId,omitempty"     jsonschema:"Filter by exact department ID."`
	ProjectID        *int32 `json:"projectId,omitempty"        jsonschema:"Filter by exact numeric project ID."`
	Search           string `json:"search,omitempty"           jsonschema:"Free-text template search passed to the List Templates API."`
	IncludeAncestors bool   `json:"includeAncestors,omitempty" jsonschema:"With exactly one of projectId or clusterId, request the API's expanded scope results."`
	SortBy           string `json:"sortBy,omitempty"           jsonschema:"Sort field. One of: name, workloadType, runnable, scopeType, clusterId, projectId, departmentId, createdAt, updatedAt, createdBy, updatedBy."`
	SortOrder        string `json:"sortOrder,omitempty"        jsonschema:"Sort direction: asc or desc. Defaults to asc."`
	Limit            int32  `json:"limit,omitempty"            jsonschema:"Maximum number of templates to return. Must be between 1 and 200 when set; defaults to 50."`
	Offset           int32  `json:"offset,omitempty"           jsonschema:"Pagination cursor from a previous response's nextOffset. Defaults to 0."`
}

func (in listTemplatesInput) Validate() error {
	if err := common.ValidatePagination(in.Limit, in.Offset); err != nil {
		return err
	}
	if in.WorkloadType != "" {
		if _, ok := templateAdapters[in.WorkloadType]; !ok {
			return unsupportedTemplateTypeError("workloadType")
		}
	}
	if in.ScopeType != "" && !workloads.ScopeType(in.ScopeType).IsValid() {
		return fmt.Errorf("scopeType must be one of: system, tenant, cluster, department, project")
	}
	if in.ClusterID != "" {
		if err := common.ValidateUUID("clusterId", in.ClusterID); err != nil {
			return err
		}
	}
	if in.ProjectID != nil && *in.ProjectID <= 0 {
		return fmt.Errorf("projectId must be positive")
	}
	if in.IncludeAncestors {
		hasCluster := in.ClusterID != ""
		hasProject := in.ProjectID != nil
		if hasCluster == hasProject {
			return fmt.Errorf("includeAncestors requires exactly one of clusterId or projectId")
		}
	}
	if in.Search != "" && strings.TrimSpace(in.Search) == "" {
		return fmt.Errorf("search must not be blank")
	}
	if in.SortBy != "" && !templateSortFields[in.SortBy] {
		return fmt.Errorf("sortBy must be one of: name, workloadType, runnable, scopeType, clusterId, projectId, departmentId, createdAt, updatedAt, createdBy, updatedBy")
	}
	if in.SortOrder != "" && in.SortOrder != "asc" && in.SortOrder != "desc" {
		return fmt.Errorf("sortOrder must be one of: asc, desc")
	}
	for _, field := range []struct{ name, value string }{
		{"name", in.Name}, {"departmentId", in.DepartmentID},
	} {
		if field.value != "" {
			if _, err := common.RequiredString(field.name, field.value); err != nil {
				return err
			}
			if err := common.ValidateFilterValue(field.name, field.value); err != nil {
				return err
			}
		}
	}
	return nil
}

type templateRunnableOutput struct {
	IsRunnable *bool   `json:"isRunnable,omitempty" jsonschema:"Whether the template can be submitted as-is."`
	Issue      *string `json:"issue,omitempty"      jsonschema:"Reason the template is not runnable, when provided by the API."`
}

type templateListItem struct {
	TemplateID   string                  `json:"templateId"             jsonschema:"Template UUID to pass to get_template or manage_template."`
	Name         string                  `json:"name"                   jsonschema:"Template name."`
	Description  *string                 `json:"description,omitempty"  jsonschema:"Template description."`
	WorkloadType string                  `json:"workloadType"           jsonschema:"Native workload type to pass to manage_template."`
	ScopeType    string                  `json:"scopeType"              jsonschema:"Template scope type."`
	ScopeID      string                  `json:"scopeId"                jsonschema:"ID of the template's scope."`
	ClusterID    *string                 `json:"clusterId,omitempty"    jsonschema:"Owning cluster UUID, when applicable."`
	DepartmentID *string                 `json:"departmentId,omitempty" jsonschema:"Owning department ID, when applicable."`
	ProjectID    *int32                  `json:"projectId,omitempty"    jsonschema:"Owning project ID, when applicable."`
	CreatedBy    string                  `json:"createdBy"              jsonschema:"User who created the template."`
	CreatedAt    string                  `json:"createdAt"              jsonschema:"Template creation timestamp in RFC3339."`
	UpdatedBy    *string                 `json:"updatedBy,omitempty"    jsonschema:"User who last updated the template."`
	UpdatedAt    *string                 `json:"updatedAt,omitempty"    jsonschema:"Last update timestamp in RFC3339."`
	Runnable     *templateRunnableOutput `json:"runnable,omitempty"     jsonschema:"Whether the template can be submitted as-is and any issue preventing it."`
}

type listTemplatesOutput struct {
	Templates  []templateListItem `json:"templates"            jsonschema:"Matching template metadata and runnable status."`
	HasMore    bool               `json:"hasMore"              jsonschema:"True when the API supplies another page offset; pass nextOffset back as offset."`
	NextOffset int32              `json:"nextOffset,omitempty" jsonschema:"Offset to pass on the next call. Only set when hasMore is true."`
}

func registerListTemplates(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[listTemplatesInput, listTemplatesOutput]{
		Name: "list_templates",
		Description: "List visible Run:ai workload templates across all native workload types and scopes. Supports exact " +
			"name, workload type, scope, cluster, department, and project filters, free-text search, sorting, ancestor " +
			"inclusion, and limit/offset pagination provided by the List Templates API. Each result includes the template " +
			"ID, scope metadata, and runnable status. Results intentionally omit the full spec and assets; call " +
			"get_template with a result's templateId to inspect the complete template or prepare an update. Call this " +
			"before manage_template when the template ID or workload type is unknown.",
		ReadOnly: true,
		Handler: func(ctx context.Context, in listTemplatesInput) (listTemplatesOutput, error) {
			return listTemplatesRun(ctx, client, in)
		},
	})
}

func listTemplatesRun(ctx context.Context, client *runaiclient.Client, in listTemplatesInput) (listTemplatesOutput, error) {
	pageLimit := common.EffectiveLimit(in.Limit)
	result, resp, err := newTemplateListRequest(ctx, client, in, pageLimit).Execute()
	if err != nil {
		return listTemplatesOutput{}, fmt.Errorf("list templates: %w", runaimcp.MapHTTPError(resp, err))
	}
	if result == nil {
		return listTemplatesOutput{}, fmt.Errorf("list templates: nil response: %w", runaimcp.ErrUpstream)
	}

	items, err := templateListItems(result.Entries, pageLimit)
	if err != nil {
		return listTemplatesOutput{}, err
	}
	hasMore, nextOffset, err := templateListPagination(result.Pagination, in.Offset, pageLimit)
	if err != nil {
		return listTemplatesOutput{}, err
	}

	return listTemplatesOutput{
		Templates:  items,
		HasMore:    hasMore,
		NextOffset: nextOffset,
	}, nil
}

func newTemplateListRequest(
	ctx context.Context,
	client *runaiclient.Client,
	in listTemplatesInput,
	pageLimit int32,
) workloads.ApiListTemplatesRequest {
	req := client.Workloads.WorkloadTemplatesAPI.ListTemplates(ctx).
		Offset(in.Offset).
		Limit(pageLimit)
	if filters := templateFilters(in); len(filters) > 0 {
		req = req.FilterBy(filters)
	}
	if in.Search != "" {
		req = req.Search(in.Search)
	}
	if in.IncludeAncestors {
		req = req.IncludeAncestors(true)
	}
	if in.SortBy != "" {
		req = req.SortBy(in.SortBy)
	}
	if in.SortOrder != "" {
		req = req.SortOrder(in.SortOrder)
	}
	return req
}

func templateListItems(entries []workloads.SupersetTemplate, pageLimit int32) ([]templateListItem, error) {
	if len(entries) > int(pageLimit) {
		return nil, fmt.Errorf(
			"list templates: response contains %d entries, exceeding requested limit %d: %w",
			len(entries), pageLimit, runaimcp.ErrUpstream)
	}

	items := make([]templateListItem, 0, len(entries))
	for i := range entries {
		items = append(items, templateListItemFromAPI(&entries[i]))
	}
	return items, nil
}

func templateListPagination(pagination *workloads.PaginationInfo, requestedOffset, pageLimit int32) (bool, int32, error) {
	if pagination == nil {
		return false, 0, nil
	}
	if pagination.Offset != nil {
		offset, err := templatePaginationInteger("offset", *pagination.Offset)
		if err != nil {
			return false, 0, fmt.Errorf("list templates pagination: %w", err)
		}
		if offset != requestedOffset {
			return false, 0, fmt.Errorf(
				"list templates pagination: response offset %d does not match requested offset %d: %w",
				offset, requestedOffset, runaimcp.ErrUpstream)
		}
	}
	if pagination.Limit != nil {
		limit, err := templatePaginationInteger("limit", *pagination.Limit)
		if err != nil {
			return false, 0, fmt.Errorf("list templates pagination: %w", err)
		}
		if limit != pageLimit {
			return false, 0, fmt.Errorf(
				"list templates pagination: response limit %d does not match requested limit %d: %w",
				limit, pageLimit, runaimcp.ErrUpstream)
		}
	}
	if pagination.NextOffset == nil {
		return false, 0, nil
	}

	nextOffset, err := templatePaginationInteger("nextOffset", *pagination.NextOffset)
	if err != nil {
		return false, 0, fmt.Errorf("list templates pagination: %w", err)
	}
	if nextOffset != 0 && nextOffset <= requestedOffset {
		return false, 0, fmt.Errorf(
			"list templates pagination: response nextOffset %d does not advance requested offset %d: %w",
			nextOffset, requestedOffset, runaimcp.ErrUpstream)
	}
	return nextOffset != 0, nextOffset, nil
}

func templateFilters(in listTemplatesInput) []string {
	filters := make([]string, 0, 6)
	if in.Name != "" {
		filters = append(filters, common.FilterEq("name", in.Name))
	}
	if in.WorkloadType != "" {
		filters = append(filters, common.FilterEq("workloadType", in.WorkloadType))
	}
	if in.ScopeType != "" {
		filters = append(filters, common.FilterEq("scopeType", in.ScopeType))
	}
	if in.ClusterID != "" {
		filters = append(filters, common.FilterEq("clusterId", in.ClusterID))
	}
	if in.DepartmentID != "" {
		filters = append(filters, common.FilterEq("departmentId", in.DepartmentID))
	}
	if in.ProjectID != nil {
		filters = append(filters, common.FilterEq("projectId", fmt.Sprintf("%d", *in.ProjectID)))
	}
	return filters
}

func templateListItemFromAPI(template *workloads.SupersetTemplate) templateListItem {
	meta := template.Meta
	out := templateListItem{
		TemplateID:   meta.Id,
		Name:         meta.Name,
		WorkloadType: string(meta.WorkloadType),
		ScopeType:    string(meta.ScopeType),
		ScopeID:      meta.ScopeId,
		CreatedBy:    meta.CreatedBy,
		CreatedAt:    meta.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
	if value, ok := meta.GetDescriptionOk(); ok && value != nil {
		out.Description = value
	}
	if value, ok := meta.GetClusterIdOk(); ok && value != nil {
		out.ClusterID = value
	}
	if value, ok := meta.GetDepartmentIdOk(); ok && value != nil {
		out.DepartmentID = value
	}
	if value, ok := meta.GetProjectIdOk(); ok && value != nil {
		out.ProjectID = value
	}
	if value, ok := meta.GetUpdatedByOk(); ok && value != nil {
		out.UpdatedBy = value
	}
	if value, ok := meta.GetUpdatedAtOk(); ok && value != nil {
		formatted := value.UTC().Format(time.RFC3339Nano)
		out.UpdatedAt = &formatted
	}
	if template.Runnable != nil {
		out.Runnable = &templateRunnableOutput{IsRunnable: template.Runnable.IsRunnable}
		if issue, ok := template.Runnable.GetIssueOk(); ok && issue != nil {
			out.Runnable.Issue = issue
		}
	}
	return out
}

func templatePaginationInteger(field string, value float32) (int32, error) {
	if value < 0 || float64(value) > float64(math.MaxInt32) || value != float32(int32(value)) {
		return 0, fmt.Errorf("response %s %v is not a non-negative 32-bit integer: %w",
			field, value, runaimcp.ErrUpstream)
	}
	return int32(value), nil
}
