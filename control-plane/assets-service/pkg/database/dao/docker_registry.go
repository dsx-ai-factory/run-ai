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
	"github.com/run-ai/runai/control-plane/assets-service/pkg/utils"
)

func CreateDockerRegistry(ctx context.Context, request api.DockerRegistryCreationRequest, keyId int32, nameValidationQuery *database.ParametrizedQuery) (*api.DockerRegistry, *api.Error) {
	id := uuid.New().String()

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	assetMeta, err := createAssetMeta(ctx, tx, id, api.ASSETKIND_DOCKER_REGISTRY, request.Meta, nameValidationQuery)
	if err != nil {
		if strings.Contains(err.Error(), alreadyExistsError) {
			return nil, database.ErrorWithCode(database.CREATE, api.ASSETKIND_DOCKER_REGISTRY, http.StatusConflict).WithDetails(err)
		}
		return nil, database.Error(database.CREATE, api.ASSETKIND_DOCKER_REGISTRY).WithDetails(err)
	}

	dockerRegistryRecord := dbmodel.DockerRegistryRecord{
		AssetId: id,
		DockerRegistrySpec: api.DockerRegistrySpec{
			ExistingSecretName: request.Spec.ExistingSecretName,
			User:               request.Spec.User,
			Password:           request.Spec.Password,
			Url:                request.Spec.Url,
			KeyId:              keyId,
		},
	}

	_, err = tx.Create(dockerRegistryRecord)
	if err != nil {
		return nil, database.Error(database.CREATE, api.ASSETKIND_DOCKER_REGISTRY).WithDetails(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_CREATIONS, api.ASSETKIND_DOCKER_REGISTRY).WithDetails(err)
	}
	committed = true

	if utils.StringHasValue(dockerRegistryRecord.Password) {
		dockerRegistryRecord.Password = convutils.Ref(encryption.SecretMask)
	}

	return &api.DockerRegistry{
		Meta:        *assetMeta,
		Spec:        dockerRegistryRecord.DockerRegistrySpec,
		ClusterInfo: dbmodel.GetSecretClusterInfo(assetMeta, dockerRegistryRecord.ExistingSecretName),
	}, nil
}

func GetDockerRegistryById(ctx context.Context, id string, options CredentialsOptions) (*api.DockerRegistry, *api.Error) {
	var combinedRecord dbmodel.DockerRegistryFullRecord

	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_DOCKER_REGISTRY,
			Id:   id,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, nil, dbtables.TableNameAssetsDockerRegistry)
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
		return nil, database.Error(database.GET, api.ASSETKIND_DOCKER_REGISTRY).WithDetailsFromDbResult(scanRows, scanErr)
	}

	if utils.StringHasValue(combinedRecord.Password) {
		combinedRecord.Password = convutils.Ref(encryption.SecretMask)
	}

	response := api.DockerRegistry{
		Meta:        combinedRecord.AssetMeta,
		Spec:        combinedRecord.DockerRegistrySpec,
		ClusterInfo: dbmodel.GetSecretClusterInfo(&combinedRecord.AssetMeta, combinedRecord.ExistingSecretName),
	}

	return &response, nil
}

func ListDockerRegistries(ctx context.Context, queryParams *query_types.QueryParams, options CredentialsOptions) (*api.DockerRegistryListResponse, *api.Error) {
	combinedList := make([]dbmodel.DockerRegistryFullRecord, 0)

	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_DOCKER_REGISTRY,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, queryParams, dbtables.TableNameAssetsDockerRegistry)
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
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list docker registry credentials", scanErr)
	}

	response := api.DockerRegistryListResponse{
		Entries: make([]api.DockerRegistry, 0),
	}

	for _, row := range combinedList {
		if options.MaskCredentials && utils.StringHasValue(row.Password) {
			row.Password = convutils.Ref(encryption.SecretMask)
		}

		response.Entries = append(response.Entries, api.DockerRegistry{
			Meta:        row.AssetMeta,
			Spec:        row.DockerRegistrySpec,
			ClusterInfo: dbmodel.GetSecretClusterInfo(&row.AssetMeta, row.ExistingSecretName),
		})
	}

	return &response, nil
}

