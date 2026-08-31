/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"

	openapi "github.com/run-ai/runai/control-plane/bff-service/pkg/client/generated"
	"github.com/stretchr/testify/require"
)

// redCircleSVGDataURL returns a data URL for a small red circle SVG,
// used as a distinct logo value in mutation tests.
func redCircleSVGDataURL() string {
	const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><circle cx="5" cy="5" r="5" fill="#ff0000"/></svg>`
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}

func TestGetBrandingSettings(t *testing.T) {
	t.Parallel()
	ctx := newTestContextWithUniqueTenant()

	settings, httpResp, err := OpenapiClient.SettingsAPI.GetBrandingSettings(ctx).Execute()

	require.NoError(t, err, "Should not return an error")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status")
	require.NotNil(t, settings, "Settings response should not be nil")

	require.True(t, settings.HasLogo(), "Logo should be present in branding settings")
	logo := settings.GetLogo()
	logoValue := logo.GetValue()
	require.Empty(t, logoValue.GetLogo(), "Logo value should be empty (default value is empty string)")
}

func TestGetBrandingSettingByKey(t *testing.T) {
	t.Parallel()
	ctx := newTestContextWithUniqueTenant()

	setting, httpResp, err := OpenapiClient.SettingsAPI.
		GetBrandingSettingByKey(ctx, openapi.BRANDINGSETTINGSKEYENUM_LOGO).
		Execute()

	require.NoError(t, err, "Should not return an error")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status")
	require.NotNil(t, setting, "Setting should not be nil")
	require.Contains(t, setting, "value", "Setting should contain 'value' field")
}

func TestUpdateBrandingSettingByKey_Logo(t *testing.T) {
	t.Parallel()
	ctx := newTestContextWithUniqueTenant()

	initialSettings, httpResp, err := OpenapiClient.SettingsAPI.GetBrandingSettings(ctx).Execute()
	require.NoError(t, err, "Should not return an error on initial GET")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK on initial GET")
	require.True(t, initialSettings.HasLogo(), "Logo should be present before update")
	initialLogo := initialSettings.GetLogo()
	initialLogoValue := initialLogo.GetValue()
	previousLogoValue := initialLogoValue.GetLogo()

	updateResp, err := OpenapiClient.SettingsAPI.
		UpdateBrandingSettingByKey(ctx, openapi.BRANDINGSETTINGSKEYENUM_LOGO).
		UpdateLogo(*openapi.NewUpdateLogo(redCircleSVGDataURL())).
		Execute()

	require.NoError(t, err, "Should not return an error on UPDATE")
	require.Equal(t, http.StatusOK, updateResp.StatusCode, "Should return 200 OK status for UPDATE")

	updatedSettings, httpResp, err := OpenapiClient.SettingsAPI.GetBrandingSettings(ctx).Execute()
	require.NoError(t, err, "Should not return an error on GET after UPDATE")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status for GET after UPDATE")
	require.True(t, updatedSettings.HasLogo(), "Logo should be present after update")
	updatedLogo := updatedSettings.GetLogo()
	updatedLogoValue := updatedLogo.GetValue()
	newLogoValue := updatedLogoValue.GetLogo()

	require.NotEmpty(t, newLogoValue, "Logo value should not be empty after update")
	require.NotEqual(t, previousLogoValue, newLogoValue, "Logo value should have changed after update")
}

func TestGetBrandingSettings_NonManagementTenantParam_Returns403(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/api/v1/branding/settings?tenantId=%s", hostUrl, BaseTenantId), nil)
	require.NoError(t, err)
	req.Header.Set("X-Runai-Subject-Id", "test@run.ai")
	// no X-Runai-Is-Management-Tenant header → not a management tenant

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusForbidden, resp.StatusCode, "Should return 403 when tenantId query param is set by a non-management tenant")
}

func TestUpdateBrandingSettingByKey_InvalidBase64_Returns400(t *testing.T) {
	t.Parallel()
	ctx := newTestContextWithUniqueTenant()

	httpResp, err := OpenapiClient.SettingsAPI.
		UpdateBrandingSettingByKey(ctx, openapi.BRANDINGSETTINGSKEYENUM_LOGO).
		UpdateLogo(*openapi.NewUpdateLogo("not-valid-base64!!!")).
		Execute()

	require.Error(t, err, "Should return an error for invalid base64")
	require.NotNil(t, httpResp, "HTTP response should not be nil on error")
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode, "Should return 400 for invalid base64")
}

func TestUpdateBrandingSettingByKey_UnsupportedMimeType_Returns415(t *testing.T) {
	t.Parallel()
	ctx := newTestContextWithUniqueTenant()

	httpResp, err := OpenapiClient.SettingsAPI.
		UpdateBrandingSettingByKey(ctx, openapi.BRANDINGSETTINGSKEYENUM_LOGO).
		UpdateLogo(*openapi.NewUpdateLogo(JPEG_DATA_URL)).
		Execute()

	require.Error(t, err, "Should return an error for unsupported MIME type")
	require.NotNil(t, httpResp, "HTTP response should not be nil on error")
	require.Equal(t, http.StatusUnsupportedMediaType, httpResp.StatusCode, "Should return 415 for unsupported MIME type")
}
