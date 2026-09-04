/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster

import (
	clusterService "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/config"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
)

type FactoryInterface interface {
	Client() ClientInterface
}

var factory FactoryInterface = RealFactory{}

func Get() FactoryInterface {
	return factory
}

type RealFactory struct{}

func (c RealFactory) Client() ClientInterface {
	return &Client{
		apiClient: clusterService.NewClusterClient(runai_config.Get[*config.ServicesConfig]().ClusterServiceUrl),
	}
}

func SetMock(mockFactory FactoryInterface) {
	factory = mockFactory
}
