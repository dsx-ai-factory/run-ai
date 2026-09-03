/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/policy_compliance"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/flat_workloads"
	policies "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"

	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	workloadsService "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/jinzhu/copier"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
)

// WorkspacesApiService is a service that implements the logic for the Workspace service
type WorkspacesApiService struct{}

// NewWorkspacesApiService creates a default api service
func NewWorkspacesApiService() *WorkspacesApiService {
	return &WorkspacesApiService{}
}

func (s *WorkspacesApiService) CreateWorkspace(ctx context.Context, request api.WorkloadCreationRequest) (api.ImplResponse, error) {
	genericWorkload, err := CreateFlatWorkloadBackwardComp(ctx, auth.WorkloadTypeWorkspace, request)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	workspace := api.Workspace{
		Meta:     genericWorkload.Meta,
		Spec:     genericWorkload.Spec,
		Workload: genericWorkload.Workload,
		Job:      genericWorkload.Job,
	}

	return api.Response(http.StatusOK, workspace), nil
}

func (s *WorkspacesApiService) CountAssets(ctx context.Context, kind string, name string, scope string,
	projectId int32, departmentId string, clusterId string,
) (api.ImplResponse, error) {
	queryParams, queryError := query_types.NewQueryParams("", name, scope, false, projectId, departmentId, clusterId, nil)
	if queryError != nil {
		return NewErrorResponse(queryError), nil
	}

	_, authError := query_types.ApplyRbacRules(ctx, queryParams)
	if authError != nil {
		return NewErrorResponse(authError), nil
	}

	assetKind, kindErr := query_types.LoadKindParameter(kind)
	if kindErr != nil {
		return NewErrorResponse(kindErr), nil
	}

	db, err := dao.CountAssetMeta(ctx, assetKind, queryParams)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	result := api.AssetCountResponse{
		Result: 0,
	}

	if countErr := db.Count(&result.Result); countErr != nil {
		dbErr := api.NewError(http.StatusServiceUnavailable, "failed to count assets", nil)
		return NewErrorResponse(dbErr.WithDetails(countErr)), nil
	}

	return api.Response(http.StatusOK, result), nil
}

func getWorkspaceByIdV2(ctx context.Context, id string) (*api.WorkspaceV2, *api.Error) {
	genericWorkload, getError := GetWorkloadById(ctx, auth.WorkloadTypeWorkspace, id)
	if getError != nil {
		return nil, getError
	}

	workspace := api.WorkspaceV2{
		Meta:   *genericWorkload.ToWorkloadMeta(),
		Assets: genericWorkload.Spec.Assets,
	}

	var err *api.Error
	workspace.SpecificRunParams, err = genericToWorkspaceRunParams(genericWorkload.Spec.SpecificEnv)
	if err != nil {
		return nil, err
	}

	return &workspace, nil
}

func (s *WorkspacesApiService) GetWorkspaceByIdV2(ctx context.Context, id string) (api.ImplResponse, error) {
	workspace, err := getWorkspaceByIdV2(ctx, id)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusOK, workspace), nil
}

func (s *WorkspacesApiService) GetWorkspaceMerged(ctx context.Context, projectId string, workloadId string, templateId string) (api.ImplResponse, error) {
	err := validateMergeParams(projectId, workloadId, templateId, false)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	//
	//   load the workspace, if workloadId is provided by the caller. otherwise, initialize it to empty workspace
	//
	workspace := &api.WorkspaceV2{
		SpecificRunParams: &api.WorkspaceSpecificRunParams{},
	}

	if workloadId != "" {
		var getError *api.Error

		workspace, getError = getWorkspaceByIdV2(ctx, workloadId)
		if getError != nil {
			return api.Response(int(getError.Code), getError), nil
		}

		projectId = workspace.Meta.ProjectId
	}

	if templateId != "" {
		template, _, getError := dao.GetTemplateAssetById(ctx, templateId)
		if getError != nil {
			return api.Response(int(getError.Code), getError), nil
		}

		if projectId != "" && template.Meta.ProjectId != nil {
			projectIdFromTemplate := fmt.Sprintf("%d", *template.Meta.ProjectId)
			if projectIdFromTemplate != projectId {
				return api.Response(http.StatusBadRequest, api.NewValidationError(fmt.Sprintf("template is not associated with projectId %s", projectId))), nil
			}
		}

		if projectId == "" {
			if template.Meta.ProjectId == nil {
				return api.Response(http.StatusBadRequest, api.NewValidationError("projectId must be provided for non-project scoped template")), nil
			}
			projectId = fmt.Sprintf("%d", *template.Meta.ProjectId)
		}

		if template.Spec.SpecificEnv != nil {
			copyErr := convutils.CopySpecs(&workspace.SpecificRunParams, template.Spec.SpecificEnv)
			if copyErr != nil {
				return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, "failed to load the spec of the template", copyErr)), nil
			}
		}

		workspace.Assets = template.Spec.Assets
	}

	effectivePolicy, policyErr := policy_compliance.GetEffectivePolicyOfProject(ctx, policies.Get().Client(), projectId, auth.WorkloadTypeWorkspace)
	if err != nil {
		// if the tenant is not using policies yet, just return the workload as is
		if policyErr.Code == http.StatusNotFound {
			return api.Response(http.StatusOK, workspace), nil
		}
		return api.Response(int(policyErr.Code), policyErr), nil
	}

	workspaceAsMap, jsonErr := gojsonq.LoadJson(workspace)
	if jsonErr != nil {
		return api.Response(int(err.Code), err), nil
	}

	result, err := flat_workloads.FlatToAssetsBased(workspaceAsMap, effectivePolicy.DefaultsV2, effectivePolicy.RulesV2)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	var resultWorkspace api.WorkspaceV2
	jsonErr = json.Unmarshal(gojsonq.ToBytes(result), &resultWorkspace)
	if jsonErr != nil {
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusOK, resultWorkspace), nil
}

