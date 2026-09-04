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

type Client struct {
	apiClient *assetsgenerated.APIClient
}

func NewClient(domain, token string) (*Client, error) {
	assetsConfig, err := config.CreateAssetsServiceConfig(domain, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create assets service config: %w", err)
	}

	return &Client{
		apiClient: assetsgenerated.NewAPIClient(assetsConfig),
	}, nil
}

func (c *Client) List(ctx context.Context, name *string, projectId *int32, clusterId *string, credType *string) (*assetsgenerated.CredentialsListResponse, error) {
	request := c.apiClient.CredentialsAPI.ListCredentialsAssets(ctx)

	if name != nil && *name != "" {
		request = request.Name(*name)
	}

	if projectId != nil {
		request = request.ProjectId(*projectId)
	}

	if clusterId != nil && *clusterId != "" {
		request = request.ClusterId(*clusterId)
	}

	if credType != nil && *credType != "" {
		request = request.Type_(*credType)
	}

	request = request.IncludeDescendants(true)
	request = request.StatusInfo(true)

	response, httpResp, err := request.Execute()
	defer func() {
		if httpResp != nil && httpResp.Body != nil {
			httpResp.Body.Close()
		}
	}()

	if err != nil {
		return nil, clierrors.HandleResponseError(httpResp, err)
	}

	return response, nil
}
