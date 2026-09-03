/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/internal/generated/pkg/api"
	authMock "github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/authorization/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/services/mocks"
	system_notification "github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/system-notification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSetSystemNotificationMessage_Success(t *testing.T) {
	managerMock := mocks.NewSystemNotificationManager(t)
	svc := &SystemNotificationAPIService{systemNotificationManager: managerMock}
	ctx := context.Background()

	msg := &system_notification.Message{
		ID:        "msg-123",
		Msg:       "Hello everyone",
		CreatedAt: 1706886400,
	}
	managerMock.On("Get", ctx).Return(msg, nil)

	response, err := svc.GetSystemNotification(ctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	body, ok := response.Body.(api.GetSystemNotification)
	require.True(t, ok)
	assert.Equal(t, "msg-123", body.Id)
	assert.Equal(t, "Hello everyone", body.Msg)
	assert.Equal(t, int64(1706886400), body.CreatedAt)
}

func TestSetSystemNotificationMessage_ManagerError(t *testing.T) {
	managerMock := mocks.NewSystemNotificationManager(t)
	svc := &SystemNotificationAPIService{systemNotificationManager: managerMock}
	ctx := context.Background()
	managerErr := errors.New("database error")

	managerMock.On("Get", ctx).Return(nil, managerErr)

	response, err := svc.GetSystemNotification(ctx)
	require.Error(t, err)
	assert.Equal(t, err, fmt.Errorf("failed to get system notification"))
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Nil(t, response.Body)
}

func TestUpdateSystemNotificationMessage_Success(t *testing.T) {
	managerMock := mocks.NewSystemNotificationManager(t)
	authMock := authMock.NewAuthManagerInterface(t)
	svc := &SystemNotificationAPIService{
		systemNotificationManager: managerMock,
		authManager:               authMock,
	}
	ctx := context.Background()
	msgStr := "New announcement"
	updateReq := api.UpdateSystemNotification{
		Msg: &msgStr,
	}

	authMock.On("HandleSystemPermissions", ctx).Return(nil)
	managerMock.On("Set", ctx, mock.MatchedBy(func(m *system_notification.Message) bool {
		return m.Msg == "New announcement" && m.ID != ""
	})).Return(nil)

	response, err := svc.UpdateSystemNotification(ctx, updateReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, updateReq, response.Body)
}

func TestUpdateSystemNotificationMessage_Unauthorized(t *testing.T) {
	authMock := authMock.NewAuthManagerInterface(t)
	svc := &SystemNotificationAPIService{
		systemNotificationManager: mocks.NewSystemNotificationManager(t),
		authManager:               authMock,
	}
	ctx := context.Background()
	msgStr := "New announcement"
	updateReq := api.UpdateSystemNotification{Msg: &msgStr}

	authMock.On("HandleSystemPermissions", ctx).Return(errors.New("Unauthorized to manage UI settings"))

	response, err := svc.UpdateSystemNotification(ctx, updateReq)
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, response.Code)
	assert.Contains(t, err.Error(), "forbidden to update system notification")
}

func TestUpdateSystemNotificationMessage_ManagerError(t *testing.T) {
	managerMock := mocks.NewSystemNotificationManager(t)
	authMock := authMock.NewAuthManagerInterface(t)
	svc := &SystemNotificationAPIService{
		systemNotificationManager: managerMock,
		authManager:               authMock,
	}
	ctx := context.Background()
	msgStr := "Fail message"
	updateReq := api.UpdateSystemNotification{
		Msg: &msgStr,
	}
	managerErr := errors.New("save failed")

	authMock.On("HandleSystemPermissions", ctx).Return(nil)
	managerMock.On("Set", ctx, mock.MatchedBy(func(m *system_notification.Message) bool {
		return m.Msg == "Fail message"
	})).Return(managerErr)

	response, err := svc.UpdateSystemNotification(ctx, updateReq)
	require.Error(t, err)
	assert.Equal(t, fmt.Errorf("failed to update system notification"), err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Nil(t, response.Body)
}

func TestDeleteSystemNotificationMessage_Success(t *testing.T) {
	managerMock := mocks.NewSystemNotificationManager(t)
	authMock := authMock.NewAuthManagerInterface(t)
	svc := &SystemNotificationAPIService{
		systemNotificationManager: managerMock,
		authManager:               authMock,
	}
	ctx := context.Background()

	authMock.On("HandleSystemPermissions", ctx).Return(nil)
	managerMock.On("Invalidate", ctx).Return(nil)

	response, err := svc.DeleteSystemNotification(ctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Nil(t, response.Body)
}

func TestDeleteSystemNotificationMessage_Unauthorized(t *testing.T) {
	authMock := authMock.NewAuthManagerInterface(t)
	svc := &SystemNotificationAPIService{
		systemNotificationManager: mocks.NewSystemNotificationManager(t),
		authManager:               authMock,
	}
	ctx := context.Background()

	authMock.On("HandleSystemPermissions", ctx).Return(errors.New("Unauthorized to manage UI settings"))

	response, err := svc.DeleteSystemNotification(ctx)
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, response.Code)
	assert.Contains(t, err.Error(), "forbidden to delete system notification")
}

func TestDeleteSystemNotificationMessage_ManagerError(t *testing.T) {
	managerMock := mocks.NewSystemNotificationManager(t)
	authMock := authMock.NewAuthManagerInterface(t)
	svc := &SystemNotificationAPIService{
		systemNotificationManager: managerMock,
		authManager:               authMock,
	}
	ctx := context.Background()
	managerErr := errors.New("cache del failed")

	authMock.On("HandleSystemPermissions", ctx).Return(nil)
	managerMock.On("Invalidate", ctx).Return(managerErr)

	response, err := svc.DeleteSystemNotification(ctx)
	require.Error(t, err)
	assert.Equal(t, fmt.Errorf("failed to delete system notification"), err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Nil(t, response.Body)
}
