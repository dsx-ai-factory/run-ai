/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package http_listener

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/advanced/http_listener/mocks"
	api "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/advanced/http_listener/openapi/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres/configprovider"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConfigNotificationsDelete_Failed(t *testing.T) {
	nd := mocks.NewNotificationKindsDao(t)
	nd.On("DeleteNotification", "c1", "t1").Return(errors.New("failed"))

	rnas := NewRegisterNotificationAPIService(nd)
	res, err := rnas.DeleteNotificationKind(context.Background(), "c1", "t1")
	apiErr := err.(*ApiError)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "failed to delete notification kind, category: c1 type t1", apiErr.Message)
}

func TestConfigNotificationsDelete_Succeed(t *testing.T) {
	nd := mocks.NewNotificationKindsDao(t)
	nd.On("DeleteNotification", "c1", "t1").Return(nil)

	rnas := NewRegisterNotificationAPIService(nd)
	res, err := rnas.DeleteNotificationKind(context.Background(), "c1", "t1")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestCreateNotificationKind_Failed(t *testing.T) {
	notificationKind := api.NotificationKind{
		Origin:      "test",
		Category:    "c1",
		Type:        "type",
		Description: "describe",
		Template:    "some template",
	}

	nd := mocks.NewNotificationKindsDao(t)
	nd.On("CreateNotification", "test", "c1", "type", "describe", "some template").Return(errors.New("failed"))

	rnas := NewRegisterNotificationAPIService(nd)
	res, err := rnas.CreateNotificationKind(context.Background(), notificationKind)
	apiErr := err.(*ApiError)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Contains(t, apiErr.Message, "failed to create notification kind")
}

func TestCreateNotificationKind_Succeed(t *testing.T) {
	notificationKind := api.NotificationKind{
		Origin:      "test",
		Category:    "c1",
		Type:        "type",
		Description: "describe",
		Template:    "some template",
	}

	nd := mocks.NewNotificationKindsDao(t)
	nd.On("CreateNotification", "test", "c1", "type", "describe", "some template").Return(nil)

	rnas := NewRegisterNotificationAPIService(nd)
	res, err := rnas.CreateNotificationKind(context.Background(), notificationKind)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestUpdateNotificationKind_Succeed(t *testing.T) {
	notificationKind := api.NotificationKind{
		Origin:      "origin",
		Category:    "cat",
		Type:        "typ",
		Description: "desc",
		Template:    "tpl",
	}

	nd := mocks.NewNotificationKindsDao(t)
	nd.On("UpdateNotification", "origin", "cat", "typ", "desc", "tpl").Return(nil)

	rnas := NewRegisterNotificationAPIService(nd)
	res, err := rnas.UpdateNotificationKind(context.Background(), notificationKind)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestUpdateNotificationKind_UpdateFails(t *testing.T) {
	notificationKind := api.NotificationKind{
		Origin:      "origin",
		Category:    "cat",
		Type:        "typ",
		Description: "desc",
		Template:    "tpl",
	}

	nd := mocks.NewNotificationKindsDao(t)
	nd.On("UpdateNotification", "origin", "cat", "typ", "desc", "tpl").Return(errors.New("db error"))

	rnas := NewRegisterNotificationAPIService(nd)
	res, err := rnas.UpdateNotificationKind(context.Background(), notificationKind)
	apiErr := err.(*ApiError)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Contains(t, apiErr.Message, "failed to update notification kind")
}

func TestUpdateNotificationKind_Conflict(t *testing.T) {
	notificationKind := api.NotificationKind{
		Origin:      "origin",
		Category:    "cat",
		Type:        "typ",
		Description: "desc",
		Template:    "tpl",
	}

	nd := mocks.NewNotificationKindsDao(t)
	nd.On("UpdateNotification", "origin", "cat", "typ", "desc", "tpl").Return(configprovider.NewError("not found", configprovider.ErrCodeNotFound))

	rnas := NewRegisterNotificationAPIService(nd)
	res, err := rnas.UpdateNotificationKind(context.Background(), notificationKind)
	apiErr := err.(*ApiError)
	assert.Equal(t, http.StatusNotFound, res.Code)
	assert.Contains(t, apiErr.Message, "not exists")
}

func TestGetNotificationsKinds_GetAllFailed(t *testing.T) {
	nd := mocks.NewNotificationKindsDao(t)
	nd.On("FetchNotifications", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("failed"))

	rnas := NewRegisterNotificationAPIService(nd)
	res, err := rnas.GetNotificationKinds(context.Background(), "", "", "")
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestGetNotificationsKinds_GetAllSucceed(t *testing.T) {
	nd := mocks.NewNotificationKindsDao(t)
	nd.On("FetchNotifications", mock.Anything, mock.Anything, mock.Anything).Return([]*models.Notification{}, nil)

	rnas := NewRegisterNotificationAPIService(nd)
	res, err := rnas.GetNotificationKinds(context.Background(), "", "", "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestGetNotificationsKinds_GetByCategoryFailed(t *testing.T) {
	nd := mocks.NewNotificationKindsDao(t)
	nd.On("FetchNotifications", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("failed"))

	rnas := NewRegisterNotificationAPIService(nd)
	res, err := rnas.GetNotificationKinds(context.Background(), "", "category", "")
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestGetNotificationsKinds_GetByCategorySucceed(t *testing.T) {
	nd := mocks.NewNotificationKindsDao(t)
	nd.On("FetchNotifications", "", "category", "").Return([]*models.Notification{}, nil)

	rnas := NewRegisterNotificationAPIService(nd)
	res, err := rnas.GetNotificationKinds(context.Background(), "", "category", "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestGetNotificationsKinds_GetByCategoryAndTypeFailed(t *testing.T) {
	nd := mocks.NewNotificationKindsDao(t)
	nd.On("FetchNotifications", "", "category", "type").Return(nil, errors.New("failed"))

	rnas := NewRegisterNotificationAPIService(nd)
	res, err := rnas.GetNotificationKinds(context.Background(), "", "category", "type")
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestGetNotificationsKinds_GetByCategoryAndTypeSucceed(t *testing.T) {
	nd := mocks.NewNotificationKindsDao(t)
	nd.On("FetchNotifications", "", "category", "type").Return([]*models.Notification{{
		Category: "c1",
		Type:     "created",
	}}, nil)

	rnas := NewRegisterNotificationAPIService(nd)
	res, err := rnas.GetNotificationKinds(context.Background(), "", "category", "type")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, res.Body.([]api.NotificationKind)[0].Category, "c1")
}
