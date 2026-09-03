/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"fmt"

	service_config "github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"gorm.io/gorm/clause"
)

func IsMigrationCompletedForTenant(tenantId int32) bool {
	mc := models.MigrationControl{}

	if err := sourceClient.
		Where("version = ?", config.Get[*service_config.ConfigData]().MigrationVersion).
		Where("tenant_id = ?", tenantId).
		Where("status = ?", models.COMPLETED).
		First(&mc); err != nil {
		logging.Infof(fmt.Sprintf("Migration version %s is not completed for tenant %d", mc.Version, mc.TenantId), err)
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

	if _, err := sourceClient.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "version"}, {Name: "tenant_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status"}),
	}).Create(mc); err != nil {
		logging.Errorf(fmt.Sprintf("Failed to save version %s for tennat %d", mc.Version, mc.TenantId), err)
		return err
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

	if _, err := sourceClient.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "version"}, {Name: "tenant_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status"}),
	}).Create(mc); err != nil {
		logging.Errorf(fmt.Sprintf("Failed to save version %s as completed", mc.Version), err)
		return err
	}
	return nil
}

func IsMigrationCompleted() bool {
	mc := models.MigrationControl{}

	if err := sourceClient.Where("version = ?", config.Get[*service_config.ConfigData]().MigrationVersion).Where("tenant_id IS NULL").Where("status = ?", models.COMPLETED).First(&mc); err != nil {
		logging.Infof(fmt.Sprintf("Migration version %s is not completed", mc.Version), err)
		return false
	}

	return true
}
