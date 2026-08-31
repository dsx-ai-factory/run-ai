/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	openapi "github.com/run-ai/runai/control-plane/bff-service/pkg/client/generated"
	"github.com/stretchr/testify/require"
)

func TestGetAppearanceSettings(t *testing.T) {
	t.Parallel()
	ctx := newTestContextWithUniqueTenant()

	settings, httpResp, err := OpenapiClient.SettingsAPI.GetAppearanceSettings(ctx).Execute()

	require.NoError(t, err, "Should not return an error")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status")
	require.NotNil(t, settings, "Settings response should not be nil")

	require.NotNil(t, settings.BannerConfig, "BannerConfig should be present in appearance settings")
}

func TestGetAppearanceSettingByKey(t *testing.T) {
	t.Parallel()
	ctx := newTestContextWithUniqueTenant()

	setting, httpResp, err := OpenapiClient.SettingsAPI.
		GetAppearanceSettingByKey(ctx, openapi.APPEARANCESETTINGSKEYENUM_BANNER_CONFIG).
		Execute()

	require.NoError(t, err, "Should not return an error")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status")
	require.NotNil(t, setting, "Setting should not be nil")
	require.Contains(t, setting, "value", "Setting should contain 'value' field")
}

func TestUpdateAppearanceSettingByKey_BannerConfig(t *testing.T) {
	t.Parallel()
	ctx := newTestContextWithUniqueTenant()

	enabled := true
	updateBody := openapi.NewUpdateBannerConfig()
	updateBody.SetEnabled(enabled)
	updateBody.SetText("INTEGRATION TEST BANNER")
	updateBody.SetBackgroundColor("#D32F2F")
	updateBody.SetFontColor("#FFFFFF")

	updateResp, err := OpenapiClient.SettingsAPI.
		UpdateAppearanceSettingByKey(ctx, openapi.APPEARANCESETTINGSKEYENUM_BANNER_CONFIG).
		UpdateBannerConfig(*updateBody).
		Execute()

	require.NoError(t, err, "Should not return an error on UPDATE")
	require.Equal(t, http.StatusOK, updateResp.StatusCode, "Should return 200 OK status for UPDATE")

	setting, httpResp, err := OpenapiClient.SettingsAPI.
		GetAppearanceSettingByKey(ctx, openapi.APPEARANCESETTINGSKEYENUM_BANNER_CONFIG).
		Execute()

	require.NoError(t, err, "Should not return an error on GET after UPDATE")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status for GET after UPDATE")
	require.NotNil(t, setting, "Setting should not be nil after update")

	value, ok := setting["value"].(map[string]interface{})
	require.True(t, ok, "Setting value should be a map")
	require.Equal(t, enabled, value["enabled"], "Enabled should match after update")
	require.Equal(t, "INTEGRATION TEST BANNER", value["text"], "Text should match after update")
	require.Equal(t, "#D32F2F", value["backgroundColor"], "BackgroundColor should match after update")
	require.Equal(t, "#FFFFFF", value["fontColor"], "FontColor should match after update")
}

func TestGetAppearanceSettings_NonManagementTenantParam_Returns403(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/api/v1/ui-appearance/settings?tenantId=%s", hostUrl, BaseTenantId), nil)
	require.NoError(t, err)
	req.Header.Set("X-Runai-Subject-Id", "test@run.ai")
	// no X-Runai-Is-Management-Tenant header → not a management tenant

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusForbidden, resp.StatusCode, "Should return 403 when tenantId query param is set by a non-management tenant")
}

func TestUpdateAppearanceSettingByKey_EnabledWithoutText_Returns400(t *testing.T) {
	t.Parallel()
	ctx := newTestContextWithUniqueTenant()

	updateBody := openapi.NewUpdateBannerConfig()
	updateBody.SetEnabled(true)
	// no text set — server requires text when enabled

	httpResp, err := OpenapiClient.SettingsAPI.
		UpdateAppearanceSettingByKey(ctx, openapi.APPEARANCESETTINGSKEYENUM_BANNER_CONFIG).
		UpdateBannerConfig(*updateBody).
		Execute()

	require.Error(t, err, "Should return an error when enabling banner without text")
	require.NotNil(t, httpResp, "HTTP response should not be nil on error")
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode, "Should return 400 when enabling banner without text")
}

func TestGetAppearanceSettingByKey_UnknownKey_Returns400(t *testing.T) {
	t.Parallel()
	tenantID := uuid.NewString()

	req, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/api/v1/ui-appearance/settings/unknownKey", hostUrl), nil)
	require.NoError(t, err)
	req.Header.Set("X-Runai-Tenant-Id", tenantID)
	req.Header.Set("X-Runai-Subject-Id", "test@run.ai")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should return 400 for unknown appearance setting key")
}

func TestUpdateAppearanceSettingByKey_DisableBanner(t *testing.T) {
	t.Parallel()
	ctx := newTestContextWithUniqueTenant()

	updateBody := openapi.NewUpdateBannerConfig()
	updateBody.SetEnabled(false)

	updateResp, err := OpenapiClient.SettingsAPI.
		UpdateAppearanceSettingByKey(ctx, openapi.APPEARANCESETTINGSKEYENUM_BANNER_CONFIG).
		UpdateBannerConfig(*updateBody).
		Execute()

	require.NoError(t, err, "Should not return an error when disabling banner")
	require.Equal(t, http.StatusOK, updateResp.StatusCode, "Should return 200 OK when disabling banner")

	setting, httpResp, err := OpenapiClient.SettingsAPI.
		GetAppearanceSettingByKey(ctx, openapi.APPEARANCESETTINGSKEYENUM_BANNER_CONFIG).
		Execute()

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)

	value, ok := setting["value"].(map[string]interface{})
	require.True(t, ok, "Setting value should be a map")
	require.Equal(t, false, value["enabled"], "Banner should be disabled after update")
}
