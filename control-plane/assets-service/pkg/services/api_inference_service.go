/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/policy_compliance"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"net/http"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/flat_workloads"
	policies "github.com/run-ai/runai/control-plane/workloads-and-policies/clients/policies"
	"github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"

	"github.com/jinzhu/copier"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/workloads-service"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/config"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	workloadsdb "github.com/run-ai/runai/control-plane/assets-service/pkg/database/workloads"
	hf "github.com/run-ai/runai/control-plane/assets-service/pkg/services/huggingface"
	backend "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	workloadsService "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

// InferencesApiService is a service that implements the logic for the InferencesAPIServicer
// This service should implement the business logic for every endpoint for the InferencseApi API.
// Include any external packages or services that will be required by this service.
type InferencesApiService struct {
	hfModelInfoApi                         hf.ModelInfoApiInterface
	semaphoreConcurrentHuggingFaceRequests chan struct{} // a semaphore that limits the number of concurrent requests to the model hugging face info api
}

// NewInferencesApiService creates a default api service
func NewInferencesApiService() api.InferencesAPIServicer {
	hfc := runai_config.Get[*config.HuggingFaceConfig]()
	hfModelApiInterface := hf.NewModelInfoApi(
		hfc.ModelInfoUrl,
		hfc.ModelFileUrl,
	)

	// use mock object if requested (for testing)
	if hfc.ModelInfoUrl == "/mock" {
		hfModelApiInterface = hf.NewModelInfoApiMock()
	}

	return &InferencesApiService{
		hfModelInfoApi:                         hfModelApiInterface,
		semaphoreConcurrentHuggingFaceRequests: make(chan struct{}, hfc.MaxConcurrentRequests),
	}
}

// CreateInference - Create inference workload.
func (s *InferencesApiService) CreateInference(ctx context.Context, request api.WorkloadCreationRequest) (api.ImplResponse, error) {
	genericWorkload, err := CreateFlatWorkloadBackwardComp(ctx, auth.WorkloadTypeInference, request)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	Inference := api.Inference{
		Meta:     genericWorkload.Meta,
		Spec:     genericWorkload.Spec,
		Workload: genericWorkload.Workload,
		Job:      genericWorkload.Job,
	}

	return api.Response(http.StatusOK, Inference), nil
}

// GetInferenceById - Get inference content.
func (s *InferencesApiService) GetInferenceById(ctx context.Context, id string) (api.ImplResponse, error) {
	genericWorkload, getError := GetWorkloadById(ctx, auth.WorkloadTypeInference, id)
	if getError != nil {
		return NewErrorResponse(getError), nil
	}

	inference := api.Inference{
		Meta:     genericWorkload.Meta,
		Spec:     genericWorkload.Spec,
		Workload: genericWorkload.Workload,
		Job:      genericWorkload.Job,
	}

	return api.Response(http.StatusOK, inference), nil
}

func (s *InferencesApiService) getInferenceByIdV2(ctx context.Context, id string) (*api.InferenceV2, *api.Error) {
	genericWorkload, getError := GetWorkloadById(ctx, auth.WorkloadTypeInference, id)
	if getError != nil {
		return nil, getError
	}

	result := api.InferenceV2{
		Meta:   *genericWorkload.ToWorkloadMeta(),
		Assets: genericWorkload.Spec.Assets,
	}

	var err *api.Error
	result.SpecificRunParams, err = genericToInferenceRunParams(genericWorkload.Spec.SpecificEnv)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *InferencesApiService) GetInferenceByIdV2(ctx context.Context, id string) (api.ImplResponse, error) {
	inference, err := s.getInferenceByIdV2(ctx, id)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusOK, inference), nil
}

// DeleteInferenceById - Delete inference workload leftover from backend.
func (s *InferencesApiService) DeleteInferenceById(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	workloadType := auth.WorkloadTypeInference

	workloadTypeAsAssetKind := auth.WorkloadTypeToAssetKind[workloadType]
	genericWorkload, getError := GetWorkloadById(ctx, auth.WorkloadTypeInference, workloadId)
	if getError != nil {
		return NewErrorResponse(getError), nil
	}

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfWorkload(genericWorkload.Meta.ProjectId), string(workloadTypeAsAssetKind), backend.ACTION_DELETE, false)
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	actionError := workloadsdb.DeleteWorkload(ctx, workloadType, workloadId)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	return NewDeletionResponse(), nil
}

