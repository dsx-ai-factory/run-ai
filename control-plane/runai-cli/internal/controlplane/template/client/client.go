/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"
	clierrors "github.com/run-ai/runai/control-plane/runai-cli/internal/errors"
	wcgenerated "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type Client struct {
	apiClient *wcgenerated.APIClient
}

func NewClient(domain, token string) (*Client, error) {
	workloadConfig, err := config.CreateWorkloadServiceConfig(domain, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create workloads service config: %w", err)
	}

	return &Client{
		apiClient: wcgenerated.NewAPIClient(workloadConfig),
	}, nil
}

func (c *Client) List(ctx context.Context, filterBy []string, includeAncestors bool, limit, offset *int32) (*wcgenerated.TemplatesListResponse, error) {
	request := c.apiClient.WorkloadTemplatesAPI.ListTemplates(ctx)

	if len(filterBy) > 0 {
		request = request.FilterBy(filterBy)
	}

	if includeAncestors {
		request = request.IncludeAncestors(true)
	}

	if limit != nil {
		request = request.Limit(*limit)
	}

	if offset != nil {
		request = request.Offset(*offset)
	}

	response, httpResp, err := request.Execute()
	if err != nil {
		return nil, clierrors.HandleResponseError(httpResp, err)
	}

	if httpResp != nil && httpResp.Body != nil {
		defer httpResp.Body.Close()
	}

	return response, nil
}
