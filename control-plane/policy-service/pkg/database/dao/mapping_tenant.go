/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

// DeleteMappingsOfTenant removes every tenant-scoped mapping row owned by the
// given tenant: mapping_selectors, mapping_field_groups, mapping_formatters
// (only rows with scope=tenant), and any leftover row in mapping_locking. system
// rows are never touched. used by the tenant deletion handler.
func DeleteMappingsOfTenant(ctx context.Context, db orm.DB, tenantId int32) *api.Error {
	selectorsKey := dbmodel.MappingSelectors{
		TenantId:  tenantId,
		ScopeType: api.MAPPINGSCOPE_TENANT,
	}
	rowsAffected, err := db.Where(&selectorsKey).Delete(&dbmodel.MappingSelectors{})
	if err != nil {
		dbErr := database.Error(database.DELETE, database.MappingSelectors).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("failed to delete mapping_selectors for tenant %d: %s: %v", tenantId, dbErr.Message, err)
		return dbErr
	}

	fieldGroupsKey := dbmodel.MappingFieldGroups{
		TenantId:  tenantId,
		ScopeType: api.MAPPINGSCOPE_TENANT,
	}
	rowsAffected, err = db.Where(&fieldGroupsKey).Delete(&dbmodel.MappingFieldGroups{})
	if err != nil {
		dbErr := database.Error(database.DELETE, database.MappingFieldGroups).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("failed to delete mapping_field_groups for tenant %d: %s: %v", tenantId, dbErr.Message, err)
		return dbErr
	}

	formattersKey := dbmodel.MappingFormatters{
		TenantId:  tenantId,
		ScopeType: api.MAPPINGSCOPE_TENANT,
	}
	rowsAffected, err = db.Where(&formattersKey).Delete(&dbmodel.MappingFormatters{})
	if err != nil {
		dbErr := database.Error(database.DELETE, database.MappingFormatters).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("failed to delete mapping_formatters for tenant %d: %s: %v", tenantId, dbErr.Message, err)
		return dbErr
	}

	rowsAffected, err = db.Where(&dbmodel.MappingLocking{TenantId: tenantId}).Delete(&dbmodel.MappingLocking{})
	if err != nil {
		dbErr := database.Error(database.DELETE, database.MappingLocking).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("failed to delete mapping_locking for tenant %d: %s: %v", tenantId, dbErr.Message, err)
		return dbErr
	}

	return nil
}
