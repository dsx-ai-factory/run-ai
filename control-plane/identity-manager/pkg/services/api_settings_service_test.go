package services

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"

	authMock "github.com/run-ai/runai/control-plane/identity-manager/pkg/business/authorization/mocks"
	businessMocks "github.com/run-ai/runai/control-plane/identity-manager/pkg/business/mocks"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/settings"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/setting/manager"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	iamMock "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/mocks"

	settingPkg "github.com/run-ai/runai/runai-common-packages/setting"
	settingPkgErr "github.com/run-ai/runai/runai-common-packages/setting/error"
	settingPkgMocks "github.com/run-ai/runai/runai-common-packages/setting/mocks"
)

const (
	realmName = "realm"
	tenantID  = "tenant-id"
	oidc      = "oidc"
)

func initSettingsServiceTests(t *testing.T) (iamServicesContainerMock *iamMock.MockServicesContainer, realmDalMock *businessMocks.RealmDal, authManagerMock *authMock.AuthManagerInterface, settingPkgMock *settingPkgMocks.SettingsManager, testSubject *SettingsAPIService, ctx context.Context) {
	mockIamServicesContainer := iamMock.NewMockServicesContainer()
	realmDalMock = businessMocks.NewRealmDal(t)
	authManagerMock = authMock.NewAuthManagerInterface(t)
	settingPkgManagerMock := settingPkgMocks.NewSettingsManager(t)
	settingPkg.SetMock(settingPkgManagerMock)

	ctx = context.WithValue(context.Background(), utils.AuthRealmKey, realmName)
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantID)

	testSubject = &SettingsAPIService{
		iamServices:    *mockIamServicesContainer.ToServicesContainer(),
		authManager:    authManagerMock,
		settingManager: *settings.NewSettingsManager(mockIamServicesContainer.ToServicesContainer()),
	}

	return mockIamServicesContainer, realmDalMock, authManagerMock, settingPkgManagerMock, testSubject, ctx
}

func Test_GetSecuritySettings_HappyFlow(t *testing.T) {
	_, _, authMock, settingPkgMock, testSubject, ctx := initSettingsServiceTests(t)

	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realmName), ptr.To(tenantID), nil)

	mockSettings := []manager.Settinger{
		&settings.LogoutRedirectUri{Uri: "https://example.com/logout"},
		&settings.AutoRedirectSSO{
			Enabled:  true,
			IdpAlias: ptr.To("test-idp-alias"),
		},
		&settings.BrowserSessionTimeout{
			IdleTimeout: ptr.To(float32(60)),
		},
	}
	settingPkgMock.On("GetAll", mock.Anything, tenantID, utils.SecurityCategory).Return(mockSettings, nil)

	response, err := testSubject.GetSecuritySettings(ctx, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.NotNil(t, response.Body)

	responseSettings, ok := response.Body.(*api.SettingsGetResponse)
	assert.True(t, ok, "Response body should be of type *api.SettingsGetResponse")

	// Verify LogoutRedirectUri setting
	assert.NotNil(t, responseSettings.LogoutRedirectUri.Value.Uri, "LogoutRedirectUri.Uri should not be nil")
	assert.Equal(t, "https://example.com/logout", *responseSettings.LogoutRedirectUri.Value.Uri)

	// Verify AutoRedirectSSO setting
	assert.NotNil(t, responseSettings.AutoRedirectSSO.Value.Enabled, "AutoRedirectSSO.Enabled should not be nil")
	assert.Equal(t, true, *responseSettings.AutoRedirectSSO.Value.Enabled)
	assert.NotNil(t, responseSettings.AutoRedirectSSO.Value.IdpAlias, "AutoRedirectSSO.IdpAlias should not be nil")
	assert.Equal(t, "test-idp-alias", *responseSettings.AutoRedirectSSO.Value.IdpAlias)

	// Verify BrowserSessionTimeout setting
	assert.NotNil(t, responseSettings.BrowserSessionTimeout.Value.IdleTimeout, "BrowserSessionTimeout.IdleTimeout should not be nil")
	assert.Equal(t, float32(60), *responseSettings.BrowserSessionTimeout.Value.IdleTimeout)

	authMock.AssertExpectations(t)
	settingPkgMock.AssertExpectations(t)
}

