/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package environments

import (
	"context"
	"fmt"
	"strings"

	assets "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

type listEnvironmentsInput struct {
	EnvironmentID        string `json:"environmentId,omitempty"       jsonschema:"Filter by exact environment asset ID (UUID)."`
	Name                 string `json:"name,omitempty"                jsonschema:"Filter by exact environment name."`
	Scope                string `json:"scope,omitempty"               jsonschema:"Filter by API scope. One of: tenant, cluster, department, project. tenant includes inherited system presets; other values are exact. Mutually exclusive with includeDescendants."` //nolint:lll
	ClusterID            string `json:"clusterId,omitempty"           jsonschema:"Filter by cluster UUID. Higher-scope environments may also be returned unless scope is set."`
	DepartmentID         string `json:"departmentId,omitempty"        jsonschema:"Filter by department ID. Higher-scope environments may also be returned unless scope is set."`
	ProjectID            *int32 `json:"projectId,omitempty"           jsonschema:"Filter by positive numeric project ID. Higher-scope environments may also be returned unless scope is set."`
	WorkloadType         string `json:"workloadType,omitempty"        jsonschema:"Filter by supported workload type. One of: Workspace, Training, Distributed, Inference, DistributedInference."`
	DistributedFramework string `json:"distributedFramework,omitempty" jsonschema:"Filter distributed environments by framework name, such as MPI or PyTorch."`
	IncludeDescendants   bool   `json:"includeDescendants,omitempty"  jsonschema:"Include environments from descendant scopes. Requires at least one of clusterId, departmentId, or projectId and excludes scope."`
	Limit                int32  `json:"limit,omitempty"               jsonschema:"Maximum number of environments to return. Defaults to 50, capped at 200."`
	Offset               int32  `json:"offset,omitempty"              jsonschema:"Pagination cursor from a previous response's nextOffset. Defaults to 0."`
}

func (in listEnvironmentsInput) Validate() error {
	if in.EnvironmentID != "" {
		if err := common.ValidateUUID("environmentId", in.EnvironmentID); err != nil {
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
		if err := common.ValidateUUID("clusterId", in.ClusterID); err != nil {
			return err
		}
	}
	if in.DepartmentID != "" {
		if _, err := common.RequiredString("departmentId", in.DepartmentID); err != nil {
			return err
		}
	}
	if in.ProjectID != nil && *in.ProjectID <= 0 {
		return fmt.Errorf("projectId must be positive")
	}
	switch in.WorkloadType {
	case "", "Workspace", "Training", "Distributed", "Inference", "DistributedInference":
	default:
		return fmt.Errorf("workloadType must be one of: Workspace, Training, Distributed, Inference, DistributedInference")
	}
	if in.DistributedFramework != "" {
		if _, err := common.RequiredString("distributedFramework", in.DistributedFramework); err != nil {
			return err
		}
		if !assets.DistributedFramework(in.DistributedFramework).IsValid() {
			return fmt.Errorf("distributedFramework must be one of: MPI, PyTorch, TF, XGBoost, JAX")
		}
		if in.WorkloadType != "" && in.WorkloadType != "Distributed" {
			return fmt.Errorf("distributedFramework requires workloadType=Distributed or an omitted workloadType")
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

type environmentListItem struct {
	EnvironmentID       string         `json:"environmentId"                   jsonschema:"Environment asset UUID to pass to manage_environment."`
	Name                string         `json:"name"                            jsonschema:"Environment name."`
	Scope               string         `json:"scope"                           jsonschema:"Environment scope."`
	Environment         map[string]any `json:"environment"                     jsonschema:"Complete environment returned by the Assets API."`
	UpdateRequest       map[string]any `json:"updateRequest,omitempty"         jsonschema:"Complete full-replacement request accepted by manage_environment update. Modify only intended fields and keep all others."` //nolint:lll
	ExpectedUpdateCount *int64         `json:"expectedUpdateCount,omitempty"   jsonschema:"API update count to pass unchanged to manage_environment update for a best-effort stale-write check."`                      //nolint:lll
	UpdateBlockedReason string         `json:"updateBlockedReason,omitempty"   jsonschema:"Why this environment cannot be updated. Present only when updateRequest is unavailable."`
}

type listEnvironmentsOutput struct {
	Environments []environmentListItem `json:"environments"          jsonschema:"Matching environments with complete specs and update preparation status."`
	TotalCount   int32                 `json:"totalCount"            jsonschema:"Total number of environments matching the filters across all local pages."`
	HasMore      bool                  `json:"hasMore"               jsonschema:"True when more environments exist beyond the returned page; pass nextOffset back as offset."`
	NextOffset   int32                 `json:"nextOffset,omitempty"  jsonschema:"Offset to pass on the next call. Only set when hasMore is true."`
}

func registerListEnvironments(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[listEnvironmentsInput, listEnvironmentsOutput]{
		Name: "list_environments",
		Description: "List visible Run:ai environment assets with complete metadata and container specifications. Supports exact " +
			"ID and name filters, API-native scope and org-unit filters, supported workload type and distributed framework filters, " +
			"descendant inclusion, and local limit/offset pagination. Supported non-system results include an update-ready full-replacement " +
			"request and expectedUpdateCount; edit only intended fields and pass both to manage_environment. Results that cannot be safely " +
			"updated include updateBlockedReason. Call without filters to list all visible environments.",
		ReadOnly: true,
		Handler: func(ctx context.Context, in listEnvironmentsInput) (listEnvironmentsOutput, error) {
			return listEnvironmentsRun(ctx, client, in)
		},
	})
}

func listEnvironmentsRun(ctx context.Context, client *runaiclient.Client, in listEnvironmentsInput) (listEnvironmentsOutput, error) {
	if in.hasOnlyIDFilter() {
		return listEnvironmentByID(ctx, client, in)
	}

	req := client.Assets.EnvironmentAPI.ListEnvironmentAssets(ctx)
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
	if in.DistributedFramework != "" {
		req = req.DistributedFramework(in.DistributedFramework).IsDistributed(true)
	}
	switch in.WorkloadType {
	case "Workspace":
		req = req.IsWorkspace(true)
	case "Training":
		req = req.IsTraining(true)
	case "Distributed":
		req = req.IsDistributed(true)
	case "Inference":
		req = req.IsInference(true)
	case "DistributedInference":
		req = req.IsDistributedInference(true)
	}

	result, resp, err := req.Execute()
	if err != nil {
		return listEnvironmentsOutput{}, fmt.Errorf("list environments: %w", runaimcp.MapHTTPError(resp, err))
	}
	if result == nil {
		return listEnvironmentsOutput{}, fmt.Errorf("list environments: nil response: %w", runaimcp.ErrUpstream)
	}
	rawResponse, err := readEnvironmentResponse(resp, "list environments")
	if err != nil {
		return listEnvironmentsOutput{}, err
	}
	rawEntries, ok := rawResponse["entries"].([]any)
	if !ok {
		return listEnvironmentsOutput{}, fmt.Errorf("list environments: entries must be an array: %w", runaimcp.ErrUpstream)
	}
	if len(rawEntries) != len(result.Entries) {
		return listEnvironmentsOutput{}, fmt.Errorf("list environments: decoded %d entries but raw response contains %d: %w",
			len(result.Entries), len(rawEntries), runaimcp.ErrUpstream)
	}

	type candidate struct {
		environment *assets.EnvironmentAsset
		raw         map[string]any
	}
	matched := make([]candidate, 0, len(result.Entries))
	for index := range result.Entries {
		environment := &result.Entries[index]
		raw, ok := rawEntries[index].(map[string]any)
		if !ok || raw == nil {
			return listEnvironmentsOutput{}, fmt.Errorf("list environments: entry %d must be an object: %w", index, runaimcp.ErrUpstream)
		}
		if in.EnvironmentID != "" && !strings.EqualFold(environment.Meta.Id, in.EnvironmentID) {
			continue
		}
		if in.Name != "" && environment.Meta.Name != in.Name {
			continue
		}
		if in.Scope != "" && !environmentMatchesScope(environment.Meta.Scope, in.Scope) {
			continue
		}
		matched = append(matched, candidate{environment: environment, raw: raw})
	}

	page, hasMore, nextOffset := common.Paginate(matched, in.Offset, common.EffectiveLimit(in.Limit))
	items := make([]environmentListItem, 0, len(page))
	for index, entry := range page {
		snapshot, err := newEnvironmentSnapshot(entry.environment, entry.raw)
		if err != nil {
			return listEnvironmentsOutput{}, fmt.Errorf("list environments: page entry %d: %w", index, err)
		}
		items = append(items, environmentListItemFromSnapshot(snapshot))
	}

	return listEnvironmentsOutput{Environments: items, TotalCount: int32(len(matched)), HasMore: hasMore, NextOffset: nextOffset}, nil
}

func (in listEnvironmentsInput) hasOnlyIDFilter() bool {
	return in.EnvironmentID != "" && in.Name == "" && in.Scope == "" && in.ClusterID == "" &&
		in.DepartmentID == "" && in.ProjectID == nil && in.WorkloadType == "" && in.DistributedFramework == "" &&
		!in.IncludeDescendants
}

func listEnvironmentByID(
	ctx context.Context,
	client *runaiclient.Client,
	in listEnvironmentsInput,
) (listEnvironmentsOutput, error) {
	snapshot, err := getEnvironment(ctx, client, in.EnvironmentID)
	if err != nil {
		if common.IsNotFound(err) {
			return listEnvironmentsOutput{Environments: []environmentListItem{}}, nil
		}
		return listEnvironmentsOutput{}, err
	}
	all := []environmentSnapshot{snapshot}
	page, hasMore, nextOffset := common.Paginate(all, in.Offset, common.EffectiveLimit(in.Limit))
	items := make([]environmentListItem, 0, len(page))
	for _, entry := range page {
		items = append(items, environmentListItemFromSnapshot(entry))
	}
	return listEnvironmentsOutput{
		Environments: items,
		TotalCount:   1,
		HasMore:      hasMore,
		NextOffset:   nextOffset,
	}, nil
}

func environmentListItemFromSnapshot(snapshot environmentSnapshot) environmentListItem {
	item := environmentListItem{
		EnvironmentID: snapshot.id,
		Name:          snapshot.name,
		Scope:         string(snapshot.scope),
		Environment:   snapshot.environment,
	}
	switch {
	case snapshot.scope == assets.SCOPE_SYSTEM:
		item.UpdateBlockedReason = "system-scoped environments are read-only"
	case snapshot.updateCount == nil:
		item.UpdateBlockedReason = "the Assets API response did not include meta.updateCount"
	case snapshot.updateBlockedReason != "":
		item.UpdateBlockedReason = snapshot.updateBlockedReason
	default:
		item.UpdateRequest = snapshot.updateRequest
		item.ExpectedUpdateCount = snapshot.updateCount
	}
	return item
}
