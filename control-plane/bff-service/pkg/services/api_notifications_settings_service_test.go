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
	notificationsSettings "github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/settings/notifications"
	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	settingPkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	settingPkgManager "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/manager"
	settingPkgMocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

func initNotificationsSettingsServiceTests(t *testing.T) (settingPkgMock *settingPkgMocks.SettingsManager, testSubject *NotificationsSettingsAPIService, authManagerMock *authMock.AuthManagerInterface, ctx context.Context) {
	authManagerMock = authMock.NewAuthManagerInterface(t)
	settingPkgManagerMock := settingPkgMocks.NewSettingsManager(t)
	settingPkg.SetMock(settingPkgManagerMock)

	ctx = context.WithValue(context.Background(), "AuthRealm", realmName)
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantID)

	testSubject = &NotificationsSettingsAPIService{
		settingManager: *settings.NewSettingsManager(),
		settingsHelper: &SettingsHelper{
			authManager: authManagerMock,
		},
	}

	return settingPkgManagerMock, testSubject, authManagerMock, ctx
}

func TestNotificationsSettingsAPIService_GetNotificationsSettings_Success(t *testing.T) {
	settingPkgManagerMock, testSubject, authMock, ctx := initNotificationsSettingsServiceTests(t)

	authMock.On("GetTenantID", ctx, "").Return(ptr.To(tenantID), nil)
	settingPkgManagerMock.On("GetAll", mock.Anything, tenantID, string(utils.NotificationsCategory)).Return(getMockedNotificationsSettings(), nil)

	response, err := testSubject.GetNotificationsSettings(ctx, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	expected := &api.NotificationsSettingsGetAllResponse{
		AdminMessage: api.GetAdminMessage{
			Value: api.GetAdminMessageValue{
				Id:            "test-id",
				Msg:           "Test message",
				CreatedAt:     1706886400000,
				EnableDismiss: true,
			},
			DisplayName: "Admin message",
			Description: "The message will be displayed every time users sign in and immediately after it is saved to users who are already signed in.",
		},
	}
	assert.Equal(t, expected, response.Body)
}

func TestNotificationsSettingsAPIService_GetNotificationsSettings_AuthorizationError_Unauthorized(t *testing.T) {
	_, testSubject, authMock, ctx := initNotificationsSettingsServiceTests(t)

	authMock.On("GetTenantID", ctx, "").Return(nil, errors.New("unauthorized"))

	response, err := testSubject.GetNotificationsSettings(ctx, "")
	assert.Error(t, err)
	assert.Equal(t, http.StatusForbidden, response.Code)

	utilsErr, ok := err.(utils.Error)
	assert.True(t, ok)
	assert.Equal(t, http.StatusForbidden, utilsErr.Code)
	assert.Contains(t, utilsErr.Message, "unauthorized to get notifications settings")
}

func TestNotificationsSettingsAPIService_GetNotificationsSettings_AuthorizationError_InternalError(t *testing.T) {
	_, testSubject, authMock, ctx := initNotificationsSettingsServiceTests(t)

	authMock.On("GetTenantID", ctx, "").Return(nil, errors.New("internal database error"))

	response, err := testSubject.GetNotificationsSettings(ctx, "")
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	utilsErr, ok := err.(utils.Error)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, utilsErr.Code)
	assert.Contains(t, utilsErr.Message, "failed to get notifications settings")
}

func TestNotificationsSettingsAPIService_GetNotificationsSettings_GetSettingsError(t *testing.T) {
	settingPkgManagerMock, testSubject, authMock, ctx := initNotificationsSettingsServiceTests(t)

	authMock.On("GetTenantID", ctx, "").Return(ptr.To(tenantID), nil)
	settingPkgManagerMock.On("GetAll", mock.Anything, tenantID, string(utils.NotificationsCategory)).Return(nil, errors.New("database connection failed"))

	response, err := testSubject.GetNotificationsSettings(ctx, "")
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	utilsErr, ok := err.(utils.Error)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, utilsErr.Code)
	assert.Contains(t, utilsErr.Message, "failed to get notifications settings")
}

