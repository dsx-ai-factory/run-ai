/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package org_unit

import (
	"context"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"sync"

	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/config"
	orgUnitService "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client"

	orgUnitApi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
)

// ClientInterface lists the assets services that workloads service needs
type ClientInterface interface {
	GetProjectsMetadataByCluster(ctx context.Context, clusterId string) ([]orgUnitApi.ProjectMetadata, *orgUnitApi.Error)

	GetAncestorsOfScope(ctx context.Context, scopeType orgUnitApi.ScopeType, scopeId string) (*orgUnitApi.Ancestors, *orgUnitApi.Error)
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
		servicesConfig := runai_config.Get[*config.ServicesConfig]()
		deFactoClient = orgUnitService.NewOrgUnitClient(servicesConfig.OrgUnitUrl)
	}
}
