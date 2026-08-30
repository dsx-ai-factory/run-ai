package tests

import (
	"fmt"
	"net/http"
	"testing"

	openapi "github.com/run-ai/runai/control-plane/bff-service/pkg/client/generated"
	"github.com/stretchr/testify/require"
)

func TestGetWorkloadsSettings(t *testing.T) {
	t.Parallel()
	ctx := newTestContextWithUniqueTenant()

	settings, httpResp, err := OpenapiClient.SettingsAPI.GetWorkloadsSettings(ctx).Execute()

	require.NoError(t, err, "Should not return an error")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status")
	require.NotNil(t, settings, "Settings response should not be nil")

	require.NotNil(t, settings.EnableNim, "EnableNim should not be nil")
	require.NotNil(t, settings.EnableNim.Value, "EnableNim.Value should not be nil")
	require.NotNil(t, settings.EnableNim.Value.Enabled, "EnableNim.Value.Enabled should not be nil")

	require.NotNil(t, settings.DeletedWorkloadsView, "DeletedWorkloadsView should not be nil")
	require.NotNil(t, settings.DeletedWorkloadsView.Value, "DeletedWorkloadsView.Value should not be nil")
	require.NotNil(t, settings.DeletedWorkloadsView.Value.Enabled, "DeletedWorkloadsView.Value.Enabled should not be nil")

	require.NotNil(t, settings.EnableModelCatalog, "EnableModelCatalog should not be nil")
	require.NotNil(t, settings.EnableModelCatalog.Value, "EnableModelCatalog.Value should not be nil")
	require.NotNil(t, settings.EnableModelCatalog.Value.Enabled, "EnableModelCatalog.Value.Enabled should not be nil")
}

func TestGetWorkloadsSettingByKey(t *testing.T) {
	tests := []struct {
		name       string
		settingKey openapi.WorkloadsSettingsKeyEnum
	}{
		{
			name:       "enableNIM setting",
			settingKey: openapi.WORKLOADSSETTINGSKEYENUM_ENABLE_NIM,
		},
		{
			name:       "deletedWorkloadsView setting",
			settingKey: openapi.WORKLOADSSETTINGSKEYENUM_DELETED_WORKLOADS_VIEW,
		},
		{
			name:       "allowGPT2Model setting",
			settingKey: openapi.WORKLOADSSETTINGSKEYENUM_ALLOW_GPT2_MODEL,
		},
		{
			name:       "enableModelCatalog setting",
			settingKey: openapi.WORKLOADSSETTINGSKEYENUM_ENABLE_MODEL_CATALOG,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := newTestContextWithUniqueTenant()

			setting, httpResp, err := OpenapiClient.SettingsAPI.GetWorkloadsSettingByKey(ctx, tt.settingKey).Execute()

			require.NoError(t, err, "Should not return an error")
			require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status")
			require.NotNil(t, setting, "Setting should not be nil")
			require.Contains(t, setting, "value", "Setting should contain 'value' field")
		})
	}
}

func TestUpdateWorkloadsSettingByKey(t *testing.T) {
	tests := []struct {
		name        string
		settingKey  openapi.WorkloadsSettingsKeyEnum
		enabled     bool
		prepareBody func(bool) openapi.UpdateWorkloadsSettingByKeyRequest
	}{
		{
			name:       "enableNIM setting",
			settingKey: openapi.WORKLOADSSETTINGSKEYENUM_ENABLE_NIM,
			enabled:    true,
			prepareBody: func(enabled bool) openapi.UpdateWorkloadsSettingByKeyRequest {
				update := openapi.NewUpdateEnableNIM()
				update.SetEnabled(enabled)
				return openapi.UpdateWorkloadsSettingByKeyRequest{
					UpdateEnableNIM: update,
				}
			},
		},
		{
			name:       "deletedWorkloadsView setting",
			settingKey: openapi.WORKLOADSSETTINGSKEYENUM_DELETED_WORKLOADS_VIEW,
			enabled:    false,
			prepareBody: func(enabled bool) openapi.UpdateWorkloadsSettingByKeyRequest {
				update := openapi.NewUpdateDeletedWorkloadsView()
				update.SetEnabled(enabled)
				return openapi.UpdateWorkloadsSettingByKeyRequest{
					UpdateDeletedWorkloadsView: update,
				}
			},
		},
		{
			name:       "allowGPT2Model setting",
			settingKey: openapi.WORKLOADSSETTINGSKEYENUM_ALLOW_GPT2_MODEL,
			enabled:    true,
			prepareBody: func(enabled bool) openapi.UpdateWorkloadsSettingByKeyRequest {
				update := openapi.NewUpdateAllowGPT2Model()
				update.SetEnabled(enabled)
				return openapi.UpdateWorkloadsSettingByKeyRequest{
					UpdateAllowGPT2Model: update,
				}
			},
		},
		{
			name:       "enableModelCatalog setting",
			settingKey: openapi.WORKLOADSSETTINGSKEYENUM_ENABLE_MODEL_CATALOG,
			enabled:    true,
			prepareBody: func(enabled bool) openapi.UpdateWorkloadsSettingByKeyRequest {
				update := openapi.NewUpdateEnableModelCatalog()
				update.SetEnabled(enabled)
				return openapi.UpdateWorkloadsSettingByKeyRequest{
					UpdateEnableModelCatalog: update,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := newTestContextWithUniqueTenant()

			updateBody := tt.prepareBody(tt.enabled)

			httpResp, err := OpenapiClient.SettingsAPI.UpdateWorkloadsSettingByKey(ctx, tt.settingKey).
				UpdateWorkloadsSettingByKeyRequest(updateBody).
				Execute()

			require.NoError(t, err, "Should not return an error on UPDATE")
			require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status for UPDATE")

			setting, httpResp, err := OpenapiClient.SettingsAPI.GetWorkloadsSettingByKey(ctx, tt.settingKey).Execute()

			require.NoError(t, err, "Should not return an error on GET after UPDATE")
			require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status for GET")
			require.Contains(t, setting, "value", "Setting should contain 'value' field")

			value := setting["value"].(map[string]interface{})
			require.Contains(t, value, FieldEnabled, "Value should contain 'enabled' field")
			require.Equal(t, tt.enabled, value[FieldEnabled], "Setting should match expected value after update")
		})
	}
}

func TestGetWorkloadsSettings_NonManagementTenantParam_Returns403(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/api/v1/ui-workloads/settings?tenantId=%s", hostUrl, BaseTenantId), nil)
	require.NoError(t, err)
	req.Header.Set("X-Runai-Subject-Id", "test@run.ai")
	// no X-Runai-Is-Management-Tenant header → not a management tenant

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusForbidden, resp.StatusCode, "Should return 403 when tenantId query param is set by a non-management tenant")
}
