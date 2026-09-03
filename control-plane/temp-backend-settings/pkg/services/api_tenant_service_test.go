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
	"time"

	cluster_openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	clusterMocks "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/business/settings"
	settingsMocks "github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/services/mocks"
	tm_openapi "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client/generated"
	tmMocks "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testTenantID    = 42
	testTenantIDStr = "42"
)

var testCreatedAt = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

func initTenantAPIServiceTests(t *testing.T) (settingsMgrMock *settingsMocks.SettingsManagerInterface, testSubject *TenantAPIService, clusterClientMock *clusterMocks.IClusterClient, tenantMgrMock *tmMocks.ITenantsManager, ctx context.Context) {
	settingsMgrMock = settingsMocks.NewSettingsManagerInterface(t)
	clusterClientMock = clusterMocks.NewIClusterClient(t)
	tenantMgrMock = tmMocks.NewITenantsManager(t)

	ctx = context.WithValue(context.Background(), runaictx.CfTenantId, testTenantIDStr)

	testSubject = &TenantAPIService{
		settingsManager: settingsMgrMock,
		clusterClient:   clusterClientMock,
		tenantMgrClient: tenantMgrMock,
	}

	return settingsMgrMock, testSubject, clusterClientMock, tenantMgrMock, ctx
}

// ==================== GetTenantSettings ====================

