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
	"net/http"
	"strconv"
	"time"

	k8s_objects_tracker "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v1/runai_workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/notifications"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	k8sObjectsApi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"

	k8sObjectsClient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	workloadsdb "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/flat_workloads"
	backend "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/workloadutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	scope "github.com/dsx-ai-factory/run-ai/runai-application-packages/scopeinfo"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NoneStorageClass is used as storage class config identifier for static provisioning of workloads
const NoneStorageClass = "runai-none"

const nonPrivate = false

type CreatedWorkloadInfo struct {
	WorkerPayloadAsBytes []byte
	WorkerAssets         []dbmodel.AssetsContainedInWorkload

	MasterPayloadAsBytes []byte
	MasterAssets         []dbmodel.AssetsContainedInWorkload
}

type CreateFlatWorkloadReplicaInfo struct {
	containedAssetIds       []string
	specificRunParams       *api.SpecificRunParams
	dataSourceOverridesById map[string]*api.DataSourceOverrides
}

func CreateFlatWorkload(ctx context.Context,
	workloadType string,
	meta *api.WorkloadCreationMeta,
	assets *api.AssetsIds,
	specificRunParams *api.SpecificRunParams,
	distributed *api.DistributedParams,
	dataSourceOverridesById map[string]*api.DataSourceOverrides,
	masterDataSourceOverridesById map[string]*api.DataSourceOverrides,
) (*CreatedWorkloadInfo, *workloadsApi.SubmissionErrorV2) {
	//
	//   check permission to create workload
	//
	targetScope := scope.NewProjectScope(meta.ProjectId)

	workloadTypeAsAssetKind := auth.WorkloadTypeToAssetKind[workloadType]

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, targetScope, string(workloadTypeAsAssetKind), backend.ACTION_CREATE, nonPrivate)
	if rbacErr != nil {
		return nil, &workloadsApi.SubmissionErrorV2{
			Code:    rbacErr.Code,
			Message: rbacErr.Message,
			Details: &rbacErr.Details,
		}
	}

	projectIdAsInt, convErr := strconv.Atoi(meta.ProjectId)
	if convErr != nil {
		return nil, &workloadsApi.SubmissionErrorV2{
			Code:    http.StatusBadRequest,
			Message: "non numeric project id",
			Details: convutils.Ref(convErr.Error()),
		}
	}

	projectIdAsNum := int32(projectIdAsInt)

	containedAssetIds, assetsError := dao.ValidateContainedAssets(ctx, convutils.Ref(workloadType), api.SCOPE_PROJECT, &meta.ClusterId, &projectIdAsNum, nil, assets)
	if assetsError != nil {
		return nil, toSubmissionError(assetsError)
	}

	workerReplicaInfo := CreateFlatWorkloadReplicaInfo{
		containedAssetIds:       containedAssetIds,
		specificRunParams:       specificRunParams,
		dataSourceOverridesById: dataSourceOverridesById,
	}

	var masterReplicaInfo *CreateFlatWorkloadReplicaInfo

	if distributed != nil && distributed.Master != nil {
		masterReplicaInfo = &CreateFlatWorkloadReplicaInfo{
			dataSourceOverridesById: masterDataSourceOverridesById,
		}

		if distributed.Master.SameAsWorker {
			if distributed.Master.Assets != nil {
				return nil, workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, "assets should not be provided when sameAsWorker is requested for the master replica")
			}
			if distributed.Master.SpecificRunParams != nil {
				return nil, workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, "specificRunParams should not be provided when sameAsWorker is requested for the master replica")
			}
		} else {
			if distributed.Master.Assets == nil {
				return nil, workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, "assets should be provided for the master replica")
			}

			masterSpecificRunParams, convErr := trainingToGenericRunParams(distributed.Master.SpecificRunParams)
			if convErr != nil {
				return nil, toSubmissionError(convErr)
			}

			masterReplicaInfo.specificRunParams = masterSpecificRunParams

			masterReplicaInfo.containedAssetIds = NewContainedIdsFromOptionalAssets(distributed.Master.Assets)
		}
	}

	return composeFlatWorkload(ctx, workloadType, meta.ClusterId, &workerReplicaInfo, masterReplicaInfo, nil)
}

