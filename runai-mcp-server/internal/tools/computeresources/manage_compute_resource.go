/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package computeresources

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	assets "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

type manageComputeResourceInput struct {
	Action              string         `json:"action"                        jsonschema:"What to do. One of: create, update, delete."`
	ComputeResourceID   string         `json:"computeResourceId,omitempty"   jsonschema:"Compute resource UUID. Required for update and delete; use list_compute_resources to resolve it. Must not be set for create."`
	Request             map[string]any `json:"request,omitempty"             jsonschema:"Complete API request for create or update. Create requires meta.name, meta.scope, and spec; spec may be null for no compute settings. Update is a full replacement requiring meta.name and an object spec; start from list_compute_resources.updateRequest and edit only intended fields."` //nolint:lll // complete full-replacement request description
	ExpectedFingerprint string         `json:"expectedFingerprint,omitempty" jsonschema:"Required for update. Pass unchanged from list_compute_resources for a best-effort stale-snapshot check. The API has no atomic conditional-update support."`                                                                                                                                 //nolint:lll // long tool-safety description; cannot wrap a struct tag
	IgnoreMissing       bool           `json:"ignoreMissing,omitempty"       jsonschema:"Delete-only. Treat an already-missing compute resource as a successful no-op."`
	Confirm             bool           `json:"confirm,omitempty"             jsonschema:"Required for update and delete when the client cannot show an interactive confirmation prompt. Create does not require confirmation."`
}

func (in manageComputeResourceInput) Validate() error {
	action := common.WriteAction(in.Action)
	if !action.Valid() {
		return fmt.Errorf("action must be one of: create, update, delete")
	}
	if in.ComputeResourceID != "" {
		if err := common.ValidateUUID("computeResourceId", in.ComputeResourceID); err != nil {
			return err
		}
	}
	if in.IgnoreMissing && action != common.ActionDelete {
		return fmt.Errorf("ignoreMissing is only valid with action=delete")
	}

	switch action {
	case common.ActionCreate:
		if in.ComputeResourceID != "" {
			return fmt.Errorf("computeResourceId must not be set for create")
		}
		if in.ExpectedFingerprint != "" {
			return fmt.Errorf("expectedFingerprint must not be set for create")
		}
		if len(in.Request) == 0 {
			return fmt.Errorf("request is required for create")
		}
		return validateComputeResourceCreateRequest(in.Request)
	case common.ActionUpdate:
		if in.ComputeResourceID == "" {
			return fmt.Errorf("computeResourceId is required for update")
		}
		if len(in.Request) == 0 {
			return fmt.Errorf("request is required for update")
		}
		if in.ExpectedFingerprint == "" {
			return fmt.Errorf("expectedFingerprint is required for the best-effort stale-snapshot check; call list_compute_resources")
		}
		if err := validateComputeResourceFingerprint(in.ExpectedFingerprint); err != nil {
			return err
		}
		return validateComputeResourceUpdateRequest(in.Request)
	case common.ActionDelete:
		if in.ComputeResourceID == "" {
			return fmt.Errorf("computeResourceId is required for delete")
		}
		if len(in.Request) != 0 {
			return fmt.Errorf("request must not be set for delete")
		}
		if in.ExpectedFingerprint != "" {
			return fmt.Errorf("expectedFingerprint must not be set for delete")
		}
		return nil
	default:
		return fmt.Errorf("unsupported action %q", in.Action)
	}
}

type manageComputeResourceOutput struct {
	Action            string         `json:"action"                      jsonschema:"The action performed: create, update, or delete."`
	Applied           bool           `json:"applied"                     jsonschema:"True when the upstream service accepted the change. False for an ignored-missing delete."`
	ComputeResourceID string         `json:"computeResourceId,omitempty" jsonschema:"Created or targeted compute resource UUID."`
	Name              string         `json:"name,omitempty"              jsonschema:"Compute resource name."`
	Before            map[string]any `json:"before"                      jsonschema:"Complete compute resource snapshot before the change. Empty object on create."`
	After             map[string]any `json:"after"                       jsonschema:"Complete compute resource snapshot after the change. Empty object on delete."`
	Diff              map[string]any `json:"diff"                        jsonschema:"Field-level changes as path -> {before, after}."`
	Warnings          []string       `json:"warnings,omitempty"          jsonschema:"Non-fatal notes, including ignored missing deletes."`
}

