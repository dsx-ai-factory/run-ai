/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

func CreateNFSAsset(ctx context.Context, request api.NfsCreationRequest, nameValidationQuery *database.ParametrizedQuery) (*api.NfsAsset, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, err := createNFSAssetInTransaction(ctx, tx, request, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_CREATIONS, api.ASSETKIND_NFS).WithDetails(err)
	}
	committedTx = true

	return result, nil
}

func createNFSAssetInTransaction(ctx context.Context, tx orm.DB, request api.NfsCreationRequest, nameValidationQuery *database.ParametrizedQuery) (*api.NfsAsset, *api.Error) {
	id := uuid.New().String()

	assetMeta, err := createAssetMeta(ctx, tx, id, api.ASSETKIND_NFS, request.Meta, nameValidationQuery)
	if err != nil {
		if strings.Contains(err.Error(), alreadyExistsError) {
			return nil, database.ErrorWithCode(database.CREATE, api.ASSETKIND_NFS, http.StatusConflict).WithDetails(err)
		}
		return nil, database.Error(database.CREATE, api.ASSETKIND_NFS).WithDetails(err)
	}

	nfsAssetRec := dbmodel.NFSAssetRec{
		AssetId: id,
		NfsAssetSpec: api.NfsAssetSpec{
			Path:      request.Spec.Path,
			ReadOnly:  utils.ApplyBoolDefault(request.Spec.ReadOnly, true),
			Server:    request.Spec.Server,
			MountPath: request.Spec.MountPath,
		},
	}

	_, err = tx.Create(nfsAssetRec)
	if err != nil {
		return nil, database.Error(database.CREATE, api.ASSETKIND_NFS).WithDetails(err)
	}

	return &api.NfsAsset{
		Meta: *assetMeta,
		Spec: request.Spec,
	}, nil
}

func GetNFSAssetById(ctx context.Context, id string) (*api.NfsAsset, *api.Error) {
	var combinedRec dbmodel.NFSCombinedRec
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_NFS,
			Id:   id,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, nil, dbtables.TableNameAssetsNFS)
	if err != nil {
		return nil, err
	}

	scanRows, scanErr := result.Scan(&combinedRec)
	if scanErr != nil || scanRows == 0 {
		return nil, database.Error(database.GET, api.ASSETKIND_NFS).WithDetailsFromDbResult(scanRows, scanErr)
	}

	response := api.NfsAsset{
		Meta: combinedRec.AssetMeta,
		Spec: combinedRec.NfsAssetSpec,
	}
	return &response, nil
}

func ListNFSAssets(ctx context.Context, queryParams *query_types.QueryParams) (*api.NfsListResponse, *api.Error) {
	combinedList := make([]dbmodel.NFSCombinedRec, 0)
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_NFS,
		},
	}

	clonePresetAssetsToTenant(ctx, auth.TenantIdOf(ctx))

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, queryParams, dbtables.TableNameAssetsNFS)
	if err != nil {
		return nil, err
	}

	_, scanErr := result.
		Order(&query.AssetMeta.Id).
		Scan(&combinedList)
	if scanErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list nfs assets", scanErr)
	}

	response := api.NfsListResponse{
		Entries: make([]api.NfsAsset, 0),
	}

	for _, row := range combinedList {
		response.Entries = append(response.Entries, api.NfsAsset{
			Meta: row.AssetMeta,
			Spec: row.NfsAssetSpec,
		})
	}

	return &response, nil
}

func UpdateNFSAsset(ctx context.Context, id string, request api.NfsUpdateRequest, nameValidationQuery *database.ParametrizedQuery) (*api.NfsAsset, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, assetMeta, err := updateNFSAssetInTransaction(ctx, tx, id, request, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	policyConflict := ValidateAssetVsPolicy(ctx, id, ScopeInfoOfAssetMeta(&result.Meta))
	if policyConflict != nil {
		return nil, api.NewError(int(policyConflict.Code), policyConflict.Message, nil)
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_NFS).WithDetails(err)
	}
	committedTx = true

	NotifyAssetModified(ctx, id, assetMeta.Kind, ScopeInfoOfDbMeta(assetMeta))

	return result, nil
}

func updateNFSAssetInTransaction(ctx context.Context, tx orm.DB, id string, request api.NfsUpdateRequest,
	nameValidationQuery *database.ParametrizedQuery,
) (*api.NfsAsset, *dbmodel.AssetMeta, *api.Error) {
	assetMeta, rowsAffected, apiErr, dbErr := updateAssetMeta(ctx, tx, id, request.Meta, nameValidationQuery, nil)
	if apiErr != nil {
		return nil, nil, apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return nil, nil, database.Error(database.UPDATE, api.ASSETKIND_NFS).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	recId := dbmodel.NFSAssetRec{AssetId: id}
	nfsAssetUpdate := dbmodel.NFSAssetUpdate{}

	if err := tx.Model(recId).Where(recId).First(&nfsAssetUpdate); err != nil {
		return nil, assetMeta, database.Error(database.UPDATE, api.ASSETKIND_NFS).WithDetails(err)
	}

	nfsAssetUpdate.NfsAssetSpec = request.Spec

	saveRows, saveErr := tx.Where(recId).Save(&nfsAssetUpdate)
	if saveErr != nil {
		return nil, assetMeta, database.Error(database.UPDATE, api.ASSETKIND_NFS).WithDetailsFromDbResult(saveRows, saveErr)
	}

	if policyErr := deletePolicyComplianceOfAsset(ctx, tx, assetMeta.Id); policyErr != nil {
		return nil, assetMeta, policyErr
	}

	return &api.NfsAsset{
		Meta: assetMeta.ToApiAssetMeta(),
		Spec: request.Spec,
	}, assetMeta, nil
}

func DeleteNFSAsset(ctx context.Context, assetMeta *dbmodel.AssetMeta, deletionType DeletionType) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	err := deleteNFSAssetInTransaction(ctx, tx, assetMeta.Id, deletionType)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return database.Error(database.COMMIT_DELETIONS, api.ASSETKIND_NFS).WithDetails(err)
	}
	committedTx = true

	NotifyAssetDeleted(ctx, assetMeta.Id, assetMeta.Kind, ScopeInfoOfDbMeta(assetMeta))

	return nil
}

func deleteNFSAssetInTransaction(ctx context.Context, tx orm.DB, id string, deletionType DeletionType) *api.Error {
	recId := dbmodel.NFSAssetRec{AssetId: id}

	rowsAffected, apiErr, dbErr := deleteAssetMeta(ctx, tx, id, deletionType)
	if apiErr != nil {
		return apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_NFS).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	delRows, delErr := tx.Where(&recId).Delete(&recId)
	if delErr != nil || delRows == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_NFS).WithDetailsFromDbResult(delRows, delErr)
	}

	if deletionType == SoftDelete {
		return nil
	}

	if policyErr := deletePolicyComplianceOfAsset(ctx, tx, id); policyErr != nil {
		return policyErr
	}

	if err := deleteWorkloadAssetsByAssetId(ctx, tx, id); err != nil {
		return err
	}

	return nil
}
