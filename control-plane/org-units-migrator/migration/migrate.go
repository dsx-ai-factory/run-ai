/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

func Migrate(conf *config.ConfigData, tenants []models.Tenant) bool {
	logging.Infof("Running migration %s on tenants %+v", conf.MigrationVersion, tenants)
	hadFailure := false

	for _, tenant := range tenants {
		if database.IsMigrationCompletedForTenant(tenant.Id) {
			logging.Infof("Tenant %s has already been migrated", tenant.Name)
			continue
		}
		startTime := time.Now()
		logging.Debugf("Starting tenant %s migration", tenant.Name)

		database.SetMigrationStatusForTenant(tenant.Id, models.STARTED)

		settings, err := database.GetTenantSettingsMap(tenant.Id)
		if err != nil {
			database.SetMigrationStatusForTenant(tenant.Id, models.FAILED)
			hadFailure = true
			continue
		}

		departments, err := database.GetTenantDepartmentsV1(tenant.Id)
		if err != nil {
			database.SetMigrationStatusForTenant(tenant.Id, models.FAILED)
			hadFailure = true
			continue
		}

		err = CreateDepartments(tenant, departments, settings)
		if err != nil {
			database.SetMigrationStatusForTenant(tenant.Id, models.FAILED)
			hadFailure = true
			continue
		}

		nodeTypes, err := database.GetTenantNodeTypes(tenant.Id)
		if err != nil {
			database.SetMigrationStatusForTenant(tenant.Id, models.FAILED)
			hadFailure = true
			continue
		}

		err = CreateNodeTypes(nodeTypes)
		if err != nil {
			database.SetMigrationStatusForTenant(tenant.Id, models.FAILED)
			hadFailure = true
			continue
		}
		projects, err := database.GetTenantProjectsV1(tenant.Id)
		if err != nil {
			database.SetMigrationStatusForTenant(tenant.Id, models.FAILED)
			hadFailure = true
			continue
		}

		err = CreateProjects(tenant, projects, settings)
		if err != nil {
			database.SetMigrationStatusForTenant(tenant.Id, models.FAILED)
			hadFailure = true
			continue
		}

		database.SetMigrationStatusForTenant(tenant.Id, models.COMPLETED)
		logging.Infof("Tenant %s migration has been completed in %d milli seconds", tenant.Name, time.Now().UnixMilli()-startTime.UnixMilli())
	}

	// In any case of failure mark migration failed for tenant
	return hadFailure
}
