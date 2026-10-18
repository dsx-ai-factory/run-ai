/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"

	authorization_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v1/runai_workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	clients "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/backend"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/conjungo"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	workloadMgrApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dinfwv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

func UpdateDistributedInferenceSpec(ctx context.Context, workloadId string, tenantSettingsClient clients.TenantSettingsClient, request api.UpdateRequest) (api.ImplResponse, error) {
	tenantId, parseErr := services_utils.ParseTenantId(ctx)
	if parseErr != nil {
		err := ToSubmissionError(parseErr)
		return api.Response(int(err.Code), err), nil
	}

	permittedScopes, scopesErr := authorization_client.ParsePermittedScopes(ctx)
	if scopesErr != nil {
		permittedScopes = nil
	}

	// Build minimal workload context to check the submitter
	wc := &workloadContext{
		workloadType: DistributedInference,
		tenantId:     int32(tenantId),
	}

	// Get the submitter for this workload type
	submitter := getAnyWorkloadSubmitter(wc)
	if submitter == nil {
		return api.Response(http.StatusInternalServerError, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, "failed to get submitter for workload type distributed inference")), nil
	}

	// Determine the original submission method that was used for creating the workload
	var origSubmissionMethod SubmissionMethods

	// for any-workload update we need both the anyWorkload from the workload manager, and the
	// snapshot from the snapshots table which holds a copy of the submitted workload in native notation
	var anyWorkload *workloadMgrApi.WorkloadV2GetResponse
	var snapshot *models_crud.WorkloadV2Snapshot

	var clusterId string

	// let's first assume the workload was submitted in the native path. in this case, we're
	// expecting to find it in the workloads table of service-crud
	nativeWorkload, dbErr := dao_crud.DbGetWorkload(ctx, tenantId, DistributedInference, workloadId, permittedScopes)
	if dbErr != nil && dbErr.Code != http.StatusNotFound {
		err := ToSubmissionError(dbErr)
		return api.Response(int(err.Code), err), nil
	}

	if nativeWorkload != nil {
		// we found the record in the crud table, so we know it has been submitted in native path
		origSubmissionMethod = SubmissionMethodNative
		clusterId = nativeWorkload.ClusterId
	} else {
		origSubmissionMethod = SubmissionMethodAnyWorkload

		// assuming any workload submission, we expect to find both a snapshot in the db and a workload
		// manager record of it
		var err *workloadMgrApi.Error
		anyWorkload, err = getWorkloadFromWorkloadManager(ctx, workloadId)
		if err != nil {
			return api.Response(int(err.Code), err.Message), nil
		}

		// if the workloads db never heard of it, and also the workload manager never heard
		// of it, we can report back that the workload cannot be found
		if anyWorkload == nil {
			return api.Response(http.StatusNotFound, "workload not found"), nil
		}

		// The workload exists in workloads-manager but is not a Knative Service.
		if !validateWorkloadGVK(anyWorkload, constants.LwsGroup, constants.LwsVersion, constants.LwsKind) {
			return api.Response(http.StatusConflict, fmt.Sprintf("workload %s cannot be updated: it is not a LeaderWorkerSet service", workloadId)), nil
		}

		clusterId = anyWorkload.Workload.Metadata.ClusterId

		var snapErr *api.Error
		snapshot, snapErr = dao_crud.DbGetWorkloadV2Snapshot(ctx, tenantId, DistributedInference, workloadId, permittedScopes)
		if snapErr != nil && snapErr.Code != http.StatusNotFound {
			apiErr := ToSubmissionError(snapErr)
			return api.Response(int(apiErr.Code), apiErr), nil
		}

		if snapshot == nil {
			// The workload exists in workloads-manager, so returning the legacy 404
			// would be misleading. Updating requires the persisted full spec because
			// this endpoint accepts only a partial spec.
			err := workloadsApi.NewSubmissionErrorV2(
				http.StatusServiceUnavailable,
				fmt.Sprintf("cannot update inference workload %s: V2 workload snapshot is missing", workloadId),
			)
			return api.Response(int(err.Code), err), nil
		}
	}

	if wc.clusterCapabilities == nil {
		wc.clusterCapabilities = checker.NewClusterCapabilitiesChecker(clusterId)
	}

	// Check workloads-manager as the source of truth for V2 workload ownership.
	// CanUpdate can derive the cluster from workloads-manager when the best-effort
	// snapshot is missing.
	checkErr := submitter.CanUpdate(ctx, origSubmissionMethod, wc, workloadId)
	if checkErr != nil {
		return api.Response(int(checkErr.Code), checkErr), nil
	}

	// Validations that apply to both paths
	if err := services_utils.ValidateClusterSupportsCapability(ctx, clusterId, checker.MinVersionUpdateDistributedInferenceSpec); err != nil {
		return api.Response(http.StatusBadRequest, err.Message), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Updating distributed inference spec for workload id <%s>", workloadId)

	// For native path (workload from DB)
	if nativeWorkload != nil {
		opErr := services_utils.CheckWorkloadOwnershipProtection(ctx, nativeWorkload.ClusterId, nativeWorkload.DepartmentId, nativeWorkload.CreatedBy)
		if opErr != nil {
			apiErr := ToSubmissionError(opErr)
			return api.Response(int(apiErr.Code), apiErr), nil
		}

		// copy the fields from the update spec into the full spec
		copyErr := copyDistributedInferenceUpdateSpecIntoFullSpec(ctx, nativeWorkload.Spec, &request.Spec)
		if copyErr != nil {
			apiErr := NewSubmissionErrorWithDetails(copyErr.Code, copyErr.Message, copyErr.Details)
			return api.Response(int(apiErr.Code), apiErr), nil
		}

		// set updateBy and updateAt for the workload
		updatedBy := runaictx.GetString(ctx, runaictx.CfSubjectID)
		if updatedBy != nil {
			nativeWorkload.UpdatedBy = *updatedBy
			nativeWorkload.UpdatedAt = time.Now()
		}

		options, checkerErr := determineComposerOptions(ctx, wc.clusterCapabilities)
		if checkerErr != nil {
			return api.Response(int(checkerErr.Code), checkerErr), nil
		}

		// Native compose and persist
		requirements, nativeErr := composeDistributedInferenceUpdateNative(ctx, nativeWorkload, options)
		if nativeErr != nil {
			return api.Response(int(nativeErr.Code), nativeErr), nil
		}

		response, persistErr := persistUpdateAndRespondNative(ctx, tenantSettingsClient, nativeWorkload, workloadId, &requirements)
		if persistErr != nil {
			return api.Response(int(persistErr.Code), persistErr), nil
		}
		acceptedAt := time.Now().UTC()
		sanitizeWorkloadResponse(ctx, response)
		recordWorkloadResponseAction(ctx, *response, api.WORKLOADACTIONTYPE_UPDATE, acceptedAt)

		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully updated distributed inference spec for workload id: <%s> name: <%s> in project id: <%s>", workloadId, nativeWorkload.Name, nativeWorkload.ProjectId)
		return api.Response(http.StatusAccepted, *response), nil
	}

	//
	// V2 / any-workload path: read snapshot, merge the partial update, recompose, send, and write back.
	//
	mergedWorkload := snapshot.ToWorkload()

	if opErr := services_utils.CheckWorkloadOwnershipProtection(ctx, mergedWorkload.ClusterId, mergedWorkload.DepartmentId, mergedWorkload.CreatedBy); opErr != nil {
		apiErr := ToSubmissionError(opErr)
		return api.Response(int(apiErr.Code), apiErr), nil
	}
	if copyErr := copyDistributedInferenceUpdateSpecIntoFullSpec(ctx, mergedWorkload.Spec, &request.Spec); copyErr != nil {
		apiErr := NewSubmissionErrorWithDetails(copyErr.Code, copyErr.Message, copyErr.Details)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(mergedWorkload.ClusterId)
	options, optionsErr := determineComposerOptions(ctx, clusterCapabilityChecker)
	if optionsErr != nil {
		return api.Response(int(optionsErr.Code), optionsErr), nil
	}
	wc.options = options
	wc.clusterCapabilities = clusterCapabilityChecker

	update, submitterErr := submitter.Update(ctx, wc, workloadId, anyWorkload, mergedWorkload)
	if submitterErr != nil {
		return api.Response(int(submitterErr.Code), submitterErr), nil
	}
	if update == nil {
		return api.Response(http.StatusInternalServerError, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, "distributed inference update produced no result")), nil
	}

	updatedWorkload, updateErr := sendUpdateToWorkloadMgr(ctx, workloadId, update)
	if updateErr != nil {
		return api.Response(int(updateErr.Code), updateErr), nil
	}
	acceptedAt := time.Now().UTC()

	// Write the merged spec back so subsequent updates / make-a-copy see it (best-effort):
	// the workload update already succeeded in workloads-manager, so a snapshot failure must
	// not turn a successful update into an error response - log it and preserve the success.
	snapshot.Spec = mergedWorkload.Spec
	response := workloadFromV2Snapshot(snapshot, workloadsApi.WorkloadDesiredPhase(updatedWorkload.DesiredPhase))
	if updatedBy := runaictx.GetString(ctx, runaictx.CfSubjectID); updatedBy != nil {
		snapshot.UpdatedBy = *updatedBy
		snapshot.UpdatedAt = time.Now()
	}
	if writeBackErr := writeBackDistributedInferenceSnapshot(ctx, snapshot, dao_crud.DbUpsertWorkloadV2Snapshot); writeBackErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to write back V2 snapshot for workload %s: %v", workloadId, writeBackErr.Message)
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully updated distributed inference spec via workload-runner for workload id: <%s>", workloadId)
	sanitizeWorkloadResponse(ctx, &response)
	recordWorkloadResponseAction(ctx, response, api.WORKLOADACTIONTYPE_UPDATE, acceptedAt)
	return api.Response(http.StatusAccepted, response), nil
}

