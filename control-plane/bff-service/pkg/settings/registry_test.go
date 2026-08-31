/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"testing"

	"github.com/run-ai/runai/control-plane/bff-service/internal/generated/pkg/api"
	ui_workloads "github.com/run-ai/runai/control-plane/bff-service/pkg/settings/ui-workloads"
	"github.com/run-ai/runai/control-plane/bff-service/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/setting/manager"
	"github.com/stretchr/testify/assert"
)

func TestSettingRegistry_Contains_Logo_Setting(t *testing.T) {
	logoHandler, exists := settingRegistry[string(api.BRANDINGSETTINGSKEYENUM_LOGO)]
	assert.True(t, exists, "settingRegistry should contain BRANDINGSETTINGSKEYENUM_LOGO")
	assert.NotNil(t, logoHandler)
	assert.NotNil(t, logoHandler.get, "logoHandler should have a get function")
	assert.NotNil(t, logoHandler.update, "logoHandler should have an update function")
}

func TestTransformerRegistry_Contains_Branding(t *testing.T) {
	transformer, exists := transformerRegistry[utils.BrandingCategory]
	assert.True(t, exists, "transformerRegistry should contain BrandingCategory")
	assert.NotNil(t, transformer, "transformer function should not be nil")
}

func TestSettingRegistry_Structure(t *testing.T) {
	assert.Len(t, settingRegistry, 7)
	for key, handler := range settingRegistry {
		t.Run(key, func(t *testing.T) {
			assert.NotNil(t, handler, "handler should not be nil for key: %s", key)
			assert.NotNil(t, handler.get, "get function should not be nil for key: %s", key)
			assert.NotNil(t, handler.update, "update function should not be nil for key: %s", key)
		})
	}
}

func TestSettingRegistry_Contains_AdminMessage_Setting(t *testing.T) {
	adminMessageHandler, exists := settingRegistry[string(api.NOTIFICATIONSSETTINGSKEYENUM_ADMIN_MESSAGE)]
	assert.True(t, exists, "settingRegistry should contain NOTIFICATIONSSETTINGSKEYENUM_ADMIN_MESSAGE")
	assert.NotNil(t, adminMessageHandler)
	assert.NotNil(t, adminMessageHandler.get, "adminMessageHandler should have a get function")
	assert.NotNil(t, adminMessageHandler.update, "adminMessageHandler should have an update function")
}

func TestSettingRegistry_Contains_EnableNIM_Setting(t *testing.T) {
	enableNIMHandler, exists := settingRegistry[string(api.WORKLOADSSETTINGSKEYENUM_ENABLE_NIM)]
	assert.True(t, exists, "settingRegistry should contain WORKLOADSSETTINGSKEYENUM_ENABLE_NIM")
	assert.NotNil(t, enableNIMHandler)
	assert.NotNil(t, enableNIMHandler.get, "enableNIMHandler should have a get function")
	assert.NotNil(t, enableNIMHandler.update, "enableNIMHandler should have an update function")
}

func TestSettingRegistry_Contains_DeletedWorkloadsView(t *testing.T) {
	deletedWorkloadsViewHandler, exists := settingRegistry[string(api.WORKLOADSSETTINGSKEYENUM_DELETED_WORKLOADS_VIEW)]
	assert.True(t, exists, "settingRegistry should contain WORKLOADSSETTINGSKEYENUM_DELETED_WORKLOADS_VIEW")
	assert.NotNil(t, deletedWorkloadsViewHandler)
	assert.NotNil(t, deletedWorkloadsViewHandler.get, "deletedWorkloadsViewHandler should have a get function")
	assert.NotNil(t, deletedWorkloadsViewHandler.update, "deletedWorkloadsViewHandler should have an update function")
}

func TestSettingRegistry_Contains_AllowGPT2Model(t *testing.T) {
	allowGPT2ModelHandler, exists := settingRegistry[string(api.WORKLOADSSETTINGSKEYENUM_ALLOW_GPT2_MODEL)]
	assert.True(t, exists, "settingRegistry should contain WORKLOADSSETTINGSKEYENUM_ALLOW_GPT2_MODEL")
	assert.NotNil(t, allowGPT2ModelHandler)
	assert.NotNil(t, allowGPT2ModelHandler.get, "allowGPT2ModelHandler should have a get function")
	assert.NotNil(t, allowGPT2ModelHandler.update, "allowGPT2ModelHandler should have an update function")
}

