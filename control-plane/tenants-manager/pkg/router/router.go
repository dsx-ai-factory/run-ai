package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/run-ai/runai/control-plane/tenants-manager/generated/pkg/api"
	auth_manager "github.com/run-ai/runai/control-plane/tenants-manager/pkg/business/authorization"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/clients"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/services"
	logginghttp "github.com/run-ai/runai/runai-common-packages/logging/http"
	runaimetrics "github.com/run-ai/runai/runai-common-packages/metrics/http"
	runaictx "github.com/run-ai/runai/runai-common-packages/runaictx/http"
	tracing "github.com/run-ai/runai/runai-common-packages/tracing/http"
)

func CreateRouter(
	tenantsApiService api.TenantsAPIServicer,
	tenantPublicApiService api.TenantAPIServicer,
	internalApiService api.InternalAPIServicer,
	logoDao dao.Logo,
	authManager auth_manager.AuthManagerInterface,
	bffClient clients.IBffClient,
) chi.Router {
	HealthApiService := services.NewHealthApiService()
	HealthApiController := api.NewHealthAPIController(HealthApiService)

	LogoApiService := services.NewLogoApiService(logoDao, &authManager, bffClient)
	LogoApiController := api.NewLogoAPIController(LogoApiService)

	TenantsApiController := api.NewTenantsAPIController(tenantsApiService)
	TenantPublicApiController := api.NewTenantAPIController(tenantPublicApiService)
	InternalApiController := api.NewInternalAPIController(internalApiService)

	// Add routes to the router for each of the controllers
	router := api.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(runaictx.ContextMiddleware)
		api.AddRoutes(r, HealthApiController, InternalApiController)
	})
	router.Group(func(r chi.Router) {
		r.Use(tracing.HttpMiddleware)
		r.Use(runaictx.ContextMiddleware)
		r.Use(runaimetrics.HttpMiddleware)
		r.Use(logginghttp.HttpMiddleware)
		api.AddRoutes(r, TenantsApiController, LogoApiController)
	})

	router.Group(func(r chi.Router) {
		r.Use(runaimetrics.HttpMiddleware)
		r.Use(tracing.HttpMiddleware)
		r.Use(runaictx.ContextMiddleware)
		r.Use(logginghttp.HttpMiddleware)
		api.AddRoutes(r, TenantPublicApiController)
	})

	return router
}
