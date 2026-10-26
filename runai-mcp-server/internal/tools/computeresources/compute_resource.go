/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package computeresources

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	assets "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

var knownComputeResourceResponseFields = map[string]struct{}{
	"compliance": {},
	"meta":       {},
	"spec":       {},
	"usageTimes": {},
	"usedBy":     {},
}

const (
	computeResourceFingerprintPrefix = "sha256:"
	computeResourceScopeField        = "scope"
	computeResourceClusterIDField    = "clusterId"
	computeResourceDepartmentIDField = "departmentId"
	computeResourceProjectIDField    = "projectId"
)

var knownComputeResourceMetaFields = map[string]struct{}{
	"autoDelete":                     {},
	computeResourceClusterIDField:    {},
	"createdAt":                      {},
	"createdBy":                      {},
	"deletedAt":                      {},
	"deletedBy":                      {},
	computeResourceDepartmentIDField: {},
	"description":                    {},
	"id":                             {},
	"kind":                           {},
	"name":                           {},
	"private":                        {},
	computeResourceProjectIDField:    {},
	"projectName":                    {},
	computeResourceScopeField:        {},
	"tenantId":                       {},
	"updateCount":                    {},
	"updatedAt":                      {},
	"updatedBy":                      {},
	"workloadSupportedTypes":         {},
}

func computeResourceUpdateSnapshot(raw map[string]any) (map[string]any, string, error) {
	if unknown := unknownComputeResourceFields(raw, knownComputeResourceResponseFields); len(unknown) != 0 {
		return nil, "", unsupportedComputeResourceFieldsError(unknown)
	}
	meta, err := computeResourceMeta(raw)
	if err != nil {
		return nil, "", err
	}
	if unknown := unknownComputeResourceFields(meta, knownComputeResourceMetaFields); len(unknown) != 0 {
		for i := range unknown {
			unknown[i] = "meta." + unknown[i]
		}
		return nil, "", unsupportedComputeResourceFieldsError(unknown)
	}
	name, err := computeResourceMetaString(meta, "name")
	if err != nil {
		return nil, "", err
	}
	requestMeta := map[string]any{"name": name}
	if description, exists := meta["description"]; exists {
		requestMeta["description"] = description
	}
	spec, ok := raw["spec"].(map[string]any)
	if !ok || spec == nil {
		return nil, "", fmt.Errorf("compute resource response spec must be an object: %w", runaimcp.ErrUpstream)
	}
	request := map[string]any{
		"meta": requestMeta,
		"spec": spec,
	}
	identity := map[string]any{}
	for _, field := range []string{
		"id",
		computeResourceScopeField,
		computeResourceClusterIDField,
		computeResourceDepartmentIDField,
		computeResourceProjectIDField,
	} {
		if value, exists := meta[field]; exists {
			identity[field] = value
		}
	}
	fingerprintValue := map[string]any{"identity": identity, "request": request}
	fingerprint, err := computeResourceJSONFingerprint(fingerprintValue)
	if err != nil {
		return nil, "", fmt.Errorf("fingerprint compute resource update request: %w", err)
	}
	return request, fingerprint, nil
}

func validateComputeResourceFingerprint(value string) error {
	if !isValidComputeResourceFingerprint(value) {
		return fmt.Errorf("expectedFingerprint must be the value returned by list_compute_resources")
	}
	return nil
}

func readComputeResourceResponse(resp *http.Response, operation string) (map[string]any, error) {
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("%s: missing response body: %w", operation, runaimcp.ErrUpstream)
	}
	defer resp.Body.Close()
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("%s: decode response: %w: %w", operation, err, runaimcp.ErrUpstream)
	}
	if raw == nil {
		return nil, fmt.Errorf("%s: nil response: %w", operation, runaimcp.ErrUpstream)
	}
	return raw, nil
}

func computeResourceJSONFingerprint(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return computeResourceFingerprintPrefix + hex.EncodeToString(sum[:]), nil
}

func isValidComputeResourceFingerprint(value string) bool {
	if !strings.HasPrefix(value, computeResourceFingerprintPrefix) {
		return false
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(value, computeResourceFingerprintPrefix))
	return err == nil && len(decoded) == sha256.Size && computeResourceFingerprintPrefix+hex.EncodeToString(decoded) == value
}

