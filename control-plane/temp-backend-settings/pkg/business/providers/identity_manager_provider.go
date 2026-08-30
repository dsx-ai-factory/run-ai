package providers

import (
	"context"
	"strconv"

	idm_openapi "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"

	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/autorization"
	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/business/settings"
)

type IdentityManagerProvider struct {
	client *idm_openapi.APIClient
}

func NewIdentityManagerProvider(client *idm_openapi.APIClient) *IdentityManagerProvider {
	return &IdentityManagerProvider{client: client}
}

func (p *IdentityManagerProvider) FetchSettings(ctx context.Context, tenantID int) (map[string]settings.SettingOverride, error) {
	result := make(map[string]settings.SettingOverride)

	req := p.client.SettingsAPI.GetSecuritySettings(ctx)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	resp, _, err := req.Execute()
	if err != nil {
		logging.Errorf("failed to fetch security settings from identity manager: %v", err)
		return result, nil
	}

	if resp.BrowserSessionTimeout != nil {
		override := settings.SettingOverride{}
		if valWrapper, ok := resp.BrowserSessionTimeout.GetValueOk(); ok && valWrapper != nil {
			if timeout, ok := valWrapper.GetIdleTimeoutOk(); ok && timeout != nil {
				override.Value = float64(*timeout)
			}
		}
		if dn, ok := resp.BrowserSessionTimeout.GetDisplayNameOk(); ok && dn != nil {
			override.DisplayName = *dn
		}
		if desc, ok := resp.BrowserSessionTimeout.GetDescriptionOk(); ok && desc != nil {
			override.Description = *desc
		}
		if override.Value != nil {
			result[settings.SettingKeyInactivityMinutesTillLogout] = override
		}
	}

	if resp.LogoutRedirectUri != nil {
		override := settings.SettingOverride{}
		if valWrapper, ok := resp.LogoutRedirectUri.GetValueOk(); ok && valWrapper != nil {
			if uri, ok := valWrapper.GetUriOk(); ok && uri != nil {
				override.Value = *uri
			}
		}
		if dn, ok := resp.LogoutRedirectUri.GetDisplayNameOk(); ok && dn != nil {
			override.DisplayName = *dn
		}
		if desc, ok := resp.LogoutRedirectUri.GetDescriptionOk(); ok && desc != nil {
			override.Description = *desc
		}
		if override.Value != nil {
			result[settings.SettingKeyLogoutURI] = override
		}
	}

	return result, nil
}
