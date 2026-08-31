/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
)

// BuildRandomSystemSecretVolumeAsset build a secret storage creating request in RUNAI scope
func BuildRandomSystemSecretVolumeAsset() api.SecretAssetCreationRequest {
	ans := buildRandomSecretVolumeAssetHelper(nil)
	ans.Meta.Scope = api.SCOPE_SYSTEM
	return ans
}

// BuildRandomTenantSecretVolumeAsset build a secret storage creating request in a tenant scope
func BuildRandomTenantSecretVolumeAsset() api.SecretAssetCreationRequest {
	ans := buildRandomSecretVolumeAssetHelper(nil)
	ans.Meta.Scope = api.SCOPE_TENANT
	return ans
}

func BuildRandomClusterSecretVolumeAsset(randoms *testutils.RandomValues, credentialAssetId *string) api.SecretAssetCreationRequest {
	ans := buildRandomSecretVolumeAssetHelper(credentialAssetId)
	ans.Meta.Scope = api.SCOPE_CLUSTER
	ans.Meta.ClusterId = convutils.Ref(randoms.ClusterId)
	return ans
}

// BuildRandomProjectSecretVolumeAsset build a secret storage creating request in a project scope
func BuildRandomProjectSecretVolumeAsset(randoms *testutils.RandomValues, credentialAssetId *string) api.SecretAssetCreationRequest {
	ans := buildRandomSecretVolumeAssetHelper(credentialAssetId)
	ans.Meta.Scope = api.SCOPE_PROJECT
	ans.Meta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
	return ans
}

// BuildRandomDepartmentSecretVolumeAsset builds a secret storage creating request in a department scope
func BuildRandomDepartmentSecretVolumeAsset(randoms *testutils.RandomValues, credentialAssetId *string) api.SecretAssetCreationRequest {
	ans := buildRandomSecretVolumeAssetHelper(credentialAssetId)
	ans.Meta.Scope = api.SCOPE_DEPARTMENT
	ans.Meta.DepartmentId = convutils.Ref(randoms.DepartmentId)
	return ans
}

// buildRandomSecretVolumeAssetHelper builds a Secret storage creating request
func buildRandomSecretVolumeAssetHelper(credentialAssetId *string) api.SecretAssetCreationRequest {
	return api.SecretAssetCreationRequest{
		Meta: api.AssetCreationRequest{
			Name: testutils.RandString(5),
		},
		Spec: api.SecretAssetSpec{
			CredentialAssetId: credentialAssetId,
			MountPath:         convutils.Ref(testutils.RandString(5)),
			DefaultMode:       convutils.Ref("0644"),
		},
	}
}
