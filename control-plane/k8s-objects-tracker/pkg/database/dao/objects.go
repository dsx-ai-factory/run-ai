/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	authorizationClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"gorm.io/gorm/clause"
)

// ListObjects query the db for objects
func ListObjects(ctx context.Context, filter *models.ObjectFilters) ([]*models.Object, *api.Error) {
	record := models.Object{}
	dbErr := api.NewError(http.StatusServiceUnavailable, record.ErrorMessage("list"), nil)

	db := runaidb.DB().WithContext(ctx)

	objects := make([]*models.Object, 0)
	if err := db.Model(record).Where(filter).Find(&objects); err != nil {
		return nil, dbErr.WithDetails(err)
	}

	return objects, nil
}

func ListStorageClassWithOptionalConfig(ctx context.Context, filter *models.ObjectFilters) ([]*models.StorageClassWithOptionalConfig, *api.Error) {
	record := models.Object{}
	dbErr := api.NewError(http.StatusServiceUnavailable, record.ErrorMessage("list"), nil)

	db := runaidb.DB().WithContext(ctx)

	filter.Type = models.ObjectTypeStorageClass

	objects := make([]*models.StorageClassWithOptionalConfig, 0)

	joinCmd := fmt.Sprintf("LEFT JOIN %s ON ", models.StorageClassConfigurationsTable)

	joinCmd += fmt.Sprintf("%s.%s = %s.%s AND ",
		models.StorageClassConfigurationsTable, models.StorageClassConfigurationTenantIdColumn, models.ObjectsTable, models.ObjectsTenantIdColumn)
	joinCmd += fmt.Sprintf("%s.%s = %s.%s AND ",
		models.StorageClassConfigurationsTable, models.StorageClassConfigurationScopeTypeColumn, models.ObjectsTable, models.ObjectScopeTypeColumn)
	joinCmd += fmt.Sprintf("%s.%s = %s.%s::text AND ",
		models.StorageClassConfigurationsTable, models.StorageClassConfigurationScopeIdColumn, models.ObjectsTable, models.ObjectsClusterIdColumn)
	joinCmd += fmt.Sprintf("%s.%s = %s.%s",
		models.StorageClassConfigurationsTable, models.StorageClassConfigurationNameColumn, models.ObjectsTable, models.ObjectsNameColumn)

	if err := db.Table(models.ObjectsTable).Select("*").Where(filter).Joins(joinCmd).Find(&objects); err != nil {
		return nil, dbErr.WithDetails(err)
	}

	return objects, nil
}

// UpsertObject upserts a single object and populates its ID with the stable
// row id. The second return value reports whether the row was newly inserted
// vs. an update of an existing one - callers that track issues or publish
// events use this to decide whether to emit "newly created" semantics.
//
// Prefer this over UpsertObjects when the caller needs the row id (for
// example to tie resource_issues rows to the object). UpsertObjects is a
// bulk-write path that does not expose the generated id and is used by the
// DeviceClass and StorageClass flows that have no per-row downstream work.
func UpsertObject(ctx context.Context, obj *models.Object) (bool, *api.Error) {
	dbErr := api.NewError(http.StatusServiceUnavailable, obj.ErrorMessage("upsert"), nil)
	db := runaidb.DB().WithContext(ctx)

	existing := &models.Object{ObjectFilters: obj.ObjectFilters}
	err := db.Model(existing).Where(existing).First(existing)
	switch {
	case err == nil:
		obj.ID = existing.ID
		if _, err := db.Model(obj).Save(obj); err != nil {
			return false, dbErr.WithDetails(err)
		}
		return false, nil
	case errors.Is(err, orm.ErrRecordNotFound):
		if _, err := db.Model(obj).Create(obj); err != nil {
			return false, dbErr.WithDetails(err)
		}
		return true, nil
	default:
		return false, dbErr.WithDetails(err)
	}
}

// UpsertObjects creates or updates multiple records
func UpsertObjects(ctx context.Context, filter *models.ObjectFilters, objects []*models.Object) *api.Error {
	if len(objects) == 0 {
		return nil
	}
	filterRecord := models.NewObject(filter)
	dbErr := api.NewError(http.StatusServiceUnavailable, filterRecord.ErrorMessage("upsert"), nil)
	db := runaidb.DB().WithContext(ctx)

	if _, err := db.Clauses(clause.OnConflict{UpdateAll: true, OnConstraint: models.ObjectsExtendedUniqueConstraint}).
		Save(objects); err != nil {
		return dbErr.WithDetails(err)
	}

	return nil
}

// Delete deletes multiple/single Object records by defined filters
func Delete(ctx context.Context, filters *models.ObjectFilters) *api.Error {
	record := models.NewObject(filters)
	dbErr := api.NewError(http.StatusServiceUnavailable, record.ErrorMessage("delete"), nil)

	db := runaidb.DB().WithContext(ctx)

	if _, err := db.Where(filters).Delete(record); err != nil {
		return dbErr.WithDetails(err)
	}

	return nil
}

// DeleteProjectObjects deletes all objects that of a project
func DeleteProjectObjects(ctx context.Context, tenantId int32, projectId string) *api.Error {
	return deleteEntitiesByScope(ctx, authorizationClient.SCOPETYPE_PROJECT, tenantId, projectId, &models.Object{ObjectFilters: models.ObjectFilters{ScopeType: authorizationClient.SCOPETYPE_PROJECT}})
}

// DeleteClusterObjects deletes all objects within a cluster
func DeleteClusterObjects(ctx context.Context, tenantId int32, clusterId string) *api.Error {
	return deleteClusterEntities(ctx, tenantId, clusterId, &models.Object{})
}

// DeleteTenantObjects deletes all objects within a tenant
func DeleteTenantObjects(ctx context.Context, tenantId int32) *api.Error {
	return deleteTenantEntities(ctx, tenantId, &models.Object{})
}

// DeleteObjectsNotInCluster deletes multiple/single Object records by defined filter.
// It also filters out of the deletion, records which are representing provided 'objects' list.
//
// We pass an empty-model receiver (not `objects`) to GORM's Delete so GORM does
// not splice primary keys from the kept slice into the WHERE clause. Callers
// that populate obj.ID on upsert (e.g. the per-row UpsertObject path used by
// RCT / RC) would otherwise end up with `... AND id IN (kept.ids) AND name NOT
// IN (kept.names)` - an impossible intersection that silently prunes nothing.
func DeleteObjectsNotInCluster(ctx context.Context, filter *models.ObjectFilters, objects []*models.Object) *api.Error {
	filterRecord := models.NewObject(filter)
	dbErr := api.NewError(http.StatusServiceUnavailable, filterRecord.ErrorMessage("delete"), nil)

	db := runaidb.DB().WithContext(ctx)

	tx := db.Model(filterRecord).Where(filter)
	if len(objects) > 0 {
		tx = tx.Not(fmt.Sprintf("%s in ?", models.ObjectsNameColumn), extractNames(objects))
	}

	if _, err := tx.Delete(&models.Object{}); err != nil {
		return dbErr.WithDetails(err)
	}
	return nil
}

func extractNames(objects []*models.Object) []string {
	result := make([]string, len(objects))
	for i, object := range objects {
		result[i] = *object.Name
	}
	return result
}
