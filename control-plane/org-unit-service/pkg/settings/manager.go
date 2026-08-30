package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting"
	"github.com/run-ai/runai/runai-common-packages/setting/manager"
)

const (
	ErrUnknownSettingKey = "unknown setting key"
	OrgUnitCategory      = "org-unit"
)

// SettingsManager handles retrieval, update, and transformation of org-unit-service settings.
type SettingsManager struct{}

// NewSettingsManager creates a new SettingsManager.
func NewSettingsManager() *SettingsManager {
	return &SettingsManager{}
}

func (s *SettingsManager) GetSettingByKey(key api.OrgUnitSettingsKeyEnum, tenantID string) (manager.Settinger, error) {
	switch key {
	case api.ORGUNITSETTINGSKEYENUM_LIMIT_QUOTA_OVER_SUBSCRIPTION:
		setting, err := setting_pkg.Get[*LimitQuotaOverSubscription](tenantID)
		if err != nil {
			return nil, err
		}
		return setting, nil
	default:
		return nil, fmt.Errorf("%s: %s", ErrUnknownSettingKey, key)
	}
}

func (s *SettingsManager) UpdateSettingByKey(ctx context.Context, key api.OrgUnitSettingsKeyEnum, tenantID string, updateRequest api.UpdateOrgUnitSettingByKeyRequest) error {
	if updateRequest.Enabled == nil {
		return errors.New("enabled is required")
	}
	switch key {
	case api.ORGUNITSETTINGSKEYENUM_LIMIT_QUOTA_OVER_SUBSCRIPTION:
		settingValue := LimitQuotaOverSubscription{Enabled: *updateRequest.Enabled}
		return setting_pkg.Update[*LimitQuotaOverSubscription](tenantID, settingValue).UpdatedByUser().Execute()
	default:
		return fmt.Errorf("%s: %s", ErrUnknownSettingKey, key)
	}
}

func (s *SettingsManager) TransformSettingsToResponse(settingsList []manager.Settinger) *api.OrgUnitSettingsGetResponse {
	response := &api.OrgUnitSettingsGetResponse{}
	for _, setting := range settingsList {
		switch v := setting.(type) {
		case *LimitQuotaOverSubscription:
			response.LimitQuotaOverSubscription = api.OrgUnitSetting{
				Enabled:     &v.Enabled,
				DisplayName: v.DisplayName(),
				Description: v.Description(),
				Stage:       string(v.Stage()),
			}
		}
	}
	return response
}
