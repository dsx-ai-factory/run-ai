package notifications

import (
	"github.com/run-ai/runai/control-plane/bff-service/internal/generated/pkg/api"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting/manager"
)

func TransformToNotificationsSettingsResponse(settingsList []setting_pkg.Settinger) any {
	response := &api.NotificationsSettingsGetAllResponse{}

	for _, s := range settingsList {
		switch v := s.(type) {
		case *AdminMessage:
			response.AdminMessage = api.GetAdminMessage{
				Value: api.GetAdminMessageValue{
					Id:            v.ID,
					Msg:           v.Msg,
					CreatedAt:     v.CreatedAt,
					EnableDismiss: v.EnableDismiss,
				},
				DisplayName: v.DisplayName(),
				Description: v.Description(),
			}
		}
	}
	return response
}
