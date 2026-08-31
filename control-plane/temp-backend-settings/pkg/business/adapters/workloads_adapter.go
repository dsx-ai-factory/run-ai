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
	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/business/settings"
	wc_openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

var settingKeyToWorkloadsAPIKey = map[string]wc_openapi.WorkloadsSettingsKeyEnum{
	settings.SettingKeyEnablePolicies:              wc_openapi.WORKLOADSSETTINGSKEYENUM_ENABLE_POLICIES,
	settings.SettingKeyWorkloadSubmissionV2Enabled: wc_openapi.WORKLOADSSETTINGSKEYENUM_WORKLOAD_SUBMISSION_V2_ENABLED,
	settings.SettingKeyEnableTemplatesV2:           wc_openapi.WORKLOADSSETTINGSKEYENUM_ENABLE_TEMPLATES_V2,
	settings.SettingKeyEnableHybridSubmission:      wc_openapi.WORKLOADSSETTINGSKEYENUM_ENABLE_HYBRID_SUBMISSION,
	settings.SettingKeyGpuOverProvisioning:         wc_openapi.WORKLOADSSETTINGSKEYENUM_GPU_OVER_PROVISIONING,
}

type WorkloadsAdapter struct {
	SettingsAPI wc_openapi.SettingsAPI
}

func (a *WorkloadsAdapter) UpdateSetting(ctx context.Context, tenantID int, key string, value interface{}) error {
	apiKey, ok := settingKeyToWorkloadsAPIKey[key]
	if !ok {
		return fmt.Errorf("unknown workloads setting key: %s", key)
	}

	boolVal, _ := value.(bool)
	body := wc_openapi.NewUpdateWorkloadsSettingByKeyRequest()
	body.SetEnabled(boolVal)

	req := a.SettingsAPI.UpdateWorkloadsServiceSettingByKey(ctx, apiKey).
		UpdateWorkloadsSettingByKeyRequest(*body)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	_, resp, err := req.Execute()
	if err != nil {
		return fmt.Errorf("UpdateWorkloadsSettingByKey(%s): %w", apiKey, err)
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("UpdateWorkloadsSettingByKey(%s) returned status %d", apiKey, resp.StatusCode)
	}
	return nil
}
