/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clusterservice

import (
	"context"
	"net/http"

	csclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
)

type ClientInterface interface {
	GetNodePoolsInternalByCluster(ctx context.Context, clusterUUID string) (*openapi.GetNodepoolsInternal200Response, error)
	GetNodePoolsMinimalByCluster(ctx context.Context, clusterUUID string) (*openapi.GetNodepoolsMinimal200Response, error)
	GetNetworkTopologies(ctx context.Context, filterBy []string) (*openapi.GetNetworkTopologies200Response, error)
	GetClusterByUuid(ctx context.Context, id string) (*openapi.DisplayedCluster, *openapi.Error)
	GetInternalCluster(ctx context.Context, clusterId string) (*openapi.DisplayedCluster, error)
	GetNodeNetworkTopologyLabelValues(ctx context.Context, clusterUuid string, nodeName string) ([]string, error)
	GetNodeInternal(ctx context.Context, clusterUuid string, nodeName string) (*openapi.NodeInternal, *http.Response, error)
}

func GetRealClient(clusterServiceUrl string) ClientInterface {
	return csclient.NewCachedClusterClient(clusterServiceUrl)
}
