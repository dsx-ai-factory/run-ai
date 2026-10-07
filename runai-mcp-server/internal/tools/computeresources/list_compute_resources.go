/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package computeresources

import (
	"context"
	"fmt"
	"strings"

	assets "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

type listComputeResourcesInput struct {
	ComputeResourceID  string `json:"computeResourceId,omitempty" jsonschema:"Filter by exact compute resource asset ID (UUID)."`
	Name               string `json:"name,omitempty"              jsonschema:"Filter by exact compute resource name."`
	Scope              string `json:"scope,omitempty"             jsonschema:"Filter by API scope. One of: tenant, cluster, department, project. tenant includes inherited system presets; other values are exact. Mutually exclusive with includeDescendants."` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	ClusterID          string `json:"clusterId,omitempty"         jsonschema:"Filter by cluster UUID. Higher-scope resources may also be returned unless scope is set."`
	DepartmentID       string `json:"departmentId,omitempty"      jsonschema:"Filter by department ID. Higher-scope resources may also be returned unless scope is set."`
	ProjectID          *int32 `json:"projectId,omitempty"         jsonschema:"Filter by positive numeric project ID. Higher-scope resources may also be returned unless scope is set."`
	IncludeDescendants bool   `json:"includeDescendants,omitempty" jsonschema:"Include compute resources from descendant scopes. Requires at least one of clusterId, departmentId, or projectId and excludes scope."`
	Limit              int32  `json:"limit,omitempty"             jsonschema:"Maximum number of compute resources to return. Defaults to 50, capped at 200."`
	Offset             int32  `json:"offset,omitempty"            jsonschema:"Pagination cursor from a previous response's nextOffset. Defaults to 0."`
}

func (in listComputeResourcesInput) Validate() error {
	if in.ComputeResourceID != "" {
		if err := common.ValidateUUID("computeResourceId", in.ComputeResourceID); err != nil {
			return err
		}
	}
	if in.Name != "" {
		if _, err := common.RequiredString("name", in.Name); err != nil {
			return err
		}
	}
	switch assets.Scope(in.Scope) {
	case "", assets.SCOPE_TENANT, assets.SCOPE_CLUSTER, assets.SCOPE_DEPARTMENT, assets.SCOPE_PROJECT:
	default:
		return fmt.Errorf("scope must be one of: tenant, cluster, department, project")
	}
	if in.ClusterID != "" {
		if err := common.ValidateUUID(computeResourceClusterIDField, in.ClusterID); err != nil {
			return err
		}
	}
	if in.ProjectID != nil && *in.ProjectID <= 0 {
		return fmt.Errorf("projectId must be positive")
	}
	if in.DepartmentID != "" {
		if _, err := common.RequiredString(computeResourceDepartmentIDField, in.DepartmentID); err != nil {
			return err
		}
	}
	if in.IncludeDescendants {
		if in.Scope != "" {
			return fmt.Errorf("includeDescendants and scope are mutually exclusive")
		}
		if in.ClusterID == "" && in.DepartmentID == "" && in.ProjectID == nil {
			return fmt.Errorf("includeDescendants requires at least one of clusterId, departmentId, or projectId")
		}
	}
	return common.ValidatePagination(in.Limit, in.Offset)
}

type computeResourceListItem struct {
	ComputeResourceID   string         `json:"computeResourceId"  jsonschema:"Compute resource UUID to pass to manage_compute_resource."`
	Name                string         `json:"name"               jsonschema:"Compute resource name."`
	Scope               string         `json:"scope"              jsonschema:"Compute resource scope."`
	ComputeResource     map[string]any `json:"computeResource"    jsonschema:"Complete read-only compute resource returned by the Assets API, including metadata and compute spec."`
	UpdateRequest       map[string]any `json:"updateRequest,omitempty"       jsonschema:"Complete full-replacement request accepted by manage_compute_resource update. Modify only the intended fields and keep all others. Omitted when this MCP version cannot safely update the resource."` //nolint:lll // complete update-safety description
	ExpectedFingerprint string         `json:"expectedFingerprint,omitempty" jsonschema:"Best-effort stale-snapshot fingerprint to pass unchanged to manage_compute_resource update. Omitted when updateRequest is unavailable."`                                                              //nolint:lll // long tool-safety description; cannot wrap a struct tag
	UpdateBlockedReason string         `json:"updateBlockedReason,omitempty" jsonschema:"Why this resource cannot be updated. Present only when updateRequest is unavailable; listing remains supported."`                                                                                     //nolint:lll // long tool-safety description; cannot wrap a struct tag
}

type listComputeResourcesOutput struct {
	ComputeResources []computeResourceListItem `json:"computeResources"       jsonschema:"Matching compute resources with complete specs and update preparation status."`
	TotalCount       int32                     `json:"totalCount"             jsonschema:"Total number of compute resources matching the filters across all local pages."`
	HasMore          bool                      `json:"hasMore"                jsonschema:"True when more compute resources exist beyond the returned page; pass nextOffset back as offset."`
	NextOffset       int32                     `json:"nextOffset,omitempty"   jsonschema:"Offset to pass on the next call. Only set when hasMore is true."`
}

type rawComputeResourceListItem struct {
	id    string
	name  string
	scope string
	raw   map[string]any
}

func registerListComputeResources(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[listComputeResourcesInput, listComputeResourcesOutput]{
		Name: "list_compute_resources",
		Description: "List visible Run:ai compute resource assets with their complete metadata and compute specification. " +
			"Supports exact ID and name filters, API-native scope, cluster, department, and project filters, descendant inclusion, and local limit/offset pagination. " +
			"Supported results include an update-ready full-replacement request and expectedFingerprint; pass both unchanged to " +
			"manage_compute_resource after editing only the intended fields. A result instead includes updateBlockedReason when this " +
			"MCP version cannot safely preserve newer API fields or the result is a read-only system preset. " +
			"Call without filters to list all visible compute resources.",
		ReadOnly: true,
		Handler: func(ctx context.Context, in listComputeResourcesInput) (listComputeResourcesOutput, error) {
			return listComputeResourcesRun(ctx, client, in)
		},
	})
}

