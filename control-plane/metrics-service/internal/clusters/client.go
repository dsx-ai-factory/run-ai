/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clusters

import (
	"context"
	"fmt"
	"time"

	api "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type Client struct {
	client *api.APIClient
}

func NewClustersClient() *Client {
	registerCacheStores()
	conf := config.Get[*ClustersConfiguration]()

	return &Client{
		client: api.NewAPIClient(&api.Configuration{
			Servers: api.ServerConfigurations{{URL: conf.ServiceUrl}},
		}),
	}
}

func (c *Client) GetClustersVersionsHistories(ctx context.Context, clusterIds []string) (map[string]*ClusterVersionHistory, error) {
	clusterVersions, ok := c.getClustersVersionsFromCache(ctx, clusterIds)
	if ok {
		return clusterVersions, nil
	}

	clusters, err := c.getClustersHistories(ctx, clusterIds)
	if err != nil {
		return nil, err
	}

	clusterVersions = c.parseVersionHistories(clusters, clusterIds)
	if len(clusterVersions) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("no cluster versions found for clusters: %v", clusterIds)
	}
	return clusterVersions, nil
}

func (c *Client) GetClusterName(ctx context.Context, clusterId string) (string, error) {
	res, err := cache.Get[api.GetClusterName200Response](ctx, clusterId)
	if err != nil {
		logging.Warnf("failed to get cluster %s name from cache: %v", clusterId, err)
	}
	if res != nil {
		return res.Name, nil
	}

	response, _, err := c.client.ClustersAPI.GetClusterName(ctx, clusterId).Execute()
	if err != nil {
		return "", err
	}

	if err := cache.Set[api.GetClusterName200Response](ctx, clusterId, *response); err != nil {
		logging.Warnf("failed to set cluster name in cache for cluster %s: %v", clusterId, err)
	}

	return response.Name, nil
}

func (c *Client) parseVersionHistories(clusters []api.ClusterHistory, clusterIds []string) map[string]*ClusterVersionHistory {
	sortedVersions := map[string]*ClusterVersionHistory{}
	clusterVersions := c.extractToMapByClusterId(clusters)
	for _, id := range clusterIds {
		cached := NewClusterVersionHistory(clusterVersions[id])

		err := cache.Set[ClusterVersionHistory](context.Background(), id, *cached)
		if err != nil {
			logging.Warnf("failed to set cluster %s version history in cache: %v", id, err)
		}
		sortedVersions[id] = cached
	}
	return sortedVersions
}

func (c *Client) getClustersVersionsFromCache(ctx context.Context, clusterIds []string) (map[string]*ClusterVersionHistory, bool) {
	clusterVersionsMap := map[string]*ClusterVersionHistory{}
	for _, id := range clusterIds {
		v, err := cache.Get[ClusterVersionHistory](ctx, id)
		if v == nil && err == nil {
			return nil, false
		}
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("failed to get cluster %v version history from cache: %v", id, err)
			return nil, false
		}
		clusterVersionsMap[id] = v
	}

	return clusterVersionsMap, true
}

func (c *Client) extractToMapByClusterId(clusters []api.ClusterHistory) map[string][]api.ClusterHistory {
	clusterVersionsMap := map[string][]api.ClusterHistory{}
	for _, c := range clusters {
		if clusterVersionsMap[c.ClusterId] == nil {
			clusterVersionsMap[c.ClusterId] = []api.ClusterHistory{}
		}
		clusterVersionsMap[c.ClusterId] = append(clusterVersionsMap[c.ClusterId], c)
	}

	return clusterVersionsMap
}

func (c *Client) getClustersHistories(ctx context.Context, clusterIds []string) ([]api.ClusterHistory, error) {
	apiRequest := c.client.InternalAPI.GetClustersHistories(ctx).ClusterId(clusterIds)
	clusters, _, err := c.client.InternalAPI.GetClustersHistoriesExecute(apiRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant clusters versions data for clusters: %v, err: %v", clusterIds, err)
	}
	return clusters, nil
}

func registerCacheStores() {
	conf := config.Get[*ClustersConfiguration]()
	if err := cache.RegisterStore[api.GetClusterName200Response](
		options.Ttl(time.Duration(conf.Cache.ClusterNamesTtlMinutes)*time.Minute),
		options.MaxStoreSizeMB(5),
		options.MaxRecordSizeKB(10),
	); err != nil {
		logging.Errorf("failed to create clusters name cache: %v", err)
	}

	if err := cache.RegisterStore[ClusterVersionHistory](
		options.Ttl(time.Duration(conf.Cache.VersionHistoriesTtlMinutes)*time.Minute),
		options.MaxStoreSizeMB(75),
		options.MaxRecordSizeKB(10),
	); err != nil {
		logging.Errorf("failed to create clusters version history cache: %v", err)
	}
}
