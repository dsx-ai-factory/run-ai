/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/internal/generated/pkg/api"
	authMock "github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/authorization/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/settings"
	ui_workloads "github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/settings/ui-workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	settingPkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	settingPkgManager "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/manager"
	settingPkgMocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

func initWorkloadsSettingsServiceTests(t *testing.T) (settingPkgMock *settingPkgMocks.SettingsManager, testSubject *WorkloadsSettingsAPIService, authManagerMock *authMock.AuthManagerInterface, ctx context.Context) {
	authManagerMock = authMock.NewAuthManagerInterface(t)
	settingPkgManagerMock := settingPkgMocks.NewSettingsManager(t)
	settingPkg.SetMock(settingPkgManagerMock)

	ctx = context.WithValue(context.Background(), "AuthRealm", realmName)
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantID)

	testSubject = &WorkloadsSettingsAPIService{
		settingManager: *settings.NewSettingsManager(),
		settingsHelper: &SettingsHelper{
			authManager: authManagerMock,
		},
	}

	return settingPkgManagerMock, testSubject, authManagerMock, ctx
}

func TestWorkloadsSettingsAPIService_GetWorkloadsSettings_Success(t *testing.T) {
	settingPkgManagerMock, testSubject, authMock, ctx := initWorkloadsSettingsServiceTests(t)

	authMock.On("GetTenantID", ctx, "").Return(ptr.To(tenantID), nil)
	settingPkgManagerMock.On("GetAll", mock.Anything, tenantID, string(utils.UIWorkloadsCategory)).Return(getMockedWorkloadsSettings(), nil)

	response, err := testSubject.GetWorkloadsSettings(ctx, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	expected := &api.WorkloadsSettingsGetAllResponse{
		EnableNim: api.GetEnableNim{
			Value: api.GetEnableNimValue{
				Enabled: ptr.To(true),
			},
			DisplayName: "NIM models",
			Description: "Enable instant deployment of NIM models via the workload submission form. For more information, see the <a data-doc-key='nimInference' target='_blank'>Deploy Inference Workloads with NVIDIA NIM</a>",
		},
	}
	assert.Equal(t, expected, response.Body)
}

func TestWorkloadsSettingsAPIService_GetWorkloadsSettings_NilTenantID(t *testing.T) {
	_, testSubject, authMock, ctx := initWorkloadsSettingsServiceTests(t)

	authMock.On("GetTenantID", ctx, "").Return(nil, nil)

	response, err := testSubject.GetWorkloadsSettings(ctx, "")
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	utilsErr, ok := err.(utils.Error)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, utilsErr.Code)
	assert.Contains(t, utilsErr.Message, "failed to get tenant ID")
}

func TestWorkloadsSettingsAPIService_GetWorkloadsSettings_AuthorizationError(t *testing.T) {
	_, testSubject, authMock, ctx := initWorkloadsSettingsServiceTests(t)

	authMock.On("GetTenantID", ctx, "").Return(nil, errors.New("Unauthorized"))

	response, err := testSubject.GetWorkloadsSettings(ctx, "")
	assert.Error(t, err)
	assert.Equal(t, http.StatusForbidden, response.Code)
}

func TestWorkloadsSettingsAPIService_GetWorkloadsSettingByKey_Success(t *testing.T) {
	tests := []struct {
		name       string
		settingKey api.WorkloadsSettingsKeyEnum
		mockReturn settingPkgManager.Settinger
	}{
		{
			name:       "enableNIM setting",
			settingKey: api.WORKLOADSSETTINGSKEYENUM_ENABLE_NIM,
			mockReturn: &ui_workloads.EnableNIM{Enabled: true},
		},
		{
			name:       "deletedWorkloadsView setting",
			settingKey: api.WORKLOADSSETTINGSKEYENUM_DELETED_WORKLOADS_VIEW,
			mockReturn: &ui_workloads.DeletedWorkloadsView{Enabled: true},
		},
		{
			name:       "enableModelCatalog setting",
			settingKey: api.WORKLOADSSETTINGSKEYENUM_ENABLE_MODEL_CATALOG,
			mockReturn: &ui_workloads.EnableModelCatalog{Enabled: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settingPkgManagerMock, testSubject, authMock, ctx := initWorkloadsSettingsServiceTests(t)

			authMock.On("GetTenantID", ctx, "").Return(ptr.To(tenantID), nil)
			settingPkgManagerMock.On("Get", mock.Anything, tenantID, mock.Anything).Return(tt.mockReturn, nil)

			response, err := testSubject.GetWorkloadsSettingByKey(ctx, tt.settingKey, "")
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, response.Code)
		})
	}
}