func TestSettingRegistry_Contains_EnableModelCatalog(t *testing.T) {
	enableModelCatalogHandler, exists := settingRegistry[string(api.WORKLOADSSETTINGSKEYENUM_ENABLE_MODEL_CATALOG)]
	assert.True(t, exists, "settingRegistry should contain WORKLOADSSETTINGSKEYENUM_ENABLE_MODEL_CATALOG")
	assert.NotNil(t, enableModelCatalogHandler)
	assert.NotNil(t, enableModelCatalogHandler.get, "enableModelCatalogHandler should have a get function")
	assert.NotNil(t, enableModelCatalogHandler.update, "enableModelCatalogHandler should have an update function")
}

func TestTransformerRegistry_Contains_Notifications(t *testing.T) {
	transformer, exists := transformerRegistry[utils.NotificationsCategory]

	assert.True(t, exists, "transformerRegistry should contain NotificationsCategory")
	assert.NotNil(t, transformer, "transformer function should not be nil")
}

func TestTransformerRegistry_Contains_Workloads(t *testing.T) {
	transformer, exists := transformerRegistry[utils.UIWorkloadsCategory]

	assert.True(t, exists, "transformerRegistry should contain WorkloadsCategory")
	assert.NotNil(t, transformer, "transformer function should not be nil")
}

func TestTransformerRegistry_Structure(t *testing.T) {
	assert.Len(t, transformerRegistry, 4)
	for category, transformer := range transformerRegistry {
		t.Run(string(category), func(t *testing.T) {
			assert.NotNil(t, transformer, "transformer should not be nil for category: %s", category)
		})
	}
}

func TestTransformToWorkloadsSettingsResponse_EmptyList(t *testing.T) {
	settingsList := []manager.Settinger{}

	result := ui_workloads.TransformToWorkloadsSettingsResponse(settingsList)

	assert.NotNil(t, result)
	resp, ok := result.(*api.WorkloadsSettingsGetAllResponse)
	assert.True(t, ok, "result should be of type *api.WorkloadsSettingsGetAllResponse")
	assert.NotNil(t, resp)
	assert.Nil(t, resp.DeletedWorkloadsView.Value.Enabled)
	assert.Nil(t, resp.AllowGPT2Model.Value.Enabled)
}

func TestTransformToWorkloadsSettingsResponse_WithEnableNIM(t *testing.T) {
	tests := []struct {
		name         string
		settingsList []manager.Settinger
		wantEnabled  bool
	}{
		{
			name: "enable NIM - enabled",
			settingsList: []manager.Settinger{
				&ui_workloads.EnableNIM{
					Enabled: true,
				},
			},
			wantEnabled: true,
		},
		{
			name: "enable NIM - disabled",
			settingsList: []manager.Settinger{
				&ui_workloads.EnableNIM{
					Enabled: false,
				},
			},
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ui_workloads.TransformToWorkloadsSettingsResponse(tt.settingsList)

			assert.NotNil(t, result)
			resp, ok := result.(*api.WorkloadsSettingsGetAllResponse)
			assert.True(t, ok, "result should be of type *api.WorkloadsSettingsGetAllResponse")
			assert.NotNil(t, resp)

			assert.NotNil(t, resp.EnableNim.Value.Enabled)
			assert.Equal(t, tt.wantEnabled, *resp.EnableNim.Value.Enabled)
		})
	}
}

func TestTransformToWorkloadsSettingsResponse_WithDeletedWorkloadsView(t *testing.T) {
	tests := []struct {
		name         string
		settingsList []manager.Settinger
		wantEnabled  bool
	}{
		{
			name: "deleted workloads view - enabled",
			settingsList: []manager.Settinger{
				&ui_workloads.DeletedWorkloadsView{
					Enabled: true,
				},
			},
			wantEnabled: true,
		},
		{
			name: "deleted workloads view - disabled",
			settingsList: []manager.Settinger{
				&ui_workloads.DeletedWorkloadsView{
					Enabled: false,
				},
			},
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ui_workloads.TransformToWorkloadsSettingsResponse(tt.settingsList)

			assert.NotNil(t, result)
			resp, ok := result.(*api.WorkloadsSettingsGetAllResponse)
			assert.True(t, ok, "result should be of type *api.WorkloadsSettingsGetAllResponse")
			assert.NotNil(t, resp)

			assert.NotNil(t, resp.DeletedWorkloadsView.Value.Enabled)
			assert.Equal(t, tt.wantEnabled, *resp.DeletedWorkloadsView.Value.Enabled)
		})
	}
}

