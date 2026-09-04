/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/datavolumes"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	middlewares "github.com/dsx-ai-factory/run-ai/runai-application-packages/old-auth-token-middleware"
	dvclient "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	logginghttp "github.com/dsx-ai-factory/run-ai/runai-common-packages/logging/http"
	runaimetrics "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/http"
	runaictx "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"
	tracing "github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing/http"
	"github.com/go-chi/chi/v5"
)

func CreateRouter() chi.Router {
	healthApiService := NewHealthApiService()
	healthApiController := api.NewHealthAPIController(healthApiService)

	AssetsApiService := NewAssetsAPIService()
	AssetsApiController := api.NewAssetsAPIController(AssetsApiService)

	CredentialsApiService := NewCredentialsApiService()
	CredentialsApiController := api.NewCredentialsAPIController(CredentialsApiService)

	S3ApiService := NewS3ApiService()
	S3ApiController := api.NewS3APIController(S3ApiService)

	GitApiService := NewGitApiService()
	GitApiController := api.NewGitAPIController(GitApiService)

	NFSApiService := NewNFSApiService()
	nfsApiController := api.NewNFSAPIController(NFSApiService)

	PVCApiService := NewPVCApiService()
	pvcApiController := api.NewPVCAPIController(PVCApiService)

	HostPathApiService := NewHostPathApiService()
	hostPathApiController := api.NewHostPathAPIController(HostPathApiService)

	configMapService := NewConfigMapService(&dao.ConfigMapsDBAccessor)
	configMapController := api.NewConfigMapAPIController(configMapService)

	secretVolumeService := NewSecretVolumeService(&dao.SecretVolumeDBAccessor)
	secretVolumeController := api.NewSecretAPIController(secretVolumeService)

	WorkspacesApiService := NewWorkspacesApiService()
	workspacesApiController := api.NewWorkspacesAPIController(WorkspacesApiService)

	TrainingsApiService := NewTrainingsApiService()
	trainingsApiController := api.NewTrainingsAPIController(TrainingsApiService)

	WorkloadApiService := NewWorkloadApiService()
	workloadApiController := api.NewWorkloadAPIController(WorkloadApiService)

	ComputeApiService := NewComputeApiService()
	ComputeApiController := api.NewComputeAPIController(ComputeApiService)

	environmentApiService := NewEnvironmentApiService()
	environmentApiController := api.NewEnvironmentAPIController(environmentApiService)

	TemplateApiService := NewTemplateApiService()
	TemplateApiController := api.NewTemplateAPIController(TemplateApiService)

	dataVolumesAdapter := datavolumes.NewAdapter(dvclient.NewDatavolumesClient(runai_config.Get[*config.ServicesConfig]().DataVolumesServiceUrl))
	datasourcesApiService := NewDatasourcesApiService(&dao.ConfigMapsDBAccessor, &dao.SecretVolumeDBAccessor, dataVolumesAdapter)
	datasourceController := api.NewDatasourcesAPIController(datasourcesApiService)

	syncApiService := NewSyncApiService()
	syncController := api.NewSynchronizationAPIController(syncApiService)

	registryApiService := NewRegistryApiService()
	registryController := api.NewRegistryAPIController(registryApiService)

	modelApiService := NewModelApiService()
	modelController := api.NewModelAPIController(modelApiService)

	inferencesApiService := NewInferencesApiService()
	inferencesApiController := api.NewInferencesAPIController(inferencesApiService)

	modelInferenceApiService := NewModelInferenceApiService(&modelApiService, &inferencesApiService)
	modelInferenceController := api.NewModelInferenceAPIController(modelInferenceApiService)

	eventsApiService := NewEventsAPIService()
	eventsApiController := api.NewEventsAPIController(eventsApiService)

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

		api.AddRoutes(router,
			AssetsApiController,
			CredentialsApiController,
			S3ApiController,
			GitApiController,
			nfsApiController,
			pvcApiController,
			hostPathApiController,
			ComputeApiController,
			environmentApiController,
			datasourceController,
			TemplateApiController,
			workspacesApiController,
			trainingsApiController,
			inferencesApiController,
			workloadApiController,
			syncController,
			registryController,
			modelController,
			modelInferenceController,
			configMapController,
			secretVolumeController,
			eventsApiController)
	})

	return router
}
