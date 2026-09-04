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
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/encryption"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/utils"
	"github.com/google/uuid"
)

func CreateAccessKey(ctx context.Context, request api.AccessKeyCreationRequest, keyId int32, nameValidationQuery *database.ParametrizedQuery) (*api.AccessKey, *api.Error) {
	id := uuid.New().String()

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	assetMeta, err := createAssetMeta(ctx, tx, id, api.ASSETKIND_ACCESS_KEY, request.Meta, nameValidationQuery)
	if err != nil {
		if strings.Contains(err.Error(), alreadyExistsError) {
			return nil, database.ErrorWithCode(database.CREATE, api.ASSETKIND_ACCESS_KEY, http.StatusConflict).WithDetails(err)
		}
		return nil, database.Error(database.CREATE, api.ASSETKIND_ACCESS_KEY).WithDetails(err)
	}

	accessKeyRecord := dbmodel.AccessKeyRecord{
		AssetId: id,
		AccessKeySpec: api.AccessKeySpec{
			ExistingSecretName: request.Spec.ExistingSecretName,
			AccessKeyId:        request.Spec.AccessKeyId,
			SecretAccessKey:    request.Spec.SecretAccessKey,
			KeyId:              keyId,
		},
	}

	_, err = tx.Create(accessKeyRecord)
	if err != nil {
		return nil, database.Error(database.CREATE, api.ASSETKIND_ACCESS_KEY).WithDetails(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_CREATIONS, api.ASSETKIND_ACCESS_KEY).WithDetails(err)
	}

	committed = true
	if utils.StringHasValue(accessKeyRecord.SecretAccessKey) {
		accessKeyRecord.SecretAccessKey = convutils.Ref(encryption.SecretMask)
	}

	return &api.AccessKey{
		Meta:        *assetMeta,
		Spec:        accessKeyRecord.AccessKeySpec,
		ClusterInfo: dbmodel.GetSecretClusterInfo(assetMeta, accessKeyRecord.ExistingSecretName),
	}, nil
}

func GetAccessKeyById(ctx context.Context, id string, options CredentialsOptions) (*api.AccessKey, *api.Error) {
	var combinedRecord dbmodel.AccessKeyFullRecord

	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_ACCESS_KEY,
			Id:   id,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, nil, dbtables.TableNameAssetsAccessKeys)
	if err != nil {
		return nil, err
	}

	if options.IncludeDeleted {
		result = result.Unscoped()
	}

	scanRows, scanErr := result.Scan(&combinedRecord)
	if scanErr != nil || scanRows == 0 {
		return nil, database.Error(database.GET, api.ASSETKIND_ACCESS_KEY).WithDetailsFromDbResult(scanRows, scanErr)
	}

	if utils.StringHasValue(combinedRecord.SecretAccessKey) {
		combinedRecord.SecretAccessKey = convutils.Ref(encryption.SecretMask)
	}

	response := api.AccessKey{
		Meta:        combinedRecord.AssetMeta,
		Spec:        combinedRecord.AccessKeySpec,
		ClusterInfo: dbmodel.GetSecretClusterInfo(&combinedRecord.AssetMeta, combinedRecord.ExistingSecretName),
	}

	return &response, nil
}

func ListAccessKeys(ctx context.Context, queryParams *query_types.QueryParams, options CredentialsOptions) (*api.AccessKeyListResponse, *api.Error) {
	combinedList := make([]dbmodel.AccessKeyFullRecord, 0)

	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_ACCESS_KEY,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, queryParams, dbtables.TableNameAssetsAccessKeys)
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
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list access keys", scanErr)
	}

	response := api.AccessKeyListResponse{
		Entries: make([]api.AccessKey, 0),
	}

	for _, row := range combinedList {
		if options.MaskCredentials && utils.StringHasValue(row.SecretAccessKey) {
			row.SecretAccessKey = convutils.Ref(encryption.SecretMask)
		}

		response.Entries = append(response.Entries, api.AccessKey{
			Meta:        row.AssetMeta,
			Spec:        row.AccessKeySpec,
			ClusterInfo: dbmodel.GetSecretClusterInfo(&row.AssetMeta, row.ExistingSecretName),
		})
	}

	return &response, nil
}

