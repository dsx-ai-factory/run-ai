/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package checker

import (
	"context"

	clusterService "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
)

const gpuComputeModeUnreportedReason = "GPU compute mode availability has not been reported"

// GpuComputeModeStatus refreshes a negative cached result before rejecting a mode,
// matching the checks for other configurable components such as Knative and LWS.
func (c *ClusterCapabilitiesChecker) GpuComputeModeStatus(ctx context.Context, mode string) (bool, string, error) {
	configuration, err := c.ClusterConfiguration(ctx)
	if err != nil {
		return false, "", err
	}
	status := gpuComputeModeStatus(configuration, mode)
	if !status.Available {
		configuration, err = FetchClusterConfigurationSyncAndStore(ctx, c.clusterId)
		if err != nil {
			return false, "", err
		}
		status = gpuComputeModeStatus(configuration, mode)
	}
	return status.Available, status.Reason, nil
}

func gpuComputeModeStatus(configuration *ClusterConfiguration, mode string) GpuComputeModeStatus {
	if configuration != nil {
		if status, found := configuration.GpuComputeModes[mode]; found {
			return status
		}
	}
	return GpuComputeModeStatus{Reason: gpuComputeModeUnreportedReason}
}

// GpuComputeModes reads runtime-reported availability independently of cluster version.
func GpuComputeModes(cluster *clusterService.DisplayedCluster) map[string]GpuComputeModeStatus {
	modes := map[string]GpuComputeModeStatus{
		"time-slicing": {Reason: gpuComputeModeUnreportedReason},
		"sm-sharing":   {Reason: gpuComputeModeUnreportedReason},
	}
	if cluster == nil {
		return modes
	}
	status := cluster.Status.Get()
	if status == nil {
		return modes
	}
	dependency, found := status.GetDependencies().Optional[GpuComputeModesDependency]
	if !found {
		return modes
	}
	for mode, unavailable := range modes {
		if !dependency.Available {
			if reason := dependency.GetReason(); reason != "" {
				unavailable.Reason = reason
			}
			modes[mode] = unavailable
			continue
		}
		if dependency.Components == nil {
			continue
		}
		if component, found := (*dependency.Components)[mode]; found {
			modes[mode] = GpuComputeModeStatus{Available: component.Available, Reason: component.GetReason()}
		}
	}
	return modes
}
