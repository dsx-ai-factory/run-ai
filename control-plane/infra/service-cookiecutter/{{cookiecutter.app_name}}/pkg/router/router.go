package router

import (
	chi "github.com/go-chi/chi/v5"
	logginghttp "github.com/run-ai/runai/runai-common-packages/logging/http"
	runaictx "github.com/run-ai/runai/runai-common-packages/runaictx/http"
	tracing "github.com/run-ai/runai/runai-common-packages/tracing/http"

	"github.com/run-ai/runai/control-plane/{{cookiecutter.app_name}}/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/{{cookiecutter.app_name}}/pkg/services"
	middlewares "github.com/run-ai/runai/runai-application-packages/old-auth-token-middleware"
)

func CreateRouter() chi.Router {
	healthApiService := services.NewHealthApiService()
	healthApiController := api.NewHealthApiController(healthApiService)

	// Init more service's controllers here

	router := api.NewRouter()
	router.Group(func(router chi.Router) {
		router.Use(tracing.HttpMiddleware)
		router.Use(runaictx.ContextMiddleware)
		router.Use(runaimetrics.HttpMiddleware)
		router.Use(logginghttp.HttpMiddleware)
		// Add routes to the router for each of the controllers
		api.AddRoutes(router,
			healthApiController,
			// Add service's controllers route here as well
		)
	})

	return router
}
