package clients

import (
	"context"

	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/generated/config_client"
)

type NotificationKinds struct {
	client *config_client.RegisterNotificationAPIService
}

func NewNotificationKinds() *NotificationKinds {
	return &NotificationKinds{client: GetConfigClient().RegisterNotificationAPI}
}

func (nk *NotificationKinds) GetNotificationKinds(ctx context.Context) ([]config_client.NotificationKind, error) {
	res, _, err := nk.client.GetNotificationKinds(ctx).Execute()
	return res, err
}
