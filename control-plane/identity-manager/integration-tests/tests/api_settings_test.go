/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"net/http"
	"testing"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

const (
	SettingsKeyAutoRedirectSSO       = "autoRedirectSSO"
	SettingsKeyBrowserSessionTimeout = "browserSessionTimeout"
	SettingsKeyLogoutRedirectUri     = "logoutRedirectUri"
	DefaultLogoutUri                 = "https://www.example.com/logout"

	TestUserName     = "test@example.com"
	FieldEnabled     = "enabled"
	FieldUri         = "uri"
	FieldIdleTimeout = "idleTimeout"
)

func TestGetSecuritySettings(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	settings, httpResp, err := OpenapiClient.SettingsAPI.GetSecuritySettings(ctx).Execute()

	require.NoError(t, err, "Should not return an error")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status")
	require.NotNil(t, settings, "Settings response should not be nil")

	require.True(t, settings.HasAutoRedirectSSO(), "AutoRedirectSSO should be present")
	sso := settings.GetAutoRedirectSSO()
	require.True(t, sso.Value.Enabled.IsSet(), "AutoRedirectSSO.Enabled should be set")
	require.True(t, sso.Value.IdpAlias.IsSet(), "AutoRedirectSSO.IdpAlias should be set")

	require.True(t, settings.HasBrowserSessionTimeout(), "BrowserSessionTimeout should be present")
	timeout := settings.GetBrowserSessionTimeout()
	require.True(t, timeout.Value.IdleTimeout.IsSet(), "BrowserSessionTimeout.IdleTimeout should be set")

	require.True(t, settings.HasLogoutRedirectUri(), "LogoutRedirectUri should be present")
	logout := settings.GetLogoutRedirectUri()
	require.True(t, logout.Value.Uri.IsSet(), "LogoutRedirectUri.Uri should be set")
}

func TestGetSecuritySettingByKey_AutoRedirectSso(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	setting, httpResp, err := OpenapiClient.SettingsAPI.GetSecuritySettingByKey(ctx, SettingsKeyAutoRedirectSSO).Execute()

	require.NoError(t, err, "Should not return an error")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status for auto-redirect-sso")
	valueMap, ok := setting["value"].(map[string]interface{})
	require.True(t, ok, "setting should contain 'value'")
	require.Contains(t, valueMap, FieldEnabled)
}

func TestUpdateSecuritySettingByKey_AutoRedirectSso_HappyFlow_Enabled(t *testing.T) {
	t.Parallel()
	enabled := true
	idpAlias := "oidc"
	updateRequest := openapi.UpdateSecuritySettingByKeyRequest{
		UpdateAutoRedirectSsoSettingValue: &openapi.UpdateAutoRedirectSsoSettingValue{
			Enabled:  *openapi.NewNullableBool(&enabled),
			IdpAlias: *openapi.NewNullableString(&idpAlias),
		},
	}
	setting, httpResp, err := updateSetting(SettingsKeyAutoRedirectSSO, updateRequest)
	require.NoError(t, err, "Should not return an error on UPDATE")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status for UPDATE")
	assertSettingField(t, setting, FieldEnabled, enabled)
	assertSettingField(t, setting, "idpAlias", idpAlias)
}

func TestUpdateSecuritySettingByKey_AutoRedirectSso_HappyFlow_Enabled_WithNilAlias(t *testing.T) {
	t.Parallel()
	enabled := true
	updateRequest := openapi.UpdateSecuritySettingByKeyRequest{
		UpdateAutoRedirectSsoSettingValue: &openapi.UpdateAutoRedirectSsoSettingValue{
			Enabled:  *openapi.NewNullableBool(&enabled),
			IdpAlias: *openapi.NewNullableString(nil),
		},
	}
	setting, httpResp, err := updateSetting(SettingsKeyAutoRedirectSSO, updateRequest)
	require.Error(t, err)
	require.Empty(t, setting)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
}

func TestUpdateSecuritySettingByKey_AutoRedirectSso_HappyFlow_Enabled_WithEmptyAlias(t *testing.T) {
	t.Parallel()
	enabled := true
	updateRequest := openapi.UpdateSecuritySettingByKeyRequest{
		UpdateAutoRedirectSsoSettingValue: &openapi.UpdateAutoRedirectSsoSettingValue{
			Enabled:  *openapi.NewNullableBool(&enabled),
			IdpAlias: *openapi.NewNullableString(ptr.To("")),
		},
	}
	setting, httpResp, err := updateSetting(SettingsKeyAutoRedirectSSO, updateRequest)
	require.Error(t, err)
	require.Empty(t, setting)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
}

