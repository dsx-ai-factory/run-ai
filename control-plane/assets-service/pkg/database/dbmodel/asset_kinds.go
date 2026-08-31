/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"cmp"
	"fmt"
	"net/http"
	"slices"

	"github.com/run-ai/runai/runai-application-packages/convutils"

	"github.com/run-ai/runai/runai-application-packages/datastructs"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
)

type (
	AssetIdToMetaMap  map[string]CombinedRec
	AssetKindToIdsMap map[api.AssetKind][]string
)

// ValidateKindOfAssets receives a list of asset ids and validates that they all exist and have given asset kind
func (m *AssetIdToMetaMap) ValidateKindOfAssets(kind api.AssetKind, ids []string, workloadVolumeAssets datastructs.Set[string], targetScope api.Scope) *api.Error {
	for _, id := range ids {
		assetMeta, ok := (*m)[id]
		if !ok {
			return api.NewError(http.StatusNotFound, fmt.Sprintf("asset %s is not associated with the requested %s or its ancestors", id, targetScope), nil)
		}
		if assetMeta.Kind != kind {
			return api.NewValidationError(fmt.Sprintf("asset %s is a %s rather than %s", id, assetMeta.Kind, kind))
		}
		if workloadVolumeAssets.Has(id) {
			if assetMeta.Kind != api.ASSETKIND_PVC {
				return api.NewValidationError(fmt.Sprintf("%s asset %s is incompatible with workloadVolume (asset kind should be PVC)", assetMeta.Kind, id))
			}
			if !convutils.Deref(assetMeta.AutoDelete, DefaultForAutoDelete) {
				return api.NewValidationError(fmt.Sprintf("workloadVolume %s is not an autoDelete asset", id))
			}
		} else {
			if convutils.Deref(assetMeta.AutoDelete, DefaultForAutoDelete) {
				return api.NewValidationError(fmt.Sprintf("autoDelete %s asset %s should not be used as datasource", kind, id))
			}
		}
	}
	return nil
}

func (m *AssetIdToMetaMap) Sorted() []*CombinedRec {
	result := make([]*CombinedRec, 0)

	for i := range *m {
		rec := (*m)[i]
		result = append(result, &rec)
	}

	slices.SortFunc(result, func(a, b *CombinedRec) int {
		if c := cmp.Compare(a.Kind, b.Kind); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Scope, b.Scope); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Name, b.Name); c != 0 {
			return c
		}
		return cmp.Compare(a.Id, b.Id)
	})

	return result
}
