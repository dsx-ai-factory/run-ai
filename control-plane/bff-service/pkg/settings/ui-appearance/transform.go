/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package appearance

import (
	"github.com/run-ai/runai/control-plane/bff-service/internal/generated/pkg/api"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting/manager"
)

func TransformToAppearanceSettingsResponse(settingsList []setting_pkg.Settinger) any {
	response := &api.AppearanceSettingsGetAllResponse{}

	for _, s := range settingsList {
		if v, ok := s.(*BannerConfig); ok {
			response.BannerConfig = api.GetBannerConfig{
				Value: api.GetBannerConfigValue{
					Enabled:         &v.Enabled,
					Text:            &v.Text,
					BackgroundColor: &v.BackgroundColor,
					FontColor:       &v.FontColor,
				},
				DisplayName: v.DisplayName(),
				Description: v.Description(),
			}
		}
	}
	return response
}
