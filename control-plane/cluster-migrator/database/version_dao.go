package database

import (
	"github.com/run-ai/runai/control-plane/cluster-migrator/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm/clause"
)

func IsNodepoolsMigrationCompletedForTenant(tenantId int32) bool {
	mc := models.NodepoolsMigrationControl{}

	err := sourceClient.
		Where("version = ?", conf.MigrationVersion).
		Where("tenant_id = ?", tenantId).
		Where("status = ?", models.COMPLETED).
		First(&mc)
	if err != nil {
		logging.Infof("Migration version %s is not completed for tenant %d: %v", mc.Version, tenantId, err)
		return false
	}

	return true
}

func SetNodepoolsMigrationStatusForTenant(tenantId int32, status models.MigrationStatus) error {
	mc := models.NodepoolsMigrationControl{
		Version:  conf.MigrationVersion,
		TenantId: &tenantId,
		Status:   status,
	}

	_, err := sourceClient.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "version"}, {Name: "tenant_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status"}),
	}).Create(mc)
	if err != nil {
		logging.Errorf("Failed to save version %s for tenant %d: %v", mc.Version, tenantId, err)
		return err
	}

	return nil
}

func SetMigrationCompleted() error {
	if !conf.ShouldMarkCompleted {
		return nil
	}

	mc := models.NodepoolsMigrationControl{
		Version:  conf.MigrationVersion,
		TenantId: nil,
		Status:   models.COMPLETED,
	}

	_, err := sourceClient.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "version"}, {Name: "tenant_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status"}),
	}).Create(mc)
	if err != nil {
		logging.Errorf("Failed to save version %s as completed: %v", mc.Version, err)
		return err
	}
	return nil
}

func IsMigrationCompleted() bool {
	mc := models.NodepoolsMigrationControl{}

	err := sourceClient.Where("version = ?", conf.MigrationVersion).Where("tenant_id IS NULL").Where("status = ?", models.COMPLETED).First(&mc)
	if err != nil {
		logging.Infof("Migration version %s is not completed: %v", mc.Version, err)
		return false
	}

	return true
}
