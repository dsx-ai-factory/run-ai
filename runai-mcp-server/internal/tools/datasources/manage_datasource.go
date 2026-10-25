/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datasources

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	assets "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

type manageDatasourceInput struct {
	Kind                assets.AssetKind   `json:"kind" jsonschema:"API asset kind. One of: nfs, pvc, s3, git, hostPath, config-map, secret-volume. Use list_datasources to resolve it."`
	Action              common.WriteAction `json:"action"                          jsonschema:"What to do. One of: create, update, delete."`
	DatasourceID        string             `json:"datasourceId,omitempty"         jsonschema:"Datasource UUID for update/delete. Mutually exclusive with name. Must not be set for create."`
	Name                string             `json:"name,omitempty" jsonschema:"Exact name for update/delete, mutually exclusive with datasourceId. Name and kind must identify one asset within the selected scope."`
	Scope               assets.Scope       `json:"scope,omitempty" jsonschema:"Optional scope for name lookup: tenant, cluster, department, or project. Use scopeId or scopeName to select a specific scope."`
	ScopeID             string             `json:"scopeId,omitempty" jsonschema:"Scope identifier for name lookup: project numeric ID, department ID, or cluster UUID. Requires scope; mutually exclusive with scopeName."`
	ScopeName           string             `json:"scopeName,omitempty" jsonschema:"Project, department, or cluster name. Create uses request.meta.scope; update/delete use scope. Mutually exclusive with explicit scope IDs."`
	Request             map[string]any     `json:"request,omitempty"               jsonschema:"Complete Assets API request for create or update. Start updates from list_datasources.updateRequest and edit only intended fields."` //nolint:lll
	ExpectedUpdateCount *int64             `json:"expectedUpdateCount,omitempty"   jsonschema:"Required for update. Pass unchanged from list_datasources for a best-effort stale-write check."`
	IgnoreMissing       bool               `json:"ignoreMissing,omitempty"         jsonschema:"Delete-only. Treat an already-missing datasource as a successful no-op."`
	Confirm             bool               `json:"confirm,omitempty"               jsonschema:"Required for update and delete without interactive confirmation."`
}

func (in manageDatasourceInput) Validate() error {
	if _, ok := datasourceAdapters[in.Kind]; !ok {
		return fmt.Errorf("unsupported datasource kind %q; supported kinds: %v", in.Kind, slices.Sorted(maps.Keys(datasourceAdapters)))
	}
	action := in.Action
	if !action.Valid() {
		return fmt.Errorf("action must be one of: create, update, delete")
	}
	if in.DatasourceID != "" {
		if err := common.ValidateUUID("datasourceId", in.DatasourceID); err != nil {
			return err
		}
	}
	if err := in.validateTarget(); err != nil {
		return err
	}
	if in.IgnoreMissing && action != common.ActionDelete {
		return fmt.Errorf("ignoreMissing is only valid with action=delete")
	}

	switch action {
	case common.ActionCreate:
		if in.DatasourceID != "" {
			return fmt.Errorf("datasourceId must not be set for create")
		}
		if in.ExpectedUpdateCount != nil {
			return fmt.Errorf("expectedUpdateCount must not be set for create")
		}
		return validateDatasourceCreateRequest(in.Kind, in.Request)
	case common.ActionUpdate:
		if in.DatasourceID == "" && in.Name == "" {
			return fmt.Errorf("datasourceId or name is required for update")
		}
		if in.ExpectedUpdateCount == nil {
			return fmt.Errorf("expectedUpdateCount is required for update; call list_datasources")
		}
		if *in.ExpectedUpdateCount < 0 {
			return fmt.Errorf("expectedUpdateCount must be non-negative")
		}
		return validateDatasourceUpdateRequest(in.Kind, in.Request)
	case common.ActionDelete:
		if in.DatasourceID == "" && in.Name == "" {
			return fmt.Errorf("datasourceId or name is required for delete")
		}
		if len(in.Request) != 0 {
			return fmt.Errorf("request must not be set for delete")
		}
		if in.ExpectedUpdateCount != nil {
			return fmt.Errorf("expectedUpdateCount must not be set for delete")
		}
		return nil
	default:
		return fmt.Errorf("unsupported action %q", in.Action)
	}
}

