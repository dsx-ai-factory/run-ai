package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/settings"
	"github.com/run-ai/runai/runai-common-packages/logging"
	runaictx "github.com/run-ai/runai/runai-common-packages/runaictx"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting"
)

// SettingsService handles CRUD operations for org-unit-related settings.
type SettingsService struct {
	settingsManager settings.SettingsManager
}

// NewSettingsService creates a new SettingsService for the settings API.
func NewSettingsService() api.SettingsAPIServicer {
	return &SettingsService{
		settingsManager: *settings.NewSettingsManager(),
	}
}

func (s *SettingsService) GetOrgUnitSettings(ctx context.Context, tenantId string) (api.ImplResponse, error) {
	resolvedTenantId, errResp := resolveSettingsTenantId(ctx, tenantId)
	if errResp != nil {
		return *errResp, nil
	}

	settingsList, err := setting_pkg.GetAll(resolvedTenantId, settings.OrgUnitCategory)
	if err != nil {
		errMessage := fmt.Sprintf("failed to get org unit settings for tenant: '%s': %v", resolvedTenantId, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
		return HandleError(err, http.StatusServiceUnavailable, "Failed to get org unit settings"), nil
	}

	return api.Response(http.StatusOK, s.settingsManager.TransformSettingsToResponse(settingsList)), nil
}

func (s *SettingsService) GetOrgUnitSettingByKey(ctx context.Context, key api.OrgUnitSettingsKeyEnum, tenantId string) (api.ImplResponse, error) {
	resolvedTenantId, errResp := resolveSettingsTenantId(ctx, tenantId)
	if errResp != nil {
		return *errResp, nil
	}

	setting, err := s.settingsManager.GetSettingByKey(key, resolvedTenantId)
	if err != nil {
		if strings.Contains(err.Error(), settings.ErrUnknownSettingKey) {
			return HandleError(err, http.StatusBadRequest, err.Error()), nil
		}
		errMessage := fmt.Sprintf("failed to get org unit setting by key: '%s' for tenant: '%s': %v", key, resolvedTenantId, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
		return HandleError(err, http.StatusServiceUnavailable, "Failed to get org unit setting by key"), nil
	}

	return api.Response(http.StatusOK, setting), nil
}

func (s *SettingsService) UpdateOrgUnitSettingByKey(ctx context.Context, key api.OrgUnitSettingsKeyEnum, value api.UpdateOrgUnitSettingByKeyRequest, tenantId string) (api.ImplResponse, error) {
	resolvedTenantId, errResp := resolveSettingsTenantId(ctx, tenantId)
	if errResp != nil {
		return *errResp, nil
	}

	if err := s.settingsManager.UpdateSettingByKey(ctx, key, resolvedTenantId, value); err != nil {
		if strings.Contains(err.Error(), settings.ErrUnknownSettingKey) || strings.Contains(err.Error(), "enabled is required") {
			return HandleError(err, http.StatusBadRequest, err.Error()), nil
		}
		errMessage := fmt.Sprintf("failed to update org unit setting by key: '%s' for tenant: '%s': %v", key, resolvedTenantId, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
		return HandleError(err, http.StatusServiceUnavailable, "Failed to update org unit setting by key"), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("org unit setting '%s' updated for tenant '%s'", key, resolvedTenantId)
	return api.Response(http.StatusOK, value), nil
}

// resolveSettingsTenantId returns the effective tenant ID: if explicitly provided, only management
// tenants may use it; otherwise the tenant ID is taken from the request context. On error it returns
// a non-nil error response to write back.
func resolveSettingsTenantId(ctx context.Context, requestedTenantId string) (string, *api.ImplResponse) {
	if requestedTenantId != "" {
		if !runaictx.IsManagementTenant(ctx) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("unauthorized to manage settings for tenant: %s", requestedTenantId)
			resp := HandleError(nil, http.StatusUnauthorized, "Unauthorized to manage settings for tenant: "+requestedTenantId)
			return "", &resp
		}
		return requestedTenantId, nil
	}

	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to resolve tenant id: %v", err)
		resp := HandleError(err, http.StatusBadRequest, "Missing tenant id")
		return "", &resp
	}
	return *tenantId, nil
}
