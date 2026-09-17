/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"fmt"
	"net/http"

	authorizationApi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/assets"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

func loadTemplate(ctx context.Context, params *createWorkloadParams, masterSameAsWorker bool,
	permittedScopes *authorizationApi.PermittedScopes,
	ancestors *orgUnitApi.Ancestors,
) (map[string]interface{}, *api.Error) {
	if params.TemplateId == nil {
		return nil, nil
	}

	// Debug logging for permitted scopes
	if permittedScopes != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Permitted scopes - System: %v, Tenants: %v, Clusters: %v, Departments: %v, Projects: %v",
			permittedScopes.System, permittedScopes.Tenants, permittedScopes.Clusters, permittedScopes.Departments, permittedScopes.Projects)
	} else {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Permitted scopes is nil")
	}

	template, err := dao_crud.GetRawTemplateById(ctx, uint(params.TenantId), *params.TemplateId, permittedScopes)
	if err != nil {
		return nil, err
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Retrieved template - ID: %s, ScopeType: %s, ScopeId: %s, WorkloadType: %s",
		template.ID, template.ScopeType, template.ScopeId, template.WorkloadType)

	if template.WorkloadType != params.WorkloadType {
		return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("%s template '%s' does not exist", params.WorkloadType, *params.TemplateId), nil)
	}

	scopeMatch := template.ScopeType == string(api.SCOPETYPE_SYSTEM)
	scopeMatch = scopeMatch || template.ScopeType == string(api.SCOPETYPE_TENANT)
	scopeMatch = scopeMatch || template.ScopeType == string(api.SCOPETYPE_CLUSTER) && template.ScopeId == *ancestors.ClusterUuid.Get()
	scopeMatch = scopeMatch || template.ScopeType == string(api.SCOPETYPE_DEPARTMENT) && template.ScopeId == *ancestors.DepartmentId.Get()
	scopeMatch = scopeMatch || template.ScopeType == string(api.SCOPETYPE_PROJECT) && template.ScopeId == *ancestors.ProjectId.Get()

	if !scopeMatch {
		return nil, api.NewValidationError(fmt.Sprintf("template cannot be used for project %s", params.Response.ProjectId))
	}

	if params.WorkloadType != string(api.WORKLOADTYPE_DISTRIBUTED) {
		return template.CombinedSpec, nil
	}

	result := map[string]interface{}{
		workerSection: map[string]interface{}(template.CombinedSpec),
	}

	if masterSameAsWorker || convutils.BoolIsTrue(template.MasterSameAsWorker) {
		result[masterSection] = map[string]interface{}(template.CombinedSpec.Clone())
	} else {
		result[masterSection] = map[string]interface{}(template.CombinedMasterSpec)
	}

	return result, nil
}

