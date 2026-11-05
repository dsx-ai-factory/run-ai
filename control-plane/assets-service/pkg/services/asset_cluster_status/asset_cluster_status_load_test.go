/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package asset_cluster_status

import (
	"context"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/response_patching"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/stretchr/testify/assert"
)

func secretsPatchingInfo(ids ...string) map[string]response_patching.ResponsePatchingPtrs {
	patchingInfo := make(map[string]response_patching.ResponsePatchingPtrs, len(ids))
	for _, id := range ids {
		patchingInfo[id] = response_patching.ResponsePatchingPtrs{
			Meta: &api.AssetMeta{Id: id, Kind: api.ASSETKIND_GENERIC_SECRET},
		}
	}
	return patchingInfo
}

// LoadAssetsStatus may be called again with the same assets for the same kind:
// the second call has nothing left to load and must succeed.
func TestLoadAssetsStatusAcceptsTheSameAssetsAgain(t *testing.T) {
	ids := []string{"secret-1", "secret-2", "secret-3"}
	acs := NewAssetClusterStatus(secretsPatchingInfo(ids...), nil)
	// What the first call for this kind stored.
	acs.assetIdsPerKind[api.ASSETKIND_GENERIC_SECRET] = datastructs.NewSet(ids...)

	err := acs.LoadAssetsStatus(context.Background(), api.ASSETKIND_GENERIC_SECRET)

	assert.Nil(t, err, "loading the same assets again failed: %v", err)
}

// A different call for the same kind is rejected, as the comment on
// LoadAssetsStatus promises.
func TestLoadAssetsStatusRejectsDifferentAssets(t *testing.T) {
	acs := NewAssetClusterStatus(secretsPatchingInfo("secret-1", "secret-2"), nil)
	acs.assetIdsPerKind[api.ASSETKIND_GENERIC_SECRET] = datastructs.NewSet("secret-1", "secret-2", "secret-3")

	err := acs.LoadAssetsStatus(context.Background(), api.ASSETKIND_GENERIC_SECRET)

	assert.NotNil(t, err, "expected the extra asset of the first load to be reported")
}

// The stored set is the one that has to match: an asset that was loaded before
// and is missing now is a different set too.
func TestLoadAssetsStatusRejectsAssetsLoadedBefore(t *testing.T) {
	acs := NewAssetClusterStatus(secretsPatchingInfo("secret-1", "secret-2"), nil)
	acs.assetIdsPerKind[api.ASSETKIND_GENERIC_SECRET] = datastructs.NewSet("secret-1", "secret-2", "secret-3")
	acs.assetIdsPerKind[api.ASSETKIND_PVC] = datastructs.NewSet("pvc-1")

	err := acs.LoadAssetsStatus(context.Background(), api.ASSETKIND_GENERIC_SECRET)

	assert.NotNil(t, err, "expected the missing asset to be reported")
}
