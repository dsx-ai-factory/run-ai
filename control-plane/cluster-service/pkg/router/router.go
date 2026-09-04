/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services"
	logginghttp "github.com/dsx-ai-factory/run-ai/runai-common-packages/logging/http"
	runaimetrics "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/http"
	runaictx "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"
	tracing "github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing/http"
	"github.com/go-chi/chi/v5"
)

func CreateRouter(clusterServices *services.ClusterServiceServices) chi.Router {
	healthApiController := api.NewHealthAPIController(clusterServices.HealthApiService)
	nodesApiController := api.NewNodesAPIController(clusterServices.NodesApiService)
	nodepoolsApiController := api.NewNodePoolsAPIController(clusterServices.NodepoolsApiService)
	clustersApiController := api.NewClustersAPIController(clusterServices.ClustersApiService)
	internalApiController := api.NewInternalAPIController(clusterServices.InternalApiService)
	cpConfigApiController := api.NewControlPlaneConfigAPIController(clusterServices.CpConfigApiService)
	clustersConfigApiController := api.NewClusterConfigAPIController(clusterServices.ClustersApiConfigService)
	networkTopologiesApiController := api.NewNetworkTopologiesAPIController(clusterServices.NetworkTopologiesService)
	settingsApiController := api.NewSettingsAPIController(clusterServices.SettingsApiService)

	// Add routes to the router for each of the controllers
	router := api.NewRouter()

	api.AddRoutes(router, healthApiController)

	router.Group(func(r chi.Router) {
		r.Use(tracing.HttpMiddleware)
		r.Use(runaictx.ContextMiddleware)
		r.Use(runaimetrics.HttpMiddleware)
		r.Use(logginghttp.HttpMiddleware)
		r.Use(addFullHostToContext)
		api.AddRoutes(r, networkTopologiesApiController, nodesApiController,
			nodepoolsApiController, clustersApiController, internalApiController,
			cpConfigApiController, clustersConfigApiController, settingsApiController)
	})

	return router
}
