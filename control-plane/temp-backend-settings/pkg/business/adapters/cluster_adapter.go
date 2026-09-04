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

	cluster_openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/autorization"
)

type ClusterAdapter struct {
	SettingsAPI cluster_openapi.SettingsAPI
}

func (a *ClusterAdapter) UpdateEnableAdvancedGPUMetrics(ctx context.Context, tenantID int, value interface{}) error {
	boolVal, ok := value.(bool)
	if !ok {
		return fmt.Errorf("UpdateEnableAdvancedGPUMetrics: expected bool, got %T", value)
	}

	body := cluster_openapi.UpdateClusterSettingByKeyRequest{}
	body.SetEnabled(boolVal)

	req := a.SettingsAPI.UpdateClusterSettingByKey(ctx, cluster_openapi.CLUSTERSETTINGSKEYENUM_ENABLE_ADVANCED_GPU_METRICS).
		UpdateClusterSettingByKeyRequest(body)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	resp, err := req.Execute()
	if err != nil {
		return fmt.Errorf("UpdateEnableAdvancedGPUMetrics: %w", err)
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("UpdateEnableAdvancedGPUMetrics returned status %d", resp.StatusCode)
	}
	return nil
}

func (a *ClusterAdapter) UpdateProjectManualOverQuota(ctx context.Context, tenantID int, value interface{}) error {
	return a.updateNodePoolSetting(ctx, tenantID, cluster_openapi.NODEPOOLSETTINGSKEYENUM_PROJECT_MANUAL_OVER_QUOTA, value)
}

func (a *ClusterAdapter) UpdateCPUResourcesQuota(ctx context.Context, tenantID int, value interface{}) error {
	return a.updateNodePoolSetting(ctx, tenantID, cluster_openapi.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA, value)
}

func (a *ClusterAdapter) updateNodePoolSetting(ctx context.Context, tenantID int, key cluster_openapi.NodePoolSettingsKeyEnum, value interface{}) error {
	boolVal, ok := value.(bool)
	if !ok {
		return fmt.Errorf("updateNodePoolSetting %s: expected bool, got %T", key, value)
	}

	body := cluster_openapi.UpdateNodePoolSettingByKeyRequest{}
	body.SetEnabled(boolVal)

	req := a.SettingsAPI.UpdateNodePoolSettingByKey(ctx, key).UpdateNodePoolSettingByKeyRequest(body)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	resp, err := req.Execute()
	if err != nil {
		return fmt.Errorf("updateNodePoolSetting %s: %w", key, err)
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("updateNodePoolSetting %s returned status %d", key, resp.StatusCode)
	}
	return nil
}
