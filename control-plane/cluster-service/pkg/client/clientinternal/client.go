/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	topologylabelscache "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/cache"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
)

type ClusterClient struct {
	ApiClient *openapi.APIClient
}

func (client *ClusterClient) GetNetworkTopologies(ctx context.Context, filterBy []string) (*openapi.GetNetworkTopologies200Response, error) {
	networkTopologies, response, err := client.ApiClient.NetworkTopologiesAPI.
		GetNetworkTopologies(ctx).
		FilterBy(filterBy).
		Execute()
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintf("Failed to get network topologies. status is %s", response.Status))
		return nil, err
	}
	return networkTopologies, nil
}

func (client *ClusterClient) GetClusterInternalConfig(ctx context.Context, clusterUUID string) (*openapi.GetInternalClusterConfigResponse, error) {
	request := client.ApiClient.ClusterConfigAPI.GetInternalClusterConfig(ctx, clusterUUID)
	internalConfig, response, err := request.Execute()
	if err != nil {
		return nil, err
	}

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	if !checkResponseStatusCodeOK(response) {
		return nil, fmt.Errorf("get internal cluster config failed, status code error: %s", response.Status)
	}

	return internalConfig, nil
}

func (client *ClusterClient) GetClusterControlPlaneConfig(ctx context.Context, clusterUUID string) (*openapi.ClusterControlPlaneConfig, error) {
	request := client.ApiClient.ControlPlaneConfigAPI.GetClusterControlPlaneConfig(ctx, clusterUUID)
	cpConfig, response, err := request.Execute()
	if err != nil {
		return nil, err
	}

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	if !checkResponseStatusCodeOK(response) {
		return nil, fmt.Errorf("get internal cluster cp config failed, status code error: %s", response.Status)
	}

	return cpConfig, nil
}

func (client *ClusterClient) CreateCluster(ctx context.Context, clusterCreateRequest *openapi.ClusterCreationRequest) (*openapi.DisplayedCluster, error) {
	request := client.ApiClient.ClustersAPI.CreateCluster(ctx)
	request = request.ClusterCreationRequest(*clusterCreateRequest)
	displayedCluster, response, err := request.Execute()
	if err != nil {
		return nil, err
	}

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	if !checkResponseStatusCodeOK(response) {
		return nil, fmt.Errorf("create cluster config failed, status code error: %s", response.Status)
	}

	return displayedCluster, nil
}

func (client *ClusterClient) GetClusterByUuid(ctx context.Context, id string) (*openapi.DisplayedCluster, *openapi.Error) {
	cachedCluster, err := cache.Get[openapi.DisplayedCluster](ctx, id)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("[ClusterClient.GetClusterByUuid] failed to get openapi.DisplayedCluster with id %s. err: %v", id, err)
	}
	if cachedCluster != nil && cachedCluster.Version.IsSet() {
		return cachedCluster, nil
	}

	clusterRequest := client.ApiClient.ClustersAPI.GetClusterByUuid(ctx, id)
	displayedCluster, response, err := client.ApiClient.ClustersAPI.GetClusterByUuidExecute(clusterRequest)
	if response != nil {
		if response.StatusCode != http.StatusOK {
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				return nil, openapi.NewError(int32(response.StatusCode), fmt.Sprintf("Failed to get cluster %s status is %s", id, response.Status))
			}

			return nil, openapi.NewError(int32(response.StatusCode), string(body))
		}

		if displayedCluster != nil {
			if err = cache.Set[openapi.DisplayedCluster](ctx, id, *displayedCluster); err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("[ClusterClient.GetClusterByUuid] failed to set openapi.DisplayedCluster with id %s. err: %v", id, err)
			}
		}

		return displayedCluster, nil
	}

	if err != nil {
		return nil, openapi.NewError(http.StatusServiceUnavailable, err.Error())
	}

	return nil, openapi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to get cluster %s details", id))
}

func (client *ClusterClient) GetClusters(ctx context.Context) ([]openapi.DisplayedCluster, error) {
	clustersRequest := client.ApiClient.ClustersAPI.GetClusters(ctx)
	displayedClusters, response, err := client.ApiClient.ClustersAPI.GetClustersExecute(clustersRequest)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintf("Failed to get clusters. status is %s", response.Status))
		return nil, err
	}

	return displayedClusters, nil
}

