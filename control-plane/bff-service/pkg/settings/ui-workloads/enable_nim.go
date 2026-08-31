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

type EnableNIM struct {
	Enabled bool `json:"enabled"`
}

func NewEnableNIM() setting_pkg.SettingerWithTransform {
	return &EnableNIM{}
}

func (e *EnableNIM) Category() string {
	return string(utils.UIWorkloadsCategory)
}

func (e *EnableNIM) Description() string {
	return "Enable instant deployment of NIM models via the workload submission form. " +
		"For more information, see the <a href='https://run-ai-docs.nvidia.com/self-hosted/workloads-in-nvidia-run-ai/using-inference/nim-inference' target='_blank'>Deploy Inference Workloads with NVIDIA NIM</a>"
}

func (e *EnableNIM) DisplayName() string {
	return "NIM models"
}

func (e *EnableNIM) Key() string {
	return string(api.WORKLOADSSETTINGSKEYENUM_ENABLE_NIM)
}

func (e *EnableNIM) SetDefault() {
	e.Enabled = true
}

func (e *EnableNIM) Stage() setting_pkg.Stage {
	return setting_pkg.GA
}

func (e *EnableNIM) Validate(value interface{}) bool {
	return true
}

func (e *EnableNIM) TransformForMigration(value interface{}) interface{} {
	settingValue, ok := value.(api.GetEnableNim)
	if !ok {
		return value
	}
	return EnableNIM{
		Enabled: *settingValue.Value.Enabled,
	}
}

func (e *EnableNIM) TransformToResponse() interface{} {
	return api.GetEnableNim{
		Value: api.GetEnableNimValue{
			Enabled: &e.Enabled,
		},
		DisplayName: e.DisplayName(),
		Description: e.Description(),
	}
}

func GetEnableNIM(tenantID string) (setting_pkg.Settinger, error) {
	return setting.Get[*EnableNIM](tenantID)
}

func UpdateEnableNIM(tenantID string, updateRequest any) error {
	req, ok := updateRequest.(api.UpdateWorkloadsSettingByKeyRequest)
	if !ok {
		return utils.ErrInvalidInput
	}
	if req.Enabled == nil {
		return utils.ErrInvalidInput
	}
	settingValue := EnableNIM{
		Enabled: *req.Enabled,
	}
	return setting.Update[*EnableNIM](tenantID, settingValue).UpdatedByUser().Execute()
}
