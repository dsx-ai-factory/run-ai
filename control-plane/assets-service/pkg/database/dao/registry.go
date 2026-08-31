/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/encryption"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/query_types"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

func CreateRegistry(ctx context.Context, request api.RegistryCreationRequest, keyId int32, nameValidationQuery *database.ParametrizedQuery) (*api.Registry, *api.Error) {
	id := uuid.New().String()

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	assetMeta, createErr := createAssetMeta(ctx, tx, id, api.ASSETKIND_REGISTRY, request.Meta, nameValidationQuery)
	if createErr != nil {
		return nil, database.Error(database.CREATE, api.ASSETKIND_REGISTRY).WithDetails(createErr)
	}

	passwordRequest := api.PasswordCreationRequest{
		Meta: request.Meta,
		Spec: api.PasswordCreationSpec{
			ExistingSecretName: nil,
			User:               &request.Spec.User,
			Password:           &request.Spec.Password,
		},
	}

	passwordRecord, err := CreatePasswordWithTransaction(ctx, tx, passwordRequest, keyId, true, nil)
	if err != nil {
		return nil, err
	}

	registryRecord := dbmodel.RegistryRecord{
		AssetId:        id,
		Url:            request.Spec.Url,
		CredentialKind: request.Spec.CredentialKind,
		CredentialId:   passwordRecord.Meta.Id,
	}

	_, regErr := tx.Create(registryRecord)
	if regErr != nil {
		return nil, database.Error(database.CREATE, api.ASSETKIND_REGISTRY).WithDetails(regErr)
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_CREATIONS, api.ASSETKIND_REGISTRY).WithDetails(err)
	}
	committed = true

	response := api.Registry{
		Meta: *assetMeta,
		Spec: api.RegistrySpec{
			Url:            registryRecord.Url,
			CredentialKind: registryRecord.CredentialKind,
			User:           *passwordRecord.Spec.User,
			Password:       *passwordRecord.Spec.Password,
		},
	}
	return &response, nil
}

func GetRegistryById(ctx context.Context, id string) (*api.Registry, *api.Error) {
	registryRecord, err := getRegistryFullRecord(ctx, runaidb.DB().WithContext(ctx), id)
	if err != nil {
		return nil, err
	}

	passwordRecord, err := GetPasswordById(ctx, registryRecord.CredentialId, CredentialsOptions{IncludePrivateUsePasswords: true})
	if err != nil {
		return nil, err
	}

	response := api.Registry{
		Meta: registryRecord.AssetMeta,
		Spec: api.RegistrySpec{
			Url:            registryRecord.Url,
			CredentialKind: registryRecord.CredentialKind,
			User:           *passwordRecord.Spec.User,
			Password:       *passwordRecord.Spec.Password,
		},
	}
	return &response, nil
}

func ListRegistries(ctx context.Context, queryParams *query_types.QueryParams) (*api.RegistryListResponse, *api.Error) {
	combinedList := make([]dbmodel.RegistryFullRecord, 0)

	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_REGISTRY,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, queryParams, dbtables.TableNameAssetsRegistry)
	if err != nil {
		return nil, err
	}

	_, scanErr := result.
		Order(&query.AssetMeta.Id).
		Scan(&combinedList)
	if scanErr != nil {
		return nil, database.Error(database.GET, api.ASSETKIND_REGISTRY).WithDetails(scanErr)
	}

	response := api.RegistryListResponse{
		Entries: make([]api.Registry, 0),
	}

	for _, combinedRecord := range combinedList {
		passwordRecord, err := GetPasswordById(ctx, combinedRecord.CredentialId, CredentialsOptions{IncludePrivateUsePasswords: true})
		if err != nil {
			return nil, err
		}

		response.Entries = append(response.Entries, api.Registry{
			Meta: combinedRecord.AssetMeta,
			Spec: api.RegistrySpec{
				Url:            combinedRecord.Url,
				CredentialKind: combinedRecord.CredentialKind,
				User:           *passwordRecord.Spec.User,
				Password:       *passwordRecord.Spec.Password,
			},
		})
	}
	return &response, nil
}