func Test_GetSecuritySettings_HappyFlow_WithTenantIdParam(t *testing.T) {
	_, _, authMock, settingPkgMock, testSubject, ctx := initSettingsServiceTests(t)
	tenantId := "tenant-id-param"

	authMock.On("GetRealmAndTenantID", ctx, tenantId).Return(ptr.To(realmName), ptr.To(tenantId), nil)

	mockSettings := []manager.Settinger{
		&settings.LogoutRedirectUri{Uri: "https://example.com/logout"},
		&settings.AutoRedirectSSO{
			Enabled:  true,
			IdpAlias: ptr.To("test-idp-alias"),
		},
		&settings.BrowserSessionTimeout{
			IdleTimeout: ptr.To(float32(60)),
		},
	}
	settingPkgMock.On("GetAll", mock.Anything, tenantId, utils.SecurityCategory).Return(mockSettings, nil)

	response, err := testSubject.GetSecuritySettings(ctx, tenantId)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.NotNil(t, response.Body)

	responseSettings, ok := response.Body.(*api.SettingsGetResponse)
	assert.True(t, ok, "Response body should be of type *api.SettingsGetResponse")

	// Verify LogoutRedirectUri setting
	assert.NotNil(t, responseSettings.LogoutRedirectUri.Value.Uri, "LogoutRedirectUri.Uri should not be nil")
	assert.Equal(t, "https://example.com/logout", *responseSettings.LogoutRedirectUri.Value.Uri)

	// Verify AutoRedirectSSO setting
	assert.NotNil(t, responseSettings.AutoRedirectSSO.Value.Enabled, "AutoRedirectSSO.Enabled should not be nil")
	assert.Equal(t, true, *responseSettings.AutoRedirectSSO.Value.Enabled)
	assert.NotNil(t, responseSettings.AutoRedirectSSO.Value.IdpAlias, "AutoRedirectSSO.IdpAlias should not be nil")
	assert.Equal(t, "test-idp-alias", *responseSettings.AutoRedirectSSO.Value.IdpAlias)

	// Verify BrowserSessionTimeout setting
	assert.NotNil(t, responseSettings.BrowserSessionTimeout.Value.IdleTimeout, "BrowserSessionTimeout.IdleTimeout should not be nil")
	assert.Equal(t, float32(60), *responseSettings.BrowserSessionTimeout.Value.IdleTimeout)

	authMock.AssertExpectations(t)
	settingPkgMock.AssertExpectations(t)
}

func Test_GetSecuritySettings_Unauthorized(t *testing.T) {
	_, _, authMock, _, testSubject, ctx := initSettingsServiceTests(t)
	tenantId := "tenant-id-param"

	authMock.On("GetRealmAndTenantID", ctx, tenantId).Return(nil, nil, fmt.Errorf("Unauthorized access"))
	response, err := testSubject.GetSecuritySettings(ctx, tenantId)

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Contains(t, err.(utils.Error).Message, "failed to get security settings")
	authMock.AssertExpectations(t)
}

func Test_GetSecuritySettings_AuthManagerGeneralError(t *testing.T) {
	_, _, authMock, _, testSubject, ctx := initSettingsServiceTests(t)
	tenantId := "tenant-id-param"

	authMock.On("GetRealmAndTenantID", ctx, tenantId).Return(ptr.To(realmName), ptr.To(tenantID), fmt.Errorf("error"))
	response, err := testSubject.GetSecuritySettings(ctx, tenantId)

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	authMock.AssertExpectations(t)
}

func Test_GetSecuritySettings_SettingsRetrievalError(t *testing.T) {
	_, _, authMock, settingPkgMock, testSubject, ctx := initSettingsServiceTests(t)

	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realmName), ptr.To(tenantID), nil)
	settingPkgMock.On("GetAll", mock.Anything, tenantID, utils.SecurityCategory).Return(([]manager.Settinger)(nil), fmt.Errorf("settings retrieval error"))

	response, err := testSubject.GetSecuritySettings(ctx, "")

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	authMock.AssertExpectations(t)
	settingPkgMock.AssertExpectations(t)
}

func Test_GetSecuritySettings_BadRequestError(t *testing.T) {
	_, _, authMock, settingPkgMock, testSubject, ctx := initSettingsServiceTests(t)

	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realmName), ptr.To(tenantID), nil)
	pkgErr := settingPkgErr.ErrorBadRequest{Message: fmt.Sprintf("failed to get settings by tenantId: '%s'. err: %v", tenantID, fmt.Errorf("error"))}
	settingPkgMock.On("GetAll", mock.Anything, tenantID, utils.SecurityCategory).Return(([]manager.Settinger)(nil), pkgErr)

	response, err := testSubject.GetSecuritySettings(ctx, "")

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	authMock.AssertExpectations(t)
	settingPkgMock.AssertExpectations(t)
}

