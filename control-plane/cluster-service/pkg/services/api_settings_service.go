/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/settings"
	org_unit_service_client "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
)

// SettingsService handles CRUD operations for cluster-related settings.
type SettingsService struct {
	settingsManager         settings.SettingsManager
	nodePoolSettingsManager settings.NodePoolSettingsManager
	orgUnitClient           org_unit_service_client.IOrgUnitClient
}

// NewSettingsService creates a new SettingsService for the settings API.
func NewSettingsService(orgUnitClient org_unit_service_client.IOrgUnitClient) api.SettingsAPIServicer {
	return &SettingsService{
		settingsManager:         *settings.NewSettingsManager(),
		nodePoolSettingsManager: *settings.NewNodePoolSettingsManager(),
		orgUnitClient:           orgUnitClient,
	}
}

func (s *SettingsService) GetClusterSettings(ctx context.Context, tenantId string) (api.ImplResponse, error) {
	resolvedTenantId, apiErr := resolveSettingsTenantId(ctx, tenantId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to resolve tenant id: %v", apiErr.Message)
		return NewErrorResponse(apiErr), nil
	}

	settingsList, err := setting_pkg.GetAll(resolvedTenantId, settings.ClusterCategory)
	if err != nil {
		errMessage := fmt.Sprintf("failed to get cluster settings for tenant: '%s': %v", resolvedTenantId, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
		return api.Response(http.StatusServiceUnavailable, api.NewError(http.StatusServiceUnavailable, errMessage, err)), nil
	}

	response := s.settingsManager.TransformSettingsToResponse(settingsList)
	return api.Response(http.StatusOK, response), nil
}

func (s *SettingsService) GetClusterSettingByKey(ctx context.Context, key api.ClusterSettingsKeyEnum, tenantId string) (api.ImplResponse, error) {
	resolvedTenantId, apiErr := resolveSettingsTenantId(ctx, tenantId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to resolve tenant id: %v", apiErr.Message)
		return NewErrorResponse(apiErr), nil
	}

	setting, err := s.settingsManager.GetSettingByKey(key, resolvedTenantId)
	if err != nil {
		if strings.Contains(err.Error(), settings.ErrUnknownSettingKey) {
			return api.Response(http.StatusBadRequest, api.NewError(http.StatusBadRequest, err.Error(), nil)), nil
		}
		errMessage := fmt.Sprintf("failed to get cluster setting by key: '%s' for tenant: '%s': %v", key, resolvedTenantId, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
		return api.Response(http.StatusServiceUnavailable, api.NewError(http.StatusServiceUnavailable, errMessage, err)), nil
	}

	return api.Response(http.StatusOK, setting), nil
}

func (s *SettingsService) UpdateClusterSettingByKey(ctx context.Context, key api.ClusterSettingsKeyEnum, value api.UpdateClusterSettingByKeyRequest, tenantId string) (api.ImplResponse, error) {
	resolvedTenantId, apiErr := resolveSettingsTenantId(ctx, tenantId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to resolve tenant id: %v", apiErr.Message)
		return NewErrorResponse(apiErr), nil
	}

	err := s.settingsManager.UpdateSettingByKey(ctx, key, resolvedTenantId, value)
	if err != nil {
		if strings.Contains(err.Error(), settings.ErrUnknownSettingKey) || strings.Contains(err.Error(), "enabled is required") {
			return api.Response(http.StatusBadRequest, api.NewError(http.StatusBadRequest, err.Error(), nil)), nil
		}
		errMessage := fmt.Sprintf("failed to update cluster setting by key: '%s' for tenant: '%s': %v", key, resolvedTenantId, err)
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
			return "", api.NewError(http.StatusUnauthorized, "Unauthorized to manage cluster settings for tenant: "+requestedTenantId, nil)
		}
		return requestedTenantId, nil
	}

	tenantId := runaictx.GetString(ctx, runaictx.CfTenantId)
	if tenantId == nil {
		return "", api.NewError(http.StatusBadRequest, "Missing tenant id", nil)
	}
	return *tenantId, nil
}

func (s *SettingsService) GetNodePoolSettings(ctx context.Context, tenantId string) (api.ImplResponse, error) {
	resolvedTenantId, apiErr := resolveSettingsTenantId(ctx, tenantId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to resolve tenant id: %v", apiErr.Message)
		return NewErrorResponse(apiErr), nil
	}

	settingsList, err := setting_pkg.GetAll(resolvedTenantId, settings.NodePoolCategory)
	if err != nil {
		errMessage := fmt.Sprintf("failed to get node pool settings for tenant: '%s': %v", resolvedTenantId, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
		return api.Response(http.StatusServiceUnavailable, api.NewError(http.StatusServiceUnavailable, errMessage, err)), nil
	}

	response := s.nodePoolSettingsManager.TransformSettingsToResponse(settingsList)
	return api.Response(http.StatusOK, response), nil
}

func (s *SettingsService) GetNodePoolSettingByKey(ctx context.Context, key api.NodePoolSettingsKeyEnum, tenantId string) (api.ImplResponse, error) {
	resolvedTenantId, apiErr := resolveSettingsTenantId(ctx, tenantId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to resolve tenant id: %v", apiErr.Message)
		return NewErrorResponse(apiErr), nil
	}

	setting, err := s.nodePoolSettingsManager.GetSettingByKey(key, resolvedTenantId)
	if err != nil {
		if strings.Contains(err.Error(), settings.ErrUnknownSettingKey) {
			return api.Response(http.StatusBadRequest, api.NewError(http.StatusBadRequest, err.Error(), nil)), nil
		}
		errMessage := fmt.Sprintf("failed to get node pool setting by key: '%s' for tenant: '%s': %v", key, resolvedTenantId, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
		return api.Response(http.StatusServiceUnavailable, api.NewError(http.StatusServiceUnavailable, errMessage, err)), nil
	}

	return api.Response(http.StatusOK, setting), nil
}

func (s *SettingsService) UpdateNodePoolSettingByKey(ctx context.Context, key api.NodePoolSettingsKeyEnum, value api.UpdateNodePoolSettingByKeyRequest, tenantId string) (api.ImplResponse, error) {
	resolvedTenantId, apiErr := resolveSettingsTenantId(ctx, tenantId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to resolve tenant id: %v", apiErr.Message)
		return NewErrorResponse(apiErr), nil
	}

	err := s.nodePoolSettingsManager.UpdateSettingByKey(ctx, key, resolvedTenantId, value)
	if err != nil {
		if strings.Contains(err.Error(), settings.ErrUnknownSettingKey) || strings.Contains(err.Error(), "enabled is required") {
			return api.Response(http.StatusBadRequest, api.NewError(http.StatusBadRequest, err.Error(), nil)), nil
		}
		errMessage := fmt.Sprintf("failed to update node pool setting by key: '%s' for tenant: '%s': %v", key, resolvedTenantId, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
		return api.Response(http.StatusServiceUnavailable, api.NewError(http.StatusServiceUnavailable, errMessage, err)), nil
	}

	// Propagate the toggle to all of the tenant's node pools. This is the in-process replacement for the
	// internal POST /internal/node-pools/settings API that temp-backend-settings used to call.
	err = s.propagateNodePoolSetting(ctx, key, resolvedTenantId, *value.Enabled)
	if err != nil {
		errMessage := fmt.Sprintf("failed to propagate node pool setting '%s' to tenant '%s' node pools: %v", key, resolvedTenantId, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
		return api.Response(http.StatusServiceUnavailable, api.NewError(http.StatusServiceUnavailable, errMessage, err)), nil
	}

	// Disabling CPU quota must also wipe the CPU/memory quota values on the tenant's projects and
	// departments. That data lives in org-unit-service, so trigger its DisableCpuQuota via the internal
	// settings endpoint. This replaces the temp-backend-settings PostSave that used to do it; both the UI
	// (direct) and legacy temp-backend write paths funnel through this handler.
	if key == api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA && !*value.Enabled {
		// org-unit resolves the tenant from the request context; for a management tenant acting on another
		// tenant, override it so org-unit wipes the resolved tenant's data, not the caller's.
		orgUnitCtx := context.WithValue(ctx, runaictx.CfTenantId, resolvedTenantId)
		if err = s.orgUnitClient.UpdateOrgUnitSettings(orgUnitCtx, dao.CpuQuotaSystemSettingFlag, "false"); err != nil {
			errMessage := fmt.Sprintf("failed to disable cpu quota on org units for tenant '%s': %v", resolvedTenantId, err)
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMessage)
			return api.Response(http.StatusServiceUnavailable, api.NewError(http.StatusServiceUnavailable, errMessage, err)), nil
		}
	}

	return api.Response(http.StatusOK, value), nil
}

// propagateNodePoolSetting applies the updated setting value to all of the tenant's node pools by
// mapping the setting key to its node pool system-setting flag and calling the existing DAO helper.
func (s *SettingsService) propagateNodePoolSetting(ctx context.Context, key api.NodePoolSettingsKeyEnum, tenantID string, enabled bool) error {
	var flag string
	switch key {
	case api.NODEPOOLSETTINGSKEYENUM_PROJECT_MANUAL_OVER_QUOTA:
		flag = dao.WeightsSystemSettingFlag
	case api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA:
		flag = dao.CpuQuotaSystemSettingFlag
	default:
		return nil
	}

	tenantIDInt, err := strconv.Atoi(tenantID)
	if err != nil {
		return fmt.Errorf("invalid tenant id '%s': %w", tenantID, err)
	}

	return dao.UpdateSystemSettingsOnTenantNodepools(ctx, int32(tenantIDInt), map[string]bool{flag: enabled})
}
