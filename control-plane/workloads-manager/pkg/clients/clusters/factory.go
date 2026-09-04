/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clusters

import (
	"sync"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	ourConfig "github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
)

var (
	once          sync.Once
	deFactoClient clusterservice.ClientInterface
)

// Get the real or mock client
func Get() clusterservice.ClientInterface {
	once.Do(initRealClient)
	return deFactoClient
}

// SetMock is used by unit test to apply mock client instead of the real client
func SetMock(mockClient clusterservice.ClientInterface) {
	deFactoClient = mockClient
}

// initialize the "real" client, invoked once if SetMock was not invoked by tests
func initRealClient() {
	if deFactoClient == nil {
		servicesConfig := config.Get[*ourConfig.Services]()
		deFactoClient = clusterservice.GetRealClient(servicesConfig.ClusterServiceUrl)
	}
}
