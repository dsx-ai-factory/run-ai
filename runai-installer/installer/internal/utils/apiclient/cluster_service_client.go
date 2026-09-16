/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package apiclient

import (
	"context"
	"fmt"
	"time"

	cluster "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	runaiclient "github.com/dsx-ai-factory/run-ai/runai-installer/clients/pkg/clients"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
)

type ClusterServiceClient struct {
	client *runaiclient.RunAIClient
}

func NewClusterServiceClientFromRunAI(client *runaiclient.RunAIClient) *ClusterServiceClient {
	return &ClusterServiceClient{
		client: client,
	}
}

func (c *ClusterServiceClient) GetClusters(ctx context.Context) ([]cluster.DisplayedCluster, int, error) {
	logging.Infof("[ClusterService] Fetching clusters")

	clusters, resp, err := c.client.ClustersAPI().GetClusters(ctx).Execute()
	if err != nil {
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		logging.Errorf("[ClusterService] Failed to get clusters: %v", err)
		return nil, statusCode, fmt.Errorf("failed to get clusters: %w", err)
	}

	logging.Infof("[ClusterService] Successfully fetched %d clusters", len(clusters))
	return clusters, resp.StatusCode, nil
}

func (c *ClusterServiceClient) GetClusterByUUID(ctx context.Context, uuid string) (*cluster.DisplayedCluster, int, error) {
	logging.Infof("[ClusterService] Fetching cluster %s", uuid)

	cluster, resp, err := c.client.ClustersAPI().GetClusterByUuid(ctx, uuid).Execute()
	if err != nil {
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		logging.Errorf("[ClusterService] Failed to get cluster: %v", err)
		return nil, statusCode, fmt.Errorf("failed to get cluster: %w", err)
	}

	logging.Infof("[ClusterService] Successfully fetched cluster %s", uuid)
	return cluster, resp.StatusCode, nil
}

func (c *ClusterServiceClient) CreateCluster(ctx context.Context, req cluster.ClusterCreationRequest) (*cluster.DisplayedCluster, int, error) {
	logging.Infof("[ClusterService] Creating cluster '%s'", req.Name)

	cluster, resp, err := c.client.ClustersAPI().CreateCluster(ctx).ClusterCreationRequest(req).Execute()
	if err != nil {
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		logging.Errorf("[ClusterService] Failed to create cluster: %v", err)
		return nil, statusCode, fmt.Errorf("failed to create cluster: %w", err)
	}

	logging.Infof("[ClusterService] Successfully created cluster '%s' (UUID: %s)", cluster.Name, cluster.Uuid)
	return cluster, resp.StatusCode, nil
}

func (c *ClusterServiceClient) GetClusterInstallInfoByUUID(ctx context.Context, uuid, version, remoteClusterURL string) (*cluster.ClusterInstallationInfoResponse, int, error) {
	logging.Infof("[ClusterService] Fetching install info for cluster %s", uuid)

	request := c.client.ClustersAPI().GetClusterInstallInfoByUuid(ctx, uuid).Version(version)
	if remoteClusterURL != "" {
		request = request.RemoteClusterUrl(remoteClusterURL)
	}

	installInfo, resp, err := request.Execute()
	if err != nil {
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		logging.Errorf("[ClusterService] Failed to get install info: %v", err)
		return nil, statusCode, fmt.Errorf("failed to get install info: %w", err)
	}

	logging.Infof("[ClusterService] Successfully fetched install info for cluster %s", uuid)
	return installInfo, resp.StatusCode, nil
}

func IsWithinLastMinute(timestampStr string) (bool, error) {
	t, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return false, fmt.Errorf("failed to parse timestamp: %w", err)
	}
	oneMinuteAgo := time.Now().Add(-1 * time.Minute)
	return t.After(oneMinuteAgo) && t.Before(time.Now()), nil
}
