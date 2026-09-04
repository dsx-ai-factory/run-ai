/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/pkg/services"
	middlewares "github.com/dsx-ai-factory/run-ai/runai-application-packages/old-auth-token-middleware"
	logginghttp "github.com/dsx-ai-factory/run-ai/runai-common-packages/logging/http"
	runaimetrics "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/http"
	runaictx "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"
	tracing "github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing/http"
	chi "github.com/go-chi/chi/v5"
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
