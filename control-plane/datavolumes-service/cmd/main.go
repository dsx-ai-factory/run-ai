/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"flag"
	"os"
	"time"

	pubsub "github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/pkg/pubsub"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	auth_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/pkg/clients"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/pkg/router"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/pkg/settings"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

func init() {
	logging.Init(nil)
	metrics.Init(nil)
	tracing.Init(nil)
	audit.Init(nil)
	config.Init()
	runaidb.SetLogger(logging.Actual())
	if err := runaidb.Init(nil); err != nil {
		logging.Fatalf("Failed to initialize database: %v", err)
	}
	err := cache.Init(nil)
	if err != nil {
		logging.Errorf("Failed to initialize cache: %v", err)
	}

	auth_client.Init(nil)

	backendClient := clients.NewBackendClient(runai_config.Get[*config.OrgHierarchyManagerConfig]().BackendUrl)
	setting.Init(&setting.SettingConfig{
		BackendClient: backendClient,
	}, nil)

	if err := setting.Register[*settings.EnableDataVolumes](); err != nil {
		logging.Errorf("Failed to register EnableDataVolumes setting: %v", err)
	}

	logging.Infof("Initializing cluster capabilities checker")

	checker.InitClusterCapabilitiesChecker(runai_config.Get[*config.OrgHierarchyManagerConfig]().ClusterServiceUrl, nil, runai_config.Get[*logging.Config]().LogLevel)
	runaipubsub.SetLogger(logging.Actual())
	_, err = runaipubsub.InitAsync(nil)
	if err != nil {
		logging.Fatalf("Failed to initialize pubsub: %v", err)
	}
}

func main() {
	logging.Infof("Starting Run:AI datavolumes-service")
	runSettingsMigration()

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully waits for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	pubsub.Register()
	defer runaipubsub.Close()

	runaihttp.ServeHTTP(
		router.CreateRouter(),
		nil,
	)

	logging.Infof("Shutting down datavolumes-service")
	os.Exit(0)
}

func runSettingsMigration() {
	if err := setting.Migrate[*settings.EnableDataVolumes](); err != nil {
		logging.Fatalf("failed to migrate EnableDataVolumes: %v", err)
	}
	logging.Infof("settings migration completed successfully")
}
