// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/config"
	clierrors "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/errors"
	catalog "github.com/dsx-ai-factory/run-ai/runai-clients/catalog-service-client/generated"
)

type Client struct {
	apiClient *catalog.APIClient
}

func NewClient(baseURL, token string) (*Client, error) {
	configuration, err := config.CreateCatalogConfig(baseURL, token)
	if err != nil {
		return nil, err
	}
	return &Client{apiClient: catalog.NewAPIClient(configuration)}, nil
}

func (c *Client) Recommend(ctx context.Context, request catalog.RightSizingRequest) (*catalog.RightSizingResponse, error) {
	result, response, err := c.apiClient.RightSizingAPI.RecommendInferenceResources(ctx).RightSizingRequest(request).Execute()
	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return nil, clierrors.HandleResponseError(response, err)
	}
	return result, nil
}
