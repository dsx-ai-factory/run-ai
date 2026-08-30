package notification_types

import (
	"context"
)

type NotificationType struct {
	Origin      string
	Category    string
	Type        string
	Description string
	Template    string
}

type NotificationKindsClienter interface {
	Get(ctx context.Context) ([]NotificationType, error)
}
