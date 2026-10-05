/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads_manager_client

import (
	"context"

	client_utils "github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/internal"
)

type IWorkloadsManagerExternalClient interface {
	BatchWorkloads(ctx context.Context, batch openapi.WorkloadsV2Batch) (*openapi.BatchResponse, *openapi.Error)

	GetWorkloadV2ById(ctx context.Context, workloadId string) (*openapi.WorkloadV2GetResponse, *openapi.Error)

	UpdateWorkloadV2(ctx context.Context, workloadId string, req openapi.WorkloadV2UpdateRequest) (*openapi.WorkloadV2, *openapi.Error)
}

type IWorkloadsManagerClient interface {
	IWorkloadsManagerExternalClient

	CreateWorkloadV2(ctx context.Context, req openapi.WorkloadV2CreateRequest) (*openapi.WorkloadV2, *openapi.Error)

	UpdateWorkloadInternal(ctx context.Context, workloadId string, req openapi.WorkloadV2UpdateRequest) (*openapi.WorkloadV2, *openapi.Error)
}

func NewWorkloadsManagerClient(baseURL string) IWorkloadsManagerClient {
	clientConfig := openapi.NewConfiguration()
	clientConfig.Servers = openapi.ServerConfigurations{{URL: client_utils.GetInternalUrl(baseURL)}}
	client := openapi.NewAPIClient(clientConfig)
	return &internal.WorkloadsManagerClient{ApiClient: client}
}

func NewWorkloadsManagerExternalClient(baseURL string) IWorkloadsManagerExternalClient {
	return NewWorkloadsManagerClient(client_utils.GetExternalUrl(baseURL))
}
