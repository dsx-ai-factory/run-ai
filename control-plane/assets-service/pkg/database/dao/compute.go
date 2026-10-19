/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/google/uuid"
)

func CreateComputeAsset(ctx context.Context, request api.ComputeCreationRequest, nameValidationQuery *database.ParametrizedQuery) (*api.ComputeAsset, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, err := createComputeAssetInTransaction(ctx, tx, request, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_CREATIONS, api.ASSETKIND_COMPUTE).WithDetails(err)
	}
	committedTx = true

	return result, nil
}

func createComputeAssetInTransaction(ctx context.Context, tx orm.DB, request api.ComputeCreationRequest, nameValidationQuery *database.ParametrizedQuery) (*api.ComputeAsset, *api.Error) {
	id := uuid.New().String()

	assetMeta, err := createAssetMeta(ctx, tx, id, api.ASSETKIND_COMPUTE, request.Meta, nameValidationQuery)
	if err != nil {
		if strings.Contains(err.Error(), alreadyExistsError) {
			return nil, database.ErrorWithCode(database.CREATE, api.ASSETKIND_COMPUTE, http.StatusConflict).WithDetails(err)
		}
		return nil, database.Error(database.CREATE, api.ASSETKIND_COMPUTE).WithDetails(err)
	}

	ComputeAssetRec := dbmodel.ComputeAssetRec{
		AssetId: id,
	}
	if request.Spec != nil {
		stored, err := dbmodel.ComputeAssetForDB(*request.Spec)
		if err != nil {
			return nil, api.NewValidationError(err.Error())
		}
		ComputeAssetRec.ComputeAssetSpec = stored
	}

	_, err = tx.Create(ComputeAssetRec)
	if err != nil {
		return nil, database.Error(database.CREATE, api.ASSETKIND_COMPUTE).WithDetails(err)
	}

	return &api.ComputeAsset{
		Meta: *assetMeta,
		Spec: dbmodel.LegacyCompatibleComputeAsset(convutils.Deref(request.Spec, api.ComputeAssetSpec{})),
	}, nil
}

func GetComputeAssetById(ctx context.Context, id string) (*api.ComputeAsset, *api.Error) {
	var combinedRec dbmodel.ComputeFullRec
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_COMPUTE,
			Id:   id,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, nil, dbtables.TableNameAssetsCompute)
	if err != nil {
		return nil, err
	}

	scanRows, scanErr := result.Scan(&combinedRec)
	if scanErr != nil || scanRows == 0 {
		return nil, database.Error(database.GET, api.ASSETKIND_COMPUTE).WithDetailsFromDbResult(scanRows, scanErr)
	}

	response := api.ComputeAsset{
		Meta: combinedRec.AssetMeta,
		Spec: dbmodel.LegacyCompatibleComputeAsset(combinedRec.ComputeAssetSpec),
	}
	return &response, nil
}

func ListComputeAssets(ctx context.Context, queryParams *query_types.QueryParams) (*api.ComputeListResponse, *api.Error) {
	combinedList := make([]dbmodel.ComputeFullRec, 0)
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_COMPUTE,
		},
	}

	clonePresetAssetsToTenant(ctx, auth.TenantIdOf(ctx))

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, queryParams, dbtables.TableNameAssetsCompute)
	if err != nil {
		return nil, err
	}

	_, scanErr := result.
		Order(&query.AssetMeta.Id).
		Scan(&combinedList)
	if scanErr != nil {
		return nil, database.Error(database.LIST, api.ASSETKIND_COMPUTE)
	}

	response := api.ComputeListResponse{
		Entries: make([]api.ComputeAsset, 0),
	}

	for _, row := range combinedList {
		response.Entries = append(response.Entries, api.ComputeAsset{
			Meta: row.AssetMeta,
			Spec: dbmodel.LegacyCompatibleComputeAsset(row.ComputeAssetSpec),
		})
	}

	return &response, nil
}

