package services

import (
	"github.com/go-chi/chi/v5"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	middlewares "github.com/run-ai/runai/runai-application-packages/old-auth-token-middleware"
	logginghttp "github.com/run-ai/runai/runai-common-packages/logging/http"
	runaimetrics "github.com/run-ai/runai/runai-common-packages/metrics/http"
	runaictx "github.com/run-ai/runai/runai-common-packages/runaictx/http"
	tracing "github.com/run-ai/runai/runai-common-packages/tracing/http"
)

func CreateRouter() chi.Router {
	healthApiService := NewHealthApiService()
	healthApiController := api.NewHealthAPIController(healthApiService)

	PolicyApiService := NewPolicyApiService()
	PolicyApiController := api.NewPolicyAPIController(PolicyApiService)

	SyncApiService := NewSyncApiService()
	SyncApiController := api.NewSynchronizationAPIController(SyncApiService)

	MappingApiService := NewMappingApiService()
	MappingApiController := api.NewSpecMappingAPIController(MappingApiService)

	router := api.NewRouter()

	// The `healthApiController` must be in its own router group without middlewares like `auth.ExtractAuthInfo`,
	// as these would block unauthenticated access required for health check.
	router.Group(func(router chi.Router) {
		api.AddRoutes(router, healthApiController)
	})

	router.Group(func(router chi.Router) {
		router.Use(tracing.HttpMiddleware)
		router.Use(runaictx.ContextMiddleware)
		router.Use(runaimetrics.HttpMiddleware)
		router.Use(logginghttp.HttpMiddleware)
		router.Use(middlewares.PreserveJwtToken) // needed for project related APIs
		router.Use(auth.ExtractAuthInfo)

		api.AddRoutes(router, PolicyApiController)
		api.AddRoutes(router, SyncApiController)
		api.AddRoutes(router, MappingApiController)
	})

	return router
}
