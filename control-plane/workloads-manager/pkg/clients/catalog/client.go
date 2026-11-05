// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package catalog

import (
	"context"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	clientutils "github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils"
	catalog "github.com/dsx-ai-factory/run-ai/runai-clients/catalog-service-client/generated"
)

type Client struct{ api *catalog.APIClient }

func New(baseURL string) *Client {
	cfg := catalog.NewConfiguration()
	cfg.Servers = catalog.ServerConfigurations{{URL: clientutils.GetInternalUrl(baseURL)}}
	cfg.HTTPClient = &http.Client{Timeout: 60 * time.Second}
	return &Client{api: catalog.NewAPIClient(cfg)}
}

func (c *Client) Recommend(ctx context.Context, request catalog.RightSizingRequest) (*catalog.RightSizingResponse, *api.Error) {
	result, response, err := c.api.RightSizingAPI.RecommendInferenceResources(ctx).RightSizingRequest(request).Execute()
	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}
	if err == nil {
		return result, nil
	}
	code := http.StatusServiceUnavailable
	if response != nil {
		code = http.StatusBadGateway
		if response.StatusCode >= http.StatusBadRequest {
			code = response.StatusCode
		}
	}
	message := "failed to retrieve NIM GPU sizing recommendation"
	if typed, ok := err.(*catalog.GenericOpenAPIError); ok {
		if model, ok := typed.Model().(catalog.Error); ok && model.Message != "" {
			message = model.Message
		}
	}
	return nil, api.NewError(code, message, err)
}
