package notification_types

import (
	"context"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/notification_types"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/generated/config_client"
)

type NotificationKindsClienter interface {
	GetNotificationKinds(ctx context.Context) ([]config_client.NotificationKind, error)
}

type Client struct {
	notificationKindsClient NotificationKindsClienter
}

func NewClient() *Client {
	return &Client{notificationKindsClient: clients.NewNotificationKinds()}
}

func (c *Client) Get(ctx context.Context) ([]notification_types.NotificationType, error) {
	notificationTypes, err := c.notificationKindsClient.GetNotificationKinds(ctx)
	if err != nil {
		return nil, err
	}

	return toNotifications(notificationTypes), nil
}

func toNotifications(notificationTypes []config_client.NotificationKind) []notification_types.NotificationType {
	res := []notification_types.NotificationType{}
	for _, nt := range notificationTypes {
		res = append(res, notification_types.NotificationType{
			Origin:      nt.Origin,
			Category:    nt.Category,
			Type:        nt.Type,
			Description: nt.Description,
			Template:    nt.Template,
		})
	}
	return res
}
