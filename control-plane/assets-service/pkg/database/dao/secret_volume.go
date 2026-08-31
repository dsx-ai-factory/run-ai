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

	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/query_types"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
)

func NewSecretVolumeDBAccessor() *secretVolumeDBAccessor {
	return &secretVolumeDBAccessor{}
}

// SecretVolumeDBAccessor is a public singleton access for secret storage db accessor
var SecretVolumeDBAccessor secretVolumeDBAccessor

// secretVolumeDBAccessor implements the service's secretVolumeDBAccessor interface
type secretVolumeDBAccessor struct{}

// CreateSecretVolumeAsset insets a new Secret storage assets into the Secret and assets metadata tables
func (sdb *secretVolumeDBAccessor) CreateSecretVolumeAsset(ctx context.Context, request api.SecretAssetCreationRequest, nameValidationQuery *database.ParametrizedQuery) (*api.SecretAsset, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, err := createSecretAssetInTransaction(ctx, tx, request, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_CREATIONS, api.ASSETKIND_SECRET_VOLUME).WithDetails(err)
	}
	committedTx = true

	return result, nil
}

// DeleteSecretVolumeAsset deletes a Secret storage asset
func (sdb *secretVolumeDBAccessor) DeleteSecretVolumeAsset(ctx context.Context, assetMeta *dbmodel.AssetMeta, deletionType DeletionType) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	err := deleteSecretAssetInTransaction(ctx, tx, assetMeta.Id, deletionType)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return database.Error(database.COMMIT_DELETIONS, api.ASSETKIND_SECRET_VOLUME).WithDetails(err)
	}
	committedTx = true

	NotifyAssetDeleted(ctx, assetMeta.Id, assetMeta.Kind, ScopeInfoOfDbMeta(assetMeta))

	return nil
}

// GetSecretVolumeAssetById returns a Secret storage asset by it asset id
func (sdb *secretVolumeDBAccessor) GetSecretVolumeAssetById(ctx context.Context, assetId string) (*api.SecretAsset, *api.Error) {
	var combinedRec dbmodel.SecretAssetCombinedRec
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_SECRET_VOLUME,
			Id:   assetId,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, nil, dbtables.TableNameAssetsSecretVolume)
	if err != nil {
		return nil, err
	}

	scanRows, scanErr := result.Scan(&combinedRec)
	if scanErr != nil || scanRows == 0 {
		return nil, database.Error(database.GET, api.ASSETKIND_SECRET_VOLUME).WithDetailsFromDbResult(scanRows, scanErr)
	}

	response := api.SecretAsset{
		Meta: combinedRec.AssetMeta,
		Spec: combinedRec.SecretAssetSpec,
	}
	return &response, nil
}

// ListSecretVolumeAssets returns a list of Secret storage assets, filtered by query params
func (sdb *secretVolumeDBAccessor) ListSecretVolumeAssets(ctx context.Context, queryParams *query_types.QueryParams) (*api.SecretAssetListResponse, *api.Error) {
	combinedList := make([]dbmodel.SecretAssetCombinedRec, 0)
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_SECRET_VOLUME,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, queryParams, dbtables.TableNameAssetsSecretVolume)
	if err != nil {
		return nil, err
	}

	_, scanErr := result.
		Order(&query.AssetMeta.Id).
		Scan(&combinedList)
	if scanErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list ConfigMap assets", scanErr)
	}

	response := api.SecretAssetListResponse{
		Entries: make([]api.SecretAsset, 0),
	}

	for _, row := range combinedList {
		response.Entries = append(response.Entries, api.SecretAsset{
			Meta: row.AssetMeta,
			Spec: row.SecretAssetSpec,
		})
	}

	return &response, nil
}

// UpdateSecretVolumeAsset updates a Secret storage assets
func (sdb *secretVolumeDBAccessor) UpdateSecretVolumeAsset(ctx context.Context, assetId string,
	request api.SecretAssetUpdateRequest, nameValidationQuery *database.ParametrizedQuery,
) (*api.SecretAsset, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, assetMeta, err := updateSecretAssetInTransaction(ctx, tx, assetId, request, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	policyConflict := ValidateAssetVsPolicy(ctx, assetId, ScopeInfoOfAssetMeta(&result.Meta))
	if policyConflict != nil {
		return nil, api.NewError(int(policyConflict.Code), policyConflict.Message, nil)
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_SECRET_VOLUME).WithDetails(err)
	}
	committedTx = true

	NotifyAssetModified(ctx, assetId, assetMeta.Kind, ScopeInfoOfDbMeta(assetMeta))

	return result, nil
}

// createSecretAssetInTransaction insets a new Secret asset into the Secrets and assets metadata tables under a managed transaction
func createSecretAssetInTransaction(ctx context.Context, tx orm.DB, request api.SecretAssetCreationRequest, nameValidationQuery *database.ParametrizedQuery) (*api.SecretAsset, *api.Error) {
	id := uuid.New().String()

	assetMeta, err := createAssetMeta(ctx, tx, id, api.ASSETKIND_SECRET_VOLUME, request.Meta, nameValidationQuery)
	if err != nil {
		if strings.Contains(err.Error(), alreadyExistsError) {
			return nil, database.ErrorWithCode(database.CREATE, api.ASSETKIND_SECRET_VOLUME, http.StatusConflict).WithDetails(err)
		}
		return nil, database.Error(database.CREATE, api.ASSETKIND_SECRET_VOLUME).WithDetails(err)
	}

	configMapAssetRec := dbmodel.SecretAssetRec{
		AssetId: id,
		SecretAssetSpec: api.SecretAssetSpec{
			CredentialAssetId: request.Spec.CredentialAssetId,
			MountPath:         request.Spec.MountPath,
			DefaultMode:       request.Spec.DefaultMode,
		},
	}

	_, err = tx.Create(configMapAssetRec)
	if err != nil {
		return nil, database.Error(database.CREATE, api.ASSETKIND_SECRET_VOLUME).WithDetails(err)
	}

	return &api.SecretAsset{
		Meta: *assetMeta,
		Spec: request.Spec,
	}, nil
}

