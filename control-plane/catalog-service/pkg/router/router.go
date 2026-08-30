package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/run-ai/runai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/clients"
	assetsClient "github.com/run-ai/runai/control-plane/catalog-service/pkg/clients/assets"
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/services"
	middlewares "github.com/run-ai/runai/runai-application-packages/old-auth-token-middleware"
	logginghttp "github.com/run-ai/runai/runai-common-packages/logging/http"
	runaimetrics "github.com/run-ai/runai/runai-common-packages/metrics/http"
	runaictx "github.com/run-ai/runai/runai-common-packages/runaictx/http"
	tracing "github.com/run-ai/runai/runai-common-packages/tracing/http"
)

func CreateRouter() chi.Router {
	healthApiService := services.NewHealthApiService()
	healthApiController := api.NewHealthAPIController(healthApiService)

	ngcClient := clients.NewNGCClient()
	assetsClientInstance := assetsClient.Get()
	imagesApiService := services.NewImagesAPIService(ngcClient, assetsClientInstance)
	imagesApiController := api.NewImagesAPIController(imagesApiService)

	registriesApiService := services.NewRegistriesAPIService()
	registriesApiController := api.NewRegistriesAPIController(registriesApiService)

	inferencesApiService := services.NewInferencesAPIService()
	inferencesApiController := api.NewInferencesAPIController(inferencesApiService)

	helmReposApiService := services.NewHelmReposAPIService(assetsClientInstance)
	helmReposApiController := api.NewHelmReposAPIController(helmReposApiService)

	settingsApiService := services.NewSettingsService()
	settingsApiController := api.NewSettingsAPIController(settingsApiService)

	router := api.NewRouter()

	router.Group(func(router chi.Router) {
		router.Use(tracing.HttpMiddleware)
		router.Use(runaictx.ContextMiddleware)
		router.Use(runaimetrics.HttpMiddleware)
		router.Use(logginghttp.HttpMiddleware)
		router.Use(middlewares.PreserveJwtToken)

		api.AddRoutes(router, healthApiController)
		api.AddRoutes(router, imagesApiController)
		api.AddRoutes(router, inferencesApiController)
		api.AddRoutes(router, registriesApiController)
		api.AddRoutes(router, helmReposApiController)
		api.AddRoutes(router, settingsApiController)
	})

	return router
}
