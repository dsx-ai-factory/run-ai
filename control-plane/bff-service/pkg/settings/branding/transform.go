/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package branding

import (
	"github.com/run-ai/runai/control-plane/bff-service/internal/generated/pkg/api"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting/manager"
)

func TransformToBrandingSettingsResponse(settingsList []setting_pkg.Settinger) any {
	response := &api.BrandingSettingsGetAllResponse{}

	for _, s := range settingsList {
		if v, ok := s.(*Logo); ok {
			response.Logo = api.GetLogo{
				Value:       api.GetLogoValue{Logo: reconstructDataURL(v.Logo)},
				DisplayName: v.DisplayName(),
				Description: v.Description(),
			}
		}
	}
	return response
}
