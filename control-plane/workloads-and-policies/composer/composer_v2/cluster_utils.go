/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer_v2

import (
	clustersApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"net/http"
)

func ValidateGpuVendorSupported(cluster *clustersApi.DisplayedCluster) *workloadsApi.Error {
	if cluster == nil {
		return nil
	}

	if !cluster.Status.IsSet() {
		return workloadsApi.NewError(http.StatusServiceUnavailable, "cluster status is not available")
	}
	if cluster.Status.Get().Dependencies == nil || cluster.Status.Get().Dependencies.Required == nil {
		return workloadsApi.NewError(http.StatusServiceUnavailable, "cluster dependencies are not available")
	}
	gpuStack := cluster.Status.Get().Dependencies.Required["gpu-stack"]
	if !gpuStack.Available {
		return workloadsApi.NewError(http.StatusBadRequest, "gpu stack is not available")
	}
	if nvidia, ok := (*gpuStack.Components)["nvidia"]; !ok || !nvidia.Available {
		return workloadsApi.NewError(http.StatusBadRequest, "nvidia gpu vendor is not available")
	}

	return nil
}
