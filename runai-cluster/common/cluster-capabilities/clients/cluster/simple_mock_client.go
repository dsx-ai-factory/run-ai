/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster

import (
	"context"
	"fmt"
	"net/http"

	clusterService "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"k8s.io/utils/ptr"
)

// SimpleMockClient only supports fixed set of clusters
type SimpleMockClient struct {
	Clusters                map[string]clusterService.DisplayedCluster
	GetClusterByIdCallCount *int
}

func (s SimpleMockClient) Client() ClientInterface {
	return s
}

func (self SimpleMockClient) GetClusterByUuid(_ context.Context, clusterId string) (*clusterService.DisplayedCluster, *clusterService.Error) {
	*self.GetClusterByIdCallCount++
	result, found := self.Clusters[clusterId]
	if !found {
		return nil, clusterService.NewError(http.StatusNotFound, fmt.Sprintf("cluster %s does not exist", clusterId))
	}

	return &result, nil
}

func (self SimpleMockClient) GetClusterInternal(_ context.Context, clusterId string) (*clusterService.DisplayedCluster, *clusterService.Error) {
	*self.GetClusterByIdCallCount++
	result, found := self.Clusters[clusterId]
	if !found {
		return nil, clusterService.NewError(http.StatusNotFound, fmt.Sprintf("cluster %s does not exist", clusterId))
	}

	return &result, nil
}

func (self SimpleMockClient) GetClusterMinimalDataById(_ context.Context, clusterId string) (*clusterService.ClustersMinimalData, *clusterService.Error) {
	displayed, found := self.Clusters[clusterId]
	if !found {
		return nil, clusterService.NewError(http.StatusNotFound, fmt.Sprintf("cluster %s does not exist", clusterId))
	}
	minimal := clusterService.ClustersMinimalData{
		Uuid:     displayed.Uuid,
		Name:     displayed.Name,
		TenantId: displayed.TenantId,
	}
	if displayed.Version.IsSet() {
		minimal.Version = displayed.Version
	}
	if displayed.Domain.IsSet() {
		minimal.Domain = displayed.Domain
	}
	return &minimal, nil
}

type SimpleMockClientFactory struct {
	SimpleMockClient
}

func (self SimpleMockClientFactory) Client() ClientInterface {
	return &self.SimpleMockClient
}

func NewSimpleMockClient() *SimpleMockClient {
	return &SimpleMockClient{
		Clusters:                make(map[string]clusterService.DisplayedCluster),
		GetClusterByIdCallCount: ptr.To(0),
	}
}

func (s SimpleMockClient) WithCluster(clusterId string, clusterInfo clusterService.DisplayedCluster) SimpleMockClient {
	s.Clusters[clusterId] = clusterInfo
	return s
}