func listComputeResourcesRun(ctx context.Context, client *runaiclient.Client, in listComputeResourcesInput) (listComputeResourcesOutput, error) {
	req := client.Assets.ComputeAPI.ListComputeAssets(ctx)
	if in.Name != "" {
		req = req.Name(in.Name)
	}
	if in.Scope != "" {
		req = req.Scope(in.Scope)
	}
	if in.ClusterID != "" {
		req = req.ClusterId(in.ClusterID)
	}
	if in.DepartmentID != "" {
		req = req.DepartmentId(in.DepartmentID)
	}
	if in.ProjectID != nil {
		req = req.ProjectId(*in.ProjectID)
	}
	if in.IncludeDescendants {
		req = req.IncludeDescendants(true)
	}

	result, resp, err := req.Execute()
	if err != nil {
		return listComputeResourcesOutput{}, fmt.Errorf("list compute resources: %w", runaimcp.MapHTTPError(resp, err))
	}
	if result == nil {
		return listComputeResourcesOutput{}, fmt.Errorf("list compute resources: nil response: %w", runaimcp.ErrUpstream)
	}
	raw, err := readComputeResourceResponse(resp, "list compute resources")
	if err != nil {
		return listComputeResourcesOutput{}, err
	}
	rawEntries, ok := raw["entries"].([]any)
	if !ok {
		return listComputeResourcesOutput{}, fmt.Errorf("list compute resources: entries must be an array: %w", runaimcp.ErrUpstream)
	}
	if len(rawEntries) != len(result.Entries) {
		return listComputeResourcesOutput{}, fmt.Errorf("list compute resources: decoded %d entries but raw response contains %d: %w",
			len(result.Entries), len(rawEntries), runaimcp.ErrUpstream)
	}

	matched := make([]rawComputeResourceListItem, 0, len(rawEntries))
	for i, value := range rawEntries {
		entry, ok := value.(map[string]any)
		if !ok {
			return listComputeResourcesOutput{}, fmt.Errorf("list compute resources: entry %d must be an object: %w", i, runaimcp.ErrUpstream)
		}
		identity, err := computeResourceIdentity(entry)
		if err != nil {
			return listComputeResourcesOutput{}, fmt.Errorf("list compute resources: entry %d: %w", i, err)
		}
		if in.ComputeResourceID != "" && !strings.EqualFold(identity.id, in.ComputeResourceID) {
			continue
		}
		if in.Name != "" && identity.name != in.Name {
			continue
		}
		if in.Scope != "" && !computeResourceMatchesScope(identity.scope, in.Scope) {
			continue
		}
		matched = append(matched, rawComputeResourceListItem{
			id: identity.id, name: identity.name, scope: identity.scope, raw: entry,
		})
	}

	page, hasMore, nextOffset := common.Paginate(matched, in.Offset, common.EffectiveLimit(in.Limit))
	items := make([]computeResourceListItem, 0, len(page))
	for _, resource := range page {
		item := computeResourceListItem{
			ComputeResourceID: resource.id,
			Name:              resource.name,
			Scope:             resource.scope,
			ComputeResource:   resource.raw,
		}
		if resource.scope == string(assets.SCOPE_SYSTEM) {
			item.UpdateBlockedReason = "system-scoped compute resources are read-only"
			items = append(items, item)
			continue
		}
		updateRequest, fingerprint, err := computeResourceUpdateSnapshot(resource.raw)
		if err != nil {
			item.UpdateBlockedReason = fmt.Sprintf("cannot prepare a safe update request: %v", err)
		} else if err := validateComputeResourceUpdateRequest(updateRequest); err != nil {
			item.UpdateBlockedReason = fmt.Sprintf("this MCP version cannot safely preserve the complete update request: %v", err)
		} else {
			item.UpdateRequest = updateRequest
			item.ExpectedFingerprint = fingerprint
		}
		items = append(items, item)
	}

	return listComputeResourcesOutput{
		ComputeResources: items,
		TotalCount:       int32(len(matched)),
		HasMore:          hasMore,
		NextOffset:       nextOffset,
	}, nil
}

func computeResourceMatchesScope(actual, requested string) bool {
	if requested == string(assets.SCOPE_TENANT) {
		return actual == requested || actual == string(assets.SCOPE_SYSTEM)
	}
	return actual == requested
}
