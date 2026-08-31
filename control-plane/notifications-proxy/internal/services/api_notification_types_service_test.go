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

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/notification_types"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/services/mocks"
	"github.com/stretchr/testify/require"
)

func TestGetNotificationTypes_Succeed(t *testing.T) {
	ctx := context.Background()
	mockManager := mocks.NewNotificationTypesManager(t)
	nts := &NotificationTypesAPIService{mockManager}

	notificationTypes := []notification_types.NotificationType{
		{Category: "category1", Type: "type1", Description: "description1"},
		{Category: "category2", Type: "type2", Description: "description2"},
	}

	mockManager.On("GetNotificationTypes", ctx).Return(notificationTypes, nil)

	res, err := nts.GetNotificationTypes(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, []api.NotificationType{
		{Category: "category1", Type: "type1", Description: "description1"},
		{Category: "category2", Type: "type2", Description: "description2"},
	}, res.Body)
}

func TestGetNotificationTypes_Failed(t *testing.T) {
	ctx := context.Background()
	mockManager := mocks.NewNotificationTypesManager(t)
	nts := &NotificationTypesAPIService{mockManager}

	mockManager.On("GetNotificationTypes", ctx).Return(nil, errors.New("error"))

	res, err := nts.GetNotificationTypes(ctx)
	require.EqualError(t, err, "failed to get notification types")
	require.Equal(t, http.StatusInternalServerError, res.Code)
}
