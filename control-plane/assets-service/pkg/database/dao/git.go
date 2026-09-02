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

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"

	scope "github.com/dsx-ai-factory/run-ai/runai-application-packages/scopeinfo"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
)

func CreateGitAsset(ctx context.Context, request api.GitCreationRequest, nameValidationQuery *database.ParametrizedQuery) (*api.GitAsset, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, err := createGitAssetInTransaction(ctx, tx, request, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_CREATIONS, api.ASSETKIND_GIT).WithDetails(err)
	}
	committedTx = true

	return result, nil
}

func createGitAssetInTransaction(ctx context.Context, tx orm.DB, request api.GitCreationRequest, nameValidationQuery *database.ParametrizedQuery) (*api.GitAsset, *api.Error) {
	id := uuid.New().String()

	assetMeta, err := createAssetMeta(ctx, tx, id, api.ASSETKIND_GIT, request.Meta, nameValidationQuery)
	if err != nil {
		if strings.Contains(err.Error(), alreadyExistsError) {
			return nil, database.ErrorWithCode(database.CREATE, api.ASSETKIND_GIT, http.StatusConflict).WithDetails(err)
		}
		return nil, database.Error(database.CREATE, api.ASSETKIND_GIT).WithDetails(err)
	}

	gitAssetRec := dbmodel.GitAssetRec{
		AssetId: id,
		GitAssetSpec: api.GitAssetSpec{
			Repository:      request.Spec.Repository,
			Branch:          request.Spec.Branch,
			Revision:        request.Spec.Revision,
			Path:            request.Spec.Path,
			PasswordAssetId: request.Spec.PasswordAssetId,
			CredentialId:    request.Spec.CredentialId,
		},
	}

	_, err = tx.Create(gitAssetRec)
	if err != nil {
		return nil, database.Error(database.CREATE, api.ASSETKIND_GIT).WithDetails(err)
	}

	return &api.GitAsset{
		Meta: *assetMeta,
		Spec: request.Spec,
	}, nil
}

func GetGitAssetById(ctx context.Context, id string) (*api.GitAsset, *api.Error) {
	var combinedRec dbmodel.GitCombinedRec
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_GIT,
			Id:   id,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, nil, dbtables.TableNameAssetsGit)
	if err != nil {
		return nil, err
	}

	scanRows, scanErr := result.Scan(&combinedRec)
	if scanErr != nil || scanRows == 0 {
		return nil, database.Error(database.GET, api.ASSETKIND_GIT).WithDetailsFromDbResult(scanRows, scanErr)
	}

	response := api.GitAsset{
		Meta: combinedRec.AssetMeta,
		Spec: combinedRec.GitAssetSpec,
	}
	return &response, nil
}

func ListGitAssets(ctx context.Context, queryParams *query_types.QueryParams) (*api.GitListResponse, *api.Error) {
	combinedList := make([]dbmodel.GitCombinedRec, 0)
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_GIT,
		},
	}

	clonePresetAssetsToTenant(ctx, auth.TenantIdOf(ctx))

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, queryParams, dbtables.TableNameAssetsGit)
	if err != nil {
		return nil, err
	}

	_, scanErr := result.
		Order(&query.AssetMeta.Id).
		Scan(&combinedList)
	if scanErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list git assets", scanErr)
	}

	response := api.GitListResponse{
		Entries: make([]api.GitAsset, 0),
	}

	for _, row := range combinedList {
		response.Entries = append(response.Entries, api.GitAsset{
			Meta: row.AssetMeta,
			Spec: row.GitAssetSpec,
		})
	}

	return &response, nil
}

func ListGitUsingPassword(ctx context.Context, specificId *string) (*api.GitListResponse, *api.Error) {
	combinedList := make([]dbmodel.GitCombinedRec, 0)

	kindFilter := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_GIT,
		},
	}

	clonePresetAssetsToTenant(ctx, auth.TenantIdOf(ctx))

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), kindFilter, nil, dbtables.TableNameAssetsGit)
	if err != nil {
		return nil, err
	}

	if specificId != nil {
		// Filter for git assets that use the specific password asset ID (either in password_asset_id or credential_id)
		result.Where(fmt.Sprintf("(%s.%s = ? OR %s.%s = ?)",
			dbtables.TableNameAssetsGit, dbtables.GitColumnPasswordAssetId,
			dbtables.TableNameAssetsGit, dbtables.GitColumnCredentialId), *specificId, *specificId)
	} else {
		// Join with assets_metadata to check the kind of credential_id
		assets_metadata := "am"
		result.Joins(fmt.Sprintf("LEFT JOIN %s AS %s ON %s.%s = %s.%s",
			dbtables.TableNameAssetsMeta, assets_metadata,
			dbtables.TableNameAssetsGit, dbtables.GitColumnCredentialId,
			assets_metadata, dbtables.AssetMetaColumnId))

		// Filter for git assets that either:
		// 1. Have a password_asset_id (non-null), OR
		// 2. Have a credential_id that points to a password asset
		result.Where(fmt.Sprintf("(%s.%s IS NOT NULL OR (%s.%s IS NOT NULL AND %s.%s = ?))",
			dbtables.TableNameAssetsGit, dbtables.GitColumnPasswordAssetId,
			dbtables.TableNameAssetsGit, dbtables.GitColumnCredentialId,
			assets_metadata, dbtables.AssetMetaColumnKind), string(api.ASSETKIND_PASSWORD))
	}

	result.Order(fmt.Sprintf("%s.%s", dbtables.TableNameAssetsGit, dbtables.GitColumnCredentialId))

	_, scanErr := result.Scan(&combinedList)
	if scanErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list git assets", scanErr)
	}

	response := api.GitListResponse{
		Entries: make([]api.GitAsset, 0),
	}

	for _, row := range combinedList {
		response.Entries = append(response.Entries, api.GitAsset{
			Meta: row.AssetMeta,
			Spec: row.GitAssetSpec,
		})
	}

	return &response, nil
}

