/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	chi "github.com/go-chi/chi/v5"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/services"
	logginghttp "github.com/dsx-ai-factory/run-ai/runai-common-packages/logging/http"
	runaimetrics "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/http"
	runaictx "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"
	tracing "github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing/http"
)

func CreateRouter() chi.Router {
	healthApiService := services.NewHealthAPIService()
	healthApiController := api.NewHealthAPIController(healthApiService)

	notificationTypesApiService := services.NewNotificationTypesAPIService()
	notificationTypesApiController := api.NewNotificationTypesAPIController(notificationTypesApiService)

	notificationStateApiService := services.NewNotificationStateAPIService()
	notificationStateApiController := api.NewNotificationStateAPIController(notificationStateApiService)

	notifiersApiService := services.NewNotificationChannelsAPIService()
	notifiersApiController := api.NewNotificationChannelsAPIController(notifiersApiService)

	emailApiService := services.NewEmailAPIService()
	emailApiController := api.NewEmailAPIController(emailApiService)

	subscriptionsApiService := services.NewSubscriptionsAPIService()
	subscriptionsApiController := api.NewSubscriptionsAPIController(subscriptionsApiService)

	slackApiService := services.NewSlackAPIService()

	slackApiController := api.NewSlackAPIController(slackApiService)

	router := api.NewRouter()

	router.Group(func(router chi.Router) {
		api.AddRoutes(router, healthApiController)
	})

	router.Group(func(router chi.Router) {
		router.Use(tracing.HttpMiddleware)
		router.Use(runaictx.ContextMiddleware)
		router.Use(runaimetrics.HttpMiddleware)
		router.Use(logginghttp.HttpMiddleware)
		api.AddRoutes(router,
			notificationTypesApiController,
			notificationStateApiController,
			notifiersApiController,
			emailApiController,
			slackApiController,
			subscriptionsApiController)
	})

	return router
}
