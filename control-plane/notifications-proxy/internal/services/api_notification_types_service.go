package services

import (
	"context"
	"fmt"
	"net/http"

	nt_client "github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients/notification_types"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/notification_types"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type NotificationTypesManager interface {
	GetNotificationTypes(ctx context.Context) ([]notification_types.NotificationType, error)
}

type NotificationTypesAPIService struct {
	manager NotificationTypesManager
}

func NewNotificationTypesAPIService() *NotificationTypesAPIService {
	return &NotificationTypesAPIService{
		manager: notification_types.NewNotificationTypesManager(nt_client.NewClient()),
	}
}

func (nts *NotificationTypesAPIService) GetNotificationTypes(ctx context.Context) (api.ImplResponse, error) {
	notificationTypes, err := nts.manager.GetNotificationTypes(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get notification types: %v", err)
		return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to get notification types")
	}

	return api.Response(http.StatusOK, toApiNotificationsType(notificationTypes)), nil
}

func toApiNotificationsType(notificationKinds []notification_types.NotificationType) []api.NotificationType {
	res := []api.NotificationType{}
	for _, nk := range notificationKinds {
		res = append(res, api.NotificationType{
			Category:    nk.Category,
			Type:        nk.Type,
			Description: nk.Description,
		})
	}
	return res
}
