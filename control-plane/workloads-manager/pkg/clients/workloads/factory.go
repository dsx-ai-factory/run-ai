/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"sync"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	ourConfig "github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/config"

	workloadsClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

// ClientInterface lists the workloads services that workloads-manager service needs
type ClientInterface interface {
	ListWorkloadTypes(ctx context.Context, filters []string) (*workloadsApi.ListWorkloadTypes200Response, *workloadsApi.Error)

	// CreateWorkload is a notification to workloads service, informing it about workload creation
	CreateWorkload(ctx context.Context, workload workloadsApi.WorkloadCreationFields) (*workloadsApi.Workload, *workloadsApi.Error)
	CreateWorkloadActions(ctx context.Context, request workloadsApi.CreateWorkloadActionsRequest) *workloadsApi.Error

	CountWorkloads(ctx context.Context, filter []string, search string, includeDeleted bool, precise bool) (int64, *workloadsApi.Error)

	GetWorkload(ctx context.Context, workloadId string) (*workloadsApi.WorkloadDetailed, *workloadsApi.Error)

	PatchWorkload(ctx context.Context, workloadId string, fields workloadsApi.WorkloadPatchFields) *workloadsApi.Error

	ProcessSyncReport(ctx context.Context, clusterId string, report workloadsApi.ClusterSyncReport) *workloadsApi.Error

	GetAvailablePrioritiesAndCategories(ctx context.Context, clusterId *string, sortOrder *string, sortBy *string,
		filters *[]string) (*workloadsApi.GetAvailablePrioritiesAndCategories200Response, *workloadsApi.Error)
}

var once sync.Once
var deFactoClient ClientInterface

// Get the real or mock client
func Get() ClientInterface {
	once.Do(initRealClient)
	return deFactoClient
}

// SetMock is used by unit test to apply mock client instead of the real client
func SetMock(mockClient ClientInterface) {
	deFactoClient = mockClient
}

// initialize the "real" client, invoked once if SetMock was not invoked by tests
func initRealClient() {
	if deFactoClient == nil {
		servicesConfig := config.Get[*ourConfig.Services]()
		deFactoClient = workloadsClient.NewWorkloadsClient(servicesConfig.WorkloadsUrl)
	}
}
