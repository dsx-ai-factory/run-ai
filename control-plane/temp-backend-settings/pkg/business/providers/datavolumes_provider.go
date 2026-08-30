package providers

import (
	"context"
	"strconv"

	dv_openapi "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"

	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/autorization"
	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/business/settings"
)

type DatavolumesProvider struct {
	client *dv_openapi.APIClient
}

func NewDatavolumesProvider(client *dv_openapi.APIClient) *DatavolumesProvider {
	return &DatavolumesProvider{client: client}
}

func (p *DatavolumesProvider) FetchSettings(ctx context.Context, tenantID int) (map[string]settings.SettingOverride, error) {
	result := make(map[string]settings.SettingOverride)

	req := p.client.SettingsAPI.GetDatavolumesSettings(ctx)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	resp, _, err := req.Execute()
	if err != nil {
		logging.Errorf("failed to fetch settings from datavolumes-service: %v", err)
		return result, nil
	}

	if resp.EnableDataVolumes != nil {
		override := settings.SettingOverride{}
		if val, ok := resp.EnableDataVolumes.GetValueOk(); ok {
			if enabled, ok := val.GetEnabledOk(); ok && enabled != nil {
				override.Value = *enabled
			}
		}
		if dn, ok := resp.EnableDataVolumes.GetDisplayNameOk(); ok && dn != nil {
			override.DisplayName = *dn
		}
		if desc, ok := resp.EnableDataVolumes.GetDescriptionOk(); ok && desc != nil {
			override.Description = *desc
		}
		if stage, ok := resp.EnableDataVolumes.GetStageOk(); ok && stage != nil {
			override.Stage = *stage
		}
		if override.Value != nil {
			result[settings.SettingKeyEnableDataVolumes] = override
		}
	}

	return result, nil
}
