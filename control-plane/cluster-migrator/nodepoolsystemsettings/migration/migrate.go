/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"context"
	"fmt"
	"time"

	migrator_config "github.com/run-ai/runai/control-plane/cluster-migrator/config"
	"github.com/run-ai/runai/control-plane/cluster-migrator/database"
	"github.com/run-ai/runai/control-plane/cluster-migrator/database/models"
	clusterServiceModels "github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

const (
	serviceMigrationsTimeout  = 10 * time.Minute
	serviceMigrationsPollRate = 5 * time.Second
)

func Migrate(conf *migrator_config.ConfigData) bool {
	if err := waitForClusterServiceMigrations(context.Background(), serviceMigrationsTimeout, serviceMigrationsPollRate); err != nil {
		logging.Errorf("Cluster-service migrations are not ready: %v", err)
		return true
	}

	tenants := database.GetTenantsToMigrate(conf.TenantsToMigrate, conf.TenantsToNotMigrate)

	return migrateTenants(conf, tenants)
}

func migrateTenants(conf *migrator_config.ConfigData, tenants []models.Tenant) bool {
	if database.IsNodepoolSystemSettingsMigrationCompleted() {
		logging.Infof("Nodepool-system-settings migration version %s has already been completed for all tenants",
			conf.MigrationVersion)
		return false
	}

	logging.Infof("Running nodepool-system-settings migration %s on tenants %+v", conf.MigrationVersion, tenants)
	hadFailure := false

	for _, tenant := range tenants {
		err := migrateTenant(tenant)

		hadFailure = hadFailure || err != nil
	}

	err := database.SetNodepoolSystemSettingsMigrationFinished(hadFailure)
	if err != nil {
		logging.Errorf("Failed to set nodepool-system-settings migration finished. err: %v", err)
		hadFailure = true
	}

	return hadFailure
}

func migrateTenant(tenant models.Tenant) (err error) {
	defer func() {
		if err != nil {
			logging.Errorf("Had an error migrating nodepool-system-settings for tenant %s. error: %v", tenant.Name, err)
			_ = database.SetNodepoolSystemSettingsMigrationStatusForTenant(tenant.Id, models.FAILED)
		} else {
			_ = database.SetNodepoolSystemSettingsMigrationStatusForTenant(tenant.Id, models.COMPLETED)
		}
	}()

	if database.IsNodepoolSystemSettingsMigrationCompletedForTenant(tenant.Id) {
		logging.Infof("Tenant %s has already been migrated for nodepool-system-settings", tenant.Name)
		return
	}

	startTime := time.Now()
	logging.Infof("Starting tenant %s nodepool-system-settings migration", tenant.Name)

	err = database.SetNodepoolSystemSettingsMigrationStatusForTenant(tenant.Id, models.STARTED)
	if err != nil {
		return
	}

	err = migrateSystemSettingsNodepools(tenant.Id)
	if err != nil {
		return
	}

	logging.Infof("Tenant %s nodepools migration has been completed in %d milli seconds", tenant.Name, time.Now().UnixMilli()-startTime.UnixMilli())
	return
}

func migrateSystemSettingsNodepools(tenantId int32) error {
	tenantSettingsMap, err := database.GetTenantSettingsMap(tenantId)
	if err != nil {
		return err
	}

	err = database.UpdateSystemSettingsOnTenantNodepools(context.Background(), tenantId, tenantSettingsMap)
	if err != nil {
		return err
	}

	return nil
}

// waitForClusterServiceMigrations blocks until the cluster-service DB schema contains
// the columns introduced by the 2.24.1 nodepool scheduling capabilities migration.
// We check for presence of both allow_adjustable_weights and allow_cpu_quota in the nodepools table.
func waitForClusterServiceMigrations(ctx context.Context, timeout time.Duration, pollInterval time.Duration) error {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context canceled while waiting for cluster-service schema: %w", ctx.Err())
		case <-timeoutTimer.C:
			return fmt.Errorf("timeout waiting for cluster-service schema (need columns %v and %v on table %v)",
				clusterServiceModels.AdjustableWeightsColumn,
				clusterServiceModels.CpuQuotaColumn,
				clusterServiceModels.NodepoolTable,
			)
		case <-ticker.C:
			db := runaidb.DB().WithContext(ctx).Actual()
			// Columns added by migration 02240001_nodepool_scheduling_capabilities.up.sql
			hasAdjustable := db.Migrator().HasColumn(&clusterServiceModels.Nodepool{}, clusterServiceModels.AdjustableWeightsColumn)
			hasCpuQuota := db.Migrator().HasColumn(&clusterServiceModels.Nodepool{}, clusterServiceModels.CpuQuotaColumn)
			hasQuota := db.Migrator().HasColumn(&clusterServiceModels.Nodepool{}, clusterServiceModels.QuotaColumn)
			if hasAdjustable && hasCpuQuota && hasQuota {
				return nil
			}
		}
	}
}