type resourceIdentity struct {
	id    string
	name  string
	scope string
}

func computeResourceIdentity(raw map[string]any) (resourceIdentity, error) {
	meta, err := computeResourceMeta(raw)
	if err != nil {
		return resourceIdentity{}, err
	}
	id, err := computeResourceMetaString(meta, "id")
	if err != nil {
		return resourceIdentity{}, err
	}
	if err := common.ValidateUUID("meta.id", id); err != nil {
		return resourceIdentity{}, fmt.Errorf("compute resource response: %w: %w", err, runaimcp.ErrUpstream)
	}
	kind, err := computeResourceMetaString(meta, "kind")
	if err != nil {
		return resourceIdentity{}, err
	}
	if assets.AssetKind(kind) != assets.ASSETKIND_COMPUTE {
		return resourceIdentity{}, fmt.Errorf("compute resource response kind must be %q (got %q): %w",
			assets.ASSETKIND_COMPUTE, kind, runaimcp.ErrUpstream)
	}
	name, err := computeResourceMetaString(meta, "name")
	if err != nil {
		return resourceIdentity{}, err
	}
	scope, err := computeResourceMetaString(meta, computeResourceScopeField)
	if err != nil {
		return resourceIdentity{}, err
	}
	switch assets.Scope(scope) {
	case assets.SCOPE_SYSTEM, assets.SCOPE_TENANT:
	case assets.SCOPE_CLUSTER:
		clusterID, err := computeResourceMetaString(meta, computeResourceClusterIDField)
		if err != nil {
			return resourceIdentity{}, err
		}
		if err := common.ValidateUUID("meta.clusterId", clusterID); err != nil {
			return resourceIdentity{}, fmt.Errorf("compute resource response: %w: %w", err, runaimcp.ErrUpstream)
		}
	case assets.SCOPE_DEPARTMENT:
		if _, err := computeResourceMetaString(meta, computeResourceDepartmentIDField); err != nil {
			return resourceIdentity{}, err
		}
	case assets.SCOPE_PROJECT:
		if projectID, ok := numberAsPositiveInt32(meta[computeResourceProjectIDField]); !ok || projectID <= 0 {
			return resourceIdentity{}, fmt.Errorf("compute resource response meta.projectId must be a positive 32-bit integer: %w", runaimcp.ErrUpstream)
		}
	default:
		return resourceIdentity{}, fmt.Errorf("compute resource response has invalid scope %q: %w", scope, runaimcp.ErrUpstream)
	}
	if private, exists := meta["private"]; exists && private == true {
		return resourceIdentity{}, fmt.Errorf("compute resource response must not be private: %w", runaimcp.ErrUpstream)
	}
	if description, exists := meta["description"]; exists && description != nil {
		value, ok := description.(string)
		if !ok || len(value) == 0 || len(value) > 250 {
			return resourceIdentity{}, fmt.Errorf("compute resource response description must contain between 1 and 250 UTF-8 bytes: %w", runaimcp.ErrUpstream)
		}
	}
	return resourceIdentity{id: id, name: name, scope: scope}, nil
}

func computeResourceMeta(raw map[string]any) (map[string]any, error) {
	if meta, ok := raw["meta"].(map[string]any); ok && meta != nil {
		return meta, nil
	}
	return nil, fmt.Errorf("compute resource response meta must be an object: %w", runaimcp.ErrUpstream)
}

func computeResourceMetaString(meta map[string]any, field string) (string, error) {
	value, err := common.RequiredString("meta."+field, meta[field])
	if err != nil {
		return "", fmt.Errorf("compute resource response: %w: %w", err, runaimcp.ErrUpstream)
	}
	return value, nil
}

func unknownComputeResourceFields(value map[string]any, known map[string]struct{}) []string {
	unknown := make([]string, 0)
	for field := range value {
		if _, ok := known[field]; !ok {
			unknown = append(unknown, field)
		}
	}
	sort.Strings(unknown)
	return unknown
}

func unsupportedComputeResourceFieldsError(fields []string) error {
	return fmt.Errorf("compute resource contains fields this MCP version cannot safely preserve during a full-replacement update (%s); upgrade the MCP server before updating: %w",
		strings.Join(fields, ", "), runaimcp.ErrConflict)
}
