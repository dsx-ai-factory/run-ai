/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"github.com/go-chi/chi/v5"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/dao"
	middlewares "github.com/run-ai/runai/runai-application-packages/old-auth-token-middleware"
	logginghttp "github.com/run-ai/runai/runai-common-packages/logging/http"
	runaimetrics "github.com/run-ai/runai/runai-common-packages/metrics/http"
	runaictx "github.com/run-ai/runai/runai-common-packages/runaictx/http"
	tracing "github.com/run-ai/runai/runai-common-packages/tracing/http"
)

func CreateRouter() chi.Router {
	healthApiService := NewHealthApiService()
	healthApiController := api.NewHealthAPIController(healthApiService)

	pvcsApiService := NewPVCService()
	pvcsApiController := api.NewPVCsAPIController(pvcsApiService)

	secretsApiService := NewSecretsService()
	secretsApiController := api.NewSecretsAPIController(secretsApiService)

	configMapsApiService := NewConfigMapsService(dao.NewConfigMapsDBAccessor())
	configMapsApiController := api.NewConfigMapsAPIController(configMapsApiService)

	clusterObjectsStatusApiService := NewClusterObjectsStatusApiService(dao.NewConfigMapsDBAccessor())
	clusterObjectsStatusApiController := api.NewClusterObjectsStatusAPIController(clusterObjectsStatusApiService)

	storageClassesApiService := NewStorageClassesApiService()
	storageClassesApiController := api.NewStorageClassesAPIController(storageClassesApiService)

	storageClassConfigurationApiService := NewStorageClassConfigurationAPIService()
	storageClassConfigurationApiController := api.NewStorageClassConfigurationAPIController(storageClassConfigurationApiService)

	router := api.NewRouter()

	router.Group(func(router chi.Router) {
		router.Use(tracing.HttpMiddleware)
		router.Use(runaictx.ContextMiddleware)
		router.Use(runaimetrics.HttpMiddleware)
		router.Use(logginghttp.HttpMiddleware)
		router.Use(middlewares.PreserveJwtToken) // needed for project related APIs
		// TODO - add csp middleware

		// Add routes to the router for each of the controllers
		api.AddRoutes(router,
			healthApiController,
			pvcsApiController,
			secretsApiController,
			storageClassConfigurationApiController,
			configMapsApiController,
			clusterObjectsStatusApiController,
			storageClassesApiController,
		)
	})

	return router
}
