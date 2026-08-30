package database

import (
	"strings"

	"github.com/run-ai/runai/control-plane/org-units-migrator/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

const (
	ProjectManualOverQuota = "project.quota.manual_over_quota"
	CPUResourcesQuota      = "project.quota.cpu_resources"
)

var MigrationSettings = []string{
	ProjectManualOverQuota,
	CPUResourcesQuota,
}

func GetTenantSettingsMap(tenantId int32) (map[string]bool, error) {
	settings := []models.TenantSetting{}

	result := sourceClient.
		Where("tenant_id = ?", tenantId).
		Where("key IN ?", MigrationSettings).
		Find(&settings)
	if result.Error != nil {
		logging.Errorf("Failed to query tenant settings for tenant %d. err: %v", tenantId, result.Error)
		return nil, result.Error
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