func registerManageComputeResource(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[manageComputeResourceInput, manageComputeResourceOutput]{
		Name: "manage_compute_resource",
		Description: "Create, fully replace, or delete one Run:ai compute resource asset. Create accepts the Assets API's " +
			"meta and spec fields at tenant, cluster, department, or project scope and does not require confirmation. " +
			"Update uses a full-replacement PUT: call list_compute_resources first, modify only the intended fields in its " +
			"updateRequest, and pass that complete request together with expectedFingerprint unchanged. The fingerprint provides " +
			"a best-effort stale-snapshot check; the Assets API has no atomic update precondition. Update and delete require " +
			"confirmation. Delete permanently removes the compute resource and may fail while workloads or templates use it. " +
			"System-scoped presets are read-only. Do not use this tool to list compute resources.",
		ReadOnly:    false,
		Destructive: true,
		Handler: func(ctx context.Context, in manageComputeResourceInput) (manageComputeResourceOutput, error) {
			return manageComputeResourceRun(ctx, client, in)
		},
	})
}

func manageComputeResourceRun(ctx context.Context, client *runaiclient.Client, in manageComputeResourceInput) (manageComputeResourceOutput, error) {
	out := newManageComputeResourceOutput(in)
	action := common.WriteAction(in.Action)
	switch action {
	case common.ActionCreate:
		return createComputeResource(ctx, client, in, out)
	case common.ActionUpdate, common.ActionDelete:
		current, err := getComputeResource(ctx, client, in.ComputeResourceID, action == common.ActionUpdate)
		if err != nil {
			if action == common.ActionDelete && in.IgnoreMissing && common.IsNotFound(err) {
				out.Warnings = append(out.Warnings, ignoredMissingComputeResourceWarning(in.ComputeResourceID))
				return out, nil
			}
			return manageComputeResourceOutput{}, err
		}
		out.Before = current.raw
		out.Name = current.name
		if current.scope == string(assets.SCOPE_SYSTEM) {
			return out, fmt.Errorf("system-scoped compute resource %q is read-only: %w",
				in.ComputeResourceID, runaimcp.ErrForbidden)
		}
		if action == common.ActionUpdate {
			return updateComputeResource(ctx, client, in, out, current)
		}
		return deleteComputeResource(ctx, client, in, out, current)
	default:
		return manageComputeResourceOutput{}, fmt.Errorf("unsupported action %q: %w", in.Action, runaimcp.ErrValidation)
	}
}

type computeResourceMutationTarget struct {
	name          string
	scope         string
	raw           map[string]any
	updateRequest map[string]any
	fingerprint   string
}

func getComputeResource(
	ctx context.Context,
	client *runaiclient.Client,
	id string,
	prepareUpdate bool,
) (computeResourceMutationTarget, error) {
	resource, resp, err := client.Assets.ComputeAPI.GetComputeAssetById(ctx, id).Execute()
	if err != nil {
		return computeResourceMutationTarget{}, fmt.Errorf("get compute resource %q: %w", id, runaimcp.MapHTTPError(resp, err))
	}
	if resource == nil {
		return computeResourceMutationTarget{}, fmt.Errorf("get compute resource %q: nil response: %w", id, runaimcp.ErrUpstream)
	}
	raw, err := readComputeResourceResponse(resp, fmt.Sprintf("get compute resource %q", id))
	if err != nil {
		return computeResourceMutationTarget{}, err
	}
	actualID, name, scope, err := computeResourceIdentity(raw)
	if err != nil {
		return computeResourceMutationTarget{}, fmt.Errorf("get compute resource %q: %w", id, err)
	}
	if !strings.EqualFold(actualID, id) {
		return computeResourceMutationTarget{}, fmt.Errorf("get compute resource %q returned id %q: %w", id, actualID, runaimcp.ErrUpstream)
	}
	target := computeResourceMutationTarget{name: name, scope: scope, raw: raw}
	if !prepareUpdate || scope == string(assets.SCOPE_SYSTEM) {
		return target, nil
	}
	updateRequest, fingerprint, err := computeResourceUpdateSnapshot(raw)
	if err != nil {
		return computeResourceMutationTarget{}, fmt.Errorf("prepare compute resource %q for update: %w", id, err)
	}
	target.updateRequest = updateRequest
	target.fingerprint = fingerprint
	return target, nil
}

