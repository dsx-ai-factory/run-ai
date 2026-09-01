/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-application-packages/dbmodels"
	"github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
	"gorm.io/gorm"
)

// WipeAllEntityCaches removes every row from entities_dependency and
// entities_cache (in that order, to satisfy the FK). it is used by the system
// seeding path on startup: when any system-scoped component (mapping, field
// group, formatter) version bumps, every tenant's cached breakdown could be
// affected and the cheapest correct response is to drop everything.
func WipeAllEntityCaches(ctx context.Context) *api.Error {
	db := runaidb.DB().WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true})

	if _, err := db.Delete(&dbmodel.EntitiesDependency{}); err != nil {
		dbErr := database.Error(database.DELETE, database.EntitiesDependency).WithDetails(err)
		logging.Errorf("failed to wipe entities_dependency: %s: %v", dbErr.Message, err)
		return dbErr
	}

	if _, err := db.Delete(&dbmodel.EntitiesCache{}); err != nil {
		dbErr := database.Error(database.DELETE, database.EntitiesCache).WithDetails(err)
		logging.Errorf("failed to wipe entities_cache: %s: %v", dbErr.Message, err)
		return dbErr
	}

	return nil
}

// DeleteEntityCachesOfTenant removes every dependency and cache row owned by
// the given tenant (in that order, to satisfy the FK). it is used by the
// tenant/cluster/project deletion handlers, which prefer to drop all of the
// tenant's caches and let them be recomputed on next read.
func DeleteEntityCachesOfTenant(ctx context.Context, db orm.DB, tenantId int32) *api.Error {
	rowsAffected, err := db.Where(&dbmodel.EntitiesDependency{TenantId: tenantId}).Delete(&dbmodel.EntitiesDependency{})
	if err != nil {
		dbErr := database.Error(database.DELETE, database.EntitiesDependency).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("failed to delete entities_dependency for tenant %d: %s: %v", tenantId, dbErr.Message, err)
		return dbErr
	}

	rowsAffected, err = db.Where(&dbmodel.EntitiesCache{TenantId: tenantId}).Delete(&dbmodel.EntitiesCache{})
	if err != nil {
		dbErr := database.Error(database.DELETE, database.EntitiesCache).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("failed to delete entities_cache for tenant %d: %s: %v", tenantId, dbErr.Message, err)
		return dbErr
	}

	return nil
}

// GetCachedBreakdown returns the cached BreakdownBreakdown for the given entity
// id, or nil if no cache exists. entityId is typically the policy id string.
//
// if the row exists but its cached_info JSON cannot be deserialized into
// BreakdownBreakdown (corrupt or schema-drifted data), the function logs the
// error and reports a cache miss (return nil, nil) so the caller falls through
// to recompute the breakdown rather than returning a zero-value result.
func GetCachedBreakdown(ctx context.Context, entityId string) (*api.BreakdownBreakdown, *api.Error) {
	row, apiErr := getCachedEntity(ctx, dbmodel.EntityTypeBreakdown, entityId)
	if apiErr != nil || row == nil {
		return nil, apiErr
	}

	// JsonBMap.Unmarshal swallows errors internally — go via json.Marshal so
	// we can detect corruption and treat it as a cache miss.
	bytes, err := json.Marshal(row.CachedInfo)
	if err != nil {
		logging.Errorf("failed to marshal cached breakdown for entity %s: %v - treating as cache miss", entityId, err)
		return nil, nil
	}

	var breakdown api.BreakdownBreakdown
	if err := json.Unmarshal(bytes, &breakdown); err != nil {
		logging.Errorf("failed to unmarshal cached breakdown for entity %s: %v - treating as cache miss", entityId, err)
		return nil, nil
	}

	return &breakdown, nil
}

