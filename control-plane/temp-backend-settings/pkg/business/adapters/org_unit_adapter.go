/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package adapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/autorization"
	orgunit_openapi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
)

type OrgUnitSettingsAdapter struct {
	OrgUnitAPI  orgunit_openapi.OrgUnitAPI
	SettingsAPI orgunit_openapi.SettingsAPI
}

func (a *OrgUnitSettingsAdapter) UpdateOrgUnitSettings(ctx context.Context, tenantID int, name string, value string) error {
	resp, err := a.OrgUnitAPI.UpdateOrgUnitSettings(ctx).
		OrgUnitSingleSetting(orgunit_openapi.OrgUnitSingleSetting{
			Name:  &name,
			Value: &value,
		}).Execute()
	if err != nil {
		return fmt.Errorf("UpdateOrgUnitSettings: %w", err)
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("UpdateOrgUnitSettings returned status %d", resp.StatusCode)
	}
	return nil
}

func (a *OrgUnitSettingsAdapter) UpdateLimitQuotaOverSubscription(ctx context.Context, tenantID int, value interface{}) error {
	boolVal, ok := value.(bool)
	if !ok {
		return fmt.Errorf("UpdateLimitQuotaOverSubscription: expected bool, got %T", value)
	}

	body := orgunit_openapi.UpdateOrgUnitSettingByKeyRequest{}
	body.SetEnabled(boolVal)

	req := a.SettingsAPI.UpdateOrgUnitSettingByKey(ctx, orgunit_openapi.ORGUNITSETTINGSKEYENUM_LIMIT_QUOTA_OVER_SUBSCRIPTION).
		UpdateOrgUnitSettingByKeyRequest(body)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	resp, err := req.Execute()
	if err != nil {
		return fmt.Errorf("UpdateLimitQuotaOverSubscription: %w", err)
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("UpdateLimitQuotaOverSubscription returned status %d", resp.StatusCode)
	}
	return nil
}
