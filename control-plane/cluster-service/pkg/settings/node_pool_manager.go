/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting"
	"github.com/run-ai/runai/runai-common-packages/setting/manager"
)

const NodePoolCategory = "node-pools"

// NodePoolSettingsManager handles retrieval, update, and transformation of node-pool-scoped settings.
type NodePoolSettingsManager struct{}

// NewNodePoolSettingsManager creates a new NodePoolSettingsManager.
func NewNodePoolSettingsManager() *NodePoolSettingsManager {
	return &NodePoolSettingsManager{}
}

func (s *NodePoolSettingsManager) GetSettingByKey(key api.NodePoolSettingsKeyEnum, tenantID string) (manager.Settinger, error) {
	switch key {
	case api.NODEPOOLSETTINGSKEYENUM_PROJECT_MANUAL_OVER_QUOTA:
		setting, err := setting_pkg.Get[*ProjectManualOverQuota](tenantID)
		if err != nil {
			return nil, err
		}
		return setting, nil
	case api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA:
		setting, err := setting_pkg.Get[*CPUResourcesQuota](tenantID)
		if err != nil {
			return nil, err
		}
		return setting, nil
	default:
		return nil, fmt.Errorf("%s: %s", ErrUnknownSettingKey, key)
	}
}

func (s *NodePoolSettingsManager) UpdateSettingByKey(ctx context.Context, key api.NodePoolSettingsKeyEnum, tenantID string, updateRequest api.UpdateNodePoolSettingByKeyRequest) error {
	if updateRequest.Enabled == nil {
		return errors.New("enabled is required")
	}
	switch key {
	case api.NODEPOOLSETTINGSKEYENUM_PROJECT_MANUAL_OVER_QUOTA:
		settingValue := ProjectManualOverQuota{Enabled: *updateRequest.Enabled}
		return setting_pkg.Update[*ProjectManualOverQuota](tenantID, settingValue).UpdatedByUser().Execute()
	case api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA:
		settingValue := CPUResourcesQuota{Enabled: *updateRequest.Enabled}
		return setting_pkg.Update[*CPUResourcesQuota](tenantID, settingValue).UpdatedByUser().Execute()
	default:
		return fmt.Errorf("%s: %s", ErrUnknownSettingKey, key)
	}
}

func (s *NodePoolSettingsManager) TransformSettingsToResponse(settingsList []manager.Settinger) *api.NodePoolSettingsGetResponse {
	response := &api.NodePoolSettingsGetResponse{}
	for _, setting := range settingsList {
		switch v := setting.(type) {
		case *ProjectManualOverQuota:
			response.ProjectManualOverQuota = api.NodePoolSetting{
				Enabled:     &v.Enabled,
				DisplayName: v.DisplayName(),
				Description: v.Description(),
				Stage:       string(v.Stage()),
			}
		case *CPUResourcesQuota:
			response.CpuResourcesQuota = api.NodePoolSetting{
				Enabled:     &v.Enabled,
				DisplayName: v.DisplayName(),
				Description: v.Description(),
				Stage:       string(v.Stage()),
			}
		}
	}
	return response
}
