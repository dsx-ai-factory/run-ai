/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting"
	"github.com/run-ai/runai/runai-common-packages/setting/manager"
)

const (
	ErrUnknownSettingKey  = "unknown setting key"
	AuthorizationCategory = "authorization"
)

// SettingsManager handles retrieval, update, and transformation of authorization settings.
type SettingsManager struct {
}

// NewSettingsManager creates a new SettingsManager.
func NewSettingsManager() *SettingsManager {
	return &SettingsManager{}
}

func (s *SettingsManager) GetSettingByKey(key api.AuthorizationSettingsKeyEnum, tenantID string) (manager.Settinger, error) {
	switch key {
	case api.AUTHORIZATIONSETTINGSKEYENUM_ENABLE_SCOPED_AUTHORIZATION:
		setting, err := setting_pkg.Get[*EnableScopedAuthorization](tenantID)
		if err != nil {
			return nil, err
		}
		return setting, nil
	case api.AUTHORIZATIONSETTINGSKEYENUM_TENANT_IS_CSP_MANAGED:
		setting, err := setting_pkg.Get[*TenantIsCspManaged](tenantID)
		if err != nil {
			return nil, err
		}
		return setting, nil
	default:
		return nil, fmt.Errorf("%s: %s", ErrUnknownSettingKey, key)
	}
}

func (s *SettingsManager) UpdateSettingByKey(ctx context.Context, key api.AuthorizationSettingsKeyEnum, tenantID string, updateRequest api.UpdateAuthorizationSettingByKeyRequest) error {
	if updateRequest.Enabled == nil {
		return errors.New("enabled is required")
	}
	switch key {
	case api.AUTHORIZATIONSETTINGSKEYENUM_ENABLE_SCOPED_AUTHORIZATION:
		settingValue := EnableScopedAuthorization{Enabled: *updateRequest.Enabled}
		return setting_pkg.Update[*EnableScopedAuthorization](tenantID, settingValue).UpdatedByUser().Execute()
	case api.AUTHORIZATIONSETTINGSKEYENUM_TENANT_IS_CSP_MANAGED:
		settingValue := TenantIsCspManaged{Enabled: *updateRequest.Enabled}
		return setting_pkg.Update[*TenantIsCspManaged](tenantID, settingValue).UpdatedByUser().Execute()
	default:
		return fmt.Errorf("%s: %s", ErrUnknownSettingKey, key)
	}
}

func (s *SettingsManager) TransformSettingsToResponse(settingsList []manager.Settinger) *api.AuthorizationSettingsGetResponse {
	response := &api.AuthorizationSettingsGetResponse{}

	for _, setting := range settingsList {
		switch v := setting.(type) {
		case *EnableScopedAuthorization:
			response.EnableScopedAuthorization = api.EnableScopedAuthorizationSetting{Enabled: &v.Enabled}
		case *TenantIsCspManaged:
			response.TenantIsCspManaged = api.TenantIsCspManagedSetting{Enabled: &v.Enabled}
		}
	}

	return response
}