func Test_GetSecuritySettings_NotFoundError(t *testing.T) {
	_, _, authMock, settingPkgMock, testSubject, ctx := initSettingsServiceTests(t)

	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realmName), ptr.To(tenantID), nil)
	pkgErr := settingPkgErr.ErrorNotFound{Message: fmt.Sprintf("failed to get settings by tenantId: '%s'. err: %v", tenantID, fmt.Errorf("error"))}
	settingPkgMock.On("GetAll", mock.Anything, tenantID, utils.SecurityCategory).Return(([]manager.Settinger)(nil), pkgErr)

	response, err := testSubject.GetSecuritySettings(ctx, "")

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusNotFound, response.Code)
	authMock.AssertExpectations(t)
	settingPkgMock.AssertExpectations(t)
}

func Test_GetSecuritySettings_ConflictError(t *testing.T) {
	_, _, authMock, settingPkgMock, testSubject, ctx := initSettingsServiceTests(t)

	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realmName), ptr.To(tenantID), nil)
	pkgErr := settingPkgErr.ErrorConflict{Message: fmt.Sprintf("failed to get settings by tenantId: '%s'. err: %v", tenantID, fmt.Errorf("error"))}
	settingPkgMock.On("GetAll", mock.Anything, tenantID, utils.SecurityCategory).Return(([]manager.Settinger)(nil), pkgErr)

	response, err := testSubject.GetSecuritySettings(ctx, "")

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusConflict, response.Code)
	authMock.AssertExpectations(t)
	settingPkgMock.AssertExpectations(t)
}

/// ********************************************************
/// Tests for UpdateSecuritySettingByKey
/// ********************************************************

func Test_UpdateSecuritySettingByKey_LogoutRedirectUri_HappyFlow(t *testing.T) {
	iamServicesMock, _, authMock, settingPkgMock, testSubject, ctx := initSettingsServiceTests(t)

	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realmName), ptr.To(tenantID), nil)

	updateRequest := api.UpdateSecuritySettingByKeyRequest{Uri: ptr.To("https://new-logout-uri.com")}
	settingValue := settings.LogoutRedirectUri{
		Uri: *updateRequest.Uri,
	}
	iamServicesMock.MockClientService.On("UpdatePostLogoutRedirectUris", ctx, realmName, utils.RUNAI_ADMIN_UI_CLIENT_ID, *updateRequest.Uri).Return(nil)
	settingPkgMock.On("Update", mock.Anything, tenantID, string(api.SETTINGSKEYENUM_LOGOUT_REDIRECT_URI), true, settingValue).Return(nil)

	response, err := testSubject.UpdateSecuritySettingByKey(ctx, api.SETTINGSKEYENUM_LOGOUT_REDIRECT_URI, updateRequest, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	responseBody := response.Body.(api.UpdateSecuritySettingByKeyRequest)
	assert.Equal(t, *updateRequest.Uri, *responseBody.Uri)
	authMock.AssertExpectations(t)
	settingPkgMock.AssertExpectations(t)
}

func Test_UpdateSecuritySettingByKey_AutoRedirectSSO_HappyFlow(t *testing.T) {
	iamServicesMock, _, authMock, settingPkgMock, testSubject, ctx := initSettingsServiceTests(t)

	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realmName), ptr.To(tenantID), nil)

	updateRequest := api.UpdateSecuritySettingByKeyRequest{
		Enabled:  ptr.To(true),
		IdpAlias: ptr.To("test-idp"),
	}
	settingValue := settings.AutoRedirectSSO{
		Enabled:  *updateRequest.Enabled,
		IdpAlias: updateRequest.IdpAlias,
	}

	iamServicesMock.MockRealmService.On("UpdateAutoRedirectSSO", ctx, realmName, updateRequest.IdpAlias, *updateRequest.Enabled).Return(nil)
	settingPkgMock.On("Update", mock.Anything, tenantID, string(api.SETTINGSKEYENUM_AUTO_REDIRECT_SSO), true, settingValue).Return(nil)

	response, err := testSubject.UpdateSecuritySettingByKey(ctx, api.SETTINGSKEYENUM_AUTO_REDIRECT_SSO, updateRequest, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	authMock.AssertExpectations(t)
	settingPkgMock.AssertExpectations(t)
	iamServicesMock.MockRealmService.AssertExpectations(t)
}

