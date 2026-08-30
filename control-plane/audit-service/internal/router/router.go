package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/run-ai/runai/control-plane/audit-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/audit-service/internal/services"
	logginghttp "github.com/run-ai/runai/runai-common-packages/logging/http"
	runaimetrics "github.com/run-ai/runai/runai-common-packages/metrics/http"
	runaictx "github.com/run-ai/runai/runai-common-packages/runaictx/http"
	tracing "github.com/run-ai/runai/runai-common-packages/tracing/http"
)

func CreateRouter(auditLogsApiService api.AuditLogsAPIServicer, internalAuditLogsApiService api.InternalAuditLogsAPIServicer) chi.Router {
	healthApiService := services.NewHealthApiService()
	healthApiController := api.NewHealthAPIController(healthApiService)

	auditLogsApiController := api.NewAuditLogsAPIController(auditLogsApiService)
	internalAuditLogsAPIController := api.NewInternalAuditLogsAPIController(internalAuditLogsApiService)

	router := api.NewRouter()

	router.Group(func(router chi.Router) {
		// TODO: consider more middlewares
		router.Use(tracing.HttpMiddleware)
		router.Use(runaictx.ContextMiddleware)
		router.Use(runaimetrics.HttpMiddleware)
		router.Use(logginghttp.HttpMiddleware)
		api.AddRoutes(router,
			healthApiController,
			auditLogsApiController,
			internalAuditLogsAPIController)
	})

	return router
}
