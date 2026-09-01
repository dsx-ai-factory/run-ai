/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-application-packages/dbmodels"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
	"gorm.io/gorm/clause"
)

// getCachedEntity returns the cache row for (entityType, entityId) scoped to
// the caller's tenant, or nil if no cache exists. a cache miss is not an error.
func getCachedEntity(ctx context.Context, entityType dbmodel.EntityType, entityId string) (*dbmodel.EntitiesCache, *api.Error) {
	key := dbmodel.EntitiesCache{
		TenantId:   auth.TenantIdOf(ctx),
		EntityType: entityType,
		EntityId:   entityId,
	}

	var row dbmodel.EntitiesCache

	err := runaidb.DB().WithContext(ctx).Where(&key).First(&row)
	if err != nil {
		if errors.Is(err, orm.ErrRecordNotFound) {
			return nil, nil
		}
		dbErr := database.Error(database.GET, database.EntitiesCache).WithDetails(err)
		logging.Errorf("failed to get entity cache for %s/%s: %s: %v", entityType, entityId, dbErr.Message, err)
		return nil, dbErr
	}

	return &row, nil
}

// UpsertEntityCache writes the cached info for (tenantId, entityType, entityId),
// creating the cache row if it does not exist and updating cached_info otherwise.
// on return, row.ID holds the actual id of the row in the database (the existing
// id is preserved on update via RETURNING; a fresh UUID is assigned on insert).
//
// db must be a *gorm.DB scoped to the right context. pass a transaction when
// the upsert must be atomic with a subsequent UpsertEntityDependencies call.
func UpsertEntityCache(ctx context.Context, db orm.DB, row *dbmodel.EntitiesCache) *api.Error {
	if row.CachedInfo == nil {
		row.CachedInfo = dbmodels.JsonBMap{}
	}
	if row.ID == uuid.Nil {
		row.ID = uuid.New()
	}
	row.TenantId = auth.TenantIdOf(ctx)

	_, err := db.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{
				{Name: dbtables.EntitiesCacheColumnTenantId},
				{Name: dbtables.EntitiesCacheColumnEntityType},
				{Name: dbtables.EntitiesCacheColumnEntityId},
			},
			DoUpdates: clause.AssignmentColumns([]string{dbtables.EntitiesCacheColumnCachedInfo}),
		},
		clause.Returning{Columns: []clause.Column{{Name: dbtables.EntitiesCacheColumnId}}},
	).
		Create(row)
	if err != nil {
		dbErr := database.Error(database.UPSERT, database.EntitiesCache).WithDetails(err)
		logging.Errorf("failed to upsert entity cache for %s/%s: %s: %v", row.EntityType, row.EntityId, dbErr.Message, err)
		return dbErr
	}

	return nil
}

// OnEntityDeleted removes the cache and all dependency rows for a given entity
// belonging to the caller's tenant. runs in a single transaction. returns nil
// (no-op) when no cache row exists for the entity.
// OnEntityDeleted removes the cache and all dependency rows for the given
// entity. db must be a *gorm.DB scoped to the right context — pass the
// caller's transaction when one exists so the invalidation is atomic with the
// entity deletion itself.
func OnEntityDeleted(ctx context.Context, db orm.DB, entityType dbmodel.EntityType, entityId string) *api.Error {
	tenantId := auth.TenantIdOf(ctx)

	cacheKey := dbmodel.EntitiesCache{
		TenantId:   tenantId,
		EntityType: entityType,
		EntityId:   entityId,
	}

	var cache dbmodel.EntitiesCache

	if err := db.Where(&cacheKey).First(&cache); err != nil {
		if errors.Is(err, orm.ErrRecordNotFound) {
			return nil // no cache to invalidate — no-op
		}
		dbErr := database.Error(database.DELETE, database.EntitiesCache).WithDetails(err)
		logging.Errorf("failed to find cache for entity %s/%s: %s: %v", entityType, entityId, dbErr.Message, err)
		return dbErr
	}

	rowsAffected, err := db.Model(&dbmodel.EntitiesDependency{}).
		Where(&dbmodel.EntitiesDependency{CacheId: cache.ID, TenantId: tenantId}).
		Delete(&dbmodel.EntitiesDependency{})
	if err != nil {
		dbErr := database.Error(database.DELETE, database.EntitiesDependency).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("failed to delete deps for entity %s/%s: %s: %v", entityType, entityId, dbErr.Message, err)
		return dbErr
	}

	rowsAffected, err = db.Model(&dbmodel.EntitiesCache{}).
		Where(&dbmodel.EntitiesCache{ID: cache.ID, TenantId: tenantId}).
		Delete(&dbmodel.EntitiesCache{})
	if err != nil {
		dbErr := database.Error(database.DELETE, database.EntitiesCache).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("failed to delete cache for entity %s/%s: %s: %v", entityType, entityId, dbErr.Message, err)
		return dbErr
	}

	return nil
}
