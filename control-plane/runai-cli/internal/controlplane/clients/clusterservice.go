/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"
	"fmt"
	"net/http"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api"
	clierrors "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/errors"
)

const (
	GroupByNode string = "Node"
	maxPageSize int32  = 500 // the list endpoints' maximum accepted limit
	maxPages           = 500 // bounds the loop if a server ignores offset
)

var GroupByList = []string{GroupByNode}

func fetchAllPages[T any](fetchPage func(offset int32) ([]T, error)) ([]T, error) {
	var all []T
	for page := 0; page < maxPages; page++ {
		items, err := fetchPage(int32(page) * maxPageSize)
		if err != nil {
			return nil, err
		}

		all = append(all, items...)
		if int32(len(items)) < maxPageSize {
			return all, nil
		}
	}

	return nil, fmt.Errorf("stopped after reading %d pages: the server keeps returning full pages", maxPages)
}

func closeResponseBody(response *http.Response) {
	if response != nil && response.Body != nil {
		_ = response.Body.Close()
	}
}

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
	nodepools, err := fetchAllPages(func(offset int32) ([]openapi.Nodepool, error) {
		request := cs.apiClient.NodePoolsAPI.GetNodepools(ctx).
			FilterBy(filters).
			SortBy(openapi.NODEPOOLSORTFILTERFIELDS_NAME).
			Offset(offset).
			Limit(maxPageSize)

		nodepoolList, response, err := request.Execute()
		closeResponseBody(response)
		if err != nil {
			return nil, clierrors.HandleResponseError(response, err)
		}

		return nodepoolList.GetNodepools(), nil
	})
	if err != nil {
		return nil, err
	}

	return &openapi.GetNodepools200Response{Nodepools: nodepools}, nil
}

func (cs *ClusterService) GetNodepoolMinimalList(ctx context.Context, filters []string) (*openapi.GetNodepoolsMinimal200Response, error) {
	nodepools, err := fetchAllPages(func(offset int32) ([]openapi.NodepoolMinimal, error) {
		request := cs.apiClient.NodePoolsAPI.GetNodepoolsMinimal(ctx).
			FilterBy(filters).
			Offset(offset).
			Limit(maxPageSize)

		nodepoolMinimalList, response, err := request.Execute()
		closeResponseBody(response)
		if err != nil {
			return nil, clierrors.HandleResponseError(response, err)
		}

		return nodepoolMinimalList.GetNodepools(), nil
	})
	if err != nil {
		return nil, err
	}

	return &openapi.GetNodepoolsMinimal200Response{Nodepools: nodepools}, nil
}

func (cs *ClusterService) GetNodeList(ctx context.Context, clusterUuid string, offset, limit int32) (*openapi.Nodes, error) {
	filters := []string{fmt.Sprintf("clusterUuid==%s", clusterUuid)}

	// Unordered without sortBy, which would make pages skip and repeat nodes.
	request := cs.apiClient.NodesAPI.GetTenantNodes(ctx).
		FilterBy(filters).
		SortBy(openapi.NODEFILTERSORTFIELDS_NAME).
		Offset(offset).
		Limit(limit)

	nodeList, response, err := request.Execute()
	closeResponseBody(response)
	if err != nil {
		return nil, clierrors.HandleResponseError(response, err)
	}

	return &openapi.Nodes{Nodes: nodeList.GetNodes()}, nil
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
