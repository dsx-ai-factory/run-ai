package adapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	bff_openapi "github.com/run-ai/runai/control-plane/bff-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/autorization"
)

type BffNotificationsAdapter struct {
	NotificationsSettingsAPI bff_openapi.SettingsAPI
}

func (a *BffNotificationsAdapter) UpdateAdminMessage(ctx context.Context, tenantID int, value interface{}) error {
	valMap, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("UpdateAdminMessage: expected map, got %T", value)
	}

	body := bff_openapi.UpdateAdminMessage{}
	if msg, ok := valMap["msg"].(string); ok {
		body.SetMsg(msg)
	}
	if enableDismiss, ok := valMap["enableDismiss"].(bool); ok {
		body.SetEnableDismiss(enableDismiss)
	}

	req := a.NotificationsSettingsAPI.UpdateNotificationsSettingByKey(
		ctx, bff_openapi.NOTIFICATIONSSETTINGSKEYENUM_ADMIN_MESSAGE,
	).UpdateAdminMessage(body)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	resp, err := req.Execute()
	if err != nil {
		return fmt.Errorf("UpdateAdminMessage: %w", err)
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("UpdateAdminMessage returned status %d", resp.StatusCode)
	}
	return nil
}
