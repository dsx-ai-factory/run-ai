/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads_service

import (
	"context"
	"fmt"
	"io"
	"net/http"

	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

// Client -> for the "real" client we need an APIClient to the workloads service
type Client struct {
	apiClient *workloadsServiceClient.APIClient
}

// Factory creates the workloads-service client used in production.
type Factory struct{}

// Client returns a workload service client
func (f Factory) Client() ClientInterface {
	return &Client{
		apiClient: workloadsServiceClient.NewAPIClient(serviceConfig),
	}
}

// InternalGetResourceClaimTemplateWorkloads lists the workloads in the request's scope whose pods
// reference the resource claim template with the request's name
func (c Client) InternalGetResourceClaimTemplateWorkloads(ctx context.Context,
	request *workloadsServiceClient.InternalResourceClaimTemplateWorkloadsRequest, offset int32, limit int32,
) (*workloadsServiceClient.ResourceClaimTemplateWorkloadsResponse, *workloadsServiceClient.Error) {
	resp, r, err := c.apiClient.WorkloadsAPI.InternalGetResourceClaimTemplateWorkloads(ctx).
		InternalResourceClaimTemplateWorkloadsRequest(*request).
		Offset(offset).
		Limit(limit).
		Execute()
	if err == nil && r != nil && r.StatusCode == http.StatusOK {
		return resp, nil
	}

	if r != nil && r.StatusCode >= http.StatusBadRequest {
		data, _ := io.ReadAll(r.Body)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to list the workloads of resource claim template %s in %s %s, status %d, data: %s",
			request.Name, request.ScopeType, request.ScopeId, r.StatusCode, data)
		return nil, workloadsServiceClient.NewError(int32(r.StatusCode), fmt.Sprintf("failed to list the workloads of resource claim template %s", request.Name))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed requesting workloads service to list the workloads of resource claim template %s in %s %s: %v",
		request.Name, request.ScopeType, request.ScopeId, err)
	return nil, workloadsServiceClient.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request the workloads of resource claim template %s", request.Name))
}
