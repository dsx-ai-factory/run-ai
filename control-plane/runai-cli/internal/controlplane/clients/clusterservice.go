/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"
	"fmt"

	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api"
	clierrors "github.com/run-ai/runai/control-plane/runai-cli/internal/errors"
)

const (
	GroupByNode string = "Node"
)

var GroupByList = []string{GroupByNode}

type ClusterService struct {
	apiClient *openapi.APIClient
}

func NewClusterService(urlStr, token string) (*ClusterService, error) {
	conf, err := config.CreateClusterConfig(urlStr, token)
	if err != nil {
		return nil, err
	}

	return &ClusterService{apiClient: openapi.NewAPIClient(conf)}, nil
}

func (cs *ClusterService) GetNodepoolList(ctx context.Context, filters []string) (*openapi.GetNodepools200Response, error) {
	nodepoolList, response, err := cs.apiClient.NodePoolsAPI.GetNodepools(ctx).FilterBy(filters).Execute()
	if err != nil {
		return nil, clierrors.HandleResponseError(response, err)
	}

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	return nodepoolList, nil
}

func (cs *ClusterService) GetNodepoolMinimalList(ctx context.Context, filters []string) (*openapi.GetNodepoolsMinimal200Response, error) {
	nodepoolMinimalList, response, err := cs.apiClient.NodePoolsAPI.GetNodepoolsMinimal(ctx).FilterBy(filters).Execute()
	if err != nil {
		return nil, clierrors.HandleResponseError(response, err)
	}

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	return nodepoolMinimalList, nil
}

func (cs *ClusterService) GetNodeList(ctx context.Context, clusterUuid string) (*openapi.Nodes, error) {
	request := cs.apiClient.NodesAPI.GetTenantNodes(ctx).FilterBy([]string{fmt.Sprintf("clusterUuid==%s", clusterUuid)})
	nodeList, response, err := request.Execute()
	if err != nil {
		return nil, clierrors.HandleResponseError(response, err)
	}

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	return nodeList, nil
}

func (cs *ClusterService) GetNodesTelemetry(ctx context.Context, clusterUuid string, telemetryTypes []openapi.NodeTelemetryType) (api.NodeTelemetriesByType, error) {
	telemetryResponsesByType := make(api.NodeTelemetriesByType)
	for _, telemetryType := range telemetryTypes {
		response, err := cs.getNodesTelemetryByType(ctx, clusterUuid, telemetryType)
		if err != nil {
			return nil, err
		}
		telemetryResponsesByType[telemetryType] = response
	}

	return telemetryResponsesByType, nil
}

func (cs *ClusterService) getNodesTelemetryByType(ctx context.Context, clusterUuid string, telemetryType openapi.NodeTelemetryType) (*openapi.TelemetryResponse, error) {
	telemetry, response, err := cs.apiClient.NodesAPI.
		GetNodeTelemetry(ctx).
		GroupBy(GroupByList).
		TelemetryType(telemetryType).
		ClusterId(clusterUuid).
		Execute()

	if err != nil {
		return nil, clierrors.HandleResponseError(response, err)
	}

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	return telemetry, nil
}