func composeFlatWorkload(ctx context.Context,
	workloadType string,
	clusterId string,
	worker *CreateFlatWorkloadReplicaInfo,
	master *CreateFlatWorkloadReplicaInfo,
	options *flat_workloads.TranslateFlatWorkloadOptions,
) (*CreatedWorkloadInfo, *workloadsApi.SubmissionErrorV2) {

	if clusterId == "" {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, "mandatory clusterId is not provided")
	}

	//
	//   this API requires agent to sync from control plane to the cluster. old cluster will not have this capability
	//   and will have to continue using V1 APIs
	//
	capabilitiesChecker := checker.NewClusterCapabilitiesChecker(clusterId)

	clusterSupportFlatWorkloads, checkerError := capabilitiesChecker.ValidateClusterCapability(ctx, checker.MinVersionOfFlatWorkloadSubmission)
	if checkerError != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusServiceUnavailable, checkerError.Error())
	}

	if !clusterSupportFlatWorkloads {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, fmt.Sprintf("this API is incompatible with the version of cluster '%s'", clusterId))
	}

	workloadInfo, translationErr := translateFlatWorkload(ctx, workloadType, worker, master, options)
	if translationErr != nil {
		return nil, translationErr
	}

	//
	// even though workload controller validates the translated workload VS storage class
	// configurations, we have to validate it also here because non-existing assets are
	// translated to existing in the flat workload and in this translation the storage class
	// nam is not included
	//
	allAssets := append(workloadInfo.WorkerAssets, workloadInfo.MasterAssets...)

	_, storageClassErr := validateStorageClassPermissions(ctx, clusterId, allAssets)
	if storageClassErr != nil {
		return nil, toSubmissionError(storageClassErr)
	}

	return workloadInfo, nil
}

func translateFlatWorkload(ctx context.Context, workloadType string,
	worker *CreateFlatWorkloadReplicaInfo,
	master *CreateFlatWorkloadReplicaInfo,
	options *flat_workloads.TranslateFlatWorkloadOptions,
) (*CreatedWorkloadInfo, *workloadsApi.SubmissionErrorV2) {
	workerAssets, credentialsAssets, err := ValidateWorkloadRequest(ctx, workloadType, worker.containedAssetIds, worker.specificRunParams)
	if err != nil {
		return nil, toSubmissionError(err)
	}

	fixPortionForNonFractionalIssue(workerAssets)

	//
	//   translate the asset based payload to workloads-service API kind of payload
	//
	_, workerPayloadAsBytes, translateErr := flat_workloads.TranslateToFlatWorkload(ctx, workloadType, workerAssets, credentialsAssets, worker.specificRunParams, worker.dataSourceOverridesById, options)
	if translateErr != nil {
		return nil, toSubmissionError(translateErr)
	}

	//
	//   for distributed, we may also have master assets + specific run params
	//
	var masterPayloadAsBytes []byte
	var masterAssets []dbmodel.AssetsContainedInWorkload

	if master != nil {
		var masterCredentialsAssets map[string]dbmodel.CredentialsAssetSpecs

		masterAssets, masterCredentialsAssets, err = ValidateWorkloadRequest(ctx, workloadType, master.containedAssetIds, master.specificRunParams)
		if err != nil {
			return nil, toSubmissionError(err)
		}

		fixPortionForNonFractionalIssue(masterAssets)

		_, masterPayloadAsBytes, translateErr = flat_workloads.TranslateToFlatWorkload(ctx, workloadType, masterAssets, masterCredentialsAssets, master.specificRunParams, master.dataSourceOverridesById, options)
		if translateErr != nil {
			return nil, toSubmissionError(translateErr)
		}
	}

	return &CreatedWorkloadInfo{
		WorkerPayloadAsBytes: workerPayloadAsBytes,
		WorkerAssets:         workerAssets,

		MasterPayloadAsBytes: masterPayloadAsBytes,
		MasterAssets:         masterAssets,
	}, nil
}

