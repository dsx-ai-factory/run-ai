package settings

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

const (
	authzSettings      = "/api/v1/authorization/settings/"
	wlSettings         = "/api/v1/workloads/settings/"
	clusterSetting     = "/api/v1/clusters/settings/"
	orgUnitSetting     = "/api/v1/org-unit/settings/"
	catalogSetting     = "/api/v1/catalog/settings/"
	dataVolumesSetting = "/api/v1/datavolumes/settings/"
	nodePoolSetting    = "/api/v1/node-pools/settings/"
	bffWlSetting       = "/api/v1/ui-workloads/settings/"

	// get-all endpoints (no trailing slash, no key) — used by list_system_settings.
	wlSettingsAll         = "/api/v1/workloads/settings"
	authzSettingsAll      = "/api/v1/authorization/settings"
	clusterSettingAll     = "/api/v1/clusters/settings"
	orgUnitSettingAll     = "/api/v1/org-unit/settings"
	catalogSettingAll     = "/api/v1/catalog/settings"
	dataVolumesSettingAll = "/api/v1/datavolumes/settings"
	nodePoolSettingAll    = "/api/v1/node-pools/settings"
	bffWlSettingAll       = "/api/v1/ui-workloads/settings"
)

// settingsGetAll serves a category's get-all response: an object keyed by each setting key, each
// value flat ({enabled,…}) or value-wrapped ({value:{enabled},…}). meta adds the
// displayName/description/stage that metadata-bearing services return (authorization returns none).
func settingsGetAll(wrap, enabled, meta bool, keys ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		obj := map[string]any{}
		for _, k := range keys {
			entry := map[string]any{}
			if wrap {
				entry["value"] = map[string]any{"enabled": enabled}
			} else {
				entry["enabled"] = enabled
			}
			if meta {
				entry["displayName"] = "Display"
				entry["description"] = "Desc"
				entry["stage"] = "GA"
			}
			obj[k] = entry
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, obj)
	}
}

var (
	permScopesPath    = runaiclienttest.PermScopesPath
	permTenantHandler = runaiclienttest.PermTenantHandler
	bptr              = runaiclienttest.BoolPtr
)

// statefulBool serves a boolean setting whose GET reflects the last PUT, so tests exercise the
// re-read-after-apply path. wrap=true nests the value as {"value":{"enabled":…}} (workloads /
// catalog / datavolumes / BFF shape); wrap=false is the flat {"enabled":…} shape. A nil initial
// value 404s on GET until the first PUT (never-set semantics).
func statefulBool(t *testing.T, wrap bool, initial *bool) (http.HandlerFunc, *map[string]any) {
	t.Helper()
	cur := initial
	var putBody map[string]any
	h := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if cur == nil {
				http.Error(w, "not set", http.StatusNotFound)
				return
			}
			body := map[string]any{"enabled": *cur}
			if wrap {
				body = map[string]any{"value": map[string]any{"enabled": *cur}}
			}
			runaiclienttest.WriteJSON(w, http.StatusOK, body)
			return
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &putBody)
		if v, ok := putBody["enabled"].(bool); ok {
			cur = &v
		}
		w.WriteHeader(http.StatusOK)
	}
	return h, &putBody
}

func TestUpdateSystemSettingsInput_Validate(t *testing.T) {
	cases := []struct {
		name    string
		in      updateSystemSettingsInput
		wantErr string
	}{
		{name: "happy", in: updateSystemSettingsInput{Category: "workloads", Key: "enablePolicies", Enabled: bptr(true)}},
		{name: "happy_workloads_manager", in: updateSystemSettingsInput{Category: "workloads-manager", Key: "enableNIM", Enabled: bptr(false)}},
		{name: "happy_deleted_workloads_view", in: updateSystemSettingsInput{Category: "workloads-manager", Key: "deletedWorkloadsView", Enabled: bptr(true)}},
		{name: "happy_node_pool", in: updateSystemSettingsInput{Category: "node-pool", Key: "cpuResourcesQuota", Enabled: bptr(true)}},
		{name: "happy_catalog", in: updateSystemSettingsInput{Category: "catalog", Key: "enableNgcIntegration", Enabled: bptr(true)}},
		{name: "bad_category", in: updateSystemSettingsInput{Category: "nope", Key: "x", Enabled: bptr(true)}, wantErr: "category must be one of"},
		{name: "key_not_in_category", in: updateSystemSettingsInput{Category: "workloads", Key: "enableScopedAuthorization", Enabled: bptr(true)}, wantErr: "not a workloads setting"},
		{name: "key_wrong_service", in: updateSystemSettingsInput{Category: "catalog", Key: "enableDataVolumes", Enabled: bptr(true)}, wantErr: "not a catalog setting"},
		{name: "read_only_key", in: updateSystemSettingsInput{Category: "authorization", Key: "tenantIsCspManaged", Enabled: bptr(true)}, wantErr: "read-only"},
		{name: "missing_enabled", in: updateSystemSettingsInput{Category: "workloads", Key: "enablePolicies"}, wantErr: "enabled is required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.Validate()
			if tc.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestUpdateSystemSettings_Apply(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("update"))
	h, putBody := statefulBool(t, false, bptr(false))
	mux.HandleFunc(authzSettings, h)

	out, err := updateSystemSettingsRun(context.Background(), client, updateSystemSettingsInput{
		Category: "authorization", Key: "enableScopedAuthorization", Enabled: bptr(true), Confirm: true,
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, false, out.Before["enabled"])
	assert.Equal(t, true, out.After["enabled"]) // re-read reflects the applied value
	assert.Equal(t, true, (*putBody)["enabled"])
}

func TestUpdateSystemSettings_DryRun(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("update"))
	mux.HandleFunc(authzSettings, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatal("dry-run must not mutate")
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"enabled": false})
	})

	out, err := updateSystemSettingsRun(context.Background(), client, updateSystemSettingsInput{
		Category: "authorization", Key: "enableScopedAuthorization", Enabled: bptr(true), DryRun: true,
	})
	require.NoError(t, err)
	assert.False(t, out.Applied)
	assert.Equal(t, false, out.Before["enabled"])
	assert.Equal(t, true, out.After["enabled"]) // preview of the intended value
}