func writeBackDistributedInferenceSnapshot(
	ctx context.Context,
	snapshot *models_crud.WorkloadV2Snapshot,
	upsert func(context.Context, *models_crud.WorkloadV2Snapshot) *api.Error,
) *workloadsApi.SubmissionErrorV2 {
	if err := upsert(ctx, snapshot); err != nil {
		return ToSubmissionError(err)
	}
	return nil
}

// composeDistributedInferenceUpdateNative runs the native composer for distributed inference updates.
func composeDistributedInferenceUpdateNative(
	ctx context.Context,
	workload *models_crud.Workload,
	options composer.FieldsComposerOptions,
) (composer.WorkloadRequirements, *workloadsApi.SubmissionErrorV2) {
	if err := validateNativeGpuMode(workload.Spec); err != nil {
		return composer.WorkloadRequirements{}, err
	}

	dInfComposer := runai_workloads.NewDinfWorkloadComposer(
		workload.Name,
		workload.CreatedBy,
		nil,
		options)

	composerErr := dInfComposer.Compose(ctx, gpucompat.LegacyWorkloadSpec(workload.Spec), &dinfwv1.DistributedInferenceWorkload{})
	if composerErr != nil {
		return composer.WorkloadRequirements{}, NewSubmissionErrorWithDetails(composerErr.Code, composerErr.Message, composerErr.Details)
	}

	return dInfComposer.Requirements(), nil
}