func UpdateComputeAsset(ctx context.Context, id string, request api.ComputeUpdateRequest, nameValidationQuery *database.ParametrizedQuery) (*api.ComputeAsset, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, assetMeta, err := updateComputeAssetInTransaction(ctx, tx, id, request, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_COMPUTE).WithDetails(err)
	}
	committedTx = true

	NotifyAssetModified(ctx, id, assetMeta.Kind, ScopeInfoOfDbMeta(assetMeta))

	return result, err
}

func updateComputeAssetInTransaction(ctx context.Context, tx orm.DB, id string,
	request api.ComputeUpdateRequest, nameValidationQuery *database.ParametrizedQuery,
) (*api.ComputeAsset, *dbmodel.AssetMeta, *api.Error) {
	assetMeta, rowsAffected, apiErr, dbErr := updateAssetMeta(ctx, tx, id, request.Meta, nameValidationQuery, nil)
	if apiErr != nil {
		return nil, nil, apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return nil, nil, database.Error(database.UPDATE, api.ASSETKIND_COMPUTE).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	recId := dbmodel.ComputeAssetRec{AssetId: id}
	computeAssetUpdate := dbmodel.ComputeAssetUpdate{}

	if err := tx.Model(recId).Where(recId).First(&computeAssetUpdate); err != nil {
		return nil, assetMeta, database.Error(database.UPDATE, api.ASSETKIND_COMPUTE).WithDetails(err)
	}

	if request.Spec != nil {
		stored, err := dbmodel.ComputeAssetForDB(*request.Spec)
		if err != nil {
			return nil, assetMeta, api.NewValidationError(err.Error())
		}
		computeAssetUpdate.ComputeAssetSpec = stored
	}

	saveRows, saveErr := tx.Where(recId).Save(&computeAssetUpdate)
	if saveErr != nil {
		return nil, assetMeta, database.Error(database.UPDATE, api.ASSETKIND_COMPUTE).WithDetailsFromDbResult(saveRows, saveErr)
	}

	if policyErr := deletePolicyComplianceOfAsset(ctx, tx, assetMeta.Id); policyErr != nil {
		return nil, assetMeta, policyErr
	}

	return &api.ComputeAsset{
		Meta: assetMeta.ToApiAssetMeta(),
		Spec: dbmodel.LegacyCompatibleComputeAsset(convutils.Deref(request.Spec, api.ComputeAssetSpec{})),
	}, assetMeta, nil
}

func DeleteComputeAsset(ctx context.Context, assetMeta *dbmodel.AssetMeta, deletionType DeletionType) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	err := deleteComputeAssetInTransaction(ctx, tx, assetMeta.Id, deletionType)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return database.Error(database.COMMIT_DELETIONS, api.ASSETKIND_COMPUTE).WithDetails(err)
	}
	committedTx = true

	NotifyAssetDeleted(ctx, assetMeta.Id, assetMeta.Kind, ScopeInfoOfDbMeta(assetMeta))

	return nil
}

func deleteComputeAssetInTransaction(ctx context.Context, tx orm.DB, id string, deletionType DeletionType) *api.Error {
	recId := dbmodel.ComputeAssetRec{AssetId: id}

	rowsAffected, apiErr, dbErr := deleteAssetMeta(ctx, tx, id, deletionType)
	if apiErr != nil {
		return apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_COMPUTE).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	if deletionType == SoftDelete {
		return nil
	}

	delRows, delErr := tx.Where(&recId).Delete(&recId)
	if delErr != nil || delRows == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_COMPUTE).WithDetailsFromDbResult(delRows, delErr)
	}

	if policyErr := deletePolicyComplianceOfAsset(ctx, tx, id); policyErr != nil {
		return policyErr
	}

	if err := deleteWorkloadAssetsByAssetId(ctx, tx, id); err != nil {
		return err
	}

	return nil
}
