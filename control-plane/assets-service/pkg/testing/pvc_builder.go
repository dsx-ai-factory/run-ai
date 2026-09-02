/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"strings"
)

func BuildRandomPVCHelper() api.PvcCreationRequest {
	name := testutils.RandString(5)
	return api.PvcCreationRequest{
		Meta: api.AssetCreationRequest{
			Name: name,
		},
		Spec: api.PvcAssetSpec{
			ClaimName: convutils.Ref(name),
			Path:      convutils.Ref(testutils.RandString(5)),
			ReadOnly:  convutils.Ref(false),
			ClaimInfo: &api.ClaimInfo{
				StorageClass: convutils.Ref(strings.ToLower(testutils.RandString(8))),
				Size:         convutils.Ref("10M"),
				AccessModes: &api.PvcAccessModes{
					ReadWriteMany: convutils.Ref(true),
					ReadOnlyMany:  convutils.Ref(true),
					ReadWriteOnce: convutils.Ref(true),
				},
				VolumeMode: convutils.Ref(api.PVCVOLUMEMODE_FILESYSTEM),
			},
		},
	}
}

func BuildRandomSystemPVCAsset() api.PvcCreationRequest {
	ans := BuildRandomPVCHelper()
	ans.Meta.Scope = api.SCOPE_SYSTEM
	return ans
}

func BuildRandomTenantPVCAsset() api.PvcCreationRequest {
	ans := BuildRandomPVCHelper()
	ans.Meta.Scope = api.SCOPE_TENANT
	return ans
}

func BuildRandomClusterPVCAsset(randoms *testutils.RandomValues) api.PvcCreationRequest {
	ans := BuildRandomPVCHelper()
	ans.Meta.Scope = api.SCOPE_CLUSTER
	ans.Meta.ClusterId = convutils.Ref(randoms.ClusterId)
	return ans
}

func BuildRandomProjectPVCAsset(randoms *testutils.RandomValues) api.PvcCreationRequest {
	ans := BuildRandomPVCHelper()
	ans.Meta.Scope = api.SCOPE_PROJECT
	ans.Meta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
	return ans
}

func BuildRandomDepartmentPVCAsset(randoms *testutils.RandomValues) api.PvcCreationRequest {
	ans := BuildRandomPVCHelper()
	ans.Meta.Scope = api.SCOPE_DEPARTMENT
	ans.Meta.DepartmentId = convutils.Ref(randoms.DepartmentId)
	return ans
}
