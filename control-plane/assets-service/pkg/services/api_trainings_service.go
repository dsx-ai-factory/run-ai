/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/policy_compliance"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/flat_workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	policies "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	workloadsService "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/jinzhu/copier"
)

var (
	assetsWorkloadCracker = jsonutils.NewJsonCracker(api.TrainingV2{})
	fldNameDistributed    = assetsWorkloadCracker.NameOf(&assetsWorkloadCracker.Struct().Distributed)
)

var (
	assetsDistributedCracker = jsonutils.NewJsonCracker(api.DistributedData{})
	fldNameMaster            = assetsDistributedCracker.NameOf(&assetsDistributedCracker.Struct().Master)
)

var (
	distributedMasterCracker = jsonutils.NewJsonCracker(api.DistributedMasterData{})
	fldNameSameAsWorker      = distributedMasterCracker.NameOf(&distributedMasterCracker.Struct().SameAsWorker)
)

const (
	distReplicaMaster = string(dbmodel.DistReplicaTypeMaster)
	distReplicaWorker = string(dbmodel.DistReplicaTypeWorker)
)

// TrainingsApiService is a service that implements the logic for the TrainingsAPIServicer
type TrainingsApiService struct{}

func NewTrainingsApiService() *TrainingsApiService {
	return &TrainingsApiService{}
}

func (s *TrainingsApiService) CreateTraining(ctx context.Context, request api.WorkloadCreationRequest) (api.ImplResponse, error) {
	workloadType := auth.WorkloadTypeTraining
	if request.Distributed != nil {
		workloadType = auth.WorkloadTypeDistributed
	}

	genericWorkload, err := CreateFlatWorkloadBackwardComp(ctx, workloadType, request)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	training := api.Training{
		Meta:     genericWorkload.Meta,
		Spec:     genericWorkload.Spec,
		Workload: genericWorkload.Workload,
		Job:      genericWorkload.Job,
	}

	return api.Response(http.StatusOK, training), nil
}