func createComputeResource(ctx context.Context, client *runaiclient.Client, in manageComputeResourceInput, out manageComputeResourceOutput) (manageComputeResourceOutput, error) {
	var request assets.ComputeCreationRequest
	if err := common.DecodeTypedRequest(in.Request, &request); err != nil {
		return manageComputeResourceOutput{}, fmt.Errorf("decode compute resource create request: %w: %w", err, runaimcp.ErrValidation)
	}
	created, resp, err := client.Assets.ComputeAPI.CreateComputeAsset(ctx).ComputeCreationRequest(request).Execute()
	if err != nil {
		return manageComputeResourceOutput{}, fmt.Errorf("create compute resource %q: %w", request.Meta.Name, runaimcp.MapHTTPError(resp, err))
	}
	if created == nil {
		return manageComputeResourceOutput{}, fmt.Errorf("create compute resource %q: nil response: %w", request.Meta.Name, runaimcp.ErrUpstream)
	}
	raw, err := readComputeResourceResponse(resp, fmt.Sprintf("create compute resource %q", request.Meta.Name))
	if err != nil {
		return manageComputeResourceOutput{}, err
	}
	id, name, scope, err := computeResourceIdentity(raw)
	if err != nil {
		return manageComputeResourceOutput{}, err
	}
	if name != request.Meta.Name {
		return manageComputeResourceOutput{}, fmt.Errorf("create compute resource %q returned name %q: %w",
			request.Meta.Name, name, runaimcp.ErrUpstream)
	}
	if scope != string(request.Meta.Scope) {
		return manageComputeResourceOutput{}, fmt.Errorf("create compute resource %q returned scope %q instead of %q: %w",
			request.Meta.Name, scope, request.Meta.Scope, runaimcp.ErrUpstream)
	}
	if err := validateCreatedComputeResourceScope(in.Request, raw); err != nil {
		return manageComputeResourceOutput{}, fmt.Errorf("create compute resource %q: %w", request.Meta.Name, err)
	}
	out.ComputeResourceID = id
	out.Name = name
	out.After = raw
	out.Diff = common.DiffMaps(nil, out.After)
	out.Applied = true
	return out, nil
}

