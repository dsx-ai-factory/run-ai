package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/bff-service/internal/generated/pkg/api"
	settings_manager "github.com/run-ai/runai/control-plane/bff-service/pkg/settings"
	"github.com/run-ai/runai/control-plane/bff-service/pkg/utils"
)

type NotificationsSettingsAPIService struct {
	settingManager settings_manager.SettingsManager
	settingsHelper *SettingsHelper
}

func NewNotificationsSettingsAPIService() *NotificationsSettingsAPIService {
	return &NotificationsSettingsAPIService{
		settingManager: *settings_manager.NewSettingsManager(),
		settingsHelper: NewSettingsHelper(),
	}
}

func (s *NotificationsSettingsAPIService) GetNotificationsSettings(ctx context.Context, tenantId string) (api.ImplResponse, error) {
	tenantID, errorResponse, err := s.settingsHelper.HandleTenantAuthorization(ctx, tenantId, "get notifications settings")
	if err != nil {
		return *errorResponse, err
	}
	if tenantID == nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to get tenant ID"}
	}
	settingsList, err := s.settingManager.GetSettings(*tenantID, utils.NotificationsCategory)
	if err != nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: fmt.Sprintf("failed to get notifications settings for tenant id: %s.", tenantId)}
	}
	response, err := s.settingManager.TransformSettingsToResponse(settingsList, utils.NotificationsCategory)
	if err != nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: fmt.Sprintf("failed to transform settings for tenant id: %s.", tenantId)}
	}
	return api.Response(http.StatusOK, response), nil
}

func (s *NotificationsSettingsAPIService) GetNotificationsSettingByKey(ctx context.Context, key api.NotificationsSettingsKeyEnum, tenantId string) (api.ImplResponse, error) {
	tenantID, errorResponse, err := s.settingsHelper.HandleTenantAuthorization(ctx, tenantId, fmt.Sprintf("get notifications setting by key: '%s'", key))
	if err != nil {
		return *errorResponse, err
	}
	if tenantID == nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to get tenant ID"}
	}
	setting, err := s.settingManager.GetSettingByKey(string(key), *tenantID)
	if err != nil {
		return s.settingsHelper.HandleSettingsError(err, tenantId, fmt.Sprintf("get notifications setting by key: '%s'", key))
	}
	return api.Response(http.StatusOK, setting), nil
}

func (s *NotificationsSettingsAPIService) UpdateNotificationsSettingByKey(ctx context.Context, key api.NotificationsSettingsKeyEnum, updateRequest api.UpdateAdminMessage, tenantId string) (api.ImplResponse, error) {
	tenantID, errorResponse, err := s.settingsHelper.HandleTenantAuthorization(ctx, tenantId, fmt.Sprintf("update notifications setting by key: '%s'", key))
	if err != nil {
		return *errorResponse, err
	}
	if tenantID == nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to get tenant ID"}
	}

	err = s.settingManager.UpdateSettingByKey(ctx, string(key), *tenantID, updateRequest)
	if err != nil {
		return s.settingsHelper.HandleSettingsError(err, tenantId, fmt.Sprintf("update notifications setting by key: '%s'", key))
	}
	return api.Response(http.StatusOK, updateRequest), nil
}
