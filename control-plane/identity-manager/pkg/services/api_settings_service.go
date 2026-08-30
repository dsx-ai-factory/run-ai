package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam"
	auth_manager "github.com/run-ai/runai/control-plane/identity-manager/pkg/business/authorization"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/reconciler"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/settings"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting"
)

type SettingsAPIService struct {
	iamServices    iam.IamServiceContainer
	authManager    auth_manager.AuthManagerInterface
	settingManager settings.SettingsManager
}

func NewSettingsAPIService(
	iamServices *iam.IamServiceContainer,
	masterReconciler *reconciler.MasterReconcilerI,
	authManager *auth_manager.AuthManagerInterface,
) *SettingsAPIService {
	return &SettingsAPIService{
		iamServices:    *iamServices,
		authManager:    *authManager,
		settingManager: *settings.NewSettingsManager(iamServices),
	}
}

func (s *SettingsAPIService) GetSecuritySettings(ctx context.Context, tenantId string) (api.ImplResponse, error) {
	realm, tenantID, err := s.authManager.GetRealmAndTenantID(ctx, tenantId)
	if err != nil {
		if s.isUnauthorizedError(err) {
			return api.Response(http.StatusUnauthorized, err.Error()), utils.Error{Code: http.StatusUnauthorized, Message: "failed to get security settings"}
		}
		return utils.HandleError(ctx, utils.SecuritySetting, utils.Realm, *realm, err, "failed to get security settings")
	}
	settingsList, err := setting_pkg.GetAll(*tenantID, utils.SecurityCategory)
	if err != nil {
		errMessage := fmt.Sprintf("failed to get security settings for realm: '%s'. err: '%v'", *realm, err)
		return utils.HandleError(ctx, utils.SecuritySetting, utils.Realm, *realm, err, errMessage)
	}
	response := s.settingManager.TransformSettingsToResponse(settingsList)
	return api.Response(http.StatusOK, response), nil
}

func (s *SettingsAPIService) GetSecuritySettingByKey(ctx context.Context, key api.SettingsKeyEnum, tenantId string) (api.ImplResponse, error) {
	realm, tenantID, err := s.authManager.GetRealmAndTenantID(ctx, tenantId)
	if err != nil {
		if s.isUnauthorizedError(err) {
			errMsg := fmt.Sprintf("unauthorized to get security setting by key: '%s'.", key)
			return api.Response(http.StatusUnauthorized, err.Error()), utils.Error{Code: http.StatusUnauthorized, Message: errMsg}
		}
		errMessage := fmt.Sprintf("failed to get security setting by key: '%s'.", key)
		return utils.HandleError(ctx, utils.SecuritySetting, utils.Tenant, tenantId, err, errMessage)
	}
	setting, err := s.settingManager.GetSettingByKey(key, *tenantID)
	if err != nil {
		if strings.Contains(err.Error(), settings.ErrUnknownSettingKey) {
			return api.Response(http.StatusBadRequest, nil), utils.Error{Code: http.StatusBadRequest, Message: err.Error()}
		}
		errMessage := fmt.Sprintf("failed to get security setting by key: '%s' for realm: %s.", key, *realm)
		return utils.HandleError(ctx, utils.SecuritySetting, "setting key", string(key), err, errMessage)
	}
	return api.Response(http.StatusOK, setting), nil
}

func (s *SettingsAPIService) UpdateSecuritySettingByKey(ctx context.Context, key api.SettingsKeyEnum, updateIdmSettingByKeyRequest api.UpdateSecuritySettingByKeyRequest, tenantId string) (api.ImplResponse, error) {
	realm, tenantID, err := s.authManager.GetRealmAndTenantID(ctx, tenantId)
	if err != nil {
		if s.isUnauthorizedError(err) {
			errMsg := fmt.Sprintf("unauthorized to update security setting by key: '%s'", key)
			return api.Response(http.StatusUnauthorized, err.Error()), utils.Error{Code: http.StatusUnauthorized, Message: errMsg}
		}
		errMessage := fmt.Sprintf("failed to update security setting by key: '%s'", key)
		return utils.HandleError(ctx, utils.SecuritySetting, utils.Tenant, tenantId, err, errMessage)
	}

	err = s.settingManager.UpdateSettingByKey(ctx, *realm, key, *tenantID, updateIdmSettingByKeyRequest)
	if err != nil {
		if strings.Contains(err.Error(), settings.ErrUnknownSettingKey) {
			return api.Response(http.StatusBadRequest, nil), utils.Error{Code: http.StatusBadRequest, Message: err.Error()}
		}
		errMessage := fmt.Sprintf("failed updating security setting by key: '%s' for realm: '%s'", key, *realm)
		return utils.HandleError(ctx, utils.SecuritySetting, "setting key", string(key), err, errMessage)
	}
	return api.Response(http.StatusOK, updateIdmSettingByKeyRequest), nil
}

func (s *SettingsAPIService) isUnauthorizedError(err error) bool {
	return strings.HasPrefix(err.Error(), "Unauthorized")
}
