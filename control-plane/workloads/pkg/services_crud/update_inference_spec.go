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
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/workloadutils"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	clients "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/backend"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/conjungo"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	workloadMgrApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

func UpdateInferenceSpec(ctx context.Context, workloadId string, tenantSettingsClient clients.TenantSettingsClient, request api.InferenceUpdateRequest) (api.ImplResponse, error) {
	updateSpecAsMap, _ := gojsonq.LoadJson(request.Spec)
	sensitiveRequestValues := sensitiveEnvironmentVariableValuesFromMap(updateSpecAsMap)

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
		workloadType: Inference,
		tenantId:     int32(tenantId),
	}

	// Get the submitter for this workload type
	submitter := getAnyWorkloadSubmitter(wc)
	if submitter == nil {
		return api.Response(http.StatusInternalServerError, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, "failed to get submitter for workload type inference")), nil
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
	nativeWorkload, dbErr := dao_crud.DbGetWorkload(ctx, tenantId, Inference, workloadId, permittedScopes)
	if dbErr != nil && dbErr.Code != http.StatusNotFound {
		err := ToSubmissionError(dbErr)
		return api.Response(int(err.Code), err), nil
	}

	if nativeWorkload != nil {
		// we found the record in the crud table, so we know it has been submitted in native path
		origSubmissionMethod = SubmissionMethodNative
		clusterId = nativeWorkload.ClusterId
	} else {
		// the workload was not found in the crud table, so we suspect it has been submitted in any-workload path.
		// but this needs to be verified
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
		if !validateWorkloadGVK(anyWorkload, knativeGroup, knativeVersion, knativeKind) {
			return api.Response(http.StatusConflict, fmt.Sprintf("workload %s cannot be updated: it is not a Knative service", workloadId)), nil
		}

		clusterId = anyWorkload.Workload.Metadata.ClusterId

		var snapErr *api.Error
		snapshot, snapErr = dao_crud.DbGetWorkloadV2Snapshot(ctx, tenantId, Inference, workloadId, permittedScopes)
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

	err := services_utils.CheckKnativeAvailable(ctx, clusterId, services_utils.KnativeFeatures{
		ScaleToZero: isScaleToZeroWorkload(request.Spec.Autoscaling),
		NodeType:    isNodeTypeConfigured(request.Spec.NodeType), //nolint:staticcheck // NodeType is deprecated but still supported for backward compatibility
	})
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Updating inference spec for workload id <%s>", workloadId)

	// For native path (workload from DB)
	if nativeWorkload != nil {
		opErr := services_utils.CheckWorkloadOwnershipProtection(ctx, nativeWorkload.ClusterId, nativeWorkload.DepartmentId, nativeWorkload.CreatedBy)
		if opErr != nil {
			apiErr := ToSubmissionError(opErr)
			return api.Response(int(apiErr.Code), apiErr), nil
		}

		// copy the fields from the update spec into the full spec
		copyErr := copyUpdateSpecIntoFullSpec(ctx, nativeWorkload.Spec, &request.Spec)
		if copyErr != nil {
			apiErr := NewSubmissionErrorWithDetails(copyErr.Code, copyErr.Message, copyErr.Details)
			apiErr = redactSensitiveValuesFromResponse(apiErr, sensitiveRequestValues)
			return api.Response(int(apiErr.Code), apiErr), nil
		}

		// validate policy
		_, policyErr := validatePolicy(ctx, nativeWorkload)
		if policyErr != nil {
			policyErr = redactSensitiveValuesFromResponse(policyErr, sensitiveRequestValues)
			return api.Response(int(policyErr.Code), policyErr), nil
		}

		// set updateBy and updateAt for the workload
		updatedBy := runaictx.GetString(ctx, runaictx.CfSubjectID)
		if updatedBy != nil {
			nativeWorkload.UpdatedBy = *updatedBy
			nativeWorkload.UpdatedAt = time.Now()
		}

		// use workload composer to validate the result spec and notify workload-service
		scErr := workloadutils.ValidateAndHandelServingConfigurationFields(ctx, nativeWorkload.Spec, nativeWorkload.ClusterId)
		if scErr != nil {
			apiErr := NewSubmissionErrorWithDetails(http.StatusBadRequest, scErr.Error(), nil)
			return api.Response(int(apiErr.Code), apiErr), nil
		}

		clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(nativeWorkload.ClusterId)

		options, checkerErr := determineComposerOptions(ctx, clusterCapabilityChecker)
		if checkerErr != nil {
			return api.Response(int(checkerErr.Code), checkerErr), nil
		}

		// Native compose and persist
		requirements, nativeErr := composeInferenceUpdateNative(ctx, nativeWorkload, options)
		if nativeErr != nil {
			nativeErr = redactSensitiveValuesFromResponse(nativeErr, sensitiveRequestValues)
			return api.Response(int(nativeErr.Code), nativeErr), nil
		}

		response, persistErr := persistUpdateAndRespondNative(ctx, tenantSettingsClient, nativeWorkload, workloadId, &requirements)
		if persistErr != nil {
			persistErr = redactSensitiveValuesFromResponse(persistErr, sensitiveRequestValues)
			return api.Response(int(persistErr.Code), persistErr), nil
		}
		acceptedAt := time.Now().UTC()
		sanitizeWorkloadResponse(ctx, response)
		recordWorkloadResponseAction(ctx, *response, api.WORKLOADACTIONTYPE_UPDATE, acceptedAt)

		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully updated inference spec for workload id: <%s> name: <%s> in project id: <%s>", workloadId, nativeWorkload.Name, nativeWorkload.ProjectId)
		return api.Response(http.StatusAccepted, *response), nil
	}

	// V2 / any-workload path: read snapshot, merge partial update, recompose, send, write snapshot back.
	mergedWorkload := snapshot.ToWorkload()

	// Ownership / policy / serving-config validation — parity with the native update path.
	if opErr := services_utils.CheckWorkloadOwnershipProtection(ctx, mergedWorkload.ClusterId, mergedWorkload.DepartmentId, mergedWorkload.CreatedBy); opErr != nil {
		apiErr := ToSubmissionError(opErr)
		return api.Response(int(apiErr.Code), apiErr), nil
	}
	if copyErr := mergeUpdateSpecIntoFullSpec(ctx, mergedWorkload.Spec, &request.Spec); copyErr != nil {
		apiErr := NewSubmissionErrorWithDetails(copyErr.Code, copyErr.Message, copyErr.Details)
		return api.Response(int(apiErr.Code), apiErr), nil
	}
	if _, policyErr := validatePolicy(ctx, mergedWorkload); policyErr != nil {
		policyErr = redactSensitiveValuesFromResponse(policyErr, sensitiveRequestValues)
		return api.Response(int(policyErr.Code), policyErr), nil
	}
	// Mutates mergedWorkload.Spec, so it must run before the submitter recomposes.
	if scErr := workloadutils.ValidateAndHandelServingConfigurationFields(ctx, mergedWorkload.Spec, mergedWorkload.ClusterId); scErr != nil {
		apiErr := NewSubmissionErrorWithDetails(http.StatusBadRequest, scErr.Error(), nil)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	dataVolumeClaimNames, dataVolumeErr := preValidateDataVolumes(
		ctx,
		mergedWorkload.ProjectId,
		mergedWorkload.Spec,
		nil,
	)
	if dataVolumeErr != nil {
		return api.Response(int(dataVolumeErr.Code), dataVolumeErr), nil
	}
	wc.dataVolumeClaimNames = dataVolumeClaimNames

	options, optErr := determineComposerOptions(ctx, wc.clusterCapabilities)
	if optErr != nil {
		return api.Response(int(optErr.Code), optErr), nil
	}
	wc.options = options

	update, submitterErr := submitter.Update(ctx, wc, workloadId, anyWorkload, mergedWorkload)
	if submitterErr != nil {
		submitterErr = redactSensitiveValuesFromResponse(submitterErr, sensitiveRequestValues)
		return api.Response(int(submitterErr.Code), submitterErr), nil
	}
	if update == nil {
		// Defensive: the inference submitter never returns nil,nil; guard against a nil deref.
		return api.Response(http.StatusInternalServerError, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, "inference update produced no result")), nil
	}

	updatedWorkload, updateErr := sendUpdateToWorkloadMgr(ctx, workloadId, update)
	if updateErr != nil {
		updateErr = redactSensitiveValuesFromResponse(updateErr, sensitiveRequestValues)
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
	if wbErr := dao_crud.DbUpsertWorkloadV2Snapshot(ctx, snapshot); wbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to write back V2 snapshot for workload %s: %v", workloadId, wbErr.Message)
	}
	sanitizeWorkloadResponse(ctx, &response)
	recordWorkloadResponseAction(ctx, response, api.WORKLOADACTIONTYPE_UPDATE, acceptedAt)

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully updated inference spec via workload-runner for workload id: <%s>", workloadId)
	return api.Response(http.StatusAccepted, response), nil
}

