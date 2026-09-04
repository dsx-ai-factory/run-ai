/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

func MergedAssetMetaOf(row AssetRecord) api.AssetMeta {
	meta := row.GetApiAssetMeta()
	meta.WorkloadSupportedTypes = row.GetApiWorkloadSupportedTypes()

	// if WorkloadSupportedTypes is not set
	// then set it to default value (Workspace and Training)
	if meta.WorkloadSupportedTypes != nil && NoWSTIsGiven(meta.WorkloadSupportedTypes) {
		setDefaultWST(&meta)
	}
	return meta
}

func MergedAssetMetaAndWST(assetMeta *dbmodel.AssetMeta, workloadSupportedTypes *dbmodel.WorkloadSupportedTypes) api.AssetMeta {
	meta := assetMeta.ToApiAssetMeta()
	meta.WorkloadSupportedTypes = workloadSupportedTypes.ToApiWorkloadSupportedTypes()

	// if WorkloadSupportedTypes is not set
	// then set it to default value (Workspace and Training)
	if NoWSTIsGiven(meta.WorkloadSupportedTypes) {
		setDefaultWST(&meta)
	}

	return meta
}

func setDefaultWST(meta *api.AssetMeta) {
	meta.WorkloadSupportedTypes = &api.WorkloadSupportedTypes{
		Workspace:     convutils.Ref(true),
		Training:      convutils.Ref(true),
		Inference:     convutils.Ref(false),
		Distributed:   convutils.Ref(false),
		DistFramework: nil,
	}
}

// NoWSTIsGiven checks if all of the WorkloadSupportedTypes are nil
func NoWSTIsGiven(wst *api.WorkloadSupportedTypes) bool {
	if wst == nil {
		return false
	}
	if wst.Distributed == nil && wst.Training == nil &&
		wst.Workspace == nil && wst.DistFramework == nil && wst.Inference == nil {
		return true
	}
	return false
}
