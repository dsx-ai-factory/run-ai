/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/internal/generated/pkg/api"
	settings_manager "github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/settings"
	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/utils"
)

type WorkloadsSettingsAPIService struct {
	settingManager settings_manager.SettingsManager
	settingsHelper *SettingsHelper
}

func NewWorkloadsSettingsAPIService() *WorkloadsSettingsAPIService {
	return &WorkloadsSettingsAPIService{
		settingManager: *settings_manager.NewSettingsManager(),
		settingsHelper: NewSettingsHelper(),
	}
}

func (s *WorkloadsSettingsAPIService) GetWorkloadsSettings(ctx context.Context, tenantId string) (api.ImplResponse, error) {
	tenantID, errorResponse, err := s.settingsHelper.HandleTenantAuthorization(ctx, tenantId, "get workloads settings")
	if err != nil {
		return *errorResponse, err
	}
	if tenantID == nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to get tenant ID"}
	}
	settingsList, err := s.settingManager.GetSettings(*tenantID, utils.UIWorkloadsCategory)
	if err != nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: fmt.Sprintf("failed to get workloads settings for tenant id: %s.", tenantId)}
	}
	response, err := s.settingManager.TransformSettingsToResponse(settingsList, utils.UIWorkloadsCategory)
	if err != nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: fmt.Sprintf("failed to transform settings for tenant id: %s.", tenantId)}
	}
	return api.Response(http.StatusOK, response), nil
}

func (s *WorkloadsSettingsAPIService) GetWorkloadsSettingByKey(ctx context.Context, key api.WorkloadsSettingsKeyEnum, tenantId string) (api.ImplResponse, error) {
	tenantID, errorResponse, err := s.settingsHelper.HandleTenantAuthorization(ctx, tenantId, fmt.Sprintf("get workloads setting by key: '%s'", key))
	if err != nil {
		return *errorResponse, err
	}
	if tenantID == nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to get tenant ID"}
	}
	setting, err := s.settingManager.GetSettingByKey(string(key), *tenantID)
	if err != nil {
		return s.settingsHelper.HandleSettingsError(err, tenantId, fmt.Sprintf("get workloads setting by key: '%s'", key))
	}

	response := s.settingsHelper.WrapSettingInResponse(setting)
	return api.Response(http.StatusOK, response), nil
}

func (s *WorkloadsSettingsAPIService) UpdateWorkloadsSettingByKey(ctx context.Context, key api.WorkloadsSettingsKeyEnum, updateRequest api.UpdateWorkloadsSettingByKeyRequest, tenantId string) (api.ImplResponse, error) {
	tenantID, errorResponse, err := s.settingsHelper.HandleTenantAuthorization(ctx, tenantId, fmt.Sprintf("update workloads setting by key: '%s'", key))
	if err != nil {
		return *errorResponse, err
	}
	if tenantID == nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to get tenant ID"}
	}

	err = s.settingManager.UpdateSettingByKey(ctx, string(key), *tenantID, updateRequest)
	if err != nil {
		return s.settingsHelper.HandleSettingsError(err, tenantId, fmt.Sprintf("update workloads setting by key: '%s'", key))
	}
	return api.Response(http.StatusOK, updateRequest), nil
}
