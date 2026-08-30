package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-common-packages/runaidb"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
)

// ModelInferenceApiService is a service that implements the logic for the ModelInferenceAPIServicer
// This service should implement the business logic for every endpoint for the ModelInferenceApi API.
// Include any external packages or services that will be required by this service.
type ModelInferenceApiService struct {
	ModelAPIServicer     *api.ModelAPIServicer
	InferenceAPIServicer *api.InferencesAPIServicer
}

// NewModelInferenceApiService creates a default api service
func NewModelInferenceApiService(modelAPIServicer *api.ModelAPIServicer, inferenceAPIServicer *api.InferencesAPIServicer) api.ModelInferenceAPIServicer {
	return &ModelInferenceApiService{
		ModelAPIServicer:     modelAPIServicer,
		InferenceAPIServicer: inferenceAPIServicer,
	}
}

func initWorkloadCreationRequest(modelInferenceRequest *api.ModelInferenceCreationRequest, modelAssetIds *api.AssetsIds, computeAsset string) api.WorkloadCreationRequest {
	workloadRequestAssets := api.AssetsIds{
		Environment: modelAssetIds.Environment,
		Compute:     convutils.Ref(computeAsset),
	}

	workloadRequest := api.WorkloadCreationRequest{
		Name:        modelInferenceRequest.Name,
		ClusterId:   modelInferenceRequest.ClusterId,
		ProjectId:   modelInferenceRequest.ProjectId,
		Namespace:   modelInferenceRequest.Namespace,
		Assets:      workloadRequestAssets,
		SpecificEnv: modelInferenceRequest.SpecificEnv,
	}

	return workloadRequest
}

// CreateModelInference - Create model inference.
func (s *ModelInferenceApiService) CreateModelInference(ctx context.Context, request api.ModelInferenceCreationRequest) (api.ImplResponse, error) {
	// get model asset that will be used to create the inference
	assetResult, assetErr := (*s.ModelAPIServicer).GetModelAssetById(ctx, request.ModelId, false, 0, "", "")
	if assetErr != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("Model asset id %s is invalid in create model inference request", request.ModelId), assetErr)), nil
	}
	if assetResult.Code != http.StatusOK {
		return assetResult, nil
	}

	// init a WorkloadCreationRequest struct
	modelAsset := assetResult.Body.(*api.ModelAsset)
	workloadRequest := initWorkloadCreationRequest(&request, &modelAsset.Spec.Assets, request.Assets.Compute)
	// if SpecificEnv or specificEnv.EnvironmentVariables is not provided, load the environment variables from the environment asset
	if workloadRequest.SpecificEnv == nil || workloadRequest.SpecificEnv.EnvironmentVariables == nil {
		envVars, err := loadEnvironmentVariablesFromEnvironmentAsset(ctx, modelAsset.Spec.Assets.Environment)
		if err != nil {
			return NewErrorResponse(err), nil
		}
		if envVars != nil {
			if workloadRequest.SpecificEnv == nil {
				workloadRequest.SpecificEnv = &api.SpecificRunParams{}
			}

			workloadRequest.SpecificEnv.EnvironmentVariables = envVars
		}
	}

	// update the workloadRequest environment variables with the model HuggingfaceId
	if modelAsset.Spec.HuggingfaceId != nil && workloadRequest.SpecificEnv != nil && workloadRequest.SpecificEnv.EnvironmentVariables != nil {
		updateEnvironmentVariablesWithModelName(workloadRequest.SpecificEnv.EnvironmentVariables, *modelAsset.Spec.HuggingfaceId)
	}

	genericWorkload, err := CreateFlatWorkloadBackwardComp(ctx, auth.WorkloadTypeInference, workloadRequest)
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

func loadEnvironmentVariablesFromEnvironmentAsset(ctx context.Context, environmentId string) (*[]api.EnvironmentVariableOfAsset, *api.Error) {
	assetIds := []string{environmentId}

	// When loading environment variables from an environment asset there aren't credentials from run param
	workloadAssets, _, dbError := dao.ListWorkloadAssets(assetIds, nil, runaidb.DB().WithContext(ctx))
	if dbError != nil {
		return nil, dbError
	}

	// find the environment asset and return its environment variables
	for _, workloadAsset := range workloadAssets {
		if workloadAsset.AssetsEnvironment != nil {
			return workloadAsset.AssetsEnvironment.EnvironmentVariables, nil
		}
	}

	// not found, return an error
	return nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("Environment asset id %s is invalid in create model inference request", environmentId), nil)
}

func updateEnvironmentVariablesWithModelName(environmentVariables *[]api.EnvironmentVariableOfAsset, modelName string) {
	filteredEnvs := (*environmentVariables)[:0]

	for _, env := range *environmentVariables {
		if env.Name != nil && *env.Name == MODEL_ASSET_ENV_VAR_RUNAI_MODEL {
			env.Value = &modelName
		}
		if env.Name != nil && *env.Name == MODEL_ASSET_ENV_VAR_RUNAI_MODEL_NAME {
			// there is no need to use the MODEL_ASSET_ENV_VAR_RUNAI_MODEL as an environment variable when we don't
			// use a local model, so we skip it
			continue
		}
		filteredEnvs = append(filteredEnvs, env)
	}

	*environmentVariables = filteredEnvs
}

func (s *ModelInferenceApiService) CreateModelInferenceV2(ctx context.Context, request api.ModelInferenceCreationRequestV2) (api.ImplResponse, error) {
	// get model asset that will be used to create the inference
	modelAssetResult, assetErr := (*s.ModelAPIServicer).GetModelAssetById(ctx, request.Assets.Model, false, 0, "", "")
	if assetErr != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("Model asset id %s is invalid in create model inference request", request.Assets.Model), assetErr)), nil
	}
	if modelAssetResult.Code != http.StatusOK {
		return modelAssetResult, nil
	}

	modelAsset := modelAssetResult.Body.(*api.ModelAsset)
	envAssetId := modelAsset.Spec.Assets.Environment

	// init a WorkloadCreationRequestV2 struct
	inferenceRequest := api.InferenceCreationRequestV2{
		Meta: request.Meta,
		Assets: api.AssetsIds{
			Environment: envAssetId,
			Compute:     convutils.Ref(request.Assets.Compute),
		},
		SpecificRunParams: request.SpecificRunParams,
	}

	// take care of required environment variables in specific run params

	// if no environment variables are provided via SpecificRunParams, load them from the environment asset in order
	// to update them with the model name
	if inferenceRequest.SpecificRunParams == nil || inferenceRequest.SpecificRunParams.EnvironmentVariables == nil {
		envVars, err := loadEnvironmentVariablesFromEnvironmentAsset(ctx, modelAsset.Spec.Assets.Environment)
		if err != nil {
			return NewErrorResponse(err), nil
		}
		if envVars != nil {
			if inferenceRequest.SpecificRunParams == nil {
				inferenceRequest.SpecificRunParams = &api.InferenceSpecificRunParams{}
			}

			inferenceRequest.SpecificRunParams.EnvironmentVariables = envVars
		}
	}

	// update the workloadRequest environment variables with the model HuggingfaceId
	if modelAsset.Spec.HuggingfaceId != nil && inferenceRequest.SpecificRunParams != nil && inferenceRequest.SpecificRunParams.EnvironmentVariables != nil {
		updateEnvironmentVariablesWithModelName(inferenceRequest.SpecificRunParams.EnvironmentVariables, *modelAsset.Spec.HuggingfaceId)
	}

	// continue as a regular inference creation
	return (*s.InferenceAPIServicer).CreateInferenceV2(ctx, inferenceRequest)
}
