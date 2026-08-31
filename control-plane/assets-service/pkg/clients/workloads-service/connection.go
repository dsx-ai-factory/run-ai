/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads_service

import (
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"net/url"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/config"
	workloadsServiceClient "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

// serviceConfig holds connection configuration to workloads service
var serviceConfig *workloadsServiceClient.Configuration

// factory is the singleton factory to provide workloads service client
var factory FactoryInterface = Factory{}

// InitConnection int the connection configuration with the workloads service
func InitConnection(urlStr string) {
	serviceUrl, err := url.Parse(urlStr)
	if err != nil {
		logging.Fatalf("failed to connect to workloads-service on '%s': %v", serviceUrl, err)
	}

	serviceConfig = workloadsServiceClient.NewConfiguration()
	serviceConfig.Scheme = serviceUrl.Scheme
	serviceConfig.Host = serviceUrl.Host
	if runai_config.Get[*config.AssetsServiceConfig]().LogLevel == "debug" {
		serviceConfig.Debug = true
	}
}

// SetMock assign the mock implementation of the factory on the singleton factory
func SetMock(clientFactory FactoryInterface) {
	serviceConfig = &workloadsServiceClient.Configuration{}
	factory = clientFactory
}

// Get return the singleton factory
func Get() FactoryInterface {
	return factory
}
