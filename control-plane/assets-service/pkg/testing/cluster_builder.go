/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-application-packages/testutils"
)

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
	result.Status.Set(status)
	return result
}
