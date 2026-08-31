/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"github.com/run-ai/runai/control-plane/bff-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/bff-service/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/setting"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting/manager"
)

type DeletedWorkloadsView struct {
	Enabled bool `json:"enabled"`
}

func NewDeletedWorkloadsView() setting_pkg.SettingerWithTransform {
	return &DeletedWorkloadsView{}
}

func (d *DeletedWorkloadsView) Category() string {
	return string(utils.UIWorkloadsCategory)
}

func (d *DeletedWorkloadsView) Description() string {
	return "Enable the deleted workloads view"
}

func (d *DeletedWorkloadsView) DisplayName() string {
	return "Deleted workloads view"
}

func (d *DeletedWorkloadsView) Key() string {
	return string(api.WORKLOADSSETTINGSKEYENUM_DELETED_WORKLOADS_VIEW)
}

func (d *DeletedWorkloadsView) SetDefault() {
	d.Enabled = true
}

func (d *DeletedWorkloadsView) Stage() setting_pkg.Stage {
	return setting_pkg.GA
}

func (d *DeletedWorkloadsView) Validate(value interface{}) bool {
	return true
}

func (d *DeletedWorkloadsView) TransformForMigration(value interface{}) interface{} {
	settingValue, ok := value.(api.GetDeletedWorkloadsView)
	if !ok {
		return value
	}
	return DeletedWorkloadsView{
		Enabled: *settingValue.Value.Enabled,
	}
}

func (d *DeletedWorkloadsView) TransformToResponse() interface{} {
	return api.GetDeletedWorkloadsView{
		Value: api.GetDeletedWorkloadsViewValue{
			Enabled: &d.Enabled,
		},
		DisplayName: d.DisplayName(),
		Description: d.Description(),
	}
}

func GetEnableDeletedWorkloadsView(tenantID string) (setting_pkg.Settinger, error) {
	return setting.Get[*DeletedWorkloadsView](tenantID)
}

func UpdateEnableDeletedWorkloadsView(tenantID string, updateRequest any) error {
	req, ok := updateRequest.(api.UpdateWorkloadsSettingByKeyRequest)
	if !ok {
		return utils.ErrInvalidInput
	}
	if req.Enabled == nil {
		return utils.ErrInvalidInput
	}
	settingValue := DeletedWorkloadsView{
		Enabled: *req.Enabled,
	}

	return setting.Update[*DeletedWorkloadsView](tenantID, settingValue).UpdatedByUser().Execute()
}
