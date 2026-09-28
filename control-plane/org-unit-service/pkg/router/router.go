/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	authorization_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	cluster_service_client "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/projects"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/reports"
	v1 "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/v1"
	tenants_manager_client "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client"
	workloads_client "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
	middlewares "github.com/dsx-ai-factory/run-ai/runai-application-packages/old-auth-token-middleware"
	datavolumes_service_client "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client"
	logginghttp "github.com/dsx-ai-factory/run-ai/runai-common-packages/logging/http"
	runaimetrics "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/http"
	runaictx "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"
	tracing "github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing/http"
	"github.com/go-chi/chi/v5"
)

func CreateRouter(
	authorizationClient authorization_client.IAuthorizationClient,
	clustersApiService cluster_service_client.IClusterClient,
	tenantsApiService tenants_manager_client.ITenantsManager,
	workloadsClient workloads_client.IWorkloadsClient,
	dvsClient datavolumes_service_client.IDatavolumesClient,
) chi.Router {
	HealthApiService := services.NewHealthApiService()
	HealthApiController := api.NewHealthAPIController(HealthApiService)

	DepartmentApiService := departments.NewDepartmentApiService(clustersApiService, workloadsClient)
	DepartmentApiController := api.NewDepartmentsAPIController(DepartmentApiService)

	projectDeletionService := services.NewProjectDeletionService(clustersApiService, workloadsClient, dvsClient)

	ProjectApiService := projects.NewProjectApiService(projectDeletionService, clustersApiService, workloadsClient)
	ProjectApiController := api.NewProjectsAPIController(ProjectApiService)

	NodeTypeApiService := services.NewNodeTypeApiService()
	NodeTypeApiController := api.NewNodeTypesAPIController(NodeTypeApiService)

	DepartmentV1ApiService := v1.NewDepartmentV1ApiService(clustersApiService)
	DepartmentV1ApiController := api.NewDepartmentsV1APIController(DepartmentV1ApiService)

	ProjectV1ApiService := v1.NewProjectV1ApiService(clustersApiService, projectDeletionService)
	ProjectV1ApiController := api.NewProjectsV1APIController(ProjectV1ApiService)

	OrgUnitApiService := services.NewOrgUnitApiService(authorizationClient, clustersApiService, tenantsApiService, clustersApiService)
	OrgUnitApiController := api.NewOrgUnitAPIController(OrgUnitApiService)

	notificationServiceApiService := services.NewNotificationsApiService(clustersApiService)
	notificationsApiController := api.NewNotificationsAPIController(notificationServiceApiService)

	ReportsApiService := reports.NewReportsApiService()
	ReportsApiController := api.NewReportsAPIController(ReportsApiService)

	SettingsApiService := services.NewSettingsService()
	SettingsApiController := api.NewSettingsAPIController(SettingsApiService)
	// Add routes to the router for each of the controllers
	router := api.NewRouter()

	router.Group(func(router chi.Router) {
		api.AddRoutes(router, HealthApiController)
	})

	router.Group(func(router chi.Router) {
		router.Use(tracing.HttpMiddleware)
		router.Use(runaictx.ContextMiddleware)
		router.Use(runaimetrics.HttpMiddleware)
		router.Use(logginghttp.HttpMiddleware)
		router.Use(middlewares.PreserveJwtToken)
		api.AddRoutes(router, DepartmentApiController, ProjectApiController, NodeTypeApiController,
			OrgUnitApiController, notificationsApiController, ReportsApiController,
			DepartmentV1ApiController, ProjectV1ApiController, SettingsApiController)
	})

	return router
}
