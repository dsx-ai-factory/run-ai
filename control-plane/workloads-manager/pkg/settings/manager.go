/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/run-ai/runai/control-plane/workloads-manager/internal/generated/pkg/api"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting"
	"github.com/run-ai/runai/runai-common-packages/setting/manager"
)

const (
	ErrUnknownSettingKey = "unknown setting key"
	WorkloadsCategory    = "workloads"
)

type SettingsManager struct {
}

func NewSettingsManager() *SettingsManager {
	return &SettingsManager{}
}

func (s *SettingsManager) GetSettingByKey(key api.WorkloadsSettingsKeyEnum, tenantID string) (manager.Settinger, error) {
	switch key {
	case api.WORKLOADSSETTINGSKEYENUM_YAML_SUBMISSION:
		setting, err := setting_pkg.Get[*YamlSubmission](tenantID)
		if err != nil {
			return nil, err
		}
		return setting, nil
	default:
		return nil, fmt.Errorf("%s: %s", ErrUnknownSettingKey, key)
	}
}

// UpdateSettingByKey - Update workloads setting by key
// Currently the only setting is yamlSubmission, when we'll add more settings, we'll need to update the value type, by setting the anyOf in the swagger spec
func (s *SettingsManager) UpdateSettingByKey(ctx context.Context, key api.WorkloadsSettingsKeyEnum, tenantID string, updateRequest api.YamlSubmissionSetting) error {
	switch key {
	case api.WORKLOADSSETTINGSKEYENUM_YAML_SUBMISSION:
		if updateRequest.Enabled != nil {
			settingValue := YamlSubmission{
				Enabled: *updateRequest.Enabled,
			}
			return setting_pkg.Update[*YamlSubmission](tenantID, settingValue).UpdatedByUser().Execute()
		}
		return errors.New("enabled is required")
	default:
		return nil
	}
}

func (s *SettingsManager) TransformSettingsToResponse(settingsList []manager.Settinger) *api.WorkloadsSettingsGetResponse {
	response := &api.WorkloadsSettingsGetResponse{}

	for _, setting := range settingsList {
		switch v := setting.(type) {
		case *YamlSubmission:
			response.YamlSubmission = transformYamlSubmission(v)
		}
	}

	return response
}

func (s *SettingsManager) TransformSettingByKeyToResponse(setting manager.Settinger) (api.GetYamlSubmission, error) {
	switch v := setting.(type) {
	case *YamlSubmission:
		return transformYamlSubmission(v), nil
	default:
		return api.GetYamlSubmission{}, fmt.Errorf("%s", ErrUnknownSettingKey)
	}
}

func transformYamlSubmission(v *YamlSubmission) api.GetYamlSubmission {
	return api.GetYamlSubmission{
		Value: api.GetYamlSubmissionValue{
			Enabled: &v.Enabled,
		},
		DisplayName: v.DisplayName(),
		Description: v.Description(),
		Stage:       string(v.Stage()),
	}
}
