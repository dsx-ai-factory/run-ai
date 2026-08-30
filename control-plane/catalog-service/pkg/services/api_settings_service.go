package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/run-ai/runai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/settings"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting"
)

// SettingsService handles CRUD operations for catalog-related settings.
type SettingsService struct {
	settingsManager settings.SettingsManager
}

// NewSettingsService creates a new SettingsService for the settings API.
func NewSettingsService() api.SettingsAPIServicer {
	return &SettingsService{
		settingsManager: *settings.NewSettingsManager(),
	}
}

func (s *SettingsService) GetCatalogSettings(ctx context.Context, tenantId string) (api.ImplResponse, error) {
	resolvedTenantId, apiErr := resolveTenantId(ctx, tenantId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to resolve tenant id: %v", apiErr.Message)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	settingsList, err := setting_pkg.GetAll(resolvedTenantId, settings.CatalogCategory)
	if err != nil {
		errMessage := fmt.Sprintf("failed to get catalog settings for tenant: '%s': %v", resolvedTenantId, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
		return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, errMessage, err)), nil
	}

	response := s.settingsManager.TransformSettingsToResponse(settingsList)
	return api.Response(http.StatusOK, response), nil
}

func (s *SettingsService) GetCatalogSettingByKey(ctx context.Context, key api.CatalogSettingsKeyEnum, tenantId string) (api.ImplResponse, error) {
	resolvedTenantId, apiErr := resolveTenantId(ctx, tenantId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to resolve tenant id: %v", apiErr.Message)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	setting, err := s.settingsManager.GetSettingByKey(key, resolvedTenantId)
	if err != nil {
		if strings.Contains(err.Error(), settings.ErrUnknownSettingKey) {
			return api.Response(http.StatusBadRequest, api.NewError(http.StatusBadRequest, err.Error(), nil)), nil
		}
		errMessage := fmt.Sprintf("failed to get catalog setting by key: '%s' for tenant: '%s': %v", key, resolvedTenantId, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
		return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, errMessage, err)), nil
	}

	response, err := s.settingsManager.TransformSettingByKeyToResponse(setting)
	if err != nil {
		errMessage := fmt.Sprintf("failed to transform catalog setting by key: '%s' for tenant: '%s': %v", key, resolvedTenantId, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
		return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, errMessage, err)), nil
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *SettingsService) UpdateCatalogSettingByKey(ctx context.Context, key api.CatalogSettingsKeyEnum, value api.UpdateCatalogSettingByKeyRequest, tenantId string) (api.ImplResponse, error) {
	resolvedTenantId, apiErr := resolveTenantId(ctx, tenantId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to resolve tenant id: %v", apiErr.Message)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	err := s.settingsManager.UpdateSettingByKey(ctx, key, resolvedTenantId, value)
	if err != nil {
		if strings.Contains(err.Error(), settings.ErrUnknownSettingKey) {
			return api.Response(http.StatusBadRequest, api.NewError(http.StatusBadRequest, err.Error(), nil)), nil
		}
		errMessage := fmt.Sprintf("failed to update catalog setting by key: '%s' for tenant: '%s': %v", key, resolvedTenantId, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
		return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, errMessage, err)), nil
	}

	return api.Response(http.StatusOK, value), nil
}

// resolveTenantId returns the effective tenant ID: if explicitly provided, only management tenants
// are allowed to use it; otherwise, the tenant ID is extracted from the request context.
func resolveTenantId(ctx context.Context, requestedTenantId string) (string, *api.Error) {
	if requestedTenantId != "" {
		if !runaictx.IsManagementTenant(ctx) {
			return "", api.NewError(http.StatusUnauthorized, "Unauthorized to manage catalog settings for tenant: "+requestedTenantId, nil)
		}
		return requestedTenantId, nil
	}

	tenantId := runaictx.GetString(ctx, runaictx.CfTenantId)
	if tenantId == nil {
		return "", api.NewError(http.StatusBadRequest, "Missing tenant id", nil)
	}
	return *tenantId, nil
}
