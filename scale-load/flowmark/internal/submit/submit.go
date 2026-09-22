/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"context"
	"errors"
	"fmt"

	clusterclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
)

const (
	ownerMarker           = "managed-by: flowmark"
	nodepoolsPageLimit    = int32(500)
	projectNamespaceLabel = "runai/queue"
)

var errProjectNotFound = errors.New("project not found")

type ProjectRef struct {
	ID        string
	Name      string
	Namespace string
	ClusterID string
}

type WorkloadRef struct {
	ID   string
	Name string
}

func ClusterID(ctx context.Context, clients *client.Clients, name string) (string, error) {
	clusters, _, err := clients.ClusterService.ClustersAPI.GetClusters(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("listing clusters: %w", err)
	}
	if name == "" {
		if len(clusters) != 1 {
			return "", fmt.Errorf("control plane has %d clusters, set the cluster knob to choose one", len(clusters))
		}
		return clusters[0].Uuid, nil
	}
	for _, cluster := range clusters {
		if cluster.Name == name {
			return cluster.Uuid, nil
		}
	}
	return "", fmt.Errorf("cluster %q not found on the control plane", name)
}

func nodepools(ctx context.Context, clients *client.Clients, clusterID string) ([]clusterclient.Nodepool, error) {
	var pools []clusterclient.Nodepool
	for offset := int32(0); ; offset += nodepoolsPageLimit {
		page, _, err := clients.ClusterService.NodePoolsAPI.GetNodepools(ctx).
			FilterBy([]string{"clusterId==" + clusterID}).Limit(nodepoolsPageLimit).Offset(offset).Execute()
		if err != nil {
			return nil, fmt.Errorf("listing nodepools of cluster %s: %w", clusterID, err)
		}
		pools = append(pools, page.Nodepools...)
		if len(page.Nodepools) < int(nodepoolsPageLimit) {
			break
		}
	}
	if len(pools) == 0 {
		return nil, fmt.Errorf("cluster %s has no nodepools", clusterID)
	}
	return pools, nil
}
