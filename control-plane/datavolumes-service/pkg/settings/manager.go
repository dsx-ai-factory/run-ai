/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/run-ai/runai/control-plane/datavolumes-service/internal/generated/pkg/api"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting"
	"github.com/run-ai/runai/runai-common-packages/setting/manager"
)

const (
	ErrUnknownSettingKey = "unknown setting key"
	DatavolumesCategory  = "datavolumes"
)

// SettingsManager handles retrieval, update, and transformation of datavolumes settings.
type SettingsManager struct {
}

// NewSettingsManager creates a new SettingsManager.
func NewSettingsManager() *SettingsManager {
	return &SettingsManager{}
}

func (s *SettingsManager) GetSettingByKey(key api.DatavolumesSettingsKeyEnum, tenantID string) (manager.Settinger, error) {
	switch key {
	case api.DATAVOLUMESSETTINGSKEYENUM_ENABLE_DATA_VOLUMES:
		setting, err := setting_pkg.Get[*EnableDataVolumes](tenantID)
		if err != nil {
			return nil, err
		}
		return setting, nil
	default:
		return nil, fmt.Errorf("%s: %s", ErrUnknownSettingKey, key)
	}
}

func (s *SettingsManager) UpdateSettingByKey(ctx context.Context, key api.DatavolumesSettingsKeyEnum, tenantID string, updateRequest api.UpdateDatavolumesSettingByKeyRequest) error {
	switch key {
	case api.DATAVOLUMESSETTINGSKEYENUM_ENABLE_DATA_VOLUMES:
		if updateRequest.Enabled != nil {
			settingValue := EnableDataVolumes{
				Enabled: *updateRequest.Enabled,
			}
			return setting_pkg.Update[*EnableDataVolumes](tenantID, settingValue).UpdatedByUser().Execute()
		}
		return errors.New("enabled is required")
	default:
		return fmt.Errorf("%s: %s", ErrUnknownSettingKey, key)
	}
}

func (s *SettingsManager) TransformSettingsToResponse(settingsList []manager.Settinger) *api.DatavolumesSettingsGetResponse {
	response := &api.DatavolumesSettingsGetResponse{}

	for _, setting := range settingsList {
		switch v := setting.(type) {
		case *EnableDataVolumes:
			response.EnableDataVolumes = transformEnableDataVolumes(v)
		}
	}

	return response
}

func (s *SettingsManager) TransformSettingByKeyToResponse(setting manager.Settinger) (api.GetEnableDataVolumes, error) {
	switch v := setting.(type) {
	case *EnableDataVolumes:
		return transformEnableDataVolumes(v), nil
	default:
		return api.GetEnableDataVolumes{}, fmt.Errorf("%s", ErrUnknownSettingKey)
	}
}

func transformEnableDataVolumes(v *EnableDataVolumes) api.GetEnableDataVolumes {
	return api.GetEnableDataVolumes{
		Value: api.GetEnableDataVolumesValue{
			Enabled: &v.Enabled,
		},
		DisplayName: v.DisplayName(),
		Description: v.Description(),
		Stage:       string(v.Stage()),
	}
}
