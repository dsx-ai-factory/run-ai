/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	clustercpconfigprovider "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/business/cluster-cp-config/provider"
	access_control "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/business/cluster-infra/access-control"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/business/cluster-infra/access-control/authn"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/business/cluster-infra/access-control/authz"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/cleaners"
	clients "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/clients/backend"
	clients_cache "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/clients/cache"
	workloads_clients "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/clients/workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services"
	status_post_update_handlers "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/cluster-config/status-post-update-handlers"
	cluster_delete "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/cluster-delete"
	cluster_install "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/cluster-installation"
	clusterstatecalculator "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/cluster-state-calculator"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/settings"
	clustersync "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/sync"
	idm_client "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
	tm_client "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client/generated"
	org_unit_client "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/pubsub"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	authorization_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/router"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
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
	metrics.Init(nil)
	tracing.Init(nil)
	audit.Init(nil)

	// Initialize config system
	config.Init("")
	// Initialize database with config
	runaidb.SetLogger(logging.Actual())
	if err := runaidb.Init(nil); err != nil {
		logging.Fatalf("Failed to initialize database: %v", err)
	}
	authorization_client.Init(nil)
	err := cache.Init(nil)
	if err != nil {
		logging.Errorf("Failed to initialize cache: %v", err)
	}
	clients_cache.Init()

	backendConfig := runai_config.Get[*config.Backend]()
	backendClient := clients.NewBackendClient(backendConfig.TempBackendSettingsUrl)
	setting.Init(&setting.SettingConfig{
		BackendClient: backendClient,
	}, nil)
	if err := setting.Register[*settings.EnableAdvancedGPUMetrics](); err != nil {
		logging.Errorf("Failed to register EnableAdvancedGPUMetrics setting: %v", err)
	}
	if err := setting.Register[*settings.ProjectManualOverQuota](); err != nil {
		logging.Errorf("Failed to register ProjectManualOverQuota setting: %v", err)
	}
	if err := setting.Register[*settings.CPUResourcesQuota](); err != nil {
		logging.Errorf("Failed to register CPUResourcesQuota setting: %v", err)
	}

	err = pubsub.Init()
	if err != nil {
		logging.Fatalf("Failed to initialize pubsub: %v", err)
	} else {
		go func() {
			start := time.Now()
			for runaipubsub.GetPubSuber() == nil {
				time.Sleep(10 * time.Second)
				logging.Debugf("Waiting for pubsub to be initialized")
			}
			logging.Infof("Pubsub successfully initialized after %f seconds", time.Since(start).Seconds())
			serviceConfig := runai_config.Get[*config.Service]()
			serviceConfig.PubsubInitialized = true
			if err := runai_config.Set(serviceConfig); err != nil {
				logging.Errorf("Failed to update service config: %v", err)
			}
		}()
	}
}

func main() {
	logging.Infof("Starting cluster service.")
	ctx, cancel := context.WithCancel(context.Background())

	runSettingsMigration()

	clusterServices := initServices(ctx)
	runaihttp.ServeHTTP(
		router.CreateRouter(clusterServices),
		nil,
	)
	defer cancel()
	runaipubsub.Close()
	os.Exit(0)
}

func runSettingsMigration() {
	logging.Infof("running settings migration...")
	var errs []error
	if err := setting.Migrate[*settings.EnableAdvancedGPUMetrics](); err != nil {
		errs = append(errs, fmt.Errorf("EnableAdvancedGPUMetrics: %s", err.Error()))
	}
	if err := setting.Migrate[*settings.ProjectManualOverQuota](); err != nil {
		errs = append(errs, fmt.Errorf("ProjectManualOverQuota: %s", err.Error()))
	}
	if err := setting.Migrate[*settings.CPUResourcesQuota](); err != nil {
		errs = append(errs, fmt.Errorf("CPUResourcesQuota: %s", err.Error()))
	}
	if len(errs) > 0 {
		logging.Fatalf("failed to migrate settings: %v", errors.Join(errs...))
	}
	logging.Infof("settings migration completed successfully")
}

