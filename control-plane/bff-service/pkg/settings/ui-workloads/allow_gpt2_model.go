/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/manager"
)

type AllowGPT2Model struct {
	Enabled bool `json:"enabled"`
}

func NewAllowGPT2Model() setting_pkg.SettingerWithTransform {
	return &AllowGPT2Model{}
}

func (a *AllowGPT2Model) Category() string {
	return string(utils.UIWorkloadsCategory)
}

func (a *AllowGPT2Model) Description() string {
	//nolint:lll // long catalog description string
	return "GPT2 model - present OpenAI's GPT2 model in the catalog in favor of demos. This model's compute requirements are relatively low (CPU resources are sufficient for it). The underlying image of this model also consists of a Chatbot UI"
}

func (a *AllowGPT2Model) DisplayName() string {
	return "Allow GPT-2 model"
}

func (a *AllowGPT2Model) Key() string {
	return string(api.WORKLOADSSETTINGSKEYENUM_ALLOW_GPT2_MODEL)
}

func (a *AllowGPT2Model) SetDefault() {
	a.Enabled = false
}

func (a *AllowGPT2Model) Stage() setting_pkg.Stage {
	return setting_pkg.Alpha
}

func (a *AllowGPT2Model) Validate(value interface{}) bool {
	return true
}

func (a *AllowGPT2Model) TransformForMigration(value interface{}) interface{} {
	settingValue, ok := value.(bool)
	if !ok {
		return value
	}
	return AllowGPT2Model{
		Enabled: settingValue,
	}
}

func (a *AllowGPT2Model) TransformToResponse() interface{} {
	return api.GetAllowGpt2Model{
		Value: api.GetAllowGpt2ModelValue{
			Enabled: &a.Enabled,
		},
		DisplayName: a.DisplayName(),
		Description: a.Description(),
	}
}

func GetAllowGPT2Model(tenantID string) (setting_pkg.Settinger, error) {
	return setting.Get[*AllowGPT2Model](tenantID)
}

func UpdateAllowGPT2Model(tenantID string, updateRequest any) error {
	req, ok := updateRequest.(api.UpdateWorkloadsSettingByKeyRequest)
	if !ok {
		return utils.ErrInvalidInput
	}
	if req.Enabled == nil {
		return utils.ErrInvalidInput
	}
	settingValue := AllowGPT2Model{
		Enabled: *req.Enabled,
	}
	return setting.Update[*AllowGPT2Model](tenantID, settingValue).UpdatedByUser().Execute()
}
