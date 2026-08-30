package adapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	authorization_openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/autorization"
)

type AuthorizationAdapter struct {
	SettingsAPI authorization_openapi.SettingsAPI
}

func (a *AuthorizationAdapter) UpdateEnableScopedAuthorization(ctx context.Context, tenantID int, value interface{}) error {
	return a.updateBooleanSetting(ctx, tenantID, authorization_openapi.AUTHORIZATIONSETTINGSKEYENUM_ENABLE_SCOPED_AUTHORIZATION, value)
}

func (a *AuthorizationAdapter) UpdateTenantIsCspManaged(ctx context.Context, tenantID int, value interface{}) error {
	return a.updateBooleanSetting(ctx, tenantID, authorization_openapi.AUTHORIZATIONSETTINGSKEYENUM_TENANT_IS_CSP_MANAGED, value)
}

func (a *AuthorizationAdapter) updateBooleanSetting(ctx context.Context, tenantID int, key authorization_openapi.AuthorizationSettingsKeyEnum, value interface{}) error {
	boolVal, ok := value.(bool)
	if !ok {
		return fmt.Errorf("update %s: expected bool, got %T", key, value)
	}

	body := authorization_openapi.UpdateAuthorizationSettingByKeyRequest{}
	body.SetEnabled(boolVal)

	req := a.SettingsAPI.UpdateAuthorizationSettingByKey(ctx, key).
		UpdateAuthorizationSettingByKeyRequest(body)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	resp, err := req.Execute()
	if err != nil {
		return fmt.Errorf("update %s: %w", key, err)
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("update %s returned status %d", key, resp.StatusCode)
	}
	return nil
}
