/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy

import (
	"context"
	"sync"

	service "github.com/run-ai/runai/runai-clients/workloads-manager-client"
	openapi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
)

// ClientInterface lists the workloads-manager services that workloads service needs
type ClientInterface interface {
	BatchWorkloads(ctx context.Context, batch openapi.WorkloadsV2Batch) (*openapi.BatchResponse, *openapi.Error)

	CreateWorkloadV2(ctx context.Context, req openapi.WorkloadV2CreateRequest) (*openapi.WorkloadV2, *openapi.Error)

	GetWorkloadV2ById(ctx context.Context, workloadId string) (*openapi.WorkloadV2GetResponse, *openapi.Error)

	UpdateWorkloadV2(ctx context.Context, workloadId string, req openapi.WorkloadV2UpdateRequest) (*openapi.WorkloadV2, *openapi.Error)

	UpdateWorkloadInternal(ctx context.Context, workloadId string, req openapi.WorkloadV2UpdateRequest) (*openapi.WorkloadV2, *openapi.Error)
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
		deFactoClient = service.NewWorkloadsManagerClient(runai_config.Get[*config.ServicesConfig]().WorkloadsManagerServiceUrl)
	}
}