func (client *ClusterClient) GetInternalClusters(ctx context.Context, includeDeleted bool) ([]openapi.DisplayedCluster, error) {
	clustersRequest := client.ApiClient.ClustersAPI.GetInternalClusters(ctx).IncludeDeleted(includeDeleted)
	displayedClusters, response, err := client.ApiClient.ClustersAPI.GetInternalClustersExecute(clustersRequest)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintf("Failed to get clusters. status is %s", response.Status))
		return nil, err
	}

	return displayedClusters, nil
}

func (client *ClusterClient) GetInternalCluster(ctx context.Context, clusterId string) (*openapi.DisplayedCluster, error) {
	clustersRequest := client.ApiClient.ClustersAPI.GetClusterInternal(ctx, clusterId)
	displayedCluster, response, err := client.ApiClient.ClustersAPI.GetClusterInternalExecute(clustersRequest)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintf("Failed to get clusters. status is %s", response.Status))
		return nil, err
	}

	return displayedCluster, nil
}

func (client *ClusterClient) PutClusterConfigStatus(ctx context.Context, clusterUUID string, putClusterConfigStatusRequest *openapi.PutClusterConfigStatusRequest) error {
	request := client.ApiClient.ClusterConfigAPI.PutClusterConfigStatus(ctx, clusterUUID)
	request = request.PutClusterConfigStatusRequest(*putClusterConfigStatusRequest)
	response, err := request.Execute()
	if err != nil {
		return err
	}
	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}
	if !checkResponseStatusCodeOK(response) {
		return fmt.Errorf("put cluster remote config failed, status code error: %s", response.Status)
	}

	return nil
}

func (client *ClusterClient) GetNodePoolsMinimalForTenants(ctx context.Context, tenantIds []string) (*openapi.GetNodepoolsMinimalInternal200Response, error) {
	request := client.ApiClient.NodePoolsAPI.GetNodepoolsMinimalInternal(ctx).TenantId(tenantIds)
	nodePools, response, err := request.Execute()
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintf("Failed to get nodepools. status is %s", response.Status))
		return nil, err
	}
	return nodePools, nil
}

func (client *ClusterClient) GetNodePoolsInternalByCluster(ctx context.Context, clusterId string) (*openapi.GetNodepoolsInternal200Response, error) {
	request := client.ApiClient.NodePoolsAPI.GetNodepoolsInternal(ctx).ClusterId(clusterId)
	nodePools, response, err := request.Execute()
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintf("Failed to get nodepools. status is %s", response.Status))
		return nil, err
	}
	return nodePools, nil
}

func (client *ClusterClient) GetNodePoolsByCluster(ctx context.Context, clusterUUID string) (*openapi.GetNodepools200Response, error) {
	nodePools, response, err := client.ApiClient.NodePoolsAPI.
		GetNodepools(ctx).
		FilterBy([]string{"clusterId==" + clusterUUID}).
		Execute()
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintf("Failed to get nodepools. status is %s", response.Status))
		return nil, err
	}
	return nodePools, nil
}

func (client *ClusterClient) GetClusterMinimalDataById(ctx context.Context, clusterId string) (*openapi.ClustersMinimalData, error) {
	clusterMinimalData, response, err := client.ApiClient.ClustersAPI.GetClusterMinimalDataById(ctx, clusterId).
		Execute()
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintf("Failed to get cluster minimal data. status is %s", response.Status))
		return nil, err
	}
	return clusterMinimalData, nil
}

func (client *ClusterClient) GetClustersMinimalData(ctx context.Context) (*openapi.GetClustersMinimalData200Response, error) {
	clusters, response, err := client.ApiClient.ClustersAPI.GetClustersMinimalData(ctx).
		Execute()
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintf("Failed to get clusters minimal data. status is %s", response.Status))
		return nil, err
	}
	return clusters, nil
}

// minimalNodepoolPageSize matches the API's default page size. A full page
// must be followed by another request, including when the last page is empty.
const minimalNodepoolPageSize int32 = 50