func Test_UpdateSecuritySettingByKey_BrowserSessionTimeout_HappyFlow(t *testing.T) {
	_, _, authMock, settingPkgMock, testSubject, ctx := initSettingsServiceTests(t)

	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realmName), ptr.To(tenantID), nil)

	idleTimeout := float32(120)
	updateRequest := api.UpdateSecuritySettingByKeyRequest{
		IdleTimeout: &idleTimeout,
	}
	settingValue := settings.BrowserSessionTimeout{
		IdleTimeout: updateRequest.IdleTimeout,
	}
	settingPkgMock.On("Update", mock.Anything, tenantID, string(api.SETTINGSKEYENUM_BROWSER_SESSION_TIMEOUT), true, settingValue).Return(nil)

	response, err := testSubject.UpdateSecuritySettingByKey(ctx, api.SETTINGSKEYENUM_BROWSER_SESSION_TIMEOUT, updateRequest, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	authMock.AssertExpectations(t)
	settingPkgMock.AssertExpectations(t)
}

func Test_UpdateSecuritySettingByKey__WithTenantIdParam_Unauthorized(t *testing.T) {
	_, _, authMock, _, testSubject, ctx := initSettingsServiceTests(t)
	tenantId := "tenant-id-param"

	updateRequest := api.UpdateSecuritySettingByKeyRequest{Uri: ptr.To("https://new-logout-uri.com")}
	authMock.On("GetRealmAndTenantID", ctx, tenantId).Return(nil, nil, fmt.Errorf("Unauthorized access"))

	response, err := testSubject.UpdateSecuritySettingByKey(ctx, api.SETTINGSKEYENUM_LOGOUT_REDIRECT_URI, updateRequest, tenantId)

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	utilsError, ok := err.(utils.Error)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, utilsError.Code)
	assert.Contains(t, utilsError.Message, "unauthorized to update security setting by key")
	authMock.AssertExpectations(t)
}

func Test_UpdateSecuritySettingByKey_SettingUpdateError(t *testing.T) {
	iamServicesMock, _, authMock, settingPkgMock, testSubject, ctx := initSettingsServiceTests(t)

	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realmName), ptr.To(tenantID), nil)

	updateRequest := api.UpdateSecuritySettingByKeyRequest{Uri: ptr.To("https://new-logout-uri.com")}
	expectedSetting := settings.LogoutRedirectUri{
		Uri: *updateRequest.Uri,
	}
	iamServicesMock.MockClientService.On("UpdatePostLogoutRedirectUris", ctx, realmName, utils.RUNAI_ADMIN_UI_CLIENT_ID, *updateRequest.Uri).Return(nil)
	pkgErr := settingPkgErr.ErrorInternal{Message: fmt.Sprintf("failed to update security setting by key: %s", api.SETTINGSKEYENUM_LOGOUT_REDIRECT_URI)}
	settingPkgMock.On("Update", mock.Anything, tenantID, string(api.SETTINGSKEYENUM_LOGOUT_REDIRECT_URI), true, expectedSetting).Return(pkgErr)

	response, err := testSubject.UpdateSecuritySettingByKey(ctx, api.SETTINGSKEYENUM_LOGOUT_REDIRECT_URI, updateRequest, "")

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	authMock.AssertExpectations(t)
	settingPkgMock.AssertExpectations(t)
}

func Test_UpdateSecuritySettingByKey_UnknownKey(t *testing.T) {
	_, _, authMock, _, testSubject, ctx := initSettingsServiceTests(t)

	invalidKey := api.SettingsKeyEnum("invalid-key")
	updateRequest := api.UpdateSecuritySettingByKeyRequest{Uri: ptr.To("https://test.com")}
	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realmName), ptr.To(tenantID), nil)

	response, err := testSubject.UpdateSecuritySettingByKey(ctx, invalidKey, updateRequest, "")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	authMock.AssertExpectations(t)
}

/// ********************************************************
/// Tests for GetSecuritySettingByKey
/// ********************************************************

