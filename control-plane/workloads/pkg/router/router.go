package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	clients "github.com/run-ai/runai/control-plane/workloads/pkg/clients/backend"
	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	"github.com/run-ai/runai/control-plane/workloads/pkg/metrics_exporter"
	"github.com/run-ai/runai/control-plane/workloads/pkg/services"
	"github.com/run-ai/runai/control-plane/workloads/pkg/services_crud"
	"github.com/run-ai/runai/control-plane/workloads/pkg/setting_cache"
	middlewares "github.com/run-ai/runai/runai-application-packages/old-auth-token-middleware"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	logginghttp "github.com/run-ai/runai/runai-common-packages/logging/http"
	metrics_pkg "github.com/run-ai/runai/runai-common-packages/metrics"
	runaimetrics "github.com/run-ai/runai/runai-common-packages/metrics/http"
	runaictx "github.com/run-ai/runai/runai-common-packages/runaictx/http"
	tracing "github.com/run-ai/runai/runai-common-packages/tracing/http"
)

// CreateRouter sets up the HTTP router with all workloads-service routes and middleware.
func CreateRouter() chi.Router {
	tenantSettingsClient := clients.NewTenantSettingsClient(runai_config.Get[*config.ServicesConfig]().TempBackendSettingsUrl)

	setting_cache.RegisterAdvancedGpuMetricsStore()

	HealthApiService := NewHealthApiService()
	HealthAPIController := api.NewHealthAPIController(HealthApiService)

	WorkspacesApiService := services_crud.NewWorkspacesService(tenantSettingsClient)
	WorkspacesAPIController := api.NewWorkspacesAPIController(WorkspacesApiService)

	TrainingsApiService := services_crud.NewTrainingsService(tenantSettingsClient)
	TrainingsAPIController := api.NewTrainingsAPIController(TrainingsApiService)

	DistributedApiService := services_crud.NewDistributedService(tenantSettingsClient)
	DistributedAPIController := api.NewDistributedAPIController(DistributedApiService)

	InferencesApiService := services_crud.NewInferencesService(tenantSettingsClient)
	InferencesAPIController := api.NewInferencesAPIController(InferencesApiService)

	DistributedInferencesApiService := services_crud.NewDistributedInferenceService(tenantSettingsClient)
	DistributedInferencesAPIController := api.NewDistributedInferencesAPIController(DistributedInferencesApiService)

	WorkloadsBatchService := services_crud.NewWorkloadsBatchAPIService()
	WorkloadsBatchAPIController := api.NewWorkloadsBatchAPIController(WorkloadsBatchService)

	UnifiedWorkloadService := services_crud.NewUnifiedAPIService()
	UnifiedWorkloadAPIController := api.NewUnifiedAPIController(UnifiedWorkloadService)

	WorkloadsApiService := services.NewWorkloadsService()
	WorkloadsAPIController := api.NewWorkloadsAPIController(WorkloadsApiService)

	WorkloadPropertiesApiService := services.NewWorkloadPropertiesService()
	WorkloadPropertiesAPIController := api.NewWorkloadPropertiesAPIController(WorkloadPropertiesApiService)

	PodsApiService := services.NewPodsService(tenantSettingsClient)
	PodsAPIController := api.NewPodsAPIController(PodsApiService)

	PodGroupsApiService := services.NewPodGroupsService()
	PodGroupsAPIController := api.NewPodGroupsAPIController(PodGroupsApiService)

	EventsApiService := services.NewEventsService()
	EventsAPIController := api.NewEventsAPIController(EventsApiService)

	RevisionsApiService := services.NewRevisionsService()
	RevisionsAPIController := api.NewRevisionsAPIController(RevisionsApiService)

	TemplatesService := services_crud.NewTemplatesService(tenantSettingsClient)
	TemplatesAPIController := api.NewWorkloadTemplatesAPIController(TemplatesService)

	ApplicationsService := services.NewAIApplicationsService()
	ApplicationsAPIController := api.NewAIApplicationsAPIController(ApplicationsService)

	SyncService := services_crud.NewSyncService()
	SyncController := api.NewSyncAPIController(SyncService)

	settingsApiService := services.NewSettingsService()
	settingsApiController := api.NewSettingsAPIController(settingsApiService)
	router := api.NewRouter()

	router.Group(func(router chi.Router) {
		api.AddRoutes(router, HealthAPIController)
	})

	router.Group(func(router chi.Router) {
		router.Use(tracing.HttpMiddleware)
		router.Use(runaictx.ContextMiddleware)
		router.Use(runaimetrics.HttpMiddleware)
		router.Use(logginghttp.HttpMiddleware)
		router.Use(middlewares.PreserveJwtToken)
		api.AddRoutes(router, WorkloadsBatchAPIController, WorkspacesAPIController, TrainingsAPIController, DistributedAPIController,
			InferencesAPIController, DistributedInferencesAPIController, WorkloadsAPIController, UnifiedWorkloadAPIController, PodsAPIController, EventsAPIController, RevisionsAPIController, PodGroupsAPIController,
			TemplatesAPIController, WorkloadPropertiesAPIController, ApplicationsAPIController, SyncController, settingsApiController)
	})

	return router
}

func CreateHelperRouter() chi.Router {
	HealthApiService := NewHealthApiService()
	HealthAPIController := api.NewHealthAPIController(HealthApiService)

	MetricsApiService := services.NewMetricsAPIService(metrics_pkg.Registry(metrics_exporter.MetricsRegistryName))
	MetricsAPIController := api.NewMetricsAPIController(MetricsApiService)

	router := api.NewRouter()
	router.Group(func(router chi.Router) {
		api.AddRoutes(router, HealthAPIController)
	})

	router.Group(func(router chi.Router) {
		router.Use(tracing.HttpMiddleware)
		router.Use(runaictx.ContextMiddleware)
		router.Use(runaimetrics.HttpMiddleware)
		router.Use(logginghttp.HttpMiddleware)
		router.Use(middlewares.PreserveJwtToken)
		api.AddRoutes(router, MetricsAPIController)
	})
	return router
}