// composeInferenceUpdateNative runs the native composer for inference updates.
func composeInferenceUpdateNative(
	ctx context.Context,
	workload *models_crud.Workload,
	options composer.FieldsComposerOptions,
) (composer.WorkloadRequirements, *workloadsApi.SubmissionErrorV2) {
	if err := validateNativeGpuMode(workload.Spec); err != nil {
		return composer.WorkloadRequirements{}, err
	}

	jobSpecComposer := runai_workloads.NewSupersetToJobSpecComposer(
		Inference,
		workload.Name,
		workload.CreatedBy,
		nil,
		nil,
		options)

	composerErr := jobSpecComposer.Compose(ctx, gpucompat.LegacyWorkloadSpec(workload.Spec), &jobs.JobSpec{})()
	if composerErr != nil {
		return composer.WorkloadRequirements{}, NewSubmissionErrorWithDetails(composerErr.Code, composerErr.Message, composerErr.Details)
	}

	return jobSpecComposer.WorkloadRequirements, nil
}

func specMergeFunc(t, s reflect.Value, o *conjungo.Options) (reflect.Value, error) {
	// iterate the map keys (root level) and overwrite each key in the target with the source one,
	// if it is not nil. there's no recursion as we just want to handle the root level in a simple manner, simple merge
	for _, k := range s.MapKeys() {
		if !s.MapIndex(k).IsNil() {
			t.SetMapIndex(k, s.MapIndex(k))
		}
	}

	return t, nil
}

