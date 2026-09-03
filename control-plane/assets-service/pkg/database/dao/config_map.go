/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/google/uuid"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
)

var ConfigMapsDBAccessor configMapsDBAccessor

// configMapsDBAccessor implements ConfigMapsDAOAccessor
type configMapsDBAccessor struct{}

// CreateConfigMapAsset insets a new ConfigMap assets into the ConfigMaps and assets metadata tables
func (cmdb *configMapsDBAccessor) CreateConfigMapAsset(ctx context.Context, request api.ConfigMapCreationRequest, nameValidationQuery *database.ParametrizedQuery) (*api.ConfigMapAsset, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, err := createConfigMapAssetInTransaction(ctx, tx, request, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_CREATIONS, api.ASSETKIND_CONFIG_MAP).WithDetails(err)
	}
	committedTx = true

	return result, nil
}

// DeleteConfigMapAsset deletes a ConfigMap asset
func (cmdb *configMapsDBAccessor) DeleteConfigMapAsset(ctx context.Context, assetMeta *dbmodel.AssetMeta, deletionType DeletionType) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	err := deleteConfigMapAssetInTransaction(ctx, tx, assetMeta.Id, deletionType)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return database.Error(database.COMMIT_DELETIONS, api.ASSETKIND_CONFIG_MAP).WithDetails(err)
	}
	committedTx = true

	NotifyAssetDeleted(ctx, assetMeta.Id, assetMeta.Kind, ScopeInfoOfDbMeta(assetMeta))

	return nil
}

// GetConfigMapAssetById returns a ConfigMap asset by it asset id
func (cmdb *configMapsDBAccessor) GetConfigMapAssetById(ctx context.Context, assetId string) (*api.ConfigMapAsset, *api.Error) {
	var combinedRec dbmodel.ConfigMapCombinedRec
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_CONFIG_MAP,
			Id:   assetId,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, nil, dbtables.TableNameAssetsConfigMap)
	if err != nil {
		return nil, err
	}

	scanRows, scanErr := result.Scan(&combinedRec)
	if scanErr != nil || scanRows == 0 {
		return nil, database.Error(database.GET, api.ASSETKIND_CONFIG_MAP).WithDetailsFromDbResult(scanRows, scanErr)
	}

	response := api.ConfigMapAsset{
		Meta: combinedRec.AssetMeta,
		Spec: combinedRec.ConfigMapAssetSpec,
	}
	return &response, nil
}

// ListConfigMapAssets returns a list of  ConfigMap assets filter by query params
func (cmdb *configMapsDBAccessor) ListConfigMapAssets(ctx context.Context, queryParams *query_types.QueryParams) (*api.ConfigMapListResponse, *api.Error) {
	combinedList := make([]dbmodel.ConfigMapCombinedRec, 0)
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_CONFIG_MAP,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, queryParams, dbtables.TableNameAssetsConfigMap)
	if err != nil {
		return nil, err
	}

	_, scanErr := result.
		Order(&query.AssetMeta.Id).
		Scan(&combinedList)
	if scanErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list ConfigMap assets", scanErr)
	}

	response := api.ConfigMapListResponse{
		Entries: make([]api.ConfigMapAsset, 0),
	}

	for _, row := range combinedList {
		response.Entries = append(response.Entries, api.ConfigMapAsset{
			Meta: row.AssetMeta,
			Spec: row.ConfigMapAssetSpec,
		})
	}

	return &response, nil
}

// UpdateConfigMapAsset updates a ConfigMap assets
func (cmdb *configMapsDBAccessor) UpdateConfigMapAsset(ctx context.Context, assetId string,
	request api.ConfigMapUpdateRequest, nameValidationQuery *database.ParametrizedQuery,
) (*api.ConfigMapAsset, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, assetMeta, err := updateConfigMapAssetInTransaction(ctx, tx, assetId, request, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	policyConflict := ValidateAssetVsPolicy(ctx, assetId, ScopeInfoOfAssetMeta(&result.Meta))
	if policyConflict != nil {
		return nil, api.NewError(int(policyConflict.Code), policyConflict.Message, nil)
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_CONFIG_MAP).WithDetails(err)
	}
	committedTx = true

	NotifyAssetModified(ctx, assetId, assetMeta.Kind, ScopeInfoOfDbMeta(assetMeta))

	return result, nil
}

