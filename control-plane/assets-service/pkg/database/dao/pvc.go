/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/google/uuid"
)

func CreatePvcAsset(
	ctx context.Context,
	request api.PvcCreationRequest,
	nameValidationQuery *database.ParametrizedQuery,
	claimNameValidationQuery *database.ParametrizedQuery,
	datavolumeId *string,
) (*api.PvcAsset, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, err := createPvcAssetInTransaction(ctx, tx, request, nameValidationQuery, claimNameValidationQuery)
	if err != nil {
		return nil, err
	}

	if datavolumeId != nil {
		err = CreatePvcDatavolumeInTransaction(tx, result.Meta.Id, *datavolumeId)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_CREATIONS, api.ASSETKIND_PVC).WithDetails(err)
	}
	committedTx = true

	return result, nil
}

func createPvcAssetInTransaction(ctx context.Context, tx orm.DB, request api.PvcCreationRequest,
	nameValidationQuery *database.ParametrizedQuery, claimNameValidationQuery *database.ParametrizedQuery,
) (*api.PvcAsset, *api.Error) {
	id := uuid.New().String()

	// For non-existing PVCs, create a claim name lock and validate claim name uniqueness
	if claimNameValidationQuery != nil {
		lockRec, lockErr := createClaimNameLock(ctx, *request.Spec.ClaimName)
		if lockErr != nil {
			return nil, api.NewError(http.StatusConflict, fmt.Sprintf("creating a PVC asset with the claim name '%s' is temporarily unavailable, try again later", *request.Spec.ClaimName), lockErr)
		}
		defer deleteClaimNameLock(ctx, lockRec.Id)

		// Validate claim name uniqueness in database
		err := ValidateClaimNameUniqueness(ctx, tx, *request.Spec.ClaimName, claimNameValidationQuery)
		if err != nil {
			return nil, api.NewError(http.StatusConflict, fmt.Sprintf("PVC asset with claim name '%s' already exists in the scope hierarchy", *request.Spec.ClaimName), err)
		}
	}

	assetMeta, err := createAssetMeta(ctx, tx, id, api.ASSETKIND_PVC, request.Meta, nameValidationQuery)
	if err != nil {
		if strings.Contains(err.Error(), alreadyExistsError) {
			return nil, database.ErrorWithCode(database.CREATE, api.ASSETKIND_PVC, http.StatusConflict).WithDetails(err)
		}
		return nil, database.Error(database.CREATE, api.ASSETKIND_PVC).WithDetails(err)
	}

	pvcAssetRec := dbmodel.PVCAssetRec{
		AssetId: id,
		PvcAssetSpec: api.PvcAssetSpec{
			ClaimName:   request.Spec.ClaimName,
			ExistingPvc: request.Spec.ExistingPvc,
			Path:        request.Spec.Path,
			ReadOnly:    utils.ApplyBoolDefault(request.Spec.ReadOnly, false),
			ClaimInfo:   request.Spec.ClaimInfo,
			Ephemeral:   request.Spec.Ephemeral,
			DataSharing: request.Spec.DataSharing,
		},
	}

	_, err = tx.Create(pvcAssetRec)
	if err != nil {
		return nil, database.Error(database.CREATE, api.ASSETKIND_PVC).WithDetails(err)
	}

	return &api.PvcAsset{
		Meta:        *assetMeta,
		Spec:        request.Spec,
		ClusterInfo: getPvcClusterInfo(request.Spec),
	}, nil
}

func GetPvcAssetById(ctx context.Context, id string, options DatasourcesOptions) (*api.PvcAsset, *api.Error) {
	var combinedRec dbmodel.PVCCombinedRec
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_PVC,
			Id:   id,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, nil, dbtables.TableNameAssetsPVC)
	if err != nil {
		return nil, err
	}

	if options.IncludeDeleted {
		result = result.Unscoped()
	}

	scanRows, scanErr := result.Scan(&combinedRec)
	if scanErr != nil || scanRows == 0 {
		return nil, database.Error(database.GET, api.ASSETKIND_PVC).WithDetailsFromDbResult(scanRows, scanErr)
	}

	response := api.PvcAsset{
		Meta:        combinedRec.AssetMeta,
		Spec:        combinedRec.PvcAssetSpec,
		ClusterInfo: getPvcClusterInfo(combinedRec.PvcAssetSpec),
	}
	return &response, nil
}

