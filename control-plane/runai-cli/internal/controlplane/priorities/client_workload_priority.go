/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package priorities

import (
	"context"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/httputils"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

const (
	RunAiSource = "Run.ai"
)

type WorkloadPrioritiesClient struct {
	apiClient *openapi.APIClient
}

func NewWorkloadPrioritiesServiceClient(urlStr, token string) (*WorkloadPrioritiesClient, error) {
	conf, err := config.CreateWorkloadServiceConfig(urlStr, token)
	if err != nil {
		return nil, err
	}
	return &WorkloadPrioritiesClient{apiClient: openapi.NewAPIClient(conf)}, nil
}

func (bsc *WorkloadPrioritiesClient) ListPriorities(ctx context.Context, clusterId string) (*openapi.GetWorkloadClusterPriorities200Response, error) {
	fb := filter.NewBuilder().WithSource(RunAiSource)
	request := bsc.apiClient.WorkloadPropertiesAPI.
		GetWorkloadClusterPriorities(ctx).
		FilterBy(fb.StringList()).
		ClusterId(clusterId)

	workloadClusterPriorities, response, err := request.Execute() //nolint:bodyclose // handle in handler
	return httputils.HandleResponse(workloadClusterPriorities, response, err)
}