func Test_GetSecuritySettingByKey_HappyFlow_LogoutRedirectUri(t *testing.T) {
	_, _, authMock, settingPkgMock, testSubject, ctx := initSettingsServiceTests(t)

	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realmName), ptr.To(tenantID), nil)
	mockLogoutRedirectUri := &settings.LogoutRedirectUri{Uri: "test-uri"}
	settingPkgMock.On("Get", mock.Anything, tenantID, mock.Anything).Return(mockLogoutRedirectUri, nil)

	response, err := testSubject.GetSecuritySettingByKey(ctx, api.SETTINGSKEYENUM_LOGOUT_REDIRECT_URI, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	responseBody := response.Body.(api.GetLogoutRedirectUriSetting)
	assert.Equal(t, mockLogoutRedirectUri.Uri, *responseBody.Value.Uri)
	assert.Equal(t, mockLogoutRedirectUri.DisplayName(), responseBody.DisplayName)
	assert.Equal(t, mockLogoutRedirectUri.Description(), responseBody.Description)
}

func Test_GetSecuritySettingByKey_HappyFlow_AutoRedirectSSO(t *testing.T) {
	_, _, authMock, settingPkgMock, testSubject, ctx := initSettingsServiceTests(t)

	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realmName), ptr.To(tenantID), nil)
	mockAutoRedirectSSO := &settings.AutoRedirectSSO{
		Enabled:  true,
		IdpAlias: ptr.To("test-idp-alias"),
	}
	settingPkgMock.On("Get", mock.Anything, tenantID, mock.Anything).Return(mockAutoRedirectSSO, nil)

	response, err := testSubject.GetSecuritySettingByKey(ctx, api.SETTINGSKEYENUM_AUTO_REDIRECT_SSO, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	responseBody := response.Body.(api.GetAutoRedirectSsoSetting)
	assert.Equal(t, mockAutoRedirectSSO.Enabled, *responseBody.Value.Enabled)
	assert.Equal(t, mockAutoRedirectSSO.IdpAlias, responseBody.Value.IdpAlias)
	assert.Equal(t, mockAutoRedirectSSO.DisplayName(), responseBody.DisplayName)
	assert.Equal(t, mockAutoRedirectSSO.Description(), responseBody.Description)
	authMock.AssertExpectations(t)
	settingPkgMock.AssertExpectations(t)
}

func Test_GetSecuritySettingByKey_HappyFlow_BrowserSessionTimeout(t *testing.T) {
	_, _, authMock, settingPkgMock, testSubject, ctx := initSettingsServiceTests(t)

	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realmName), ptr.To(tenantID), nil)
	idleTimeout := float32(90)
	mockBrowserSessionTimeout := &settings.BrowserSessionTimeout{
		IdleTimeout: &idleTimeout,
	}
	settingPkgMock.On("Get", mock.Anything, tenantID, mock.Anything).Return(mockBrowserSessionTimeout, nil)

	response, err := testSubject.GetSecuritySettingByKey(ctx, api.SETTINGSKEYENUM_BROWSER_SESSION_TIMEOUT, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	responseBody := response.Body.(api.GetBrowserSessionTimeoutSetting)
	assert.Equal(t, *mockBrowserSessionTimeout.IdleTimeout, *responseBody.Value.IdleTimeout)
	assert.Equal(t, mockBrowserSessionTimeout.DisplayName(), responseBody.DisplayName)
	assert.Equal(t, mockBrowserSessionTimeout.Description(), responseBody.Description)
	authMock.AssertExpectations(t)
	settingPkgMock.AssertExpectations(t)
}

func Test_GetSecuritySettingByKey_Unauthorized(t *testing.T) {
	_, _, authMock, _, testSubject, ctx := initSettingsServiceTests(t)
	tenantId := "tenant-id-param"
	authMock.On("GetRealmAndTenantID", ctx, tenantId).Return(nil, nil, fmt.Errorf("Unauthorized access"))
	response, err := testSubject.GetSecuritySettingByKey(ctx, api.SETTINGSKEYENUM_LOGOUT_REDIRECT_URI, tenantId)

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	utilsError, ok := err.(utils.Error)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, utilsError.Code)
	assert.Contains(t, utilsError.Message, "unauthorized to get security setting by key")
}

func Test_GetSecuritySettingByKey_AuthManagerGeneralError(t *testing.T) {
	_, _, authMock, _, testSubject, ctx := initSettingsServiceTests(t)

	authMock.On("GetRealmAndTenantID", ctx, "test-tenant").Return(nil, nil, fmt.Errorf("some general error"))
	response, err := testSubject.GetSecuritySettingByKey(ctx, api.SETTINGSKEYENUM_LOGOUT_REDIRECT_URI, "test-tenant")

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}
