/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"fmt"

	clusterclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
)

func (c *Clients) Cluster(ctx context.Context, name string) (clusterclient.DisplayedCluster, error) {
	clusters, _, err := c.ClusterService.ClustersAPI.GetClusters(ctx).Execute()
	if err != nil {
		return clusterclient.DisplayedCluster{}, fmt.Errorf("listing clusters: %w", err)
	}
	if name == "" {
		if len(clusters) != 1 {
			return clusterclient.DisplayedCluster{}, fmt.Errorf("control plane has %d clusters, choose one with the cluster knob or --cluster", len(clusters))
		}
		return clusters[0], nil
	}
	for _, cluster := range clusters {
		if cluster.Name == name {
			return cluster, nil
		}
	}
	return clusterclient.DisplayedCluster{}, fmt.Errorf("cluster %q not found on the control plane", name)
}
