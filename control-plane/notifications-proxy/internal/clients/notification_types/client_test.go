package notification_types

import (
	"context"
	"errors"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients/notification_types/mocks"
	nt "github.com/run-ai/runai/control-plane/notifications-proxy/internal/notification_types"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/generated/config_client"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewNotificationKindsClienter(t)
	c := &Client{notificationKindsClient: mockClient}

	notificationKinds := []config_client.NotificationKind{
		{Origin: "origin1", Category: "category1", Type: "type1", Description: "description1", Template: "template1"},
		{Origin: "origin2", Category: "category2", Type: "type2", Description: "description2", Template: "template2"},
	}
	mockClient.On("GetNotificationKinds", ctx).Return(notificationKinds, nil)

	res, err := c.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, []nt.NotificationType{
		{Origin: "origin1", Category: "category1", Type: "type1", Description: "description1", Template: "template1"},
		{Origin: "origin2", Category: "category2", Type: "type2", Description: "description2", Template: "template2"},
	}, res)
	mockClient.AssertExpectations(t)
}

func TestGet_NoNotificationTypes(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewNotificationKindsClienter(t)
	c := &Client{notificationKindsClient: mockClient}

	mockClient.On("GetNotificationKinds", ctx).Return([]config_client.NotificationKind{}, nil)

	res, err := c.Get(ctx)
	require.NoError(t, err)
	require.Empty(t, res)
	mockClient.AssertExpectations(t)
}

func TestGet_Failed(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewNotificationKindsClienter(t)
	c := &Client{notificationKindsClient: mockClient}

	mockClient.On("GetNotificationKinds", ctx).Return(nil, errors.New("backend error"))

	res, err := c.Get(ctx)
	require.EqualError(t, err, "backend error")
	require.Nil(t, res)
	mockClient.AssertExpectations(t)
}