func updateComputeResource(ctx context.Context, client *runaiclient.Client, in manageComputeResourceInput, out manageComputeResourceOutput, current computeResourceMutationTarget) (manageComputeResourceOutput, error) {
	if current.fingerprint != in.ExpectedFingerprint {
		return out, fmt.Errorf("compute resource %q changed since list_compute_resources; list it again before updating: %w",
			in.ComputeResourceID, runaimcp.ErrConflict)
	}
	out.After = in.Request
	out.Diff = common.DiffMaps(current.updateRequest, in.Request)
	if err := common.ConfirmWrite(ctx, in.Confirm, "update", "compute resource", current.name, in.ComputeResourceID, out.Diff); err != nil {
		return out, err
	}

	confirmed, err := getComputeResource(ctx, client, in.ComputeResourceID, true)
	if err != nil {
		return out, err
	}
	if confirmed.fingerprint != in.ExpectedFingerprint {
		return out, fmt.Errorf("compute resource %q changed while awaiting confirmation; list it again before updating: %w",
			in.ComputeResourceID, runaimcp.ErrConflict)
	}

	var request assets.ComputeUpdateRequest
	if err := common.DecodeTypedRequest(in.Request, &request); err != nil {
		return manageComputeResourceOutput{}, fmt.Errorf("decode compute resource update request: %w: %w", err, runaimcp.ErrValidation)
	}
	updated, resp, err := client.Assets.ComputeAPI.UpdateComputeAssetById(ctx, in.ComputeResourceID).
		ComputeUpdateRequest(request).Execute()
	if err != nil {
		return manageComputeResourceOutput{}, fmt.Errorf("update compute resource %q: %w",
			in.ComputeResourceID, runaimcp.MapHTTPError(resp, err))
	}
	if updated == nil {
		return manageComputeResourceOutput{}, fmt.Errorf("update compute resource %q: nil response: %w",
			in.ComputeResourceID, runaimcp.ErrUpstream)
	}
	raw, err := readComputeResourceResponse(resp, fmt.Sprintf("update compute resource %q", in.ComputeResourceID))
	if err != nil {
		return manageComputeResourceOutput{}, err
	}
	updatedID, name, scope, err := computeResourceIdentity(raw)
	if err != nil {
		return manageComputeResourceOutput{}, err
	}
	if !strings.EqualFold(updatedID, in.ComputeResourceID) {
		return manageComputeResourceOutput{}, fmt.Errorf("update compute resource %q returned id %q: %w",
			in.ComputeResourceID, updatedID, runaimcp.ErrUpstream)
	}
	if name != request.Meta.Name {
		return manageComputeResourceOutput{}, fmt.Errorf("update compute resource %q returned name %q instead of %q: %w",
			in.ComputeResourceID, name, request.Meta.Name, runaimcp.ErrUpstream)
	}
	if scope != current.scope {
		return manageComputeResourceOutput{}, fmt.Errorf("update compute resource %q changed scope from %q to %q: %w",
			in.ComputeResourceID, current.scope, scope, runaimcp.ErrUpstream)
	}
	if err := validateUpdatedComputeResourceRequest(in.Request, raw); err != nil {
		return manageComputeResourceOutput{}, fmt.Errorf("update compute resource %q: %w", in.ComputeResourceID, err)
	}
	out.Name = name
	out.After = raw
	out.Diff = common.DiffMaps(out.Before, out.After)
	out.Applied = true
	return out, nil
}

func deleteComputeResource(ctx context.Context, client *runaiclient.Client, in manageComputeResourceInput, out manageComputeResourceOutput, current computeResourceMutationTarget) (manageComputeResourceOutput, error) {
	out.Diff = common.DiffMaps(out.Before, nil)
	details := common.WithConfirmWarning(out.Before,
		"deleting a compute resource is permanent and may fail while a workload, template, or other asset uses it")
	if err := common.ConfirmWrite(ctx, in.Confirm, "delete", "compute resource", current.name, in.ComputeResourceID, details); err != nil {
		return out, err
	}
	_, resp, err := client.Assets.ComputeAPI.DeleteComputeAssetById(ctx, in.ComputeResourceID).Execute()
	if err != nil {
		mapped := runaimcp.MapHTTPError(resp, err)
		if in.IgnoreMissing && common.IsNotFound(mapped) {
			out.Warnings = append(out.Warnings, ignoredMissingComputeResourceWarning(in.ComputeResourceID))
			return out, nil
		}
		return manageComputeResourceOutput{}, fmt.Errorf("delete compute resource %q: %w", in.ComputeResourceID, mapped)
	}
	out.Applied = true
	return out, nil
}

func ignoredMissingComputeResourceWarning(id string) string {
	return fmt.Sprintf("compute resource %q not found; nothing to delete", id)
}

func newManageComputeResourceOutput(in manageComputeResourceInput) manageComputeResourceOutput {
	return manageComputeResourceOutput{
		Action:            in.Action,
		ComputeResourceID: in.ComputeResourceID,
		Before:            map[string]any{},
		After:             map[string]any{},
		Diff:              map[string]any{},
	}
}

