/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	authz "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	bff "github.com/run-ai/runai/control-plane/bff-service/pkg/client/generated"
	clusterservice "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	catalog "github.com/run-ai/runai/runai-clients/catalog-service-client/generated"
	datavolumes "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
	orgunit "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	workloads "github.com/run-ai/runai/runai-clients/workloads-client/generated"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

// Categories of tenant-wide boolean settings, each backed by its own service. Security
// (heterogeneous values) and UI cosmetics are intentionally out of scope.
const (
	catWorkloads        = "workloads"
	catAuthorization    = "authorization"
	catCluster          = "cluster"
	catOrgUnit          = "org-unit"
	catCatalog          = "catalog"
	catDataVolumes      = "datavolumes"
	catNodePool         = "node-pool"
	catWorkloadsManager = "workloads-manager"
)

// settingKey is one setting within a category. updatable=false means the platform exposes it
// for reading but it is not an operator-toggleable feature flag (e.g. a provisioning-owned
// tenant attribute): list shows it, update rejects it.
type settingKey struct {
	name      string
	updatable bool
}

// categoryDef is a category and its keys in stable listing order — the single source of truth
// for which categories/keys exist and which are writable. Read/update dispatch keys off the
// category name only.
type categoryDef struct {
	name string
	keys []settingKey
}

// rw builds a list of writable keys (the common case).
func rw(names ...string) []settingKey {
	keys := make([]settingKey, len(names))
	for i, n := range names {
		keys[i] = settingKey{name: n, updatable: true}
	}
	return keys
}

var categoryDefs = []categoryDef{
	{catWorkloads, rw("enablePolicies", "workloadSubmissionV2Enabled", "enableTemplatesV2", "enableHybridSubmission", "gpuOverProvisioning", "applicationSubmission")},
	// tenantIsCspManaged is a provisioning-owned tenant attribute, not an operator feature flag: listable, not updatable.
	{catAuthorization, []settingKey{{"enableScopedAuthorization", true}, {"tenantIsCspManaged", false}}},
	{catCluster, rw("enableAdvancedGpuMetrics")},
	{catOrgUnit, rw("limitQuotaOverSubscription")},
	{catCatalog, rw("enableNgcIntegration")},
	{catDataVolumes, rw("enableDataVolumes")},
	{catNodePool, rw("projectManualOverQuota", "cpuResourcesQuota")},
	{catWorkloadsManager, rw("enableNIM", "deletedWorkloadsView", "allowGPT2Model", "enableModelCatalog")},
}

func categories() []string {
	out := make([]string, len(categoryDefs))
	for i, c := range categoryDefs {
		out[i] = c.name
	}
	return out
}

func categoryDefByName(name string) (categoryDef, bool) {
	for _, c := range categoryDefs {
		if c.name == name {
			return c, true
		}
	}
	return categoryDef{}, false
}

func validCategory(c string) bool {
	_, ok := categoryDefByName(c)
	return ok
}

func keyNames(category string) []string {
	def, _ := categoryDefByName(category)
	names := make([]string, len(def.keys))
	for i, k := range def.keys {
		names[i] = k.name
	}
	return names
}

func validKey(category, key string) bool {
	def, ok := categoryDefByName(category)
	if !ok {
		return false
	}
	for _, k := range def.keys {
		if k.name == key {
			return true
		}
	}
	return false
}

func keyUpdatable(category, key string) bool {
	def, ok := categoryDefByName(category)
	if !ok {
		return false
	}
	for _, k := range def.keys {
		if k.name == key {
			return k.updatable
		}
	}
	return false
}

// settingValue is a setting's current state plus its self-describing metadata. enabled is nil
// when the platform has no value set. displayName/description/stage are nil for services that
// don't return them (authorization) — the rest of the platform reports all three, including the
// lifecycle stage (GA, Beta, Alpha, Legacy).
type settingValue struct {
	enabled     *bool
	displayName *string
	description *string
	stage       *string
}

// readSetting returns only the current enabled value — used by the update path's before/after
// snapshots, which don't need metadata. See readSettingValue for the not-found rationale.
func readSetting(ctx context.Context, client *runaiclient.Client, category, key string) (*bool, error) {
	v, err := readSettingValue(ctx, client, category, key)
	return v.enabled, err
}

