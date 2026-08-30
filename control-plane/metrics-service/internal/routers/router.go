package routers

import (
	"github.com/go-chi/chi/v5"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
	logginghttp "github.com/run-ai/runai/runai-common-packages/logging/http"
	runaimetrics "github.com/run-ai/runai/runai-common-packages/metrics/http"
	"github.com/run-ai/runai/runai-common-packages/runaictx/http"
	tracing "github.com/run-ai/runai/runai-common-packages/tracing/http"
)

func CreateRouter(healthController api.Router, metricsController api.Router, reportsController api.Router, metricStoresController api.Router) chi.Router {
	// Add routes to the router for each of the controllers
	router := api.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(tracing.HttpMiddleware)
		r.Use(http.ContextMiddleware)
		r.Use(runaimetrics.HttpMiddleware)
		r.Use(logginghttp.HttpMiddleware)
		api.AddRoutes(r, metricsController, healthController, reportsController, metricStoresController)
	})

	return router
}