func validateComputeResourceCreateRequest(request map[string]any) error {
	if err := validateComputeResourceRequestShape(request, true); err != nil {
		return err
	}
	var typed assets.ComputeCreationRequest
	if err := common.DecodeTypedRequest(request, &typed); err != nil {
		return fmt.Errorf("request must match the compute resource create API: %w", err)
	}
	return nil
}

func validateComputeResourceUpdateRequest(request map[string]any) error {
	if err := validateComputeResourceRequestShape(request, false); err != nil {
		return err
	}
	var typed assets.ComputeUpdateRequest
	if err := common.DecodeTypedRequest(request, &typed); err != nil {
		return fmt.Errorf("request must match the compute resource update API: %w", err)
	}
	return nil
}

func validateComputeResourceRequestShape(request map[string]any, create bool) error {
	meta, ok := request["meta"].(map[string]any)
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
	if private, exists := meta["private"]; exists && private == true {
		return fmt.Errorf("request.meta.private=true is not supported for compute resources")
	}
	if workloadTypes, exists := meta["workloadSupportedTypes"]; exists && workloadTypes != nil {
		return fmt.Errorf("request.meta.workloadSupportedTypes is not supported for compute resources")
	}
	if _, exists := request["spec"]; !exists {
		return fmt.Errorf("request.spec is required; use an empty object for no compute settings")
	}
	if !create {
		if spec, ok := request["spec"].(map[string]any); !ok || spec == nil {
			return fmt.Errorf("request.spec must be an object for update; use an empty object to clear all compute settings")
		}
		for _, field := range []string{"scope", "clusterId", "departmentId", "projectId"} {
			if _, exists := meta[field]; exists {
				return fmt.Errorf("request.meta.%s is create-only; update cannot move a compute resource between scopes", field)
			}
		}
		return nil
	}

	scope, err := common.RequiredString("request.meta.scope", meta["scope"])
	if err != nil {
		return err
	}
	clusterIDValue, clusterIDSet := nonNilMapField(meta, "clusterId")
	clusterID, _ := clusterIDValue.(string)
	departmentIDValue, departmentIDSet := nonNilMapField(meta, "departmentId")
	departmentID, _ := departmentIDValue.(string)
	projectIDValue, projectIDSet := nonNilMapField(meta, "projectId")
	projectID, hasProjectID := numberAsPositiveInt32(projectIDValue)
	switch assets.Scope(scope) {
	case assets.SCOPE_TENANT:
		if clusterIDSet || departmentIDSet || projectIDSet {
			return fmt.Errorf("tenant-scoped compute resources must not set clusterId, departmentId, or projectId")
		}
	case assets.SCOPE_CLUSTER:
		if err := common.ValidateUUID("request.meta.clusterId", clusterID); err != nil {
			return err
		}
		if departmentIDSet || projectIDSet {
			return fmt.Errorf("cluster-scoped compute resources must not set departmentId or projectId")
		}
	case assets.SCOPE_DEPARTMENT:
		if _, err := common.RequiredString("request.meta.departmentId", departmentID); err != nil {
			return err
		}
		if clusterIDSet || projectIDSet {
			return fmt.Errorf("department-scoped compute resources must not set clusterId or projectId")
		}
	case assets.SCOPE_PROJECT:
		if !hasProjectID || projectID <= 0 {
			return fmt.Errorf("request.meta.projectId must be a positive 32-bit integer")
		}
		if departmentIDSet {
			return fmt.Errorf("project-scoped compute resources must not set departmentId")
		}
		if clusterIDSet {
			return fmt.Errorf("project-scoped compute resources must not set clusterId; the API ignores it and derives the cluster from projectId")
		}
	case assets.SCOPE_SYSTEM:
		return fmt.Errorf("request.meta.scope=system is not supported by manage_compute_resource")
	default:
		return fmt.Errorf("request.meta.scope must be one of: tenant, cluster, department, project")
	}
	return nil
}

