/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"os"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	authorizationClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/pubsub"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

func init() {
	logging.Init(nil)
	metrics.Init(nil)
	tracing.Init(nil)
	audit.Init(nil)
	authorizationClient.Init(nil)

	config.Init()
	servicesConfig := runai_config.Get[*config.ServicesConfig]()

	runaidb.SetLogger(logging.Actual())
	if err := runaidb.Init(nil); err != nil {
		logging.Fatalf("Failed to initialize database: %v", err)
	}

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