type manageDatasourceOutput struct {
	Kind         assets.AssetKind   `json:"kind" jsonschema:"API asset kind of the targeted datasource."`
	Action       common.WriteAction `json:"action"                  jsonschema:"The action performed: create, update, or delete."`
	Applied      bool               `json:"applied"                 jsonschema:"True when the upstream service accepted the change. False for an ignored-missing delete."`
	DatasourceID string             `json:"datasourceId,omitempty" jsonschema:"Created or targeted datasource UUID."`
	Name         string             `json:"name,omitempty"          jsonschema:"Datasource name."`
	Datasource   map[string]any     `json:"datasource,omitempty"   jsonschema:"Datasource returned by the Assets API after create or update."`
	Warnings     []string           `json:"warnings,omitempty"      jsonschema:"Non-fatal notes, including ignored missing deletes or accepted writes with unverifiable responses."`
}

func registerManageDatasource(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[manageDatasourceInput, *manageDatasourceOutput]{
		Name: "manage_datasource",
		Description: "Create, update, or delete an NFS, PVC, S3, Git, host path, ConfigMap, or secret-volume asset. " +
			"Update/delete select by UUID or exact name, optionally with scope and scopeId/scopeName; ambiguous names require more context. " +
			"Create accepts API meta and spec, optionally resolving scopeName; update requires list_datasources.updateRequest and expectedUpdateCount. " +
			"ConfigMap/secret creation requires cluster, department, or project scope; their updates change metadata only. " +
			"PVC updates accept path and claimInfo size/addedAttrValues. " +
			"Update/delete require confirmation. System assets are read-only.",
		ReadOnly:    false,
		Destructive: true,
		Handler: func(ctx context.Context, in manageDatasourceInput) (*manageDatasourceOutput, error) {
			return manageDatasourceRun(ctx, client, in)
		},
	})
}

func manageDatasourceRun(ctx context.Context, client *runaiclient.Client, in manageDatasourceInput) (*manageDatasourceOutput, error) {
	if in.Action != common.ActionCreate && in.Name != "" {
		id, err := resolveDatasourceID(ctx, client, in)
		if err != nil {
			if in.Action == common.ActionDelete && in.IgnoreMissing && common.IsNotFound(err) {
				return ignoredMissingDatasourceOutput(in), nil
			}
			return newManageDatasourceOutput(in), err
		}
		in.DatasourceID = id
	}
	switch in.Action {
	case common.ActionCreate:
		return createDatasource(ctx, client, in)
	case common.ActionUpdate:
		return updateDatasource(ctx, client, in)
	case common.ActionDelete:
		return deleteDatasource(ctx, client, in)
	default:
		return newManageDatasourceOutput(in), fmt.Errorf("unsupported action %q: %w", in.Action, runaimcp.ErrValidation)
	}
}

func createDatasource(ctx context.Context, client *runaiclient.Client, in manageDatasourceInput) (*manageDatasourceOutput, error) {
	out := newManageDatasourceOutput(in)
	request, err := datasourceCreateRequest(ctx, client, in)
	if err != nil {
		return out, err
	}
	in.Request = request
	meta, _ := request["meta"].(map[string]any)
	out.Name, _ = meta["name"].(string)
	resp, err := datasourceAdapters[in.Kind].create(ctx, client, request)

	return datasourceMutationResult(in, out, resp, err)
}

func datasourceCreateRequest(ctx context.Context, client *runaiclient.Client, in manageDatasourceInput) (map[string]any, error) {
	if in.ScopeName == "" {
		return in.Request, nil
	}
	meta := maps.Clone(in.Request["meta"].(map[string]any))
	scope, _ := meta["scope"].(string)
	filters, err := datasourceScopeFilters(ctx, client, manageDatasourceInput{Scope: assets.Scope(scope), ScopeName: in.ScopeName})
	if err != nil {
		return nil, err
	}
	switch filters.Scope {
	case assets.SCOPE_PROJECT:
		meta["projectId"] = *filters.ProjectID
	case assets.SCOPE_DEPARTMENT:
		meta["departmentId"] = filters.DepartmentID
	case assets.SCOPE_CLUSTER:
		meta["clusterId"] = filters.ClusterID
	}
	request := maps.Clone(in.Request)
	request["meta"] = meta
	return request, nil
}