// FlatWorkloadUpdateDb updates the workloads related tables with the information of the workload.
// we mainly need it for usage info and for copy&edit.
func FlatWorkloadUpdateDb(ctx context.Context, workloadType string, meta *api.WorkloadMeta, specificRunParams *api.SpecificRunParams,
	workerAssets []dbmodel.AssetsContainedInWorkload,
	masterAssets []dbmodel.AssetsContainedInWorkload,
	distributed *api.DistributedData, dataSourceOverridesById map[string]*api.DataSourceOverrides, masterDataSourceOverridesById map[string]*api.DataSourceOverrides,
) *api.Error {
	projectIdAsNum, convErr := strconv.Atoi(meta.ProjectId)
	if convErr != nil {
		return api.NewError(http.StatusBadRequest, "non numeric project id", convErr)
	}

	v1CreationRequest := api.WorkloadCreationRequest{
		Name:        meta.Name,
		ClusterId:   meta.ClusterId,
		ProjectId:   int32(projectIdAsNum),
		SpecificEnv: specificRunParams,
	}
	if distributed != nil {
		v1CreationRequest.Distributed = &api.Distributed{
			NumWorkers:    distributed.NumWorkers,
			NoMaster:      convutils.Ref(distributed.Master == nil),
			DistFramework: distributed.Framework,
		}
		if distributed.Master != nil {
			v1CreationRequest.Distributed.Master = &api.DistMaster{
				SpecificEnv: &api.SpecificRunParams{},
			}
			if distributed.Master.SpecificRunParams != nil {
				copyErr := convutils.CopySpecs(&v1CreationRequest.Distributed.Master.SpecificEnv, distributed.Master.SpecificRunParams)
				if copyErr != nil {
					return api.NewError(http.StatusBadRequest, fmt.Sprintf("failed to copy master specific env: %v", copyErr), copyErr)
				}
			}
		}
	}

	_, dbError := workloadsdb.CreateWorkload(ctx, workloadType, &v1CreationRequest, meta.WorkloadId, workerAssets, masterAssets, nil, dataSourceOverridesById, masterDataSourceOverridesById)
	if dbError != nil {
		return dbError
	}

	return nil
}

