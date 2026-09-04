/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"flag"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"

	auth_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	service_conf "github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/config"
	pubsub "github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/pubsub"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/router"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/settings"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

func init() {
	logging.Init(nil)
	if _, err := metrics.Init(nil); err != nil {
		logging.Errorf("Failed to initialize metrics: %v", err)
	}
	if _, err := tracing.Init(nil); err != nil {
		logging.Errorf("Failed to initialize tracing: %v", err)
	}
	service_conf.Init()

	servicesConfig := config.Get[*service_conf.Services]()

	auth_client.Init(&auth_client.Config{
		Url: servicesConfig.AuthorizationUrl,
	})

	if err := cache.Init(nil); err != nil {
		logging.Errorf("Failed to initialize cache: %v", err)
	}

	services.RegisterSuspendDefCacheStore()

	policies.InitConnection(servicesConfig.PolicyServiceUrl)

	checker.InitClusterCapabilitiesChecker(servicesConfig.ClusterServiceUrl, &checker.ClustersCacheConfig{
		ExpirationSec: servicesConfig.ClusterCacheExpirationSec,
		CleanupSec:    servicesConfig.ClusterCacheCleanupSec,
	}, config.Get[*logging.Config]().LogLevel)

	runaidb.SetLogger(logging.Actual())
	if err := runaidb.Init(nil); err != nil {
		logging.Fatalf("failed to initialize database: %v", err)
	}

	setting.Init(nil, nil)
	if err := setting.Register[*settings.YamlSubmission](); err != nil {
		logging.Errorf("Failed to register YamlSubmission setting: %v", err)
	}

	runaipubsub.SetLogger(logging.Actual())
	_, err := runaipubsub.InitAsync(nil)
	if err != nil {
		logging.Fatalf("Failed to initialize pubsub: %v", err)
	}
}

func main() {
	logging.Infof("Starting Run:AI workloads-manager")

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully waits for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	pubsub.Register()
	defer runaipubsub.Close()

	runaihttp.ServeHTTP(
		router.CreateRouter(),
		nil,
	)
}