func TestUpdateSecuritySettingByKey_AutoRedirectSso_HappyFlow_Enabled_WithInvalidAlias(t *testing.T) {
	t.Parallel()
	enabled := true
	updateRequest := openapi.UpdateSecuritySettingByKeyRequest{
		UpdateAutoRedirectSsoSettingValue: &openapi.UpdateAutoRedirectSsoSettingValue{
			Enabled:  *openapi.NewNullableBool(&enabled),
			IdpAlias: *openapi.NewNullableString(ptr.To("@@")),
		},
	}
	setting, httpResp, err := updateSetting(SettingsKeyAutoRedirectSSO, updateRequest)
	require.Error(t, err)
	require.Empty(t, setting)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
}

func TestUpdateSecuritySettingByKey_AutoRedirectSso_HappyFlow_Disabled_WithAlias(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("auto-redirect-sso-with-alias")

	enabled := false
	idpAlias := "oidc"
	updateRequest := openapi.UpdateSecuritySettingByKeyRequest{
		UpdateAutoRedirectSsoSettingValue: &openapi.UpdateAutoRedirectSsoSettingValue{
			Enabled:  *openapi.NewNullableBool(&enabled),
			IdpAlias: *openapi.NewNullableString(&idpAlias),
		},
	}
	setting, httpResp, err := updateSettingWithContext(ctx, SettingsKeyAutoRedirectSSO, updateRequest)
	require.NoError(t, err, "Should not return an error on UPDATE")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status for UPDATE")
	assertSettingField(t, setting, FieldEnabled, enabled)
	assertSettingField(t, setting, "idpAlias", idpAlias)
}

func TestUpdateSecuritySettingByKey_AutoRedirectSso_HappyFlow_Disabled_WithNilAlias(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("auto-redirect-sso-with-nil-alias")

	enabled := false
	updateRequest := openapi.UpdateSecuritySettingByKeyRequest{
		UpdateAutoRedirectSsoSettingValue: &openapi.UpdateAutoRedirectSsoSettingValue{
			Enabled:  *openapi.NewNullableBool(&enabled),
			IdpAlias: *openapi.NewNullableString(nil),
		},
	}
	setting, httpResp, err := updateSettingWithContext(ctx, SettingsKeyAutoRedirectSSO, updateRequest)
	require.NoError(t, err, "Should not return an error on UPDATE")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status for UPDATE")
	assertSettingField(t, setting, FieldEnabled, enabled)
}

func TestUpdateSecuritySettingByKey_AutoRedirectSso_HappyFlow_Disabled_WithEmptyAlias(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("auto-redirect-sso-with-empty-alias")

	enabled := false
	emptyStr := ""
	updateRequest := openapi.UpdateSecuritySettingByKeyRequest{
		UpdateAutoRedirectSsoSettingValue: &openapi.UpdateAutoRedirectSsoSettingValue{
			Enabled:  *openapi.NewNullableBool(&enabled),
			IdpAlias: *openapi.NewNullableString(&emptyStr),
		},
	}
	setting, httpResp, err := updateSettingWithContext(ctx, SettingsKeyAutoRedirectSSO, updateRequest)
	require.NoError(t, err, "Should not return an error on UPDATE")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status for UPDATE")
	assertSettingField(t, setting, FieldEnabled, enabled)
	assertSettingField(t, setting, "idpAlias", "")
}

func TestGetSecuritySettingByKey_BrowserSessionTimeout(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	setting, httpResp, err := OpenapiClient.SettingsAPI.GetSecuritySettingByKey(ctx, SettingsKeyBrowserSessionTimeout).Execute()

	require.NoError(t, err, "Should not return an error")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status for browser-session-timeout")
	valueMap, ok := setting["value"].(map[string]interface{})
	require.True(t, ok, "setting should contain 'value'")
	require.Contains(t, valueMap, FieldIdleTimeout)
}

func TestUpdateSecuritySettingByKey_BrowserSessionTimeout(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("browser-session-timeout-update")

	testTimeout := float32(1234)
	updateRequest := openapi.UpdateSecuritySettingByKeyRequest{
		UpdateBrowserSessionTimeoutSettingValue: &openapi.UpdateBrowserSessionTimeoutSettingValue{
			IdleTimeout: *openapi.NewNullableFloat32(&testTimeout),
		},
	}
	updateAndVerifySetting(ctx, t, SettingsKeyBrowserSessionTimeout, updateRequest, FieldIdleTimeout, float64(testTimeout))

	defaultTimeout := float32(3600)
	updateRequest = openapi.UpdateSecuritySettingByKeyRequest{
		UpdateBrowserSessionTimeoutSettingValue: &openapi.UpdateBrowserSessionTimeoutSettingValue{
			IdleTimeout: *openapi.NewNullableFloat32(&defaultTimeout),
		},
	}
	updateAndVerifySetting(ctx, t, SettingsKeyBrowserSessionTimeout, updateRequest, FieldIdleTimeout, float64(defaultTimeout))
}

