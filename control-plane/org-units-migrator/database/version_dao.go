package database

import (
	"fmt"

	service_config "github.com/run-ai/runai/control-plane/org-units-migrator/config"
	"github.com/run-ai/runai/control-plane/org-units-migrator/database/models"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"gorm.io/gorm/clause"
)

func IsMigrationCompletedForTenant(tenantId int32) bool {
	mc := models.MigrationControl{}

	result := sourceClient.
		Where("version = ?", config.Get[*service_config.ConfigData]().MigrationVersion).
		Where("tenant_id = ?", tenantId).
		Where("status = ?", models.COMPLETED).
		First(&mc)
	if result.Error != nil {
		logging.Infof(fmt.Sprintf("Migration version %s is not completed for tenant %d", mc.Version, mc.TenantId), result.Error)
		return false
	}

	return true
}

func SetMigrationStatusForTenant(tenantId int32, status models.MigrationStatus) error {
	mc := models.MigrationControl{
		Version:  config.Get[*service_config.ConfigData]().MigrationVersion,
		TenantId: &tenantId,
		Status:   status,
	}

	result := sourceClient.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "version"}, {Name: "tenant_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status"}),
	}).Create(mc)
	if result.Error != nil {
		logging.Errorf(fmt.Sprintf("Failed to save version %s for tennat %d", mc.Version, mc.TenantId), result.Error)
		return result.Error
	}

	return nil
}

func SetMigrationCompleted() error {
	if !config.Get[*service_config.ConfigData]().ShouldMarkCompleted {
		return nil
	}

	mc := models.MigrationControl{
		Version:  config.Get[*service_config.ConfigData]().MigrationVersion,
		TenantId: nil,
		Status:   models.COMPLETED,
	}

	result := sourceClient.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "version"}, {Name: "tenant_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status"}),
	}).Create(mc)
	if result.Error != nil {
		logging.Errorf(fmt.Sprintf("Failed to save version %s as completed", mc.Version), result.Error)
		return result.Error
	}
	return nil
}

func IsMigrationCompleted() bool {
	mc := models.MigrationControl{}

	result := sourceClient.Where("version = ?", config.Get[*service_config.ConfigData]().MigrationVersion).Where("tenant_id IS NULL").Where("status = ?", models.COMPLETED).First(&mc)
	if result.Error != nil {
		logging.Infof(fmt.Sprintf("Migration version %s is not completed", mc.Version), result.Error)
		return false
	}

	return true
}
