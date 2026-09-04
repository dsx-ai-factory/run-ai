/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"context"
	"fmt"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/generated/pkg/api"
	iamMock "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/manager"
	settingPkgMocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

const (
	testTenantID = "test-tenant-id"
	testRealm    = "test-realm"
)

func setupManagerTest(t *testing.T) (*SettingsManager, *iamMock.MockServicesContainer, *settingPkgMocks.SettingsManager) {
	mockIamServicesContainer := iamMock.NewMockServicesContainer()
	settingPkgMock := settingPkgMocks.NewSettingsManager(t)
	setting_pkg.SetMock(settingPkgMock)

	mgr := NewSettingsManager(mockIamServicesContainer.ToServicesContainer())

	return mgr, mockIamServicesContainer, settingPkgMock
}

/// ********************************************************
/// Tests for GetSettingByKey
/// ********************************************************

func Test_GetSettingByKey_LogoutRedirectUri_HappyFlow(t *testing.T) {
	mgr, _, settingPkgMock := setupManagerTest(t)

	expectedSetting := &LogoutRedirectUri{Uri: "https://example.com/logout"}
	settingPkgMock.On("Get", mock.Anything, testTenantID, mock.Anything).Return(expectedSetting, nil)

	result, err := mgr.GetSettingByKey(api.SETTINGSKEYENUM_LOGOUT_REDIRECT_URI, testTenantID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	logoutRedirectUri, ok := result.(api.GetLogoutRedirectUriSetting)
	assert.True(t, ok)
	assert.Equal(t, expectedSetting.Uri, *logoutRedirectUri.Value.Uri)
	assert.Equal(t, expectedSetting.DisplayName(), logoutRedirectUri.DisplayName)
	assert.Equal(t, expectedSetting.Description(), logoutRedirectUri.Description)
	settingPkgMock.AssertExpectations(t)
}

func Test_GetSettingByKey_AutoRedirectSSO_HappyFlow(t *testing.T) {
	mgr, _, settingPkgMock := setupManagerTest(t)

	expectedSetting := &AutoRedirectSSO{
		Enabled:  true,
		IdpAlias: ptr.To("test-idp"),
	}
	settingPkgMock.On("Get", mock.Anything, testTenantID, mock.Anything).Return(expectedSetting, nil)

	result, err := mgr.GetSettingByKey(api.SETTINGSKEYENUM_AUTO_REDIRECT_SSO, testTenantID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	autoRedirectSSO, ok := result.(api.GetAutoRedirectSsoSetting)
	assert.True(t, ok)
	assert.Equal(t, expectedSetting.Enabled, *autoRedirectSSO.Value.Enabled)
	assert.Equal(t, expectedSetting.IdpAlias, autoRedirectSSO.Value.IdpAlias)
	assert.Equal(t, expectedSetting.DisplayName(), autoRedirectSSO.DisplayName)
	assert.Equal(t, expectedSetting.Description(), autoRedirectSSO.Description)
	settingPkgMock.AssertExpectations(t)
}

func Test_GetSettingByKey_BrowserSessionTimeout_HappyFlow(t *testing.T) {
	mgr, _, settingPkgMock := setupManagerTest(t)

	idleTimeout := float32(120)
	expectedSetting := &BrowserSessionTimeout{
		IdleTimeout: &idleTimeout,
	}
	settingPkgMock.On("Get", mock.Anything, testTenantID, mock.Anything).Return(expectedSetting, nil)

	result, err := mgr.GetSettingByKey(api.SETTINGSKEYENUM_BROWSER_SESSION_TIMEOUT, testTenantID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	browserSessionTimeout, ok := result.(api.GetBrowserSessionTimeoutSetting)
	assert.True(t, ok)
	assert.Equal(t, *expectedSetting.IdleTimeout, *browserSessionTimeout.Value.IdleTimeout)
	assert.Equal(t, expectedSetting.DisplayName(), browserSessionTimeout.DisplayName)
	assert.Equal(t, expectedSetting.Description(), browserSessionTimeout.Description)
	settingPkgMock.AssertExpectations(t)
}

func Test_GetSettingByKey_UnknownKey(t *testing.T) {
	mgr, _, _ := setupManagerTest(t)

	invalidKey := api.SettingsKeyEnum("invalid-key")
	result, err := mgr.GetSettingByKey(invalidKey, testTenantID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unknown setting key")
}

func Test_GetSettingByKey_SettingPackageError(t *testing.T) {
	mgr, _, settingPkgMock := setupManagerTest(t)

	expectedError := fmt.Errorf("setting package error")
	settingPkgMock.On("Get", mock.Anything, testTenantID, mock.Anything).Return(nil, expectedError)

	result, err := mgr.GetSettingByKey(api.SETTINGSKEYENUM_LOGOUT_REDIRECT_URI, testTenantID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedError, err)
	settingPkgMock.AssertExpectations(t)
}

/// ********************************************************
/// Tests for UpdateSettingByKey
/// ********************************************************

func Test_UpdateSettingByKey_LogoutRedirectUri_HappyFlow(t *testing.T) {
	mgr, iamServicesMock, settingPkgMock := setupManagerTest(t)

	updateRequest := api.UpdateSecuritySettingByKeyRequest{
		Uri: ptr.To("https://new-uri.com"),
	}

	expectedSetting := LogoutRedirectUri{Uri: *updateRequest.Uri}
	iamServicesMock.MockClientService.On("UpdatePostLogoutRedirectUris", mock.Anything, testRealm, utils.RUNAI_ADMIN_UI_CLIENT_ID, *updateRequest.Uri).Return(nil)
	settingPkgMock.On("Update", mock.Anything, testTenantID, string(api.SETTINGSKEYENUM_LOGOUT_REDIRECT_URI), true, expectedSetting).Return(nil)

	err := mgr.UpdateSettingByKey(context.Background(), testRealm, api.SETTINGSKEYENUM_LOGOUT_REDIRECT_URI, testTenantID, updateRequest)

	assert.NoError(t, err)
	iamServicesMock.MockClientService.AssertExpectations(t)
	settingPkgMock.AssertExpectations(t)
}

func Test_UpdateSettingByKey_LogoutRedirectUri_NilUri(t *testing.T) {
	mgr, _, _ := setupManagerTest(t)

	updateRequest := api.UpdateSecuritySettingByKeyRequest{
		Uri: nil,
	}

	err := mgr.UpdateSettingByKey(context.Background(), testRealm, api.SETTINGSKEYENUM_LOGOUT_REDIRECT_URI, testTenantID, updateRequest)

	assert.Error(t, err)
	assert.Equal(t, utils.ErrInvalidInput, err)
}

func Test_UpdateSettingByKey_AutoRedirectSSO_HappyFlow_Enabled(t *testing.T) {
	mgr, iamServicesMock, settingPkgMock := setupManagerTest(t)

	updateRequest := api.UpdateSecuritySettingByKeyRequest{
		Enabled:  ptr.To(true),
		IdpAlias: ptr.To("test-idp"),
	}

	expectedSetting := AutoRedirectSSO{
		Enabled:  *updateRequest.Enabled,
		IdpAlias: updateRequest.IdpAlias,
	}

	iamServicesMock.MockRealmService.On("UpdateAutoRedirectSSO", mock.Anything, testRealm, updateRequest.IdpAlias, *updateRequest.Enabled).Return(nil)
	settingPkgMock.On("Update", mock.Anything, testTenantID, string(api.SETTINGSKEYENUM_AUTO_REDIRECT_SSO), true, expectedSetting).Return(nil)

	err := mgr.UpdateSettingByKey(context.Background(), testRealm, api.SETTINGSKEYENUM_AUTO_REDIRECT_SSO, testTenantID, updateRequest)

	assert.NoError(t, err)
	iamServicesMock.MockRealmService.AssertExpectations(t)
	settingPkgMock.AssertExpectations(t)
}

func Test_UpdateSettingByKey_AutoRedirectSSO_HappyFlow_Disabled(t *testing.T) {
	mgr, iamServicesMock, settingPkgMock := setupManagerTest(t)

	updateRequest := api.UpdateSecuritySettingByKeyRequest{
		Enabled:  ptr.To(false),
		IdpAlias: nil,
	}

	expectedSetting := AutoRedirectSSO{
		Enabled:  *updateRequest.Enabled,
		IdpAlias: nil,
	}

	iamServicesMock.MockRealmService.On("UpdateAutoRedirectSSO", mock.Anything, testRealm, updateRequest.IdpAlias, *updateRequest.Enabled).Return(nil)
	settingPkgMock.On("Update", mock.Anything, testTenantID, string(api.SETTINGSKEYENUM_AUTO_REDIRECT_SSO), true, expectedSetting).Return(nil)

	err := mgr.UpdateSettingByKey(context.Background(), testRealm, api.SETTINGSKEYENUM_AUTO_REDIRECT_SSO, testTenantID, updateRequest)

	assert.NoError(t, err)
	iamServicesMock.MockRealmService.AssertExpectations(t)
	settingPkgMock.AssertExpectations(t)
}

func Test_UpdateSettingByKey_AutoRedirectSSO_NilEnabled(t *testing.T) {
	mgr, _, _ := setupManagerTest(t)

	updateRequest := api.UpdateSecuritySettingByKeyRequest{
		Enabled:  nil,
		IdpAlias: ptr.To("test-idp"),
	}

	err := mgr.UpdateSettingByKey(context.Background(), testRealm, api.SETTINGSKEYENUM_AUTO_REDIRECT_SSO, testTenantID, updateRequest)

	assert.Error(t, err)
	assert.Equal(t, utils.ErrInvalidInput, err)
}

func Test_UpdateSettingByKey_AutoRedirectSSO_ValidationFailure(t *testing.T) {
	mgr, _, _ := setupManagerTest(t)

	updateRequest := api.UpdateSecuritySettingByKeyRequest{
		Enabled:  ptr.To(true),
		IdpAlias: ptr.To(""),
	}

	err := mgr.UpdateSettingByKey(context.Background(), testRealm, api.SETTINGSKEYENUM_AUTO_REDIRECT_SSO, testTenantID, updateRequest)

	assert.Error(t, err)
	assert.Equal(t, utils.ErrInvalidInput, err)
}

func Test_UpdateSettingByKey_AutoRedirectSSO_UpdateAutoRedirectSSOError(t *testing.T) {
	mgr, iamServicesMock, _ := setupManagerTest(t)

	updateRequest := api.UpdateSecuritySettingByKeyRequest{
		Enabled:  ptr.To(true),
		IdpAlias: ptr.To("test-idp"),
	}

	expectedError := fmt.Errorf("realm service error")
	iamServicesMock.MockRealmService.On("UpdateAutoRedirectSSO", mock.Anything, testRealm, updateRequest.IdpAlias, *updateRequest.Enabled).Return(expectedError)

	err := mgr.UpdateSettingByKey(context.Background(), testRealm, api.SETTINGSKEYENUM_AUTO_REDIRECT_SSO, testTenantID, updateRequest)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	iamServicesMock.MockRealmService.AssertExpectations(t)
}

func Test_UpdateSettingByKey_BrowserSessionTimeout_HappyFlow(t *testing.T) {
	mgr, _, settingPkgMock := setupManagerTest(t)

	idleTimeout := float32(90)
	updateRequest := api.UpdateSecuritySettingByKeyRequest{
		IdleTimeout: &idleTimeout,
	}

	expectedSetting := BrowserSessionTimeout{IdleTimeout: &idleTimeout}
	settingPkgMock.On("Update", mock.Anything, testTenantID, string(api.SETTINGSKEYENUM_BROWSER_SESSION_TIMEOUT), true, expectedSetting).Return(nil)

	err := mgr.UpdateSettingByKey(context.Background(), testRealm, api.SETTINGSKEYENUM_BROWSER_SESSION_TIMEOUT, testTenantID, updateRequest)

	assert.NoError(t, err)
	settingPkgMock.AssertExpectations(t)
}

func Test_UpdateSettingByKey_BrowserSessionTimeout_NilIdleTimeout(t *testing.T) {
	mgr, _, _ := setupManagerTest(t)

	updateRequest := api.UpdateSecuritySettingByKeyRequest{
		IdleTimeout: nil,
	}

	err := mgr.UpdateSettingByKey(context.Background(), testRealm, api.SETTINGSKEYENUM_BROWSER_SESSION_TIMEOUT, testTenantID, updateRequest)

	assert.Error(t, err)
	assert.Equal(t, utils.ErrInvalidInput, err)
}

func Test_UpdateSettingByKey_UnknownKey(t *testing.T) {
	mgr, _, _ := setupManagerTest(t)

	invalidKey := api.SettingsKeyEnum("invalid-key")
	updateRequest := api.UpdateSecuritySettingByKeyRequest{
		Uri: ptr.To("https://test.com"),
	}

	err := mgr.UpdateSettingByKey(context.Background(), testRealm, invalidKey, testTenantID, updateRequest)

	assert.NoError(t, err)
}

/// ********************************************************
/// Tests for TransformSettingsToResponse
/// ********************************************************

func Test_TransformSettingsToResponse_AllSettings(t *testing.T) {
	mgr, _, _ := setupManagerTest(t)

	idleTimeout := float32(90)
	settingsList := []manager.Settinger{
		&LogoutRedirectUri{Uri: "https://example.com/logout"},
		&AutoRedirectSSO{
			Enabled:  true,
			IdpAlias: ptr.To("test-idp"),
		},
		&BrowserSessionTimeout{
			IdleTimeout: &idleTimeout,
		},
	}

	response := mgr.TransformSettingsToResponse(settingsList)

	assert.NotNil(t, response)
	assert.NotNil(t, response.LogoutRedirectUri.Value.Uri)
	assert.Equal(t, "https://example.com/logout", *response.LogoutRedirectUri.Value.Uri)
	assert.NotNil(t, response.AutoRedirectSSO.Value.Enabled)
	assert.Equal(t, true, *response.AutoRedirectSSO.Value.Enabled)
	assert.NotNil(t, response.AutoRedirectSSO.Value.IdpAlias)
	assert.Equal(t, "test-idp", *response.AutoRedirectSSO.Value.IdpAlias)
	assert.NotNil(t, response.BrowserSessionTimeout.Value.IdleTimeout)
	assert.Equal(t, float32(90), *response.BrowserSessionTimeout.Value.IdleTimeout)
}

func Test_TransformSettingsToResponse_SomeSettings(t *testing.T) {
	mgr, _, _ := setupManagerTest(t)

	settingsList := []manager.Settinger{
		&LogoutRedirectUri{Uri: "https://example.com/logout"},
		&AutoRedirectSSO{
			Enabled:  false,
			IdpAlias: nil,
		},
	}

	response := mgr.TransformSettingsToResponse(settingsList)

	assert.NotNil(t, response)
	assert.NotNil(t, response.LogoutRedirectUri.Value.Uri)
	assert.Equal(t, "https://example.com/logout", *response.LogoutRedirectUri.Value.Uri)
	assert.NotNil(t, response.AutoRedirectSSO.Value.Enabled)
	assert.Equal(t, false, *response.AutoRedirectSSO.Value.Enabled)
	assert.Nil(t, response.BrowserSessionTimeout.Value.IdleTimeout)
}

func Test_TransformSettingsToResponse_EmptySettings(t *testing.T) {
	mgr, _, _ := setupManagerTest(t)

	settingsList := []manager.Settinger{}

	response := mgr.TransformSettingsToResponse(settingsList)

	assert.NotNil(t, response)
	assert.Nil(t, response.LogoutRedirectUri.Value.Uri)
	assert.Nil(t, response.AutoRedirectSSO.Value.Enabled)
	assert.Nil(t, response.AutoRedirectSSO.Value.IdpAlias)
	assert.Nil(t, response.BrowserSessionTimeout.Value.IdleTimeout)
}

func Test_TransformSettingsToResponse_IgnoresUnknownType(t *testing.T) {
	mgr, _, _ := setupManagerTest(t)

	settingsList := []manager.Settinger{
		&LogoutRedirectUri{Uri: "https://example.com/logout"},
	}

	response := mgr.TransformSettingsToResponse(settingsList)

	assert.NotNil(t, response)
	assert.NotNil(t, response.LogoutRedirectUri.Value.Uri)
	assert.Equal(t, "https://example.com/logout", *response.LogoutRedirectUri.Value.Uri)
}
