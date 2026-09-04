/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/internal/generated/pkg/api"
	settings_manager "github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/settings"
	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/utils"
)

type BrandingSettingsAPIService struct {
	settingManager settings_manager.SettingsManager
	settingsHelper *SettingsHelper
}

func NewBrandingSettingsAPIService() *BrandingSettingsAPIService {
	return &BrandingSettingsAPIService{
		settingManager: *settings_manager.NewSettingsManager(),
		settingsHelper: NewSettingsHelper(),
	}
}

func (s *BrandingSettingsAPIService) GetBrandingSettings(ctx context.Context, tenantId string) (api.ImplResponse, error) {
	tenantID, errorResponse, err := s.settingsHelper.HandleTenantAuthorization(ctx, tenantId, "get branding settings")
	if err != nil {
		return *errorResponse, err
	}
	if tenantID == nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to get tenant ID"}
	}
	settingsList, err := s.settingManager.GetSettings(*tenantID, utils.BrandingCategory)
	if err != nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: fmt.Sprintf("failed to get branding settings for tenant id: %s.", tenantId)}
	}
	response, err := s.settingManager.TransformSettingsToResponse(settingsList, utils.BrandingCategory)
	if err != nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: fmt.Sprintf("failed to transform settings for tenant id: %s.", tenantId)}
	}
	return api.Response(http.StatusOK, response), nil
}

func (s *BrandingSettingsAPIService) GetBrandingSettingByKey(ctx context.Context, key api.BrandingSettingsKeyEnum, tenantId string) (api.ImplResponse, error) {
	tenantID, errorResponse, err := s.settingsHelper.HandleTenantAuthorization(ctx, tenantId, fmt.Sprintf("get branding setting by key: '%s'", key))
	if err != nil {
		return *errorResponse, err
	}
	if tenantID == nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to get tenant ID"}
	}
	settingValue, err := s.settingManager.GetSettingByKey(string(key), *tenantID)
	if err != nil {
		return s.settingsHelper.HandleSettingsError(err, tenantId, fmt.Sprintf("get branding setting by key: '%s'", key))
	}

	response := s.settingsHelper.WrapSettingInResponse(settingValue)
	return api.Response(http.StatusOK, response), nil
}

func (s *BrandingSettingsAPIService) UpdateBrandingSettingByKey(ctx context.Context, key api.BrandingSettingsKeyEnum, updateLogo api.UpdateLogo, tenantId string) (api.ImplResponse, error) {
	tenantID, errorResponse, err := s.settingsHelper.HandleTenantAuthorization(ctx, tenantId, fmt.Sprintf("update branding setting by key: '%s'", key))
	if err != nil {
		return *errorResponse, err
	}
	if tenantID == nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to get tenant ID"}
	}

	err = s.settingManager.UpdateSettingByKey(ctx, string(key), *tenantID, updateLogo)
	if err != nil {
		var utilsErr utils.Error
		if errors.As(err, &utilsErr) {
			return api.Response(utilsErr.Code, nil), utilsErr
		}
		return s.settingsHelper.HandleSettingsError(err, tenantId, fmt.Sprintf("update branding setting by key: '%s'", key))
	}
	return api.Response(http.StatusOK, nil), nil
}
