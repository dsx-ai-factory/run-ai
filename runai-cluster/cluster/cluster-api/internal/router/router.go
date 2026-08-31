/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	"net/http"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/api"

	clusterstatus "github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/cluster/status"
	connectionmetrics "github.com/run-ai/runai/runai-common-packages/metrics/connection"
	httpmetrics "github.com/run-ai/runai/runai-common-packages/metrics/http"

	"github.com/go-chi/chi/v5"
	openapi "github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/health"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/status"
)

const (
	PathStatus       = "/status"
	PathLogs         = "/logs"
	PathLogsDownload = "/logs/download"
	PathPortforward  = "/port-forward"
	PathExec         = "/exec"
	PathAttach       = "/attach"
)

type UrlHandler interface {
	Get(w http.ResponseWriter, r *http.Request)
	Post(w http.ResponseWriter, r *http.Request)
}

func CreateRouter(container Container) chi.Router {
	r := container.Router

	healthApiService := health.NewApiService()
	healthApiController := openapi.NewHealthAPIController(healthApiService)

	statusApiService := status.NewStatusApiService()
	statusApiController := openapi.NewClusterAPIAPIController(statusApiService)

	clusterStatus := clusterstatus.NewApiService()
	clusterStatusController := openapi.NewStatusAPIController(clusterStatus)

	r.Group(func(r chi.Router) {
		r.Use(connectionmetrics.MetricMiddleware)
		r.Use(container.AuthMiddleware)
		r.Use(container.ProjectMiddleware)
		r.Use(container.WorkloadAttributesMiddleware)
		r.Use(container.AuthzMiddleware)
		r.Use(container.WorkloadMiddleware)
		r.Use(container.MetricsMiddleware)

		r.Route(api.ClusterAPIBaseUrl, func(r chi.Router) {
			for path, handler := range container.MapHandler {
				r.Get(path, handler.Get)
				r.Post(path, handler.Post)
			}
		})

		r = openapi.AddRoutes(r, statusApiController) //nolint
	})

	r.Group(func(r chi.Router) {
		r.Use(httpmetrics.HttpMiddleware)
		r.Use(container.AuthMiddleware)

		r = openapi.AddRoutes(r, //nolint
			clusterStatusController,
		)
	})

	r.Group(func(r chi.Router) {
		r.Use(httpmetrics.HttpMiddleware)
		r = openapi.AddRoutes(r, //nolint
			healthApiController,
		)
	})

	if container.K8sObjectService != nil && container.K8sObjectMiddleware != nil {
		k8sObjectController := openapi.NewK8sObjectAPIController(container.K8sObjectService)
		r.Group(func(r chi.Router) {
			r.Use(httpmetrics.HttpMiddleware)
			r.Use(container.AuthMiddleware)
			r.Use(container.K8sObjectMiddleware)
			r = openapi.AddRoutes(r, k8sObjectController) //nolint
		})
	}

	return r
}
