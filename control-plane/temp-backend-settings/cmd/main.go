/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"os"

	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	authorization_openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	bff_openapi "github.com/run-ai/runai/control-plane/bff-service/pkg/client/generated"
	cluster_service_client "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
	cluster_openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	idm_openapi "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	tm_client "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client"
	client_utils "github.com/run-ai/runai/runai-application-packages/client-utils"
	catalog_openapi "github.com/run-ai/runai/runai-clients/catalog-service-client/generated"
	dv_openapi "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
	orgunit_openapi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	wc_openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/metrics"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	runaihttp "github.com/run-ai/runai/runai-common-packages/runaihttp"
	"github.com/run-ai/runai/runai-common-packages/tracing"

	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/config"
	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/dao"
	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/router"
	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/services"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

func init() {
	logging.Init(nil)
	_, _ = metrics.Init(nil)
	_, _ = tracing.Init(nil)
	config.InitConfig()
	runaidb.SetLogger(logging.Actual())
	if err := runaidb.Init(nil); err != nil {
		logging.Fatalf("failed to initialize database: %v", err)
	}
	authorization_client.Init(nil)
}

func main() {
	logging.Infof("Starting temp-backend-settings")

	svcConfig := runai_config.Get[*config.ServicesConfig]()
	clients := initClients(svcConfig)
	settingsDao := dao.NewTenantSettingsDAO()

	tenantService := services.NewTenantAPIService(services.TenantAPIServiceDeps{
		Clients:        clients,
		SettingsDao:    settingsDao,
		ShowAlphaFlags: svcConfig.ShowAlphaFlags,
	})
	internalService := services.NewInternalAPIService(settingsDao)

	runaihttp.ServeHTTP(
		router.CreateRouter(tenantService, internalService),
		nil,
	)
	logging.Infof("Shutting down temp-backend-settings")
	os.Exit(0)
}

func initClients(svcConfig *config.ServicesConfig) *services.Clients {
	clusterAPIConfig := cluster_openapi.NewConfiguration()
	clusterAPIConfig.Servers = cluster_openapi.ServerConfigurations{{URL: client_utils.GetInternalUrl(svcConfig.ClusterServiceURL)}}

	idmConfig := idm_openapi.NewConfiguration()
	idmConfig.Servers = idm_openapi.ServerConfigurations{{URL: svcConfig.IdentityManagerURL}}

	bffConfig := bff_openapi.NewConfiguration()
	bffConfig.Servers = bff_openapi.ServerConfigurations{{URL: svcConfig.BffServiceURL}}

	orgUnitConfig := orgunit_openapi.NewConfiguration()
	orgUnitConfig.Servers = orgunit_openapi.ServerConfigurations{{URL: client_utils.GetInternalUrl(svcConfig.OrgUnitServiceURL)}}

	catalogConfig := catalog_openapi.NewConfiguration()
	catalogConfig.Servers = catalog_openapi.ServerConfigurations{{URL: svcConfig.CatalogServiceURL}}

	dvConfig := dv_openapi.NewConfiguration()
	dvConfig.Servers = dv_openapi.ServerConfigurations{{URL: svcConfig.DatavolumesServiceURL}}

	wcConfig := wc_openapi.NewConfiguration()
	wcConfig.Servers = wc_openapi.ServerConfigurations{{URL: svcConfig.WorkloadsServiceURL}}

	authorizationConfig := authorization_openapi.NewConfiguration()
	authorizationConfig.Servers = authorization_openapi.ServerConfigurations{{URL: svcConfig.AuthorizationServiceURL}}

	return &services.Clients{
		ClusterClient:       cluster_service_client.NewClusterClient(svcConfig.ClusterServiceURL),
		ClusterAPIClient:    cluster_openapi.NewAPIClient(clusterAPIConfig),
		TenantMgrClient:     tm_client.NewTenantsManagerClient(svcConfig.TenantManagerURL),
		IdmClient:           idm_openapi.NewAPIClient(idmConfig),
		BffClient:           bff_openapi.NewAPIClient(bffConfig),
		OrgUnitClient:       orgunit_openapi.NewAPIClient(orgUnitConfig),
		CatalogClient:       catalog_openapi.NewAPIClient(catalogConfig),
		DatavolumesClient:   dv_openapi.NewAPIClient(dvConfig),
		WorkloadsClient:     wc_openapi.NewAPIClient(wcConfig),
		AuthorizationClient: authorization_openapi.NewAPIClient(authorizationConfig),
	}
}
