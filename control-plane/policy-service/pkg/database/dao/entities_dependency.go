/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"gorm.io/gorm/clause"
)

// UpsertEntityDependencies bulk-upserts dependency edges for the caller's
// tenant. callers are expected to populate CacheId from the cache row returned
// by UpsertEntityCache so the FK to entities_cache(id) is satisfied.
//
// db must be a *gorm.DB scoped to the right context. pass the same transaction
// used for the preceding UpsertEntityCache call so both operations are atomic.
//
// rows that collide on the (component_type, component_id, cache_id) primary
// key are silently ignored - dependency rows have no payload beyond their key,
// so re-asserting an existing edge is a no-op.
func UpsertEntityDependencies(ctx context.Context, db orm.DB, rows []dbmodel.EntitiesDependency) *api.Error {
	if len(rows) == 0 {
		return nil
	}

	tenantId := auth.TenantIdOf(ctx)
	for i := range rows {
		rows[i].TenantId = tenantId
	}

	if _, err := db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&rows); err != nil {
		dbErr := database.Error(database.UPSERT, database.EntitiesDependency).WithDetails(err)
		logging.Errorf("failed to upsert %d entity dependency rows: %s: %v", len(rows), dbErr.Message, err)
		return dbErr
	}

	return nil
}

// MarkPoliciesPendingByComponent sets every policy whose cached breakdown
// depends on the given component to PendingUpdate, in all tenants, and returns
// how many policies were marked. it is meant for system seeding, which runs as
// the system tenant but changes components every tenant depends on. the caches
// and dependency rows are left in place.
func MarkPoliciesPendingByComponent(_ context.Context, db orm.DB, componentType dbmodel.ComponentType, componentId string) (int, *api.Error) {
	var policyIds []string
	if err := db.Table(dbtables.EntitiesCacheTable).
		Joins(fmt.Sprintf("JOIN %s ON %s.%s = %s.%s",
			dbtables.EntitiesDependencyTable,
			dbtables.EntitiesDependencyTable, dbtables.EntitiesDependencyColumnCacheId,
			dbtables.EntitiesCacheTable, dbtables.EntitiesCacheColumnId)).
		Where(fmt.Sprintf("%s.%s = ?", dbtables.EntitiesCacheTable, dbtables.EntitiesCacheColumnEntityType), dbmodel.EntityTypeBreakdown).
		Where(fmt.Sprintf("%s.%s = ?", dbtables.EntitiesDependencyTable, dbtables.EntitiesDependencyColumnComponentType), componentType).
		Where(fmt.Sprintf("%s.%s = ?", dbtables.EntitiesDependencyTable, dbtables.EntitiesDependencyColumnComponentId), componentId).
		Distinct().
		Pluck(fmt.Sprintf("%s.%s", dbtables.EntitiesCacheTable, dbtables.EntitiesCacheColumnEntityId), &policyIds); err != nil {
		dbErr := database.Error(database.LIST, database.EntitiesCache).WithDetails(err)
		logging.Errorf("failed to find policies depending on component %s/%s: %s: %v", componentType, componentId, dbErr.Message, err)
		return 0, dbErr
	}

	if len(policyIds) == 0 {
		return 0, nil
	}

	if _, err := db.Table(dbtables.PoliciesTable).
		Where(fmt.Sprintf("%s.%s IN ?", dbtables.PoliciesTable, dbtables.PoliciesColumnId), policyIds).
		UpdateColumns(map[string]interface{}{
			dbtables.PoliciesColumnValidationStatus: string(api.POLICYPHASE_PENDING_UPDATE),
			dbtables.PoliciesColumnValidationIssues: nil,
		}); err != nil {
		dbErr := database.Error(database.UPDATE, database.Policies).WithDetails(err)
		logging.Errorf("failed to mark %d policies as %s: %s: %v", len(policyIds), api.POLICYPHASE_PENDING_UPDATE, dbErr.Message, err)
		return 0, dbErr
	}

	return len(policyIds), nil
}

