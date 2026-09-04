/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"gorm.io/gorm/schema"
)

const deleteBatchSize = 100

// AssetIDer is an interface to a model that has asset id as attribute
type AssetIDer interface {
	// AssetIdColumn returns the name of the asset id column
	AssetIdColumn() string
}

// TableAssetIDer unified Tabler and AssetIDer
type TableAssetIDer interface {
	AssetIDer
	schema.Tabler
}

// DeleteAssetsByScope deletes all the assets under a target scope
func DeleteAssetsByScope(ctx context.Context, scopeType api.ScopeType, scopeId string) error {
	db := runaidb.DB().WithContext(ctx)
	var err error

	// splitting the assets into 3 layers to avoid foreign key constraints
	// Top layer depends on lower level
	// Note that workloads are deleted before invoking this function
	layerThree := []api.AssetKind{
		api.ASSETKIND_MODEL,
		api.ASSETKIND_WORKLOAD_TEMPLATE,
	}
	layerTwo := []api.AssetKind{
		api.ASSETKIND_PVC,
		api.ASSETKIND_CONFIG_MAP,
		api.ASSETKIND_NFS,
		api.ASSETKIND_S3,
		api.ASSETKIND_GIT,
		api.ASSETKIND_SECRET_VOLUME,
		api.ASSETKIND_HOST_PATH,
		api.ASSETKIND_REGISTRY,
	}
	layerOne := []api.AssetKind{
		api.ASSETKIND_COMPUTE,
		api.ASSETKIND_ENVIRONMENT,
		api.ASSETKIND_ACCESS_KEY,
		api.ASSETKIND_PASSWORD,
		api.ASSETKIND_GENERIC_SECRET,
		api.ASSETKIND_DOCKER_REGISTRY,
		api.ASSETKIND_NGC_API_KEY,
	}

	for _, layer := range [][]api.AssetKind{layerThree, layerTwo, layerOne} {
		for _, kind := range layer {
			err = deleteAssetsByScope(ctx, db, kind, scopeType, scopeId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// deleteAssetsByScope deletes all assets of a given kind that are in a given scope
func deleteAssetsByScope(ctx context.Context, db orm.DB, kind api.AssetKind, scopeType api.ScopeType, scopeId string) error {
	return db.Transaction(func(tx orm.DB) error {
		scopeIdColumn, err := dbtables.ScopeIdColumnName(scopeType)
		if err != nil {
			return err
		}
		tenantId := auth.TenantIdOf(ctx)

		for { // Loop until no more assets to delete
			var idsToDelete []string

			// Fetch the asset IDs
			if err := tx.Model(&dbmodel.AssetMetaRec{}).
				Select(dbtables.AssetMetaColumnId). // Ensure only IDs are fetched
				Where(fmt.Sprintf("%s = ?", dbtables.AssetMetaColumnTenantId), tenantId).
				Where(fmt.Sprintf("%s = ?", scopeIdColumn), scopeId).
				Where(fmt.Sprintf("%s = ?", dbtables.AssetMetaColumnKind), kind).
				Order(fmt.Sprintf("%s ASC", dbtables.AssetMetaColumnId)).
				Limit(deleteBatchSize). // Limit the number of records per query to avoid overloading the DB
				Pluck(dbtables.AssetMetaColumnId, &idsToDelete); err != nil {
				return fmt.Errorf("failed to fetch IDs of assets to delete: %w", err)
			}

			// Stop if no more assets to delete
			if len(idsToDelete) == 0 {
				break
			}

			// Delete assets from bookkeeping table
			if err := deleteAssetsFromBookkeepingTable(tx, idsToDelete); err != nil {
				return err
			}

			// Delete from asset-specific table
			assetModel := dbmodel.GetAssetModel(kind)
			if _, err := tx.Where(fmt.Sprintf("%s IN (?)", dbtables.AssetIdColumnName), idsToDelete).
				Delete(assetModel); err != nil {
				return fmt.Errorf("failed to delete assets from %s: %w", kind, err)
			}

			// Delete from metadata table
			if _, err := tx.Where(fmt.Sprintf("%s IN (?)", dbtables.AssetMetaColumnId), idsToDelete).
				Unscoped().Delete(&dbmodel.AssetMetaRec{}); err != nil {
				return fmt.Errorf("failed to delete assets from metadata: %w", err)
			}
		}
		return nil
	})
}

func deleteAssetsFromBookkeepingTable(tx orm.DB, paginatedIDs []string) error {
	// list of bookkeeping tables that need to be cleaned up
	// workloads_assets and workload template are deleted as part of the main flow
	tables := []TableAssetIDer{
		dbmodel.WorkloadSupportedTypesRec{},
		dbmodel.LastUsedByWorkloadRecord{},
		dbmodel.PolicyComplianceRecord{},
		dbmodel.TemplateContentRec{},
	}

	for _, table := range tables {
		if _, err := tx.Where(fmt.Sprintf("%s IN (?)", table.AssetIdColumn()), paginatedIDs).Delete(&table); err != nil {
			return fmt.Errorf("failed to delete assets from bookkeeping table %s: %w", table.TableName(), err)
		}
	}
	return nil
}

// DeleteWorkloadsByScope deletes all workloads (and related workload assets) of a given scope
func DeleteWorkloadsByScope(ctx context.Context, scopeType api.ScopeType, scopeId string) error {
	db := runaidb.DB().WithContext(ctx)
	return db.Transaction(func(tx orm.DB) error {
		var idsToDelete []string // ids of the workloads to delete
		tenantId := auth.TenantIdOf(ctx)

		scopeIdColumn, err := dbtables.ScopeIdColumnName(scopeType)
		if err != nil {
			return err
		}

		for { // Loop until no more workloads to delete
			// Get the ids of the workloads that are in the given scope
			if err := tx.Model(&dbmodel.WorkloadRec{}).
				Where(fmt.Sprintf("%s = ?", dbtables.WorkloadsColumnTenantId), tenantId).
				Where(fmt.Sprintf("%s = ?", scopeIdColumn), scopeId).
				Order(fmt.Sprintf("%s ASC", dbtables.WorkloadsColumnId)).
				Limit(deleteBatchSize). // Limit the number of records per query to avoid overloading the DB
				Pluck(dbtables.WorkloadsColumnId, &idsToDelete); err != nil {
				return fmt.Errorf("failed to fetch IDs of assets to delete: %w", err)
			}

			// Nothing to delete
			if len(idsToDelete) == 0 {
				break
			}

			// Delete all workload's assets first
			if _, err := tx.Where(fmt.Sprintf("%s IN (?)", dbtables.WorkloadsAssetsColumnWorkloadId), idsToDelete).
				Delete(&dbmodel.WorkloadsAssetsRec{}); err != nil {
				return fmt.Errorf("failed to delete assets from metadata: %w", err)
			}

			// Delete the workloads themselves
			if _, err := tx.Where(fmt.Sprintf("%s IN (?)", dbtables.WorkloadsColumnId), idsToDelete).
				Delete(&dbmodel.WorkloadRec{}); err != nil {
				return fmt.Errorf("failed to delete workloads from metadata: %w", err)
			}
		}
		return nil
	})
}
