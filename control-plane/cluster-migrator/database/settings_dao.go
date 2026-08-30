package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/run-ai/runai/control-plane/cluster-migrator/database/models"
	clusterServiceModels "github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm/clause"
)

const (
	CpuQuotaSystemSettingFlag = "project.quota.cpu_resources"
	WeightsSystemSettingFlag  = "project.quota.manual_over_quota"
)

var MigrationSettings = []string{
	CpuQuotaSystemSettingFlag,
	WeightsSystemSettingFlag,
}

func GetTenantSettingsMap(tenantId int32) (map[string]bool, error) {
	settings := []models.TenantSetting{}

	err := sourceClient.
		Table(models.TenantSetting{}.TableName()).
		Where("tenant_id = ?", tenantId).
		Where("key IN ?", MigrationSettings).
		Find(&settings)
	if err != nil {
		logging.Errorf("Failed to query tenant settings for tenant %d. err: %v", tenantId, err)
		return nil, err
	}

	settingsMap := map[string]bool{}

	for _, key := range MigrationSettings {
		settingsMap[key] = false
	}

	for _, setting := range settings {
		settingsMap[setting.Key] = strings.ToLower(setting.Value) == "true"
	}
	return settingsMap, nil
}

func UpdateSystemSettingsOnTenantNodepools(ctx context.Context, tenantId int32, settingsToSet map[string]bool) error {
	updateMap := map[string]interface{}{
		"updated_at": time.Now(),
	}

	for settingName, enabled := range settingsToSet {
		switch settingName {
		case CpuQuotaSystemSettingFlag:
			if enabled {
				updateMap[clusterServiceModels.CpuQuotaColumn] = clause.Expr{
					SQL: fmt.Sprintf("CASE WHEN %s = true THEN true ELSE false END", clusterServiceModels.QuotaColumn),
				}
			} else {
				updateMap[clusterServiceModels.CpuQuotaColumn] = false
			}
		case WeightsSystemSettingFlag:
			updateMap[clusterServiceModels.AdjustableWeightsColumn] = enabled
		default:
			return fmt.Errorf("unknown setting enum value %q", settingName)
		}
	}

	_, err := targetClient.WithContext(ctx).
		Table(clusterServiceModels.NodepoolTable).
		Where("tenant_id = ?", tenantId).
		Where("deleted_at IS NULL").
		UpdateColumns(updateMap)
	return err
}
