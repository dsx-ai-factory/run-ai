/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/generated/config_client"
	events_http_client "github.com/run-ai/runai/control-plane/notifications-service/pkg/client/generated/notify"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNotify_ShouldCallNotify(t *testing.T) {
	ctx := context.Background()

	confClientMock := mocks.NewConfigClienter(t)
	confClientMock.On("ShouldNotify", mock.Anything).Return(true, nil).Once()

	eventsClientMock := mocks.NewNotificationsEventClienter(t)
	eventsClientMock.On("NotifyBatch", []events_http_client.Notification{{
		TenantId:  "t",
		EventType: "type",
		Properties: map[string]interface{}{
			"tenant_id":   "t",
			"user_id":     "u",
			"title":       "Title",
			"category":    "category",
			"description": "description",
			"type":        "type",
		},
	}}).Once().Return(nil, nil)

	nc := client.NewNotifyClient(eventsClientMock, confClientMock, func(err error) {}, client.NotifyClientConfig{RetryIntervalSec: 1, BatchSize: 1}, ctx)

	go nc.Notify(&client.Notification{
		Title:       "Title",
		Tenant:      "t",
		User:        "u",
		Category:    "category",
		Description: "description",
		Type:        "type",
	})

	assert.Eventually(t, func() bool { return len(eventsClientMock.Calls) > 0 }, 5*time.Second, 500*time.Millisecond)
}

func TestNotify_ShouldCallNotifyAndHandleError(t *testing.T) {
	ctx := context.Background()

	confClientMock := mocks.NewConfigClienter(t)
	confClientMock.On("ShouldNotify", mock.Anything).Return(true, nil).Once()

	eventClientMock := mocks.NewNotificationsEventClienter(t)
	eventClientMock.On("NotifyBatch", mock.Anything).Once().Return(&events_http_client.NotifyBatch200Response{Succeed: 0}, errors.New("client failed"))

	nc := client.NewNotifyClient(eventClientMock, confClientMock, func(err error) {
		assert.Equal(t, err.Error(), "failed to ingest batch: client failed. Retrying\nmsgs left: 1")
	}, client.NotifyClientConfig{RetryIntervalSec: 1, BatchSize: 1}, ctx)

	nc.Notify(&client.Notification{
		Title:       "Title",
		Tenant:      "t",
		User:        "u",
		Category:    "category",
		Type:        "type",
		Description: "description",
	})

	assert.Eventually(t, func() bool { return len(eventClientMock.Calls) > 0 }, 5*time.Second, 500*time.Millisecond)
}

func TestNotify_ShouldCallNotifyAndHandleErrorWithRetry(t *testing.T) {
	ctx := context.Background()

	eventClientMock := mocks.NewNotificationsEventClienter(t)
	eventClientMock.On("NotifyBatch", mock.Anything).Times(3).Return(&events_http_client.NotifyBatch200Response{Succeed: 0}, errors.New("client failed")).Times(3)

	confClientMock := mocks.NewConfigClienter(t)
	confClientMock.On("ShouldNotify", mock.Anything).Return(true, nil)

	nc := client.NewNotifyClient(eventClientMock, confClientMock, func(err error) {
		assert.Equal(t, err.Error(), "failed to ingest batch: client failed. Retrying\nmsgs left: 1")
	}, client.NotifyClientConfig{
		RetryIntervalSec: 1,
		RetryThreshold:   2,
		BatchSize:        1,
	}, ctx)

	nc.Notify(&client.Notification{
		Title:    "Title",
		Tenant:   "t",
		User:     "u",
		Category: "category",
		Type:     "type",
	})

	assert.Eventually(t, func() bool { return len(eventClientMock.Calls) > 2 }, 5*time.Second, 500*time.Millisecond)
}

func TestNotify_ShouldCallNotifyBatchSucceedAfterSecondTry(t *testing.T) {
	ctx := context.Background()

	confClientMock := mocks.NewConfigClienter(t)
	confClientMock.On("ShouldNotify", mock.Anything).Twice().Return(true, nil)

	eventClientMock := mocks.NewNotificationsEventClienter(t)
	eventClientMock.On("NotifyBatch", mock.Anything).Once().Return(nil, nil)

	nc := client.NewNotifyClient(eventClientMock, confClientMock, func(err error) {}, client.NotifyClientConfig{
		BatchSize: 2, BatchIntervalSec: 60, RetryIntervalSec: 1,
	}, ctx)

	for i := 0; i < 2; i++ {
		nc.Notify(&client.Notification{
			Title:       "Title",
			Tenant:      "t",
			User:        "u",
			Category:    "category",
			Description: "description",
			Type:        "type",
		})
	}

	assert.Eventually(t, func() bool { return len(eventClientMock.Calls) > 0 }, 5*time.Second, 500*time.Millisecond)
}

