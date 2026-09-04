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

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/settings"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

// SettingsService handles CRUD operations for workloads-related settings.
type SettingsService struct {
	settingsManager settings.SettingsManagerInterface
}

// NewSettingsService creates a new SettingsService for the settings API.
func NewSettingsService() api.SettingsAPIServicer {
	return &SettingsService{
		settingsManager: settings.NewSettingsManager(),
	}
}

func (s *SettingsService) GetWorkloadsServiceSettings(ctx context.Context, tenantId string) (api.ImplResponse, error) {
	resolvedTenantId, apiErr := resolveTenantId(ctx, tenantId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to resolve tenant id: %v", apiErr.Message)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	settingsList, err := s.settingsManager.GetAllSettings(resolvedTenantId)
	if err != nil {
		errMessage := fmt.Sprintf("failed to get workloads settings for tenant: '%s': %v", resolvedTenantId, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
		return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, errMessage, err)), nil
	}

	response := s.settingsManager.TransformSettingsToResponse(settingsList)
	return api.Response(http.StatusOK, response), nil
}

func (s *SettingsService) GetWorkloadsServiceSettingByKey(ctx context.Context, key api.WorkloadsSettingsKeyEnum, tenantId string) (api.ImplResponse, error) {
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
		errMessage := fmt.Sprintf("failed to get workloads setting by key: '%s' for tenant: '%s': %v", key, resolvedTenantId, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
		return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, errMessage, err)), nil
	}

	response, err := s.settingsManager.TransformSettingByKeyToResponse(setting)
	if err != nil {
		errMessage := fmt.Sprintf("failed to transform workloads setting by key: '%s' for tenant: '%s': %v", key, resolvedTenantId, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
		return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, errMessage, err)), nil
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *SettingsService) UpdateWorkloadsServiceSettingByKey(ctx context.Context, key api.WorkloadsSettingsKeyEnum, value api.UpdateWorkloadsSettingByKeyRequest, tenantId string) (api.ImplResponse, error) {
	resolvedTenantId, apiErr := resolveTenantId(ctx, tenantId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to resolve tenant id: %v", apiErr.Message)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	err := s.settingsManager.UpdateSettingByKey(ctx, key, resolvedTenantId, value)
	if err != nil {
		if strings.Contains(err.Error(), settings.ErrUnknownSettingKey) ||
			strings.Contains(err.Error(), settings.ErrValidationFailed) {
			return api.Response(http.StatusBadRequest, api.NewError(http.StatusBadRequest, err.Error(), nil)), nil
		}
		errMessage := fmt.Sprintf("failed to update workloads setting by key: '%s' for tenant: '%s': %v", key, resolvedTenantId, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
		return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, errMessage, err)), nil
	}

	return api.Response(http.StatusOK, value), nil
}

// resolveTenantId returns the effective tenant ID: if explicitly provided, only management tenants
// are allowed to use it; otherwise, the tenant ID is extracted from the request context.
func resolveTenantId(ctx context.Context, requestedTenantId string) (string, *api.Error) {
	if requestedTenantId != "" {
		if runaictx.IsManagementTenant(ctx) {
			return requestedTenantId, nil
		}
		return "", api.NewError(http.StatusUnauthorized, "Unauthorized to manage workloads settings for tenant: "+requestedTenantId, nil)
	}

	tenantId := runaictx.GetString(ctx, runaictx.CfTenantId)
	if tenantId == nil {
		return "", api.NewError(http.StatusBadRequest, "Missing tenant id", nil)
	}
	return *tenantId, nil
}
