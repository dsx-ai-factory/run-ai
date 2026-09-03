/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/settings"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
)

// SettingsService handles CRUD operations for authorization-related settings.
type SettingsService struct {
	settingsManager settings.SettingsManager
}

// NewSettingsService creates a new SettingsService for the settings API.
func NewSettingsService() api.SettingsAPIServicer {
	return &SettingsService{
		settingsManager: *settings.NewSettingsManager(),
	}
}

func (s *SettingsService) GetAuthorizationSettings(ctx context.Context, tenantId string) (api.ImplResponse, error) {
	resolvedTenantId, apiErr := resolveSettingsTenantId(ctx, tenantId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to resolve tenant id: %v", apiErr.Message)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	settingsList, err := setting_pkg.GetAll(resolvedTenantId, settings.AuthorizationCategory)
	if err != nil {
		errMessage := fmt.Sprintf("failed to get authorization settings for tenant: '%s': %v", resolvedTenantId, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
		return api.Response(http.StatusServiceUnavailable, api.NewError(http.StatusServiceUnavailable, errMessage, err)), nil
	}

	response := s.settingsManager.TransformSettingsToResponse(settingsList)
	return api.Response(http.StatusOK, response), nil
}

func (s *SettingsService) GetAuthorizationSettingByKey(ctx context.Context, key api.AuthorizationSettingsKeyEnum, tenantId string) (api.ImplResponse, error) {
	resolvedTenantId, apiErr := resolveSettingsTenantId(ctx, tenantId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to resolve tenant id: %v", apiErr.Message)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	setting, err := s.settingsManager.GetSettingByKey(key, resolvedTenantId)
	if err != nil {
		if strings.Contains(err.Error(), settings.ErrUnknownSettingKey) {
			return api.Response(http.StatusBadRequest, api.NewError(http.StatusBadRequest, err.Error(), nil)), nil
		}
		errMessage := fmt.Sprintf("failed to get authorization setting by key: '%s' for tenant: '%s': %v", key, resolvedTenantId, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
		return api.Response(http.StatusServiceUnavailable, api.NewError(http.StatusServiceUnavailable, errMessage, err)), nil
	}

	return api.Response(http.StatusOK, setting), nil
}

func (s *SettingsService) UpdateAuthorizationSettingByKey(ctx context.Context, key api.AuthorizationSettingsKeyEnum, value api.UpdateAuthorizationSettingByKeyRequest, tenantId string) (api.ImplResponse, error) {
	resolvedTenantId, apiErr := resolveSettingsTenantId(ctx, tenantId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to resolve tenant id: %v", apiErr.Message)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	err := s.settingsManager.UpdateSettingByKey(ctx, key, resolvedTenantId, value)
	if err != nil {
		if strings.Contains(err.Error(), settings.ErrUnknownSettingKey) || strings.Contains(err.Error(), "enabled is required") {
			return api.Response(http.StatusBadRequest, api.NewError(http.StatusBadRequest, err.Error(), nil)), nil
		}
		errMessage := fmt.Sprintf("failed to update authorization setting by key: '%s' for tenant: '%s': %v", key, resolvedTenantId, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
		return api.Response(http.StatusServiceUnavailable, api.NewError(http.StatusServiceUnavailable, errMessage, err)), nil
	}

	return api.Response(http.StatusOK, value), nil
}

// resolveSettingsTenantId returns the effective tenant ID: if explicitly provided, only management
// tenants are allowed to use it; otherwise, the tenant ID is extracted from the request context.
func resolveSettingsTenantId(ctx context.Context, requestedTenantId string) (string, *api.Error) {
	if requestedTenantId != "" {
		if !runaictx.IsManagementTenant(ctx) {
			return "", api.NewError(http.StatusUnauthorized, "Unauthorized to manage authorization settings for tenant: "+requestedTenantId, nil)
		}
		return requestedTenantId, nil
	}

	tenantId := runaictx.GetString(ctx, runaictx.CfTenantId)
	if tenantId == nil {
		return "", api.NewError(http.StatusBadRequest, "Missing tenant id", nil)
	}
	return *tenantId, nil
}
