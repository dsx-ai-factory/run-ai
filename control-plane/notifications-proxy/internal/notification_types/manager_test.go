package notification_types_test

import (
	"context"
	"errors"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/notification_types"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/notification_types/mocks"
	"github.com/stretchr/testify/require"
)

func TestGetNotificationTypes_Succeed(t *testing.T) {
	ctx := context.Background()
	mockClienter := mocks.NewNotificationKindsClienter(t)
	nts := notification_types.NewNotificationTypesManager(mockClienter)

	notificationTypes := []notification_types.NotificationType{
		{Category: "category1", Type: "type1", Description: "description1"},
	}

	mockClienter.On("Get", ctx).Return(notificationTypes, nil)

	res, err := nts.GetNotificationTypes(ctx)
	require.NoError(t, err)
	require.Equal(t, []notification_types.NotificationType{
		{Category: "category1", Type: "type1", Description: "description1"},
	}, res)
}

func TestGetNotificationTypes_Failed(t *testing.T) {
	ctx := context.Background()
	mockClienter := mocks.NewNotificationKindsClienter(t)
	nts := notification_types.NewNotificationTypesManager(mockClienter)

	mockClienter.On("Get", ctx).Return(nil, errors.New("error"))

	res, err := nts.GetNotificationTypes(ctx)
	require.EqualError(t, err, "failed to get notification kinds: error")
	require.Nil(t, res)
}
