/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
)

// this function is a workaround for asset-service compute assets which contains request type = PORTION
// while the asset itself is non-fractional. it attempts to create incorrect workload, because request
// type PORTION requires value in the gpuPortionRequest field. to avoid this issue, in case asset contains
// request type PORTION without having value in gpuPortionRequest - reset the request type to nil, indication
// non-fractional gpu request
// also when sending request type PORTION AND gpuPortionRequest == 1 AND gpuDevices > 1,
// we need to reset the request type and portionRequest in order to use multiple whole devices.
func fixPortionForNonFractionalIssue(assets []dbmodel.AssetsContainedInWorkload) {
	for idx, asset := range assets {
		if asset.Kind != api.ASSETKIND_COMPUTE || asset.AssetsCompute == nil {
			continue
		}

		computeSpec := asset.AssetsCompute.ComputeAssetSpec
		if computeSpec.GpuRequestType == nil {
			continue
		}
		if *computeSpec.GpuRequestType == api.GPUREQUESTTYPE_PORTION && computeSpec.GpuPortionRequest == nil {
			assets[idx].AssetsCompute.GpuRequestType = nil
			return
		}

		// When sending gpuDevicesRequest > 1 and sending portion == 1 - we ignore the portion since it is 100% of the devices.
		if computeSpec.GpuDevicesRequest != nil && *computeSpec.GpuDevicesRequest > 1 && computeSpec.GpuPortionRequest != nil && *computeSpec.GpuPortionRequest == 1 {
			assets[idx].AssetsCompute.GpuRequestType = nil
			assets[idx].AssetsCompute.GpuPortionRequest = nil
		}
	}
}

func StorageClassNameForMessage(pvc *api.PvcAssetSpec) string {
	var storageClassNameForMessage string
	if *pvc.ClaimInfo.StorageClass == "" {
		storageClassNameForMessage = "static provisioning"
	} else {
		storageClassNameForMessage = fmt.Sprintf("storage class %s", *pvc.ClaimInfo.StorageClass)
	}

	return storageClassNameForMessage
}
