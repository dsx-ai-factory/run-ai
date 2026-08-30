package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/bff-service/internal/generated/pkg/api"
	settings_manager "github.com/run-ai/runai/control-plane/bff-service/pkg/settings"
	"github.com/run-ai/runai/control-plane/bff-service/pkg/utils"
)

type AppearanceSettingsAPIService struct {
	settingManager settings_manager.SettingsManager
	settingsHelper *SettingsHelper
}

func NewAppearanceSettingsAPIService() *AppearanceSettingsAPIService {
	return &AppearanceSettingsAPIService{
		settingManager: *settings_manager.NewSettingsManager(),
		settingsHelper: NewSettingsHelper(),
	}
}

func (s *AppearanceSettingsAPIService) GetAppearanceSettings(ctx context.Context, tenantId string) (api.ImplResponse, error) {
	tenantID, errorResponse, err := s.settingsHelper.HandleTenantAuthorization(ctx, tenantId, "get appearance settings")
	if err != nil {
		return *errorResponse, err
	}
	if tenantID == nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to get tenant ID"}
	}
	settingsList, err := s.settingManager.GetSettings(*tenantID, utils.UIAppearanceCategory)
	if err != nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: fmt.Sprintf("failed to get appearance settings for tenant id: %s.", tenantId)}
	}
	response, err := s.settingManager.TransformSettingsToResponse(settingsList, utils.UIAppearanceCategory)
	if err != nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: fmt.Sprintf("failed to transform settings for tenant id: %s.", tenantId)}
	}
	return api.Response(http.StatusOK, response), nil
}

func (s *AppearanceSettingsAPIService) GetAppearanceSettingByKey(ctx context.Context, key api.AppearanceSettingsKeyEnum, tenantId string) (api.ImplResponse, error) {
	tenantID, errorResponse, err := s.settingsHelper.HandleTenantAuthorization(ctx, tenantId, fmt.Sprintf("get appearance setting by key: '%s'", key))
	if err != nil {
		return *errorResponse, err
	}
	if tenantID == nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to get tenant ID"}
	}
	setting, err := s.settingManager.GetSettingByKey(string(key), *tenantID)
	if err != nil {
		return s.settingsHelper.HandleSettingsError(err, tenantId, fmt.Sprintf("get appearance setting by key: '%s'", key))
	}

	response := s.settingsHelper.WrapSettingInResponse(setting)
	return api.Response(http.StatusOK, response), nil
}

func (s *AppearanceSettingsAPIService) UpdateAppearanceSettingByKey(ctx context.Context, key api.AppearanceSettingsKeyEnum, updateRequest api.UpdateBannerConfig, tenantId string) (api.ImplResponse, error) {
	tenantID, errorResponse, err := s.settingsHelper.HandleTenantAuthorization(ctx, tenantId, fmt.Sprintf("update appearance setting by key: '%s'", key))
	if err != nil {
		return *errorResponse, err
	}
	if tenantID == nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to get tenant ID"}
	}

	err = s.settingManager.UpdateSettingByKey(ctx, string(key), *tenantID, updateRequest)
	if err != nil {
		if errors.Is(err, utils.ErrBannerTextRequiredWhenEnabled) {
			return api.Response(http.StatusBadRequest, nil), utils.Error{Code: http.StatusBadRequest, Message: err.Error()}
		}
		return s.settingsHelper.HandleSettingsError(err, tenantId, fmt.Sprintf("update appearance setting by key: '%s'", key))
	}
	return api.Response(http.StatusOK, updateRequest), nil
}
