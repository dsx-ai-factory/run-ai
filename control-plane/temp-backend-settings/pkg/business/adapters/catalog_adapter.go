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

	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/autorization"
	catalog_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/catalog-service-client/generated"
)

type CatalogAdapter struct {
	SettingsAPI catalog_openapi.SettingsAPI
}

func (a *CatalogAdapter) UpdateEnableNgcIntegration(ctx context.Context, tenantID int, value interface{}) error {
	boolVal, ok := value.(bool)
	if !ok {
		return fmt.Errorf("UpdateEnableNgcIntegration: expected bool, got %T", value)
	}

	body := catalog_openapi.UpdateCatalogSettingByKeyRequest{}
	body.SetEnabled(boolVal)

	req := a.SettingsAPI.UpdateCatalogSettingByKey(ctx, catalog_openapi.CATALOGSETTINGSKEYENUM_ENABLE_NGC_INTEGRATION).
		UpdateCatalogSettingByKeyRequest(body)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	resp, err := req.Execute()
	if err != nil {
		return fmt.Errorf("UpdateEnableNgcIntegration: %w", err)
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("UpdateEnableNgcIntegration returned status %d", resp.StatusCode)
	}
	return nil
}
