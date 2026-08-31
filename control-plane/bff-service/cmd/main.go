/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/*
 * BFF Service - Backend for Frontend Service
 * API version: 1.0.0
 * Contact: support@run.ai
 */

package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	"github.com/run-ai/runai/control-plane/bff-service/internal/auth"
	"github.com/run-ai/runai/control-plane/bff-service/internal/config"
	"github.com/run-ai/runai/control-plane/bff-service/internal/router"
	"github.com/run-ai/runai/control-plane/bff-service/pkg/clients"
	"github.com/run-ai/runai/control-plane/bff-service/pkg/settings/branding"
	"github.com/run-ai/runai/control-plane/bff-service/pkg/settings/notifications"
	ui_appearance "github.com/run-ai/runai/control-plane/bff-service/pkg/settings/ui-appearance"
	ui_workloads "github.com/run-ai/runai/control-plane/bff-service/pkg/settings/ui-workloads"
	system_notification "github.com/run-ai/runai/control-plane/bff-service/pkg/system-notification"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/metrics"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaihttp"
	"github.com/run-ai/runai/runai-common-packages/setting"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

var logoMigrator *branding.LogoMigrator

func init() {
	logging.Init(nil)
	metrics.Init(nil)
	config.InitConfig()
	runaidb.SetLogger(logging.Actual())
	if err := runaidb.Init(nil); err != nil {
		logging.Fatalf("failed to initialize database: %v", err)
	}
	initCache()

	authorization_client.Init(nil)
	if runai_config.Get[*config.Configuration]().UseMocks {
		clients.SetMockBackendClient(&clients.MockBackendClient{})
		clients.SetMockTenantsManagerClient(&clients.MockTenantsManagerClient{})
		authorization_client.SetMockAuthorizationClient(auth.MockService{})
	}
	backendConfig := runai_config.Get[*config.Backend]()
	backendClient := clients.NewBackendClient(backendConfig.BackendUrl)
	tmClient := clients.NewTenantsManagerClient(backendConfig.TenantManagerUrl)
	logoMigrator = branding.NewLogoMigrator(tmClient)

	setting.Init(&setting.SettingConfig{
		BackendClient: backendClient,
	}, nil)

	setting.Register[*notifications.AdminMessage]()
	setting.Register[*ui_workloads.EnableNIM]()
	setting.Register[*ui_workloads.DeletedWorkloadsView]()
	setting.Register[*ui_workloads.AllowGPT2Model]()
	setting.Register[*ui_workloads.EnableModelCatalog]()
	setting.Register[*ui_appearance.BannerConfig]()
	setting.Register[*branding.Logo]()
}

func main() {
	runSettingsMigration()
	serveHTTP()
}

func serveHTTP() {
	serviceConfig := runai_config.Get[*config.Service]()
	bootConfig := runai_config.Get[*config.BootConfig]()
	logging.Infof("serving HTTP for Run:ai version %s", serviceConfig.RunaiVersion)
	runaihttp.ServeHTTP(
		router.CreateRouter(serviceConfig, bootConfig),
		nil,
	)
	os.Exit(0)
}

func runSettingsMigration() {
	logging.Infof("running settings migration...")
	var errs []error
	if err := setting.Migrate[*notifications.AdminMessage](); err != nil {
		errs = append(errs, fmt.Errorf("adminMessage: %s", err.Error()))
	}
	if err := setting.Migrate[*ui_workloads.EnableNIM](); err != nil {
		errs = append(errs, fmt.Errorf("enableNIM: %s", err.Error()))
	}
	if err := setting.Migrate[*ui_workloads.DeletedWorkloadsView](); err != nil {
		errs = append(errs, fmt.Errorf("deletedWorkloadsView: %s", err.Error()))
	}
	if err := setting.Migrate[*ui_workloads.AllowGPT2Model](); err != nil {
		errs = append(errs, fmt.Errorf("allowGPT2Model: %s", err.Error()))
	}
	if err := setting.Migrate[*ui_workloads.EnableModelCatalog](); err != nil {
		errs = append(errs, fmt.Errorf("enableModelCatalog: %s", err.Error()))
	}
	if err := logoMigrator.Migrate(); err != nil {
		errs = append(errs, fmt.Errorf("logo: %s", err.Error()))
	}

	if len(errs) > 0 {
		logging.Fatalf("failed to migrate settings: %v", errors.Join(errs...))
	}
	logging.Infof("settings migration completed successfully")
}

func initCache() {
	cache.SetLogger(logging.Actual())
	if _, err := cache.InitAsync(nil); err != nil {
		logging.Errorf("failed to initialize cache: %v", err)
		return
	}

	cache.RegisterStoreAsync[system_notification.Message](options.Ttl(24 * 7 * time.Hour))
}
