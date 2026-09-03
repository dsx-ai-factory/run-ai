/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"context"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/internal/generated/pkg/api"
	ui_workloads "github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/settings/ui-workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/manager"
	"github.com/stretchr/testify/assert"
)

func TestNewSettingsManager(t *testing.T) {
	manager := NewSettingsManager()
	assert.NotNil(t, manager)
}

func TestSettingsManager_GetSettingByKey(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		tenantID    string
		wantErr     bool
		errContains string
	}{
		{
			name:        "unknown setting key",
			key:         "unknown-key",
			tenantID:    "tenant-1",
			wantErr:     true,
			errContains: ErrUnknownSettingKey,
		},
		{
			name:     "valid key exists in registry",
			key:      string(api.NOTIFICATIONSSETTINGSKEYENUM_ADMIN_MESSAGE),
			tenantID: "tenant-1",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewSettingsManager()

			if tt.wantErr {
				setting, err := manager.GetSettingByKey(tt.key, tt.tenantID)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, setting)
			} else {
				_, exists := settingRegistry[tt.key]
				assert.True(t, exists, "key should exist in registry")
			}
		})
	}
}

func TestSettingsManager_UpdateSettingByKey(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		tenantID      string
		updateRequest any
		wantErr       bool
		errContains   string
	}{
		{
			name:          "unknown setting key",
			key:           "unknown-key",
			tenantID:      "tenant-1",
			updateRequest: api.UpdateAdminMessage{},
			wantErr:       true,
			errContains:   ErrUnknownSettingKey,
		},
		{
			name:          "valid key - invalid request type",
			key:           string(api.WORKLOADSSETTINGSKEYENUM_ENABLE_NIM),
			tenantID:      "tenant-1",
			updateRequest: "invalid-type",
			wantErr:       true,
			errContains:   "invalid input",
		},
		{
			name:          "valid key - nil enabled field",
			key:           string(api.WORKLOADSSETTINGSKEYENUM_ENABLE_NIM),
			tenantID:      "tenant-1",
			updateRequest: api.UpdateWorkloadsSettingByKeyRequest{Enabled: nil},
			wantErr:       true,
			errContains:   "invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewSettingsManager()
			err := manager.UpdateSettingByKey(context.Background(), tt.key, tt.tenantID, tt.updateRequest)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			}
		})
	}
}

func TestSettingsManager_TransformSettingsToResponse(t *testing.T) {
	tests := []struct {
		name         string
		settingsList []manager.Settinger
		category     utils.CategoryKey
		wantErr      bool
		errContains  string
		validateResp func(*testing.T, any)
	}{
		{
			name:         "unknown category",
			settingsList: []manager.Settinger{},
			category:     utils.CategoryKey("unknown-category"),
			wantErr:      true,
			errContains:  "no transformer found for category",
		},
		{
			name:         "UI Workloads category - empty list",
			settingsList: []manager.Settinger{},
			category:     utils.UIWorkloadsCategory,
			wantErr:      false,
			validateResp: func(t *testing.T, resp any) {
				workloadsResp, ok := resp.(*api.WorkloadsSettingsGetAllResponse)
				assert.True(t, ok)
				assert.NotNil(t, workloadsResp)
			},
		},
		{
			name: "UI Workloads category - with EnableNIM enabled",
			settingsList: []manager.Settinger{
				&ui_workloads.EnableNIM{Enabled: true},
			},
			category: utils.UIWorkloadsCategory,
			wantErr:  false,
			validateResp: func(t *testing.T, resp any) {
				workloadsResp, ok := resp.(*api.WorkloadsSettingsGetAllResponse)
				assert.True(t, ok)
				assert.NotNil(t, workloadsResp)
				assert.NotNil(t, workloadsResp.EnableNim.Value.Enabled)
				assert.True(t, *workloadsResp.EnableNim.Value.Enabled)
			},
		},
		{
			name: "UI Workloads category - with EnableNIM disabled",
			settingsList: []manager.Settinger{
				&ui_workloads.EnableNIM{Enabled: false},
			},
			category: utils.UIWorkloadsCategory,
			wantErr:  false,
			validateResp: func(t *testing.T, resp any) {
				workloadsResp, ok := resp.(*api.WorkloadsSettingsGetAllResponse)
				assert.True(t, ok)
				assert.NotNil(t, workloadsResp)
				assert.NotNil(t, workloadsResp.EnableNim.Value.Enabled)
				assert.False(t, *workloadsResp.EnableNim.Value.Enabled)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewSettingsManager()
			resp, err := manager.TransformSettingsToResponse(tt.settingsList, tt.category)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateResp != nil {
					tt.validateResp(t, resp)
				}
			}
		})
	}
}
