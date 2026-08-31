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
	"strconv"
	"strings"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/settings"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting"
)

type SettingsService struct {
	settingsManager settings.SettingsManager
}

func NewSettingsService() api.SettingsAPIServicer {
	return &SettingsService{
		settingsManager: *settings.NewSettingsManager(),
	}
}

// GetWorkloadsSettings - Get workloads settings
func (s *SettingsService) GetWorkloadsSettings(ctx context.Context, tenantId string) (api.ImplResponse, error) {
	requestedTenantId, apiErr := s.extractTenantIdFromCtxAndRequestedId(ctx, &tenantId)
	if apiErr != nil {
		return api.Response(int(apiErr.Code), apiErr.Message), apiErr
	}

	settingsList, err := setting_pkg.GetAll(*requestedTenantId, settings.WorkloadsCategory)
	if err != nil {
		errMessage := fmt.Sprintf("failed to get workloads settings for tenant: '%s'. err: '%v'", *requestedTenantId, err)
		return api.Response(http.StatusInternalServerError, nil), errors.New(errMessage)
	}

	response := s.settingsManager.TransformSettingsToResponse(settingsList)
	return api.Response(http.StatusOK, response), nil
}

// GetWorkloadsSettingByKey - Get workloads setting by key
func (s *SettingsService) GetWorkloadsSettingByKey(ctx context.Context, key api.WorkloadsSettingsKeyEnum, tenantId string) (api.ImplResponse, error) {
	requestedTenantId, apiErr := s.extractTenantIdFromCtxAndRequestedId(ctx, &tenantId)
	if apiErr != nil {
		return api.Response(int(apiErr.Code), apiErr.Message), apiErr
	}

	setting, err := s.settingsManager.GetSettingByKey(key, *requestedTenantId)
	if err != nil {
		if strings.Contains(err.Error(), settings.ErrUnknownSettingKey) {
			return api.Response(http.StatusBadRequest, nil), errors.New(err.Error())
		}
		errMessage := fmt.Sprintf("failed to get workloads setting by key: '%s' for tenant: '%s'. err: '%v'", key, *requestedTenantId, err)
		return api.Response(http.StatusInternalServerError, nil), errors.New(errMessage)
	}

	response, err := s.settingsManager.TransformSettingByKeyToResponse(setting)
	if err != nil {
		errMessage := fmt.Sprintf("failed to transform workloads setting by key: '%s' for tenant: '%s'. err: '%v'", key, *requestedTenantId, err)
		return api.Response(http.StatusInternalServerError, nil), errors.New(errMessage)
	}

	return api.Response(http.StatusOK, response), nil
}

// UpdateWorkloadsSettingByKey - Update workloads setting by key
// Currently the only setting is yamlSubmission, when we'll add more settings, we'll need to update the value type, by setting the anyOf in the swagger spec
func (s *SettingsService) UpdateWorkloadsSettingByKey(ctx context.Context, key api.WorkloadsSettingsKeyEnum, value api.YamlSubmissionSetting, tenantId string) (api.ImplResponse, error) {
	requestedTenantId, apiErr := s.extractTenantIdFromCtxAndRequestedId(ctx, &tenantId)
	if apiErr != nil {
		return api.Response(int(apiErr.Code), apiErr.Message), apiErr
	}

	err := s.settingsManager.UpdateSettingByKey(ctx, key, *requestedTenantId, value)
	if err != nil {
		return api.Response(http.StatusInternalServerError, nil), errors.New(err.Error())
	}

	return api.Response(http.StatusOK, value), nil
}

// extractTenantIdFromCtxAndRequestedId - Validates that the user is allowed to send tenantId and returns the requested tenantId or the tenantId from the context
func (s *SettingsService) extractTenantIdFromCtxAndRequestedId(ctx context.Context, requestedTenantId *string) (*string, *api.Error) {
	if requestedTenantId != nil && *requestedTenantId != "" {
		if runaictx.IsManagementTenant(ctx) {
			return requestedTenantId, nil
		}
		return nil, api.NewError(http.StatusUnauthorized, "Unauthorized to manage workloads settings for tenant: "+*requestedTenantId, nil)
	}

	tenantId := strconv.Itoa(int(auth.TenantIdOf(ctx)))
	return &tenantId, nil
}
