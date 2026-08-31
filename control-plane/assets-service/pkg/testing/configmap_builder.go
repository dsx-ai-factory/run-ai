/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	"strings"
)

// BuildRandomSystemConfigMapAsset build a ConfigMap creating request in RUNAI scope
func BuildRandomSystemConfigMapAsset() api.ConfigMapCreationRequest {
	ans := buildRandomConfigMapHelper()
	ans.Meta.Scope = api.SCOPE_SYSTEM
	return ans
}

// BuildRandomTenantConfigMapAsset build a ConfigMap creating request in a tenant scope
func BuildRandomTenantConfigMapAsset() api.ConfigMapCreationRequest {
	ans := buildRandomConfigMapHelper()
	ans.Meta.Scope = api.SCOPE_TENANT
	return ans
}

func BuildRandomClusterConfigMapAsset(randoms *testutils.RandomValues) api.ConfigMapCreationRequest {
	ans := buildRandomConfigMapHelper()
	ans.Meta.Scope = api.SCOPE_CLUSTER
	ans.Meta.ClusterId = convutils.Ref(randoms.ClusterId)
	return ans
}

// BuildRandomProjectConfigMapAsset build a ConfigMap creating request in a project scope
func BuildRandomProjectConfigMapAsset(randoms *testutils.RandomValues) api.ConfigMapCreationRequest {
	ans := buildRandomConfigMapHelper()
	ans.Meta.Scope = api.SCOPE_PROJECT
	ans.Meta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
	return ans
}

// BuildRandomDepartmentConfigMapAsset builds a ConfigMap creating request in a department scope
func BuildRandomDepartmentConfigMapAsset(randoms *testutils.RandomValues) api.ConfigMapCreationRequest {
	ans := buildRandomConfigMapHelper()
	ans.Meta.Scope = api.SCOPE_DEPARTMENT
	ans.Meta.DepartmentId = convutils.Ref(randoms.DepartmentId)
	return ans
}

// buildRandomConfigMapHelper builds a ConfigMap creating request
func buildRandomConfigMapHelper() api.ConfigMapCreationRequest {
	name := convutils.Ref(strings.ToLower(testutils.RandString(8)))
	return api.ConfigMapCreationRequest{
		Meta: api.AssetCreationRequest{
			Name: *name,
		},
		Spec: api.ConfigMapAssetSpec{
			ConfigMap:   name,
			MountPath:   convutils.Ref(testutils.RandString(5)),
			SubPath:     convutils.Ref(testutils.RandString(5)),
			DefaultMode: convutils.Ref("0644"),
		},
	}
}
