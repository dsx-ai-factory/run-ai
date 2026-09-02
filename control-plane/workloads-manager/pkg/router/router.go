/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/middleware"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/services"
	middlewares "github.com/dsx-ai-factory/run-ai/runai-application-packages/old-auth-token-middleware"
	logginghttp "github.com/dsx-ai-factory/run-ai/runai-common-packages/logging/http"
	runaimetrics "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/http"
	runaictx "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"
	tracing "github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing/http"
)

func CreateRouter() chi.Router {
	healthApiService := services.NewHealthApiService()
	healthApiController := api.NewHealthAPIController(healthApiService)

	router := api.NewRouter()

	settingsService := services.NewSettingsService()
	settingsController := api.NewSettingsAPIController(settingsService)

	workloadsV2Service := services.NewWorkloadsV2Service()
	workloadsV2Controller := api.NewWorkloadsV2APIController(workloadsV2Service)

	nimServicesService := services.NewNimServicesService()
	nimServicesController := api.NewNVIDIANIMAPIController(nimServicesService)

	router.Group(func(router chi.Router) {
		router.Use(tracing.HttpMiddleware)
		router.Use(runaictx.ContextMiddleware)
		router.Use(middleware.RequestContextMiddleware)
		router.Use(runaimetrics.HttpMiddleware)
		router.Use(logginghttp.HttpMiddleware)
		router.Use(middlewares.PreserveJwtToken)

		api.AddRoutes(router,
			healthApiController,
			settingsController,
			workloadsV2Controller,
			nimServicesController)
	})

	return router
}