func TestUpdateSystemSettings_ConfirmGate(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("update"))
	mux.HandleFunc(authzSettings, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method) // no mutation without confirm
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"enabled": false})
	})

	_, err := updateSystemSettingsRun(context.Background(), client, updateSystemSettingsInput{
		Category: "authorization", Key: "enableScopedAuthorization", Enabled: bptr(true),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.Contains(t, err.Error(), "update of") // not "delete of"
}

func TestUpdateSystemSettings_Forbidden(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler()) // grant nothing

	_, err := updateSystemSettingsRun(context.Background(), client, updateSystemSettingsInput{
		Category: "authorization", Key: "enableScopedAuthorization", Enabled: bptr(true), Confirm: true,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}

func TestUpdateSystemSettings_ApplyWorkloadsManager(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("update"))
	h, putBody := statefulBool(t, true, bptr(false)) // BFF nests value.enabled
	mux.HandleFunc(bffWlSetting, h)

	out, err := updateSystemSettingsRun(context.Background(), client, updateSystemSettingsInput{
		Category: "workloads-manager", Key: "enableNIM", Enabled: bptr(true), Confirm: true,
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, false, out.Before["enabled"])
	assert.Equal(t, true, out.After["enabled"])
	assert.Equal(t, true, (*putBody)["enabled"]) // polymorphic UpdateEnableNIM marshals to {enabled}
}

func TestUpdateSystemSettings_ApplyNodePool(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("update"))
	h, putBody := statefulBool(t, false, bptr(false))
	mux.HandleFunc(nodePoolSetting, h)

	out, err := updateSystemSettingsRun(context.Background(), client, updateSystemSettingsInput{
		Category: "node-pool", Key: "cpuResourcesQuota", Enabled: bptr(true), Confirm: true,
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, true, out.After["enabled"])
	assert.Equal(t, true, (*putBody)["enabled"])
}

func TestUpdateSystemSettings_SetWhenUnset(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("update"))
	h, putBody := statefulBool(t, false, nil) // never-set → GET 404 until first PUT
	mux.HandleFunc(orgUnitSetting, h)

	out, err := updateSystemSettingsRun(context.Background(), client, updateSystemSettingsInput{
		Category: "org-unit", Key: "limitQuotaOverSubscription", Enabled: bptr(true), Confirm: true,
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Nil(t, out.Before["enabled"]) // was unset
	assert.Equal(t, true, out.After["enabled"])
	assert.Equal(t, true, (*putBody)["enabled"])
}

// listAllHandlers registers a get-all handler for every category, so a full list_system_settings
// call succeeds. Shapes mirror the real services: workloads/catalog/datavolumes/BFF value-wrap and
// carry metadata; cluster/org-unit/node-pool are flat with metadata; authorization is flat with
// none. The BFF spells enableNIM as "enableNim".
func listAllHandlers(mux *http.ServeMux) {
	mux.HandleFunc(wlSettingsAll, settingsGetAll(true, true, true,
		"enablePolicies", "workloadSubmissionV2Enabled", "enableTemplatesV2", "enableHybridSubmission", "gpuOverProvisioning", "applicationSubmission"))
	mux.HandleFunc(authzSettingsAll, settingsGetAll(false, false, false, "enableScopedAuthorization", "tenantIsCspManaged"))
	mux.HandleFunc(clusterSettingAll, settingsGetAll(false, true, true, "enableAdvancedGpuMetrics"))
	mux.HandleFunc(orgUnitSettingAll, settingsGetAll(false, false, true, "limitQuotaOverSubscription"))
	mux.HandleFunc(catalogSettingAll, settingsGetAll(true, true, true, "enableNgcIntegration"))
	mux.HandleFunc(dataVolumesSettingAll, settingsGetAll(true, false, true, "enableDataVolumes"))
	mux.HandleFunc(nodePoolSettingAll, settingsGetAll(false, true, true, "projectManualOverQuota", "cpuResourcesQuota"))
	mux.HandleFunc(bffWlSettingAll, settingsGetAll(true, false, true, "enableNim", "deletedWorkloadsView", "allowGPT2Model", "enableModelCatalog"))
}

func TestListSystemSettings(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("read"))
	listAllHandlers(mux)

	out, err := listSystemSettingsRun(context.Background(), client, listSystemSettingsInput{})
	require.NoError(t, err)
	assert.Empty(t, out.Warnings)
	// 6 workloads + 2 authorization + 1 cluster + 1 org-unit + 1 catalog + 1 datavolumes + 2 node-pool + 4 workloads-manager
	assert.Len(t, out.Settings, 18)
	byKey := map[string]systemSetting{}
	for _, s := range out.Settings {
		byKey[s.Key] = s
	}
	require.NotNil(t, byKey["enablePolicies"].Enabled)
	assert.True(t, *byKey["enablePolicies"].Enabled) // workloads value-wrapper unwrapped
	require.NotNil(t, byKey["enableNgcIntegration"].Enabled)
	assert.True(t, *byKey["enableNgcIntegration"].Enabled) // catalog value-wrapper unwrapped
	require.NotNil(t, byKey["cpuResourcesQuota"].Enabled)
	assert.True(t, *byKey["cpuResourcesQuota"].Enabled) // node-pool direct enabled
	require.NotNil(t, byKey["enableNIM"].Enabled)
	assert.False(t, *byKey["enableNIM"].Enabled) // BFF untyped-map value unwrapped
	// deletedWorkloadsView is now surfaced alongside the other workloads-manager keys
	_, hasDeletedView := byKey["deletedWorkloadsView"]
	assert.True(t, hasDeletedView)
	// tenantIsCspManaged is listable even though it is not updatable
	_, listed := byKey["tenantIsCspManaged"]
	assert.True(t, listed)

	// metadata surfaces for wrapped (workloads) and flat (cluster) shapes alike, now that the
	// get-all endpoint populates it — including cluster, which the by-key endpoint left bare.
	require.NotNil(t, byKey["enablePolicies"].Stage)
	assert.Equal(t, "GA", *byKey["enablePolicies"].Stage)
	require.NotNil(t, byKey["enablePolicies"].DisplayName)
	assert.Equal(t, "Display", *byKey["enablePolicies"].DisplayName)
	require.NotNil(t, byKey["enableAdvancedGpuMetrics"].DisplayName)
	assert.Equal(t, "Display", *byKey["enableAdvancedGpuMetrics"].DisplayName)
	require.NotNil(t, byKey["enableNIM"].DisplayName) // BFF enableNim alias resolved
	// ...and is absent for authorization, whose response carries only enabled
	assert.Nil(t, byKey["enableScopedAuthorization"].Stage)
	assert.Nil(t, byKey["enableScopedAuthorization"].DisplayName)
}

func TestListSystemSettings_UnsetReturnsNull(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("read"))
	mux.HandleFunc(orgUnitSettingAll, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not set", http.StatusNotFound)
	})

	out, err := listSystemSettingsRun(context.Background(), client, listSystemSettingsInput{Category: "org-unit"})
	require.NoError(t, err) // a 404 (never-set) must not abort the listing
	require.Len(t, out.Settings, 1)
	assert.Nil(t, out.Settings[0].Enabled)
	assert.Empty(t, out.Warnings)
}

