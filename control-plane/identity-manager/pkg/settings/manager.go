/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/manager"
)

const (
	ErrUnknownSettingKey = "unknown setting key"
)

type SettingsManager struct {
	iamServices *iam.IamServiceContainer
}

func NewSettingsManager(iamServices *iam.IamServiceContainer) *SettingsManager {
	return &SettingsManager{
		iamServices: iamServices,
	}
}

func (s *SettingsManager) GetSettingByKey(key api.SettingsKeyEnum, tenantID string) (interface{}, error) {
	switch key {
	case api.SETTINGSKEYENUM_LOGOUT_REDIRECT_URI:
		setting, err := setting_pkg.Get[*LogoutRedirectUri](tenantID)
		if err != nil {
			return nil, err
		}
		return api.GetLogoutRedirectUriSetting{
			Value: api.GetLogoutRedirectUriSettingValue{
				Uri: &setting.Uri,
			},
			DisplayName: setting.DisplayName(),
			Description: setting.Description(),
		}, nil
	case api.SETTINGSKEYENUM_AUTO_REDIRECT_SSO:
		setting, err := setting_pkg.Get[*AutoRedirectSSO](tenantID)
		if err != nil {
			return nil, err
		}
		return api.GetAutoRedirectSsoSetting{
			Value: api.GetAutoRedirectSsoSettingValue{
				Enabled:  &setting.Enabled,
				IdpAlias: setting.IdpAlias,
			},
			DisplayName: setting.DisplayName(),
			Description: setting.Description(),
		}, nil
	case api.SETTINGSKEYENUM_BROWSER_SESSION_TIMEOUT:
		setting, err := setting_pkg.Get[*BrowserSessionTimeout](tenantID)
		if err != nil {
			return nil, err
		}
		return api.GetBrowserSessionTimeoutSetting{
			Value: api.GetBrowserSessionTimeoutSettingValue{
				IdleTimeout: setting.IdleTimeout,
			},
			DisplayName: setting.DisplayName(),
			Description: setting.Description(),
		}, nil
	default:
		return nil, fmt.Errorf("unknown setting key: %s", key)
	}
}

func (s *SettingsManager) UpdateSettingByKey(ctx context.Context, realm string, key api.SettingsKeyEnum, tenantID string, updateRequest api.UpdateSecuritySettingByKeyRequest) error {
	switch key {
	case api.SETTINGSKEYENUM_LOGOUT_REDIRECT_URI:
		if updateRequest.Uri != nil {
			err := s.iamServices.ClientService.UpdatePostLogoutRedirectUris(ctx, realm, utils.RUNAI_ADMIN_UI_CLIENT_ID, *updateRequest.Uri)
			if err != nil {
				return err
			}
			settingValue := LogoutRedirectUri{
				Uri: *updateRequest.Uri,
			}
			return setting_pkg.Update[*LogoutRedirectUri](tenantID, settingValue).UpdatedByUser().Execute()
		}
		return utils.ErrInvalidInput
	case api.SETTINGSKEYENUM_AUTO_REDIRECT_SSO:
		if updateRequest.Enabled != nil {
			settingValue := AutoRedirectSSO{
				Enabled:  *updateRequest.Enabled,
				IdpAlias: updateRequest.IdpAlias,
			}
			if !setting_pkg.Validate[*AutoRedirectSSO](&settingValue) {
				return utils.ErrInvalidInput
			}
			err := s.iamServices.RealmService.UpdateAutoRedirectSSO(ctx, realm, settingValue.IdpAlias, settingValue.Enabled)
			if err != nil {
				return err
			}
			return setting_pkg.Update[*AutoRedirectSSO](tenantID, settingValue).UpdatedByUser().Execute()
		}
		return utils.ErrInvalidInput
	case api.SETTINGSKEYENUM_BROWSER_SESSION_TIMEOUT:
		if updateRequest.IdleTimeout != nil {
			settingValue := BrowserSessionTimeout{
				IdleTimeout: updateRequest.IdleTimeout,
			}
			return setting_pkg.Update[*BrowserSessionTimeout](tenantID, settingValue).UpdatedByUser().Execute()
		}
		return utils.ErrInvalidInput
	default:
		return nil
	}
}

func (s *SettingsManager) TransformSettingsToResponse(settingsList []manager.Settinger) *api.SettingsGetResponse {
	response := &api.SettingsGetResponse{}

	for _, setting := range settingsList {
		displayName := setting.DisplayName()
		description := setting.Description()

		switch v := setting.(type) {
		case *LogoutRedirectUri:
			response.LogoutRedirectUri = api.GetLogoutRedirectUriSetting{
				Value: api.GetLogoutRedirectUriSettingValue{
					Uri: &v.Uri,
				},
				DisplayName: displayName,
				Description: description,
			}
		case *AutoRedirectSSO:
			response.AutoRedirectSSO = api.GetAutoRedirectSsoSetting{
				Value: api.GetAutoRedirectSsoSettingValue{
					Enabled:  &v.Enabled,
					IdpAlias: v.IdpAlias,
				},
				DisplayName: displayName,
				Description: description,
			}
		case *BrowserSessionTimeout:
			response.BrowserSessionTimeout = api.GetBrowserSessionTimeoutSetting{
				Value: api.GetBrowserSessionTimeoutSettingValue{
					IdleTimeout: v.IdleTimeout,
				},
				DisplayName: displayName,
				Description: description,
			}
		}
	}

	return response
}