// createConfigMapAssetInTransaction insets a new ConfigMap assets into the ConfigMaps and assets metadata tables under a managed transaction
func createConfigMapAssetInTransaction(ctx context.Context, tx orm.DB, request api.ConfigMapCreationRequest, nameValidationQuery *database.ParametrizedQuery) (*api.ConfigMapAsset, *api.Error) {
	id := uuid.New().String()

	assetMeta, err := createAssetMeta(ctx, tx, id, api.ASSETKIND_CONFIG_MAP, request.Meta, nameValidationQuery)
	if err != nil {
		if strings.Contains(err.Error(), alreadyExistsError) {
			return nil, database.ErrorWithCode(database.CREATE, api.ASSETKIND_CONFIG_MAP, http.StatusConflict).WithDetails(err)
		}
		return nil, database.Error(database.CREATE, api.ASSETKIND_CONFIG_MAP).WithDetails(err)
	}

	configMapAssetRec := dbmodel.ConfigMapAssetRec{
		AssetId: id,
		ConfigMapAssetSpec: api.ConfigMapAssetSpec{
			ConfigMap:   request.Spec.ConfigMap,
			MountPath:   request.Spec.MountPath,
			DefaultMode: request.Spec.DefaultMode,
			SubPath:     request.Spec.SubPath,
		},
	}

	_, err = tx.Create(configMapAssetRec)
	if err != nil {
		return nil, database.Error(database.CREATE, api.ASSETKIND_CONFIG_MAP).WithDetails(err)
	}

	return &api.ConfigMapAsset{
		Meta: *assetMeta,
		Spec: request.Spec,
	}, nil
}

// deleteConfigMapAssetInTransaction deletes a ConfigMap asset from both ConfigMaps and assets metadata tables
func deleteConfigMapAssetInTransaction(ctx context.Context, tx orm.DB, assetId string, deletionType DeletionType) *api.Error {
	recId := dbmodel.ConfigMapAssetRec{AssetId: assetId}

	rowsAffected, apiErr, dbErr := deleteAssetMeta(ctx, tx, assetId, deletionType)
	if apiErr != nil {
		return apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_CONFIG_MAP).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	if deletionType == SoftDelete {
		return nil
	}

	delRows, delErr := tx.Where(&recId).Delete(&recId)
	if delErr != nil || delRows == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_CONFIG_MAP).WithDetailsFromDbResult(delRows, delErr)
	}

	if policyErr := deletePolicyComplianceOfAsset(ctx, tx, assetId); policyErr != nil {
		return policyErr
	}

	if err := deleteWorkloadAssetsByAssetId(ctx, tx, assetId); err != nil {
		return err
	}

	return nil
}

// updateConfigMapAssetInTransaction updates a ConfigMap asset from both ConfigMaps and assets metadata tables
func updateConfigMapAssetInTransaction(ctx context.Context, tx orm.DB, assetId string, request api.ConfigMapUpdateRequest,
	nameValidationQuery *database.ParametrizedQuery,
) (*api.ConfigMapAsset, *dbmodel.AssetMeta, *api.Error) {
	assetMeta, rowsAffected, apiErr, dbErr := updateAssetMeta(ctx, tx, assetId, request.Meta, nameValidationQuery, nil)
	if apiErr != nil {
		return nil, nil, apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return nil, nil, database.Error(database.UPDATE, api.ASSETKIND_CONFIG_MAP).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	recId := dbmodel.ConfigMapAssetRec{AssetId: assetId}
	configMapAssetUpdate := dbmodel.ConfigMapAssetUpdate{}

	if err := tx.Model(recId).Where(recId).First(&configMapAssetUpdate); err != nil {
		return nil, assetMeta, database.Error(database.UPDATE, api.ASSETKIND_CONFIG_MAP).WithDetails(err)
	}

	/* At this point we are not supporting spec update
	   configMapAssetUpdate.ConfigMapAssetSpec = request.Spec

	   result = tx.Where(recId).Save(&configMapAssetUpdate)
	   if result.Error != nil {
	       return nil, database.Error(database.UPDATE, api.ASSETKIND_CONFIG_MAP).WithDetailsFromDb(result)
	   }

	   dbErr := deletePolicyComplianceOfAsset(ctx, tx, assetMeta.Id)
	   if dbErr != nil {
	       return nil, dbErr
	   }
	*/

	return &api.ConfigMapAsset{
		Meta: assetMeta.ToApiAssetMeta(),
		Spec: configMapAssetUpdate.ConfigMapAssetSpec,
	}, assetMeta, nil
}
