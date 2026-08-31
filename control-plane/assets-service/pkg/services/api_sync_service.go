/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	"github.com/run-ai/runai/runai-common-packages/logging"

	"github.com/run-ai/runai/runai-application-packages/convutils"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/query_types"
)

type SyncApiService struct{}

func NewSyncApiService() *SyncApiService {
	return &SyncApiService{}
}

func (s *SyncApiService) SyncValidate(ctx context.Context, validateSyncRequest api.ValidateSyncRequest) (api.ImplResponse, error) {
	switch validateSyncRequest.AssetKind {
	case api.ASSETKIND_PVC:
		return handlePvcValidateSync(ctx, validateSyncRequest)
	case api.ASSETKIND_GENERIC_SECRET, api.ASSETKIND_ACCESS_KEY, api.ASSETKIND_PASSWORD, api.ASSETKIND_DOCKER_REGISTRY:
		return handleSecretValidateSync(ctx, validateSyncRequest)
	default:
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("Invalid assetKind: %s", validateSyncRequest.AssetKind), nil)), nil
	}
}

func (s *SyncApiService) SyncAssetsResourcesIds(ctx context.Context, assetResourcesIdsSyncRequest api.AssetResourcesIdsSyncRequest) (api.ImplResponse, error) {
	tenantId := auth.TenantIdOf(ctx)

	dbModels := make([]dbmodel.AssetResourceIdsRecord, 0)
	for _, assetResourceId := range assetResourcesIdsSyncRequest.Items {
		dbModels = append(dbModels, dbmodel.AssetResourceIdsRecord{
			AssetId:      assetResourceId.AssetId,
			ResourceId:   assetResourceId.ResourceId,
			TenantId:     tenantId,
			LastSyncedAt: time.Now(),
		})
	}

	err := dao.UpsertAssetResourceIds(ctx, dbModels)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, nil), nil
}

func handlePvcValidateSync(ctx context.Context, validateSyncRequest api.ValidateSyncRequest) (api.ImplResponse, error) {
	projectIdInt, err := strconv.Atoi(validateSyncRequest.ProjectId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, err.Error(), err)), nil
	}

	queryParams, queryError := query_types.NewQueryParams("", "", "project", false, int32(projectIdInt), "", validateSyncRequest.ClusterId, nil)
	if queryError != nil {
		return NewErrorResponse(queryError), nil
	}

	pvcAssets, apiErr := dao.ListPvcAssetsByClaimNames(ctx, queryParams, validateSyncRequest.AssetNames)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	assetClaimNamesSet := datastructs.NewSet(validateSyncRequest.AssetNames...)

	for _, asset := range pvcAssets {
		if asset.Spec.ClaimName == nil {
			continue
		}
		if assetClaimNamesSet.Has(*asset.Spec.ClaimName) {
			logging.Infof("found in DB %s asset with claim name '%s' that was marked for deletion in project id '%v'", api.ASSETKIND_PVC, *asset.Spec.ClaimName, validateSyncRequest.ProjectId)
			assetClaimNamesSet.Delete(*asset.Spec.ClaimName)
		}
	}
	validatedNames := make([]string, 0)
	for asset := range assetClaimNamesSet {
		validatedNames = append(validatedNames, asset)
	}

	response := &api.ValidateSyncResponse{
		ValidatedNames: validatedNames,
	}

	return api.Response(http.StatusOK, response), nil
}

func handleSecretValidateSync(ctx context.Context, validateSyncRequest api.ValidateSyncRequest) (api.ImplResponse, error) {
	tenantId := auth.TenantIdOf(ctx)

	projectIdInt, err := strconv.Atoi(validateSyncRequest.ProjectId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, err.Error(), err)), nil
	}

	basicQuery := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind:      validateSyncRequest.AssetKind,
			TenantId:  convutils.Ref(tenantId),
			ClusterId: convutils.Ref(validateSyncRequest.ClusterId),
			ProjectId: convutils.Ref(int32(projectIdInt)),
			Scope:     projectScope,
		},
	}

	metadataRecords, apiErr := dao.ListAssetMeta(ctx, basicQuery)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	assetsSet := datastructs.NewSet(validateSyncRequest.AssetNames...)

	for _, assetMeta := range *metadataRecords {
		if assetsSet.Has(assetMeta.Name) {
			logging.Infof("found in DB %s asset with name '%s' that was marked for deletion in project id '%v'", validateSyncRequest.AssetKind, assetMeta.Name, validateSyncRequest.ProjectId)
			assetsSet.Delete(assetMeta.Name)
		}
	}
	validatedNames := make([]string, 0)
	for asset := range assetsSet {
		validatedNames = append(validatedNames, asset)
	}

	response := &api.ValidateSyncResponse{
		ValidatedNames: validatedNames,
	}

	return api.Response(http.StatusOK, response), nil
}
