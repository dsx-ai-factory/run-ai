package router

import (
	"github.com/go-chi/chi/v5"
	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	cluster_service_client "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/projects"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/reports"
	v1 "github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/v1"
	tenants_manager_client "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client"
	middlewares "github.com/run-ai/runai/runai-application-packages/old-auth-token-middleware"
	datavolumes_service_client "github.com/run-ai/runai/runai-clients/datavolumes-service-client"
	workloads_client "github.com/run-ai/runai/runai-clients/workloads-client"
	logginghttp "github.com/run-ai/runai/runai-common-packages/logging/http"
	runaimetrics "github.com/run-ai/runai/runai-common-packages/metrics/http"
	runaictx "github.com/run-ai/runai/runai-common-packages/runaictx/http"
	tracing "github.com/run-ai/runai/runai-common-packages/tracing/http"
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
