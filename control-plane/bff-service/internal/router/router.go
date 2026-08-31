/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/run-ai/runai/control-plane/bff-service/internal/config"
	"github.com/run-ai/runai/control-plane/bff-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/bff-service/internal/services"
	pkgServices "github.com/run-ai/runai/control-plane/bff-service/pkg/services"
	runaictx "github.com/run-ai/runai/runai-common-packages/runaictx/http"
)

func CreateRouter(
	serviceConfig *config.Service,
	bootConfig *config.BootConfig,
) chi.Router {
	HealthApiService := services.NewHealthAPIService()
	HealthApiController := api.NewHealthAPIController(HealthApiService)

	VersionApiService := services.NewVersionAPIService(serviceConfig.RunaiVersion)
	VersionApiController := api.NewVersionAPIController(VersionApiService)

	bootConfigAPIService := services.NewBootConfigAPIService(serviceConfig.RunaiVersion, bootConfig)
	bootConfigAPIController := api.NewBootConfigAPIController(bootConfigAPIService)

	SettingsApiService := pkgServices.NewSettingsAPIService()
	SettingsApiController := api.NewSettingsAPIController(SettingsApiService)

	systemNotificationApiService := pkgServices.NewSystemNotificationAPIService()
	systemNotificationsApiController := api.NewSystemNotificationAPIController(systemNotificationApiService)

	router := api.NewRouter()

	api.AddRoutes(router, HealthApiController, VersionApiController, bootConfigAPIController)

	router.Group(func(r chi.Router) {
		r.Use(runaictx.ContextMiddleware)
		api.AddRoutes(r, SettingsApiController, systemNotificationsApiController)
	})

	return router
}
