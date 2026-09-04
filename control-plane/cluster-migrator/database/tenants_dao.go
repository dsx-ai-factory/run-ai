/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

func GetTenantsToMigrate(tenantsToMigrate []string, tenantsToNotMigrate []string) []models.Tenant {
	tenants := []models.Tenant{}
	query := globalClient.Model(models.Tenant{})
	// If both are not set, we will migrate all tenants, so SH does not need any special configuration
	if len(tenantsToMigrate) != 0 {
		query = query.Where("id IN ?", tenantsToMigrate)
	}

	if len(tenantsToNotMigrate) != 0 {
		query = query.Where("id NOT IN ?", tenantsToNotMigrate)
	}

	if err := query.Find(&tenants); err != nil {
		logging.Fatalf("Failed to query tenants for migration. err: %v", err)
	}

	return tenants
}