func TestNotificationsSettingsAPIService_GetNotificationsSettings_EmptySettings(t *testing.T) {
	settingPkgManagerMock, testSubject, authMock, ctx := initNotificationsSettingsServiceTests(t)

	authMock.On("GetTenantID", ctx, "").Return(ptr.To(tenantID), nil)
	settingPkgManagerMock.On("GetAll", mock.Anything, tenantID, string(utils.NotificationsCategory)).Return([]settingPkgManager.Settinger{}, nil)

	response, err := testSubject.GetNotificationsSettings(ctx, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	expected := &api.NotificationsSettingsGetAllResponse{}
	assert.Equal(t, expected, response.Body)
}

func TestNotificationsSettingsAPIService_GetNotificationsSettings_WithTenantId_Success(t *testing.T) {
	settingPkgManagerMock, testSubject, authMock, ctx := initNotificationsSettingsServiceTests(t)

	providedTenantId := "provided-tenant-id"
	authMock.On("GetTenantID", ctx, providedTenantId).Return(ptr.To(providedTenantId), nil)
	settingPkgManagerMock.On("GetAll", mock.Anything, providedTenantId, string(utils.NotificationsCategory)).Return(getMockedNotificationsSettings(), nil)

	response, err := testSubject.GetNotificationsSettings(ctx, providedTenantId)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestNotificationsSettingsAPIService_GetNotificationsSettingByKey_Success(t *testing.T) {
	settingPkgManagerMock, testSubject, authMock, ctx := initNotificationsSettingsServiceTests(t)

	authMock.On("GetTenantID", ctx, "").Return(ptr.To(tenantID), nil)
	settingPkgManagerMock.On("Get", mock.Anything, tenantID, string(api.NOTIFICATIONSSETTINGSKEYENUM_ADMIN_MESSAGE)).Return(&notificationsSettings.AdminMessage{
		ID:            "test-id",
		Msg:           "Test message",
		CreatedAt:     1706886400000,
		EnableDismiss: true,
	}, nil)

	response, err := testSubject.GetNotificationsSettingByKey(ctx, api.NOTIFICATIONSSETTINGSKEYENUM_ADMIN_MESSAGE, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestNotificationsSettingsAPIService_GetNotificationsSettingByKey_UnknownSettingKey(t *testing.T) {
	_, testSubject, authMock, ctx := initNotificationsSettingsServiceTests(t)

	authMock.On("GetTenantID", ctx, "").Return(ptr.To(tenantID), nil)

	unknownKey := api.NotificationsSettingsKeyEnum("unknown-setting-key")
	response, err := testSubject.GetNotificationsSettingByKey(ctx, unknownKey, "")
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	utilsErr, ok := err.(utils.Error)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, utilsErr.Code)
	assert.Contains(t, utilsErr.Message, "unknown setting key")
}

func TestNotificationsSettingsAPIService_UpdateNotificationsSettingByKey_Success(t *testing.T) {
	settingPkgManagerMock, testSubject, authMock, ctx := initNotificationsSettingsServiceTests(t)

	authMock.On("GetTenantID", ctx, "").Return(ptr.To(tenantID), nil)
	settingPkgManagerMock.On("Update", mock.Anything, tenantID, string(api.NOTIFICATIONSSETTINGSKEYENUM_ADMIN_MESSAGE), true, mock.AnythingOfType("notifications.AdminMessage")).Return(nil)

	updateRequest := api.UpdateAdminMessage{Msg: ptr.To("New message"), EnableDismiss: ptr.To(true)}
	response, err := testSubject.UpdateNotificationsSettingByKey(ctx, api.NOTIFICATIONSSETTINGSKEYENUM_ADMIN_MESSAGE, updateRequest, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestNotificationsSettingsAPIService_UpdateNotificationsSettingByKey_AuthorizationError_Unauthorized(t *testing.T) {
	_, testSubject, authMock, ctx := initNotificationsSettingsServiceTests(t)

	authMock.On("GetTenantID", ctx, "").Return(nil, errors.New("unauthorized"))

	updateRequest := api.UpdateAdminMessage{Msg: ptr.To("Test message"), EnableDismiss: ptr.To(true)}
	response, err := testSubject.UpdateNotificationsSettingByKey(ctx, api.NOTIFICATIONSSETTINGSKEYENUM_ADMIN_MESSAGE, updateRequest, "")
	assert.Error(t, err)
	assert.Equal(t, http.StatusForbidden, response.Code)

	utilsErr, ok := err.(utils.Error)
	assert.True(t, ok)
	assert.Equal(t, http.StatusForbidden, utilsErr.Code)
	assert.Contains(t, utilsErr.Message, "unauthorized to update notifications setting by key")
}

func TestNotificationsSettingsAPIService_UpdateNotificationsSettingByKey_UnknownSettingKey(t *testing.T) {
	_, testSubject, authMock, ctx := initNotificationsSettingsServiceTests(t)

	authMock.On("GetTenantID", ctx, "").Return(ptr.To(tenantID), nil)

	unknownKey := api.NotificationsSettingsKeyEnum("unknown-setting-key")
	updateRequest := api.UpdateAdminMessage{Msg: ptr.To("Test message"), EnableDismiss: ptr.To(true)}
	response, err := testSubject.UpdateNotificationsSettingByKey(ctx, unknownKey, updateRequest, "")
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	utilsErr, ok := err.(utils.Error)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, utilsErr.Code)
	assert.Contains(t, utilsErr.Message, "unknown setting key")
}

func getMockedNotificationsSettings() []settingPkgManager.Settinger {
	return []settingPkgManager.Settinger{
		&notificationsSettings.AdminMessage{
			ID:            "test-id",
			Msg:           "Test message",
			CreatedAt:     1706886400000,
			EnableDismiss: true,
		},
	}
}
