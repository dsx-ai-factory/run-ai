/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster

import (
	"context"
	"fmt"
	"io"
	"net/http"

	clusterService "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

// Client -> for the "real" policies client we need an APIClient to the cluster service
type Client struct {
	apiClient *clusterService.APIClient
}

// Factory -> the "real" client factory. during testing it is replaced with factory which
// generated mock clients
type Factory struct{}

func (self Factory) Client() ClientInterface {
	return &Client{
		apiClient: clusterService.NewAPIClient(serverConfig),
	}
}

func (self Client) GetClusterByUuid(ctx context.Context, clusterId string) (*clusterService.DisplayedCluster, *clusterService.Error) {
	resp, r, err := self.apiClient.ClustersAPI.GetClusterByUuid(ctx, clusterId).Verbosity("full").Execute()
	if r != nil {
		if r.StatusCode == http.StatusOK {
			return resp, nil
		}
		if r.StatusCode == http.StatusNotFound {
			return nil, clusterService.NewError(http.StatusNotFound, fmt.Sprintf("cluster %s does not exist", clusterId))
		}
		data, _ := io.ReadAll(r.Body)

		logging.Errorf("failed to get info of cluster %s: %s", clusterId, data)
		return nil, clusterService.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get info of cluster %s", clusterId))
	}

	logging.Errorf("failed to request info of cluster %s: %v", clusterId, err)
	return nil, clusterService.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request info of cluster %s", clusterId))
}

func (self Client) GetClusterInternal(ctx context.Context, clusterId string) (*clusterService.DisplayedCluster, *clusterService.Error) {
	resp, r, err := self.apiClient.ClustersAPI.GetClusterInternal(ctx, clusterId).Execute()
	if r != nil {
		if r.StatusCode == http.StatusOK {
			return resp, nil
		}
		if r.StatusCode == http.StatusNotFound {
			return nil, clusterService.NewError(http.StatusNotFound, fmt.Sprintf("cluster %s does not exist", clusterId))
		}
		data, _ := io.ReadAll(r.Body)

		logging.Errorf("failed to get info of cluster %s: %s", clusterId, data)
		return nil, clusterService.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get info of cluster %s", clusterId))
	}

	logging.Errorf("failed to request info of cluster %s: %v", clusterId, err)
	return nil, clusterService.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request info of cluster %s", clusterId))
}

func (self Client) GetClusterMinimalDataById(ctx context.Context, clusterId string) (*clusterService.ClustersMinimalData, *clusterService.Error) {
	resp, r, err := self.apiClient.ClustersAPI.GetClusterMinimalDataById(ctx, clusterId).Execute()
	if r != nil {
		if r.StatusCode == http.StatusOK {
			return resp, nil
		}
		if r.StatusCode == http.StatusNotFound {
			return nil, clusterService.NewError(http.StatusNotFound, fmt.Sprintf("cluster %s does not exist", clusterId))
		}
		data, _ := io.ReadAll(r.Body)

		logging.Errorf("failed to get minimal info of cluster %s: %s", clusterId, data)
		return nil, clusterService.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get minimal info of cluster %s", clusterId))
	}

	logging.Errorf("failed to request minimal info of cluster %s: %v", clusterId, err)
	return nil, clusterService.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request minimal info of cluster %s", clusterId))
}