func TestGetSecuritySettingByKey_LogoutRedirectUri(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	setting, httpResp, err := OpenapiClient.SettingsAPI.GetSecuritySettingByKey(ctx, SettingsKeyLogoutRedirectUri).Execute()

	require.NoError(t, err, "Should not return an error")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status for logout-redirect-uri")
	valueMap, ok := setting["value"].(map[string]interface{})
	require.True(t, ok, "setting should contain 'value'")
	require.Contains(t, valueMap, FieldUri)
}

func TestUpdateSecuritySettingByKey_LogoutRedirectUri(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("logout-redirect-uri-update")

	testUri := DefaultLogoutUri
	updateRequest := openapi.UpdateSecuritySettingByKeyRequest{
		UpdateLogoutRedirectUriSettingValue: &openapi.UpdateLogoutRedirectUriSettingValue{
			Uri: *openapi.NewNullableString(&testUri),
		},
	}
	updateAndVerifySetting(ctx, t, SettingsKeyLogoutRedirectUri, updateRequest, FieldUri, testUri)

	defaultUri := DefaultLogoutUri
	updateRequest = openapi.UpdateSecuritySettingByKeyRequest{
		UpdateLogoutRedirectUriSettingValue: &openapi.UpdateLogoutRedirectUriSettingValue{
			Uri: *openapi.NewNullableString(&defaultUri),
		},
	}
	updateAndVerifySetting(ctx, t, SettingsKeyLogoutRedirectUri, updateRequest, FieldUri, defaultUri)
}

func TestGetSecuritySettingByKey_InvalidKey(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	_, httpResp, err := OpenapiClient.SettingsAPI.GetSecuritySettingByKey(ctx, "invalid-key").
		Execute()

	require.Error(t, err, "Should return an error for invalid key")
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode, "Should return 400 Bad Request status for invalid key")
}

func TestUpdateSecuritySettingByKey_AutoRedirectSso_Invalid(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	enabled := true
	updateRequest := openapi.UpdateSecuritySettingByKeyRequest{
		UpdateAutoRedirectSsoSettingValue: &openapi.UpdateAutoRedirectSsoSettingValue{
			Enabled:  *openapi.NewNullableBool(&enabled),
			IdpAlias: *openapi.NewNullableString(nil),
		},
	}
	httpResp, err := OpenapiClient.SettingsAPI.UpdateSecuritySettingByKey(ctx, SettingsKeyAutoRedirectSSO).
		UpdateSecuritySettingByKeyRequest(updateRequest).
		Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
}

func TestUpdateSecuritySettingByKey_BrowserSessionTimeout_Zero_DisablesTimeout(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("browser-session-timeout-zero")

	zeroTimeout := float32(0)
	updateRequest := openapi.UpdateSecuritySettingByKeyRequest{
		UpdateBrowserSessionTimeoutSettingValue: &openapi.UpdateBrowserSessionTimeoutSettingValue{
			IdleTimeout: *openapi.NewNullableFloat32(&zeroTimeout),
		},
	}
	updateAndVerifySetting(ctx, t, SettingsKeyBrowserSessionTimeout, updateRequest, FieldIdleTimeout, float64(zeroTimeout))
}

func TestUpdateSecuritySettingByKey_BrowserSessionTimeout_Invalid(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	invalidTimeout := float32(-1)
	updateRequest := openapi.UpdateSecuritySettingByKeyRequest{
		UpdateBrowserSessionTimeoutSettingValue: &openapi.UpdateBrowserSessionTimeoutSettingValue{
			IdleTimeout: *openapi.NewNullableFloat32(&invalidTimeout),
		},
	}
	httpResp, err := OpenapiClient.SettingsAPI.UpdateSecuritySettingByKey(ctx, SettingsKeyBrowserSessionTimeout).
		UpdateSecuritySettingByKeyRequest(updateRequest).
		Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
}

func TestUpdateSecuritySettingByKey_LogoutRedirectUri_Invalid(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	invalidUri := "not-a-valid-url"
	updateRequest := openapi.UpdateSecuritySettingByKeyRequest{
		UpdateLogoutRedirectUriSettingValue: &openapi.UpdateLogoutRedirectUriSettingValue{
			Uri: *openapi.NewNullableString(&invalidUri),
		},
	}
	httpResp, err := OpenapiClient.SettingsAPI.UpdateSecuritySettingByKey(ctx, SettingsKeyLogoutRedirectUri).
		UpdateSecuritySettingByKeyRequest(updateRequest).
		Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
}

