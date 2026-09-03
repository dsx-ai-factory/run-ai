/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/internal/services"
	logginghttp "github.com/dsx-ai-factory/run-ai/runai-common-packages/logging/http"
	runaictx "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"
	tracing "github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing/http"
	"github.com/go-chi/chi/v5"
)

func CreateRouter(tenantService api.TenantAPIServicer, internalService api.InternalAPIServicer) chi.Router {
	healthApiService := services.NewHealthAPIService()
	healthApiController := api.NewHealthAPIController(healthApiService)

	tenantApiController := api.NewTenantAPIController(tenantService)
	internalApiController := api.NewInternalAPIController(internalService)

	router := api.NewRouter()

	api.AddRoutes(router, healthApiController)

	router.Group(func(r chi.Router) {
		r.Use(tracing.HttpMiddleware)
		r.Use(runaictx.ContextMiddleware)
		r.Use(logginghttp.HttpMiddleware)
		api.AddRoutes(r, tenantApiController, internalApiController)
	})

	return router
}
