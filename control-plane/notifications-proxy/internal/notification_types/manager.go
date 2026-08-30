package notification_types

import (
	"context"
	"fmt"
)

type NotificationTypesManager struct {
	notificationTypesClient NotificationKindsClienter
}

func NewNotificationTypesManager(client NotificationKindsClienter) *NotificationTypesManager {
	return &NotificationTypesManager{
		notificationTypesClient: client,
	}
}

func (ntm *NotificationTypesManager) GetNotificationTypes(ctx context.Context) ([]NotificationType, error) {
	notificationKinds, err := ntm.notificationTypesClient.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification kinds: %w", err)
	}
	return notificationKinds, nil
}
