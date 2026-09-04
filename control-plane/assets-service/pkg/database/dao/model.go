/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
	"github.com/google/uuid"
)

func CreateModelAsset(ctx context.Context, request api.ModelCreationRequest,
	workloadAssets []dbmodel.AssetsContainedInWorkload, nameValidationQuery *database.ParametrizedQuery,
) (*api.ModelAsset, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, err := createModelAssetInTransaction(ctx, tx, request, workloadAssets, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_CREATIONS, api.ASSETKIND_MODEL).WithDetails(err)
	}
	committedTx = true

	return result, nil
}

func createModelAssetInTransaction(ctx context.Context, tx orm.DB, request api.ModelCreationRequest,
	workloadAssets []dbmodel.AssetsContainedInWorkload, nameValidationQuery *database.ParametrizedQuery,
) (*api.ModelAsset, *api.Error) {
	id := uuid.New().String()

	assetMeta, err := createAssetMeta(ctx, tx, id, api.ASSETKIND_MODEL, request.Meta, nameValidationQuery)
	if err != nil {
		return nil, database.Error(database.CREATE, api.ASSETKIND_MODEL).WithDetails(err)
	}

	modelAssetRec := dbmodel.ModelAssetRec{
		AssetId:       id,
		Title:         request.Spec.Title,
		Description:   convutils.Deref(request.Spec.Description, ""),
		Thumbnail:     convutils.Deref(request.Spec.Thumbnail, ""),
		HuggingfaceId: convutils.Deref(request.Spec.HuggingfaceId, ""),
		License:       convutils.Deref(request.Spec.License, ""),
	}

	if setErr := setAssets(workloadAssets, &modelAssetRec); setErr != nil {
		return nil, setErr
	}

	_, err = tx.Create(modelAssetRec)
	if err != nil {
		return nil, database.Error(database.CREATE, api.ASSETKIND_MODEL).WithDetails(err)
	}

	return &api.ModelAsset{
		Meta: *assetMeta,
		Spec: request.Spec,
	}, nil
}

func setAssets(workloadAssets []dbmodel.AssetsContainedInWorkload, modelAssetRec *dbmodel.ModelAssetRec) *api.Error {
	for _, asset := range workloadAssets {
		switch asset.Kind {
		case api.ASSETKIND_COMPUTE:
			modelAssetRec.ComputeAssetId = asset.Id

		case api.ASSETKIND_ENVIRONMENT:
			modelAssetRec.EnvironmentAssetId = asset.Id

		default:
			return api.NewError(http.StatusBadRequest, fmt.Sprintf("unsupported asset kind: %s", asset.Kind), nil)
		}
	}

	return nil
}

func GetModelAssetById(ctx context.Context, id string) (*api.ModelAsset, *api.Error) {
	var combinedRec dbmodel.ModelFullRec
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_MODEL,
			Id:   id,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, nil, dbtables.TableNameModel)
	if err != nil {
		return nil, err
	}

	scanRows, scanErr := result.Scan(&combinedRec)
	if scanErr != nil || scanRows == 0 {
		return nil, database.Error(database.GET, api.ASSETKIND_MODEL).WithDetailsFromDbResult(scanRows, scanErr)
	}

	spec := combinedRec.ModelAssetSpec
	spec.Assets.Environment = combinedRec.EnvironmentAssetId
	spec.Assets.Compute = convutils.Ref(combinedRec.ComputeAssetId)
	response := api.ModelAsset{
		Meta: combinedRec.AssetMeta,
		Spec: spec,
	}

	return &response, nil
}

func ListModels(ctx context.Context, queryParams *query_types.QueryParams, environmentAssetId string) (*api.ModelListResponse, *api.Error) {
	combinedList := make([]dbmodel.ModelFullRec, 0)
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_MODEL,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, queryParams, dbtables.TableNameModel)
	if err != nil {
		return nil, err
	}

	_, scanErr := result.
		Order(&query.AssetMeta.Id).
		Scan(&combinedList)
	if scanErr != nil {
		return nil, database.Error(database.LIST, api.ASSETKIND_MODEL)
	}

	response := api.ModelListResponse{
		Entries: make([]api.ModelAsset, 0),
	}

	for _, row := range combinedList {
		if environmentAssetId != "" && row.EnvironmentAssetId != environmentAssetId {
			continue
		}
		spec := row.ModelAssetSpec
		spec.Assets.Environment = row.EnvironmentAssetId
		spec.Assets.Compute = convutils.Ref(row.ComputeAssetId)

		response.Entries = append(response.Entries, api.ModelAsset{
			Meta: row.AssetMeta,
			Spec: spec,
		})
	}

	return &response, nil
}

func DeleteModelAssetById(ctx context.Context, id string, deletionType DeletionType) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	err := deleteModelAssetInTransaction(ctx, tx, id, deletionType)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return database.Error(database.COMMIT_DELETIONS, api.ASSETKIND_MODEL).WithDetails(err)
	}
	committedTx = true

	return nil
}

func deleteModelAssetInTransaction(ctx context.Context, tx orm.DB, id string, deletionType DeletionType) *api.Error {
	recId := dbmodel.ModelAssetRec{AssetId: id}

	rowsAffected, apiErr, dbErr := deleteAssetMeta(ctx, tx, id, deletionType)
	if apiErr != nil {
		return apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_MODEL).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	if deletionType == SoftDelete {
		return nil
	}

	delRows, delErr := tx.Where(&recId).Delete(&recId)
	if delErr != nil || delRows == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_MODEL).WithDetailsFromDbResult(delRows, delErr)
	}

	if policyErr := deletePolicyComplianceOfAsset(ctx, tx, id); policyErr != nil {
		return policyErr
	}

	if err := deleteWorkloadAssetsByAssetId(ctx, tx, id); err != nil {
		return err
	}

	return nil
}

func UpdateModelAssetById(ctx context.Context, id string, request api.ModelUpdateRequest, nameValidationQuery *database.ParametrizedQuery) (*api.ModelAsset, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, err := updateModelAssetInTransaction(ctx, tx, id, request, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_MODEL).WithDetails(err)
	}
	committedTx = true

	return result, nil
}

func updateModelAssetInTransaction(ctx context.Context, tx orm.DB, id string, request api.ModelUpdateRequest, nameValidationQuery *database.ParametrizedQuery) (*api.ModelAsset, *api.Error) {
	assetMeta, rowsAffected, apiErr, dbErr := updateAssetMeta(ctx, tx, id, request.Meta, nameValidationQuery, nil)
	if apiErr != nil {
		return nil, apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return nil, database.Error(database.UPDATE, api.ASSETKIND_MODEL).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	recId := dbmodel.ModelAssetRec{AssetId: id}

	modelAssetRec := dbmodel.ModelAssetRec{}

	if err := tx.Model(recId).Where(recId).First(&modelAssetRec); err != nil {
		return nil, database.Error(database.UPDATE, api.ASSETKIND_MODEL).WithDetails(err)
	}

	_, err := tx.Where(recId).Save(&modelAssetRec)
	if err != nil {
		return nil, database.Error(database.UPDATE, api.ASSETKIND_MODEL).WithDetails(err)
	}

	if err := deletePolicyComplianceOfAsset(ctx, tx, assetMeta.Id); err != nil {
		return nil, err
	}

	return &api.ModelAsset{
		Meta: assetMeta.ToApiAssetMeta(),
		Spec: request.Spec,
	}, nil
}
