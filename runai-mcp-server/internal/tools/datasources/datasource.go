/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datasources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	assets "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

const dataVolumeType = "dataVolume"

type datasourceType struct {
	kind    assets.AssetKind
	specKey string
}

// Query type, metadata kind, and spec key use different API spellings.
var datasourceTypes = map[string]datasourceType{
	"hostPath": {assets.ASSETKIND_HOST_PATH, "hostPath"}, "nfs": {assets.ASSETKIND_NFS, "nfs"}, "pvc": {assets.ASSETKIND_PVC, "pvc"},
	"git": {assets.ASSETKIND_GIT, "git"}, "s3": {assets.ASSETKIND_S3, "s3"}, "configMap": {assets.ASSETKIND_CONFIG_MAP, "config_map"},
	"secretVolume": {assets.ASSETKIND_SECRET_VOLUME, "secret"}, dataVolumeType: {assets.ASSETKIND_DATA_VOLUME, dataVolumeType},
}

type datasourceSnapshot struct {
	kind                assets.AssetKind
	id                  string
	name                string
	scope               assets.Scope
	updateCount         *int64
	datasource          map[string]any
	updateRequest       map[string]any
	updateBlockedReason string
}

var knownDatasourceResponseFields = map[string]struct{}{
	"compliance":  {},
	"meta":        {},
	"spec":        {},
	"status":      {},
	"usageTimes":  {},
	"usedBy":      {},
	"clusterInfo": {},
}

func newDatasourceSnapshot(raw map[string]any, kind assets.AssetKind) (datasourceSnapshot, error) {
	meta, ok := raw["meta"].(map[string]any)
	if !ok {
		return datasourceSnapshot{}, fmt.Errorf("datasource response meta must be an object: %w", runaimcp.ErrUpstream)
	}
	snapshot, err := datasourceIdentity(meta, kind)
	if err != nil {
		return datasourceSnapshot{}, fmt.Errorf("datasource response: %v: %w", err, runaimcp.ErrUpstream)
	}
	spec, exists := raw["spec"]
	if !exists {
		return datasourceSnapshot{}, fmt.Errorf("datasource response spec is required: %w", runaimcp.ErrUpstream)
	}
	if spec != nil {
		if _, ok := spec.(map[string]any); !ok {
			return datasourceSnapshot{}, fmt.Errorf("datasource response spec must be an object or null: %w", runaimcp.ErrUpstream)
		}
	}
	snapshot.datasource = raw
	return snapshot, nil
}

func prepareDatasourceUpdate(snapshot datasourceSnapshot) datasourceSnapshot {
	if snapshot.scope == assets.SCOPE_SYSTEM {
		snapshot.updateBlockedReason = "system-scoped datasources are read-only"
		return snapshot
	}
	if snapshot.updateCount == nil {
		snapshot.updateBlockedReason = "the Assets API response did not include meta.updateCount"
		return snapshot
	}
	if _, ok := datasourceAdapters[snapshot.kind]; !ok {
		snapshot.updateBlockedReason = "this datasource kind is not supported for updates"
		return snapshot
	}
	spec, _ := snapshot.datasource["spec"].(map[string]any)
	if spec == nil && snapshot.kind != assets.ASSETKIND_CONFIG_MAP && snapshot.kind != assets.ASSETKIND_SECRET_VOLUME {
		snapshot.updateBlockedReason = "the datasource spec is null; a safe replacement request cannot be prepared"
		return snapshot
	}
	request, err := datasourceUpdateSnapshot(snapshot)
	snapshot.updateRequest = request
	if err != nil {
		snapshot.updateBlockedReason = err.Error()
	}
	return snapshot
}

func datasourceIdentity(meta map[string]any, kind assets.AssetKind) (datasourceSnapshot, error) {
	id, _ := meta["id"].(string)
	if err := common.ValidateUUID("meta.id", id); err != nil {
		return datasourceSnapshot{}, err
	}
	name, err := common.RequiredString("meta.name", meta["name"])
	if err != nil {
		return datasourceSnapshot{}, err
	}
	actualKind, err := common.RequiredString("meta.kind", meta["kind"])
	if err != nil {
		return datasourceSnapshot{}, err
	}
	if assets.AssetKind(actualKind) != kind {
		return datasourceSnapshot{}, fmt.Errorf("meta.kind must be %q", kind)
	}
	scope, _ := meta["scope"].(string)
	if scope == "" || !assets.Scope(scope).IsValid() {
		return datasourceSnapshot{}, fmt.Errorf("meta.scope is invalid")
	}
	snapshot := datasourceSnapshot{id: id, name: name, kind: kind, scope: assets.Scope(scope)}
	if value, exists := meta["updateCount"]; exists && value != nil {
		switch value.(type) {
		case json.Number, int, int32, int64:
		default:
			return datasourceSnapshot{}, fmt.Errorf("meta.updateCount must be a non-negative 64-bit integer")
		}
		count, err := strconv.ParseInt(fmt.Sprint(value), 10, 64)
		if err != nil || count < 0 {
			return datasourceSnapshot{}, fmt.Errorf("meta.updateCount must be a non-negative 64-bit integer")
		}
		snapshot.updateCount = &count
	}
	return snapshot, nil
}