// CacheBreakdown persists a BreakdownBreakdown and its dependency edges.
// The caller owns the transaction: pass the same *gorm.DB used for the rest
// of the surrounding operation so the cache row, the dependency edges, and
// any sibling writes commit (or roll back) together. The two writes done
// here would otherwise be auto-committed independently and could leave the
// cache row without its deps if the second write failed.
//
// the correct flow is:
//  1. upsert the entities_cache row. the ON CONFLICT ... RETURNING id path
//     returns the *existing* row's id when the row already exists (e.g. from a
//     concurrent call). using a pre-generated id and stamping deps with it
//     would leave them pointing at a non-existing id when the conflict fires.
//  2. stamp every dependency with the actual id returned by step 1.
//  3. upsert the dependency edges.
func CacheBreakdown(ctx context.Context, tx orm.DB, entityId string, breakdown *api.BreakdownBreakdown,
	dependencies []dbmodel.EntitiesDependency,
) (uuid.UUID, *api.Error) {
	cachedInfo, err := gojsonq.LoadJson(breakdown)
	if err != nil {
		return uuid.Nil, api.NewError(http.StatusInternalServerError, "failed to serialize breakdown for cache", err)
	}

	tenantId := auth.TenantIdOf(ctx)

	row := dbmodel.EntitiesCache{
		EntityType: dbmodel.EntityTypeBreakdown,
		EntityId:   entityId,
		CachedInfo: dbmodels.JsonBMap(cachedInfo),
	}

	// step 1: upsert cache — RETURNING id gives the actual DB id
	if apiErr := UpsertEntityCache(ctx, tx, &row); apiErr != nil {
		return uuid.Nil, apiErr
	}

	actualCacheId := row.ID

	// step 2: stamp deps with the real cache id returned above
	for i := range dependencies {
		dependencies[i].CacheId = actualCacheId
		dependencies[i].TenantId = tenantId
	}

	// step 3: upsert deps
	if apiErr := UpsertEntityDependencies(ctx, tx, dependencies); apiErr != nil {
		return uuid.Nil, apiErr
	}

	return actualCacheId, nil
}

// DeleteBreakdownCachesOfSubtree removes breakdown cache rows (and their
// dependency edges) for the given policy ids, which are expected to be the
// descendants returned by SoftDeleteEffectivePoliciesOfSubtree. used by the
// delete/update/create-policy flows on a parent scope: a descendant's
// cached breakdown is derived from its own effective rules, and those
// effective rules change the moment the ancestor is touched — leaving the
// cache row in place would serve stale breakdowns on the next read.
//
// the caller's transaction is reused (db is *gorm.DB) so the cache cleanup
// commits or rolls back atomically with the surrounding policy change.
func DeleteBreakdownCachesOfSubtree(ctx context.Context, db orm.DB, policyIds []string) *api.Error {
	if len(policyIds) == 0 {
		return nil
	}

	tenantId := auth.TenantIdOf(ctx)

	// step 1: resolve the policy ids to cache ids so we can also clear the
	// matching dependency rows (cache_id is the FK target). policies that
	// have no cached breakdown simply contribute no cache_ids — no-op.
	var cacheIds []uuid.UUID
	if err := db.Model(&dbmodel.EntitiesCache{}).
		Where(&dbmodel.EntitiesCache{TenantId: tenantId, EntityType: dbmodel.EntityTypeBreakdown}).
		Where(fmt.Sprintf("%s IN ?", dbtables.EntitiesCacheColumnEntityId), policyIds).
		Pluck(dbtables.EntitiesCacheColumnId, &cacheIds); err != nil {
		dbErr := database.Error(database.LIST, database.EntitiesCache).WithDetails(err)
		logging.Errorf("failed to list subtree cache ids: %s: %v", dbErr.Message, err)
		return dbErr
	}

	if len(cacheIds) == 0 {
		return nil
	}

	// step 2: drop dependency edges first to honor the FK to entities_cache.
	rowsAffected, err := db.Where(&dbmodel.EntitiesDependency{TenantId: tenantId}).
		Where(fmt.Sprintf("%s IN ?", dbtables.EntitiesDependencyColumnCacheId), cacheIds).
		Delete(&dbmodel.EntitiesDependency{})
	if err != nil {
		dbErr := database.Error(database.DELETE, database.EntitiesDependency).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("failed to delete subtree dependency rows: %s: %v", dbErr.Message, err)
		return dbErr
	}

	// step 3: drop the cache rows themselves.
	rowsAffected, err = db.Where(&dbmodel.EntitiesCache{TenantId: tenantId}).
		Where(fmt.Sprintf("%s IN ?", dbtables.EntitiesCacheColumnId), cacheIds).
		Delete(&dbmodel.EntitiesCache{})
	if err != nil {
		dbErr := database.Error(database.DELETE, database.EntitiesCache).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("failed to delete subtree cache rows: %s: %v", dbErr.Message, err)
		return dbErr
	}

	return nil
}
