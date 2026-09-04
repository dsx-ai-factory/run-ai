/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	clusterService "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
	clusterServiceApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
)

type Client struct {
	apiClient clusterService.IClusterClient
}

func (c Client) GetClusters(ctx context.Context) ([]clusterServiceApi.DisplayedCluster, *clusterServiceApi.Error) {
	result, err := c.apiClient.GetClusters(ctx)
	if err != nil {
		return nil, clusterServiceApi.NewError(http.StatusServiceUnavailable, err.Error())
	}
	return result, nil
}

func (c Client) GetClusterByUuid(ctx context.Context, id string) (*clusterServiceApi.DisplayedCluster, *clusterServiceApi.Error) {
	result, err := c.apiClient.GetClusterByUuid(ctx, id)
	if err != nil {
		return nil, clusterServiceApi.NewError(http.StatusServiceUnavailable, err.Message)
	}
	return result, nil
}

func (c Client) GetClustersMinimalData(ctx context.Context) (*clusterServiceApi.GetClustersMinimalData200Response, error) {
	result, err := c.apiClient.GetClustersMinimalData(ctx)
	if err != nil {
		return nil, clusterServiceApi.NewError(http.StatusServiceUnavailable, err.Error())
	}
	return result, nil
}

func (c Client) GetClusterMinimalDataById(ctx context.Context, clusterId string) (*clusterServiceApi.ClustersMinimalData, error) {
	result, err := c.apiClient.GetClusterMinimalDataById(ctx, clusterId)
	if err != nil {
		return nil, clusterServiceApi.NewError(http.StatusServiceUnavailable, err.Error())
	}
	return result, nil
}

func GetClustersInfo(ctx context.Context, optionalUuid *string) (map[string]string, *api.Error) {
	result := make(map[string]string)

	var relevantClusters []clusterServiceApi.ClustersMinimalData

	if optionalUuid != nil {
		clusterInfo, err := Get().Client().GetClusterMinimalDataById(ctx, *optionalUuid)
		if err != nil {
			return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to obtain info of cluster %s", *optionalUuid), err)
		}
		relevantClusters = append(relevantClusters, *clusterInfo)
	} else {
		clusterData, err := Get().Client().GetClustersMinimalData(ctx)
		if err != nil {
			return nil, api.NewError(http.StatusServiceUnavailable, "failed to obtain the list of available clusters", err)
		}
		relevantClusters = (*clusterData).Clusters
	}

	for _, curCluster := range relevantClusters {
		// if the cluster has no version it most likely means that the cluster is not connected therefore we skip it
		clusterVersion := convutils.Deref(curCluster.Version.Get(), "")
		if curCluster.Uuid != "" && clusterVersion != "" {
			result[curCluster.Uuid] = clusterVersion
		}
	}

	return result, nil
}