func loadAssets(ctx context.Context, params *createWorkloadParams,
	masterSameAsWorker bool,
) (map[string]interface{}, *api.Error) {
	if params.AssetIds == nil && params.MasterAssetIds == nil {
		return nil, nil
	}

	if params.MasterAssetIds != nil {
		if params.WorkloadType != string(api.WORKLOADTYPE_DISTRIBUTED) {
			return nil, api.NewValidationError("masterAssetIds is not supported by this workload type")
		}

		if masterSameAsWorker {
			return nil, api.NewValidationError("masterAssetIds cannot be provided with masterSameAsWorker=true")
		}
	}

	flatSpec := make(map[string]interface{})

	assetFlatReq := assetsApi.AssetFlatRequest{
		ClusterId:          params.Response.ClusterId,
		ProjectId:          params.Response.ProjectId,
		WorkloadType:       assetsApi.WorkloadType(params.WorkloadType),
		MasterSameAsWorker: &masterSameAsWorker,
	}

	// for distributed inference we only allow assets that are supported by inference
	requiredAssetType := params.WorkloadType
	if requiredAssetType == string(api.WORKLOADTYPE_DISTRIBUTED_INFERENCE) {
		requiredAssetType = string(api.WORKLOADTYPE_INFERENCE)
	}

	//
	//   load the assets of the worker (and if masterSameAsWorker then also request their master representation)
	//
	assetFlatReq.AssetIds = getAssetIds(params.AssetIds)

	if params.AssetIds != nil && len(assetFlatReq.AssetIds) == 0 {
		return nil, api.NewValidationError("at least one asset ID is required when submitting with assets")
	}

	if len(assetFlatReq.AssetIds) != 0 {
		assetFlat, assetErr := assets.Get().RequestAssetFlatFormat(ctx, assetFlatReq)
		if assetErr != nil {
			return nil, api.NewError(int(assetErr.Code), assetErr.Message, fmt.Errorf("%v", assetErr.Details))
		}

		checkErr := checkAssetsValidity(ctx, assetFlat.AssetsMeta, requiredAssetType, params.AssetIds)
		if checkErr != nil {
			return nil, checkErr
		}

		var jsonErr error

		if params.WorkloadType == string(api.WORKLOADTYPE_DISTRIBUTED) {
			if masterSameAsWorker {
				flatSpec, jsonErr = gojsonq.LoadJson(assetFlat) // we need both master and worker sections
			} else {
				flatSpec[workerSection], jsonErr = gojsonq.LoadJson(assetFlat.Worker)
			}
		} else if params.WorkloadType == string(api.WORKLOADTYPE_DISTRIBUTED_INFERENCE) {
			flatSpec[workerSection], jsonErr = gojsonq.LoadJson(assetFlat.Worker)
		} else {
			flatSpec, jsonErr = gojsonq.LoadJson(assetFlat.Worker) // non-distributed we need only the worker sections
		}

		if jsonErr != nil {
			return nil, api.NewError(http.StatusInternalServerError, "failed to load assets payload", jsonErr)
		}
	}

	//
	//   handle the case that master has different set of assets
	//
	assetFlatReq.AssetIds = getAssetIds(params.MasterAssetIds)

	if params.MasterAssetIds != nil && len(assetFlatReq.AssetIds) == 0 {
		return nil, api.NewValidationError("at least one asset ID is required when submitting with masterAssets")
	}

	if len(assetFlatReq.AssetIds) != 0 {
		assetFlat, assetErr := assets.Get().RequestAssetFlatFormat(ctx, assetFlatReq)
		if assetErr != nil {
			return nil, api.NewError(int(assetErr.Code), assetErr.Message, fmt.Errorf("%v", assetErr.Details))
		}

		checkErr := checkAssetsValidity(ctx, assetFlat.AssetsMeta, requiredAssetType, params.MasterAssetIds)
		if checkErr != nil {
			return nil, checkErr
		}

		masterFlatSpec, jsonErr := gojsonq.LoadJson(assetFlat.Master)
		if jsonErr != nil {
			return nil, api.NewError(http.StatusInternalServerError, "failed to load assets payload", jsonErr)
		}

		flatSpec[masterSection] = masterFlatSpec
	}

	//
	//   handle the case that leader has different set of assets
	//
	assetFlatReq.AssetIds = getAssetIds(params.LeaderAssetIds)

	if params.LeaderAssetIds != nil && len(assetFlatReq.AssetIds) == 0 {
		return nil, api.NewValidationError("at least one asset ID is required when submitting with assets.leader")
	}

	if len(assetFlatReq.AssetIds) != 0 {
		if params.WorkloadType != string(api.WORKLOADTYPE_DISTRIBUTED_INFERENCE) {
			return nil, api.NewValidationError("leader assets are only supported for distributed inference workloads")
		}

		assetFlat, assetErr := assets.Get().RequestAssetFlatFormat(ctx, assetFlatReq)
		if assetErr != nil {
			return nil, api.NewError(int(assetErr.Code), assetErr.Message, fmt.Errorf("%v", assetErr.Details))
		}

		checkErr := checkAssetsValidity(ctx, assetFlat.AssetsMeta, requiredAssetType, params.LeaderAssetIds)
		if checkErr != nil {
			return nil, checkErr
		}

		// note! this is no typo. the internal API works like this: it returns a struct with master and worker.
		// the only thing that goes into the master is master section of distributed. all the rest, including the
		// leader goes into "worker" which acts as the "main" container for the result.
		leaderFlatSpec, jsonErr := gojsonq.LoadJson(assetFlat.Worker)
		if jsonErr != nil {
			return nil, api.NewError(http.StatusInternalServerError, "failed to load assets payload", jsonErr)
		}

		flatSpec[leaderSection] = leaderFlatSpec
	}

	cleanEmptyArraysMap(flatSpec)

	return flatSpec, nil
}

func getAssetIds(assetIds *api.AssetSubmissionIds) []string {
	result := make([]string, 0)

	if assetIds == nil {
		return result
	}

	if assetIds.Environment != nil {
		result = append(result, *assetIds.Environment)
	}
	if assetIds.Compute != nil {
		result = append(result, *assetIds.Compute)
	}
	if assetIds.Datasources != nil {
		result = append(result, assetIds.Datasources...)
	}

	return result
}

func checkAssetsValidity(ctx context.Context, assetsMeta []assetsApi.AssetMeta, workloadType string, requestedIds *api.AssetSubmissionIds) *api.Error {
	idToMeta := make(map[string]*assetsApi.AssetMeta)
	for _, assetMeta := range assetsMeta {
		idToMeta[assetMeta.Id] = &assetMeta
	}

	if requestedIds.Environment != nil {
		envMeta := idToMeta[*requestedIds.Environment]
		if envMeta.Kind != assetsApi.ASSETKIND_ENVIRONMENT {
			return api.NewValidationError(fmt.Sprintf("asset %s is not an environment asset", envMeta.Id))
		}

		workloadTypeErr := validateAssetVsWorkloadType(ctx, envMeta, workloadType)
		if workloadTypeErr != nil {
			return workloadTypeErr
		}
	}

	if requestedIds.Compute != nil {
		computeMeta := idToMeta[*requestedIds.Compute]
		if computeMeta.Kind != assetsApi.ASSETKIND_COMPUTE {
			return api.NewValidationError(fmt.Sprintf("asset %s is not a compute asset", computeMeta.Id))
		}
	}

	if requestedIds.Datasources != nil {
		for _, datasourceId := range requestedIds.Datasources {
			datasourceMeta := idToMeta[datasourceId]
			_, isDatasource := composer_v2.MapAssetKindToDatasourcesList[string(datasourceMeta.Kind)]
			if !isDatasource {
				return api.NewValidationError(fmt.Sprintf("asset %s is not a datasource asset", datasourceMeta.Id))
			}
		}
	}

	return nil
}
