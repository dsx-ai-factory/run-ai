/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package providers

import (
	"context"
	"strconv"

	orgunit_openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"

	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/autorization"
	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/business/settings"
)

type OrgUnitProvider struct {
	client *orgunit_openapi.APIClient
}

func NewOrgUnitProvider(client *orgunit_openapi.APIClient) *OrgUnitProvider {
	return &OrgUnitProvider{client: client}
}

func (p *OrgUnitProvider) FetchSettings(ctx context.Context, tenantID int) (map[string]settings.SettingOverride, error) {
	result := make(map[string]settings.SettingOverride)

	req := p.client.SettingsAPI.GetOrgUnitSettings(ctx)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	resp, _, err := req.Execute()
	if err != nil {
		logging.Errorf("failed to fetch org unit settings from org-unit-service: %v", err)
		return result, nil
	}

	if resp.LimitQuotaOverSubscription != nil {
		if enabled, ok := resp.LimitQuotaOverSubscription.GetEnabledOk(); ok && enabled != nil {
			result[settings.SettingKeyLimitQuotaOverSubscription] = settings.SettingOverride{Value: *enabled}
		}
	}

	return result, nil
}
