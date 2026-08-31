/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package org_unit

import (
	"context"
	"github.com/run-ai/runai/runai-clients/org-unit-service-client/structs"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"sync"

	"github.com/run-ai/runai/control-plane/policy-service/pkg/config"
	orgUnitService "github.com/run-ai/runai/runai-clients/org-unit-service-client"
	orgUnitApi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
)

// ClientInterface lists the assets services that workloads service needs
type ClientInterface interface {
	GetOrgTreeMap(ctx context.Context) (structs.OrgUnitMap, map[structs.OrgUnitKey]string, *orgUnitApi.Error)
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
		deFactoClient = orgUnitService.NewOrgUnitClient(runai_config.Get[*config.ServicesConfig]().OrgUnitServiceUrl)
	}
}
