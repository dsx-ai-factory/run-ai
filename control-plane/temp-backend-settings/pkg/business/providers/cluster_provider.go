/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package providers

import (
	"context"
	"strconv"

	cluster_openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"

	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/autorization"
	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/business/settings"
)

type ClusterProvider struct {
	client *cluster_openapi.APIClient
}

func NewClusterProvider(client *cluster_openapi.APIClient) *ClusterProvider {
	return &ClusterProvider{client: client}
}

func (p *ClusterProvider) FetchSettings(ctx context.Context, tenantID int) (map[string]settings.SettingOverride, error) {
	result := make(map[string]settings.SettingOverride)

	req := p.client.SettingsAPI.GetClusterSettings(ctx)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	resp, _, err := req.Execute()
	if err != nil {
		logging.Errorf("failed to fetch cluster settings from cluster-service: %v", err)
		return result, nil
	}

	if resp.EnableAdvancedGpuMetrics != nil {
		if enabled, ok := resp.EnableAdvancedGpuMetrics.GetEnabledOk(); ok && enabled != nil {
			result[settings.SettingKeyEnableAdvancedGPUMetrics] = settings.SettingOverride{Value: *enabled}
		}
	}

	p.fetchNodePoolSettings(ctx, tenantID, result)

	return result, nil
}

func (p *ClusterProvider) fetchNodePoolSettings(ctx context.Context, tenantID int, result map[string]settings.SettingOverride) {
	req := p.client.SettingsAPI.GetNodePoolSettings(ctx)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	resp, _, err := req.Execute()
	if err != nil {
		logging.Errorf("failed to fetch node pool settings from cluster-service: %v", err)
		return
	}

	if resp.ProjectManualOverQuota != nil {
		if enabled, ok := resp.ProjectManualOverQuota.GetEnabledOk(); ok && enabled != nil {
			result[settings.SettingKeyProjectManualOverQuota] = settings.SettingOverride{Value: *enabled}
		}
	}
	if resp.CpuResourcesQuota != nil {
		if enabled, ok := resp.CpuResourcesQuota.GetEnabledOk(); ok && enabled != nil {
			result[settings.SettingKeyCPUResourcesQuota] = settings.SettingOverride{Value: *enabled}
		}
	}
}
