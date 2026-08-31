/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloadutils

import (
	"context"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/workloadutils/api"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"

	"net/http"
)

// CheckLWSAvailable validates that LeaderWorkerSet (LWS) is available in the cluster for distributed inference
func CheckLWSAvailable(ctx context.Context, clusterId string, notAvailableMsg string) *api.Error {
	// Try cached version first
	available, apiErr := getLWSStatus(ctx, clusterId)
	if apiErr != nil {
		return apiErr
	}

	// If we fail based on cached data, perform ONE sync request
	if !available {
		availableSync, apiErrSync := getLWSStatusSync(ctx, clusterId)
		if apiErrSync != nil {
			return apiErrSync
		}

		// Use fresh data for final validation
		available = availableSync
	}

	if !available {
		return api.NewError(http.StatusBadRequest, notAvailableMsg, nil)
	}

	return nil
}

func getLWSStatus(ctx context.Context, clusterId string) (bool, *api.Error) {
	clusterChecker := checker.NewClusterCapabilitiesChecker(clusterId)
	config, err := clusterChecker.ClusterConfiguration(ctx)
	if err != nil {
		return false, api.NewError(http.StatusServiceUnavailable, err.Error(), nil)
	}

	return config.LWSAvailable, nil
}

// getLWSStatusSync retrieves LWS status without using cache and stores it in cache
func getLWSStatusSync(ctx context.Context, clusterId string) (bool, *api.Error) {
	config, err := checker.FetchClusterConfigurationSyncAndStore(ctx, clusterId)
	if err != nil {
		return false, api.NewError(http.StatusServiceUnavailable, err.Error(), nil)
	}

	return config.LWSAvailable, nil
}
