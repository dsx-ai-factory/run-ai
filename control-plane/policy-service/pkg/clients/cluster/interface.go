/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster

import (
	"context"

	clusterService "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
)

type ClientInterface interface {
	GetClusters(ctx context.Context) ([]clusterService.DisplayedCluster, *clusterService.Error)
	GetClusterByUuid(ctx context.Context, id string) (*clusterService.DisplayedCluster, *clusterService.Error)
	GetClustersMinimalData(ctx context.Context) (*clusterService.GetClustersMinimalData200Response, error)
	GetClusterMinimalDataById(ctx context.Context, clusterId string) (*clusterService.ClustersMinimalData, error)
}
