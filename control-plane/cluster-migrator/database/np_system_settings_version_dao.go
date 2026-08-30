package database

import (
	"github.com/run-ai/runai/control-plane/cluster-migrator/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm/clause"
)

func IsNodepoolSystemSettingsMigrationCompletedForTenant(tenantId int32) bool {
	mc := models.NodepoolSystemSettingsMigrationControl{}

	err := targetClient.
		Where("version = ?", conf.MigrationVersion).
		Where("tenant_id = ?", tenantId).
		Where("status = ?", models.COMPLETED).
		First(&mc)
	if err != nil {
		logging.Infof("Migration version %s of nodepool-system-settings is not completed for tenant %d: %v",
			mc.Version, tenantId, err)
		return false
	}

	return true
}

func SetNodepoolSystemSettingsMigrationStatusForTenant(tenantId int32, status models.MigrationStatus) error {
	mc := models.NodepoolSystemSettingsMigrationControl{
		Version:  conf.MigrationVersion,
		TenantId: &tenantId,
		Status:   status,
	}

	_, err := targetClient.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "version"}, {Name: "tenant_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status"}),
	}).Create(mc)
	if err != nil {
		logging.Errorf("Failed to save version %s for tenant %d for nodepool-system-settings migration: %v",
			mc.Version, tenantId, err)
		return err
	}

	return nil
}

func IsNodepoolSystemSettingsMigrationCompleted() bool {
	mc := models.NodepoolSystemSettingsMigrationControl{}

	err := targetClient.
		Where("version = ?", conf.MigrationVersion).
		Where("tenant_id IS NULL").
		Where("status = ?", models.COMPLETED).
		First(&mc)
	if err != nil {
		logging.Infof("Migration version %s of nodepool-system-settings is not completed for all tenants: %v",
			mc.Version, err)
		return false
	}

	return true
}

func SetNodepoolSystemSettingsMigrationFinished(hadFailure bool) error {
	if !conf.ShouldMarkCompleted {
		return nil
	}

	status := models.COMPLETED
	if hadFailure {
		status = models.FAILED
	}

	mc := models.NodepoolSystemSettingsMigrationControl{
		Version:  conf.MigrationVersion,
		TenantId: nil,
		Status:   status,
	}

	_, err := targetClient.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "version"}, {Name: "tenant_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status"}),
	}).Create(mc)
	if err != nil {
		logging.Errorf("Failed to save version %s as %v for nodepool-system-settings: %v",
			mc.Version, status, err)
		return err
	}
	return nil
}