func TestTenantAPIService_GetTenantSettings_Success(t *testing.T) {
	settingsMgrMock, testSubject, clusterClientMock, tenantMgrMock, ctx := initTenantAPIServiceTests(t)

	clusterClientMock.On("GetClustersMinimalData", ctx).Return(getMockedClustersMinimalData(), nil)
	tenantMgrMock.On("GetTenantByIdInternal", ctx, int32(testTenantID)).Return(getMockedTenant(), &http.Response{}, nil)
	settingsMgrMock.On("GetAllSettings", ctx, testTenantID, mock.Anything, []string{"2.17.0"}, false).Return(getMockedTenantSettings(), nil)

	response, err := testSubject.GetTenantSettings(ctx, false, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	result, ok := response.Body.([]api.Setting)
	assert.True(t, ok)
	assert.Len(t, result, 1)
	assert.Equal(t, "tenant.enable_policies", result[0].Key)
}

func TestTenantAPIService_GetTenantSettings_MissingTenantIDInContext(t *testing.T) {
	_, testSubject, _, _, _ := initTenantAPIServiceTests(t)

	ctx := context.Background()

	response, err := testSubject.GetTenantSettings(ctx, false, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	body, ok := response.Body.(BackendError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, body.Status)
	assert.Contains(t, body.Description, "invalid tenant id")
	assert.Equal(t, "BackendError", body.ErrorType)
}

func TestTenantAPIService_GetTenantSettings_TenantStaticDataError(t *testing.T) {
	_, testSubject, clusterClientMock, tenantMgrMock, ctx := initTenantAPIServiceTests(t)

	clusterClientMock.On("GetClustersMinimalData", ctx).Return(getMockedClustersMinimalData(), nil)
	tenantMgrMock.On("GetTenantByIdInternal", ctx, int32(testTenantID)).Return(nil, nil, errors.New("tenant not found"))

	response, err := testSubject.GetTenantSettings(ctx, false, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	body, ok := response.Body.(BackendError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, body.Status)
	assert.Contains(t, body.Description, "failed to get tenant data")
	assert.Equal(t, "BackendError", body.ErrorType)
}

func TestTenantAPIService_GetTenantSettings_SettingsManagerError(t *testing.T) {
	settingsMgrMock, testSubject, clusterClientMock, tenantMgrMock, ctx := initTenantAPIServiceTests(t)

	clusterClientMock.On("GetClustersMinimalData", ctx).Return(getMockedClustersMinimalData(), nil)
	tenantMgrMock.On("GetTenantByIdInternal", ctx, int32(testTenantID)).Return(getMockedTenant(), &http.Response{}, nil)
	settingsMgrMock.On("GetAllSettings", ctx, testTenantID, mock.Anything, []string{"2.17.0"}, false).Return(nil, errors.New("database error"))

	response, err := testSubject.GetTenantSettings(ctx, false, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	body, ok := response.Body.(BackendError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, body.Status)
	assert.Contains(t, body.Description, "failed to get tenant settings")
	assert.Equal(t, "BackendError", body.ErrorType)
}

func TestTenantAPIService_GetTenantSettings_ClusterVersionsError_StillSucceeds(t *testing.T) {
	settingsMgrMock, testSubject, clusterClientMock, tenantMgrMock, ctx := initTenantAPIServiceTests(t)

	clusterClientMock.On("GetClustersMinimalData", ctx).Return(nil, errors.New("cluster service unavailable"))
	tenantMgrMock.On("GetTenantByIdInternal", ctx, int32(testTenantID)).Return(getMockedTenant(), &http.Response{}, nil)
	settingsMgrMock.On("GetAllSettings", ctx, testTenantID, mock.Anything, []string{}, false).Return(getMockedTenantSettings(), nil)

	response, err := testSubject.GetTenantSettings(ctx, false, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestTenantAPIService_GetTenantSettings_ExcludeMigrated(t *testing.T) {
	settingsMgrMock, testSubject, clusterClientMock, tenantMgrMock, ctx := initTenantAPIServiceTests(t)

	clusterClientMock.On("GetClustersMinimalData", ctx).Return(getMockedClustersMinimalData(), nil)
	tenantMgrMock.On("GetTenantByIdInternal", ctx, int32(testTenantID)).Return(getMockedTenant(), &http.Response{}, nil)
	settingsMgrMock.On("GetAllSettings", ctx, testTenantID, mock.Anything, []string{"2.17.0"}, true).Return(getMockedTenantSettings(), nil)

	response, err := testSubject.GetTenantSettings(ctx, true, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestTenantAPIService_GetTenantSettings_EmptySettings(t *testing.T) {
	settingsMgrMock, testSubject, clusterClientMock, tenantMgrMock, ctx := initTenantAPIServiceTests(t)

	clusterClientMock.On("GetClustersMinimalData", ctx).Return(getMockedClustersMinimalData(), nil)
	tenantMgrMock.On("GetTenantByIdInternal", ctx, int32(testTenantID)).Return(getMockedTenant(), &http.Response{}, nil)
	settingsMgrMock.On("GetAllSettings", ctx, testTenantID, mock.Anything, []string{"2.17.0"}, false).Return([]api.Setting{}, nil)

	response, err := testSubject.GetTenantSettings(ctx, false, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	result, ok := response.Body.([]api.Setting)
	assert.True(t, ok)
	assert.Empty(t, result)
}

func TestTenantAPIService_GetTenantSettings_ManagementTenant_WithTenantIdParam(t *testing.T) {
	settingsMgrMock, testSubject, clusterClientMock, tenantMgrMock, _ := initTenantAPIServiceTests(t)

	ctx := context.WithValue(context.Background(), runaictx.CfIsManagementTenant, "true")

	clusterClientMock.On("GetClustersMinimalData", ctx).Return(getMockedClustersMinimalData(), nil)
	tenantMgrMock.On("GetTenantByIdInternal", ctx, int32(99)).Return(getMockedTenant(), &http.Response{}, nil)
	settingsMgrMock.On("GetAllSettings", ctx, 99, mock.Anything, []string{"2.17.0"}, false).Return(getMockedTenantSettings(), nil)

	response, err := testSubject.GetTenantSettings(ctx, false, "99")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	result, ok := response.Body.([]api.Setting)
	assert.True(t, ok)
	assert.Len(t, result, 1)
}

func TestTenantAPIService_GetTenantSettings_ManagementTenant_NoParam_FallsBackToContext(t *testing.T) {
	settingsMgrMock, testSubject, clusterClientMock, tenantMgrMock, _ := initTenantAPIServiceTests(t)

	ctx := context.WithValue(context.Background(), runaictx.CfIsManagementTenant, "true")
	ctx = context.WithValue(ctx, runaictx.CfTenantId, testTenantIDStr)

	clusterClientMock.On("GetClustersMinimalData", ctx).Return(getMockedClustersMinimalData(), nil)
	tenantMgrMock.On("GetTenantByIdInternal", ctx, int32(testTenantID)).Return(getMockedTenant(), &http.Response{}, nil)
	settingsMgrMock.On("GetAllSettings", ctx, testTenantID, mock.Anything, []string{"2.17.0"}, false).Return(getMockedTenantSettings(), nil)

	response, err := testSubject.GetTenantSettings(ctx, false, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	result, ok := response.Body.([]api.Setting)
	assert.True(t, ok)
	assert.Len(t, result, 1)
}

func TestTenantAPIService_GetTenantSettings_ManagementTenant_NoParam_NoContext(t *testing.T) {
	_, testSubject, _, _, _ := initTenantAPIServiceTests(t)

	ctx := context.WithValue(context.Background(), runaictx.CfIsManagementTenant, "true")

	response, err := testSubject.GetTenantSettings(ctx, false, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	body, ok := response.Body.(BackendError)
	assert.True(t, ok)
	assert.Contains(t, body.Description, "invalid tenant id")
}

// ==================== GetTenantSettingsByKey ====================

func TestTenantAPIService_GetTenantSettingsByKey_Success(t *testing.T) {
	settingsMgrMock, testSubject, _, tenantMgrMock, ctx := initTenantAPIServiceTests(t)

	tenantMgrMock.On("GetTenantByIdInternal", ctx, int32(testTenantID)).Return(getMockedTenant(), &http.Response{}, nil)
	settingsMgrMock.On("GetAllSettings", ctx, testTenantID, mock.Anything, []string{}, false).Return(getMockedTenantSettings(), nil)

	response, err := testSubject.GetTenantSettingsByKey(ctx, "tenant.enable_policies", "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.NotNil(t, response.Body)
}

func TestTenantAPIService_GetTenantSettingsByKey_NotFound(t *testing.T) {
	settingsMgrMock, testSubject, _, tenantMgrMock, ctx := initTenantAPIServiceTests(t)

	tenantMgrMock.On("GetTenantByIdInternal", ctx, int32(testTenantID)).Return(getMockedTenant(), &http.Response{}, nil)
	settingsMgrMock.On("GetAllSettings", ctx, testTenantID, mock.Anything, []string{}, false).Return(getMockedTenantSettings(), nil)

	response, err := testSubject.GetTenantSettingsByKey(ctx, "non.existent.key", "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.Code)

	body, ok := response.Body.(BackendError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusNotFound, body.Status)
	assert.Contains(t, body.Description, "not found")
	assert.Equal(t, "BackendError", body.ErrorType)
}

func TestTenantAPIService_GetTenantSettingsByKey_MissingTenantIDInContext(t *testing.T) {
	_, testSubject, _, _, _ := initTenantAPIServiceTests(t)

	ctx := context.Background()

	response, err := testSubject.GetTenantSettingsByKey(ctx, "tenant.enable_policies", "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	body, ok := response.Body.(BackendError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, body.Status)
	assert.Contains(t, body.Description, "invalid tenant id")
	assert.Equal(t, "BackendError", body.ErrorType)
}

func TestTenantAPIService_GetTenantSettingsByKey_TenantStaticDataError(t *testing.T) {
	_, testSubject, _, tenantMgrMock, ctx := initTenantAPIServiceTests(t)

	tenantMgrMock.On("GetTenantByIdInternal", ctx, int32(testTenantID)).Return(nil, nil, errors.New("tenant service error"))

	response, err := testSubject.GetTenantSettingsByKey(ctx, "tenant.enable_policies", "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	body, ok := response.Body.(BackendError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, body.Status)
	assert.Contains(t, body.Description, "failed to get tenant data")
	assert.Equal(t, "BackendError", body.ErrorType)
}

func TestTenantAPIService_GetTenantSettingsByKey_SettingsManagerError(t *testing.T) {
	settingsMgrMock, testSubject, _, tenantMgrMock, ctx := initTenantAPIServiceTests(t)

	tenantMgrMock.On("GetTenantByIdInternal", ctx, int32(testTenantID)).Return(getMockedTenant(), &http.Response{}, nil)
	settingsMgrMock.On("GetAllSettings", ctx, testTenantID, mock.Anything, []string{}, false).Return(nil, errors.New("database error"))

	response, err := testSubject.GetTenantSettingsByKey(ctx, "tenant.enable_policies", "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	body, ok := response.Body.(BackendError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, body.Status)
	assert.Contains(t, body.Description, "failed to get tenant settings")
	assert.Equal(t, "BackendError", body.ErrorType)
}

func TestTenantAPIService_GetTenantSettingsByKey_EmptySettings(t *testing.T) {
	settingsMgrMock, testSubject, _, tenantMgrMock, ctx := initTenantAPIServiceTests(t)

	tenantMgrMock.On("GetTenantByIdInternal", ctx, int32(testTenantID)).Return(getMockedTenant(), &http.Response{}, nil)
	settingsMgrMock.On("GetAllSettings", ctx, testTenantID, mock.Anything, []string{}, false).Return([]api.Setting{}, nil)

	response, err := testSubject.GetTenantSettingsByKey(ctx, "tenant.enable_policies", "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestTenantAPIService_GetTenantSettingsByKey_ManagementTenant_WithTenantIdParam(t *testing.T) {
	settingsMgrMock, testSubject, _, tenantMgrMock, _ := initTenantAPIServiceTests(t)

	ctx := context.WithValue(context.Background(), runaictx.CfIsManagementTenant, "true")

	tenantMgrMock.On("GetTenantByIdInternal", ctx, int32(99)).Return(getMockedTenant(), &http.Response{}, nil)
	settingsMgrMock.On("GetAllSettings", ctx, 99, mock.Anything, []string{}, false).Return(getMockedTenantSettings(), nil)

	response, err := testSubject.GetTenantSettingsByKey(ctx, "tenant.enable_policies", "99")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.NotNil(t, response.Body)
}

func TestTenantAPIService_GetTenantSettingsByKey_ManagementTenant_NoParam_FallsBackToContext(t *testing.T) {
	settingsMgrMock, testSubject, _, tenantMgrMock, _ := initTenantAPIServiceTests(t)

	ctx := context.WithValue(context.Background(), runaictx.CfIsManagementTenant, "true")
	ctx = context.WithValue(ctx, runaictx.CfTenantId, testTenantIDStr)

	tenantMgrMock.On("GetTenantByIdInternal", ctx, int32(testTenantID)).Return(getMockedTenant(), &http.Response{}, nil)
	settingsMgrMock.On("GetAllSettings", ctx, testTenantID, mock.Anything, []string{}, false).Return(getMockedTenantSettings(), nil)

	response, err := testSubject.GetTenantSettingsByKey(ctx, "tenant.enable_policies", "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.NotNil(t, response.Body)
}

func TestTenantAPIService_GetTenantSettingsByKey_ManagementTenant_NoParam_NoContext(t *testing.T) {
	_, testSubject, _, _, _ := initTenantAPIServiceTests(t)

	ctx := context.WithValue(context.Background(), runaictx.CfIsManagementTenant, "true")

	response, err := testSubject.GetTenantSettingsByKey(ctx, "tenant.enable_policies", "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	body, ok := response.Body.(BackendError)
	assert.True(t, ok)
	assert.Contains(t, body.Description, "invalid tenant id")
}

// ==================== UpdateTenantSetting ====================

func TestTenantAPIService_UpdateTenantSetting_Success(t *testing.T) {
	settingsMgrMock, testSubject, clusterClientMock, _, ctx := initTenantAPIServiceTests(t)

	value := interface{}(true)
	req := api.TenantSettingCreationRequest{Key: "tenant.enable_policies", Value: &value}

	clusterClientMock.On("GetClustersMinimalData", ctx).Return(getMockedClustersMinimalData(), nil)
	settingsMgrMock.On("SetSetting", ctx, testTenantID, "tenant.enable_policies", true, []string{"2.17.0"}).Return(nil)

	response, err := testSubject.UpdateTenantSetting(ctx, "", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	body, ok := response.Body.(api.TenantSettingCreationResponse)
	assert.True(t, ok)
	assert.Equal(t, "done", body.Msg)
}

func TestTenantAPIService_UpdateTenantSetting_MissingTenantIDInContext(t *testing.T) {
	_, testSubject, _, _, _ := initTenantAPIServiceTests(t)

	ctx := context.Background()
	value := interface{}(true)
	req := api.TenantSettingCreationRequest{Key: "tenant.enable_policies", Value: &value}

	response, err := testSubject.UpdateTenantSetting(ctx, "", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	body, ok := response.Body.(BackendError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, body.Status)
	assert.Contains(t, body.Description, "invalid tenant id")
	assert.Equal(t, "BackendError", body.ErrorType)
}

func TestTenantAPIService_UpdateTenantSetting_SetSettingError(t *testing.T) {
	settingsMgrMock, testSubject, clusterClientMock, _, ctx := initTenantAPIServiceTests(t)

	value := interface{}(true)
	req := api.TenantSettingCreationRequest{Key: "unknown.key", Value: &value}

	clusterClientMock.On("GetClustersMinimalData", ctx).Return(getMockedClustersMinimalData(), nil)
	settingsMgrMock.On("SetSetting", ctx, testTenantID, "unknown.key", true, []string{"2.17.0"}).Return(errors.New("unknown setting key: unknown.key"))

	response, err := testSubject.UpdateTenantSetting(ctx, "", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	body, ok := response.Body.(BackendError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, body.Status)
	assert.Contains(t, body.Description, "unknown setting key")
	assert.Equal(t, "BackendError", body.ErrorType)
}

func TestTenantAPIService_UpdateTenantSetting_PostSaveError(t *testing.T) {
	settingsMgrMock, testSubject, clusterClientMock, _, ctx := initTenantAPIServiceTests(t)

	value := interface{}(true)
	req := api.TenantSettingCreationRequest{Key: "tenant.enable_policies", Value: &value}

	clusterClientMock.On("GetClustersMinimalData", ctx).Return(getMockedClustersMinimalData(), nil)
	settingsMgrMock.On("SetSetting", ctx, testTenantID, "tenant.enable_policies", true, []string{"2.17.0"}).Return(&settings.PostSaveError{Err: errors.New("post save failed")})

	response, err := testSubject.UpdateTenantSetting(ctx, "", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	body, ok := response.Body.(BackendError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, body.Status)
	assert.Contains(t, body.Description, "post save failed")
	assert.Equal(t, "BackendError", body.ErrorType)
}

func TestTenantAPIService_UpdateTenantSetting_NilValue(t *testing.T) {
	settingsMgrMock, testSubject, clusterClientMock, _, ctx := initTenantAPIServiceTests(t)

	req := api.TenantSettingCreationRequest{Key: "tenant.enable_policies", Value: nil}

	clusterClientMock.On("GetClustersMinimalData", ctx).Return(getMockedClustersMinimalData(), nil)
	settingsMgrMock.On("SetSetting", ctx, testTenantID, "tenant.enable_policies", nil, []string{"2.17.0"}).Return(nil)

	response, err := testSubject.UpdateTenantSetting(ctx, "", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestTenantAPIService_UpdateTenantSetting_ClusterVersionsError_StillProceeds(t *testing.T) {
	settingsMgrMock, testSubject, clusterClientMock, _, ctx := initTenantAPIServiceTests(t)

	value := interface{}(true)
	req := api.TenantSettingCreationRequest{Key: "tenant.enable_policies", Value: &value}

	clusterClientMock.On("GetClustersMinimalData", ctx).Return(nil, errors.New("cluster service unavailable"))
	settingsMgrMock.On("SetSetting", ctx, testTenantID, "tenant.enable_policies", true, []string{}).Return(nil)

	response, err := testSubject.UpdateTenantSetting(ctx, "", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestTenantAPIService_UpdateTenantSetting_ManagementTenant_Success(t *testing.T) {
	settingsMgrMock, testSubject, clusterClientMock, _, _ := initTenantAPIServiceTests(t)

	ctx := context.WithValue(context.Background(), runaictx.CfIsManagementTenant, "true")

	value := interface{}(true)
	req := api.TenantSettingCreationRequest{Key: "tenant.enable_policies", Value: &value}

	clusterClientMock.On("GetClustersMinimalData", ctx).Return(getMockedClustersMinimalData(), nil)
	settingsMgrMock.On("SetSetting", ctx, 99, "tenant.enable_policies", true, []string{"2.17.0"}).Return(nil)

	response, err := testSubject.UpdateTenantSetting(ctx, "99", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestTenantAPIService_UpdateTenantSetting_ManagementTenant_TenantIdFromContext(t *testing.T) {
	settingsMgrMock, testSubject, clusterClientMock, _, ctx := initTenantAPIServiceTests(t)

	ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

	value := interface{}(true)
	req := api.TenantSettingCreationRequest{Key: "tenant.enable_policies", Value: &value}

	clusterClientMock.On("GetClustersMinimalData", ctx).Return(getMockedClustersMinimalData(), nil)
	settingsMgrMock.On("SetSetting", ctx, testTenantID, "tenant.enable_policies", true, []string{"2.17.0"}).Return(nil)

	response, err := testSubject.UpdateTenantSetting(ctx, "", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	body, ok := response.Body.(api.TenantSettingCreationResponse)
	assert.True(t, ok)
	assert.Equal(t, "done", body.Msg)
}

func TestTenantAPIService_UpdateTenantSetting_ManagementTenant_InvalidTenantIdParam(t *testing.T) {
	_, testSubject, _, _, _ := initTenantAPIServiceTests(t)

	ctx := context.WithValue(context.Background(), runaictx.CfIsManagementTenant, "true")

	value := interface{}(true)
	req := api.TenantSettingCreationRequest{Key: "tenant.enable_policies", Value: &value}

	response, err := testSubject.UpdateTenantSetting(ctx, "not-a-number", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	body, ok := response.Body.(BackendError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, body.Status)
	assert.Contains(t, body.Description, "invalid tenant Id")
	assert.Equal(t, "BackendError", body.ErrorType)
}

// ==================== Helper Functions ====================

func getMockedClustersMinimalData() *cluster_openapi.GetClustersMinimalData200Response {
	cluster := cluster_openapi.ClustersMinimalData{
		Uuid:     "cluster-uuid-1",
		Name:     "test-cluster",
		TenantId: testTenantID,
	}
	cluster.SetVersion("2.17.0")
	return &cluster_openapi.GetClustersMinimalData200Response{
		Clusters: []cluster_openapi.ClustersMinimalData{cluster},
	}
}

func getMockedTenant() *tm_openapi.Tenant {
	return &tm_openapi.Tenant{
		Id:        testTenantID,
		Name:      "test-tenant",
		CreatedAt: testCreatedAt,
	}
}

func getMockedTenantSettings() []api.Setting {
	trueVal := interface{}(true)
	return []api.Setting{
		{
			Key:         "tenant.enable_policies",
			Source:      "Tenant",
			Category:    "Tenant",
			Label:       "Enable Policies",
			Description: "Enable policies for the tenant",
			Type:        "Boolean",
			Value:       &trueVal,
			Active:      true,
		},
	}
}
