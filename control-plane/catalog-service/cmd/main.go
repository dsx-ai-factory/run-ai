/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"flag"
	"time"

	authorizationClient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/clients"
	catalogConfig "github.com/run-ai/runai/control-plane/catalog-service/pkg/config"
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/router"
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/services"
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/settings"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/metrics"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	runaihttp "github.com/run-ai/runai/runai-common-packages/runaihttp"
	"github.com/run-ai/runai/runai-common-packages/setting"
	"github.com/run-ai/runai/runai-common-packages/tracing"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

func init() {
	logging.Init(nil)
	_, err := metrics.Init(nil)
	if err != nil {
		logging.Errorf("Failed to initialize metrics: %v", err)
	}
	_, err = tracing.Init(nil)
	if err != nil {
		logging.Errorf("Failed to initialize tracing: %v", err)
	}
	catalogConfig.Init()

	authorizationClient.Init(&authorizationClient.Config{
		Url: config.Get[*catalogConfig.Services]().AuthorizationUrl,
	})

	if err := cache.Init(nil); err != nil {
		logging.Errorf("Failed to initialize cache: %v", err)
	}
	runaidb.SetLogger(logging.Actual())
	if err := runaidb.Init(nil); err != nil {
		logging.Fatalf("Failed to initialize database: %v", err)
	}

	backendClient := clients.NewBackendClient(config.Get[*catalogConfig.Services]().TempBackendSettingsUrl)
	setting.Init(&setting.SettingConfig{
		BackendClient: backendClient,
	}, nil)

	if err := setting.Register[*settings.EnableNgcIntegration](); err != nil {
		logging.Errorf("Failed to register EnableNgcIntegration setting: %v", err)
	}

	services.InitRegistriesPresets()
	services.InitHelmRepoPresets()
}

func main() {
	logging.Infof("Starting Run:AI catalog-service")
	runSettingsMigration()

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully waits for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	runaihttp.ServeHTTP(
		router.CreateRouter(),
		nil,
	)
}

func runSettingsMigration() {
	if err := setting.Migrate[*settings.EnableNgcIntegration](); err != nil {
		logging.Fatalf("failed to migrate EnableNgcIntegration: %v", err)
	}
	logging.Infof("settings migration completed successfully")
}