func UpdateDockerRegistry(ctx context.Context, id string, request api.DockerRegistryUpdateRequest, keyId int32) (*api.DockerRegistry, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	_, rowsAffected, apiErr, dbErr := updateCredentialAssetMeta(ctx, tx, id, request.Meta)
	if apiErr != nil {
		return nil, apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return nil, database.Error(database.UPDATE, api.ASSETKIND_DOCKER_REGISTRY).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	//
	//   for credentials, it is possible that spec is nil, and in this case it will only rename without changing the spec
	//
	if true /* RUN-20970: At this point we are not supporting spec update -> request.Spec == nil */ {
		if err := tx.Commit(); err != nil {
			return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_DOCKER_REGISTRY).WithDetails(err)
		}
		committed = true

		// no need to notify asset modify because there was no change of the spec

		return GetDockerRegistryById(ctx, id, CredentialsOptions{})
	}

	/* At this point we are not supporting spec update
	   recId := dbmodel.DockerRegistryRecord{AssetId: id}
	   dockerRegistryUpdate := dbmodel.DockerRegistryUpdate{}

	   result = tx.Model(recId).Where(recId).First(&dockerRegistryUpdate)
	   if result.Error != nil || result.RowsAffected == 0 {
	   	return nil, database.Error(database.UPDATE, api.ASSETKIND_DOCKER_REGISTRY).WithDetailsFromDb(result)
	   }

	   dockerRegistryUpdate.DockerRegistrySpec = api.DockerRegistrySpec{
	   	User:     request.Spec.User,
	   	Password: request.Spec.Password,
	   	Url:      request.Spec.Url,
	   	KeyId:    keyId,
	   }
	   result = tx.Where(recId).Save(&dockerRegistryUpdate)
	   if result.Error != nil {
	   	return nil, database.Error(database.UPDATE, api.ASSETKIND_DOCKER_REGISTRY).WithDetailsFromDb(result)
	   }

	   result = tx.Commit()
	   if result.Error != nil {
	   	return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_DOCKER_REGISTRY).WithDetailsFromDb(result)
	   }
	   committed = true

	   if utils.StringHasValue(dockerRegistryUpdate.Password) {
	   	dockerRegistryUpdate.Password = convutils.Ref(encryption.SecretMask)
	   }

	   return &api.DockerRegistry{
	   	Meta: assetMeta.ToApiAssetMeta(),
	   	Spec: dockerRegistryUpdate.DockerRegistrySpec,
	   }, nil
	*/
	return nil, nil
}

func DeleteDockerRegistry(ctx context.Context, assetMeta *api.AssetMeta, deletionType DeletionType) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	rowsAffected, apiErr, dbErr := deleteAssetMeta(ctx, tx, assetMeta.Id, deletionType)
	if apiErr != nil {
		return apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_DOCKER_REGISTRY).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	if deletionType == HardDelete {
		recId := dbmodel.DockerRegistryRecord{AssetId: assetMeta.Id}

		delRows, delErr := tx.Where(&recId).Delete(&recId)
		if delErr != nil || delRows == 0 {
			return database.Error(database.DELETE, api.ASSETKIND_DOCKER_REGISTRY).WithDetailsFromDbResult(delRows, delErr)
		}
	}

	if err := tx.Commit(); err != nil {
		return database.Error(database.COMMIT_DELETIONS, api.ASSETKIND_DOCKER_REGISTRY).WithDetails(err)
	}
	committed = true

	NotifyAssetDeleted(ctx, assetMeta.Id, assetMeta.Kind, ScopeInfoOfAssetMeta(assetMeta))

	return nil
}
