/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

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

func assetRecToAssetRef(assetsRef *api.AssetsRef, assets []dbmodel.AssetMetaRec, assetOverridesMapById map[string]*api.DataSourceOverrides) *api.Error {
	for _, asset := range assets {

		assetOverrides, ok := assetOverridesMapById[asset.Id]
		if !ok {
			assetOverrides = nil
		}

		switch asset.Kind {
		case api.ASSETKIND_ENVIRONMENT:
			assetsRef.Environment = api.EnvironmentAssetRef{
				Id:   asset.Id,
				Name: asset.Name,
			}
		case api.ASSETKIND_COMPUTE:
			assetsRef.Compute = &api.AssetsRefCompute{
				Id:   asset.Id,
				Name: asset.Name,
			}
		case api.ASSETKIND_PVC:
			if convutils.Deref(asset.AutoDelete, dbmodel.DefaultForAutoDelete) {
				// assetOverrides is not relevant for workload Volumes, therefore its not added here
				assetsRef.WorkloadVolumes = appendToWorkloadVolumes(assetsRef.WorkloadVolumes, asset.Id)
			} else {
				assetsRef.Datasources = appendToDatasourceRef(assetsRef.Datasources, asset, assetOverrides)
			}
		case api.ASSETKIND_S3, api.ASSETKIND_GIT, api.ASSETKIND_NFS, api.ASSETKIND_HOST_PATH, api.ASSETKIND_CONFIG_MAP, api.ASSETKIND_SECRET_VOLUME:
			assetsRef.Datasources = appendToDatasourceRef(assetsRef.Datasources, asset, assetOverrides)
		default:
			return api.NewError(http.StatusInternalServerError, fmt.Sprintf("unsupported asset kind: %s", asset.Kind), nil)
		}
	}
	sort.Slice(assetsRef.Datasources, func(i, j int) bool {
		return assetsRef.Datasources[i].Id < assetsRef.Datasources[j].Id
	})
	return nil
}

// WorkloadRecToAssetsRef reuse the assetRecToAssetsRef function for workload asset
func WorkloadRecToAssetsRef(assetsRef *api.AssetsRef, workloadAssets []dbmodel.CombinedWorkloadsRec, currentDistReplicaType *dbmodel.DistReplicaType) *api.Error {
	//
	// re-arrange the input in a way that assetRecToAssetIds expect to receive
	//
	assets := make([]dbmodel.AssetMetaRec, 0, len(workloadAssets))
	assetOverridesById := make(map[string]*api.DataSourceOverrides)
	for _, asset := range workloadAssets {
		if currentDistReplicaType != nil && asset.WorkloadsAssetsRec.DistReplicaType != *currentDistReplicaType {
			// skip assets that are not relevant to the current dist replica type
			continue
		}
		assets = append(assets, dbmodel.AssetMetaRec{
			AssetMeta: dbmodel.AssetMeta{
				Id:         asset.WorkloadsAssetsRec.AssetId,
				Name:       asset.AssetName,
				Kind:       asset.WorkloadsAssetsRec.AssetKind,
				AutoDelete: convutils.Ref(asset.WorkloadsAssetsRec.AutoDelete),
			},
		})
		if asset.WorkloadsAssetsRec.Overrides != nil {
			assetOverridesById[asset.WorkloadsAssetsRec.AssetId] = asset.WorkloadsAssetsRec.Overrides
		}
	}

	return assetRecToAssetRef(assetsRef, assets, assetOverridesById)
}

func templateAssetsRecToMap(containment []dbmodel.TemplateAssetsCombinedRec, envIdToTools EnvironmentToolsMap) (TemplateToAssetsMap, *api.Error) {
	resultMap := make(TemplateToAssetsMap)
	for _, record := range containment {
		var curTemplate api.AssetsRef
		if _, ok := resultMap[record.TemplateAssetId]; ok {
			curTemplate = resultMap[record.TemplateAssetId]
		}
		switch record.AssetMeta.Kind {
		case api.ASSETKIND_ENVIRONMENT:
			curTemplate.Environment = api.EnvironmentAssetRef{Id: record.AssetMeta.Id, Name: record.AssetMeta.Name}
			toolTypes := envIdToTools[record.AssetMeta.Id]
			if toolTypes != nil {
				curTemplate.Environment.ToolTypes = &toolTypes
			}
		case api.ASSETKIND_COMPUTE:
			curTemplate.Compute = &api.AssetsRefCompute{Id: record.AssetMeta.Id, Name: record.AssetMeta.Name}
		case api.ASSETKIND_PVC:
			if convutils.Deref(record.AssetMeta.AutoDelete, DefaultAutoDelete) {
				curTemplate.WorkloadVolumes = appendToWorkloadVolumes(curTemplate.WorkloadVolumes, record.AssetMeta.Id)
			} else {
				curTemplate.Datasources = appendToDatasourceRef(curTemplate.Datasources, record.AssetMeta, nil)
			}
		case api.ASSETKIND_S3, api.ASSETKIND_GIT, api.ASSETKIND_NFS, api.ASSETKIND_HOST_PATH, api.ASSETKIND_CONFIG_MAP, api.ASSETKIND_SECRET_VOLUME:
			curTemplate.Datasources = appendToDatasourceRef(curTemplate.Datasources, record.AssetMeta, nil)
		default:
			return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("unsupported asset kind %s: asset %s template %s",
				record.AssetMeta.Kind, record.AssetMeta.Id, record.TemplateAssetId), nil)
		}

		resultMap[record.TemplateAssetId] = curTemplate
	}
	return resultMap, nil
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

func appendToDatasourceRef(assetsRef []api.DatasourceRef, asset dbmodel.AssetMetaRec, overrideData *api.DataSourceOverrides) []api.DatasourceRef {
	return append(assetsRef, api.DatasourceRef{Id: asset.Id, Name: asset.Name, Kind: asset.Kind, Overrides: overrideData})
}

func appendToWorkloadVolumes(workloadVolumes []string, assetId string) []string {
	if workloadVolumes == nil {
		return []string{assetId}
	} else {
		return append(workloadVolumes, assetId)
	}
}
