package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting"
	"github.com/run-ai/runai/runai-common-packages/setting/manager"
)

const (
	ErrUnknownSettingKey = "unknown setting key"
	ClusterCategory      = "clusters"
)

// SettingsManager handles retrieval, update, and transformation of cluster-service settings.
type SettingsManager struct{}

// NewSettingsManager creates a new SettingsManager.
func NewSettingsManager() *SettingsManager {
	return &SettingsManager{}
}

func (s *SettingsManager) GetSettingByKey(key api.ClusterSettingsKeyEnum, tenantID string) (manager.Settinger, error) {
	switch key {
	case api.CLUSTERSETTINGSKEYENUM_ENABLE_ADVANCED_GPU_METRICS:
		setting, err := setting_pkg.Get[*EnableAdvancedGPUMetrics](tenantID)
		if err != nil {
			return nil, err
		}
		return setting, nil
	default:
		return nil, fmt.Errorf("%s: %s", ErrUnknownSettingKey, key)
	}
}

func (s *SettingsManager) UpdateSettingByKey(ctx context.Context, key api.ClusterSettingsKeyEnum, tenantID string, updateRequest api.UpdateClusterSettingByKeyRequest) error {
	if updateRequest.Enabled == nil {
		return errors.New("enabled is required")
	}
	switch key {
	case api.CLUSTERSETTINGSKEYENUM_ENABLE_ADVANCED_GPU_METRICS:
		settingValue := EnableAdvancedGPUMetrics{Enabled: *updateRequest.Enabled}
		return setting_pkg.Update[*EnableAdvancedGPUMetrics](tenantID, settingValue).UpdatedByUser().Execute()
	default:
		return fmt.Errorf("%s: %s", ErrUnknownSettingKey, key)
	}
}

func (s *SettingsManager) TransformSettingsToResponse(settingsList []manager.Settinger) *api.ClusterSettingsGetResponse {
	response := &api.ClusterSettingsGetResponse{}
	for _, setting := range settingsList {
		switch v := setting.(type) {
		case *EnableAdvancedGPUMetrics:
			response.EnableAdvancedGpuMetrics = api.EnableAdvancedGpuMetricsSetting{
				Enabled:     &v.Enabled,
				DisplayName: v.DisplayName(),
				Description: v.Description(),
				Stage:       string(v.Stage()),
			}
		}
	}
	return response
}