// CreateFlatWorkloadBackwardComp replaces the V1 CreateWorkload function by utilizing the workload composer instead of
// the soon to become obsolete sections creator
func CreateFlatWorkloadBackwardComp(ctx context.Context, workloadType string, request api.WorkloadCreationRequest) (*dbmodel.GenericWorkload, *api.Error) {
	workloadTypeAsAssetKind := auth.WorkloadTypeToAssetKind[workloadType]

	//
	//   load scope information
	//
	projectIdAsNum := request.ProjectId
	targetScope := scope.NewProjectScope(fmt.Sprintf("%d", projectIdAsNum))

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, targetScope, string(workloadTypeAsAssetKind), backend.ACTION_CREATE, nonPrivate)
	if rbacErr != nil {
		return nil, api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))
	}

	containedAssetIds, assetsError := dao.ValidateContainedAssets(ctx, convutils.Ref(workloadType), api.SCOPE_PROJECT, &request.ClusterId, &projectIdAsNum, nil, &request.Assets)
	if assetsError != nil {
		return nil, assetsError
	}

	workloadAssets, credentialsAssets, err := ValidateWorkloadRequest(ctx, workloadType, containedAssetIds, request.SpecificEnv)
	if err != nil {
		return nil, err
	}

	fixPortionForNonFractionalIssue(workloadAssets)

	//
	//   translate the asset based payload to workloads-service API kind of payload
	//
	var workloadSpec *workloadsApi.SupersetSpec

	dataSourceOverridesById := CreateDatasourceOverrideMapById(request.Assets.Datasources)
	masterDataSourceOverridesById := dataSourceOverridesById
	if (request.Distributed != nil) && (request.Distributed.Master != nil) && request.Distributed.Master.Assets.Datasources != nil {
		masterDataSourceOverridesById = CreateDatasourceOverrideMapById(request.Distributed.Master.Assets.Datasources)
	}

	_, payloadAsBytes, translateErr := flat_workloads.TranslateToFlatWorkload(ctx, workloadType, workloadAssets, credentialsAssets, request.SpecificEnv, dataSourceOverridesById, nil)
	if translateErr != nil {
		return nil, translateErr
	}

	jsonErr := json.Unmarshal(payloadAsBytes, &workloadSpec)
	if jsonErr != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to load the workload spec", jsonErr)
	}

	//
	//   also for distributed need to update the distributed master spec with details that in v1 were
	//   stored on the request
	//
	if request.Distributed != nil {
		framework := workloadsApi.DistributedFramework(convutils.Deref(request.Distributed.DistFramework, ""))
		workloadSpec.DistributedFramework = *workloadsApi.NewNullableDistributedFramework(&framework)

		workloadSpec.NumWorkers = *workloadsApi.NewNullableInt32(&request.Distributed.NumWorkers)
	}

	//
	//   for distributed, we may also have master assets + specific run params
	//
	masterAssets, masterSpec, masterErr := translateMasterToFlatWorkload(ctx, &request, workloadAssets, workloadSpec)
	if masterErr != nil {
		return nil, masterErr
	}

	allAssets := append(workloadAssets, masterAssets...)

	scErr := workloadutils.ValidateAndHandelServingConfigurationFields(ctx, workloadSpec, request.ClusterId)
	if scErr != nil {
		return nil, api.NewError(http.StatusBadRequest, scErr.Error(), scErr)
	}

	configs, storageClassErr := validateStorageClassPermissions(ctx, request.ClusterId, allAssets)
	if storageClassErr != nil {
		return nil, storageClassErr
	}

	workloadId := uuid.New()

	clusterWorkload := jobs.JobWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      request.Name,
			Namespace: request.Namespace,
		},
		Spec: jobs.JobSpec{},
	}

	controlPlaneWorkload := workloadsApi.SupersetWorkload{
		Spec: workloadSpec,
	}

	composerWorkloadType := workloadType
	if request.Distributed != nil {
		composerWorkloadType = composer.Distributed
	}

	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(request.ClusterId)

	options, checkerErr := determineComposerOptions(ctx, clusterCapabilityChecker)
	if checkerErr != nil {
		return nil, checkerErr
	}

	complimentaryAssetInfo, complimentaryErr := composer.LoadComplimentaryAssetInfo(workloadSpec.Annotations, true)
	if complimentaryErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to load asset info: %v", err)
	}

	notificationsComposer := notifications.NewNotificationsComposer(
		workloadType,
		request.Name,
		workloadId.String(),
		auth.SubjectIdOf(ctx),
		request.ClusterId,
		targetScope.ScopeId,
		"",
		time.Now().Truncate(time.Millisecond),
		nil,
		options)

	switch workloadsApi.WorkloadType(composerWorkloadType) {
	case workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE:
		return nil, api.NewError(http.StatusInternalServerError, "distributed inference does not support asset submission", nil)
	case workloadsApi.WORKLOADTYPE_DISTRIBUTED:
		distributedComposer := runai_workloads.NewDistributedWorkloadComposer(
			request.Name, auth.SubjectIdOf(ctx), notificationsComposer, options)

		distributedComposer.
			WithComplimentaryAssetInfo(workloadsApi.REPLICATYPE_WORKER, complimentaryAssetInfo)

		if masterSpec != nil {
			complimentaryAssetInfoMaster, err := composer.LoadComplimentaryAssetInfo(masterSpec.Annotations, true)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to load asset info of master spec: %v", err)
			}

			distributedComposer.WithComplimentaryAssetInfo(workloadsApi.REPLICATYPE_MASTER, complimentaryAssetInfoMaster)

			controlPlaneWorkload.MasterSpec.Set(masterSpec)
		} else if !convutils.BoolIsTrue(request.Distributed.NoMaster) {
			controlPlaneWorkload.MasterSpecSameAsWorker.Set(convutils.Ref(true))
		}

		distributedComposer.WithStorageClassConfiguration(configs)

		composerErr := distributedComposer.Compose(ctx, &controlPlaneWorkload, &clusterWorkload)
		if composerErr != nil {
			return nil, api.NewError(int(composerErr.Code), composerErr.Message, nil)
		}
	default:
		jobSpecComposer := runai_workloads.NewSupersetToJobSpecComposer(
			composerWorkloadType, request.Name, auth.SubjectIdOf(ctx), nil, nil, options)

		jobSpecComposer.WithStorageClassConfiguration(configs)

		composerErr := jobSpecComposer.Compose(ctx, workloadSpec, &clusterWorkload.Spec)()
		if composerErr == nil {
			composerErr = notificationsComposer.Compose(ctx, nil, complimentaryAssetInfo, workloadSpec)()
		}
		if composerErr != nil {
			return nil, api.NewError(int(composerErr.Code), composerErr.Message, nil)
		}
	}

	genericWorkload, dbError := workloadsdb.CreateWorkload(ctx, workloadType, &request, workloadId.String(),
		workloadAssets, masterAssets, &clusterWorkload, dataSourceOverridesById, masterDataSourceOverridesById)
	if dbError != nil {
		return nil, dbError
	}

	zap.S().Debugf("created %s %s for %s:%s", workloadType, genericWorkload.Meta.Id, genericWorkload.Meta.Name, request.Namespace)

	// Notify workload service in an async manner as failure is an option
	notificationCtx := context.WithoutCancel(ctx)
	go notifyWorkloadService(notificationCtx, *notificationsComposer.Output)

	//
	//   when creating new workload it will be good to check if there are auto-delete assets that no longer used.
	//   we do not do it in delete workload because workload can be forced to delete (by cluster sync, for example)
	//   so we are not premised that it is user who initiated the workload delete. we just do it as part of creation.
	//   ideally, there should be a microservice that takes ownership of the workload table and all related operations
	//   (create, delete, list) which are currently implemented by the NodeJS backend. once this happens. the auto delete
	//   should move to that microservice.
	//
	var autoDeleteErr *api.Error
	if agentSyncSupported(ctx, &request.ClusterId) {
		autoDeleteErr = dao.AutoDeletePvcAssets(ctx, dao.SoftDelete)
	} else {
		autoDeleteErr = dao.AutoDeletePvcAssets(ctx, dao.HardDelete)
	}
	if autoDeleteErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to auto-delete assets: %v", autoDeleteErr)
	}

	return genericWorkload, nil
}

