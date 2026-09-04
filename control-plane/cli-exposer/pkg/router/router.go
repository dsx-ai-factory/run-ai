/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	"context"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/cli-exposer/pkg/document"

	"github.com/go-chi/chi/v5"
	"github.com/dsx-ai-factory/run-ai/control-plane/cli-exposer/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cli-exposer/pkg/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/cli-exposer/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/cli-exposer/pkg/version"
	middlewares "github.com/dsx-ai-factory/run-ai/runai-application-packages/old-auth-token-middleware"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	logginghttp "github.com/dsx-ai-factory/run-ai/runai-common-packages/logging/http"
	runaimetrics "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/http"
	runaictx "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"
	tracing "github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing/http"
)

func CreateRouter() chi.Router {
	healthApiService := services.NewHealthApiService()
	healthApiController := api.NewHealthAPIController(healthApiService)

	validators, err := services.GetRegexValidators()
	if err != nil {
		logging.Fatalf("Error creating validators: %v", err)
	}

	cliVersion, err := version.GetVersion()
	if err != nil {
		logging.Fatalf("error getting CLI binary version: %v", err)
	}

	htmlDocContent, err := document.GetHtmlContent()
	if err != nil {
		logging.Fatalf("error getting html content: %v", err)
	}

	cliApiService := services.NewCliApiService(validators, cliVersion, htmlDocContent)
	cliApiController := api.NewResearcherCommandLineInterfaceAPIController(cliApiService)

	router := api.NewRouter()

	router.Group(func(router chi.Router) {
		router.Use(tracing.HttpMiddleware)
		router.Use(runaictx.ContextMiddleware)
		router.Use(runaimetrics.HttpMiddleware)
		router.Use(logginghttp.HttpMiddleware)
		router.Use(middlewares.PreserveJwtToken)
		router.Use(PreserveHostUrl)
		router.Use(PreserveUserAgent)

		router.Get("/api/v1/cli/*", func(w http.ResponseWriter, r *http.Request) {
			// Redirect to a new path
			http.Redirect(w, r, strings.Replace(r.URL.Path, "/cli/", "/cli-exposer/", 1), http.StatusMovedPermanently)
		})

		// Add routes to the router for each of the controllers
		router = api.AddRoutes(router, //nolint:staticcheck // used by reference
			healthApiController,
			cliApiController,
			// Add service's controllers route here as well
		)
	})

	return router
}

// PreserveHostUrl is a middleware that preserves the host URL in the request context
func PreserveHostUrl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), consts.ContextHostUrl, r.Host)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// PreserveUserAgent is a middleware that preserves the user agent in the request context
func PreserveUserAgent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), consts.UserAgentHeader, r.Header.Get(string(consts.UserAgentHeader)))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