// readSettingValue returns a setting's value and metadata, or a zero value when the platform
// hasn't set one. Some services signal "never set" with a 404 rather than a defaulted 200, so a
// not-found is folded into the unset case — otherwise listing all settings would abort on the
// first unset one, and an unset setting could never be turned on.
func readSettingValue(ctx context.Context, client *runaiclient.Client, category, key string) (settingValue, error) {
	v, err := readSettingValueRaw(ctx, client, category, key)
	if common.IsNotFound(err) {
		return settingValue{}, nil
	}
	return v, err
}

func readSettingValueRaw(ctx context.Context, client *runaiclient.Client, category, key string) (settingValue, error) {
	switch category {
	case catWorkloads:
		s, httpResp, err := client.Workloads.SettingsAPI.GetWorkloadsServiceSettingByKey(ctx, workloads.WorkloadsSettingsKeyEnum(key)).Execute()
		if err != nil {
			return settingValue{}, fmt.Errorf("read workloads setting %q: %w", key, runaimcp.MapHTTPError(httpResp, err))
		}
		if s == nil {
			return settingValue{}, nil
		}
		var enabled *bool
		if s.Value != nil {
			enabled = s.Value.Enabled.Get()
		}
		return settingValue{enabled: enabled, displayName: s.DisplayName, description: s.Description, stage: s.Stage}, nil
	case catAuthorization:
		s, httpResp, err := client.Authorization.SettingsAPI.GetAuthorizationSettingByKey(ctx, authz.AuthorizationSettingsKeyEnum(key)).Execute()
		if err != nil {
			return settingValue{}, fmt.Errorf("read authorization setting %q: %w", key, runaimcp.MapHTTPError(httpResp, err))
		}
		if s == nil {
			return settingValue{}, nil
		}
		return settingValue{enabled: s.Enabled.Get()}, nil
	case catCluster:
		s, httpResp, err := client.Cluster.SettingsAPI.GetClusterSettingByKey(ctx, clusterservice.ClusterSettingsKeyEnum(key)).Execute()
		if err != nil {
			return settingValue{}, fmt.Errorf("read cluster setting %q: %w", key, runaimcp.MapHTTPError(httpResp, err))
		}
		if s == nil {
			return settingValue{}, nil
		}
		return settingValue{enabled: s.Enabled.Get(), displayName: s.DisplayName, description: s.Description, stage: s.Stage}, nil
	case catOrgUnit:
		s, httpResp, err := client.OrgUnit.SettingsAPI.GetOrgUnitSettingByKey(ctx, orgunit.OrgUnitSettingsKeyEnum(key)).Execute()
		if err != nil {
			return settingValue{}, fmt.Errorf("read org-unit setting %q: %w", key, runaimcp.MapHTTPError(httpResp, err))
		}
		if s == nil {
			return settingValue{}, nil
		}
		return settingValue{enabled: s.Enabled.Get(), displayName: s.DisplayName, description: s.Description, stage: s.Stage}, nil
	case catCatalog:
		s, httpResp, err := client.Catalog.SettingsAPI.GetCatalogSettingByKey(ctx, catalog.CatalogSettingsKeyEnum(key)).Execute()
		if err != nil {
			return settingValue{}, fmt.Errorf("read catalog setting %q: %w", key, runaimcp.MapHTTPError(httpResp, err))
		}
		if s == nil {
			return settingValue{}, nil
		}
		var enabled *bool
		if s.Value != nil {
			enabled = s.Value.Enabled.Get()
		}
		return settingValue{enabled: enabled, displayName: s.DisplayName, description: s.Description, stage: s.Stage}, nil
	case catDataVolumes:
		s, httpResp, err := client.DataVolumes.SettingsAPI.GetDatavolumesSettingByKey(ctx, datavolumes.DatavolumesSettingsKeyEnum(key)).Execute()
		if err != nil {
			return settingValue{}, fmt.Errorf("read datavolumes setting %q: %w", key, runaimcp.MapHTTPError(httpResp, err))
		}
		if s == nil {
			return settingValue{}, nil
		}
		var enabled *bool
		if s.Value != nil {
			enabled = s.Value.Enabled.Get()
		}
		return settingValue{enabled: enabled, displayName: s.DisplayName, description: s.Description, stage: s.Stage}, nil
	case catNodePool:
		s, httpResp, err := client.Cluster.SettingsAPI.GetNodePoolSettingByKey(ctx, clusterservice.NodePoolSettingsKeyEnum(key)).Execute()
		if err != nil {
			return settingValue{}, fmt.Errorf("read node-pool setting %q: %w", key, runaimcp.MapHTTPError(httpResp, err))
		}
		if s == nil {
			return settingValue{}, nil
		}
		return settingValue{enabled: s.Enabled.Get(), displayName: s.DisplayName, description: s.Description, stage: s.Stage}, nil
	case catWorkloadsManager:
		resp, httpResp, err := client.BFF.SettingsAPI.GetWorkloadsSettingByKey(ctx, bff.WorkloadsSettingsKeyEnum(key)).Execute()
		if err != nil {
			return settingValue{}, fmt.Errorf("read workloads-manager setting %q: %w", key, runaimcp.MapHTTPError(httpResp, err))
		}
		return settingValueFromBFFMap(resp)
	}
	return settingValue{}, fmt.Errorf("unknown settings category %q: %w", category, runaimcp.ErrValidation)
}

