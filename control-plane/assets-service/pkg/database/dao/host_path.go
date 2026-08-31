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
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/query_types"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

func CreateHostPathAsset(ctx context.Context, request api.HostPathCreationRequest, nameValidationQuery *database.ParametrizedQuery) (*api.HostPathAsset, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, err := createHostPathAssetInTransaction(ctx, tx, request, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_CREATIONS, api.ASSETKIND_HOST_PATH).WithDetails(err)
	}
	committedTx = true

	return result, nil
}

func createHostPathAssetInTransaction(ctx context.Context, tx orm.DB, request api.HostPathCreationRequest, nameValidationQuery *database.ParametrizedQuery) (*api.HostPathAsset, *api.Error) {
	id := uuid.New().String()

	assetMeta, err := createAssetMeta(ctx, tx, id, api.ASSETKIND_HOST_PATH, request.Meta, nameValidationQuery)
	if err != nil {
		if strings.Contains(err.Error(), alreadyExistsError) {
			return nil, database.ErrorWithCode(database.CREATE, api.ASSETKIND_HOST_PATH, http.StatusConflict).WithDetails(err)
		}
		return nil, database.Error(database.CREATE, api.ASSETKIND_HOST_PATH).WithDetails(err)
	}

	hostPathAssetRec := dbmodel.HostPathAssetRec{
		AssetId: id,
		HostPathSpec: api.HostPathSpec{
			Path:             request.Spec.Path,
			ReadOnly:         request.Spec.ReadOnly,
			MountPath:        request.Spec.MountPath,
			MountPropagation: request.Spec.MountPropagation,
		},
	}

	_, err = tx.Create(hostPathAssetRec)
	if err != nil {
		return nil, database.Error(database.CREATE, api.ASSETKIND_HOST_PATH).WithDetails(err)
	}

	return &api.HostPathAsset{
		Meta: *assetMeta,
		Spec: request.Spec,
	}, nil
}

func GetHostPathById(ctx context.Context, id string) (*api.HostPathAsset, *api.Error) {
	var combinedRec dbmodel.HostPathCombinedRec
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_HOST_PATH,
			Id:   id,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, nil, dbtables.TableNameAssetsHostPath)
	if err != nil {
		return nil, err
	}

	scanRows, scanErr := result.Scan(&combinedRec)
	if scanErr != nil || scanRows == 0 {
		return nil, database.Error(database.GET, api.ASSETKIND_HOST_PATH).WithDetailsFromDbResult(scanRows, scanErr)
	}

	response := api.HostPathAsset{
		Meta: combinedRec.AssetMeta,
		Spec: combinedRec.HostPathSpec,
	}
	return &response, nil
}

func ListHostPathAssets(ctx context.Context, queryParams *query_types.QueryParams) (*api.HostPathListResponse, *api.Error) {
	combinedList := make([]dbmodel.HostPathCombinedRec, 0)
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_HOST_PATH,
		},
	}

	clonePresetAssetsToTenant(ctx, auth.TenantIdOf(ctx))

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, queryParams, dbtables.TableNameAssetsHostPath)
	if err != nil {
		return nil, err
	}

	_, scanErr := result.
		Order(&query.AssetMeta.Id).
		Scan(&combinedList)
	if scanErr != nil {
		return nil, database.Error(database.LIST, api.ASSETKIND_HOST_PATH).WithDetails(scanErr)
	}

	response := api.HostPathListResponse{
		Entries: make([]api.HostPathAsset, 0),
	}

	for _, row := range combinedList {
		response.Entries = append(response.Entries, api.HostPathAsset{
			Meta: row.AssetMeta,
			Spec: row.HostPathSpec,
		})
	}

	return &response, nil
}

func UpdateHostPath(ctx context.Context, id string, request api.HostPathUpdateRequest, nameValidationQuery *database.ParametrizedQuery) (*api.HostPathAsset, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, assetMeta, err := updateHostPathInTransaction(ctx, tx, id, request, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	policyConflict := ValidateAssetVsPolicy(ctx, id, ScopeInfoOfAssetMeta(&result.Meta))
	if policyConflict != nil {
		return nil, api.NewError(int(policyConflict.Code), policyConflict.Message, nil)
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_HOST_PATH).WithDetails(err)
	}
	committedTx = true

	NotifyAssetModified(ctx, id, assetMeta.Kind, ScopeInfoOfDbMeta(assetMeta))

	return result, nil
}

func updateHostPathInTransaction(ctx context.Context, tx orm.DB, id string, request api.HostPathUpdateRequest,
	nameValidationQuery *database.ParametrizedQuery,
) (*api.HostPathAsset, *dbmodel.AssetMeta, *api.Error) {
	assetMeta, rowsAffected, apiErr, dbErr := updateAssetMeta(ctx, tx, id, request.Meta, nameValidationQuery, nil)
	if apiErr != nil {
		return nil, nil, apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return nil, nil, database.Error(database.UPDATE, api.ASSETKIND_HOST_PATH).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	recId := dbmodel.HostPathAssetRec{AssetId: id}
	hostPathAssetUpdate := dbmodel.HostPathAssetUpdate{}

	if err := tx.Model(recId).Where(recId).First(&hostPathAssetUpdate); err != nil {
		return nil, assetMeta, database.Error(database.UPDATE, api.ASSETKIND_HOST_PATH).WithDetails(err)
	}

	hostPathAssetUpdate.HostPathSpec = request.Spec

	saveRows, saveErr := tx.Where(recId).Save(&hostPathAssetUpdate)
	if saveErr != nil {
		return nil, assetMeta, database.Error(database.UPDATE, api.ASSETKIND_HOST_PATH).WithDetailsFromDbResult(saveRows, saveErr)
	}

	if policyErr := deletePolicyComplianceOfAsset(ctx, tx, assetMeta.Id); policyErr != nil {
		return nil, assetMeta, policyErr
	}

	return &api.HostPathAsset{
		Meta: assetMeta.ToApiAssetMeta(),
		Spec: request.Spec,
	}, assetMeta, nil
}

func DeleteHostPathAsset(ctx context.Context, assetMeta *dbmodel.AssetMeta, deletionType DeletionType) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	err := deleteHostPathAssetInTransaction(ctx, tx, assetMeta.Id, deletionType)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return database.Error(database.COMMIT_DELETIONS, api.ASSETKIND_HOST_PATH).WithDetails(err)
	}
	committedTx = true

	NotifyAssetDeleted(ctx, assetMeta.Id, assetMeta.Kind, ScopeInfoOfDbMeta(assetMeta))

	return nil
}

func deleteHostPathAssetInTransaction(ctx context.Context, tx orm.DB, id string, deletionType DeletionType) *api.Error {
	recId := dbmodel.HostPathAssetRec{AssetId: id}

	rowsAffected, apiErr, dbErr := deleteAssetMeta(ctx, tx, id, deletionType)
	if apiErr != nil {
		return apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_HOST_PATH).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	if deletionType == SoftDelete {
		return nil
	}

	delRows, delErr := tx.Where(&recId).Delete(&recId)
	if delErr != nil || delRows == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_HOST_PATH).WithDetailsFromDbResult(delRows, delErr)
	}

	if policyErr := deletePolicyComplianceOfAsset(ctx, tx, id); policyErr != nil {
		return policyErr
	}

	if err := deleteWorkloadAssetsByAssetId(ctx, tx, id); err != nil {
		return err
	}

	return nil
}