func translateMasterToFlatWorkload(ctx context.Context, request *api.WorkloadCreationRequest,
	workerAssets []dbmodel.AssetsContainedInWorkload, workerSpec *workloadsApi.SupersetSpec,
) ([]dbmodel.AssetsContainedInWorkload, *workloadsApi.MasterSpec, *api.Error) {
	// option 1: no distributed at all
	if request.Distributed == nil {
		return nil, nil, nil
	}

	if request.Distributed.NumWorkers < 0 {
		return nil, nil, api.NewError(http.StatusBadRequest, "NumWorkers must not be negative", nil)
	}

	// option 2: no master requested
	if convutils.BoolIsTrue(request.Distributed.NoMaster) {
		framework := convutils.Deref(request.Distributed.DistFramework, "")

		if framework == api.DISTRIBUTEDFRAMEWORK_MPI || framework == api.DISTRIBUTEDFRAMEWORK_XG_BOOST {
			return nil, nil, api.NewError(http.StatusBadRequest, "NoMaster is not available for the chosen framework", nil)
		}

		if request.Distributed.Master != nil {
			return nil, nil, api.NewError(http.StatusBadRequest, "master should not be provided when noMaster is requested", nil)
		}
		return nil, nil, nil
	}

	// option 3: user want master to be identical to the worker
	if request.Distributed.Master == nil {
		return nil, nil, nil
	}

	// option 4: user want master different than worker
	var masterAssets []dbmodel.AssetsContainedInWorkload
	var err *api.Error

	if request.Distributed.Master.Assets.Environment == "" {
		return nil, nil, api.NewError(http.StatusBadRequest, "assets should be provided for the master replica", nil)
	}

	var masterCredentialsAssets map[string]dbmodel.CredentialsAssetSpecs

	masterAssets, masterCredentialsAssets, err = ValidateWorkloadRequest(ctx, auth.WorkloadTypeTraining,
		NewContainedIdsFromAssets(&request.Distributed.Master.Assets), request.Distributed.Master.SpecificEnv)
	if err != nil {
		return nil, nil, err
	}

	fixPortionForNonFractionalIssue(masterAssets)

	dataSourceOverridesById := CreateDatasourceOverrideMapById(request.Distributed.Master.Assets.Datasources)
	_, masterPayloadAsBytes, translateErr := flat_workloads.TranslateToFlatWorkload(ctx, auth.WorkloadTypeDistributed, masterAssets, masterCredentialsAssets,
		request.Distributed.Master.SpecificEnv, dataSourceOverridesById, nil)
	if translateErr != nil {
		return nil, nil, translateErr
	}

	var masterSpec workloadsApi.MasterSpec

	jsonErr := json.Unmarshal(masterPayloadAsBytes, &masterSpec)
	if jsonErr != nil {
		return nil, nil, api.NewError(http.StatusInternalServerError, "failed to load the workload spec", jsonErr)
	}

	return masterAssets, &masterSpec, nil
}

