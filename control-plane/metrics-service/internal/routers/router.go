/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package routers

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/generated/pkg/api"
	logginghttp "github.com/dsx-ai-factory/run-ai/runai-common-packages/logging/http"
	runaimetrics "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/http"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"
	tracing "github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing/http"
	"github.com/go-chi/chi/v5"
)

func CreateRouter(healthController api.Router, metricsController api.Router, reportsController api.Router, metricStoresController api.Router) chi.Router {
	// Add routes to the router for each of the controllers
	router := api.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(tracing.HttpMiddleware)
		r.Use(http.ContextMiddleware)
		r.Use(runaimetrics.HttpMiddleware)
		r.Use(logginghttp.HttpMiddleware)
		api.AddRoutes(r, metricsController, healthController, reportsController, metricStoresController)
	})

	return router
}
