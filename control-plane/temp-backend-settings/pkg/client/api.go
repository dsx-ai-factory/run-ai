/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package temp_backend_settings_client

import (
	"context"
	"encoding/json"
	"fmt"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/client/generated"
)

type ITempBackendSettingsClient interface {
	GetAllTenantsSettingByKey(ctx context.Context, key string) (*openapi.TenantSettingsByKey, error)
	GetTenantSettings(ctx context.Context, excludeMigrated bool) ([]openapi.Setting, error)
	GetTenantSettingsByKey(ctx context.Context, settingKey string) (bool, error)
	UpdateTenantSetting(ctx context.Context, tenantId string, req openapi.TenantSettingCreationRequest) (*openapi.TenantSettingCreationResponse, error)
}

type TempBackendSettingsClient struct {
	api *openapi.APIClient
}

func NewTempBackendSettingsClient(baseURL string) *TempBackendSettingsClient {
	config := openapi.NewConfiguration()
	config.Servers = openapi.ServerConfigurations{
		{URL: baseURL},
	}
	api := openapi.NewAPIClient(config)
	return &TempBackendSettingsClient{
		api: api,
	}
}

func (c *TempBackendSettingsClient) GetAllTenantsSettingByKey(ctx context.Context, key string) (*openapi.TenantSettingsByKey, error) {
	resp, _, err := c.api.InternalAPI.GetSettingByKeyForAllTenants(ctx, key).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get all tenants setting by key: %w", err)
	}

	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var result openapi.TenantSettingsByKey
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &result, nil
}

func (c *TempBackendSettingsClient) GetTenantSettings(ctx context.Context, excludeMigrated bool) ([]openapi.Setting, error) {
	resp, _, err := c.api.TenantAPI.GetTenantSettings(ctx).ExcludeMigrated(excludeMigrated).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant settings: %w", err)
	}
	return resp, nil
}

func (c *TempBackendSettingsClient) GetTenantSettingsByKey(ctx context.Context, settingKey string) (bool, error) {
	resp, _, err := c.api.TenantAPI.GetTenantSettingsByKey(ctx, settingKey).Execute()
	if err != nil {
		return false, fmt.Errorf("failed to get tenant setting by key: %w", err)
	}
	return resp, nil
}

func (c *TempBackendSettingsClient) UpdateTenantSetting(ctx context.Context, tenantId string, req openapi.TenantSettingCreationRequest) (*openapi.TenantSettingCreationResponse, error) {
	resp, _, err := c.api.TenantAPI.UpdateTenantSetting(ctx).TenantId(tenantId).TenantSettingCreationRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update tenant setting: %w", err)
	}
	return resp, nil
}
