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
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"

	"github.com/run-ai/runai/runai-application-packages/convutils"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/encryption"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/query_types"
)

func CreateGenericSecret(ctx context.Context, request api.GenericSecretCreationRequest, keyId int32, nameValidationQuery *database.ParametrizedQuery) (*api.GenericSecret, *api.Error) {
	id := uuid.New().String()

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	assetMeta, err := createAssetMeta(ctx, tx, id, api.ASSETKIND_GENERIC_SECRET, request.Meta, nameValidationQuery)
	if err != nil {
		if strings.Contains(err.Error(), alreadyExistsError) {
			return nil, database.ErrorWithCode(database.CREATE, api.ASSETKIND_GENERIC_SECRET, http.StatusConflict).WithDetails(err)
		}
		return nil, database.Error(database.CREATE, api.ASSETKIND_GENERIC_SECRET).WithDetails(err)
	}

	genericSecretRecord := dbmodel.GenericSecretRecord{
		AssetId: id,
		GenericSecretSpec: api.GenericSecretSpec{
			ExistingSecretName: request.Spec.ExistingSecretName,
			KeyValuePairs:      request.Spec.KeyValuePairs,
			KeyId:              keyId,
		},
	}

	_, err = tx.Create(genericSecretRecord)
	if err != nil {
		return nil, database.Error(database.CREATE, api.ASSETKIND_GENERIC_SECRET).WithDetails(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_CREATIONS, api.ASSETKIND_GENERIC_SECRET).WithDetails(err)
	}

	committed = true
	if genericSecretRecord.KeyValuePairs != nil {
		for i := range *genericSecretRecord.KeyValuePairs {
			(*genericSecretRecord.GenericSecretSpec.KeyValuePairs)[i].Key = convutils.Ref(encryption.SecretMask)
			(*genericSecretRecord.GenericSecretSpec.KeyValuePairs)[i].Value = convutils.Ref(encryption.SecretMask)
		}
	}

	return &api.GenericSecret{
		Meta:        *assetMeta,
		Spec:        genericSecretRecord.GenericSecretSpec,
		ClusterInfo: dbmodel.GetSecretClusterInfo(assetMeta, genericSecretRecord.ExistingSecretName),
	}, nil
}

func GetGenericSecretById(ctx context.Context, id string, options CredentialsOptions) (*api.GenericSecret, *api.Error) {
	var combinedRecord dbmodel.GenericSecretFullRecord

	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_GENERIC_SECRET,
			Id:   id,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, nil, dbtables.TableNameAssetsGenericSecret)
	if err != nil {
		return nil, err
	}
	// Either it's a non-private or owned by the subject
	result.Where(fmt.Sprintf("COALESCE(%v.%v, FALSE) = false OR %s.%s = ?",
		dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnPrivate,
		dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnCreatedBy), auth.SubjectIdOf(ctx))

	if options.IncludeDeleted {
		result = result.Unscoped()
	}

	scanRows, scanErr := result.Scan(&combinedRecord)
	if scanErr != nil || scanRows == 0 {
		return nil, database.Error(database.GET, api.ASSETKIND_GENERIC_SECRET).WithDetailsFromDbResult(scanRows, scanErr)
	}

	if combinedRecord.KeyValuePairs != nil {
		for i := range *combinedRecord.KeyValuePairs {
			(*combinedRecord.GenericSecretSpec.KeyValuePairs)[i].Key = convutils.Ref(encryption.SecretMask)
			(*combinedRecord.GenericSecretSpec.KeyValuePairs)[i].Value = convutils.Ref(encryption.SecretMask)
		}
	}

	apiAssetMeta := combinedRecord.ToApiAssetMeta()

	response := api.GenericSecret{
		Meta:        apiAssetMeta,
		Spec:        combinedRecord.GenericSecretSpec,
		ClusterInfo: dbmodel.GetSecretClusterInfo(&apiAssetMeta, combinedRecord.ExistingSecretName),
	}

	return &response, nil
}

func ListGenericSecrets(ctx context.Context, queryParams *query_types.QueryParams, options CredentialsOptions) (*api.GenericSecretListResponse, *api.Error) {
	combinedList := make([]dbmodel.GenericSecretFullRecord, 0)

	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_GENERIC_SECRET,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, queryParams, dbtables.TableNameAssetsGenericSecret)
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
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list generic secrets", scanErr)
	}

	response := api.GenericSecretListResponse{
		Entries: make([]api.GenericSecret, 0),
	}

	for _, row := range combinedList {
		if options.MaskCredentials && row.KeyValuePairs != nil {
			for i := range *row.KeyValuePairs {
				(*row.GenericSecretSpec.KeyValuePairs)[i].Key = convutils.Ref(encryption.SecretMask)
				(*row.GenericSecretSpec.KeyValuePairs)[i].Value = convutils.Ref(encryption.SecretMask)
			}
		}

		apiAssetMeta := row.ToApiAssetMeta()

		response.Entries = append(response.Entries, api.GenericSecret{
			Meta:        apiAssetMeta,
			Spec:        row.GenericSecretSpec,
			ClusterInfo: dbmodel.GetSecretClusterInfo(&apiAssetMeta, row.ExistingSecretName),
		})
	}

	return &response, nil
}

