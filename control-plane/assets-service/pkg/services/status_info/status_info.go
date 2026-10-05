/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package status_info

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/asset_cluster_status"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/response_patching"
)

// PatchStatusInfoMultipleAssets performs PatchStatusInfo on set of assets containing mixed asset kinds,
// so it first needs to sort out the patching pointers based on asset kind, then calls PatchUsageInfo per kind
func PatchStatusInfoMultipleAssets(ctx context.Context, assetsClusterStatus *asset_cluster_status.AssetClusterStatus,
	patchingInfo map[string]response_patching.ResponsePatchingPtrs,
) *api.Error {
	patchingInfoPerKind := map[api.AssetKind]map[string]response_patching.ResponsePatchingPtrs{}

	for assetId, assetInfo := range patchingInfo {
		assetKind := assetInfo.Meta.Kind

		if patchingInfoPerKind[assetKind] == nil {
			patchingInfoPerKind[assetKind] = map[string]response_patching.ResponsePatchingPtrs{
				assetId: assetInfo,
			}
		} else {
			patchingInfoPerKind[assetKind][assetId] = assetInfo
		}
	}

	for assetKind, patchingInfoOfThisKind := range patchingInfoPerKind {
		err := PatchStatusInfo(ctx, assetsClusterStatus, patchingInfoOfThisKind, assetKind)
		if err != nil {
			return err
		}
	}

	return nil
}

// PatchStatusInfo The cluster adds a prefix when creating the secret name, so we need to map the secret type to the prefix
func PatchStatusInfo(ctx context.Context, assetsClusterStatus *asset_cluster_status.AssetClusterStatus,
	patchingInfo map[string]response_patching.ResponsePatchingPtrs, assetKind api.AssetKind,
) *api.Error {
	if len(patchingInfo) == 0 {
		return nil
	}

	shouldLoadClusterResources := false
	relevantAssetIds := datastructs.Set[string]{}

	for _, curPatch := range patchingInfo {
		if curPatch.Meta.Kind == assetKind {
			shouldLoadClusterResources = true

			// at this point we do not sync tenant and system assets to the cluster, therefore
			// the relevant assets should all have cluster id

			if curPatch.Meta.ClusterId != nil || convutils.BoolIsTrue(curPatch.Meta.Private) {
				relevantAssetIds.Add(curPatch.Meta.Id)
			}
		}
	}

	if !shouldLoadClusterResources {
		return nil
	}

	syncBasedStatus, numPatched := LoadSyncStatus(ctx, relevantAssetIds, patchingInfo)
	if numPatched == len(relevantAssetIds) {
		return nil // all relevant assets were patched, so no need to proceed to objects tracker status
	}

	// Kinds without a k8s-objects-tracker projection (e.g. resource-claim-template)
	// draw their status exclusively from sync_status. Skipping the tracker
	// step avoids a wasted HTTP call with an empty resourceType — the
	// assets that LoadSyncStatus didn't cover simply carry no status.
	if !asset_cluster_status.HasClusterResourceMapping(assetKind) {
		return nil
	}

	assetsClusterStatus.StoreSyncBasedInfo(syncBasedStatus)

	assetsClusterStatus.LoadAssetsStatus(ctx, assetKind)

	assetStatuses, err := assetsClusterStatus.GetAssetsStatus()
	if err != nil {
		return err
	}

	for curPatchInfo := range patchingInfo {
		patchPtr := patchingInfo[curPatchInfo]
		if *patchPtr.Status != nil {
			continue // already filled by sync status
		}
		assetStatus := assetStatuses[curPatchInfo]
		if assetStatus != nil {
			*patchPtr.Status = assetStatus
		}
	}

	return nil
}

// LoadSyncStatus patch the status info based on sync status table.
// it returns the number of asset that it.
func LoadSyncStatus(ctx context.Context, relevantIds datastructs.Set[string],
	patchingInfo map[string]response_patching.ResponsePatchingPtrs,
) (map[string]*dbmodel.SyncStatusRec, int) {
	patched := 0
	tenantId := auth.TenantIdOf(ctx)

	statusInfo, err := dao.ListSyncStatusByAssetIds(ctx, tenantId, relevantIds.Keys())
	if err != nil {
		return nil, patched // error was already logged by the dao, we have nothing else to do
	}

	for assetId, curPatch := range patchingInfo {
		assetStatus, found := statusInfo[assetId]
		if !found {
			continue
		}

		assetStatus.DetermineStatus(ctx, curPatch.Meta, curPatch.Status)

		if *curPatch.Status != nil {
			patched += 1
		}
	}

	return statusInfo, patched
}