// copyUpdateSpecIntoFullSpec copies the fields from the inference update spec into the full spec.
func copyUpdateSpecIntoFullSpec(ctx context.Context, fullSpec *workloadsApi.SupersetSpec, updateSpec *api.InferenceUpdateSpecSpec) *workloadsApi.Error {
	return mergeUpdateSpecIntoFullSpec(ctx, fullSpec, updateSpec)
}

// mergeUpdateSpecIntoFullSpec merges all non-nil root-level fields from an (untyped) partial update spec into the full SupersetSpec.
func mergeUpdateSpecIntoFullSpec(ctx context.Context, fullSpec *workloadsApi.SupersetSpec, updateSpec any) *workloadsApi.Error {
	// convert the given types to map[string]interface{} in order to merge them using conjungo
	updateSpecAsMap, err := gojsonq.LoadJson(updateSpec)
	if err != nil {
		return workloadsApi.NewError(http.StatusBadRequest, err.Error())
	}

	fullSpecAsMap, err := gojsonq.LoadJson(*fullSpec)
	if err != nil {
		return workloadsApi.NewError(http.StatusBadRequest, err.Error())
	}

	if err := gpucompat.ValidateMemoryRequestInput(updateSpecAsMap); err != nil {
		return workloadsApi.NewError(http.StatusBadRequest, err.Error())
	}
	updateSpecAsMap, err = gpucompat.NormalizeLegacyGPUFields(updateSpecAsMap)
	if err != nil {
		return workloadsApi.NewError(http.StatusBadRequest, err.Error())
	}
	fullSpecAsMap, err = gpucompat.NormalizeLegacyGPUFields(fullSpecAsMap)
	if err != nil {
		return workloadsApi.NewError(http.StatusBadRequest, err.Error())
	}

	if err := restoreRedactedSensitiveEnvironmentVariableValues(updateSpecAsMap, fullSpecAsMap); err != nil {
		return workloadsApi.NewError(http.StatusBadRequest, err.Error())
	}

	logSpecAsJson(ctx, fullSpecAsMap, "base spec")

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

	logSpecAsJson(ctx, updateSpecAsMap, "update spec")
	logSpecAsJson(ctx, fullSpecAsMap, "result spec")

	return nil
}

