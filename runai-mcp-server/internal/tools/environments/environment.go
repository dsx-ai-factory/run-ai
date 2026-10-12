/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package environments

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	assets "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

type environmentSnapshot struct {
	id                  string
	name                string
	scope               assets.Scope
	updateCount         *int64
	environment         map[string]any
	updateRequest       map[string]any
	updateBlockedReason string
}

var knownEnvironmentResponseFields = map[string]struct{}{
	"compliance": {},
	"meta":       {},
	"spec":       {},
	"status":     {},
	"usageTimes": {},
	"usedBy":     {},
}

var knownEnvironmentMetaFields = map[string]struct{}{
	"autoDelete":             {},
	"clusterId":              {},
	"createdAt":              {},
	"createdBy":              {},
	"deletedAt":              {},
	"deletedBy":              {},
	"departmentId":           {},
	"description":            {},
	"id":                     {},
	"kind":                   {},
	"name":                   {},
	"private":                {},
	"projectId":              {},
	"projectName":            {},
	"scope":                  {},
	"tenantId":               {},
	"updateCount":            {},
	"updatedAt":              {},
	"updatedBy":              {},
	"workloadSupportedTypes": {},
}

func newEnvironmentSnapshot(environment *assets.EnvironmentAsset, raw map[string]any) (environmentSnapshot, error) {
	if environment == nil {
		return environmentSnapshot{}, fmt.Errorf("nil environment response: %w", runaimcp.ErrUpstream)
	}
	if raw == nil {
		return environmentSnapshot{}, fmt.Errorf("nil environment response body: %w", runaimcp.ErrUpstream)
	}
	meta, ok := raw["meta"].(map[string]any)
	if !ok || meta == nil {
		return environmentSnapshot{}, fmt.Errorf("environment response meta must be an object: %w", runaimcp.ErrUpstream)
	}
	id, err := common.RequiredString("environment response meta.id", meta["id"])
	if err != nil {
		return environmentSnapshot{}, fmt.Errorf("%w: %w", err, runaimcp.ErrUpstream)
	}
	if err := common.ValidateUUID("environment response meta.id", id); err != nil {
		return environmentSnapshot{}, fmt.Errorf("%w: %w", err, runaimcp.ErrUpstream)
	}
	if !strings.EqualFold(id, environment.Meta.Id) {
		return environmentSnapshot{}, fmt.Errorf("environment response id %q does not match decoded id %q: %w",
			id, environment.Meta.Id, runaimcp.ErrUpstream)
	}
	name, err := common.RequiredString("environment response meta.name", meta["name"])
	if err != nil {
		return environmentSnapshot{}, fmt.Errorf("%w: %w", err, runaimcp.ErrUpstream)
	}
	if name != environment.Meta.Name {
		return environmentSnapshot{}, fmt.Errorf("environment response name %q does not match decoded name %q: %w",
			name, environment.Meta.Name, runaimcp.ErrUpstream)
	}
	kind, err := common.RequiredString("environment response meta.kind", meta["kind"])
	if err != nil {
		return environmentSnapshot{}, fmt.Errorf("%w: %w", err, runaimcp.ErrUpstream)
	}
	if kind != string(assets.ASSETKIND_ENVIRONMENT) || environment.Meta.Kind != assets.ASSETKIND_ENVIRONMENT {
		return environmentSnapshot{}, fmt.Errorf("environment response kind must be %q: %w",
			assets.ASSETKIND_ENVIRONMENT, runaimcp.ErrUpstream)
	}
	scope, err := common.RequiredString("environment response meta.scope", meta["scope"])
	if err != nil {
		return environmentSnapshot{}, fmt.Errorf("%w: %w", err, runaimcp.ErrUpstream)
	}
	if assets.Scope(scope) != environment.Meta.Scope || !validEnvironmentResponseScope(environment.Meta.Scope) {
		return environmentSnapshot{}, fmt.Errorf("environment response scope %q does not match decoded valid scope %q: %w",
			scope, environment.Meta.Scope, runaimcp.ErrUpstream)
	}

	snapshot := environmentSnapshot{
		id:          environment.Meta.Id,
		name:        environment.Meta.Name,
		scope:       environment.Meta.Scope,
		environment: raw,
	}
	if updateCount, ok := environment.Meta.GetUpdateCountOk(); ok {
		value := *updateCount
		snapshot.updateCount = &value
	}
	if snapshot.scope == assets.SCOPE_SYSTEM || snapshot.updateCount == nil {
		return snapshot, nil
	}

	request, err := environmentUpdateSnapshot(raw)
	if err != nil {
		snapshot.updateBlockedReason = err.Error()
		return snapshot, nil
	}
	snapshot.updateRequest = request
	return snapshot, nil
}

func validEnvironmentResponseScope(scope assets.Scope) bool {
	switch scope {
	case assets.SCOPE_SYSTEM, assets.SCOPE_TENANT, assets.SCOPE_CLUSTER, assets.SCOPE_DEPARTMENT, assets.SCOPE_PROJECT:
		return true
	default:
		return false
	}
}

func environmentUpdateSnapshot(raw map[string]any) (map[string]any, error) {
	if unknown := unknownEnvironmentFields(raw, knownEnvironmentResponseFields); len(unknown) != 0 {
		return nil, unsupportedEnvironmentFieldsError(unknown)
	}
	meta, _ := raw["meta"].(map[string]any)
	if unknown := unknownEnvironmentFields(meta, knownEnvironmentMetaFields); len(unknown) != 0 {
		for index := range unknown {
			unknown[index] = "meta." + unknown[index]
		}
		return nil, unsupportedEnvironmentFieldsError(unknown)
	}
	name, err := common.RequiredString("environment response meta.name", meta["name"])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", err, runaimcp.ErrUpstream)
	}
	spec, ok := raw["spec"].(map[string]any)
	if !ok || spec == nil {
		return nil, fmt.Errorf("environment response spec must be an object: %w", runaimcp.ErrUpstream)
	}
	requestMeta := map[string]any{"name": name}
	if description, exists := meta["description"]; exists {
		requestMeta["description"] = description
	}
	request := map[string]any{"meta": requestMeta, "spec": spec}
	if _, err := decodeEnvironmentUpdateRequest(request); err != nil {
		return nil, fmt.Errorf("this MCP version cannot safely preserve the complete update request: %w", err)
	}
	return request, nil
}

func unknownEnvironmentFields(value map[string]any, known map[string]struct{}) []string {
	unknown := make([]string, 0)
	for field := range value {
		if _, ok := known[field]; !ok {
			unknown = append(unknown, field)
		}
	}
	sort.Strings(unknown)
	return unknown
}

func unsupportedEnvironmentFieldsError(fields []string) error {
	return fmt.Errorf("environment contains fields this MCP version cannot safely preserve during a full-replacement update (%s); upgrade the MCP server before updating: %w",
		strings.Join(fields, ", "), runaimcp.ErrConflict)
}

func readEnvironmentResponse(resp *http.Response, operation string) (map[string]any, error) {
	var raw map[string]any
	if err := common.ReadJSONResponse(resp, &raw); err != nil {
		return nil, fmt.Errorf("%s: %w: %w", operation, err, runaimcp.ErrUpstream)
	}
	if raw == nil {
		return nil, fmt.Errorf("%s: nil response: %w", operation, runaimcp.ErrUpstream)
	}
	return raw, nil
}

func environmentMatchesScope(actual assets.Scope, requested string) bool {
	if requested == string(assets.SCOPE_TENANT) {
		return actual == assets.SCOPE_TENANT || actual == assets.SCOPE_SYSTEM
	}
	return actual == assets.Scope(requested)
}
