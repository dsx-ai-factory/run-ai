package router

import (
	chi "github.com/go-chi/chi/v5"
	"github.com/run-ai/runai/control-plane/datavolumes-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/datavolumes-service/pkg/services"
	middlewares "github.com/run-ai/runai/runai-application-packages/old-auth-token-middleware"
	logginghttp "github.com/run-ai/runai/runai-common-packages/logging/http"
	runaimetrics "github.com/run-ai/runai/runai-common-packages/metrics/http"
	runaictx "github.com/run-ai/runai/runai-common-packages/runaictx/http"
	tracing "github.com/run-ai/runai/runai-common-packages/tracing/http"
)

// CreateRouter sets up the HTTP router with all datavolumes-service routes and middleware.
func CreateRouter() chi.Router {
	healthApiService := services.NewHealthApiService()
	healthApiController := api.NewHealthAPIController(healthApiService)

	datavolumesService := services.NewDatavolumesApiService()
	datavolumesApiController := api.NewDatavolumesAPIController(datavolumesService)

	settingsApiService := services.NewSettingsService()
	settingsApiController := api.NewSettingsAPIController(settingsApiService)

	router := api.NewRouter()

	router.Group(func(router chi.Router) {
		api.AddRoutes(router, healthApiController)
	})

	router.Group(func(router chi.Router) {
		router.Use(tracing.HttpMiddleware)
		router.Use(runaictx.ContextMiddleware)
		router.Use(runaimetrics.HttpMiddleware)
		router.Use(logginghttp.HttpMiddleware)
		router.Use(middlewares.PreserveJwtToken)

		api.AddRoutes(router,
			datavolumesApiController,
			settingsApiController,
		)
	})

	return router
}