func (client *ClusterClient) GetNodePoolsMinimalByCluster(ctx context.Context, clusterUUID string) (*openapi.GetNodepoolsMinimal200Response, error) {
	result := &openapi.GetNodepoolsMinimal200Response{Nodepools: []openapi.NodepoolMinimal{}}
	for offset := int32(0); ; offset += minimalNodepoolPageSize {
		page, response, err := client.ApiClient.NodePoolsAPI.
			GetNodepoolsMinimal(ctx).
			FilterBy([]string{"clusterId==" + clusterUUID}).
			Offset(offset).
			Limit(minimalNodepoolPageSize).
			Execute()
		if response != nil && response.Body != nil {
			response.Body.Close()
		}
		if err != nil {
			return nil, fmt.Errorf("get nodepools minimal at offset %d: %w", offset, err)
		}
		if response == nil || response.StatusCode != http.StatusOK || page == nil {
			return nil, fmt.Errorf("invalid nodepools minimal response at offset %d", offset)
		}
		result.Nodepools = append(result.Nodepools, page.Nodepools...)
		if len(page.Nodepools) < int(minimalNodepoolPageSize) {
			return result, nil
		}
	}
}

func (client *ClusterClient) GetNodepoolMinimal(ctx context.Context, nodepoolId string) (*openapi.NodepoolMinimal, error) {
	nodePool, response, err := client.ApiClient.NodePoolsAPI.
		GetNodepoolMinimal(ctx, nodepoolId).
		Execute()
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintf("Failed to get nodepool <%s> minimal data. status is %s", nodepoolId, response.Status))
		return nil, err
	}
	return nodePool, nil
}

func (client *ClusterClient) GetClustersCount(ctx context.Context, groupBy string) ([]openapi.ClustersCount, *openapi.Error) {
	countRequest := client.ApiClient.ClustersAPI.GetClustersCount(ctx).GroupBy(groupBy)
	clustersCount, response, err := client.ApiClient.ClustersAPI.GetClustersCountExecute(countRequest)
	if err != nil {
		return nil, openapi.NewError(http.StatusServiceUnavailable, err.Error())
	}
	if response == nil {
		return nil, openapi.NewError(http.StatusServiceUnavailable, "Failed to get parse response of get cluster count request")
	}
	if response.StatusCode != http.StatusOK {
		return nil, openapi.NewError(int32(response.StatusCode), fmt.Sprintf("Failed to get clusters count. status is %s", response.Status))
	}
	return clustersCount, nil
}

func checkResponseStatusCodeOK(response *http.Response) bool {
	return response.StatusCode >= http.StatusOK || response.StatusCode <= http.StatusIMUsed
}

func (client *ClusterClient) GetNodeNetworkTopologyLabelValues(ctx context.Context, clusterUuid string, nodeName string) ([]string, error) {
	tlCache := topologylabelscache.NewNodeNetworkTopologyLabelValuesCache(clusterUuid, nodeName)

	// Try to get from cache first
	cachedLabels, err := tlCache.GetCachedNetworkTopologyLabelValues(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("[ClusterClient.GetNodeNetworkTopologyLabelValues] failed to get NodeNetworkTopologyLabelValues from cache. err: %v", err)
	}
	if cachedLabels != nil {
		return cachedLabels.LabelValues, nil
	}

	// If not in cache, fetch from API
	node, response, err := client.GetNodeInternal(ctx, clusterUuid, nodeName)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get nodeName %s in clusterUuid %s. status is %s", nodeName, clusterUuid, response.Status)
	}

	// Extract topology labels from the node
	if node == nil {
		return nil, fmt.Errorf("node %s not found", nodeName)
	}

	// Cache the topology labels
	labelsToCache := topologylabelscache.NodeNetworkTopologyLabelValues{
		LabelValues: node.NetworkTopologyLabelValues,
	}
	tlCache.SetCachedNetworkTopologyLabelValues(ctx, labelsToCache)

	return node.NetworkTopologyLabelValues, nil
}

func (client *ClusterClient) GetNodeInternal(ctx context.Context, clusterUuid string, nodeName string) (*openapi.NodeInternal, *http.Response, error) {
	return client.ApiClient.NodesAPI.GetNodeInternal(ctx, clusterUuid, nodeName).Execute()
}