func logSpecAsJson(ctx context.Context, spec map[string]interface{}, message string) {
	fields := make([]string, 0, len(spec))
	for field := range spec {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Update inference spec - %s fields: %v", message, fields)
}

func restoreRedactedSensitiveEnvironmentVariableValues(updateSpec map[string]interface{}, storedSpec map[string]interface{}) error {
	return restoreRedactedSensitiveEnvironmentVariableValue(updateSpec, storedSpec)
}

func restoreRedactedSensitiveEnvironmentVariableValue(update interface{}, stored interface{}) error {
	switch typedUpdate := update.(type) {
	case map[string]interface{}:
		typedStored, _ := stored.(map[string]interface{})
		for key, child := range typedUpdate {
			var storedChild interface{}
			if typedStored != nil {
				storedChild = typedStored[key]
			}
			if key == "environmentVariables" {
				updateVariables, _ := child.([]interface{})
				storedVariables, _ := storedChild.([]interface{})
				if err := restoreRedactedSensitiveEnvironmentVariableArray(updateVariables, storedVariables); err != nil {
					return err
				}
				continue
			}
			if err := restoreRedactedSensitiveEnvironmentVariableValue(child, storedChild); err != nil {
				return err
			}
		}
	case []interface{}:
		storedSlice, _ := stored.([]interface{})
		for index, child := range typedUpdate {
			var storedChild interface{}
			if index < len(storedSlice) {
				storedChild = storedSlice[index]
			}
			if err := restoreRedactedSensitiveEnvironmentVariableValue(child, storedChild); err != nil {
				return err
			}
		}
	}
	return nil
}

func restoreRedactedSensitiveEnvironmentVariableArray(updateVariables []interface{}, storedVariables []interface{}) error {
	for _, candidate := range updateVariables {
		environmentVariable, ok := candidate.(map[string]interface{})
		if !ok || !isNameOnlySensitiveEnvironmentVariable(environmentVariable) {
			continue
		}

		name := environmentVariable["name"].(string)
		value, found := storedSensitiveEnvironmentVariableValue(name, storedVariables)
		if !found {
			return fmt.Errorf("%s requires a value or credential reference", name)
		}
		environmentVariable["value"] = value
	}
	return nil
}

func isNameOnlySensitiveEnvironmentVariable(environmentVariable map[string]interface{}) bool {
	name, ok := environmentVariable["name"].(string)
	if !ok {
		return false
	}
	if !isSensitivePlaintextEnvironmentVariable(name) {
		return false
	}

	for _, field := range []string{"value", "secret", "userCredential", "configMap", "podFieldRef"} {
		if _, exists := environmentVariable[field]; exists {
			return false
		}
	}
	if exclude, exists := environmentVariable["exclude"].(bool); exists && exclude {
		return false
	}
	return true
}

func storedSensitiveEnvironmentVariableValue(name string, storedVariables []interface{}) (string, bool) {
	for _, candidate := range storedVariables {
		environmentVariable, ok := candidate.(map[string]interface{})
		if !ok || environmentVariable["name"] != name {
			continue
		}
		value, found := environmentVariable["value"].(string)
		return value, found
	}
	return "", false
}
