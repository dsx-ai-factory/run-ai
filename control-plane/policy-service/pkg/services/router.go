/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	middlewares "github.com/dsx-ai-factory/run-ai/runai-application-packages/old-auth-token-middleware"
	logginghttp "github.com/dsx-ai-factory/run-ai/runai-common-packages/logging/http"
	runaimetrics "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/http"
	runaictx "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"
	tracing "github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing/http"
	"github.com/go-chi/chi/v5"
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