func toSubmissionError(err *api.Error) *workloadsApi.SubmissionErrorV2 {
	return &workloadsApi.SubmissionErrorV2{
		Code:    err.Code,
		Message: err.Message,
		Details: &err.Details,
	}
}

// this function is a workaround for asset-service compute assets which contains request type = PORTION
// while the asset itself is non-fractional. it attempts to create incorrect workload, because request
// type PORTION requires value in the gpuPortionRequest field. to avoid this issue, in case asset contains
// request type PORTION without having value in gpuPortionRequest - reset the request type to nil, indication
// non-fractional gpu request
// also when sending request type PORTION AND gpuPortionRequest == 1 AND gpuDevices > 1,
// we need to reset the request type and portionRequest in order to use multiple whole devices.
func fixPortionForNonFractionalIssue(assets []dbmodel.AssetsContainedInWorkload) {
	for idx, asset := range assets {
		if asset.Kind != api.ASSETKIND_COMPUTE || asset.AssetsCompute == nil {
			continue
		}

		computeSpec := asset.AssetsCompute.ComputeAssetSpec
		if computeSpec.GpuRequestType == nil {
			continue
		}
		if *computeSpec.GpuRequestType == api.GPUREQUESTTYPE_PORTION && computeSpec.GpuPortionRequest == nil {
			assets[idx].AssetsCompute.GpuRequestType = nil
			return
		}

		// When sending gpuDevicesRequest > 1 and sending portion == 1 - we ignore the portion since it is 100% of the devices.
		if computeSpec.GpuDevicesRequest != nil && *computeSpec.GpuDevicesRequest > 1 && computeSpec.GpuPortionRequest != nil && *computeSpec.GpuPortionRequest == 1 {
			assets[idx].AssetsCompute.GpuRequestType = nil
			assets[idx].AssetsCompute.GpuPortionRequest = nil
		}
	}
}

func NewContainedIdsFromOptionalAssets(src *api.OptionalAssetIds) []string {
	return NewContainedIds(src.Environment, src.Compute, src.Datasources)
}

func NewContainedIdsFromAssets(src *api.AssetsIds) []string {
	return NewContainedIds(src.Environment, src.Compute, src.Datasources)
}

func NewContainedIds(environment string, compute *string, datasources []api.AssetIdAndKind) []string {
	result := make([]string, 0, len(datasources)+2)

	result = append(result, environment)

	if compute != nil {
		result = append(result, *compute)
	}

	for _, datasource := range datasources {
		result = append(result, datasource.Id)
	}

	return result
}

// CreateDatasourceOverrideMapById creates a map between the asset id and it override data
func CreateDatasourceOverrideMapById(datasources []api.AssetIdAndKind) map[string]*api.DataSourceOverrides {
	dataSourceOverridesById := make(map[string]*api.DataSourceOverrides)
	for _, datasource := range datasources {
		if datasource.Overrides != nil {
			dataSourceOverridesById[datasource.Id] = datasource.Overrides
		}
	}
	return dataSourceOverridesById
}