func UpdateAccessKey(ctx context.Context, id string, request api.AccessKeyUpdateRequest, keyId int32) (*api.AccessKey, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	_, rowsAffected, apiErr, dbErr := updateCredentialAssetMeta(ctx, tx, id, request.Meta)
	if apiErr != nil {
		return nil, apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return nil, database.Error(database.UPDATE, api.ASSETKIND_ACCESS_KEY).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	//
	//   for credentials, it is possible that spec is nil, and in this case it will only rename without changing the spec
	//
	/* t this point we are not supporting spec update -> request.Spec == nil */
	{
		if err := tx.Commit(); err != nil {
			return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_ACCESS_KEY).WithDetails(err)
		}
		committed = true

		// no need to notify about asset modified because the spec remained the same

		return GetAccessKeyById(ctx, id, CredentialsOptions{})
	}

	/* At this point we are not supporting spec update
	   recId := dbmodel.AccessKeyRecord{AssetId: id}
	   accessKeyUpdate := dbmodel.AccessKeyUpdate{}

	   result = tx.Model(recId).Where(recId).First(&accessKeyUpdate)
	   if result.Error != nil || result.RowsAffected == 0 {
	   	return nil, database.Error(database.UPDATE, api.ASSETKIND_ACCESS_KEY).WithDetailsFromDb(result)
	   }

	   accessKeyUpdate.AccessKeySpec = api.AccessKeySpec{
	   	AccessKeyId:     request.Spec.AccessKeyId,
	   	SecretAccessKey: request.Spec.SecretAccessKey,
	   	KeyId:           keyId,
	   }
	   result = tx.Where(recId).Save(&accessKeyUpdate)
	   if result.Error != nil {
	   	return nil, database.Error(database.UPDATE, api.ASSETKIND_ACCESS_KEY).WithDetailsFromDb(result)
	   }

	   result = tx.Commit()
	   if result.Error != nil {
	   	return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_ACCESS_KEY).WithDetailsFromDb(result)
	   }

	   committed = true
	   if utils.StringHasValue(accessKeyUpdate.SecretAccessKey) {
	   	accessKeyUpdate.SecretAccessKey = convutils.Ref(encryption.SecretMask)
	   }

	   return &api.AccessKey{
	   	Meta: assetMeta.ToApiAssetMeta(),
	   	Spec: accessKeyUpdate.AccessKeySpec,
	   }, nil
	*/
	// return nil, nil
}

func DeleteAccessKey(ctx context.Context, assetMeta *api.AssetMeta, deletionType DeletionType) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	inuse, err := checkAccessKeyInUse(ctx, tx, assetMeta.Id)
	if err != nil {
		return err
	}
	if inuse {
		return api.NewError(http.StatusPreconditionFailed, "Cannot delete access key since it is inuse", nil)
	}

	rowsAffected, apiErr, dbErr := deleteAssetMeta(ctx, tx, assetMeta.Id, deletionType)
	if apiErr != nil {
		return apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_ACCESS_KEY).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	if deletionType == HardDelete {
		recId := dbmodel.AccessKeyRecord{AssetId: assetMeta.Id}

		delRows, delErr := tx.Where(&recId).Delete(&recId)
		if delErr != nil || delRows == 0 {
			return database.Error(database.DELETE, api.ASSETKIND_ACCESS_KEY).WithDetailsFromDbResult(delRows, delErr)
		}
	}

	if err := tx.Commit(); err != nil {
		return database.Error(database.COMMIT_DELETIONS, api.ASSETKIND_ACCESS_KEY).WithDetails(err)
	}
	committed = true

	NotifyAssetDeleted(ctx, assetMeta.Id, assetMeta.Kind, ScopeInfoOfAssetMeta(assetMeta))

	return nil
}

func checkAccessKeyInUse(ctx context.Context, tx orm.DB, accessKeyId string) (bool, *api.Error) {
	// Access keys is used by S3 and Secret Volumes assets
	count, err := countS3AssetsUsingAccessKeyId(tx, accessKeyId)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	count, err = countSecretVolumeAssetsUsingCredentialId(tx, accessKeyId)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	count, err = countEnvironmentAssetsUsingCredentialAsEnvVar(ctx, accessKeyId)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	return false, nil
}
