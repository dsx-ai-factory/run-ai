/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	"os"

	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"

	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"
	authorizationClient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/config"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/pubsub"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/services"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/metrics"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaihttp"
	"github.com/run-ai/runai/runai-common-packages/tracing"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

func init() {
	logging.Init(nil)
	metrics.Init(nil)
	tracing.Init(nil)
	audit.Init(nil)
	config.Init()
	servicesConfig := runai_config.Get[*config.ServicesConfig]()

	runaidb.SetLogger(logging.Actual())
	if err := runaidb.Init(nil); err != nil {
		logging.Fatalf("Failed to initialize database: %v", err)
	}
	authorizationClient.Init(servicesConfig.AuthorizationClientConfig())

	if err := cache.Init(nil); err != nil {
		logging.Errorf("Failed to initialize cache: %v", err)
	}

	clusterCacheConf := runai_config.Get[*config.ClusterCache]().CacheConfig
	conf := runai_config.Get[*config.K8sObjectsTrackerConfig]()
	checker.InitClusterCapabilitiesChecker(
		servicesConfig.ClusterServiceUrl,
		&checker.ClustersCacheConfig{ExpirationSec: clusterCacheConf.ExpirationSec, CleanupSec: clusterCacheConf.CleanupSec},
		conf.LogLevel)

	runaipubsub.SetLogger(logging.Actual())
	_, err := runaipubsub.InitAsync(nil)
	if err != nil {
		logging.Fatalf("Failed to initialize pubsub: %v", err)
	}
}

func main() {
	logging.Infof("Starting k8s-objects-tracker service.")

	pubsub.Register()
	defer runaipubsub.Close()

	runaihttp.ServeHTTP(
		services.CreateRouter(),
		nil,
	)

	os.Exit(0)
}