// ListWorkspaces is a placeholder, currently implemented by the "old" backend, so it should never arrive here
func (s *WorkspacesApiService) ListWorkspaces(context.Context /*clusterIdParam WAIT_FOR_BEING_DELETED old backend*/, string /*nodeIdParam*/, string /*filterParam*/, string,
	/*sortByParam*/ string /*sortDirectionParam*/, string /*fromParam*/, int32 /*limitParam*/, int32,
) (api.ImplResponse, error) {
	return api.Response(http.StatusNotImplemented, nil), nil
}

func (s *WorkspacesApiService) CreateWorkspaceV2(ctx context.Context, request api.WorkspaceCreationRequestV2) (api.ImplResponse, error) {
	specificParams, convErr := workspaceToGenericRunParams(request.SpecificRunParams)
	if convErr != nil {
		return NewErrorResponse(convErr), nil
	}
	dataSourceOverridesById := CreateDatasourceOverrideMapById(request.Assets.Datasources)

	createdWorkloadInfo, err := CreateFlatWorkload(ctx, auth.WorkloadTypeWorkspace, &request.Meta, &request.Assets, specificParams, nil, dataSourceOverridesById, nil)
	if err != nil {
		return NewSubmissionErrorResponse(err), nil
	}

	workspaceRequest := workloadsService.WorkspaceCreationRequest{
		Name:                 request.Meta.Name,
		UseGivenNameAsPrefix: convutils.Ref(request.Meta.UseGivenNameAsPrefix),
		ProjectId:            request.Meta.ProjectId,
		ClusterId:            request.Meta.ClusterId,
		Spec:                 &workloadsService.WorkspaceSpecSpec{},
	}

	jsonErr := json.Unmarshal(createdWorkloadInfo.WorkerPayloadAsBytes, workspaceRequest.Spec)
	if jsonErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to compose workload payload: %v", jsonErr), nil)), nil
	}

	workspace, submitErr := workloadsServiceClient.Get().Client().CreateWorkspace(ctx, &workspaceRequest)
	if submitErr != nil {
		return api.Response(int(submitErr.Code), submitErr), nil
	}

	result := api.WorkspaceV2{
		Assets:            ContainedAssetsToAssetRef(createdWorkloadInfo.WorkerAssets, dataSourceOverridesById),
		SpecificRunParams: request.SpecificRunParams,
	}

	copyErr := copier.Copy(&result.Meta, &workspace)
	if copyErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "failed to load workload creation response", copyErr)), nil
	}

	dbErr := FlatWorkloadUpdateDb(ctx, auth.WorkloadTypeWorkspace, &result.Meta, specificParams,
		createdWorkloadInfo.WorkerAssets, nil, nil, dataSourceOverridesById, nil)
	if dbErr != nil {
		return api.Response(http.StatusInternalServerError, dbErr), nil
	}

	//
	//   when creating new workload it will be good to check if there are auto-delete assets that no longer used.
	//   we do not do it in delete workload because workload can be forced to delete (by cluster sync, for example)
	//   so we are not premised that it is user who initiated the workload delete. we just do it as part of creation.
	//   ideally, there should be a microservice that takes ownership of the workload table and all related operations
	//   (create, delete, list) which are currently implemented by the NodeJS backend. once this happens. the auto delete
	//   should move to that microservice.
	//
	var autoDeleteErr *api.Error
	if agentSyncSupported(ctx, &request.Meta.ClusterId) {
		autoDeleteErr = dao.AutoDeletePvcAssets(ctx, dao.SoftDelete)
	} else {
		autoDeleteErr = dao.AutoDeletePvcAssets(ctx, dao.HardDelete)
	}
	if autoDeleteErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to auto-delete assets: %v", autoDeleteErr)
	}

	return api.Response(http.StatusOK, result), nil
}

func workspaceToGenericRunParams(workspaceParams *api.WorkspaceSpecificRunParams) (*api.SpecificRunParams, *api.Error) {
	if workspaceParams == nil {
		return nil, nil
	}

	var result api.SpecificRunParams

	err := copier.CopyWithOption(&result, workspaceParams, copier.Option{
		IgnoreEmpty: true,
	})
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to parse workload creation payload", err)
	}
	return &result, nil
}

func genericToWorkspaceRunParams(genericParams *api.SpecificRunParams) (*api.WorkspaceSpecificRunParams, *api.Error) {
	if genericParams == nil {
		return nil, nil
	}

	var result api.WorkspaceSpecificRunParams

	err := copier.CopyWithOption(&result, genericParams, copier.Option{
		IgnoreEmpty: true,
	})
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to parse workload creation payload", err)
	}
	return &result, nil
}

func (s *WorkspacesApiService) UpdateWorkspaceStatus(ctx context.Context, clusterId string, update api.WorkloadStatusUpdate) (api.ImplResponse, error) {
	return updateWorkloadStatus(ctx, clusterId, update)
}

func (s *WorkspacesApiService) SyncWorkspaceStatus(ctx context.Context, clusterId string, updates []api.WorkloadStatusUpdate) (api.ImplResponse, error) {
	return syncWorkloadStatus(ctx, clusterId, updates)
}

func (s *WorkspacesApiService) DeleteWorkspaceById(ctx context.Context, clusterId string, workloadId string) (api.ImplResponse, error) {
	return deleteWorkload(ctx, clusterId, workloadId)
}