func GetPvcAssetSpecById(ctx context.Context, id string) (*dbmodel.PVCAssetRec, *api.Error) {
	spec := dbmodel.PVCAssetRec{
		AssetId: id,
	}

	db := runaidb.DB().WithContext(ctx)

	rows, scanErr := db.Model(&spec).Where(spec).Scan(&spec)
	if scanErr != nil || rows == 0 {
		return nil, database.Error(database.GET, api.ASSETKIND_PVC).WithDetailsFromDbResult(rows, scanErr)
	}

	return &spec, nil
}

type DatasourcesOptions struct {
	IncludeDeleted bool
}

func ListPvcAssets(ctx context.Context, queryParams *query_types.QueryParams, options DatasourcesOptions) (*api.PvcListResponse, *api.Error) {
	combinedList := make([]dbmodel.PVCCombinedRec, 0)
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_PVC,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, queryParams, dbtables.TableNameAssetsPVC)
	if err != nil {
		return nil, err
	}

	if options.IncludeDeleted {
		result = result.Unscoped()
	}

	_, scanErr := result.
		Order(&query.AssetMeta.Id).
		Scan(&combinedList)
	if scanErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list PVC assets", scanErr)
	}

	response := api.PvcListResponse{
		Entries: make([]api.PvcAsset, 0),
	}

	for _, row := range combinedList {
		response.Entries = append(response.Entries, api.PvcAsset{
			Meta:        row.AssetMeta,
			Spec:        row.PvcAssetSpec,
			ClusterInfo: getPvcClusterInfo(row.PvcAssetSpec),
		})
	}

	return &response, nil
}

func ListPvcAssetsByClaimNames(ctx context.Context, queryParams *query_types.QueryParams, claimNames []string) ([]api.PvcAsset, *api.Error) {
	combinedList := make([]dbmodel.PVCCombinedRec, 0)
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_PVC,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, queryParams, dbtables.TableNameAssetsPVC)
	if err != nil {
		return nil, err
	}

	_, scanErr := result.
		Where(fmt.Sprintf("%s.%s IN ?", dbtables.TableNameAssetsPVC, dbmodel.ClaimNameColumn), claimNames).
		Order(&query.AssetMeta.Id).
		Scan(&combinedList)
	if scanErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list PVC assets by claim names", scanErr)
	}

	pvcAssets := make([]api.PvcAsset, 0)

	for _, row := range combinedList {
		pvcAssets = append(pvcAssets, api.PvcAsset{
			Meta: row.AssetMeta,
			Spec: row.PvcAssetSpec,
		})
	}

	return pvcAssets, nil
}

// FoundEphemeralPvcsByIds checks if any of the PVCs with the given IDs is ephemeral
func FoundEphemeralPvcsByIds(ctx context.Context, assetIds []string) (bool, *api.Error) {
	combinedList := make([]dbmodel.PVCCombinedRec, 0)
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_PVC,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, nil, dbtables.TableNameAssetsPVC)
	if err != nil {
		return false, err
	}

	_, scanErr := result.
		Where("id IN ? AND ephemeral = true", assetIds).
		Order(&query.AssetMeta.Id).
		Scan(&combinedList)
	if scanErr != nil {
		return false, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list PVC assets", scanErr)
	}

	return len(combinedList) > 0, nil
}

func UpdatePvcAsset(ctx context.Context, id string, meta api.AssetUpdateRequest, spec *dbmodel.PVCAssetRec, nameValidationQuery *database.ParametrizedQuery) (*api.PvcAsset, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, assetMeta, err := updatePvcAssetInTransaction(ctx, tx, id, meta, spec, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	policyConflict := ValidateAssetVsPolicy(ctx, id, ScopeInfoOfAssetMeta(&result.Meta))
	if policyConflict != nil {
		return nil, api.NewError(int(policyConflict.Code), policyConflict.Message, nil)
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_PVC).WithDetails(err)
	}
	committedTx = true

	NotifyAssetModified(ctx, id, assetMeta.Kind, ScopeInfoOfDbMeta(assetMeta))

	return result, nil
}