// copyDistributedInferenceUpdateSpecIntoFullSpec copies the fields from the update spec into the full spec.
// For distributed inference updates, we only allow updating the replicas field
func copyDistributedInferenceUpdateSpecIntoFullSpec(ctx context.Context, fullSpec *workloadsApi.SupersetSpec, updateSpec *api.UpdateSpecSpec) *workloadsApi.Error {
	// convert the given types to map[string]interface{} in order to merge them using conjungo
	updateSpecAsMap, err := gojsonq.LoadJson(*updateSpec)
	if err != nil {
		return workloadsApi.NewError(http.StatusBadRequest, err.Error())
	}

	fullSpecAsMap, err := gojsonq.LoadJson(*fullSpec)
	if err != nil {
		return workloadsApi.NewError(http.StatusBadRequest, err.Error())
	}

	logDistributedInferenceSpecAsJson(ctx, fullSpecAsMap, "base spec")

	// set merge options. as we only want to merge the root level, we use a custom merge function only for the Map kind
	mergeOptions := conjungo.NewOptions()
	mergeOptions.SetKindMergeFunc(reflect.Map, specMergeFunc)

	err = conjungo.Merge(&fullSpecAsMap, updateSpecAsMap, mergeOptions)
	if err != nil {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to merge update spec into full spec: %v", err))
	}

	// convert the result full spec back to its original type
	jsonBytes, err := json.Marshal(fullSpecAsMap)
	if err != nil {
		return workloadsApi.NewError(http.StatusBadRequest, err.Error())
	}

	*fullSpec = workloadsApi.SupersetSpec{} // the original fullSpec object must be reset before unmarshalling the new data into it
	err = json.Unmarshal(jsonBytes, &fullSpec)
	if err != nil {
		return workloadsApi.NewError(http.StatusBadRequest, err.Error())
	}

	logDistributedInferenceSpecAsJson(ctx, updateSpecAsMap, "update spec")
	logDistributedInferenceSpecAsJson(ctx, fullSpecAsMap, "result spec")

	return nil
}

func logDistributedInferenceSpecAsJson(ctx context.Context, spec map[string]interface{}, message string) {
	fields := make([]string, 0, len(spec))
	for field := range spec {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Update distributed inference spec - %s fields: %v", message, fields)
}
