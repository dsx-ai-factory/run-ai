package services

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/run-ai/runai/control-plane/bff-service/internal/generated/pkg/api"
	authMock "github.com/run-ai/runai/control-plane/bff-service/pkg/authorization/mocks"
	"github.com/run-ai/runai/control-plane/bff-service/pkg/settings"
	ui_appearance "github.com/run-ai/runai/control-plane/bff-service/pkg/settings/ui-appearance"
	"github.com/run-ai/runai/control-plane/bff-service/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	settingPkg "github.com/run-ai/runai/runai-common-packages/setting"
	settingPkgManager "github.com/run-ai/runai/runai-common-packages/setting/manager"
	settingPkgMocks "github.com/run-ai/runai/runai-common-packages/setting/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

func initAppearanceSettingsServiceTests(t *testing.T) (settingPkgMock *settingPkgMocks.SettingsManager, testSubject *AppearanceSettingsAPIService, authManagerMock *authMock.AuthManagerInterface, ctx context.Context) {
	authManagerMock = authMock.NewAuthManagerInterface(t)
	settingPkgManagerMock := settingPkgMocks.NewSettingsManager(t)
	settingPkg.SetMock(settingPkgManagerMock)

	ctx = context.WithValue(context.Background(), "AuthRealm", realmName)
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantID)

	testSubject = &AppearanceSettingsAPIService{
		settingManager: *settings.NewSettingsManager(),
		settingsHelper: &SettingsHelper{
			authManager: authManagerMock,
		},
	}

	return settingPkgManagerMock, testSubject, authManagerMock, ctx
}

func TestAppearanceSettingsAPIService_GetAppearanceSettings_Success(t *testing.T) {
	settingPkgManagerMock, testSubject, authManagerMock, ctx := initAppearanceSettingsServiceTests(t)

	authManagerMock.On("GetTenantID", ctx, "").Return(ptr.To(tenantID), nil)
	settingPkgManagerMock.On("GetAll", mock.Anything, tenantID, string(utils.UIAppearanceCategory)).Return(getMockedAppearanceSettings(), nil)

	response, err := testSubject.GetAppearanceSettings(ctx, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	expected := &api.AppearanceSettingsGetAllResponse{
		BannerConfig: api.GetBannerConfig{
			Value: api.GetBannerConfigValue{
				Enabled:         ptr.To(true),
				Text:            ptr.To("PRODUCTION"),
				BackgroundColor: ptr.To("#D32F2F"),
				FontColor:       ptr.To("#FFFFFF"),
			},
			DisplayName: "Banner configuration",
			Description: "Configure a persistent banner displayed at the top of the Run:AI UI for all authenticated users.",
		},
	}
	assert.Equal(t, expected, response.Body)
}

func TestAppearanceSettingsAPIService_GetAppearanceSettings_NilTenantID(t *testing.T) {
	_, testSubject, authManagerMock, ctx := initAppearanceSettingsServiceTests(t)

	authManagerMock.On("GetTenantID", ctx, "").Return(nil, nil)

	response, err := testSubject.GetAppearanceSettings(ctx, "")
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	utilsErr, ok := err.(utils.Error)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, utilsErr.Code)
	assert.Contains(t, utilsErr.Message, "failed to get tenant ID")
}

func TestAppearanceSettingsAPIService_GetAppearanceSettings_AuthorizationError(t *testing.T) {
	_, testSubject, authManagerMock, ctx := initAppearanceSettingsServiceTests(t)

	authManagerMock.On("GetTenantID", ctx, "").Return(nil, errors.New("Unauthorized"))

	response, err := testSubject.GetAppearanceSettings(ctx, "")
	assert.Error(t, err)
	assert.Equal(t, http.StatusForbidden, response.Code)
}

