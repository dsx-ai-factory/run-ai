/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/config"
	clierrors "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/errors"
	assetsgenerated "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
)

// Client is the thin client wrapper for datasource API operations
type Client struct {
	apiClient *assetsgenerated.APIClient
}

// NewClient creates a new datasource client that wraps the assets API client
func NewClient(domain, token string) (*Client, error) {
	assetsConfig, err := config.CreateAssetsServiceConfig(domain, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create assets service config: %w", err)
	}

	return &Client{
		apiClient: assetsgenerated.NewAPIClient(assetsConfig),
	}, nil
}

// List makes an API call to list datasources with given parameters
// Returns the raw API response - no transformation
func (c *Client) List(ctx context.Context, name *string, projectId *int32, clusterId *string, dsType *string) (*assetsgenerated.DatasourceListResponse, error) {
	request := c.apiClient.DatasourcesAPI.ListDatasourceAssets(ctx)

	if name != nil && *name != "" {
		request = request.Name(*name)
	}

	if projectId != nil {
		request = request.ProjectId(*projectId)
	}

	if clusterId != nil && *clusterId != "" {
		request = request.ClusterId(*clusterId)
	}

	if dsType != nil && *dsType != "" {
		request = request.Type_(*dsType)
	}

	// Always include descendants
	request = request.IncludeDescendants(true)

	response, httpResp, err := request.Execute()
	if err != nil {
		return nil, clierrors.HandleResponseError(httpResp, err)
	}

	if httpResp != nil && httpResp.Body != nil {
		defer httpResp.Body.Close()
	}

	return response, nil
}