func TestListSystemSettings_CategoryFilter(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("read"))
	mux.HandleFunc(clusterSettingAll, settingsGetAll(false, true, true, "enableAdvancedGpuMetrics"))

	out, err := listSystemSettingsRun(context.Background(), client, listSystemSettingsInput{Category: "cluster"})
	require.NoError(t, err)
	assert.Len(t, out.Settings, 1)
	assert.Equal(t, "enableAdvancedGpuMetrics", out.Settings[0].Key)
	require.NotNil(t, out.Settings[0].DisplayName)
}

func TestListSystemSettings_Forbidden(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler()) // no read grant

	_, err := listSystemSettingsRun(context.Background(), client, listSystemSettingsInput{})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}

func TestListSystemSettings_DegradesOnError(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("read"))
	// cluster service is down (500): its category degrades to a warning. The other categories are
	// unregistered → 404 → folded to unset (null) rows, so the listing still returns them.
	mux.HandleFunc(clusterSettingAll, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})

	out, err := listSystemSettingsRun(context.Background(), client, listSystemSettingsInput{})
	require.NoError(t, err)
	require.Len(t, out.Warnings, 1)
	assert.Contains(t, out.Warnings[0], "cluster")
	for _, s := range out.Settings {
		assert.NotEqual(t, "cluster", s.Category) // the failed category is skipped
	}
	assert.NotEmpty(t, out.Settings) // other categories still present
}

func TestSettingsToolsLint(t *testing.T) {
	errs := runaimcp.Lint(func(r *runaimcp.Registry) {
		registerListSystemSettings(r, nil)
		registerUpdateSystemSettings(r, nil)
	})
	assert.Empty(t, errs)
}