// readCategorySettings fetches all of a category's settings in a single get-all call, keyed by
// setting key. list_system_settings uses this rather than per-key reads: unlike the by-key
// endpoint (whose response some services return without displayName/description/stage), get-all
// populates that metadata for every service, matching what the UI shows — and costs one request
// per category instead of one per key. A not-found (nothing ever set) folds to an empty map.
func readCategorySettings(ctx context.Context, client *runaiclient.Client, category string) (map[string]settingValue, error) {
	raw, err := fetchCategorySettings(ctx, client, category)
	if common.IsNotFound(err) {
		return map[string]settingValue{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := make(map[string]settingValue, len(raw))
	for k, entry := range raw {
		out[k] = settingValueFromEnvelope(entry)
	}
	return out, nil
}

// Every category's get-all response marshals to an object keyed by setting key, so one JSON
// round-trip normalizes the differing typed shapes (flat, value-wrapped, and the BFF's untyped
// map) into a single map[key]raw.
func fetchCategorySettings(ctx context.Context, client *runaiclient.Client, category string) (map[string]json.RawMessage, error) {
	switch category {
	case catWorkloads:
		resp, httpResp, err := client.Workloads.SettingsAPI.GetWorkloadsServiceSettings(ctx).Execute()
		return settingsGetAllToMap(resp, httpResp, err, category)
	case catAuthorization:
		resp, httpResp, err := client.Authorization.SettingsAPI.GetAuthorizationSettings(ctx).Execute()
		return settingsGetAllToMap(resp, httpResp, err, category)
	case catCluster:
		resp, httpResp, err := client.Cluster.SettingsAPI.GetClusterSettings(ctx).Execute()
		return settingsGetAllToMap(resp, httpResp, err, category)
	case catOrgUnit:
		resp, httpResp, err := client.OrgUnit.SettingsAPI.GetOrgUnitSettings(ctx).Execute()
		return settingsGetAllToMap(resp, httpResp, err, category)
	case catCatalog:
		resp, httpResp, err := client.Catalog.SettingsAPI.GetCatalogSettings(ctx).Execute()
		return settingsGetAllToMap(resp, httpResp, err, category)
	case catDataVolumes:
		resp, httpResp, err := client.DataVolumes.SettingsAPI.GetDatavolumesSettings(ctx).Execute()
		return settingsGetAllToMap(resp, httpResp, err, category)
	case catNodePool:
		resp, httpResp, err := client.Cluster.SettingsAPI.GetNodePoolSettings(ctx).Execute()
		return settingsGetAllToMap(resp, httpResp, err, category)
	case catWorkloadsManager:
		resp, httpResp, err := client.BFF.SettingsAPI.GetWorkloadsSettings(ctx).Execute()
		return settingsGetAllToMap(resp, httpResp, err, category)
	}
	return nil, fmt.Errorf("unknown settings category %q: %w", category, runaimcp.ErrValidation)
}

func settingsGetAllToMap(resp any, httpResp *http.Response, err error, category string) (map[string]json.RawMessage, error) {
	if err != nil {
		return nil, fmt.Errorf("read %s settings: %w", category, runaimcp.MapHTTPError(httpResp, err))
	}
	b, mErr := json.Marshal(resp)
	if mErr != nil {
		return nil, fmt.Errorf("encode %s settings: %w", category, runaimcp.ErrInternal)
	}
	var m map[string]json.RawMessage
	if uErr := json.Unmarshal(b, &m); uErr != nil {
		return nil, fmt.Errorf("decode %s settings: %w", category, runaimcp.ErrUpstream)
	}
	return m, nil
}

// settingEnvelope is the common subset of every service's per-setting object: enabled either at
// the top level (flat services) or nested under value (workloads/catalog/datavolumes/BFF), plus
// the shared metadata.
type settingEnvelope struct {
	Enabled *bool `json:"enabled"`
	Value   *struct {
		Enabled *bool `json:"enabled"`
	} `json:"value"`
	DisplayName *string `json:"displayName"`
	Description *string `json:"description"`
	Stage       *string `json:"stage"`
}

func settingValueFromEnvelope(raw json.RawMessage) settingValue {
	var e settingEnvelope
	_ = json.Unmarshal(raw, &e)
	enabled := e.Enabled
	if enabled == nil && e.Value != nil {
		enabled = e.Value.Enabled
	}
	return settingValue{enabled: enabled, displayName: e.DisplayName, description: e.Description, stage: e.Stage}
}

// responseKey maps a catalog key to the field name a category's get-all response uses for it.
// They match everywhere except the BFF, whose response spells enableNIM as "enableNim".
func responseKey(category, key string) string {
	if category == catWorkloadsManager && key == "enableNIM" {
		return "enableNim"
	}
	return key
}

func updateSetting(ctx context.Context, client *runaiclient.Client, category, key string, enabled bool) error {
	switch category {
	case catWorkloads:
		body := workloads.UpdateWorkloadsSettingByKeyRequest{Enabled: *workloads.NewNullableBool(&enabled)}
		_, httpResp, err := client.Workloads.SettingsAPI.UpdateWorkloadsServiceSettingByKey(ctx, workloads.WorkloadsSettingsKeyEnum(key)).UpdateWorkloadsSettingByKeyRequest(body).Execute()
		if err != nil {
			return fmt.Errorf("update workloads setting %q: %w", key, runaimcp.MapHTTPError(httpResp, err))
		}
		return nil
	case catAuthorization:
		body := authz.UpdateAuthorizationSettingByKeyRequest{Enabled: *authz.NewNullableBool(&enabled)}
		httpResp, err := client.Authorization.SettingsAPI.UpdateAuthorizationSettingByKey(ctx, authz.AuthorizationSettingsKeyEnum(key)).UpdateAuthorizationSettingByKeyRequest(body).Execute()
		if err != nil {
			return fmt.Errorf("update authorization setting %q: %w", key, runaimcp.MapHTTPError(httpResp, err))
		}
		return nil
	case catCluster:
		body := clusterservice.UpdateClusterSettingByKeyRequest{Enabled: *clusterservice.NewNullableBool(&enabled)}
		httpResp, err := client.Cluster.SettingsAPI.UpdateClusterSettingByKey(ctx, clusterservice.ClusterSettingsKeyEnum(key)).UpdateClusterSettingByKeyRequest(body).Execute()
		if err != nil {
			return fmt.Errorf("update cluster setting %q: %w", key, runaimcp.MapHTTPError(httpResp, err))
		}
		return nil
	case catOrgUnit:
		body := orgunit.UpdateOrgUnitSettingByKeyRequest{Enabled: *orgunit.NewNullableBool(&enabled)}
		httpResp, err := client.OrgUnit.SettingsAPI.UpdateOrgUnitSettingByKey(ctx, orgunit.OrgUnitSettingsKeyEnum(key)).UpdateOrgUnitSettingByKeyRequest(body).Execute()
		if err != nil {
			return fmt.Errorf("update org-unit setting %q: %w", key, runaimcp.MapHTTPError(httpResp, err))
		}
		return nil
	case catCatalog:
		body := catalog.UpdateCatalogSettingByKeyRequest{Enabled: *catalog.NewNullableBool(&enabled)}
		httpResp, err := client.Catalog.SettingsAPI.UpdateCatalogSettingByKey(ctx, catalog.CatalogSettingsKeyEnum(key)).UpdateCatalogSettingByKeyRequest(body).Execute()
		if err != nil {
			return fmt.Errorf("update catalog setting %q: %w", key, runaimcp.MapHTTPError(httpResp, err))
		}
		return nil
	case catDataVolumes:
		body := datavolumes.UpdateDatavolumesSettingByKeyRequest{Enabled: *datavolumes.NewNullableBool(&enabled)}
		httpResp, err := client.DataVolumes.SettingsAPI.UpdateDatavolumesSettingByKey(ctx, datavolumes.DatavolumesSettingsKeyEnum(key)).UpdateDatavolumesSettingByKeyRequest(body).Execute()
		if err != nil {
			return fmt.Errorf("update datavolumes setting %q: %w", key, runaimcp.MapHTTPError(httpResp, err))
		}
		return nil
	case catNodePool:
		body := clusterservice.UpdateNodePoolSettingByKeyRequest{Enabled: *clusterservice.NewNullableBool(&enabled)}
		httpResp, err := client.Cluster.SettingsAPI.UpdateNodePoolSettingByKey(ctx, clusterservice.NodePoolSettingsKeyEnum(key)).UpdateNodePoolSettingByKeyRequest(body).Execute()
		if err != nil {
			return fmt.Errorf("update node-pool setting %q: %w", key, runaimcp.MapHTTPError(httpResp, err))
		}
		return nil
	case catWorkloadsManager:
		body, err := bffWorkloadsUpdateBody(key, enabled)
		if err != nil {
			return err
		}
		httpResp, err := client.BFF.SettingsAPI.UpdateWorkloadsSettingByKey(ctx, bff.WorkloadsSettingsKeyEnum(key)).UpdateWorkloadsSettingByKeyRequest(body).Execute()
		if err != nil {
			return fmt.Errorf("update workloads-manager setting %q: %w", key, runaimcp.MapHTTPError(httpResp, err))
		}
		return nil
	}
	return fmt.Errorf("unknown settings category %q: %w", category, runaimcp.ErrValidation)
}

// settingValueFromBFFMap extracts value.enabled plus the displayName/description/stage metadata
// from the BFF's untyped by-key response (GetWorkloadsSettingByKey returns map[string]interface{}).
func settingValueFromBFFMap(resp map[string]interface{}) (settingValue, error) {
	enabled, err := enabledFromBFFMap(resp)
	if err != nil {
		return settingValue{}, err
	}
	return settingValue{
		enabled:     enabled,
		displayName: stringFromMap(resp, "displayName"),
		description: stringFromMap(resp, "description"),
		stage:       stringFromMap(resp, "stage"),
	}, nil
}

func stringFromMap(m map[string]interface{}, key string) *string {
	v, ok := m[key].(string)
	if !ok {
		return nil
	}
	return &v
}

// enabledFromBFFMap extracts value.enabled from the BFF's untyped by-key response. A missing
// value means unset (nil); a present-but-wrong-shaped value is an upstream error, not silently
// "unset".
func enabledFromBFFMap(resp map[string]interface{}) (*bool, error) {
	raw, present := resp["value"]
	if !present {
		return nil, nil
	}
	value, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected workloads-manager settings response shape: %w", runaimcp.ErrUpstream)
	}
	enabledRaw, present := value["enabled"]
	if !present {
		return nil, nil
	}
	enabled, ok := enabledRaw.(bool)
	if !ok {
		return nil, fmt.Errorf("unexpected workloads-manager settings response shape: %w", runaimcp.ErrUpstream)
	}
	return &enabled, nil
}

// bffWorkloadsUpdateBody builds the BFF's polymorphic update body for a workloads-manager key.
func bffWorkloadsUpdateBody(key string, enabled bool) (bff.UpdateWorkloadsSettingByKeyRequest, error) {
	switch key {
	case "enableNIM":
		return bff.UpdateWorkloadsSettingByKeyRequest{
			UpdateEnableNIM: &bff.UpdateEnableNIM{Enabled: *bff.NewNullableBool(&enabled)},
		}, nil
	case "deletedWorkloadsView":
		return bff.UpdateWorkloadsSettingByKeyRequest{
			UpdateDeletedWorkloadsView: &bff.UpdateDeletedWorkloadsView{Enabled: *bff.NewNullableBool(&enabled)},
		}, nil
	case "allowGPT2Model":
		return bff.UpdateWorkloadsSettingByKeyRequest{
			UpdateAllowGPT2Model: &bff.UpdateAllowGPT2Model{Enabled: *bff.NewNullableBool(&enabled)},
		}, nil
	case "enableModelCatalog":
		return bff.UpdateWorkloadsSettingByKeyRequest{
			UpdateEnableModelCatalog: &bff.UpdateEnableModelCatalog{Enabled: *bff.NewNullableBool(&enabled)},
		}, nil
	}
	return bff.UpdateWorkloadsSettingByKeyRequest{}, fmt.Errorf("unknown workloads-manager setting %q: %w", key, runaimcp.ErrValidation)
}
