/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients"
	assetsClient "github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/assets"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/services"
	middlewares "github.com/dsx-ai-factory/run-ai/runai-application-packages/old-auth-token-middleware"
	logginghttp "github.com/dsx-ai-factory/run-ai/runai-common-packages/logging/http"
	runaimetrics "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/http"
	runaictx "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"
	tracing "github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing/http"
	"github.com/go-chi/chi/v5"
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

	rightSizingApiController := api.NewRightSizingAPIController(services.NewRightSizingAPIService())

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
		api.AddRoutes(router, rightSizingApiController)
		api.AddRoutes(router, registriesApiController)
		api.AddRoutes(router, helmReposApiController)
		api.AddRoutes(router, settingsApiController)
	})

	return router
}