func ListGitUsingGenericSecret(ctx context.Context, specificId *string) (*api.GitListResponse, *api.Error) {
	combinedList := make([]dbmodel.GitCombinedRec, 0)

	kindFilter := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_GIT,
		},
	}

	clonePresetAssetsToTenant(ctx, auth.TenantIdOf(ctx))

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), kindFilter, nil, dbtables.TableNameAssetsGit)
	if err != nil {
		return nil, err
	}

	if specificId != nil {
		// Filter for git assets that use the specific generic secret asset ID
		result.Where(fmt.Sprintf("(%s.%s = ?)",
			dbtables.TableNameAssetsGit, dbtables.GitColumnCredentialId), *specificId)
	} else {
		// Join with assets_metadata to check the kind of credential_id
		assets_metadata := "am"
		result.Joins(fmt.Sprintf("LEFT JOIN %s AS %s ON %s.%s = %s.%s",
			dbtables.TableNameAssetsMeta, assets_metadata,
			dbtables.TableNameAssetsGit, dbtables.GitColumnCredentialId,
			assets_metadata, dbtables.AssetMetaColumnId))

		// Filter for git assets have a credential_id that points to a generic asset
		result.Where(fmt.Sprintf("%s.%s IS NOT NULL AND %s.%s = ?",
			dbtables.TableNameAssetsGit, dbtables.GitColumnCredentialId,
			assets_metadata, dbtables.AssetMetaColumnKind), string(api.ASSETKIND_GENERIC_SECRET))
	}

	result.Order(fmt.Sprintf("%s.%s", dbtables.TableNameAssetsGit, dbtables.GitColumnCredentialId))

	_, scanErr := result.Scan(&combinedList)
	if scanErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list git assets", scanErr)
	}

	response := api.GitListResponse{
		Entries: make([]api.GitAsset, 0),
	}

	for _, row := range combinedList {
		response.Entries = append(response.Entries, api.GitAsset{
			Meta: row.AssetMeta,
			Spec: row.GitAssetSpec,
		})
	}

	return &response, nil
}

func UpdateGitAsset(ctx context.Context, id string, request api.GitUpdateRequest, validationHook AssetVsCredentialsValidationHook, nameValidationQuery *database.ParametrizedQuery) (*api.GitAsset, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, assetMeta, err := updateGitAssetInTransaction(ctx, tx, id, request, validationHook, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	policyConflict := ValidateAssetVsPolicy(ctx, id, ScopeInfoOfAssetMeta(&result.Meta))
	if policyConflict != nil {
		return nil, api.NewError(int(policyConflict.Code), policyConflict.Message, nil)
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_GIT).WithDetails(err)
	}
	committedTx = true

	NotifyAssetModified(ctx, id, assetMeta.Kind, ScopeInfoOfDbMeta(assetMeta))

	return result, nil
}

func updateGitAssetInTransaction(ctx context.Context, tx orm.DB, id string, request api.GitUpdateRequest, validationHook AssetVsCredentialsValidationHook,
	nameValidationQuery *database.ParametrizedQuery,
) (*api.GitAsset, *dbmodel.AssetMeta, *api.Error) {
	assetMeta, rowsAffected, apiErr, dbErr := updateAssetMeta(ctx, tx, id, request.Meta, nameValidationQuery, nil)
	if apiErr != nil {
		return nil, nil, apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return nil, nil, database.Error(database.UPDATE, api.ASSETKIND_GIT).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	if validationHook != nil {
		scopeInfo := ScopeInfoOfDbMeta(assetMeta)
		validationErr := validationHook(ctx, assetMeta.Scope, scopeInfo.ScopeId, dbmodel.GetGitCredentialId(request.Spec))
		if validationErr != nil {
			return nil, assetMeta, validationErr
		}
	}

	recId := dbmodel.GitAssetRec{AssetId: id}
	gitAssetUpdate := dbmodel.GitAssetUpdate{}

	if err := tx.Model(recId).Where(recId).First(&gitAssetUpdate); err != nil {
		return nil, assetMeta, database.Error(database.UPDATE, api.ASSETKIND_GIT).WithDetails(err)
	}

	gitAssetUpdate.GitAssetSpec = request.Spec

	saveRows, saveErr := tx.Where(recId).Save(&gitAssetUpdate)
	if saveErr != nil {
		return nil, assetMeta, database.Error(database.UPDATE, api.ASSETKIND_GIT).WithDetailsFromDbResult(saveRows, saveErr)
	}

	if policyErr := deletePolicyComplianceOfAsset(ctx, tx, assetMeta.Id); policyErr != nil {
		return nil, assetMeta, policyErr
	}

	return &api.GitAsset{
		Meta: assetMeta.ToApiAssetMeta(),
		Spec: request.Spec,
	}, assetMeta, nil
}

func DeleteGitAsset(ctx context.Context, assetMeta *dbmodel.AssetMeta, deletionType DeletionType) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	err := deleteGitAssetInTransaction(ctx, tx, assetMeta.Id, deletionType)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return database.Error(database.COMMIT_DELETIONS, api.ASSETKIND_GIT).WithDetails(err)
	}
	committedTx = true

	NotifyAssetDeleted(ctx, assetMeta.Id, assetMeta.Kind, ScopeInfoOfDbMeta(assetMeta))

	return nil
}

