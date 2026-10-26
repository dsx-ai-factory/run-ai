/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
)

// defaultTestKubeVersion is what BuildDisplayedCluster reports as the
// cluster's kube version by default. Chosen to satisfy DRA's floor
// (MinDraK8sVersion = "1.34.0") so RCT create-path tests don't 503 out
// on the platform-info requirement in ValidateAssetCreateRequestMeta.
// Tests that need a specific kube version (e.g. below-floor rejection)
// build the DisplayedCluster manually or mutate the result before
// installing it in the mock.
const defaultTestKubeVersion = "1.34.0"

func BuildDisplayedCluster(randoms *testutils.RandomValues, version string, subdomainEnabled bool) openapi.DisplayedCluster {
	result := openapi.DisplayedCluster{
		Uuid:     randoms.ClusterId,
		TenantId: int32(randoms.TenantId),
	}
	result.Version.Set(&version)
	config := openapi.NewClusterReportedConfig()
	config.SubdomainEnabled.Set(&subdomainEnabled)
	status := openapi.NewClusterDisplayedStatus()
	status.SetConfig(*config)
	platform := openapi.NewClusterReportedStatusPlatform()
	kubeVer := defaultTestKubeVersion
	platform.KubeVersion.Set(&kubeVer)
	status.SetPlatform(*platform)
	result.Status.Set(status)
	return result
}
