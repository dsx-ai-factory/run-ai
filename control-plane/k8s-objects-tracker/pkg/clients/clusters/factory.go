/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clusters

import (
	"context"
	"sync"

	clusterService "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
	clusterApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/config"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
)

type ClientInterface interface {
	GetClusterByUuid(ctx context.Context, id string) (*clusterApi.DisplayedCluster, *clusterApi.Error)
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
		deFactoClient = clusterService.NewCachedClusterClient(servicesConfig.ClusterServiceUrl)
	}
}