func validateCreatedComputeResourceScope(request, response map[string]any) error {
	requestMeta, _ := request["meta"].(map[string]any)
	responseMeta, err := computeResourceMeta(response)
	if err != nil {
		return err
	}
	requestedDescription, requested := optionalComputeResourceDescription(requestMeta)
	responseDescription, returned := optionalComputeResourceDescription(responseMeta)
	if requested != returned || (requested && requestedDescription != responseDescription) {
		return fmt.Errorf("compute resource response description does not match the create request: %w", runaimcp.ErrUpstream)
	}
	scope, _ := requestMeta["scope"].(string)
	switch assets.Scope(scope) {
	case assets.SCOPE_CLUSTER:
		expected, _ := requestMeta["clusterId"].(string)
		actual, err := computeResourceMetaString(responseMeta, "clusterId")
		if err != nil {
			return err
		}
		if !strings.EqualFold(actual, expected) {
			return fmt.Errorf("compute resource response clusterId is %q, expected %q: %w", actual, expected, runaimcp.ErrUpstream)
		}
	case assets.SCOPE_DEPARTMENT:
		expected, _ := requestMeta["departmentId"].(string)
		actual, err := computeResourceMetaString(responseMeta, "departmentId")
		if err != nil {
			return err
		}
		if actual != expected {
			return fmt.Errorf("compute resource response departmentId is %q, expected %q: %w", actual, expected, runaimcp.ErrUpstream)
		}
	case assets.SCOPE_PROJECT:
		expected, _ := numberAsPositiveInt32(requestMeta["projectId"])
		actual, ok := numberAsPositiveInt32(responseMeta["projectId"])
		if !ok || actual != expected {
			return fmt.Errorf("compute resource response projectId is invalid or does not match %d: %w", expected, runaimcp.ErrUpstream)
		}
	}
	return nil
}

func validateUpdatedComputeResourceRequest(request, response map[string]any) error {
	responseRequest, _, err := computeResourceUpdateSnapshot(response)
	if err != nil {
		return err
	}
	requestSpecFingerprint, err := computeResourceJSONFingerprint(request["spec"])
	if err != nil {
		return err
	}
	responseSpecFingerprint, err := computeResourceJSONFingerprint(responseRequest["spec"])
	if err != nil {
		return err
	}
	if requestSpecFingerprint != responseSpecFingerprint {
		return fmt.Errorf("compute resource response spec does not match the requested replacement: %w", runaimcp.ErrUpstream)
	}
	requestMeta, _ := request["meta"].(map[string]any)
	responseMeta, _ := responseRequest["meta"].(map[string]any)
	requestedDescription, requested := optionalComputeResourceDescription(requestMeta)
	responseDescription, returned := optionalComputeResourceDescription(responseMeta)
	if requested != returned || (requested && requestedDescription != responseDescription) {
		return fmt.Errorf("compute resource response description does not match the requested replacement: %w", runaimcp.ErrUpstream)
	}
	return nil
}

func optionalComputeResourceDescription(meta map[string]any) (string, bool) {
	value, exists := meta["description"]
	if !exists || value == nil {
		return "", false
	}
	description, ok := value.(string)
	return description, ok
}

func nonNilMapField(value map[string]any, field string) (any, bool) {
	fieldValue, exists := value[field]
	return fieldValue, exists && fieldValue != nil
}

func numberAsPositiveInt32(value any) (int32, bool) {
	var number float64
	switch value := value.(type) {
	case float64:
		number = value
	case float32:
		number = float64(value)
	case int:
		number = float64(value)
	case int32:
		number = float64(value)
	case int64:
		number = float64(value)
	case json.Number:
		parsed, err := strconv.ParseFloat(string(value), 64)
		if err != nil {
			return 0, false
		}
		number = parsed
	default:
		return 0, false
	}
	if !common.IsFinite(number) || number <= 0 || number > math.MaxInt32 || math.Trunc(number) != number {
		return 0, false
	}
	return int32(number), true
}
