/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datasources

import (
	"fmt"

	assets "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

func validateDatasourceCreateRequest(kind assets.AssetKind, raw map[string]any) error {
	if err := validateDatasourceRequest(raw, common.ActionCreate, kind); err != nil {
		return err
	}
	meta := raw["meta"].(map[string]any)
	if meta["scope"] == string(assets.SCOPE_SYSTEM) {
		return fmt.Errorf("system-scoped datasources are read-only")
	}
	return datasourceAdapters[kind].validateCreate(raw)
}

func validateDatasourceUpdateRequest(kind assets.AssetKind, raw map[string]any) error {
	if err := validateDatasourceRequest(raw, common.ActionUpdate, kind); err != nil {
		return err
	}
	return datasourceAdapters[kind].validateUpdate(raw)
}

func validateDatasourceRequest(raw map[string]any, action common.WriteAction, kind assets.AssetKind) error {
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
	if action == common.ActionUpdate && (kind == assets.ASSETKIND_CONFIG_MAP || kind == assets.ASSETKIND_SECRET_VOLUME) {
		return nil
	}
	if spec, ok := raw["spec"].(map[string]any); !ok || spec == nil {
		return fmt.Errorf("request.spec is required and must be an object")
	}
	return nil
}

func (in manageDatasourceInput) validateTarget() error {
	if in.DatasourceID != "" && in.Name != "" {
		return fmt.Errorf("datasourceId and name are mutually exclusive")
	}
	if in.Action == common.ActionCreate {
		if in.Name != "" || in.Scope != "" || in.ScopeID != "" {
			return fmt.Errorf("name/scope/scopeId selectors are for update/delete; create uses request.meta and optional scopeName")
		}
		if in.ScopeName == "" {
			return nil
		}
		meta, _ := in.Request["meta"].(map[string]any)
		for _, field := range []string{"projectId", "departmentId", "clusterId"} {
			if meta[field] != nil {
				return fmt.Errorf("scopeName and request.meta.%s are mutually exclusive", field)
			}
		}
		scope, _ := meta["scope"].(string)
		return validateDatasourceScopeName(assets.Scope(scope), in.ScopeName)
	}
	if in.Name != "" {
		if _, err := common.RequiredString("name", in.Name); err != nil {
			return err
		}
	}
	if in.Scope == "" && in.ScopeID == "" && in.ScopeName == "" {
		return nil
	}
	if in.Name == "" {
		return fmt.Errorf("scope selectors require name")
	}
	if !in.Scope.IsValid() {
		return fmt.Errorf("scope must be a valid asset scope")
	}
	if in.ScopeID != "" && in.ScopeName != "" {
		return fmt.Errorf("scopeId and scopeName are mutually exclusive")
	}
	if in.ScopeID == "" && in.ScopeName == "" {
		return nil
	}
	if in.Scope != assets.SCOPE_PROJECT && in.Scope != assets.SCOPE_DEPARTMENT && in.Scope != assets.SCOPE_CLUSTER {
		return fmt.Errorf("scopeId/scopeName require project, department, or cluster scope")
	}
	if in.ScopeName != "" {
		return validateDatasourceScopeName(in.Scope, in.ScopeName)
	}
	switch in.Scope {
	case assets.SCOPE_PROJECT:
		return common.ValidateNumericString("scopeId", in.ScopeID)
	case assets.SCOPE_DEPARTMENT:
		_, err := common.RequiredString("scopeId", in.ScopeID)
		return err
	}
	return nil
}

func validateDatasourceScopeName(scope assets.Scope, name string) error {
	if scope != assets.SCOPE_PROJECT && scope != assets.SCOPE_DEPARTMENT && scope != assets.SCOPE_CLUSTER {
		return fmt.Errorf("scopeName requires project, department, or cluster scope")
	}
	if _, err := common.RequiredString("scopeName", name); err != nil {
		return err
	}
	return common.ValidateFilterValue("scopeName", name)
}
