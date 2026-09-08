/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package notify

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/clients/notify/mocks"
	notifyclient "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client"
	notifications_service_mocks "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNotify(t *testing.T) {
	ctx := context.Background()
	notifyClientMock := notifications_service_mocks.NewNotifyClienter(t)
	notifyclient.SetMock(notifyClientMock)
	userRoutesClientMock := mocks.NewUserRoutesClienter(t)

	notifier := notifier{
		userRoutesClient:            userRoutesClientMock,
		deleteWelcomeUserRouteDelay: 0,
	}

	userRoutesClientMock.On("CreateRoute", ctx, "1", "test@example.com", mock.Anything).Return(nil)
	userRoutesClientMock.On("DeleteRoute", mock.Anything, "1", "test@example.com", mock.Anything).Return(nil)
	notifyClientMock.On("Notify", mock.Anything).Return()

	err := notifier.Welcome(ctx, "1", "test@example.com", "sink-id")
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond) // Ensure that the welcome user route deletion is processed
}

func TestNotify_FailedToCreateRoute(t *testing.T) {
	ctx := context.Background()
	userRoutesClientMock := mocks.NewUserRoutesClienter(t)

	notifier := notifier{
		userRoutesClient: userRoutesClientMock,
	}

	userRoutesClientMock.On("CreateRoute", ctx, "1", "test@example.com", mock.Anything).Return(fmt.Errorf("failed to create route"))
	err := notifier.Welcome(ctx, "1", "test@example.com", "sink-id")
	assert.EqualError(t, err, "failed to create route")
}
