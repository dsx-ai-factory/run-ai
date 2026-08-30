package providers

import (
	"context"
	"strconv"

	bff_openapi "github.com/run-ai/runai/control-plane/bff-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"

	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/autorization"
	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/business/settings"
)

type BffWorkloadsProvider struct {
	client *bff_openapi.APIClient
}

func NewBffWorkloadsProvider(client *bff_openapi.APIClient) *BffWorkloadsProvider {
	return &BffWorkloadsProvider{client: client}
}

func (p *BffWorkloadsProvider) FetchSettings(ctx context.Context, tenantID int) (map[string]settings.SettingOverride, error) {
	result := make(map[string]settings.SettingOverride)

	req := p.client.SettingsAPI.GetWorkloadsSettings(ctx)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	resp, _, err := req.Execute()
	if err != nil {
		logging.Errorf("failed to fetch workloads settings from bff: %v", err)
		return result, nil
	}

	if resp.EnableNim != nil {
		override := settings.SettingOverride{}
		if valWrapper, ok := resp.EnableNim.GetValueOk(); ok && valWrapper != nil {
			if v, ok := valWrapper.GetEnabledOk(); ok && v != nil {
				override.Value = *v
			}
		}
		if dn, ok := resp.EnableNim.GetDisplayNameOk(); ok && dn != nil {
			override.DisplayName = *dn
		}
		if desc, ok := resp.EnableNim.GetDescriptionOk(); ok && desc != nil {
			override.Description = *desc
		}
		if override.Value != nil {
			result[settings.SettingKeyEnableNIM] = override
		}
	}

	if resp.DeletedWorkloadsView != nil {
		override := settings.SettingOverride{}
		if valWrapper, ok := resp.DeletedWorkloadsView.GetValueOk(); ok && valWrapper != nil {
			if v, ok := valWrapper.GetEnabledOk(); ok && v != nil {
				override.Value = *v
			}
		}
		if dn, ok := resp.DeletedWorkloadsView.GetDisplayNameOk(); ok && dn != nil {
			override.DisplayName = *dn
		}
		if desc, ok := resp.DeletedWorkloadsView.GetDescriptionOk(); ok && desc != nil {
			override.Description = *desc
		}
		if override.Value != nil {
			result[settings.SettingKeyDeletedWorkloadsView] = override
		}
	}

	if resp.AllowGPT2Model != nil {
		override := settings.SettingOverride{}
		if valWrapper, ok := resp.AllowGPT2Model.GetValueOk(); ok && valWrapper != nil {
			if v, ok := valWrapper.GetEnabledOk(); ok && v != nil {
				override.Value = *v
			}
		}
		if dn, ok := resp.AllowGPT2Model.GetDisplayNameOk(); ok && dn != nil {
			override.DisplayName = *dn
		}
		if desc, ok := resp.AllowGPT2Model.GetDescriptionOk(); ok && desc != nil {
			override.Description = *desc
		}
		if stage, ok := resp.AllowGPT2Model.GetStageOk(); ok && stage != nil {
			override.Stage = *stage
		}
		if override.Value != nil {
			result[settings.SettingKeyAllowGPT2Model] = override
		}
	}

	if resp.EnableModelCatalog != nil {
		override := settings.SettingOverride{}
		if valWrapper, ok := resp.EnableModelCatalog.GetValueOk(); ok && valWrapper != nil {
			if v, ok := valWrapper.GetEnabledOk(); ok && v != nil {
				override.Value = *v
			}
		}
		if dn, ok := resp.EnableModelCatalog.GetDisplayNameOk(); ok && dn != nil {
			override.DisplayName = *dn
		}
		if desc, ok := resp.EnableModelCatalog.GetDescriptionOk(); ok && desc != nil {
			override.Description = *desc
		}
		if stage, ok := resp.EnableModelCatalog.GetStageOk(); ok && stage != nil {
			override.Stage = *stage
		}
		if override.Value != nil {
			result[settings.SettingKeyEnableModelCatalog] = override
		}
	}

	return result, nil
}