func TestNotify_ShouldCallNotifyBatchByTime(t *testing.T) {
	ctx := context.Background()

	confClientMock := mocks.NewConfigClienter(t)
	confClientMock.On("ShouldNotify", mock.Anything).Times(3).Return(true, nil)

	eventClientMock := mocks.NewNotificationsEventClienter(t)
	eventClientMock.On("NotifyBatch", mock.Anything).Once().Return(nil, nil)

	nc := client.NewNotifyClient(eventClientMock, confClientMock, func(err error) {}, client.NotifyClientConfig{
		BatchSize: 3, BatchIntervalSec: 60, RetryIntervalSec: 1,
	}, ctx)

	for i := 0; i < 3; i++ {
		nc.Notify(&client.Notification{
			Title:       "Title",
			Tenant:      "t",
			User:        "u",
			Category:    "category",
			Description: "description",
			Type:        "type",
		})
	}

	assert.Eventually(t, func() bool { return len(eventClientMock.Calls) > 0 }, 5*time.Second, 500*time.Millisecond)
}

func TestNotify_ShouldNotCallNotifyBatch(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())

	confClientMock := mocks.NewConfigClienter(t)
	confClientMock.On("ShouldNotify", mock.Anything).Twice().Return(true, nil)

	eventClientMock := mocks.NewNotificationsEventClienter(t)
	eventClientMock.On("NotifyBatch", mock.Anything).Once().Return(nil, nil)

	nc := client.NewNotifyClient(eventClientMock, confClientMock, func(err error) {}, client.NotifyClientConfig{
		BatchSize: 3, BatchIntervalSec: 60, RetryIntervalSec: 1,
	}, ctx)

	for i := 0; i < 2; i++ {
		nc.Notify(&client.Notification{
			Title:       "Title",
			Tenant:      "t",
			User:        "u",
			Category:    "category",
			Description: "description",
			Type:        "type",
		})
	}

	time.Sleep(time.Second)
	eventClientMock.AssertNumberOfCalls(t, "NotifyBatch", 0)

	cancelCtx()
	nc.WaitDone()
}

func TestNotify_ShouldNotCallNotifyBatchNoRoutes(t *testing.T) {
	ctx := context.Background()

	confClientMock := mocks.NewConfigClienter(t)
	confClientMock.On("ShouldNotify", mock.Anything).Twice().Return(false, nil)

	eventClientMock := mocks.NewNotificationsEventClienter(t)

	nc := client.NewNotifyClient(eventClientMock, confClientMock, func(err error) {}, client.NotifyClientConfig{RetryIntervalSec: 1}, ctx)

	for i := 0; i < 2; i++ {
		nc.Notify(&client.Notification{
			Title:       "Title",
			Tenant:      "t",
			User:        "u",
			Category:    "category",
			Description: "description",
			Type:        "type",
		})
	}

	assert.Eventually(t, func() bool { return len(confClientMock.Calls) > 1 }, 5*time.Second, 500*time.Millisecond)
}

func TestNotify_ShouldFlushPendingNotificationsOnContextCancel(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())

	confClientMock := mocks.NewConfigClienter(t)
	confClientMock.On("ShouldNotify", mock.Anything).Once().Return(true, nil)

	eventsClientMock := mocks.NewNotificationsEventClienter(t)
	eventsClientMock.On("NotifyBatch", mock.Anything).Once().Return(nil, nil)

	nc := client.NewNotifyClient(eventsClientMock, confClientMock, func(err error) {}, client.NotifyClientConfig{
		BatchSize: 10, BatchIntervalSec: 30, RetryIntervalSec: 1,
	}, ctx)

	nc.Notify(&client.Notification{
		Title:       "Title",
		Tenant:      "t",
		User:        "u",
		Category:    "category",
		Description: "description",
		Type:        "type",
	})

	time.Sleep(1 * time.Second) // wait for the notification to be batched before cancelling context

	ctxCancel()
	assert.Eventually(t, func() bool { return len(eventsClientMock.Calls) > 0 }, 5*time.Second, 500*time.Millisecond)
}