func getTrainingByIdV2(ctx context.Context, id string) (*api.TrainingV2, *api.Error) {
	genericWorkload, getError := GetWorkloadById(ctx, auth.WorkloadTypeTraining, id)
	if getError != nil {
		return nil, getError
	}

	training := api.TrainingV2{
		Meta:   *genericWorkload.ToWorkloadMeta(),
		Assets: genericWorkload.Spec.Assets,
	}

	var err *api.Error
	training.SpecificRunParams, err = genericToTrainingRunParams(genericWorkload.Spec.SpecificEnv)
	if err != nil {
		return nil, err
	}

	if genericWorkload.Spec.Distributed != nil {
		training.Distributed = &api.DistributedData{
			NumWorkers: genericWorkload.Spec.Distributed.NumWorkers,
			Framework:  genericWorkload.Spec.Distributed.DistFramework,
		}

		noMasterRequested := convutils.BoolIsTrue(genericWorkload.Spec.Distributed.NoMaster)

		if !noMasterRequested { // no master must be specifically requested, otherwise we assume there is master replica
			training.Distributed.Master = &api.DistributedMasterData{}

			if genericWorkload.Spec.Distributed.Master.Assets == nil {
				training.Distributed.Master.SameAsWorker = true
			} else {
				training.Distributed.Master.Assets = genericWorkload.Spec.Distributed.Master.Assets

				training.Distributed.Master.SpecificRunParams, err = genericToTrainingRunParams(genericWorkload.Spec.Distributed.Master.SpecificEnv)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return &training, nil
}

func (s *TrainingsApiService) GetTrainingByIdV2(ctx context.Context, id string) (api.ImplResponse, error) {
	response, err := getTrainingByIdV2(ctx, id)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *TrainingsApiService) GetTrainingMerged(ctx context.Context, projectId string, workloadId string, templateId string, isDistributed bool, isDistributedInference bool) (api.ImplResponse, error) {
	err := validateMergeParams(projectId, workloadId, templateId, isDistributed)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	//
	//   load the workspace, if workloadId is provided by the caller. otherwise, initialize it to empty workspace
	//
	training := &api.TrainingV2{
		SpecificRunParams: &api.TrainingSpecificRunParams{},
	}

	if workloadId != "" {
		var getError *api.Error

		training, getError = getTrainingByIdV2(ctx, workloadId)
		if getError != nil {
			return api.Response(int(getError.Code), getError), nil
		}

		projectId = training.Meta.ProjectId
		isDistributed = training.Distributed != nil && training.Distributed.Framework != nil
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

		copyErr := convutils.CopySpecs(&training.SpecificRunParams, template.Spec.SpecificEnv)
		if copyErr != nil {
			return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, "failed to load the spec of the template", copyErr)), nil
		}

		training.Assets = template.Spec.Assets

		if isDistributed {
			// template that is used by both master and worker we have to copy the specific run params
			// to the master as well.
			training.Distributed = &api.DistributedData{
				Master: &api.DistributedMasterData{
					SpecificRunParams: training.SpecificRunParams,
					Assets:            &api.AssetsRefOptional{},
				},
			}

			copier.Copy(training.Distributed.Master.Assets, template.Spec.Assets)
		}
	}

	workloadType := auth.WorkloadTypeTraining
	if isDistributed {
		workloadType = auth.WorkloadTypeDistributed
	}

	effectivePolicy, policyErr := policy_compliance.GetEffectivePolicyOfProject(ctx, policies.Get().Client(), projectId, workloadsService.WorkloadType(workloadType))
	if policyErr != nil {
		// if the tenant is not using policies yet, just return the workload as is
		if policyErr.Code == http.StatusNotFound {
			return api.Response(http.StatusOK, training), nil
		}
		return api.Response(int(policyErr.Code), policyErr), nil
	}

	trainingAsMap, jsonErr := gojsonq.LoadJson(training)
	if jsonErr != nil {
		return api.Response(int(err.Code), err), nil
	}

	var result map[string]interface{}

	if isDistributed {
		defaults := convutils.MapMe(effectivePolicy.DefaultsV2.AsMAp())
		rules := convutils.MapMe(effectivePolicy.RulesV2.AsMAp())

		result, err = flat_workloads.FlatToAssetsBased(trainingAsMap, convutils.MapMe(defaults[distReplicaWorker]), convutils.MapMe(rules[distReplicaWorker]))
		if err != nil {
			return api.Response(int(err.Code), err), nil
		}

		distributedSection := jsonutils.AddSection(result, fldNameDistributed)
		masterAsMap := jsonutils.AddSection(distributedSection, fldNameMaster)

		masterSameAsWorker := training.Distributed != nil && training.Distributed.Master != nil && training.Distributed.Master.SameAsWorker

		if masterSameAsWorker {
			distributedSection[fldNameMaster] = map[string]interface{}{
				fldNameSameAsWorker: true,
			}
		} else {
			mergedMaster, err := flat_workloads.FlatToAssetsBased(masterAsMap, convutils.MapMe(defaults[distReplicaMaster]), convutils.MapMe(rules[distReplicaMaster]))
			if err != nil {
				return api.Response(int(err.Code), err), nil
			}

			distributedSection[fldNameMaster] = mergedMaster
		}
	} else {
		result, err = flat_workloads.FlatToAssetsBased(trainingAsMap, effectivePolicy.DefaultsV2, effectivePolicy.RulesV2)
		if err != nil {
			return api.Response(int(err.Code), err), nil
		}
	}

	var resultTraining api.TrainingV2
	jsonErr = json.Unmarshal(gojsonq.ToBytes(result), &resultTraining)
	if jsonErr != nil {
		return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, "failed to compose response", jsonErr)), nil
	}

	return api.Response(http.StatusOK, resultTraining), nil
}

// ListTrainings is a placeholder, currently implemented by the "old" backend, so it should never arrive here
func (s *TrainingsApiService) ListTrainings(context.Context /*clusterIdParam WAIT_FOR_BEING_DELETED old backend*/, string /*nodeIdParam*/, string /*filterParam*/, string,
	/*sortByParam*/ string /*sortDirectionParam*/, string /*fromParam*/, int32 /*limitParam*/, int32,
) (api.ImplResponse, error) {
	return api.Response(http.StatusNotImplemented, nil), nil
}

