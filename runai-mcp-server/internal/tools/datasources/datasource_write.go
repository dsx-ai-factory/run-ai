/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datasources

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"reflect"
	"strings"

	assets "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

func flattenDatasourceFields(value map[string]any) map[string]any {
	fields := make(map[string]any)
	var visit func(string, map[string]any)
	visit = func(prefix string, object map[string]any) {
		for key, value := range object {
			path := key
			if prefix != "" {
				path = prefix + "." + key
			}
			if nested, ok := value.(map[string]any); ok && len(nested) > 0 {
				visit(path, nested)
			} else {
				fields[path] = value
			}
		}
	}
	visit("", value)
	return fields
}

func datasourceUpdatePayload(kind assets.AssetKind, request map[string]any, current datasourceSnapshot) map[string]any {
	if kind != assets.ASSETKIND_PVC {
		return request
	}
	spec, _ := request["spec"].(map[string]any)
	claim, _ := spec["claimInfo"].(map[string]any)
	size, hasSize := claim["size"]
	currentSpec, _ := current.datasource["spec"].(map[string]any)
	currentClaim, _ := currentSpec["claimInfo"].(map[string]any)
	if !hasSize || !reflect.DeepEqual(size, currentClaim["size"]) {
		return request
	}

	// Sending an unchanged size still invokes the API's Bound-PVC resize checks.
	// Copy only the modified maps so the caller's request remains intact.
	payload := maps.Clone(request)
	spec = maps.Clone(spec)
	claim = maps.Clone(claim)
	delete(claim, "size")
	spec["claimInfo"] = claim
	payload["spec"] = spec
	return payload
}

func confirmDatasourceDelete(ctx context.Context, client *runaiclient.Client, in manageDatasourceInput, current datasourceSnapshot) error {
	details := common.WithConfirmWarning(current.datasource, "deleting a datasource is permanent and may fail while it is in use")
	if err := common.ConfirmWrite(ctx, in.Confirm, string(common.ActionDelete), "datasource", current.name, current.id, details); err != nil {
		return err
	}
	if !runaimcp.ClientCanElicitForm(runaimcp.SessionFromContext(ctx)) {
		return nil
	}

	// A prompt can remain open while another caller renames or changes the asset.
	latest, err := getDatasource(ctx, client, in.Kind, in.DatasourceID)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(current.datasource["meta"], latest.datasource["meta"]) ||
		!reflect.DeepEqual(current.datasource["spec"], latest.datasource["spec"]) {
		return fmt.Errorf("datasource %q changed during confirmation; inspect it again before deleting: %w", in.DatasourceID, runaimcp.ErrConflict)
	}
	return nil
}

func datasourceMutationAccepted(resp *http.Response) bool {
	return resp != nil && resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices
}

func datasourceMutationResult(in manageDatasourceInput, out *manageDatasourceOutput, resp *http.Response, callErr error) (*manageDatasourceOutput, error) {
	if !datasourceMutationAccepted(resp) {
		if callErr == nil {
			callErr = fmt.Errorf("missing successful HTTP response")
		}
		return out, fmt.Errorf("%s datasource: %w", in.Action, runaimcp.MapHTTPError(resp, callErr))
	}
	out.Applied = true
	snapshot, err := readDatasourceSnapshot(resp, string(in.Action)+" datasource", in.Kind)
	if err != nil || !datasourceMutationMatches(in, snapshot) {
		// A decode failure after HTTP success must not encourage a duplicate mutation.
		out.Datasource = nil
		out.Warnings = []string{fmt.Sprintf("Assets API accepted the %s but its response could not be verified. Do not retry automatically; use list_datasources to inspect the result.", in.Action)}
		return out, nil
	}
	result := datasourceMutationOutput(in.Action, snapshot)
	result.Applied = true
	return result, nil
}

func datasourceMutationMatches(in manageDatasourceInput, snapshot datasourceSnapshot) bool {
	meta, _ := in.Request["meta"].(map[string]any)
	name, _ := meta["name"].(string)
	if snapshot.name != name {
		return false
	}
	if in.Action == common.ActionUpdate {
		return strings.EqualFold(snapshot.id, in.DatasourceID)
	}
	var requested assets.AssetCreationRequest
	var actual assets.AssetMeta
	if err := common.DecodeTypedRequest(meta, &requested); err != nil {
		return false
	}
	if err := common.DecodeTypedRequest(snapshot.datasource["meta"], &actual); err != nil {
		return false
	}
	if actual.Scope != requested.Scope {
		return false
	}
	switch requested.Scope {
	case assets.SCOPE_PROJECT:
		return actual.GetProjectId() == requested.GetProjectId()
	case assets.SCOPE_DEPARTMENT:
		return actual.GetDepartmentId() == requested.GetDepartmentId()
	case assets.SCOPE_CLUSTER:
		return strings.EqualFold(actual.GetClusterId(), requested.GetClusterId())
	default:
		return requested.Scope == assets.SCOPE_TENANT
	}
}
