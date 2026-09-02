/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/dsx-ai-factory/run-ai/control-plane/gateway-hub/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/gateway-hub/internal/redirect"
	"github.com/dsx-ai-factory/run-ai/control-plane/gateway-hub/internal/tenant"
	"github.com/dsx-ai-factory/run-ai/control-plane/gateway-hub/pkg/config"
	runaiconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	logginghttp "github.com/dsx-ai-factory/run-ai/runai-common-packages/logging/http"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics"
	runaictxhttp "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing"
	tracinghttp "github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing/http"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

func init() {
	logging.Init(nil)
	_, _ = metrics.Init(nil)
	_, _ = tracing.Init(nil)
	config.InitConfig()
}

func main() {
	cfg := runaiconfig.Get[*config.Config]()
	logging.Infof("Starting gateway-hub")
	runaihttp.ServeHTTP(createRouter(cfg), nil)
}

func createRouter(cfg *config.Config) chi.Router {
	authHandler := auth.NewHandler(cfg.KeycloakURL, cfg.ManagementRealm)
	tenantHandler := tenant.NewHandler(cfg.TenantsManagerURL)
	redirectHandler := redirect.NewHandler(cfg.ClusterServiceURL)

	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(tracinghttp.HttpMiddleware)
		r.Use(runaictxhttp.ContextMiddleware)
		r.Use(logginghttp.HttpMiddleware)

		r.Get("/internal/plugins/authentication-handler", authHandler.ServeHTTP)
		r.Get("/internal/plugins/extract-tenant-id", tenantHandler.ServeHTTP)
		r.Get("/internal/plugins/workload-redirect-validation", redirectHandler.ValidateRedirect)
	})

	return r
}
