/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package providers

import (
	"context"
	"strconv"

	authorization_openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"

	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/autorization"
	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/business/settings"
)

type AuthorizationProvider struct {
	client *authorization_openapi.APIClient
}

func NewAuthorizationProvider(client *authorization_openapi.APIClient) *AuthorizationProvider {
	return &AuthorizationProvider{client: client}
}

func (p *AuthorizationProvider) FetchSettings(ctx context.Context, tenantID int) (map[string]settings.SettingOverride, error) {
	result := make(map[string]settings.SettingOverride)

	req := p.client.SettingsAPI.GetAuthorizationSettings(ctx)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	resp, _, err := req.Execute()
	if err != nil {
		logging.Errorf("failed to fetch authorization settings from authorization service: %v", err)
		return result, nil
	}

	if resp.EnableScopedAuthorization != nil {
		if enabled, ok := resp.EnableScopedAuthorization.GetEnabledOk(); ok && enabled != nil {
			result[settings.SettingKeyEnableScopedAuthorization] = settings.SettingOverride{Value: *enabled}
		}
	}

	return result, nil
}