func UpdateGenericSecret(ctx context.Context, id string, request api.GenericSecretUpdateRequest, keyId int32) (*api.GenericSecret, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	_, rowsAffected, apiErr, dbErr := updateCredentialAssetMeta(ctx, tx, id, request.Meta)
	if apiErr != nil {
		return nil, apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return nil, database.Error(database.UPDATE, api.ASSETKIND_GENERIC_SECRET).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	//
	//   for credentials, it is possible that spec is nil, and in this case it will only rename without changing the spec
	//
	if true /* RUN-20970: At this point we are not supporting spec update -> request.Spec == nil */ {
		if err := tx.Commit(); err != nil {
			return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_GENERIC_SECRET).WithDetails(err)
		}
		committed = true

		// no need to notify asset modified because spec remained unchanged

		return GetGenericSecretById(ctx, id, CredentialsOptions{})
	}

	/* At this point we are not supporting spec update
	   recId := dbmodel.GenericSecretRecord{AssetId: id}
	   genericSecretUpdate := dbmodel.GenericSecretUpdate{}

	   result = tx.Model(recId).Where(recId).First(&genericSecretUpdate)
	   if result.Error != nil || result.RowsAffected == 0 {
	   	return nil, database.Error(database.UPDATE, api.ASSETKIND_GENERIC_SECRET).WithDetailsFromDb(result)
	   }

	   genericSecretUpdate.GenericSecretSpec = api.GenericSecretSpec{
	   	KeyValuePairs: request.Spec.KeyValuePairs,
	   	KeyId:         keyId,
	   }

	   result = tx.Where(recId).Save(&genericSecretUpdate)
	   if result.Error != nil {
	   	return nil, database.Error(database.UPDATE, api.ASSETKIND_GENERIC_SECRET).WithDetailsFromDb(result)
	   }

	   result = tx.Commit()
	   if result.Error != nil {
	   	return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_GENERIC_SECRET).WithDetailsFromDb(result)
	   }

	   committed = true

	   if genericSecretUpdate.GenericSecretSpec.KeyValuePairs != nil {
	   	for i := range *genericSecretUpdate.GenericSecretSpec.KeyValuePairs {
	   		(*genericSecretUpdate.GenericSecretSpec.KeyValuePairs)[i].Key = convutils.Ref(encryption.SecretMask)
	   		(*genericSecretUpdate.GenericSecretSpec.KeyValuePairs)[i].Value = convutils.Ref(encryption.SecretMask)
	   	}
	   }

	   return &api.GenericSecret{
	   	Meta: assetMeta.ToApiAssetMeta(),
	   	Spec: genericSecretUpdate.GenericSecretSpec,
	   }, nil
	*/
	return nil, nil
}

func DeleteGenericSecret(ctx context.Context, assetMeta *api.AssetMeta, deletionType DeletionType) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	inuse, err := checkGenericSecretInUse(ctx, tx, assetMeta.Id)
	if err != nil {
		return err
	}
	if inuse {
		return api.NewError(http.StatusPreconditionFailed, "Cannot delete generic secret since it is being used by other resources", nil)
	}

	rowsAffected, apiErr, dbErr := deleteAssetMeta(ctx, tx, assetMeta.Id, deletionType)
	if apiErr != nil {
		return apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_GENERIC_SECRET).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	if deletionType == HardDelete {
		recId := dbmodel.GenericSecretRecord{AssetId: assetMeta.Id}

		delRows, delErr := tx.Where(&recId).Delete(&recId)
		if delErr != nil || delRows == 0 {
			return database.Error(database.DELETE, api.ASSETKIND_GENERIC_SECRET).WithDetailsFromDbResult(delRows, delErr)
		}
	}

	if err := tx.Commit(); err != nil {
		return database.Error(database.COMMIT_DELETIONS, api.ASSETKIND_GENERIC_SECRET).WithDetails(err)
	}
	committed = true

	NotifyAssetDeleted(ctx, assetMeta.Id, assetMeta.Kind, ScopeInfoOfAssetMeta(assetMeta))

	return nil
}

func checkGenericSecretInUse(ctx context.Context, tx orm.DB, Id string) (bool, *api.Error) {
	// Generic Secret is used by Secret Volume
	count, err := countSecretVolumeAssetsUsingCredentialId(tx, Id)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	count, err = countEnvironmentAssetsUsingCredentialAsEnvVar(ctx, Id)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	count, err = countGitAssetsUsingGenericSecret(tx, Id)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	return false, nil
}
