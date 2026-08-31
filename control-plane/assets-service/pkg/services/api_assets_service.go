/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	uuid2 "github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/flat_workloads"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/query_types"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
)

var dummyWorkload = uuid2.NewString()

type AssetsAPIService struct{}

func NewAssetsAPIService() *AssetsAPIService {
	return &AssetsAPIService{}
}

func (s *SyncApiService) ListAssetsById(ctx context.Context, idsAsString string) (api.ImplResponse, error) {
	if idsAsString == "" {
		return NewErrorResponse(api.NewValidationError("required comma separated list of ids is missing")), nil
	}

	ids := strings.Split(idsAsString, ",")

	for index, id := range ids {
		uuid, err := uuid2.Parse(id)
		if err != nil {
			return NewErrorResponse(api.NewValidationError(fmt.Sprintf("malformed uuid %s: %v", id, err))), nil
		}
		ids[index] = uuid.String()
	}

	idsToMeta, err := dao.ListAssetsMetaByIds(ctx, ids, false /*no need temporary deleted assets*/, nil)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	response := api.AssetListByIdsResponse{
		Entries: make([]api.AssetMeta, 0, len(idsToMeta)),
	}

	for _, meta := range idsToMeta {
		response.Entries = append(response.Entries, meta.AssetMeta)
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *AssetsAPIService) RequestAssetFlatFormat(ctx context.Context, assetFlatRequest api.AssetFlatRequest) (api.ImplResponse, error) {
	resp, err := s.AssetFlatFormatImpl(ctx, assetFlatRequest)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}
	return api.Response(http.StatusOK, resp), nil
}

func (s *AssetsAPIService) AssetFlatFormatImpl(ctx context.Context, assetFlatRequest api.AssetFlatRequest) (*api.AssetFlatResponse, *api.Error) {
	assetIds := make([]string, 0)

	if assetFlatRequest.AssetId != "" {
		assetIds = append(assetIds, assetFlatRequest.AssetId)
	}

	if assetFlatRequest.AssetIds != nil {
		assetIds = append(assetIds, assetFlatRequest.AssetIds...)
	}

	if len(assetIds) == 0 {
		return nil, api.NewError(http.StatusBadRequest, "mandatory assetId is not provided", nil)
	}

	var projectIdInt32 int32
	if assetFlatRequest.ProjectId != "" {
		projectIdInt, castingErr := strconv.Atoi(assetFlatRequest.ProjectId)
		if castingErr != nil {
			return nil, api.NewError(http.StatusBadRequest, "invalid project ID", castingErr)
		}
		projectIdInt32 = int32(projectIdInt)
	}

	// filter only by the asset ids
	assets := strings.Join(assetIds, ",")
	queryParams, queryError := query_types.NewQueryParams(assets, "", "", false, 0, "", "", nil)
	if queryError != nil {
		return nil, api.NewError(int(queryError.Code), queryError.Message, errors.New(queryError.Details))
	}

	// make sure to reduce the list by rbac
	_, authError := query_types.ApplyRbacRules(ctx, queryParams)
	if authError != nil {
		return nil, api.NewError(int(authError.Code), authError.Message, errors.New(authError.Details))
	}

	assetsMeta, getError := dao.ListContainedAssetsByIds(ctx, assetIds, queryParams)
	if getError != nil {
		return nil, getError
	}

	//
	//   sanity validation on the provided assets
	//
	response := api.AssetFlatResponse{
		AssetsMeta: make([]api.AssetMeta, 0, len(assetIds)),
	}

	allIds := datastructs.NewSet[string]()
	templateFound := false

	for _, assetId := range assetIds {
		assetMeta, ok := assetsMeta[assetId]
		if !ok {
			return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("assetId %s not found", assetId), nil)
		}

		response.AssetsMeta = append(response.AssetsMeta, assetMeta.AssetMeta)

		//
		//  if template is provided, it has to be the only asset specified in the API payload
		//  we allow multiple assets OR template, not both in the same API
		//
		if assetMeta.Kind == api.ASSETKIND_WORKLOAD_TEMPLATE {
			if len(assetsMeta) != 1 {
				return nil, api.NewError(http.StatusBadRequest, "multiple assetIds cannot be provided when assetId of a workload template is specified", nil)
			}
			templateFound = true
		}

		if allIds.Has(assetMeta.Id) {
			return nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("assetId %s provided more than once", assetMeta.Id), nil)
		}

		allIds.Add(assetMeta.Id)
	}

	worker := CreateFlatWorkloadReplicaInfo{}
	var master *CreateFlatWorkloadReplicaInfo

	//
	//   for template, we need the list of assets comprising the template. for single asset
	//   we just need to convert the single asset to flat format
	//
	if templateFound {
		template, containedAssetIds, err := dao.GetTemplateAssetById(ctx, assetIds[0])
		if err != nil {
			return nil, err
		}
		worker.containedAssetIds = containedAssetIds
		worker.specificRunParams = template.Spec.SpecificEnv
	} else {
		worker.containedAssetIds = assetIds
	}

	switch assetFlatRequest.WorkloadType {
	case api.WORKLOADTYPE_DISTRIBUTED:
		master = &worker
	case api.WORKLOADTYPE_DISTRIBUTED_INFERENCE:
		// For Distributed Inference, we need to populate both leader and worker
		// Create a copy for the leader
		leader := worker
		master = &leader
	}

	//
	//   use the create workload function for translating to the flat format. the create workload
	//   does all the translation for both worker and optional master
	//
	asFlat, err := translateFlatWorkload(ctx, string(assetFlatRequest.WorkloadType), &worker, master, &flat_workloads.TranslateFlatWorkloadOptions{
		DisableComplimentaryInfo: true,
	})
	if err != nil {
		return nil, api.NewError(int(err.Code), err.Message, nil)
	}

	//
	//   now compose the response based on the returned flat response
	//
	jsonErr := json.Unmarshal(asFlat.WorkerPayloadAsBytes, &response.Worker)
	if jsonErr != nil {
		return nil, api.NewError(http.StatusNotImplemented, jsonErr.Error(), nil)
	}

	if asFlat.MasterPayloadAsBytes != nil {
		// For Distributed Training, populate master
		response.Master = &api.MasterSpec{}

		jsonErr := json.Unmarshal(asFlat.MasterPayloadAsBytes, &response.Master)
		if jsonErr != nil {
			return nil, api.NewError(http.StatusNotImplemented, jsonErr.Error(), nil)
		}
	}

	//
	//   update last-used table for this asset, because this API is invoked whenever an asset is used by the user
	//   in the flat-based submission forms. in case of error do logging only, do not fail the API as a whole
	//
	if assetFlatRequest.ProjectId != "" {
		lastUsedErr := dao.UpsertAssetLastUsedByWorkload(ctx, assetFlatRequest.AssetId, string(assetFlatRequest.WorkloadType), dummyWorkload, projectIdInt32, time.Now())
		if lastUsedErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update last used of asset %s for workload type %s: %v",
				assetFlatRequest.AssetId, assetFlatRequest.WorkloadType, lastUsedErr.Message)
		}
	}

	return &response, nil
}
