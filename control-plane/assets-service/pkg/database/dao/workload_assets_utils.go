/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
)

func assetIdsContainedInWorkload(assets *api.AssetsIds) ([]string, dbmodel.AssetKindToIdsMap, datastructs.Set[string]) {
	assetIds := make([]string, 0)
	assetKindToIds := make(dbmodel.AssetKindToIdsMap)
	workloadVolumeAssets := datastructs.NewSet[string]()

	assetIds = appendContainedAsset(assetIds, &assetKindToIds, &assets.Environment, api.ASSETKIND_ENVIRONMENT)
	assetIds = appendContainedAsset(assetIds, &assetKindToIds, assets.Compute, api.ASSETKIND_COMPUTE)

	if assets.Datasources != nil {
		for _, datasource := range assets.Datasources {
			assetIds = appendContainedAsset(assetIds, &assetKindToIds, &datasource.Id, datasource.Kind)
		}
	}

	if assets.WorkloadVolumes != nil {
		for _, assetId := range assets.WorkloadVolumes {
			assetIds = appendContainedAsset(assetIds, &assetKindToIds, &assetId, api.ASSETKIND_PVC)
			workloadVolumeAssets.Add(assetId)
		}
	}

	return assetIds, assetKindToIds, workloadVolumeAssets
}

func appendContainedAsset(allAssetIds []string, kindToIds *dbmodel.AssetKindToIdsMap, id *string, kind api.AssetKind) []string {
	if id != nil {
		allAssetIds = append(allAssetIds, *id)
		existingIds, ok := (*kindToIds)[kind]
		if ok {
			existingIds = append(existingIds, *id)
		} else {
			existingIds = []string{*id}
		}
		(*kindToIds)[kind] = existingIds
	}
	return allAssetIds
}

func appendContainedAssets(allAssetIds []string, kindToIds *dbmodel.AssetKindToIdsMap, ids []string, kind api.AssetKind) []string { //nolint:unused
	if ids != nil {
		allAssetIds = append(allAssetIds, ids...)
		existingIds, ok := (*kindToIds)[kind]
		if ok {
			existingIds = append(existingIds, ids...)
		} else {
			existingIds = ids
		}
		(*kindToIds)[kind] = existingIds
	}
	return allAssetIds
}
