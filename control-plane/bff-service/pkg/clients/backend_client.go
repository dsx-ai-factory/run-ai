/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"
	"encoding/json"
	"fmt"

	api "github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/client"
	settings_pkg "github.com/run-ai/runai/runai-common-packages/setting"
	settings_pkg_manager "github.com/run-ai/runai/runai-common-packages/setting/manager"
)

type BackendClient interface {
	settings_pkg.BackendClient
}

type DefaultBackendClient struct {
	apiClient *api.TempBackendSettingsClient
}

var mockBackendClient BackendClient

func SetMockBackendClient(mock BackendClient) {
	mockBackendClient = mock
}

func NewBackendClient(baseURL string) BackendClient {
	if mockBackendClient != nil {
		return mockBackendClient
	}
	return &DefaultBackendClient{
		apiClient: api.NewTempBackendSettingsClient(baseURL),
	}
}

func (client *DefaultBackendClient) GetSettingByKeyForAllTenants(ctx context.Context, key string) (*settings_pkg_manager.SettingValuesForAllTenantsBackendResponse, error) {
	result, err := client.apiClient.GetAllTenantsSettingByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var response settings_pkg_manager.SettingValuesForAllTenantsBackendResponse
	if err := json.Unmarshal(jsonBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal backend response: %w", err)
	}
	return &response, nil
}