func TestGetSecuritySettingByKey_NonExistentKey(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	key := "nonExistentSettingKey"
	_, httpResp, err := OpenapiClient.SettingsAPI.GetSecuritySettingByKey(ctx, openapi.SettingsKeyEnum(key)).Execute()
	require.Error(t, err, "Should return an error for non-existent key")
	require.NotNil(t, httpResp)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode, "Should return 404 Not Found for non-existent key")
}

func TestUpdateSecuritySettingByKey_NullEmptyValues(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	enabled := true
	updateRequest := openapi.UpdateSecuritySettingByKeyRequest{
		UpdateAutoRedirectSsoSettingValue: &openapi.UpdateAutoRedirectSsoSettingValue{
			Enabled:  *openapi.NewNullableBool(&enabled),
			IdpAlias: *openapi.NewNullableString(openapi.PtrString("")),
		},
	}
	httpResp, err := OpenapiClient.SettingsAPI.UpdateSecuritySettingByKey(ctx, openapi.SettingsKeyEnum(SettingsKeyAutoRedirectSSO)).
		UpdateSecuritySettingByKeyRequest(updateRequest).
		Execute()
	require.Error(t, err, "Should return an error for empty idpAlias")
	require.NotNil(t, httpResp)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode, "Should return 400 Bad Request for empty idpAlias")
}

func updateAndVerifySetting[T comparable](ctx context.Context, t *testing.T, key string, updateRequest openapi.UpdateSecuritySettingByKeyRequest, field string, expected T) {
	setting, httpResp, err := updateSettingWithContext(ctx, key, updateRequest)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.NotNil(t, setting, "Response should not be nil")
	valueMap, ok := setting["value"].(map[string]interface{})
	require.True(t, ok, "setting should contain 'value'")
	require.Contains(t, valueMap, field, "Response value should contain "+field)
	actual, ok := valueMap[field].(T)
	if !ok {
		t.Fatalf("field %s should be of type %T", field, expected)
	}
	require.Equal(t, expected, actual, "%s value should match updated value", field)

	settingsKey := openapi.SettingsKeyEnum(key)
	getSetting, _, err := OpenapiClient.SettingsAPI.GetSecuritySettingByKey(ctx, settingsKey).Execute()
	require.NoError(t, err)
	getValueMap, ok := getSetting["value"].(map[string]interface{})
	require.True(t, ok, "GET setting should contain 'value' map")
	getActual, ok := getValueMap[field].(T)
	if !ok {
		t.Fatalf("field %s should be of type %T in GET after update", field, expected)
	}
	require.Equal(t, expected, getActual, "%s value should match updated value in GET after update", field)
}

func updateSetting(key string, req openapi.UpdateSecuritySettingByKeyRequest) (map[string]interface{}, *http.Response, error) {
	ctx := newTestContext()
	settingsKey := openapi.SettingsKeyEnum(key)
	httpResp, err := OpenapiClient.SettingsAPI.UpdateSecuritySettingByKey(ctx, settingsKey).
		UpdateSecuritySettingByKeyRequest(req).
		Execute()
	setting := map[string]interface{}{}
	if err == nil {
		setting, _, _ = OpenapiClient.SettingsAPI.GetSecuritySettingByKey(ctx, settingsKey).Execute()
	}
	return setting, httpResp, err
}

func assertSettingField(t *testing.T, setting map[string]interface{}, field string, expected interface{}) {
	valueMap, ok := setting["value"].(map[string]interface{})
	require.True(t, ok, "setting should contain 'value'")
	actual, ok := valueMap[field]
	require.True(t, ok, "field %s should be present in value", field)
	require.Equal(t, expected, actual, "field %s value should match", field)
}

func updateSettingWithContext(ctx context.Context, key string, req openapi.UpdateSecuritySettingByKeyRequest) (map[string]interface{}, *http.Response, error) {
	settingsKey := openapi.SettingsKeyEnum(key)
	httpResp, err := OpenapiClient.SettingsAPI.UpdateSecuritySettingByKey(ctx, settingsKey).
		UpdateSecuritySettingByKeyRequest(req).
		Execute()
	setting := map[string]interface{}{}
	if err == nil {
		setting, _, _ = OpenapiClient.SettingsAPI.GetSecuritySettingByKey(ctx, settingsKey).Execute()
	}
	return setting, httpResp, err
}