func initServices(ctx context.Context) *services.ClusterServiceServices {
	authConfig := runai_config.Get[*config.Auth]()
	appConfig := runai_config.Get[*config.App]()
	backendConfig := runai_config.Get[*config.Backend]()
	serviceConfig := runai_config.Get[*config.Service]()

	idmClient := idm_client.NewIdmClient(authConfig.IdmServiceUrl)
	orgUnitClient := org_unit_client.NewOrgUnitClient(appConfig.OrgUnitServiceUrl)
	authorizationClient := authorization_client.NewAuthorizationClient()
	tenantSettingsClient := clients.NewTenantSettingsClient(backendConfig.TempBackendSettingsUrl)
	tmClient := tm_client.NewAPIClient(&tm_client.Configuration{
		Servers: tm_client.ServerConfigurations{{URL: appConfig.TenantsManagerServiceUrl}},
	})

	clusterAuthnManager := authn.NewAuthnManager(idmClient)
	clusterAuthzManager := authz.NewAuthzManager(authorizationClient)
	clusterAccessControlManager := access_control.NewClusterAccessControlManager(clusterAuthnManager, clusterAuthzManager)

	healthApiService := services.NewHealthApiService()
	workloadsClient := workloads_clients.NewWorkloadsClient(appConfig.WorkloadsServiceUrl)
	nodesApiService := services.NewNodeService(tenantSettingsClient, workloadsClient)
	nodepoolsApiService := services.NewNodePoolsApiService(tenantSettingsClient, orgUnitClient)
	nodepoolSync := clustersync.NewNodePoolsSync(nodepoolsApiService)
	go nodepoolSync.Start()

	clusterDeleter := cluster_delete.NewClusterDeleter(authorizationClient, orgUnitClient, nodepoolsApiService)
	pubsub.Register(clusterDeleter)

	clusterInstaller := cluster_install.NewClusterInstallService(cluster_install.NewHelmActions())
	clustersApiService := services.NewClustersApiService(
		clusterstatecalculator.NewClusterStateCalculator(), idmClient,
		clusterDeleter, clusterInstaller, tmClient, nodepoolsApiService, orgUnitClient)

	internalApiService := services.NewInternalApiService(clusterAccessControlManager)
	cpConfigProvider := clustercpconfigprovider.NewClusterControlPlaneConfigProvider(clusterAuthnManager, clusterAccessControlManager, tenantSettingsClient, tmClient)
	cpConfigApiService := services.NewControlPlaneConfigApiService(cpConfigProvider)
	postUpdateHandler := status_post_update_handlers.NewCompositeHandler(clusterDeleter)
	clustersApiConfigService := services.NewClusterConfigApiService(postUpdateHandler)
	networkTopologiesService := services.NewNetworkTopologiesApiService()
	settingsApiService := services.NewSettingsService(orgUnitClient)

	clusterCleaner := cleaners.NewClusterCleaner(clusterDeleter,
		serviceConfig.MaxWaitForClusterGracefulDelete,
		serviceConfig.ClusterCleanerRunTiming)
	clusterCleaner.Init(ctx)

	return &services.ClusterServiceServices{
		HealthApiService:         healthApiService,
		NodesApiService:          nodesApiService,
		NodepoolsApiService:      nodepoolsApiService,
		ClusterDeleter:           clusterDeleter,
		ClusterInstaller:         clusterInstaller,
		ClustersApiService:       clustersApiService,
		InternalApiService:       internalApiService,
		CpConfigProvider:         cpConfigProvider,
		CpConfigApiService:       cpConfigApiService,
		ClustersApiConfigService: clustersApiConfigService,
		NetworkTopologiesService: networkTopologiesService,
		SettingsApiService:       settingsApiService,
	}
}
