/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/internal/generated/pkg/api"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/manager"
)

func TransformToWorkloadsSettingsResponse(settingsList []setting_pkg.Settinger) any {
	response := &api.WorkloadsSettingsGetAllResponse{}

	for _, s := range settingsList {
		switch v := s.(type) {
		case *EnableNIM:
			response.EnableNim = api.GetEnableNim{
				Value: api.GetEnableNimValue{
					Enabled: &v.Enabled,
				},
				DisplayName: v.DisplayName(),
				Description: v.Description(),
			}
		case *DeletedWorkloadsView:
			response.DeletedWorkloadsView = api.GetDeletedWorkloadsView{
				Value: api.GetDeletedWorkloadsViewValue{
					Enabled: &v.Enabled,
				},
				DisplayName: v.DisplayName(),
				Description: v.Description(),
			}
		case *AllowGPT2Model:
			response.AllowGPT2Model = api.GetAllowGpt2Model{
				Value: api.GetAllowGpt2ModelValue{
					Enabled: &v.Enabled,
				},
				DisplayName: v.DisplayName(),
				Description: v.Description(),
				Stage:       string(v.Stage()),
			}
		case *EnableModelCatalog:
			response.EnableModelCatalog = api.GetEnableModelCatalog{
				Value: api.GetEnableModelCatalogValue{
					Enabled: &v.Enabled,
				},
				DisplayName: v.DisplayName(),
				Description: v.Description(),
				Stage:       string(v.Stage()),
			}
		}
	}
	return response
}
