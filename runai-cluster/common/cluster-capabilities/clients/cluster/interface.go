/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster

import (
	"context"

	clusterService "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
)

type FactoryInterface interface {
	Client() ClientInterface
}

type ClientInterface interface {
	GetClusterByUuid(ctx context.Context, clusterId string) (*clusterService.DisplayedCluster, *clusterService.Error)
	GetClusterInternal(ctx context.Context, clusterId string) (*clusterService.DisplayedCluster, *clusterService.Error)
	GetClusterMinimalDataById(ctx context.Context, clusterId string) (*clusterService.ClustersMinimalData, *clusterService.Error)
}