func deleteGitAssetInTransaction(ctx context.Context, tx orm.DB, id string, deletionType DeletionType) *api.Error {
	recId := dbmodel.GitAssetRec{AssetId: id}

	rowsAffected, apiErr, dbErr := deleteAssetMeta(ctx, tx, id, deletionType)
	if apiErr != nil {
		return apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_GIT).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	if deletionType == SoftDelete {
		return nil
	}

	delRows, delErr := tx.Where(&recId).Delete(&recId)
	if delErr != nil || delRows == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_GIT).WithDetailsFromDbResult(delRows, delErr)
	}

	if policyErr := deletePolicyComplianceOfAsset(ctx, tx, id); policyErr != nil {
		return policyErr
	}

	if err := deleteWorkloadAssetsByAssetId(ctx, tx, id); err != nil {
		return err
	}

	return nil
}

// countGitAssetsUsingCredentialByKind counts git assets using a specific credential ID of a given kind
func countGitAssetsUsingCredentialByKind(tx orm.DB, credentialId string, credentialKind api.AssetKind) (int64, *api.Error) {
	count := int64(0)
	countErr := tx.Table(dbtables.TableNameAssetsGit).
		Joins(fmt.Sprintf("JOIN %s ON %s.%s = %s.%s",
			dbtables.TableNameAssetsMeta,
			dbtables.TableNameAssetsGit, dbtables.GitColumnCredentialId,
			dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnId)).
		Where(fmt.Sprintf("%s.%s = ? AND %s.%s = ?",
			dbtables.TableNameAssetsGit, dbtables.GitColumnCredentialId,
			dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnKind), credentialId, string(credentialKind)).
		Count(&count)

	if countErr != nil {
		return 0, database.Error(database.COUNT, api.ASSETKIND_GIT).WithDetails(countErr)
	}

	return count, nil
}

// countGitAssetsUsingPassword counts the number of git assets which uses a given password credentials asset
func countGitAssetsUsingPassword(tx orm.DB, passwordAssetId string) (int64, *api.Error) {
	// Instated of having a single complex query, we will count git assets in two parts:
	// 1. Count git assets using credential_id that points to password assets
	// 2. Count git assets using password_asset_id directly
	// the benefit of this approach is that it is more readable and maintainable,
	// * deleting password assets is a rare operation
	// * as password_asset_id is deprecated, later on it could be easily removed

	// Count git assets using credential_id that points to password assets
	credentialIdCount, err := countGitAssetsUsingCredentialByKind(tx, passwordAssetId, api.ASSETKIND_PASSWORD)
	if err != nil {
		return 0, err
	}

	// Count git assets using password credential
	gitAssetRec := dbmodel.GitAssetRec{
		GitAssetSpec: api.GitAssetSpec{
			PasswordAssetId: convutils.Ref(passwordAssetId),
		},
	}

	passwordAssetIdCount := int64(0)
	if countErr := tx.Model(&gitAssetRec).
		Where(api.GitAssetSpec{PasswordAssetId: gitAssetRec.GitAssetSpec.PasswordAssetId}).
		Count(&passwordAssetIdCount); countErr != nil {
		return 0, database.Error(database.COUNT, api.ASSETKIND_GIT).WithDetails(countErr)
	}

	return credentialIdCount + passwordAssetIdCount, nil
}

// countGitAssetsUsingGenericSecret counts the number of git assets which uses a given generic secret credentials asset
func countGitAssetsUsingGenericSecret(tx orm.DB, genericSecretAssetId string) (int64, *api.Error) {
	// Use the helper function to count git assets using credential_id that points to generic secret assets
	return countGitAssetsUsingCredentialByKind(tx, genericSecretAssetId, api.ASSETKIND_GENERIC_SECRET)
}

func ScopeInfoOfDbMeta(meta *dbmodel.AssetMeta) *scope.ScopeInfo {
	return scope.ScopeInfoOf(scope.Scope(meta.Scope), meta.ProjectId, meta.DepartmentId, meta.ClusterId, meta.TenantId)
}
