/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"github.com/run-ai/runai/control-plane/org-units-migrator/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

func GetTenantsToMigrate(tenantsToMigrate []string, tenantsToNotMigrate []string) []models.Tenant {
	tenants := []models.Tenant{}
	query := globalClient.Model(models.Tenant{})

	// If both are not set, we will migrate all tenants, so SH does not need any special configuration
	if len(tenantsToMigrate) != 0 {
		query.Where("id IN ?", tenantsToMigrate)
	}

	if len(tenantsToNotMigrate) != 0 {
		query.Where("id NOT IN ?", tenantsToNotMigrate)
	}

	result := query.Find(&tenants)
	if result.Error != nil {
		logging.Fatalf("Failed to query tenants for migration. err: %v", result.Error)
	}

	return tenants
}
