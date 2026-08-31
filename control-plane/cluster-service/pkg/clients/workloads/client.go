/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"

	workloads_client "github.com/run-ai/runai/runai-clients/workloads-client"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type WorkloadsClient interface {
	// InternalUpdateNodePods updates topology information for all pods running on a specific node
	InternalUpdateNodePods(ctx context.Context, clusterId string, nodeName string, networkTopologyLabelValues []string) *openapi.Error
}

type DefaultWorkloadsClient struct {
	client workloads_client.IWorkloadsClient
}

func NewWorkloadsClient(baseURL string) WorkloadsClient {
	return &DefaultWorkloadsClient{
		client: workloads_client.NewWorkloadsClient(baseURL),
	}
}

func (wc *DefaultWorkloadsClient) InternalUpdateNodePods(ctx context.Context, clusterId string, nodeName string, networkTopologyLabelValues []string) *openapi.Error {
	return wc.client.InternalUpdateNodePods(ctx, clusterId, nodeName, networkTopologyLabelValues)
}
