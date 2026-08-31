/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster

import (
	"net/url"

	clusterService "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

var serverConfig *clusterService.Configuration

var factory FactoryInterface = Factory{}

func InitConnection(clusterServiceUrlStr string, logLevel string) {
	clusterServiceUrl, err := url.Parse(clusterServiceUrlStr)
	if err != nil {
		logging.Fatalf("failed to connect to cluster-service on '%s': %v", clusterServiceUrl, err)
	}

	serverConfig = clusterService.NewConfiguration()
	serverConfig.Scheme = clusterServiceUrl.Scheme
	serverConfig.Host = clusterServiceUrl.Host
	if logLevel == "debug" {
		serverConfig.Debug = true
	}
}

func SetMock(clientFactory FactoryInterface) {
	serverConfig = &clusterService.Configuration{}
	factory = clientFactory
}

func Get() FactoryInterface {
	return factory
}