func updateDatasource(ctx context.Context, client *runaiclient.Client, in manageDatasourceInput) (*manageDatasourceOutput, error) {
	out := newManageDatasourceOutput(in)
	current, err := getDatasource(ctx, client, in.Kind, in.DatasourceID)
	if err != nil {
		return out, err
	}
	if in.Name != "" && current.name != in.Name {
		return out, fmt.Errorf("datasource %q changed name after lookup; inspect it again: %w", in.DatasourceID, runaimcp.ErrConflict)
	}
	out = datasourceMutationOutput(in.Action, current)
	current = prepareDatasourceUpdate(current)
	if err := validateMutableDatasource(current, in); err != nil {
		return out, err
	}
	diff := common.DiffMaps(flattenDatasourceFields(current.updateRequest), flattenDatasourceFields(in.Request))
	if err := common.ConfirmWrite(ctx, in.Confirm, string(common.ActionUpdate), "datasource", current.name, current.id, diff); err != nil {
		return out, err
	}

	if runaimcp.ClientCanElicitForm(runaimcp.SessionFromContext(ctx)) {
		current, err = getDatasource(ctx, client, in.Kind, in.DatasourceID)
		if err != nil {
			return out, err
		}
		out = datasourceMutationOutput(in.Action, current)
		current = prepareDatasourceUpdate(current)
		if err := validateMutableDatasource(current, in); err != nil {
			return out, err
		}
	}

	request := datasourceUpdatePayload(in.Kind, in.Request, current)
	resp, err := datasourceAdapters[in.Kind].update(ctx, client, in.DatasourceID, request)

	return datasourceMutationResult(in, out, resp, err)
}

func deleteDatasource(ctx context.Context, client *runaiclient.Client, in manageDatasourceInput) (*manageDatasourceOutput, error) {
	out := newManageDatasourceOutput(in)
	current, err := getDatasource(ctx, client, in.Kind, in.DatasourceID)
	if err != nil {
		if in.IgnoreMissing && common.IsNotFound(err) {
			return ignoredMissingDatasourceOutput(in), nil
		}
		return out, err
	}
	if in.Name != "" && current.name != in.Name {
		return out, fmt.Errorf("datasource %q changed name after lookup; inspect it again: %w", in.DatasourceID, runaimcp.ErrConflict)
	}
	out = datasourceMutationOutput(in.Action, current)
	if current.scope == assets.SCOPE_SYSTEM {
		return out, fmt.Errorf("system-scoped datasource %q is read-only: %w", in.DatasourceID, runaimcp.ErrForbidden)
	}
	if err := confirmDatasourceDelete(ctx, client, in, current); err != nil {
		if in.IgnoreMissing && common.IsNotFound(err) {
			return ignoredMissingDatasourceOutput(in), nil
		}
		return out, err
	}
	resp, err := datasourceAdapters[in.Kind].delete(ctx, client, in.DatasourceID)
	if err != nil && !datasourceMutationAccepted(resp) {
		mapped := runaimcp.MapHTTPError(resp, err)
		if in.IgnoreMissing && common.IsNotFound(mapped) {
			return ignoredMissingDatasourceOutput(in), nil
		}
		return out, fmt.Errorf("delete datasource %q: %w", in.DatasourceID, mapped)
	}
	out.Datasource = nil
	out.Applied = true
	return out, nil
}

func getDatasource(ctx context.Context, client *runaiclient.Client, kind assets.AssetKind, id string) (datasourceSnapshot, error) {
	resp, err := datasourceAdapters[kind].get(ctx, client, id)
	if err != nil {
		return datasourceSnapshot{}, fmt.Errorf("get datasource %q: %w", id, runaimcp.MapHTTPError(resp, err))
	}
	snapshot, err := readDatasourceSnapshot(resp, fmt.Sprintf("get datasource %q", id), kind)
	if err != nil {
		return datasourceSnapshot{}, err
	}
	if !strings.EqualFold(snapshot.id, id) {
		return datasourceSnapshot{}, fmt.Errorf("get datasource %q returned id %q: %w", id, snapshot.id, runaimcp.ErrUpstream)
	}
	return snapshot, nil
}

func validateMutableDatasource(snapshot datasourceSnapshot, in manageDatasourceInput) error {
	if snapshot.scope == assets.SCOPE_SYSTEM {
		return fmt.Errorf("system-scoped datasource %q is read-only: %w", in.DatasourceID, runaimcp.ErrForbidden)
	}
	if snapshot.updateBlockedReason != "" {
		return fmt.Errorf("datasource %q cannot be safely updated: %s: %w",
			in.DatasourceID, snapshot.updateBlockedReason, runaimcp.ErrConflict)
	}
	if *snapshot.updateCount != *in.ExpectedUpdateCount {
		return fmt.Errorf("datasource %q changed since list_datasources; list it again before updating: %w",
			in.DatasourceID, runaimcp.ErrConflict)
	}
	return nil
}