func (s *TrainingsApiService) CreateTrainingV2(ctx context.Context, request api.TrainingCreationRequestV2) (api.ImplResponse, error) {
	specificParams, convErr := trainingToGenericRunParams(request.SpecificRunParams)
	if convErr != nil {
		return NewErrorResponse(convErr), nil
	}

	dataSourceOverridesById := CreateDatasourceOverrideMapById(request.Assets.Datasources)
	var masterDataSourceOverridesById map[string]*api.DataSourceOverrides = nil
	if request.Distributed != nil {
		masterDataSourceOverridesById = dataSourceOverridesById
		if request.Distributed.Master != nil && request.Distributed.Master.Assets != nil && request.Distributed.Master.Assets.Datasources != nil {
			masterDataSourceOverridesById = CreateDatasourceOverrideMapById(request.Distributed.Master.Assets.Datasources)
		}
	}

	createdWorkloadInfo, err := CreateFlatWorkload(ctx, auth.WorkloadTypeTraining, &request.Meta, &request.Assets, specificParams, request.Distributed, dataSourceOverridesById, masterDataSourceOverridesById)
	if err != nil {
		return NewSubmissionErrorResponse(err), nil
	}

	result := api.TrainingV2{
		Assets:            ContainedAssetsToAssetRef(createdWorkloadInfo.WorkerAssets, dataSourceOverridesById),
		SpecificRunParams: request.SpecificRunParams,
	}

	var workloadType string

	if request.Distributed == nil {
		workloadType = auth.WorkloadTypeTraining

		trainingRequest := workloadsService.TrainingCreationRequest{
			Name:                 request.Meta.Name,
			UseGivenNameAsPrefix: convutils.Ref(request.Meta.UseGivenNameAsPrefix),
			ProjectId:            request.Meta.ProjectId,
			ClusterId:            request.Meta.ClusterId,
			Spec:                 &workloadsService.TrainingSpecSpec{},
		}

		jsonErr := json.Unmarshal(createdWorkloadInfo.WorkerPayloadAsBytes, trainingRequest.Spec)
		if jsonErr != nil {
			return NewErrorResponse(api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to compose workload payload: %v", jsonErr), nil)), nil
		}

		training, submitErr := workloadsServiceClient.Get().Client().CreateTraining(ctx, &trainingRequest)
		if submitErr != nil {
			return api.Response(int(submitErr.Code), submitErr), nil
		}

		copyErr := copier.Copy(&result.Meta, &training)
		if copyErr != nil {
			return NewErrorResponse(api.NewError(http.StatusInternalServerError, "failed to load workload creation response", copyErr)), nil
		}
	} else {
		workloadType = auth.WorkloadTypeDistributed

		result.Distributed = &api.DistributedData{
			NumWorkers: request.Distributed.NumWorkers,
			Framework:  request.Distributed.Framework,
		}

		if request.Distributed.Master != nil {
			result.Distributed.Master = &api.DistributedMasterData{
				SameAsWorker: request.Distributed.Master.SameAsWorker,
			}

			if !request.Distributed.Master.SameAsWorker {
				result.Distributed.Master.SpecificRunParams = request.Distributed.Master.SpecificRunParams

				assetsRef := ContainedAssetsToAssetRef(createdWorkloadInfo.MasterAssets, masterDataSourceOverridesById)
				result.Distributed.Master.Assets = &api.AssetsRefOptional{}
				copier.Copy(result.Distributed.Master.Assets, &assetsRef)
			}
		}

		distributedRequest := workloadsService.DistributedCreationRequest{
			Name:                 request.Meta.Name,
			UseGivenNameAsPrefix: convutils.Ref(request.Meta.UseGivenNameAsPrefix),
			ProjectId:            request.Meta.ProjectId,
			ClusterId:            request.Meta.ClusterId,
			Spec:                 &workloadsService.DistributedSpecSpec{},
		}

		jsonErr := json.Unmarshal(createdWorkloadInfo.WorkerPayloadAsBytes, distributedRequest.Spec)
		if jsonErr != nil {
			return NewErrorResponse(api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to compose workload payload: %v", jsonErr), nil)), nil
		}

		distributedRequest.Spec.NumWorkers = *workloadsService.NewNullableInt32(&request.Distributed.NumWorkers)

		framework := utils.Deref(request.Distributed.Framework)
		distributedRequest.Spec.DistributedFramework = *workloadsService.NewNullableDistributedFramework(convutils.Ref(workloadsService.DistributedFramework(framework)))

		if request.Distributed.Master != nil {
			if request.Distributed.Master.SameAsWorker {
				distributedRequest.MasterSpecSameAsWorker = *workloadsService.NewNullableBool(convutils.Ref(true))
			} else {
				distributedRequest.MasterSpec = workloadsService.NullableMasterSpec{}

				jsonErr := json.Unmarshal(createdWorkloadInfo.MasterPayloadAsBytes, &distributedRequest.MasterSpec)
				if jsonErr != nil {
					return NewErrorResponse(api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to compose workload master payload: %v", jsonErr), nil)), nil
				}
			}
		}

		distributed, submitErr := workloadsServiceClient.Get().Client().CreateDistributed(ctx, &distributedRequest)
		if submitErr != nil {
			return api.Response(int(submitErr.Code), submitErr), nil
		}

		copyErr := copier.Copy(&result.Meta, &distributed)
		if copyErr != nil {
			return NewErrorResponse(api.NewError(http.StatusInternalServerError, "failed to load workload creation response", copyErr)), nil
		}
	}

	dbErr := FlatWorkloadUpdateDb(ctx, workloadType, &result.Meta, specificParams,
		createdWorkloadInfo.WorkerAssets, createdWorkloadInfo.MasterAssets, result.Distributed, dataSourceOverridesById, masterDataSourceOverridesById)
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

func trainingToGenericRunParams(trainingParams *api.TrainingSpecificRunParams) (*api.SpecificRunParams, *api.Error) {
	if trainingParams == nil {
		return nil, nil
	}

	var result api.SpecificRunParams

	err := copier.CopyWithOption(&result, trainingParams, copier.Option{
		IgnoreEmpty: true,
	})
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to parse workload creation payload", err)
	}

	return &result, nil
}

func genericToTrainingRunParams(genericParams *api.SpecificRunParams) (*api.TrainingSpecificRunParams, *api.Error) {
	if genericParams == nil {
		return nil, nil
	}

	var result api.TrainingSpecificRunParams

	err := copier.CopyWithOption(&result, genericParams, copier.Option{
		IgnoreEmpty: true,
	})
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to parse workload creation payload", err)
	}
	return &result, nil
}

func validateMergeParams(projectId string, workloadId string, templateId string, isDistributed bool) *api.Error {
	if projectId == "" && workloadId == "" && templateId == "" {
		return api.NewValidationError("either workloadId, projectId or templateId must be provided")
	}

	if workloadId != "" {
		if projectId != "" {
			return api.NewValidationError("workloadId and projectId are mutually exclusive")
		}
		if templateId != "" {
			return api.NewValidationError("workloadId and templateId are mutually exclusive")
		}
	}

	if isDistributed && workloadId != "" {
		return api.NewValidationError("isDistributed and workloadId are mutually exclusive")
	}

	return nil
}

func (s *TrainingsApiService) UpdateTrainingStatus(ctx context.Context, clusterId string, update api.WorkloadStatusUpdate) (api.ImplResponse, error) {
	return updateWorkloadStatus(ctx, clusterId, update)
}

func (s *TrainingsApiService) SyncTrainingStatus(ctx context.Context, clusterId string, updates []api.WorkloadStatusUpdate) (api.ImplResponse, error) {
	return syncWorkloadStatus(ctx, clusterId, updates)
}

func (s *TrainingsApiService) DeleteTrainingById(ctx context.Context, clusterId string, workloadId string) (api.ImplResponse, error) {
	return deleteWorkload(ctx, clusterId, workloadId)
}

func (s *TrainingsApiService) UpdateDistributedStatus(ctx context.Context, clusterId string, update api.WorkloadStatusUpdate) (api.ImplResponse, error) {
	return updateWorkloadStatus(ctx, clusterId, update)
}

func (s *TrainingsApiService) SyncDistributedStatus(ctx context.Context, clusterId string, updates []api.WorkloadStatusUpdate) (api.ImplResponse, error) {
	return syncWorkloadStatus(ctx, clusterId, updates)
}

func (s *TrainingsApiService) DeleteDistributedById(ctx context.Context, clusterId string, workloadId string) (api.ImplResponse, error) {
	return deleteWorkload(ctx, clusterId, workloadId)
}
