package adapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	idm_openapi "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/autorization"
)

type IdentityManagerAdapter struct {
	SettingsAPI idm_openapi.SettingsAPI
}

func (a *IdentityManagerAdapter) UpdateLogoutRedirectUri(ctx context.Context, tenantID int, uri interface{}) error {
	var uriStr string
	if uri != nil {
		uriStr = fmt.Sprintf("%v", uri)
	}
	body := idm_openapi.UpdateSecuritySettingByKeyRequest{
		UpdateLogoutRedirectUriSettingValue: &idm_openapi.UpdateLogoutRedirectUriSettingValue{},
	}
	body.UpdateLogoutRedirectUriSettingValue.SetUri(uriStr)

	req := a.SettingsAPI.UpdateSecuritySettingByKey(ctx, idm_openapi.SETTINGSKEYENUM_LOGOUT_REDIRECT_URI).
		UpdateSecuritySettingByKeyRequest(body)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	resp, err := req.Execute()
	if err != nil {
		return fmt.Errorf("UpdateLogoutRedirectUri: %w", err)
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("UpdateLogoutRedirectUri returned status %d", resp.StatusCode)
	}
	return nil
}

func (a *IdentityManagerAdapter) UpdateBrowserSessionTimeout(ctx context.Context, tenantID int, timeout interface{}) error {
	var timeoutVal float32
	switch v := timeout.(type) {
	case float64:
		timeoutVal = float32(v)
	case float32:
		timeoutVal = v
	case int:
		timeoutVal = float32(v)
	case int64:
		timeoutVal = float32(v)
	default:
		return fmt.Errorf("invalid timeout value type: %T", timeout)
	}

	body := idm_openapi.UpdateSecuritySettingByKeyRequest{
		UpdateBrowserSessionTimeoutSettingValue: &idm_openapi.UpdateBrowserSessionTimeoutSettingValue{},
	}
	body.UpdateBrowserSessionTimeoutSettingValue.SetIdleTimeout(timeoutVal)

	req := a.SettingsAPI.UpdateSecuritySettingByKey(ctx, idm_openapi.SETTINGSKEYENUM_BROWSER_SESSION_TIMEOUT).
		UpdateSecuritySettingByKeyRequest(body)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	resp, err := req.Execute()
	if err != nil {
		return fmt.Errorf("UpdateBrowserSessionTimeout: %w", err)
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("UpdateBrowserSessionTimeout returned status %d", resp.StatusCode)
	}
	return nil
}