func TestWorkloadsSettingsAPIService_GetWorkloadsSettingByKey_NilTenantID(t *testing.T) {
	tests := []struct {
		name       string
		settingKey api.WorkloadsSettingsKeyEnum
	}{
		{
			name:       "enableNIM setting",
			settingKey: api.WORKLOADSSETTINGSKEYENUM_ENABLE_NIM,
		},
		{
			name:       "deletedWorkloadsView setting",
			settingKey: api.WORKLOADSSETTINGSKEYENUM_DELETED_WORKLOADS_VIEW,
		},
		{
			name:       "enableModelCatalog setting",
			settingKey: api.WORKLOADSSETTINGSKEYENUM_ENABLE_MODEL_CATALOG,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, testSubject, authMock, ctx := initWorkloadsSettingsServiceTests(t)

			authMock.On("GetTenantID", ctx, "").Return(nil, nil)

			response, err := testSubject.GetWorkloadsSettingByKey(ctx, tt.settingKey, "")
			assert.Error(t, err)
			assert.Equal(t, http.StatusInternalServerError, response.Code)

			utilsErr, ok := err.(utils.Error)
			assert.True(t, ok)
			assert.Equal(t, http.StatusInternalServerError, utilsErr.Code)
			assert.Contains(t, utilsErr.Message, "failed to get tenant ID")
		})
	}
}

func TestWorkloadsSettingsAPIService_UpdateWorkloadsSettingByKey_Success(t *testing.T) {
	tests := []struct {
		name       string
		settingKey api.WorkloadsSettingsKeyEnum
		mockType   string
	}{
		{
			name:       "enableNIM setting",
			settingKey: api.WORKLOADSSETTINGSKEYENUM_ENABLE_NIM,
			mockType:   "workloads.EnableNIM",
		},
		{
			name:       "deletedWorkloadsView setting",
			settingKey: api.WORKLOADSSETTINGSKEYENUM_DELETED_WORKLOADS_VIEW,
			mockType:   "workloads.DeletedWorkloadsView",
		},
		{
			name:       "enableModelCatalog setting",
			settingKey: api.WORKLOADSSETTINGSKEYENUM_ENABLE_MODEL_CATALOG,
			mockType:   "workloads.EnableModelCatalog",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settingPkgManagerMock, testSubject, authMock, ctx := initWorkloadsSettingsServiceTests(t)

			authMock.On("GetTenantID", ctx, "").Return(ptr.To(tenantID), nil)
			settingPkgManagerMock.On("Update", mock.Anything, tenantID, string(tt.settingKey), true, mock.AnythingOfType(tt.mockType)).Return(nil)

			updateRequest := api.UpdateWorkloadsSettingByKeyRequest{Enabled: ptr.To(true)}
			response, err := testSubject.UpdateWorkloadsSettingByKey(ctx, tt.settingKey, updateRequest, "")
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, response.Code)
		})
	}
}

func TestWorkloadsSettingsAPIService_UpdateWorkloadsSettingByKey_NilTenantID(t *testing.T) {
	tests := []struct {
		name       string
		settingKey api.WorkloadsSettingsKeyEnum
	}{
		{
			name:       "enableNIM setting",
			settingKey: api.WORKLOADSSETTINGSKEYENUM_ENABLE_NIM,
		},
		{
			name:       "deletedWorkloadsView setting",
			settingKey: api.WORKLOADSSETTINGSKEYENUM_DELETED_WORKLOADS_VIEW,
		},
		{
			name:       "enableModelCatalog setting",
			settingKey: api.WORKLOADSSETTINGSKEYENUM_ENABLE_MODEL_CATALOG,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, testSubject, authMock, ctx := initWorkloadsSettingsServiceTests(t)

			authMock.On("GetTenantID", ctx, "").Return(nil, nil)

			updateRequest := api.UpdateWorkloadsSettingByKeyRequest{Enabled: ptr.To(true)}
			response, err := testSubject.UpdateWorkloadsSettingByKey(ctx, tt.settingKey, updateRequest, "")
			assert.Error(t, err)
			assert.Equal(t, http.StatusInternalServerError, response.Code)

			utilsErr, ok := err.(utils.Error)
			assert.True(t, ok)
			assert.Equal(t, http.StatusInternalServerError, utilsErr.Code)
			assert.Contains(t, utilsErr.Message, "failed to get tenant ID")
		})
	}
}

func getMockedWorkloadsSettings() []settingPkgManager.Settinger {
	return []settingPkgManager.Settinger{
		&ui_workloads.EnableNIM{
			Enabled: true,
		},
	}
}