func (s *InferencesApiService) CreateInferenceV2(ctx context.Context, request api.InferenceCreationRequestV2) (api.ImplResponse, error) {
	specificParams, convErr := inferenceToGenericRunParams(request.SpecificRunParams)
	if convErr != nil {
		return NewErrorResponse(convErr), nil
	}

	dataSourceOverridesById := CreateDatasourceOverrideMapById(request.Assets.Datasources)

	createdWorkloadInfo, err := CreateFlatWorkload(ctx, auth.WorkloadTypeInference, &request.Meta, &request.Assets, specificParams, nil, dataSourceOverridesById, nil)
	if err != nil {
		return NewSubmissionErrorResponse(err), nil
	}

	inferenceRequest := workloadsService.InferenceCreationRequest{
		Name:                 request.Meta.Name,
		UseGivenNameAsPrefix: convutils.Ref(request.Meta.UseGivenNameAsPrefix),
		ProjectId:            request.Meta.ProjectId,
		ClusterId:            request.Meta.ClusterId,
		Spec:                 &workloadsService.InferenceSpecSpec{},
	}

	jsonErr := json.Unmarshal(createdWorkloadInfo.WorkerPayloadAsBytes, inferenceRequest.Spec)
	if jsonErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to compose workload payload: %v", jsonErr), nil)), nil
	}

	inference, submitErr := workloadsServiceClient.Get().Client().CreateInference(ctx, &inferenceRequest)
	if submitErr != nil {
		return NewErrorResponse(api.NewError(int(submitErr.Code), submitErr.Message, nil)), nil
	}

	result := api.InferenceV2{
		Assets:            ContainedAssetsToAssetRef(createdWorkloadInfo.WorkerAssets, dataSourceOverridesById),
		SpecificRunParams: request.SpecificRunParams,
	}

	copyErr := copier.Copy(&result.Meta, &inference)
	if copyErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "failed to load workload creation response", copyErr)), nil
	}

	dbErr := FlatWorkloadUpdateDb(ctx, auth.WorkloadTypeInference, &result.Meta, specificParams,
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

func inferenceToGenericRunParams(inferenceParams *api.InferenceSpecificRunParams) (*api.SpecificRunParams, *api.Error) {
	if inferenceParams == nil {
		return nil, nil
	}

	var result api.SpecificRunParams

	err := copier.CopyWithOption(&result, inferenceParams, copier.Option{
		IgnoreEmpty: true,
	})
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to parse workload creation payload", err)
	}
	return &result, nil
}

func genericToInferenceRunParams(genericParams *api.SpecificRunParams) (*api.InferenceSpecificRunParams, *api.Error) {
	if genericParams == nil {
		return nil, nil
	}

	var result api.InferenceSpecificRunParams

	err := copier.CopyWithOption(&result, genericParams, copier.Option{
		IgnoreEmpty: true,
	})
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to parse workload creation payload", err)
	}
	return &result, nil
}

func (s *InferencesApiService) GetHuggingFaceModelInfo(ctx context.Context, huggingFaceModelInfoRequest api.HuggingFaceModelInfoRequest) (api.ImplResponse, error) {
	// we limit the max concurrent requests of this type to avoid overloading the hugging face api.
	// the limit is enforced by a semaphore and is configured in the config

	// try to acquire semaphore without blocking. instead, we return an error if the max concurrency is reached.
	select {
	case s.semaphoreConcurrentHuggingFaceRequests <- struct{}{}:
		defer func() { <-s.semaphoreConcurrentHuggingFaceRequests }() // release semaphore when done
		modelInfo, err := s.hfModelInfoApi.GetModelInfo(ctx, huggingFaceModelInfoRequest)
		if err != nil {
			return NewErrorResponse(err), nil
		}

		return api.Response(http.StatusOK, modelInfo), nil
	default:
		return NewErrorResponse(api.NewError(http.StatusServiceUnavailable, "too many requests", nil)), nil
	}
}

func (s *InferencesApiService) GetInferenceMerged(ctx context.Context, projectId string, workloadId string, templateId string) (api.ImplResponse, error) {
	err := validateMergeParams(projectId, workloadId, templateId, false)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	//
	//   load the workload, if workloadId is provided by the caller. otherwise, initialize it to empty workload
	//
	inference := &api.InferenceV2{
		SpecificRunParams: &api.InferenceSpecificRunParams{},
	}

	if workloadId != "" {
		var getError *api.Error

		inference, getError = s.getInferenceByIdV2(ctx, workloadId)
		if getError != nil {
			return api.Response(int(getError.Code), getError), nil
		}

		projectId = inference.Meta.ProjectId
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

		copyErr := convutils.CopySpecs(&inference.SpecificRunParams, template.Spec.SpecificEnv)
		if copyErr != nil {
			return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, "failed to load the spec of the template", copyErr)), nil
		}

		inference.Assets = template.Spec.Assets
	}

	effectivePolicy, policyErr := policy_compliance.GetEffectivePolicyOfProject(ctx, policies.Get().Client(), projectId, auth.WorkloadTypeInference)
	if policyErr != nil {
		// if the tenant is not using policies yet, just return the workload as is
		if policyErr.Code == http.StatusNotFound {
			return api.Response(http.StatusOK, inference), nil
		}
		return api.Response(int(policyErr.Code), policyErr), nil
	}

	inferenceAsMap, jsonErr := gojsonq.LoadJson(inference)
	if jsonErr != nil {
		return api.Response(int(err.Code), err), nil
	}

	result, err := flat_workloads.FlatToAssetsBased(inferenceAsMap, effectivePolicy.DefaultsV2, effectivePolicy.RulesV2)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	var resultInference api.InferenceV2
	jsonErr = json.Unmarshal(gojsonq.ToBytes(result), &resultInference)
	if jsonErr != nil {
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusOK, resultInference), nil
}
