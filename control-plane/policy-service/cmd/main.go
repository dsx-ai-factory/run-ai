/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"os"

	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"
	authorizationClient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/config"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/pubsub"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/services"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/orgunit"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/metrics"
	"github.com/run-ai/runai/runai-common-packages/runaihttp"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
	"github.com/run-ai/runai/runai-common-packages/tracing"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

func init() {
	logging.Init(nil)
	metrics.Init(nil)
	tracing.Init(nil)
	audit.Init(nil)
	authorizationClient.Init(nil)

	config.Init()
	servicesConfig := runai_config.Get[*config.ServicesConfig]()

	dao_utils.Init(config.Database())

	dao.HandleSystemPolicy()

	services.HandleSystemMapping()

	orgunit.InitConnection(servicesConfig.OrgUnitServiceUrl)

	runaipubsub.SetLogger(logging.Actual())
	_, err := runaipubsub.InitAsync(nil)
	if err != nil {
		logging.Fatalf("Failed to initialize pubsub: %v", err)
	}
}

func main() {
	logging.Infof("policy-service has started")
	pubsub.Register()

	stopChan := make(chan struct{})

	runaihttp.ServeHTTP(
		services.CreateRouter(),
		nil,
	)

	stopChan <- struct{}{}
	close(stopChan)

	runaipubsub.Close()
	os.Exit(0)
}
