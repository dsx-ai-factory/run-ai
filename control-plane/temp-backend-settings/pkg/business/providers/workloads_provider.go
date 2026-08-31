/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package providers

import (
	"context"
	"strconv"

	wc_openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"

	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/autorization"
	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/business/settings"
)

type WorkloadsProvider struct {
	client *wc_openapi.APIClient
}

func NewWorkloadsProvider(client *wc_openapi.APIClient) *WorkloadsProvider {
	return &WorkloadsProvider{client: client}
}

func (p *WorkloadsProvider) FetchSettings(ctx context.Context, tenantID int) (map[string]settings.SettingOverride, error) {
	result := make(map[string]settings.SettingOverride)

	req := p.client.SettingsAPI.GetWorkloadsServiceSettings(ctx)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	resp, _, err := req.Execute()
	if err != nil {
		logging.Errorf("failed to fetch workloads settings from workloads service: %v", err)
		return result, nil
	}

	extractBool(resp.EnablePolicies, settings.SettingKeyEnablePolicies, result)
	extractBool(resp.WorkloadSubmissionV2Enabled, settings.SettingKeyWorkloadSubmissionV2Enabled, result)
	extractBool(resp.EnableTemplatesV2, settings.SettingKeyEnableTemplatesV2, result)
	extractBool(resp.EnableHybridSubmission, settings.SettingKeyEnableHybridSubmission, result)
	extractBool(resp.GpuOverProvisioning, settings.SettingKeyGpuOverProvisioning, result)

	return result, nil
}

type boolSettingGetter interface {
	GetValueOk() (*wc_openapi.GetEnableSettingsValue, bool)
	GetDisplayNameOk() (*string, bool)
	GetDescriptionOk() (*string, bool)
	GetStageOk() (*string, bool)
}

func extractBool(setting boolSettingGetter, key string, result map[string]settings.SettingOverride) {
	if setting == nil {
		return
	}
	override := settings.SettingOverride{}
	if val, ok := setting.GetValueOk(); ok {
		if enabled, ok := val.GetEnabledOk(); ok && enabled != nil {
			override.Value = *enabled
		}
	}
	if dn, ok := setting.GetDisplayNameOk(); ok && dn != nil {
		override.DisplayName = *dn
	}
	if desc, ok := setting.GetDescriptionOk(); ok && desc != nil {
		override.Description = *desc
	}
	if stage, ok := setting.GetStageOk(); ok && stage != nil {
		override.Stage = *stage
	}
	if override.Value != nil {
		result[key] = override
	}
}