func TestAppearanceSettingsAPIService_GetAppearanceSettingByKey_Success(t *testing.T) {
	settingPkgManagerMock, testSubject, authManagerMock, ctx := initAppearanceSettingsServiceTests(t)

	authManagerMock.On("GetTenantID", ctx, "").Return(ptr.To(tenantID), nil)
	settingPkgManagerMock.On("Get", mock.Anything, tenantID, mock.Anything).Return(
		&ui_appearance.BannerConfig{Enabled: true, Text: "PRODUCTION", BackgroundColor: "#D32F2F", FontColor: "#FFFFFF"},
		nil,
	)

	response, err := testSubject.GetAppearanceSettingByKey(ctx, api.APPEARANCESETTINGSKEYENUM_BANNER_CONFIG, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestAppearanceSettingsAPIService_GetAppearanceSettingByKey_NilTenantID(t *testing.T) {
	_, testSubject, authManagerMock, ctx := initAppearanceSettingsServiceTests(t)

	authManagerMock.On("GetTenantID", ctx, "").Return(nil, nil)

	response, err := testSubject.GetAppearanceSettingByKey(ctx, api.APPEARANCESETTINGSKEYENUM_BANNER_CONFIG, "")
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	utilsErr, ok := err.(utils.Error)
	assert.True(t, ok)
	assert.Contains(t, utilsErr.Message, "failed to get tenant ID")
}

func TestAppearanceSettingsAPIService_UpdateAppearanceSettingByKey_Success(t *testing.T) {
	settingPkgManagerMock, testSubject, authManagerMock, ctx := initAppearanceSettingsServiceTests(t)

	authManagerMock.On("GetTenantID", ctx, "").Return(ptr.To(tenantID), nil)
	settingPkgManagerMock.On("Get", mock.Anything, tenantID, mock.Anything).Return(
		&ui_appearance.BannerConfig{Enabled: false, Text: "", BackgroundColor: "", FontColor: ""},
		nil,
	)
	settingPkgManagerMock.On("Update", mock.Anything, tenantID, string(api.APPEARANCESETTINGSKEYENUM_BANNER_CONFIG), true, mock.AnythingOfType("appearance.BannerConfig")).Return(nil)

	updateRequest := api.UpdateBannerConfig{
		Enabled:         ptr.To(true),
		Text:            ptr.To("PRODUCTION"),
		BackgroundColor: ptr.To("#D32F2F"),
		FontColor:       ptr.To("#FFFFFF"),
	}
	response, err := testSubject.UpdateAppearanceSettingByKey(ctx, api.APPEARANCESETTINGSKEYENUM_BANNER_CONFIG, updateRequest, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestAppearanceSettingsAPIService_UpdateAppearanceSettingByKey_NilTenantID(t *testing.T) {
	_, testSubject, authManagerMock, ctx := initAppearanceSettingsServiceTests(t)

	authManagerMock.On("GetTenantID", ctx, "").Return(nil, nil)

	updateRequest := api.UpdateBannerConfig{Enabled: ptr.To(true)}
	response, err := testSubject.UpdateAppearanceSettingByKey(ctx, api.APPEARANCESETTINGSKEYENUM_BANNER_CONFIG, updateRequest, "")
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	utilsErr, ok := err.(utils.Error)
	assert.True(t, ok)
	assert.Contains(t, utilsErr.Message, "failed to get tenant ID")
}

func TestAppearanceSettingsAPIService_UpdateAppearanceSettingByKey_EnabledWithEmptyText(t *testing.T) {
	settingPkgManagerMock, testSubject, authManagerMock, ctx := initAppearanceSettingsServiceTests(t)

	authManagerMock.On("GetTenantID", ctx, "").Return(ptr.To(tenantID), nil)
	settingPkgManagerMock.On("Get", mock.Anything, tenantID, mock.Anything).Return(
		&ui_appearance.BannerConfig{Enabled: false, Text: "", BackgroundColor: "", FontColor: ""},
		nil,
	)
	updateRequest := api.UpdateBannerConfig{
		Enabled: ptr.To(true),
		Text:    ptr.To(""),
	}
	response, err := testSubject.UpdateAppearanceSettingByKey(ctx, api.APPEARANCESETTINGSKEYENUM_BANNER_CONFIG, updateRequest, "")
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	utilsErr, ok := err.(utils.Error)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, utilsErr.Code)
	assert.Equal(t, utils.ErrBannerTextRequiredWhenEnabled.Error(), utilsErr.Message)
}

func getMockedAppearanceSettings() []settingPkgManager.Settinger {
	return []settingPkgManager.Settinger{
		&ui_appearance.BannerConfig{
			Enabled:         true,
			Text:            "PRODUCTION",
			BackgroundColor: "#D32F2F",
			FontColor:       "#FFFFFF",
		},
	}
}