func UpdateRegistry(ctx context.Context, id string, request api.RegistryUpdateRequest, keyId int32, nameValidationQuery *database.ParametrizedQuery) (*api.Registry, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	assetMeta, rowsAffected, apiErr, dbErr := updateAssetMeta(ctx, tx, id, request.Meta, nameValidationQuery, nil)
	if apiErr != nil {
		return nil, apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return nil, database.Error(database.UPDATE, api.ASSETKIND_REGISTRY).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	recId := dbmodel.RegistryRecord{AssetId: id}
	registryUpdate := dbmodel.RegistryUpdate{}

	if err := tx.Model(recId).Where(recId).First(&registryUpdate); err != nil {
		return nil, database.Error(database.UPDATE, api.ASSETKIND_REGISTRY).WithDetails(err)
	}

	passwordOptions := CredentialsOptions{IncludePrivateUsePasswords: true}

	/* At this point we are not supporting spec update
	       passwordRequest := api.PasswordUpdateRequest{
	   		Meta: request.Meta,
	   		Spec: &api.PasswordUpdateSpec{
	   			ExistingSecretName: nil,
	   			User:               &request.Spec.User,
	   			Password:           &request.Spec.Password,
	   		},
	   	}
	   	passwordRecord, err := updatePasswordInTransaction(ctx, tx, registryUpdate.CredentialId, passwordRequest, keyId, passwordOptions)
	   	if err != nil {
	   		return nil, err
	   	}

	   	registryUpdate.Url = request.Spec.Url
	   	registryUpdate.CredentialKind = request.Spec.CredentialKind
	   	registryUpdate.CredentialId = passwordRecord.Meta.Id

	   	result = tx.Where(recId).Save(&registryUpdate)
	   	if result.Error != nil {
	   		return nil, database.Error(database.UPDATE, api.ASSETKIND_REGISTRY).WithDetailsFromDb(result)
	   	}
	*/

	passwordRecord, err := GetPasswordById(ctx, registryUpdate.CredentialId, passwordOptions)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_REGISTRY).WithDetails(err)
	}
	committed = true

	response := api.Registry{
		Meta: assetMeta.ToApiAssetMeta(),
		Spec: api.RegistrySpec{
			Url:            registryUpdate.Url,
			User:           *passwordRecord.Spec.User,
			CredentialKind: registryUpdate.CredentialKind,
			Password:       *passwordRecord.Spec.Password,
		},
	}
	return &response, nil
}

func DeleteRegistry(ctx context.Context, id string, deletionType DeletionType) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	registryRecord, err := getRegistryFullRecord(ctx, tx, id)
	if err != nil {
		return err
	}

	rowsAffected, apiErr, dbErr := deleteAssetMeta(ctx, tx, id, deletionType)
	if apiErr != nil {
		return apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_REGISTRY).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	if deletionType == SoftDelete {
		return nil
	}

	registryId := dbmodel.RegistryRecord{AssetId: id}

	delRows, delErr := tx.Where(&registryId).Delete(&registryId)
	if delErr != nil || delRows == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_REGISTRY).WithDetailsFromDbResult(delRows, delErr)
	}

	// no need to modify asset deleted because this is an internal password asset used by the registry itself
	if err := deletePasswordInTransaction(ctx, tx, registryRecord.CredentialId, deletionType, CredentialsOptions{IncludePrivateUsePasswords: true}); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return database.Error(database.COMMIT_DELETIONS, api.ASSETKIND_REGISTRY).WithDetails(err)
	}
	committed = true

	return nil
}

func getRegistryFullRecord(ctx context.Context, tx orm.DB, id string) (*dbmodel.RegistryFullRecord, *api.Error) {
	var combinedRecord dbmodel.RegistryFullRecord

	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Id:   id,
			Kind: api.ASSETKIND_REGISTRY,
		},
	}

	result, err := GetAssetMeta(ctx, tx, query, nil, dbtables.TableNameAssetsRegistry)
	if err != nil {
		return nil, err
	}

	scanRows, scanErr := result.Scan(&combinedRecord)
	if scanErr != nil || scanRows == 0 {
		return nil, database.Error(database.GET, api.ASSETKIND_REGISTRY).WithDetailsFromDbResult(scanRows, scanErr)
	}

	response := dbmodel.RegistryFullRecord{
		AssetMeta:      combinedRecord.AssetMeta,
		Url:            combinedRecord.Url,
		CredentialKind: combinedRecord.CredentialKind,
		CredentialId:   combinedRecord.CredentialId,
	}
	return &response, nil
}

// GetRegistryVisiblePassword Returns an api.Registry with password unmasked but encrypted.
// This is for internal use only and should never be sent in an API response.
func GetRegistryVisiblePassword(ctx context.Context, id string) (*api.Registry, *api.Error) {
	registryRecord, err := getRegistryFullRecord(ctx, runaidb.DB().WithContext(ctx), id)
	if err != nil {
		return nil, err
	}

	passwordRecord, err := GetPasswordByIdVisible(ctx, registryRecord.CredentialId)
	if err != nil {
		return nil, err
	}

	tenantKeys, keysErr := encryption.GetTenantEncryptionKeys(ctx)
	if keysErr != nil {
		return nil, keysErr
	}

	password, decryptErr := encryption.Decrypt(ctx, tenantKeys, passwordRecord.Spec.KeyId, passwordRecord.Spec.Password, string(api.ASSETKIND_REGISTRY))
	if decryptErr != nil {
		return nil, decryptErr
	}

	return &api.Registry{
		Meta: registryRecord.AssetMeta,
		Spec: api.RegistrySpec{
			Url:            registryRecord.Url,
			CredentialKind: registryRecord.CredentialKind,
			User:           *passwordRecord.Spec.User,
			Password:       password,
		},
	}, nil
}
