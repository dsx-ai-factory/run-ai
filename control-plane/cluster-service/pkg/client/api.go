/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster_service_client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	cache2 "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/cache"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/clientinternal"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	client_utils "github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

const (
	authorizationHeader       = "Authorization"
	bearerTokenStr            = "Bearer %s"
	DisplayedClusterStoreName = "clusters-client-displayed-cluster"
	TopologyLabelsStoreName   = "clusters-client-topology-labels"
)

type IClusterExternalClient interface {
	GetClusterInternalConfig(ctx context.Context, clusterUUID string) (*openapi.GetInternalClusterConfigResponse, error)
	GetClusterControlPlaneConfig(ctx context.Context, clusterUUID string) (*openapi.ClusterControlPlaneConfig, error)
	CreateCluster(ctx context.Context, clusterCreateRequest *openapi.ClusterCreationRequest) (*openapi.DisplayedCluster, error)
	GetClusterByUuid(ctx context.Context, id string) (*openapi.DisplayedCluster, *openapi.Error)
	GetClusters(ctx context.Context) ([]openapi.DisplayedCluster, error)
	PutClusterConfigStatus(ctx context.Context, clusterUUID string, putClusterConfigStatusRequest *openapi.PutClusterConfigStatusRequest) error
	GetNodePoolsByCluster(ctx context.Context, clusterUUID string) (*openapi.GetNodepools200Response, error)
	GetNodePoolsMinimalByCluster(ctx context.Context, clusterUUID string) (*openapi.GetNodepoolsMinimal200Response, error)
	GetNodepoolMinimal(ctx context.Context, nodepoolId string) (*openapi.NodepoolMinimal, error)
	GetClustersMinimalData(ctx context.Context) (*openapi.GetClustersMinimalData200Response, error)
	GetClusterMinimalDataById(ctx context.Context, clusterId string) (*openapi.ClustersMinimalData, error)
	GetNetworkTopologies(ctx context.Context, filterBy []string) (*openapi.GetNetworkTopologies200Response, error)
	GetNodeNetworkTopologyLabelValues(ctx context.Context, clusterUuid string, nodeName string) ([]string, error)
	GetNodeInternal(ctx context.Context, clusterUuid string, nodeName string) (*openapi.NodeInternal, *http.Response, error)
}

type IClusterClient interface {
	IClusterExternalClient
	GetNodePoolsMinimalForTenants(ctx context.Context, tenantIds []string) (*openapi.GetNodepoolsMinimalInternal200Response, error)
	GetNodePoolsInternalByCluster(ctx context.Context, clusterId string) (*openapi.GetNodepoolsInternal200Response, error)
	GetClustersCount(ctx context.Context, groupBy string) ([]openapi.ClustersCount, *openapi.Error)
	GetInternalClusters(ctx context.Context, includeDeleted bool) ([]openapi.DisplayedCluster, error)
	GetInternalCluster(ctx context.Context, clusterId string) (*openapi.DisplayedCluster, error)
}

func NewClusterExternalClient(baseURL string, token string) IClusterExternalClient {
	config := clusterClientConfig(baseURL, true, &token)
	return &internal.ClusterClient{
		ApiClient: openapi.NewAPIClient(config),
	}
}

func NewClusterClient(baseURL string) IClusterClient {
	clientConfig := openapi.NewConfiguration()
	clientConfig.Servers = openapi.ServerConfigurations{{URL: client_utils.GetInternalUrl(baseURL)}}
	client := openapi.NewAPIClient(clientConfig)
	return &internal.ClusterClient{ApiClient: client}
}

func NewCachedClusterClient(baseURL string) IClusterClient {
	if err := cache.Init(nil); err != nil {
		logging.Errorf("Failed to initialize cache: %v", err)
	}

	if err := InitSharedClusterClientCache(); err != nil {
		logging.Warnf("Failed to register shared cache store: %v", err)
	}

	config := clusterClientConfig(baseURL, false, nil)
	return &internal.ClusterClient{
		ApiClient: openapi.NewAPIClient(config),
	}
}

func InitSharedClusterClientCache() error {
	var errs error
	if err := cache.RegisterSharedStore[openapi.DisplayedCluster](DisplayedClusterStoreName, options.Ttl(10*time.Minute), options.MaxRecordSizeKB(20), options.MaxStoreSizeMB(200)); err != nil {
		errs = errors.Join(errs, fmt.Errorf("failed to register cache store for type DisplayedCluster: %w", err))
	}

	if err := cache.RegisterSharedStore[cache2.NodeNetworkTopologyLabelValues](TopologyLabelsStoreName, options.Ttl(10*time.Minute), options.MaxRecordSizeKB(1), options.MaxStoreSizeMB(50)); err != nil {
		errs = errors.Join(errs, fmt.Errorf("failed to register cache store for type NodeNetworkTopologyLabelValues: %w", err))
	}
	return errs
}

func clusterClientConfig(baseURL string, external bool, token *string) *openapi.Configuration {
	configuration := openapi.NewConfiguration()
	if external {
		configuration.Servers = openapi.ServerConfigurations{{URL: client_utils.GetExternalUrl(baseURL)}}
	} else {
		configuration.Servers = openapi.ServerConfigurations{{URL: client_utils.GetInternalUrl(baseURL)}}
	}

	if token != nil {
		configuration.AddDefaultHeader(authorizationHeader, fmt.Sprintf(bearerTokenStr, *token))
	}
	return configuration
}

func NewClusterServiceClientWithConfig(config *openapi.Configuration) IClusterClient {
	return &internal.ClusterClient{ApiClient: openapi.NewAPIClient(config)}
}

func NewClusterServiceExternalClientWithConfig(config *openapi.Configuration) IClusterExternalClient {
	return NewClusterServiceClientWithConfig(config)
}
