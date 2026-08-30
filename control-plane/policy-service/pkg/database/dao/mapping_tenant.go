package dao

import (
	"context"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"gorm.io/gorm"
)

// DeleteMappingsOfTenant removes every tenant-scoped mapping row owned by the
// given tenant: mapping_selectors, mapping_field_groups, mapping_formatters
// (only rows with scope=tenant), and any leftover row in mapping_locking. system
// rows are never touched. used by the tenant deletion handler.
func DeleteMappingsOfTenant(ctx context.Context, db *gorm.DB, tenantId int32) *api.Error {
	selectorsKey := dbmodel.MappingSelectors{
		TenantId:  tenantId,
		ScopeType: api.MAPPINGSCOPE_TENANT,
	}
	res := db.Where(&selectorsKey).Delete(&dbmodel.MappingSelectors{})
	if res.Error != nil {
		dbErr := database.Error(database.DELETE, database.MappingSelectors).WithDetailsFromDb(res)
		logging.Errorf("failed to delete mapping_selectors for tenant %d: %s: %v", tenantId, dbErr.Message, res.Error)
		return dbErr
	}

	fieldGroupsKey := dbmodel.MappingFieldGroups{
		TenantId:  tenantId,
		ScopeType: api.MAPPINGSCOPE_TENANT,
	}
	res = db.Where(&fieldGroupsKey).Delete(&dbmodel.MappingFieldGroups{})
	if res.Error != nil {
		dbErr := database.Error(database.DELETE, database.MappingFieldGroups).WithDetailsFromDb(res)
		logging.Errorf("failed to delete mapping_field_groups for tenant %d: %s: %v", tenantId, dbErr.Message, res.Error)
		return dbErr
	}

	formattersKey := dbmodel.MappingFormatters{
		TenantId:  tenantId,
		ScopeType: api.MAPPINGSCOPE_TENANT,
	}
	res = db.Where(&formattersKey).Delete(&dbmodel.MappingFormatters{})
	if res.Error != nil {
		dbErr := database.Error(database.DELETE, database.MappingFormatters).WithDetailsFromDb(res)
		logging.Errorf("failed to delete mapping_formatters for tenant %d: %s: %v", tenantId, dbErr.Message, res.Error)
		return dbErr
	}

	res = db.Where(&dbmodel.MappingLocking{TenantId: tenantId}).Delete(&dbmodel.MappingLocking{})
	if res.Error != nil {
		dbErr := database.Error(database.DELETE, database.MappingLocking).WithDetailsFromDb(res)
		logging.Errorf("failed to delete mapping_locking for tenant %d: %s: %v", tenantId, dbErr.Message, res.Error)
		return dbErr
	}

	return nil
}