func TestTransformToWorkloadsSettingsResponse_WithAllowGPT2Model(t *testing.T) {
	tests := []struct {
		name         string
		settingsList []manager.Settinger
		wantEnabled  bool
	}{
		{
			name: "allow GPT-2 model - enabled",
			settingsList: []manager.Settinger{
				&ui_workloads.AllowGPT2Model{
					Enabled: true,
				},
			},
			wantEnabled: true,
		},
		{
			name: "allow GPT-2 model - disabled",
			settingsList: []manager.Settinger{
				&ui_workloads.AllowGPT2Model{
					Enabled: false,
				},
			},
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ui_workloads.TransformToWorkloadsSettingsResponse(tt.settingsList)

			assert.NotNil(t, result)
			resp, ok := result.(*api.WorkloadsSettingsGetAllResponse)
			assert.True(t, ok, "result should be of type *api.WorkloadsSettingsGetAllResponse")
			assert.NotNil(t, resp)

			assert.NotNil(t, resp.AllowGPT2Model.Value.Enabled)
			assert.Equal(t, tt.wantEnabled, *resp.AllowGPT2Model.Value.Enabled)
		})
	}
}

func TestTransformToWorkloadsSettingsResponse_WithEnableModelCatalog(t *testing.T) {
	tests := []struct {
		name         string
		settingsList []manager.Settinger
		wantEnabled  bool
	}{
		{
			name: "enable model catalog - enabled",
			settingsList: []manager.Settinger{
				&ui_workloads.EnableModelCatalog{
					Enabled: true,
				},
			},
			wantEnabled: true,
		},
		{
			name: "enable model catalog - disabled",
			settingsList: []manager.Settinger{
				&ui_workloads.EnableModelCatalog{
					Enabled: false,
				},
			},
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ui_workloads.TransformToWorkloadsSettingsResponse(tt.settingsList)

			assert.NotNil(t, result)
			resp, ok := result.(*api.WorkloadsSettingsGetAllResponse)
			assert.True(t, ok, "result should be of type *api.WorkloadsSettingsGetAllResponse")
			assert.NotNil(t, resp)

			assert.NotNil(t, resp.EnableModelCatalog.Value.Enabled)
			assert.Equal(t, tt.wantEnabled, *resp.EnableModelCatalog.Value.Enabled)
		})
	}
}

func TestTransformToWorkloadsSettingsResponse_WithMultipleSettings(t *testing.T) {
	settingsList := []manager.Settinger{
		&ui_workloads.EnableNIM{
			Enabled: true,
		},
		&ui_workloads.DeletedWorkloadsView{
			Enabled: false,
		},
		&ui_workloads.AllowGPT2Model{
			Enabled: true,
		},
		&ui_workloads.EnableModelCatalog{
			Enabled: true,
		},
	}

	result := ui_workloads.TransformToWorkloadsSettingsResponse(settingsList)
	assert.NotNil(t, result)
	resp, ok := result.(*api.WorkloadsSettingsGetAllResponse)
	assert.True(t, ok, "result should be of type *api.WorkloadsSettingsGetAllResponse")
	assert.NotNil(t, resp)

	assert.NotNil(t, resp.EnableNim.Value.Enabled)
	assert.True(t, *resp.EnableNim.Value.Enabled)

	assert.NotNil(t, resp.DeletedWorkloadsView.Value.Enabled)
	assert.False(t, *resp.DeletedWorkloadsView.Value.Enabled)

	assert.NotNil(t, resp.AllowGPT2Model.Value.Enabled)
	assert.True(t, *resp.AllowGPT2Model.Value.Enabled)

	assert.NotNil(t, resp.EnableModelCatalog.Value.Enabled)
	assert.True(t, *resp.EnableModelCatalog.Value.Enabled)
}
