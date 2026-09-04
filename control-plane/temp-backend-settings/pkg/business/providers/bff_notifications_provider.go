/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package providers

import (
	"context"
	"strconv"

	bff_openapi "github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"

	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/autorization"
	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/business/settings"
)

type BffNotificationsProvider struct {
	client *bff_openapi.APIClient
}

func NewBffNotificationsProvider(client *bff_openapi.APIClient) *BffNotificationsProvider {
	return &BffNotificationsProvider{client: client}
}

func (p *BffNotificationsProvider) FetchSettings(ctx context.Context, tenantID int) (map[string]settings.SettingOverride, error) {
	result := make(map[string]settings.SettingOverride)

	req := p.client.SettingsAPI.GetNotificationsSettings(ctx)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	resp, _, err := req.Execute()
	if err != nil {
		logging.Errorf("failed to fetch notifications settings from bff: %v", err)
		return result, nil
	}

	if resp.AdminMessage != nil {
		override := settings.SettingOverride{}
		if valWrapper, ok := resp.AdminMessage.GetValueOk(); ok && valWrapper != nil {
			override.Value = *valWrapper
		}
		if dn, ok := resp.AdminMessage.GetDisplayNameOk(); ok && dn != nil {
			override.DisplayName = *dn
		}
		if desc, ok := resp.AdminMessage.GetDescriptionOk(); ok && desc != nil {
			override.Description = *desc
		}
		if override.Value != nil {
			result[settings.SettingKeyAdminMessage] = override
		}
	}

	return result, nil
}