func validateStorageClassPermissions(ctx context.Context, clusterId string, containedAssets []dbmodel.AssetsContainedInWorkload) (map[string]k8sObjectsApi.StorageClassConfiguration, *api.Error) {
	//
	//  collect the storage classes being used
	//
	storageClasses := collectStorageClasses(containedAssets)

	if len(storageClasses) == 0 {
		return nil, nil
	}

	storageClassesUnique := storageClasses.Keys()

	//
	//  need to validate that the storage classes are allowed
	//
	configs, err := k8s_objects_tracker.Get().GetStorageClassesConfigInt(ctx, clusterId, storageClassesUnique)
	if err != nil {
		return nil, api.NewError(int(err.Code), err.Message, nil)
	}

	for _, asset := range containedAssets {
		pvc := asset.AssetsPvc
		if pvc == nil || pvc.ClaimInfo == nil || pvc.ClaimInfo.StorageClass == nil {
			continue
		}

		storageClassName := *pvc.ClaimInfo.StorageClass
		if storageClassName == "" {
			storageClassName = NoneStorageClass
		}

		config, configDefined := configs[storageClassName]

		// if no permissions defined for this storage class, the assumption is not restricted
		if !configDefined || !convutils.HasValue(config.Permissions) {
			continue
		}

		// in assets that are not volumes, we need to validate against the allowedForWorkloads flag. note that
		// for existing pvcs we typically won't get the storage class name, so this validation is mainly for
		// non-existing pvcs.
		if convutils.BoolIsFalse(asset.AutoDelete) {
			if !k8sObjectsClient.IsStorageClassActionPermitted(config.Permissions.Get().AllowedForWorkloads) {
				return nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("%s is forbidden for running workloads", StorageClassNameForMessage(&pvc.PvcAssetSpec)), nil)
			}
			continue
		}

		// for volumes. we distinguish ephemeral from persistent volumes
		if convutils.BoolIsTrue(pvc.Ephemeral) {
			if !k8sObjectsClient.IsStorageClassActionPermitted(config.Permissions.Get().AllowedForEphemeralVolumes) {
				return nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("creating ephemeral volumes on %s is forbidden", StorageClassNameForMessage(&pvc.PvcAssetSpec)), nil)
			}
		} else {
			if !k8sObjectsClient.IsStorageClassActionPermitted(config.Permissions.Get().AllowedForPersistentVolumes) {
				return nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("creating persistent volumes on %s is forbidden", StorageClassNameForMessage(&pvc.PvcAssetSpec)), nil)
			}
		}
	}

	return configs, nil
}

func collectStorageClasses(containedAssets []dbmodel.AssetsContainedInWorkload) datastructs.Set[string] {
	result := datastructs.NewSet[string]()

	for _, asset := range containedAssets {
		if asset.AssetsPvc == nil {
			continue
		}

		if asset.AssetsPvc.ClaimInfo != nil {
			if convutils.IsEmptyOrNull(asset.AssetsPvc.ClaimInfo.StorageClass) {
				result.Add(NoneStorageClass)
			} else {
				result.Add(*asset.AssetsPvc.ClaimInfo.StorageClass)
			}
		}
	}

	return result
}

func StorageClassNameForMessage(pvc *api.PvcAssetSpec) string {
	var storageClassNameForMessage string
	if *pvc.ClaimInfo.StorageClass == "" {
		storageClassNameForMessage = "static provisioning"
	} else {
		storageClassNameForMessage = fmt.Sprintf("storage class %s", *pvc.ClaimInfo.StorageClass)
	}

	return storageClassNameForMessage
}

func determineComposerOptions(ctx context.Context, ClusterCapabilitiesChecker *checker.ClusterCapabilitiesChecker) (composer.FieldsComposerOptions, *api.Error) {
	options := composer.FieldsComposerOptions{
		UserCredentialsSupported: true,
	}

	checkerErr := composer_v2.SetClusterDependentOptions(ctx, &options, ClusterCapabilitiesChecker)
	if checkerErr != nil {
		return options, api.NewError(http.StatusServiceUnavailable, checkerErr.Error(), nil)
	}

	return options, nil
}
