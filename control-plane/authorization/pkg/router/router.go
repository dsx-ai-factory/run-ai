/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	authorization_config "github.com/run-ai/runai/control-plane/authorization/pkg/config"
	"github.com/run-ai/runai/control-plane/authorization/pkg/services"
	cluster_service_client "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/config"
	logginghttp "github.com/run-ai/runai/runai-common-packages/logging/http"
	runaimetrics "github.com/run-ai/runai/runai-common-packages/metrics/http"
	runaictx "github.com/run-ai/runai/runai-common-packages/runaictx/http"
	tracing "github.com/run-ai/runai/runai-common-packages/tracing/http"
)

func CreateRouter() chi.Router {
	HealthApiService := services.NewHealthApiService()
	HealthApiController := api.NewHealthAPIController(HealthApiService)

	tenantSettings := services.NewTenantSettings()

	clustersClient := cluster_service_client.NewAPIClient(&cluster_service_client.Configuration{
		Servers: cluster_service_client.ServerConfigurations{
			{
				URL: config.Get[*authorization_config.Server]().ClusterServiceUrl,
			},
		},
	})

	RolesApiService := services.NewRolesApiService(clustersClient, tenantSettings)
	RolesApiController := api.NewRolesAPIController(RolesApiService)

	AccessRulesApiService := services.NewAccessRulesApiService(tenantSettings)
	AccessRulesApiController := api.NewAccessRulesAPIController(AccessRulesApiService)

	InternalApiService := services.NewInternalApiService(tenantSettings)
	InternalApiController := api.NewInternalAPIController(InternalApiService)

	PermissionsApiService := services.NewPermissionsApiService()
	PermissionsApiController := api.NewPermissionsAPIController(PermissionsApiService)

	MeApiService := services.NewMeApiService()
	MeApiController := api.NewMeAPIController(MeApiService)

	SettingsApiService := services.NewSettingsService()
	SettingsApiController := api.NewSettingsAPIController(SettingsApiService)

	router := api.NewRouter()

	router.Group(func(router chi.Router) {
		router.Use(tracing.HttpMiddleware)
		router.Use(runaictx.ContextMiddleware)
		router.Use(runaimetrics.HttpMiddleware)
		router.Use(logginghttp.HttpMiddleware)
		// Add routes to the router for each of the controllers
		api.AddRoutes(router,
			HealthApiController,
			RolesApiController,
			AccessRulesApiController,
			InternalApiController,
			PermissionsApiController,
			MeApiController,
			SettingsApiController,
		)
	})

	return router
}
