/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	clients "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/backend"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/setting_cache"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	runaictx "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"
	"github.com/go-chi/chi/v5"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
)

func CreateRouter() chi.Router {
	tenantSettingsClient := clients.NewTenantSettingsClient(runai_config.Get[*config.ServicesConfig]().TempBackendSettingsUrl)
	return createRouterInt(tenantSettingsClient)
}

func createRouterInt(tenantSettingsClient clients.TenantSettingsClient) chi.Router {
	setting_cache.RegisterAdvancedGpuMetricsStore()

	WorkspacesApiService := NewWorkspacesService(tenantSettingsClient)
	WorkspacesAPIController := api.NewWorkspacesAPIController(WorkspacesApiService)

	TrainingsApiService := NewTrainingsService(tenantSettingsClient)
	TrainingsAPIController := api.NewTrainingsAPIController(TrainingsApiService)

	DistributedApiService := NewDistributedService(tenantSettingsClient)
	DistributedAPIController := api.NewDistributedAPIController(DistributedApiService)

	InferencesApiService := NewInferencesService(tenantSettingsClient)
	InferencesAPIController := api.NewInferencesAPIController(InferencesApiService)

	WorkloadsBatchService := NewWorkloadsBatchAPIService()
	WorkloadsBatchAPIController := api.NewWorkloadsBatchAPIController(WorkloadsBatchService)

	UnifiedWorkloadService := NewUnifiedAPIService()
	UnifiedWorkloadAPIController := api.NewUnifiedAPIController(UnifiedWorkloadService)

	TemplatesService := NewTemplatesService(tenantSettingsClient)
	TemplatesAPIController := api.NewWorkloadTemplatesAPIController(TemplatesService)

	// Add routes to the router for each of the controllers
	router := api.NewRouter()

	router.Group(func(router chi.Router) {
		router.Use(runaictx.ContextMiddleware)

		api.AddRoutes(router, WorkloadsBatchAPIController, WorkspacesAPIController, TrainingsAPIController, DistributedAPIController,
			InferencesAPIController, UnifiedWorkloadAPIController, TemplatesAPIController)
	})

	return router
}
