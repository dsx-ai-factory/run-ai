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

type manageEnvironmentInput struct {
	Action              string         `json:"action"                          jsonschema:"What to do. One of: create, update, delete."`
	EnvironmentID       string         `json:"environmentId,omitempty"         jsonschema:"Environment asset UUID. Required for update and delete; use list_environments to resolve it. Must not be set for create."`
	Request             map[string]any `json:"request,omitempty"               jsonschema:"Complete Assets API request for create or update. Start updates from list_environments.updateRequest and edit only intended fields."` //nolint:lll
	ExpectedUpdateCount *int64         `json:"expectedUpdateCount,omitempty"   jsonschema:"Required for update. Pass unchanged from list_environments for a best-effort stale-write check."`
	IgnoreMissing       bool           `json:"ignoreMissing,omitempty"         jsonschema:"Delete-only. Treat an already-missing environment as a successful no-op."`
	Confirm             bool           `json:"confirm,omitempty"               jsonschema:"Required for update and delete without interactive confirmation."`
}

func (in manageEnvironmentInput) Validate() error {
	action := common.WriteAction(in.Action)
	if !action.Valid() {
		return fmt.Errorf("action must be one of: create, update, delete")
	}
	if in.EnvironmentID != "" {
		if err := common.ValidateUUID("environmentId", in.EnvironmentID); err != nil {
			return err
		}
	}
	if in.IgnoreMissing && action != common.ActionDelete {
		return fmt.Errorf("ignoreMissing is only valid with action=delete")
	}

	switch action {
	case common.ActionCreate:
		if in.EnvironmentID != "" {
			return fmt.Errorf("environmentId must not be set for create")
		}
		if in.ExpectedUpdateCount != nil {
			return fmt.Errorf("expectedUpdateCount must not be set for create")
		}
		_, err := decodeEnvironmentCreateRequest(in.Request)
		return err
	case common.ActionUpdate:
		if in.EnvironmentID == "" {
			return fmt.Errorf("environmentId is required for update")
		}
		if in.ExpectedUpdateCount == nil {
			return fmt.Errorf("expectedUpdateCount is required for update; call list_environments")
		}
		if *in.ExpectedUpdateCount < 0 {
			return fmt.Errorf("expectedUpdateCount must be non-negative")
		}
		_, err := decodeEnvironmentUpdateRequest(in.Request)
		return err
	case common.ActionDelete:
		if in.EnvironmentID == "" {
			return fmt.Errorf("environmentId is required for delete")
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

type manageEnvironmentOutput struct {
	Action        string         `json:"action"                  jsonschema:"The action performed: create, update, or delete."`
	Applied       bool           `json:"applied"                 jsonschema:"True when the upstream service accepted the change. False for an ignored-missing delete."`
	EnvironmentID string         `json:"environmentId,omitempty" jsonschema:"Created or targeted environment UUID."`
	Name          string         `json:"name,omitempty"          jsonschema:"Environment name."`
	Environment   map[string]any `json:"environment,omitempty"   jsonschema:"Environment returned by the Assets API after create or update."`
	Warnings      []string       `json:"warnings,omitempty"      jsonschema:"Non-fatal notes, including ignored missing deletes."`
}

func registerManageEnvironment(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[manageEnvironmentInput, manageEnvironmentOutput]{
		Name: "manage_environment",
		Description: "Create, fully replace, or delete one Run:ai environment asset through the Assets API. Create requires " +
			"meta.name, meta.scope, meta.workloadSupportedTypes, and spec. Update requires the complete update request returned by " +
			"list_environments plus its expectedUpdateCount. Update and delete require confirmation. System-scoped environments " +
			"are read-only. Do not use this tool to list environments.",
		ReadOnly:    false,
		Destructive: true,
		Handler: func(ctx context.Context, in manageEnvironmentInput) (manageEnvironmentOutput, error) {
			return manageEnvironmentRun(ctx, client, in)
		},
	})
}

func manageEnvironmentRun(ctx context.Context, client *runaiclient.Client, in manageEnvironmentInput) (manageEnvironmentOutput, error) {
	switch common.WriteAction(in.Action) {
	case common.ActionCreate:
		return createEnvironment(ctx, client, in)
	case common.ActionUpdate:
		return updateEnvironment(ctx, client, in)
	case common.ActionDelete:
		return deleteEnvironment(ctx, client, in)
	default:
		return manageEnvironmentOutput{}, fmt.Errorf("unsupported action %q: %w", in.Action, runaimcp.ErrValidation)
	}
}

func createEnvironment(ctx context.Context, client *runaiclient.Client, in manageEnvironmentInput) (manageEnvironmentOutput, error) {
	request, err := decodeEnvironmentCreateRequest(in.Request)
	if err != nil {
		return manageEnvironmentOutput{}, fmt.Errorf("create environment request: %w", err)
	}
	created, resp, err := client.Assets.EnvironmentAPI.CreateEnvironmentAsset(ctx).EnvironmentCreationRequest(request).Execute()
	if err != nil {
		return manageEnvironmentOutput{}, fmt.Errorf("create environment %q: %w", request.Meta.Name, runaimcp.MapHTTPError(resp, err))
	}
	raw, err := readEnvironmentResponse(resp, fmt.Sprintf("create environment %q", request.Meta.Name))
	if err != nil {
		return manageEnvironmentOutput{}, err
	}
	snapshot, err := newEnvironmentSnapshot(created, raw)
	if err != nil {
		return manageEnvironmentOutput{}, fmt.Errorf("create environment %q: %w", request.Meta.Name, err)
	}
	if snapshot.name != request.Meta.Name || snapshot.scope != request.Meta.Scope {
		return manageEnvironmentOutput{}, fmt.Errorf("create environment %q returned name %q and scope %q: %w",
			request.Meta.Name, snapshot.name, snapshot.scope, runaimcp.ErrUpstream)
	}
	out := environmentMutationOutput(in.Action, snapshot)
	out.Applied = true
	return out, nil
}

func updateEnvironment(ctx context.Context, client *runaiclient.Client, in manageEnvironmentInput) (manageEnvironmentOutput, error) {
	request, err := decodeEnvironmentUpdateRequest(in.Request)
	if err != nil {
		return manageEnvironmentOutput{}, fmt.Errorf("update environment request: %w", err)
	}
	current, err := getEnvironment(ctx, client, in.EnvironmentID)
	if err != nil {
		return manageEnvironmentOutput{}, err
	}
	if err := validateMutableEnvironment(current, in); err != nil {
		return environmentMutationOutput(in.Action, current), err
	}
	diff := common.DiffMaps(current.updateRequest, in.Request)
	if err := common.ConfirmWrite(ctx, in.Confirm, "update", "environment", current.name, current.id, diff); err != nil {
		return environmentMutationOutput(in.Action, current), err
	}

	current, err = getEnvironment(ctx, client, in.EnvironmentID)
	if err != nil {
		return manageEnvironmentOutput{}, err
	}
	if err := validateMutableEnvironment(current, in); err != nil {
		return environmentMutationOutput(in.Action, current), err
	}

	updated, resp, err := client.Assets.EnvironmentAPI.UpdateEnvironmentAssetById(ctx, in.EnvironmentID).
		EnvironmentUpdateRequest(request).Execute()
	if err != nil {
		return manageEnvironmentOutput{}, fmt.Errorf("update environment %q: %w", in.EnvironmentID, runaimcp.MapHTTPError(resp, err))
	}
	raw, err := readEnvironmentResponse(resp, fmt.Sprintf("update environment %q", in.EnvironmentID))
	if err != nil {
		return manageEnvironmentOutput{}, err
	}
	snapshot, err := newEnvironmentSnapshot(updated, raw)
	if err != nil {
		return manageEnvironmentOutput{}, fmt.Errorf("update environment %q: %w", in.EnvironmentID, err)
	}
	if !strings.EqualFold(snapshot.id, in.EnvironmentID) {
		return manageEnvironmentOutput{}, fmt.Errorf("update environment %q returned id %q: %w",
			in.EnvironmentID, snapshot.id, runaimcp.ErrUpstream)
	}
	out := environmentMutationOutput(in.Action, snapshot)
	out.Applied = true
	return out, nil
}

func deleteEnvironment(ctx context.Context, client *runaiclient.Client, in manageEnvironmentInput) (manageEnvironmentOutput, error) {
	current, err := getEnvironment(ctx, client, in.EnvironmentID)
	if err != nil {
		if in.IgnoreMissing && common.IsNotFound(err) {
			return ignoredMissingEnvironmentOutput(in), nil
		}
		return manageEnvironmentOutput{}, err
	}
	out := environmentMutationOutput(in.Action, current)
	if current.scope == assets.SCOPE_SYSTEM {
		return out, fmt.Errorf("system-scoped environment %q is read-only: %w", in.EnvironmentID, runaimcp.ErrForbidden)
	}
	details := common.WithConfirmWarning(current.environment, "deleting an environment is permanent and may fail while it is in use")
	if err := common.ConfirmWrite(ctx, in.Confirm, "delete", "environment", current.name, current.id, details); err != nil {
		return out, err
	}
	_, resp, err := client.Assets.EnvironmentAPI.DeleteEnvironmentAssetById(ctx, in.EnvironmentID).Execute()
	if err != nil {
		mapped := runaimcp.MapHTTPError(resp, err)
		if in.IgnoreMissing && common.IsNotFound(mapped) {
			return ignoredMissingEnvironmentOutput(in), nil
		}
		return manageEnvironmentOutput{}, fmt.Errorf("delete environment %q: %w", in.EnvironmentID, mapped)
	}
	out.Environment = nil
	out.Applied = true
	return out, nil
}

func getEnvironment(ctx context.Context, client *runaiclient.Client, id string) (environmentSnapshot, error) {
	environment, resp, err := client.Assets.EnvironmentAPI.GetEnvironmentAssetById(ctx, id).Execute()
	if err != nil {
		return environmentSnapshot{}, fmt.Errorf("get environment %q: %w", id, runaimcp.MapHTTPError(resp, err))
	}
	raw, err := readEnvironmentResponse(resp, fmt.Sprintf("get environment %q", id))
	if err != nil {
		return environmentSnapshot{}, err
	}
	snapshot, err := newEnvironmentSnapshot(environment, raw)
	if err != nil {
		return environmentSnapshot{}, fmt.Errorf("get environment %q: %w", id, err)
	}
	if !strings.EqualFold(snapshot.id, id) {
		return environmentSnapshot{}, fmt.Errorf("get environment %q returned id %q: %w", id, snapshot.id, runaimcp.ErrUpstream)
	}
	return snapshot, nil
}

func validateMutableEnvironment(snapshot environmentSnapshot, in manageEnvironmentInput) error {
	if snapshot.scope == assets.SCOPE_SYSTEM {
		return fmt.Errorf("system-scoped environment %q is read-only: %w", in.EnvironmentID, runaimcp.ErrForbidden)
	}
	if snapshot.updateCount == nil {
		return fmt.Errorf("environment %q has no API update count: %w", in.EnvironmentID, runaimcp.ErrConflict)
	}
	if snapshot.updateBlockedReason != "" {
		return fmt.Errorf("environment %q cannot be safely updated: %s: %w",
			in.EnvironmentID, snapshot.updateBlockedReason, runaimcp.ErrConflict)
	}
	if *snapshot.updateCount != *in.ExpectedUpdateCount {
		return fmt.Errorf("environment %q changed since list_environments; list it again before updating: %w",
			in.EnvironmentID, runaimcp.ErrConflict)
	}
	return nil
}

func decodeEnvironmentCreateRequest(raw map[string]any) (assets.EnvironmentCreationRequest, error) {
	var request assets.EnvironmentCreationRequest
	if err := validateEnvironmentRequest(raw, "create"); err != nil {
		return request, err
	}
	meta, _ := raw["meta"].(map[string]any)
	if _, ok := nonNilEnvironmentField(meta, "workloadSupportedTypes"); !ok {
		return request, fmt.Errorf("request.meta.workloadSupportedTypes is required")
	}
	if err := validateEnvironmentCreateScope(raw); err != nil {
		return request, err
	}
	if err := common.DecodeTypedRequest(raw, &request); err != nil {
		return request, fmt.Errorf("request must match the environment create API: %w", err)
	}
	return request, nil
}

func decodeEnvironmentUpdateRequest(raw map[string]any) (assets.EnvironmentUpdateRequest, error) {
	var request assets.EnvironmentUpdateRequest
	if err := decodeEnvironmentRequest(raw, "update", &request); err != nil {
		return request, err
	}
	return request, nil
}

func decodeEnvironmentRequest(raw map[string]any, action string, request any) error {
	if err := validateEnvironmentRequest(raw, action); err != nil {
		return err
	}
	if err := common.DecodeTypedRequest(raw, request); err != nil {
		return fmt.Errorf("request must match the environment %s API: %w", action, err)
	}
	return nil
}

func validateEnvironmentRequest(raw map[string]any, action string) error {
	if len(raw) == 0 {
		return fmt.Errorf("request is required for %s", action)
	}
	meta, ok := raw["meta"].(map[string]any)
	if !ok || meta == nil {
		return fmt.Errorf("request.meta is required and must be an object")
	}
	if _, err := common.RequiredString("request.meta.name", meta["name"]); err != nil {
		return err
	}
	if description, exists := meta["description"]; exists && description != nil {
		value, ok := description.(string)
		if !ok || len(value) == 0 || len(value) > 250 {
			return fmt.Errorf("request.meta.description must contain between 1 and 250 UTF-8 bytes or be null")
		}
	}
	if action == "create" {
		if _, err := common.RequiredString("request.meta.scope", meta["scope"]); err != nil {
			return err
		}
	}
	if spec, ok := raw["spec"].(map[string]any); !ok || spec == nil {
		return fmt.Errorf("request.spec is required and must be an object")
	}
	return nil
}

func validateEnvironmentCreateScope(raw map[string]any) error {
	meta, _ := raw["meta"].(map[string]any)
	if private, exists := meta["private"]; exists && private == true {
		return fmt.Errorf("request.meta.private=true is not supported for environments")
	}
	scope, err := common.RequiredString("request.meta.scope", meta["scope"])
	if err != nil {
		return err
	}
	clusterID, clusterSet := nonNilEnvironmentField(meta, "clusterId")
	departmentID, departmentSet := nonNilEnvironmentField(meta, "departmentId")
	projectID, projectSet := nonNilEnvironmentField(meta, "projectId")
	switch assets.Scope(scope) {
	case assets.SCOPE_TENANT:
		if clusterSet || departmentSet || projectSet {
			return fmt.Errorf("tenant-scoped environments must not set clusterId, departmentId, or projectId")
		}
	case assets.SCOPE_CLUSTER:
		cluster, _ := clusterID.(string)
		if err := common.ValidateUUID("request.meta.clusterId", cluster); err != nil {
			return err
		}
		if departmentSet || projectSet {
			return fmt.Errorf("cluster-scoped environments must not set departmentId or projectId")
		}
	case assets.SCOPE_DEPARTMENT:
		department, _ := departmentID.(string)
		if _, err := common.RequiredString("request.meta.departmentId", department); err != nil {
			return err
		}
		if clusterSet || projectSet {
			return fmt.Errorf("department-scoped environments must not set clusterId or projectId")
		}
	case assets.SCOPE_PROJECT:
		if !validEnvironmentProjectID(projectID) {
			return fmt.Errorf("request.meta.projectId must be a positive 32-bit integer")
		}
		if clusterSet || departmentSet {
			return fmt.Errorf("project-scoped environments must not set clusterId or departmentId")
		}
	default:
		return fmt.Errorf("request.meta.scope must be one of: tenant, cluster, department, project")
	}
	return nil
}

func nonNilEnvironmentField(value map[string]any, field string) (any, bool) {
	fieldValue, exists := value[field]
	return fieldValue, exists && fieldValue != nil
}

func validEnvironmentProjectID(value any) bool {
	const maxProjectID = int64(1<<31 - 1)

	switch value := value.(type) {
	case int:
		return value > 0 && int64(value) <= maxProjectID
	case int32:
		return value > 0
	case int64:
		return value > 0 && value <= maxProjectID
	case float64:
		return value > 0 && value <= float64(maxProjectID) && value == float64(int64(value))
	default:
		return false
	}
}

func environmentMutationOutput(action string, snapshot environmentSnapshot) manageEnvironmentOutput {
	return manageEnvironmentOutput{
		Action:        action,
		EnvironmentID: snapshot.id,
		Name:          snapshot.name,
		Environment:   snapshot.environment,
	}
}

func ignoredMissingEnvironmentOutput(in manageEnvironmentInput) manageEnvironmentOutput {
	return manageEnvironmentOutput{
		Action:        in.Action,
		EnvironmentID: in.EnvironmentID,
		Warnings:      []string{fmt.Sprintf("environment %q not found; nothing to delete", in.EnvironmentID)},
	}
}