func datasourceUpdateSnapshot(snapshot datasourceSnapshot) (map[string]any, error) {
	raw, kind := snapshot.datasource, snapshot.kind
	var unknown []string
	for field := range raw {
		if _, known := knownDatasourceResponseFields[field]; !known {
			unknown = append(unknown, field)
		}
	}
	if len(unknown) != 0 {
		sort.Strings(unknown)
		return nil, fmt.Errorf("datasource contains fields this MCP version cannot safely preserve during an update (%s); upgrade the MCP server before updating: %w",
			strings.Join(unknown, ", "), runaimcp.ErrConflict)
	}
	meta, _ := raw["meta"].(map[string]any)
	var decoded assets.AssetMeta
	if err := common.DecodeTypedRequest(meta, &decoded); err != nil {
		return nil, err
	}
	spec, _ := raw["spec"].(map[string]any)
	requestMeta := map[string]any{"name": snapshot.name}
	if description, exists := meta["description"]; exists {
		requestMeta["description"] = description
	}
	request := map[string]any{"meta": requestMeta}
	updateSpec, err := datasourceUpdateSpec(kind, spec)
	if err != nil {
		return nil, err
	}
	if updateSpec != nil {
		request["spec"] = updateSpec
	}

	if err := validateDatasourceUpdateRequest(kind, request); err != nil {
		return nil, fmt.Errorf("this MCP version cannot safely preserve the complete update request: %w", err)
	}
	return request, nil
}

func datasourceUpdateSpec(kind assets.AssetKind, spec map[string]any) (map[string]any, error) {
	switch kind {
	case assets.ASSETKIND_CONFIG_MAP:
		var decoded assets.ConfigMapAssetSpec
		return nil, common.DecodeTypedRequest(spec, &decoded)
	case assets.ASSETKIND_SECRET_VOLUME:
		var decoded assets.SecretAssetSpec
		return nil, common.DecodeTypedRequest(spec, &decoded)
	case assets.ASSETKIND_PVC:
		var decoded assets.PVCAssetSpec
		if err := common.DecodeTypedRequest(spec, &decoded); err != nil {
			return nil, err
		}
		update := map[string]any{"path": spec["path"]}
		if claim, ok := spec["claimInfo"].(map[string]any); ok {
			fields := map[string]any{}
			for _, field := range []string{"size", "addedAttrValues"} {
				if value, exists := claim[field]; exists {
					fields[field] = value
				}
			}
			update["claimInfo"] = fields
		}
		return update, nil
	default:
		return spec, nil
	}
}

func readDatasourceResponse(resp *http.Response, operation string) (map[string]any, error) {
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("%s: missing response body: %w", operation, runaimcp.ErrUpstream)
	}
	defer resp.Body.Close()
	// Keep int64 update counters exact for the stale-snapshot check.
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	var raw map[string]any
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("%s: decode response: %v: %w", operation, err, runaimcp.ErrUpstream)
	}

	if raw == nil {
		return nil, fmt.Errorf("%s: nil response: %w", operation, runaimcp.ErrUpstream)
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("%s: unexpected content after response object: %w", operation, runaimcp.ErrUpstream)
	}
	return raw, nil
}

func readDatasourceSnapshot(resp *http.Response, operation string, kind assets.AssetKind) (datasourceSnapshot, error) {
	raw, err := readDatasourceResponse(resp, operation)
	if err != nil {
		return datasourceSnapshot{}, err
	}
	snapshot, err := newDatasourceSnapshot(raw, kind)
	if err != nil {
		return datasourceSnapshot{}, fmt.Errorf("%s: %w", operation, err)
	}
	return snapshot, nil
}

func datasourceSpecKey(kind assets.AssetKind) string {
	for _, entry := range datasourceTypes {
		if entry.kind == kind {
			return entry.specKey
		}
	}
	return ""
}