func newManageDatasourceOutput(in manageDatasourceInput) *manageDatasourceOutput {
	return &manageDatasourceOutput{Kind: in.Kind, Action: in.Action, DatasourceID: in.DatasourceID, Name: in.Name}
}

func datasourceMutationOutput(action common.WriteAction, snapshot datasourceSnapshot) *manageDatasourceOutput {
	return &manageDatasourceOutput{
		Kind:         snapshot.kind,
		Action:       action,
		DatasourceID: snapshot.id,
		Name:         snapshot.name,
		Datasource:   snapshot.datasource,
	}
}

func ignoredMissingDatasourceOutput(in manageDatasourceInput) *manageDatasourceOutput {
	out := newManageDatasourceOutput(in)
	target := in.DatasourceID
	if target == "" {
		target = in.Name
	}
	out.Warnings = []string{fmt.Sprintf("datasource %q not found; nothing to delete", target)}
	return out
}

func resolveDatasourceID(ctx context.Context, client *runaiclient.Client, in manageDatasourceInput) (string, error) {
	filters, err := datasourceScopeFilters(ctx, client, in)
	if err != nil {
		return "", err
	}
	filters.Name = in.Name
	for queryType, entry := range datasourceTypes {
		if entry.kind == in.Kind {
			filters.Type = queryType
			break
		}
	}
	entries, err := fetchDatasources(ctx, client, filters)
	if err != nil {
		return "", err
	}
	var matches []string
	for _, entry := range entries {
		meta, _ := entry.(map[string]any)["meta"].(map[string]any)
		if meta["name"] != in.Name || meta["kind"] != string(in.Kind) {
			continue
		}
		// Scope queries may include ancestors; a mutation must select the asset's own scope.
		if in.Scope != "" && meta["scope"] != string(in.Scope) {
			continue
		}
		if filters.ProjectID != nil && fmt.Sprint(meta["projectId"]) != fmt.Sprint(*filters.ProjectID) {
			continue
		}
		if filters.DepartmentID != "" && meta["departmentId"] != filters.DepartmentID {
			continue
		}
		clusterID, _ := meta["clusterId"].(string)
		if filters.ClusterID != "" && !strings.EqualFold(clusterID, filters.ClusterID) {
			continue
		}
		snapshot, err := datasourceIdentity(meta, in.Kind)
		if err != nil {
			return "", fmt.Errorf("resolve datasource %q: %v: %w", in.Name, err, runaimcp.ErrUpstream)
		}
		matches = append(matches, snapshot.id)
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("datasource %q of kind %q: %w", in.Name, in.Kind, runaimcp.ErrNotFound)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("ambiguous_datasource: name %q matches multiple %s assets (%s); specify scope with scopeId/scopeName or use datasourceId: %w",
			in.Name, in.Kind, strings.Join(matches, ", "), runaimcp.ErrValidation)
	}
}

func datasourceScopeFilters(ctx context.Context, client *runaiclient.Client, in manageDatasourceInput) (listDatasourcesInput, error) {
	filters := listDatasourcesInput{Scope: in.Scope}
	id := in.ScopeID
	if in.ScopeName != "" {
		var err error
		if in.Scope == assets.SCOPE_CLUSTER {
			id, err = common.ResolveClusterID(ctx, client, in.ScopeName)
		} else {
			id, err = common.ResolveOrgUnitID(ctx, client, string(in.Scope), in.ScopeName, "")
		}
		if err != nil {
			return filters, err
		}
		if strings.TrimSpace(id) == "" {
			return filters, fmt.Errorf("resolve %s %q: missing scope ID: %w", in.Scope, in.ScopeName, runaimcp.ErrUpstream)
		}
	}
	if id == "" {
		return filters, nil
	}
	switch in.Scope {
	case assets.SCOPE_PROJECT:
		if err := common.ValidateNumericString("projectId", id); err != nil {
			return filters, fmt.Errorf("%v: %w", err, runaimcp.ErrValidation)
		}
		projectID, err := strconv.ParseInt(id, 10, 32)
		if err != nil {
			return filters, fmt.Errorf("convert projectId to the API integer type: %v: %w", err, runaimcp.ErrValidation)
		}
		value := int32(projectID)
		filters.ProjectID = &value
	case assets.SCOPE_DEPARTMENT:
		filters.DepartmentID = id
	case assets.SCOPE_CLUSTER:
		filters.ClusterID = id
	}
	return filters, nil
}