func CreatePvcDatavolumeInTransaction(tx orm.DB, pvcAssetId string, datavolumeId string) *api.Error {
	pvcDatavolumeRec := &dbmodel.PVCDatavolumeRec{
		AssetId:      pvcAssetId,
		DatavolumeId: datavolumeId,
	}

	_, err := tx.Model(&dbmodel.PVCDatavolumeRec{}).Create(pvcDatavolumeRec)
	if err != nil {
		return database.Error(database.CREATE, api.ASSETKIND_PVC).WithDetails(err)
	}

	return nil
}

func GetPvcDatavolumeId(ctx context.Context, pvcAssetId string) (*string, *api.Error) {
	db := runaidb.DB().WithContext(ctx)
	pvcDvRec := dbmodel.PVCDatavolumeRec{AssetId: pvcAssetId}

	_, err := db.Model(pvcDvRec).Where(pvcDvRec).Scan(&pvcDvRec)
	if err != nil {
		return nil, database.Error(database.GET, api.ASSETKIND_PVC).WithDetails(err)
	}

	return &pvcDvRec.DatavolumeId, nil
}

func updatePvcAssetInTransaction(ctx context.Context, tx orm.DB, id string, meta api.AssetUpdateRequest, updatedSpec *dbmodel.PVCAssetRec,
	nameValidationQuery *database.ParametrizedQuery,
) (*api.PvcAsset, *dbmodel.AssetMeta, *api.Error) {
	assetMeta, rowsAffected, apiErr, dbErr := updateAssetMeta(ctx, tx, id, meta, nameValidationQuery, nil)
	if apiErr != nil {
		return nil, nil, apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return nil, nil, database.Error(database.UPDATE, api.ASSETKIND_PVC).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	recId := dbmodel.PVCAssetRec{AssetId: id}

	saveRows, saveErr := tx.Where(recId).Save(&updatedSpec)
	if saveErr != nil {
		return nil, assetMeta, database.Error(database.UPDATE, api.ASSETKIND_PVC).WithDetailsFromDbResult(saveRows, saveErr)
	}

	if policyErr := deletePolicyComplianceOfAsset(ctx, tx, assetMeta.Id); policyErr != nil {
		return nil, assetMeta, policyErr
	}

	return &api.PvcAsset{
		Meta: assetMeta.ToApiAssetMeta(),
		Spec: updatedSpec.PvcAssetSpec,
	}, assetMeta, nil
}

func DeletePvcAsset(ctx context.Context, assetMeta *api.AssetMeta, deletionType DeletionType) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	err := deletePvcAssetInTransaction(ctx, tx, assetMeta.Id, deletionType)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return database.Error(database.COMMIT_DELETIONS, api.ASSETKIND_PVC).WithDetails(err)
	}
	committedTx = true

	NotifyAssetDeleted(ctx, assetMeta.Id, assetMeta.Kind, ScopeInfoOfAssetMeta(assetMeta))

	return nil
}

func deletePvcAssetInTransaction(ctx context.Context, tx orm.DB, id string, deletionType DeletionType) *api.Error {
	rowsAffected, apiErr, dbErr := deleteAssetMeta(ctx, tx, id, deletionType)
	if apiErr != nil {
		return apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_PVC).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	if deletionType == HardDelete {
		recId := dbmodel.PVCAssetRec{AssetId: id}

		delRows, delErr := tx.Where(&recId).Delete(&recId)
		if delErr != nil || delRows == 0 {
			return database.Error(database.DELETE, api.ASSETKIND_PVC).WithDetailsFromDbResult(delRows, delErr)
		}
	}

	if policyErr := deletePolicyComplianceOfAsset(ctx, tx, id); policyErr != nil {
		return policyErr
	}

	if err := deleteWorkloadAssetsByAssetId(ctx, tx, id); err != nil {
		return err
	}

	return nil
}

func getPvcClusterInfo(pvcSpec api.PvcAssetSpec) *api.ClusterInfo {
	return &api.ClusterInfo{
		Resources: []api.ClusterResourceInfo{
			{
				Name: pvcSpec.ClaimName,
			},
		},
	}
}
