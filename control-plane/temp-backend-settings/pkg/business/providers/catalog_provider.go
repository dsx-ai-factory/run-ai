/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package providers

import (
	"context"
	"strconv"

	catalog_openapi "github.com/run-ai/runai/runai-clients/catalog-service-client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"

	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/autorization"
	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/business/settings"
)

type CatalogProvider struct {
	client *catalog_openapi.APIClient
}

func NewCatalogProvider(client *catalog_openapi.APIClient) *CatalogProvider {
	return &CatalogProvider{client: client}
}

func (p *CatalogProvider) FetchSettings(ctx context.Context, tenantID int) (map[string]settings.SettingOverride, error) {
	result := make(map[string]settings.SettingOverride)

	req := p.client.SettingsAPI.GetCatalogSettings(ctx)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	resp, _, err := req.Execute()
	if err != nil {
		logging.Errorf("failed to fetch catalog settings from catalog-service: %v", err)
		return result, nil
	}

	if resp.EnableNgcIntegration != nil {
		override := settings.SettingOverride{}
		if val, ok := resp.EnableNgcIntegration.GetValueOk(); ok {
			if enabled, ok := val.GetEnabledOk(); ok && enabled != nil {
				override.Value = *enabled
			}
		}
		if dn, ok := resp.EnableNgcIntegration.GetDisplayNameOk(); ok && dn != nil {
			override.DisplayName = *dn
		}
		if desc, ok := resp.EnableNgcIntegration.GetDescriptionOk(); ok && desc != nil {
			override.Description = *desc
		}
		if stage, ok := resp.EnableNgcIntegration.GetStageOk(); ok && stage != nil {
			override.Stage = *stage
		}
		if override.Value != nil {
			result[settings.SettingKeyEnableNgcIntegration] = override
		}
	}

	return result, nil
}