func TestRegisterNotifications_NoRegisteredNotificationForOriginWithoutRemoveMissing(t *testing.T) {
	ctx := context.Background()
	cclientMock := mocks.NewConfigClienter(t)
	nc := client.NewNotifyClient(mocks.NewNotificationsEventClienter(t), cclientMock, func(err error) {}, client.NotifyClientConfig{
		BatchSize: 10, BatchIntervalSec: 30, RetryIntervalSec: 1,
	}, ctx)

	cclientMock.On("UpdateNotification", "originTest", "category", "type", "d", "t").Once().Return(nil)
	err := nc.RegisterNotifications("originTest", []*client.NotificationKind{{
		Category: "category", Type: "type", Description: "d", Template: "t",
	}}, nil)
	assert.NoError(t, err)
}

func TestRegisterNotifications_NoRegisteredNotificationForOriginWithRemoveMissing(t *testing.T) {
	ctx := context.Background()
	removeMissing := true

	confClientMock := mocks.NewConfigClienter(t)

	confClientMock.On("GetAllOriginNotification", "originTest").Once().Return([]config_client.NotificationKind{}, nil)
	confClientMock.On("UpdateNotification", "originTest", "category", "type", "d", "t").Once().Return(nil)

	nc := client.NewNotifyClient(mocks.NewNotificationsEventClienter(t), confClientMock, func(err error) {}, client.NotifyClientConfig{
		BatchSize: 10, BatchIntervalSec: 30, RetryIntervalSec: 1,
	}, ctx)

	err := nc.RegisterNotifications("originTest", []*client.NotificationKind{{
		Category: "category", Type: "type", Description: "d", Template: "t",
	}}, &removeMissing)

	assert.NoError(t, err)
}

func TestRegisterNotifications_ExistRegisteredNotificationForOriginWithRemoveMissing(t *testing.T) {
	ctx := context.Background()
	removeMissing := true
	confClientMock := mocks.NewConfigClienter(t)

	confClientMock.On("GetAllOriginNotification", "originTest").Once().Return([]config_client.NotificationKind{{
		Category: "category", Type: "type1", Description: "d", Template: "t",
	}, {
		Category: "category", Type: "type2", Description: "d", Template: "t",
	}, {
		Category: "category", Type: "type3", Description: "d", Template: "t",
	}}, nil)
	confClientMock.On("DeleteNotification", "category", mock.Anything).Times(3).Return(nil)
	confClientMock.On("UpdateNotification", "originTest", "category", "type-new", "d", "t").Once().Return(nil)

	nc := client.NewNotifyClient(mocks.NewNotificationsEventClienter(t), confClientMock, func(err error) {}, client.NotifyClientConfig{
		BatchSize: 10, BatchIntervalSec: 30, RetryIntervalSec: 1,
	}, ctx)

	err := nc.RegisterNotifications("originTest", []*client.NotificationKind{{
		Category: "category", Type: "type-new", Description: "d", Template: "t",
	}}, &removeMissing)

	assert.NoError(t, err)
}

func TestRegisterNotifications_UpdateBeforeCreate(t *testing.T) {
	ctx := context.Background()
	removeMissing := true
	confClientMock := mocks.NewConfigClienter(t)

	confClientMock.On("GetAllOriginNotification", "originTest").Once().Return([]config_client.NotificationKind{{
		Category: "category", Type: "type-new", Description: "d", Template: "t",
	}}, nil)
	confClientMock.On("UpdateNotification", "originTest", "category", "type-new", "d", "t").Once().Return(nil)

	nc := client.NewNotifyClient(mocks.NewNotificationsEventClienter(t), confClientMock, func(err error) {}, client.NotifyClientConfig{
		BatchSize: 10, BatchIntervalSec: 30, RetryIntervalSec: 1,
	}, ctx)

	err := nc.RegisterNotifications("originTest", []*client.NotificationKind{{
		Category: "category", Type: "type-new", Description: "d", Template: "t",
	}}, &removeMissing)

	assert.NoError(t, err)
}