// deleteSecretAssetInTransaction deletes a Secret asset from both Secrets and assets metadata tables
func deleteSecretAssetInTransaction(ctx context.Context, tx orm.DB, assetId string, deletionType DeletionType) *api.Error {
	recId := dbmodel.SecretAssetRec{AssetId: assetId}

	rowsAffected, apiErr, dbErr := deleteAssetMeta(ctx, tx, assetId, deletionType)
	if apiErr != nil {
		return apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_SECRET_VOLUME).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	if deletionType == SoftDelete {
		return nil
	}

	delRows, delErr := tx.Where(&recId).Delete(&recId)
	if delErr != nil || delRows == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_SECRET_VOLUME).WithDetailsFromDbResult(delRows, delErr)
	}

	if policyErr := deletePolicyComplianceOfAsset(ctx, tx, assetId); policyErr != nil {
		return policyErr
	}

	if err := deleteWorkloadAssetsByAssetId(ctx, tx, assetId); err != nil {
		return err
	}

	return nil
}

// updateSecretAssetInTransaction updates a Secret asset from both Secrets and assets metadata tables
func updateSecretAssetInTransaction(ctx context.Context, tx orm.DB, assetId string, request api.SecretAssetUpdateRequest,
	nameValidationQuery *database.ParametrizedQuery,
) (*api.SecretAsset, *dbmodel.AssetMeta, *api.Error) {
	assetMeta, rowsAffected, apiErr, dbErr := updateAssetMeta(ctx, tx, assetId, request.Meta, nameValidationQuery, nil)
	if apiErr != nil {
		return nil, nil, apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return nil, nil, database.Error(database.UPDATE, api.ASSETKIND_SECRET_VOLUME).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	recId := dbmodel.SecretAssetRec{AssetId: assetId}
	secretAssetUpdate := dbmodel.SecretAssetUpdate{}

	if err := tx.Model(recId).Where(recId).First(&secretAssetUpdate); err != nil {
		return nil, assetMeta, database.Error(database.UPDATE, api.ASSETKIND_SECRET_VOLUME).WithDetails(err)
	}

	/* At this point we are not supporting spec update
	   secretAssetUpdate.SecretAssetSpec = request.Spec

	   result = tx.Where(recId).Save(&secretAssetUpdate)
	   if result.Error != nil {
	   	return nil, database.Error(database.UPDATE, api.ASSETKIND_CONFIG_MAP).WithDetailsFromDb(result)
	   }

	   dbErr := deletePolicyComplianceOfAsset(ctx, tx, assetMeta.Id)
	   if dbErr != nil {
	   	return nil, dbErr
	   }
	*/

	return &api.SecretAsset{
		Meta: assetMeta.ToApiAssetMeta(),
		Spec: secretAssetUpdate.SecretAssetSpec,
	}, assetMeta, nil
}

func countSecretVolumeAssetsUsingCredentialId(tx orm.DB, credentialAssetId string) (int64, *api.Error) {
	secretVolumeAssetRec := dbmodel.SecretAssetRec{
		SecretAssetSpec: api.SecretAssetSpec{
			CredentialAssetId: convutils.Ref(credentialAssetId),
		},
	}

	count := int64(0)
	if countErr := tx.Model(&secretVolumeAssetRec).
		Where(api.SecretAssetSpec{CredentialAssetId: secretVolumeAssetRec.SecretAssetSpec.CredentialAssetId}).
		Count(&count); countErr != nil {
		return 0, database.Error(database.COUNT, api.ASSETKIND_SECRET_VOLUME).WithDetails(countErr)
	}

	return count, nil
}

// ListSecretVolumeUsingCredentialId returns a list of Secret storage assets that are using a specific credential asset id
func ListSecretVolumeUsingCredentialId(ctx context.Context, credentialId *string) (*api.SecretAssetListResponse, *api.Error) {
	combinedList := make([]dbmodel.SecretAssetCombinedRec, 0)

	kindFilter := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_SECRET_VOLUME,
		},
	}

	clonePresetAssetsToTenant(ctx, auth.TenantIdOf(ctx))

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), kindFilter, nil, dbtables.TableNameAssetsSecretVolume)
	if err != nil {
		return nil, err
	}

	if credentialId != nil {
		result.Where(fmt.Sprintf("%s.%s = ?", dbtables.TableNameAssetsSecretVolume, dbtables.SecretVolumeColumnCredentialAssetId), *credentialId)
	} else {
		result.Where(fmt.Sprintf("%s.%s IS NOT NULL", dbtables.TableNameAssetsSecretVolume, dbtables.SecretVolumeColumnCredentialAssetId))
	}

	result.Order(fmt.Sprintf("%s.%s", dbtables.TableNameAssetsSecretVolume, dbtables.SecretVolumeColumnCredentialAssetId))

	_, scanErr := result.Scan(&combinedList)
	if scanErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list secret volume assets", scanErr)
	}

	response := api.SecretAssetListResponse{
		Entries: make([]api.SecretAsset, 0),
	}

	for _, row := range combinedList {
		response.Entries = append(response.Entries, api.SecretAsset{
			Meta: row.AssetMeta,
			Spec: row.SecretAssetSpec,
		})
	}

	return &response, nil
}
