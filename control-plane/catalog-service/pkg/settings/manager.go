/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/run-ai/runai/control-plane/catalog-service/internal/generated/pkg/api"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting"
	"github.com/run-ai/runai/runai-common-packages/setting/manager"
)

const (
	ErrUnknownSettingKey = "unknown setting key"
	CatalogCategory      = "catalog"
)

// SettingsManager handles retrieval, update, and transformation of catalog settings.
type SettingsManager struct {
}

// NewSettingsManager creates a new SettingsManager.
func NewSettingsManager() *SettingsManager {
	return &SettingsManager{}
}

func (s *SettingsManager) GetSettingByKey(key api.CatalogSettingsKeyEnum, tenantID string) (manager.Settinger, error) {
	switch key {
	case api.CATALOGSETTINGSKEYENUM_ENABLE_NGC_INTEGRATION:
		setting, err := setting_pkg.Get[*EnableNgcIntegration](tenantID)
		if err != nil {
			return nil, err
		}
		return setting, nil
	default:
		return nil, fmt.Errorf("%s: %s", ErrUnknownSettingKey, key)
	}
}

func (s *SettingsManager) UpdateSettingByKey(ctx context.Context, key api.CatalogSettingsKeyEnum, tenantID string, updateRequest api.UpdateCatalogSettingByKeyRequest) error {
	switch key {
	case api.CATALOGSETTINGSKEYENUM_ENABLE_NGC_INTEGRATION:
		if updateRequest.Enabled != nil {
			settingValue := EnableNgcIntegration{
				Enabled: *updateRequest.Enabled,
			}
			return setting_pkg.Update[*EnableNgcIntegration](tenantID, settingValue).UpdatedByUser().Execute()
		}
		return errors.New("enabled is required")
	default:
		return fmt.Errorf("%s: %s", ErrUnknownSettingKey, key)
	}
}

func (s *SettingsManager) TransformSettingsToResponse(settingsList []manager.Settinger) *api.CatalogSettingsGetResponse {
	response := &api.CatalogSettingsGetResponse{}

	for _, setting := range settingsList {
		switch v := setting.(type) {
		case *EnableNgcIntegration:
			response.EnableNgcIntegration = transformEnableNgcIntegration(v)
		}
	}

	return response
}

func (s *SettingsManager) TransformSettingByKeyToResponse(setting manager.Settinger) (api.GetEnableNgcIntegration, error) {
	switch v := setting.(type) {
	case *EnableNgcIntegration:
		return transformEnableNgcIntegration(v), nil
	default:
		return api.GetEnableNgcIntegration{}, fmt.Errorf("%s", ErrUnknownSettingKey)
	}
}

func transformEnableNgcIntegration(v *EnableNgcIntegration) api.GetEnableNgcIntegration {
	return api.GetEnableNgcIntegration{
		Value: api.GetEnableNgcIntegrationValue{
			Enabled: &v.Enabled,
		},
		DisplayName: v.DisplayName(),
		Description: v.Description(),
		Stage:       string(v.Stage()),
	}
}
