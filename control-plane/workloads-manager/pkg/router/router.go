package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/run-ai/runai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/middleware"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/services"
	middlewares "github.com/run-ai/runai/runai-application-packages/old-auth-token-middleware"
	logginghttp "github.com/run-ai/runai/runai-common-packages/logging/http"
	runaimetrics "github.com/run-ai/runai/runai-common-packages/metrics/http"
	runaictx "github.com/run-ai/runai/runai-common-packages/runaictx/http"
	tracing "github.com/run-ai/runai/runai-common-packages/tracing/http"
)

func CreateRouter() chi.Router {
	healthApiService := services.NewHealthApiService()
	healthApiController := api.NewHealthAPIController(healthApiService)

	router := api.NewRouter()

	settingsService := services.NewSettingsService()
	settingsController := api.NewSettingsAPIController(settingsService)

	workloadsV2Service := services.NewWorkloadsV2Service()
	workloadsV2Controller := api.NewWorkloadsV2APIController(workloadsV2Service)

	nimServicesService := services.NewNimServicesService()
	nimServicesController := api.NewNVIDIANIMAPIController(nimServicesService)

	router.Group(func(router chi.Router) {
		router.Use(tracing.HttpMiddleware)
		router.Use(runaictx.ContextMiddleware)
		router.Use(middleware.RequestContextMiddleware)
		router.Use(runaimetrics.HttpMiddleware)
		router.Use(logginghttp.HttpMiddleware)
		router.Use(middlewares.PreserveJwtToken)

		api.AddRoutes(router,
			healthApiController,
			settingsController,
			workloadsV2Controller,
			nimServicesController)
	})

	return router
}
