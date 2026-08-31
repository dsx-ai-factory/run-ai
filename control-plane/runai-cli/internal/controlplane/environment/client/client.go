/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/httputils"
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

type Client struct {
	apiClient *openapi.APIClient
}

func NewClient(urlStr, token string) (*Client, error) {
	conf, err := config.CreateAssetsServiceConfig(urlStr, token)
	if err != nil {
		return nil, err
	}
	return &Client{apiClient: openapi.NewAPIClient(conf)}, nil
}

func (c *Client) List(ctx context.Context, projectID *int32, clusterID *string, name *string, workloadType *string) (*openapi.EnvironmentListResponse, error) {
	request := c.apiClient.EnvironmentAPI.ListEnvironmentAssets(ctx)

	if projectID != nil {
		request = request.ProjectId(*projectID)
	}
	if clusterID != nil {
		request = request.ClusterId(*clusterID)
	}
	if name != nil {
		request = request.Name(*name)
	}

	if workloadType != nil {
		request = setWorkloadType(request, *workloadType)
	}

	request = request.IncludeDescendants(true)

	environments, response, err := request.Execute() //nolint:bodyclose // handle in handler
	return httputils.HandleResponse(environments, response, err)
}

func setWorkloadType(request openapi.ApiListEnvironmentAssetsRequest, workloadType string) openapi.ApiListEnvironmentAssetsRequest {
	switch strings.ToLower(workloadType) {
	case "workspace":
		return request.IsWorkspace(true)
	case "training":
		return request.IsTraining(true)
	case "inference", "distributedinference":
		return request.IsInference(true)
	case "distributed":
		return request.IsDistributed(true)
	default:
		return request
	}
}