// OnComponentDeleted invalidates every entity whose cache depends on the given
// component, scoped to the caller's tenant. it executes four steps on db:
//
//  1. collect all cache_ids that have at least one dep on this component.
//  2. get the entity_ids (policy IDs) from those caches for status update.
//  3. update those policies' status to PendingUpdate with no issues.
//  4. delete every dep row for those caches (not just the ones matching the
//     component) so no orphan dep rows remain after the cache is gone.
//  5. delete the cache rows.
//
// db must be a *gorm.DB scoped to the right context. when the caller already
// owns a transaction, pass it here so the invalidation is atomic with the
// main change and happens before the caller's commit. when there is no
// surrounding transaction, pass runaidb.DB().WithContext(ctx).
func OnComponentDeleted(ctx context.Context, db orm.DB, componentType dbmodel.ComponentType, componentId string) *api.Error {
	tenantId := auth.TenantIdOf(ctx)

	// step 1: find affected cache ids (scoped to the tenant)
	var cacheIds []string

	queryIds := dbmodel.EntitiesDependency{
		TenantId:      tenantId,
		ComponentType: componentType,
		ComponentId:   componentId,
	}

	if err := db.Model(&queryIds).Where(&queryIds).Distinct().
		Pluck(dbtables.EntitiesDependencyColumnCacheId, &cacheIds); err != nil {
		dbErr := database.Error(database.DELETE, database.EntitiesDependency).WithDetails(err)
		logging.Errorf("failed to find caches for component %s/%s: %s: %v", componentType, componentId, dbErr.Message, err)
		return dbErr
	}

	if len(cacheIds) == 0 {
		return nil
	}

	// step 2: get policy IDs (entity_ids) from the cache rows that will be deleted
	var policyIds []string
	if err := db.Model(&dbmodel.EntitiesCache{}).
		Where(&dbmodel.EntitiesCache{TenantId: tenantId, EntityType: dbmodel.EntityTypeBreakdown}).
		Where(fmt.Sprintf("%s IN ?", dbtables.EntitiesCacheColumnId), cacheIds).
		Pluck(dbtables.EntitiesCacheColumnEntityId, &policyIds); err != nil {
		dbErr := database.Error(database.LIST, database.EntitiesCache).WithDetails(err)
		logging.Errorf("failed to get policy ids from cache for component %s/%s: %s: %v", componentType, componentId, dbErr.Message, err)
		return dbErr
	}

	// step 3: update policies' status to PendingUpdate
	if len(policyIds) > 0 {
		pendingStatus := &api.PolicyStatus{
			Phase:  api.POLICYPHASE_PENDING_UPDATE,
			Issues: []string{},
		}
		if apiErr := UpdatePoliciesStatus(ctx, db, policyIds, pendingStatus); apiErr != nil {
			return apiErr
		}
	}

	// step 4: delete ALL dep rowsAffected for those caches
	rowsAffected, err := db.Where(&dbmodel.EntitiesDependency{TenantId: tenantId}).
		Where(fmt.Sprintf("%s IN ?", dbtables.EntitiesDependencyColumnCacheId), cacheIds).
		Delete(&dbmodel.EntitiesDependency{})
	if err != nil {
		dbErr := database.Error(database.DELETE, database.EntitiesDependency).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("failed to delete dep rows for component %s/%s: %s: %v", componentType, componentId, dbErr.Message, err)
		return dbErr
	}

	// step 5: delete the cache rows
	rowsAffected, err = db.Where(&dbmodel.EntitiesCache{TenantId: tenantId}).
		Where(fmt.Sprintf("%s IN ?", dbtables.EntitiesCacheColumnId), cacheIds).
		Delete(&dbmodel.EntitiesCache{})
	if err != nil {
		dbErr := database.Error(database.DELETE, database.EntitiesCache).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("failed to delete cache rows for component %s/%s: %s: %v", componentType, componentId, dbErr.Message, err)
		return dbErr
	}

	return nil
}
