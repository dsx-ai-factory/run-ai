/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	logginghttp "github.com/dsx-ai-factory/run-ai/runai-common-packages/logging/http"
	runaictx "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"
	tracing "github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing/http"
	chi "github.com/go-chi/chi/v5"

	"github.com/dsx-ai-factory/run-ai/control-plane/{{cookiecutter.app_name}}/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/{{cookiecutter.app_name}}/pkg/services"
	middlewares "github.com/dsx-ai-factory/run-ai/runai-application-packages/old-auth-token-middleware"
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
