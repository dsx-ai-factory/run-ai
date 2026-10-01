/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads_client

import (
	"context"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/internal"
	client_utils "github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils"
)

type IWorkloadsExternalClient interface {
}

type IWorkloadsClient interface {
	IWorkloadsExternalClient
	GetAncestors(ctx context.Context, resourceId string) (*openapi.Ancestors, error)
	GetInferenceAncestors(ctx context.Context, resourceId string) (*openapi.Ancestors, error)
	GetWorkloads(ctx context.Context, limit, offset int32, filter []string, includeDeleted bool, sortBy string, sortOrder string) (*openapi.GetWorkloads200Response, error)
	GetWorkload(ctx context.Context, workloadId string) (*openapi.WorkloadDetailed, *openapi.Error)
	CountWorkloads(ctx context.Context, filter []string, search string, includeDeleted bool, precise bool) (int64, *openapi.Error)

	ListWorkloadTypes(ctx context.Context, filters []string) (*openapi.ListWorkloadTypes200Response, *openapi.Error)
	GetAvailablePrioritiesAndCategories(ctx context.Context, clusterId *string, sortOrder *string, sortBy *string, filters *[]string) (*openapi.GetAvailablePrioritiesAndCategories200Response, *openapi.Error)

	// CreateWorkload is a notification to workloads service, informing it about workload creation
	CreateWorkload(ctx context.Context, workload openapi.WorkloadCreationFields) (*openapi.Workload, *openapi.Error)

	PatchWorkload(ctx context.Context, workloadId string, fields openapi.WorkloadPatchFields) *openapi.Error

	ProcessSyncReport(ctx context.Context, clusterId string, report openapi.ClusterSyncReport) *openapi.Error

	// InternalEnrichOrgUnitRanking enriches org unit ranking items with workload counts
	InternalEnrichOrgUnitRanking(ctx context.Context, request openapi.InternalEnrichOrgUnitRankingRequest) (*openapi.InternalEnrichOrgUnitRankingResponse, error)

	// InternalUpdateNodePods updates topology information for all pods running on a specific node
	InternalUpdateNodePods(ctx context.Context, clusterId string, nodeName string, networkTopologyLabelValues []string) *openapi.Error
}

func NewWorkloadsClient(baseURL string) IWorkloadsClient {
	clientConfig := openapi.NewConfiguration()
	clientConfig.Servers = openapi.ServerConfigurations{{URL: client_utils.GetInternalUrl(baseURL)}}
	client := openapi.NewAPIClient(clientConfig)
	return &internal.WorkloadsClient{ApiClient: client}
}

func NewWorkloadsExternalClient(baseURL string) IWorkloadsExternalClient {
	return NewWorkloadsClient(client_utils.GetExternalUrl(baseURL))
}
