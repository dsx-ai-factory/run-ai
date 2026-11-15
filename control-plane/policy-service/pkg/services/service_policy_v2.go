/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dao"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/policy_compliance"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/tree"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
)

type PolicyApiService struct{}

func NewPolicyApiService() api.PolicyAPIServicer {
	return &PolicyApiService{}
}

func toPolicyMeta(dbMeta *dbmodel.Policy) api.PolicyMeta {
	return api.PolicyMeta{
		Scope:        dbMeta.Scope,
		ProjectId:    dbMeta.ProjectId,
		DepartmentId: dbMeta.DepartmentId,
		ClusterId:    dbMeta.ClusterId,
		Name:         dbMeta.Name,
		Id:           dbMeta.ID.String(),
		CreatedBy:    dbMeta.CreatedBy,
		CreatedAt:    dbMeta.CreatedAt,
		UpdatedBy:    dbMeta.UpdatedBy,
		UpdatedAt:    dbMeta.UpdatedAt,
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//           Workspace
////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (pa *PolicyApiService) ApplyWorkspacePolicyV2(ctx context.Context, meta api.PolicyCreationFields, policy *api.WorkspacePolicyDefaultsAndRulesV2,
	resetFields []string, options tree.ApplyPolicyOptions,
) (api.ImplResponse, error) {
	result, err := pa.applyWorkloadPolicyV2(ctx, meta, policy, resetFields, options, tree.WorkloadTypeWorkspace)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	if options.ValidateOnly {
		return api.Response(http.StatusNoContent, nil), nil
	}

	return toWorkspaceResponseV2(result), nil
}

func (pa *PolicyApiService) UpdateWorkspacePolicyV2(ctx context.Context, validateOnly bool, approveClusterDeletion bool, request api.WorkspacePolicyChangeRequestV2) (api.ImplResponse, error) {
	return pa.ApplyWorkspacePolicyV2(ctx, request.Meta, request.Policy, request.Reset, tree.ApplyPolicyOptions{
		ValidateOnly: validateOnly,
		Overwrite:    false,

		ApproveClusterDeletion: approveClusterDeletion,
	})
}

func (pa *PolicyApiService) OverwriteWorkspacePolicyV2(ctx context.Context, validateOnly bool, approveClusterDeletion bool, request api.WorkspacePolicyOverwriteRequestV2) (api.ImplResponse, error) {
	return pa.ApplyWorkspacePolicyV2(ctx, request.Meta, request.Policy, []string{}, tree.ApplyPolicyOptions{
		ValidateOnly: validateOnly,
		Overwrite:    true,

		ApproveClusterDeletion: approveClusterDeletion,
	})
}

func (pa *PolicyApiService) GetWorkspacePolicyV2(ctx context.Context, scope string, departmentId string, projectId string, clusterId string, noCreationMode bool) (api.ImplResponse, error) {
	projectIdNumeric, err := projectIdToNum(projectId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	getPolicyRequest := tree.NewGenericPolicyRequest(tree.WorkloadTypeWorkspace, scope, departmentId, int32(projectIdNumeric), clusterId, noCreationMode, noNeedClusterApproval)

	result, getError := getPolicy(ctx, getPolicyRequest, getOptions{})
	if getError != nil {
		return NewErrorResponse(getError), nil
	}

	return toWorkspaceResponseV2(result), nil
}

func (pa *PolicyApiService) DeleteWorkspacePolicy(ctx context.Context, scope string, departmentId string, projectId string, clusterId string, approveClusterDeletion bool) (api.ImplResponse, error) {
	projectIdNumeric, errConv := projectIdToNum(projectId)
	if errConv != nil {
		return NewErrorResponse(errConv), nil
	}

	deletePolicyRequest := tree.NewGenericPolicyRequest(tree.WorkloadTypeWorkspace, scope, departmentId, int32(projectIdNumeric), clusterId, false, approveClusterDeletion)

	return deletePolicy(ctx, deletePolicyRequest), nil
}

func toWorkspaceV2(record *dbmodel.PolicyWithOptionalEffective) api.WorkspacePolicyV2 {
	response := api.WorkspacePolicyV2{
		Meta:               toPolicyMeta(&record.Policy),
		Policy:             &api.WorkspacePolicyDefaultsAndRulesV2{},
		Effective:          &api.WorkspacePolicyDefaultsAndRulesV2{},
		EffectiveUpdatedAt: record.UpdatedAt,
	}

	jsonBMapToResponse(&response.Policy.Defaults, record.Policy.DefaultsV2)
	jsonBMapToResponse(&response.Policy.Rules, record.Policy.RulesV2)

	jsonBMapToResponse(&response.Effective.Defaults, record.EffectivePolicy.DefaultsV2)
	jsonBMapToResponse(&response.Effective.Rules, record.EffectivePolicy.RulesV2)

	if record.Policy.ImposedAssets != nil && record.Policy.ImposedAssets.Worker != nil && len(*record.Policy.ImposedAssets.Worker) != 0 {
		response.Policy.ImposedAssets = record.Policy.ImposedAssets.Worker
		response.Effective.ImposedAssets = record.Policy.ImposedAssets.Worker
	}

	response.Effective.DefaultFromOrder = toResponse(record.DefaultFromOrder)
	response.Effective.Status = toResponseStatus(record.ValidationError)

	return response
}

func toWorkspaceResponseV2(record *dbmodel.PolicyWithOptionalEffective) api.ImplResponse {
	return api.Response(http.StatusOK, toWorkspaceV2(record))
}

func toResponse(src *api.DefaultFromOrder) *api.DefaultFromOrder {
	if src == nil {
		return nil
	}

	var result api.DefaultFromOrder

	// we want to put in the response only maps that have at least one itemized rule
	// or at least one rule for the flat rules
	// Keep dependency paths consistent with legacy defaultFrom references in responses.

	if src.Worker != nil {
		if len(src.Worker) > 1 || len(src.Worker[""]) > 0 {
			result.Worker = gpucompat.LegacyGPUDefaultFromOrder(src.Worker)
		}
	}

	if src.Master != nil {
		if len(src.Master) > 1 || len(src.Master[""]) > 0 {
			result.Master = gpucompat.LegacyGPUDefaultFromOrder(src.Master)
		}
	}

	if result.Worker == nil && result.Master == nil {
		return nil
	}

	return &result
}

func toResponseStatus(validationError *string) *api.PolicyValidationStatus {
	if validationError == nil {
		return nil
	}

	return &api.PolicyValidationStatus{
		Validation: &api.PolicyValidationStatusValidation{
			ErrorMessage: *validationError,
		},
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//           Training
////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (pa *PolicyApiService) ApplyTrainingPolicyV2(ctx context.Context, meta api.PolicyCreationFields, policy *api.TrainingPolicyDefaultsAndRulesV2,
	resetFields []string, options tree.ApplyPolicyOptions,
) (api.ImplResponse, error) {
	result, err := pa.applyWorkloadPolicyV2(ctx, meta, policy, resetFields, options, tree.WorkloadTypeTraining)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	if options.ValidateOnly {
		return api.Response(http.StatusNoContent, nil), nil
	}

	return toTrainingResponseV2(result), nil
}

func (pa *PolicyApiService) UpdateTrainingPolicyV2(ctx context.Context, validateOnly bool, approveClusterDeletion bool, request api.TrainingPolicyChangeRequestV2) (api.ImplResponse, error) {
	return pa.ApplyTrainingPolicyV2(ctx, request.Meta, request.Policy, request.Reset, tree.ApplyPolicyOptions{
		ValidateOnly: validateOnly,
		Overwrite:    false,

		ApproveClusterDeletion: approveClusterDeletion,
	})
}

func (pa *PolicyApiService) OverwriteTrainingPolicyV2(ctx context.Context, validateOnly bool, approveClusterDeletion bool, request api.TrainingPolicyOverwriteRequestV2) (api.ImplResponse, error) {
	return pa.ApplyTrainingPolicyV2(ctx, request.Meta, request.Policy, []string{}, tree.ApplyPolicyOptions{
		ValidateOnly: validateOnly,
		Overwrite:    true,

		ApproveClusterDeletion: approveClusterDeletion,
	})
}

func (pa *PolicyApiService) GetTrainingPolicyV2(ctx context.Context, scope string, departmentId string, projectId string, clusterId string, noCreationMode bool) (api.ImplResponse, error) {
	projectIdNumeric, err := projectIdToNum(projectId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	getPolicyRequest := tree.NewGenericPolicyRequest(tree.WorkloadTypeTraining, scope, departmentId, int32(projectIdNumeric), clusterId, noCreationMode, noNeedClusterApproval)

	result, getError := getPolicy(ctx, getPolicyRequest, getOptions{})
	if getError != nil {
		return NewErrorResponse(getError), nil
	}

	return toTrainingResponseV2(result), nil
}

func (pa *PolicyApiService) DeleteTrainingPolicy(ctx context.Context, scope string, departmentId string, projectId string, clusterId string, approveClusterDeletion bool) (api.ImplResponse, error) {
	projectIdNumeric, errConv := projectIdToNum(projectId)
	if errConv != nil {
		return NewErrorResponse(errConv), nil
	}
	deletePolicyRequest := tree.NewGenericPolicyRequest(tree.WorkloadTypeTraining, scope, departmentId, int32(projectIdNumeric), clusterId, false, approveClusterDeletion)

	return deletePolicy(ctx, deletePolicyRequest), nil
}

func toTrainingV2(record *dbmodel.PolicyWithOptionalEffective) api.TrainingPolicyV2 {
	response := api.TrainingPolicyV2{
		Meta:               toPolicyMeta(&record.Policy),
		Policy:             &api.TrainingPolicyDefaultsAndRulesV2{},
		Effective:          &api.TrainingPolicyDefaultsAndRulesV2{},
		EffectiveUpdatedAt: record.UpdatedAt,
	}

	jsonBMapToResponse(&response.Policy.Defaults, record.Policy.DefaultsV2)
	jsonBMapToResponse(&response.Policy.Rules, record.Policy.RulesV2)

	jsonBMapToResponse(&response.Effective.Defaults, record.EffectivePolicy.DefaultsV2)
	jsonBMapToResponse(&response.Effective.Rules, record.EffectivePolicy.RulesV2)

	if record.Policy.ImposedAssets != nil && record.Policy.ImposedAssets.Worker != nil && len(*record.Policy.ImposedAssets.Worker) != 0 {
		response.Policy.ImposedAssets = record.Policy.ImposedAssets.Worker
		response.Effective.ImposedAssets = record.Policy.ImposedAssets.Worker
	}

	response.Effective.DefaultFromOrder = toResponse(record.DefaultFromOrder)
	response.Effective.Status = toResponseStatus(record.ValidationError)

	return response
}

func toTrainingResponseV2(record *dbmodel.PolicyWithOptionalEffective) api.ImplResponse {
	return api.Response(http.StatusOK, toTrainingV2(record))
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//           Inference
////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (pa *PolicyApiService) ApplyInferencePolicyV2(ctx context.Context, meta api.PolicyCreationFields, policy *api.InferencePolicyDefaultsAndRulesV2,
	resetFields []string, options tree.ApplyPolicyOptions,
) (api.ImplResponse, error) {
	result, err := pa.applyWorkloadPolicyV2(ctx, meta, policy, resetFields, options, tree.WorkloadTypeInference)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	if options.ValidateOnly {
		return api.Response(http.StatusNoContent, nil), nil
	}

	return toInferenceResponseV2(result), nil
}

func (pa *PolicyApiService) UpdateInferencePolicyV2(ctx context.Context, validateOnly bool, approveClusterDeletion bool, request api.InferencePolicyChangeRequestV2) (api.ImplResponse, error) {
	return pa.ApplyInferencePolicyV2(ctx, request.Meta, request.Policy, request.Reset, tree.ApplyPolicyOptions{
		ValidateOnly: validateOnly,
		Overwrite:    false,

		ApproveClusterDeletion: approveClusterDeletion,
	})
}

func (pa *PolicyApiService) OverwriteInferencePolicyV2(ctx context.Context, validateOnly bool, approveClusterDeletion bool, request api.InferencePolicyOverwriteRequestV2) (api.ImplResponse, error) {
	return pa.ApplyInferencePolicyV2(ctx, request.Meta, request.Policy, []string{}, tree.ApplyPolicyOptions{
		ValidateOnly: validateOnly,
		Overwrite:    true,

		ApproveClusterDeletion: approveClusterDeletion,
	})
}

func (pa *PolicyApiService) GetInferencePolicyV2(ctx context.Context, scope string, departmentId string, projectId string, clusterId string, noCreationMode bool) (api.ImplResponse, error) {
	projectIdNumeric, err := projectIdToNum(projectId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	getPolicyRequest := tree.NewGenericPolicyRequest(tree.WorkloadTypeInference, scope, departmentId, int32(projectIdNumeric), clusterId, noCreationMode, noNeedClusterApproval)

	result, getError := getPolicy(ctx, getPolicyRequest, getOptions{})
	if getError != nil {
		return NewErrorResponse(getError), nil
	}

	return toInferenceResponseV2(result), nil
}

func (pa *PolicyApiService) DeleteInferencePolicy(ctx context.Context, scope string, departmentId string, projectId string, clusterId string, approveClusterDeletion bool) (api.ImplResponse, error) {
	projectIdNumeric, errConv := projectIdToNum(projectId)
	if errConv != nil {
		return NewErrorResponse(errConv), nil
	}
	deletePolicyRequest := tree.NewGenericPolicyRequest(tree.WorkloadTypeInference, scope, departmentId, int32(projectIdNumeric), clusterId, false, approveClusterDeletion)

	return deletePolicy(ctx, deletePolicyRequest), nil
}

func toInferenceV2(record *dbmodel.PolicyWithOptionalEffective) api.InferencePolicyV2 {
	response := api.InferencePolicyV2{
		Meta:               toPolicyMeta(&record.Policy),
		Policy:             &api.InferencePolicyDefaultsAndRulesV2{},
		Effective:          &api.InferencePolicyDefaultsAndRulesV2{},
		EffectiveUpdatedAt: record.UpdatedAt,
	}

	jsonBMapToResponse(&response.Policy.Defaults, record.Policy.DefaultsV2)
	jsonBMapToResponse(&response.Policy.Rules, record.Policy.RulesV2)

	jsonBMapToResponse(&response.Effective.Defaults, record.EffectivePolicy.DefaultsV2)
	jsonBMapToResponse(&response.Effective.Rules, record.EffectivePolicy.RulesV2)

	if record.Policy.ImposedAssets != nil && record.Policy.ImposedAssets.Worker != nil && len(*record.Policy.ImposedAssets.Worker) != 0 {
		response.Policy.ImposedAssets = record.Policy.ImposedAssets.Worker
		response.Effective.ImposedAssets = record.Policy.ImposedAssets.Worker
	}

	response.Effective.DefaultFromOrder = toResponse(record.DefaultFromOrder)
	response.Effective.Status = toResponseStatus(record.ValidationError)

	return response
}

func toInferenceResponseV2(record *dbmodel.PolicyWithOptionalEffective) api.ImplResponse {
	return api.Response(http.StatusOK, toInferenceV2(record))
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//           Distributed
////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (pa *PolicyApiService) OverwriteDistributedPolicyV2(ctx context.Context, validateOnly bool, approveClusterDeletion bool, request api.DistributedPolicyOverwriteRequestV2) (api.ImplResponse, error) {
	return pa.ApplyDistributedPolicy(ctx, request.Meta, request.Policy, []string{}, tree.ApplyPolicyOptions{
		ValidateOnly: validateOnly,
		Overwrite:    true,

		ApproveClusterDeletion: approveClusterDeletion,
	})
}

func (pa *PolicyApiService) UpdateDistributedPolicyV2(ctx context.Context, validateOnly bool, approveClusterDeletion bool, request api.DistributedPolicyChangeRequestV2) (api.ImplResponse, error) {
	return pa.ApplyDistributedPolicy(ctx, request.Meta, request.Policy, request.Reset, tree.ApplyPolicyOptions{
		ValidateOnly: validateOnly,
		Overwrite:    false,

		ApproveClusterDeletion: approveClusterDeletion,
	})
}

func (pa *PolicyApiService) ApplyDistributedPolicy(ctx context.Context, meta api.PolicyCreationFields, policy *api.DistributedPolicyDefaultsAndRulesV2,
	resetFields []string, options tree.ApplyPolicyOptions,
) (api.ImplResponse, error) {
	defaults := &api.DistributedPolicyDefaultsV2{}
	rules := &api.DistributedPolicyRulesV2{}

	if policy != nil {
		if policy.Rules != nil {
			rules = policy.Rules
		}
		if policy.Defaults != nil {
			defaults = policy.Defaults
		}
	}

	defaultsMap, err := gojsonq.LoadJson(defaults)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusServiceUnavailable, "failed to parse defaults", err)), nil
	}

	rulesMap, err := gojsonq.LoadJson(rules)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusServiceUnavailable, "failed to parse rules", err)), nil
	}

	request := tree.ApplyPolicyRequest{
		GenericPolicyRequest: tree.GenericPolicyRequest{
			WorkloadType: tree.WorkloadTypeDistributed,
			PolicyMeta:   meta,

			ApproveClusterDeletion: options.ApproveClusterDeletion,
		},
		Defaults:      defaultsMap,
		Rules:         rulesMap,
		ImposedAssets: policy.ImposedAssets,
		FieldsToReset: resetFields,
	}

	result, applyError := applyPolicy(ctx, &request, options)
	if applyError != nil {
		return NewErrorResponse(applyError), nil
	}

	if options.ValidateOnly {
		return api.Response(http.StatusNoContent, nil), nil
	}

	return toDistributedResponse(result), nil
}

func (pa *PolicyApiService) DeleteDistributedPolicy(ctx context.Context, scope string, departmentId string, projectId string, clusterId string, approveClusterDeletion bool) (api.ImplResponse, error) {
	projectIdNumeric, errConv := projectIdToNum(projectId)
	if errConv != nil {
		return NewErrorResponse(errConv), nil
	}

	deletePolicyRequest := tree.NewGenericPolicyRequest(tree.WorkloadTypeDistributed, scope, departmentId, int32(projectIdNumeric), clusterId, false, approveClusterDeletion)

	return deletePolicy(ctx, deletePolicyRequest), nil
}

func (pa *PolicyApiService) GetDistributedPolicyV2(ctx context.Context, scope string, departmentId string, projectId string, clusterId string, noCreationMode bool) (api.ImplResponse, error) {
	projectIdNumeric, err := projectIdToNum(projectId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	getPolicyRequest := tree.NewGenericPolicyRequest(tree.WorkloadTypeDistributed, scope, departmentId, int32(projectIdNumeric), clusterId, noCreationMode, noNeedClusterApproval)

	result, getError := getPolicy(ctx, getPolicyRequest, getOptions{})
	if getError != nil {
		return NewErrorResponse(getError), nil
	}

	return toDistributedResponse(result), nil
}

func toDistributed(record *dbmodel.PolicyWithOptionalEffective) api.DistributedPolicyV2 {
	response := api.DistributedPolicyV2{
		Meta:               toPolicyMeta(&record.Policy),
		Policy:             &api.DistributedPolicyDefaultsAndRulesV2{},
		Effective:          &api.DistributedPolicyDefaultsAndRulesV2{},
		EffectiveUpdatedAt: record.UpdatedAt,
	}

	jsonBMapToResponse(&response.Policy.Defaults, record.Policy.DefaultsV2)
	jsonBMapToResponse(&response.Policy.Rules, record.Policy.RulesV2)

	jsonBMapToResponse(&response.Effective.Defaults, record.EffectivePolicy.DefaultsV2)
	jsonBMapToResponse(&response.Effective.Rules, record.EffectivePolicy.RulesV2)

	if record.Policy.ImposedAssets != nil {
		response.Policy.ImposedAssets = record.Policy.ImposedAssets
		response.Effective.ImposedAssets = record.EffectivePolicy.ImposedAssets
	}

	response.Effective.DefaultFromOrder = toResponse(record.DefaultFromOrder)
	response.Effective.Status = toResponseStatus(record.ValidationError)

	return response
}

func toDistributedResponse(record *dbmodel.PolicyWithOptionalEffective) api.ImplResponse {
	return api.Response(http.StatusOK, toDistributed(record))
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//           Distributed Inference
////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (pa *PolicyApiService) ApplyDistributedInferencePolicyV2(ctx context.Context, meta api.PolicyCreationFields, policy *api.DistributedInferencePolicyDefaultsAndRulesV2,
	resetFields []string, options tree.ApplyPolicyOptions,
) (api.ImplResponse, error) {
	defaults := &api.DistributedInferenceDefaultsV2{}
	rules := &api.DistributedInferenceRulesV2{}
	imposedAssets := &api.DistributedImposedAssets{}

	if policy != nil {
		if policy.Rules != nil {
			rules = policy.Rules
		}
		if policy.Defaults != nil {
			defaults = policy.Defaults
		}
		if policy.ImposedAssets != nil {
			imposedAssets.Worker = policy.ImposedAssets.Worker
			imposedAssets.Master = policy.ImposedAssets.Leader
		}
	}

	defaultsMap, err := gojsonq.LoadJson(defaults)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusServiceUnavailable, "failed to parse defaults", err)), nil
	}

	rulesMap, err := gojsonq.LoadJson(rules)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusServiceUnavailable, "failed to parse rules", err)), nil
	}

	request := tree.ApplyPolicyRequest{
		GenericPolicyRequest: tree.GenericPolicyRequest{
			WorkloadType: tree.WorkloadTypeDistributedInference,
			PolicyMeta:   meta,

			ApproveClusterDeletion: options.ApproveClusterDeletion,
		},
		Defaults:      defaultsMap,
		Rules:         rulesMap,
		ImposedAssets: imposedAssets,
		FieldsToReset: resetFields,
	}

	result, applyError := applyPolicy(ctx, &request, options)
	if applyError != nil {
		return NewErrorResponse(applyError), nil
	}

	if options.ValidateOnly {
		return api.Response(http.StatusNoContent, nil), nil
	}

	return toDistributedInferenceResponseV2(result), nil
}

func (pa *PolicyApiService) UpdateDistributedInferencePolicyV2(ctx context.Context, validateOnly bool, approveClusterDeletions bool, request api.DistributedInferencePolicyChangeRequestV2,
) (api.ImplResponse, error) {
	return pa.ApplyDistributedInferencePolicyV2(ctx, request.Meta, request.Policy, request.Reset, tree.ApplyPolicyOptions{
		ValidateOnly:           validateOnly,
		Overwrite:              false,
		ApproveClusterDeletion: approveClusterDeletions,
	})
}

func (pa *PolicyApiService) OverwriteDistributedInferencePolicyV2(ctx context.Context, validateOnly bool, approveClusterDeletions bool, request api.DistributedInferencePolicyOverwriteRequestV2,
) (api.ImplResponse, error) {
	return pa.ApplyDistributedInferencePolicyV2(ctx, request.Meta, request.Policy, []string{}, tree.ApplyPolicyOptions{
		ValidateOnly:           validateOnly,
		Overwrite:              true,
		ApproveClusterDeletion: approveClusterDeletions,
	})
}

func (pa *PolicyApiService) GetDistributedInferencePolicyV2(ctx context.Context, scope string, departmentId string, projectId string, clusterId string, noCreationMode bool) (api.ImplResponse, error) {
	projectIdNumeric, err := projectIdToNum(projectId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	getPolicyRequest := tree.NewGenericPolicyRequest(tree.WorkloadTypeDistributedInference, scope, departmentId, int32(projectIdNumeric), clusterId, noCreationMode, noNeedClusterApproval)

	result, getError := getPolicy(ctx, getPolicyRequest, getOptions{})
	if getError != nil {
		return NewErrorResponse(getError), nil
	}

	return toDistributedInferenceResponseV2(result), nil
}

func (pa *PolicyApiService) DeleteDistributedInferencePolicy(ctx context.Context, scope string, departmentId string, projectId string, clusterId string, approveClusterDeletion bool) (api.ImplResponse, error) {
	projectIdNumeric, errConv := projectIdToNum(projectId)
	if errConv != nil {
		return NewErrorResponse(errConv), nil
	}

	deletePolicyRequest := tree.NewGenericPolicyRequest(tree.WorkloadTypeDistributedInference, scope, departmentId, projectIdNumeric, clusterId, false, approveClusterDeletion)

	return deletePolicy(ctx, deletePolicyRequest), nil
}

func toDistributedInferenceV2(record *dbmodel.PolicyWithOptionalEffective) api.DistributedInferencePolicyV2 {
	response := api.DistributedInferencePolicyV2{
		Meta:               toPolicyMeta(&record.Policy),
		Policy:             &api.DistributedInferencePolicyDefaultsAndRulesV2{},
		Effective:          &api.DistributedInferencePolicyDefaultsAndRulesV2{},
		EffectiveUpdatedAt: record.UpdatedAt,
	}

	jsonBMapToResponse(&response.Policy.Defaults, record.Policy.DefaultsV2)
	jsonBMapToResponse(&response.Policy.Rules, record.Policy.RulesV2)

	jsonBMapToResponse(&response.Effective.Defaults, record.EffectivePolicy.DefaultsV2)
	jsonBMapToResponse(&response.Effective.Rules, record.EffectivePolicy.RulesV2)

	if record.Policy.ImposedAssets != nil {
		response.Policy.ImposedAssets = &api.DistributedInferenceImposedAssets{}
		response.Effective.ImposedAssets = &api.DistributedInferenceImposedAssets{}
		if record.Policy.ImposedAssets.Worker != nil && len(*record.Policy.ImposedAssets.Worker) != 0 {
			response.Policy.ImposedAssets.Worker = record.Policy.ImposedAssets.Worker
			response.Effective.ImposedAssets.Worker = record.Policy.ImposedAssets.Worker
		}
		if record.Policy.ImposedAssets.Master != nil && len(*record.Policy.ImposedAssets.Master) != 0 {
			response.Policy.ImposedAssets.Leader = record.Policy.ImposedAssets.Master
			response.Effective.ImposedAssets.Leader = record.Policy.ImposedAssets.Master
		}
	}

	response.Effective.DefaultFromOrder = toResponse(record.DefaultFromOrder)
	response.Effective.Status = toResponseStatus(record.ValidationError)

	return response
}

func toDistributedInferenceResponseV2(record *dbmodel.PolicyWithOptionalEffective) api.ImplResponse {
	return api.Response(http.StatusOK, toDistributedInferenceV2(record))
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (pa *PolicyApiService) applyWorkloadPolicyV2(ctx context.Context, meta api.PolicyCreationFields, genericPolicy interface{},
	resetFields []string, options tree.ApplyPolicyOptions, workloadType tree.WorkloadType,
) (*dbmodel.PolicyWithOptionalEffective, *api.Error) {
	defaultsMap := make(map[string]interface{})
	rulesMap := make(map[string]interface{})
	imposedAssets := &api.DistributedImposedAssets{}

	switch policy := genericPolicy.(type) {
	case api.WorkspacePolicyDefaultsAndRulesV2:
		rulesMap = loadFromJson(policy.Rules)
		defaultsMap = loadFromJson(policy.Defaults)
		imposedAssets.Worker = policy.ImposedAssets
	case api.TrainingPolicyDefaultsAndRulesV2:
		rulesMap = loadFromJson(policy.Rules)
		defaultsMap = loadFromJson(policy.Defaults)
		imposedAssets.Worker = policy.ImposedAssets
	case api.InferencePolicyDefaultsAndRulesV2:
		rulesMap = loadFromJson(policy.Rules)
		defaultsMap = loadFromJson(policy.Defaults)
		imposedAssets.Worker = policy.ImposedAssets
	case api.DistributedInferencePolicyDefaultsAndRulesV2:
		rulesMap = loadFromJson(policy.Rules)
		defaultsMap = loadFromJson(policy.Defaults)
		if policy.ImposedAssets != nil {
			imposedAssets.Worker = policy.ImposedAssets.Worker
			imposedAssets.Master = policy.ImposedAssets.Leader
		}
	case api.NimServicePolicyDefaultsAndRules:
		rulesMap = loadFromJson(policy.Rules)
		defaultsMap = loadFromJson(policy.Defaults)
		/*WAIT_FOR_OFER imposed assets not supported for nim service policy
		if policy.ImposedAssets != nil {
			imposedAssets.Worker = policy.ImposedAssets.Worker
			imposedAssets.Master = policy.ImposedAssets.Leader
		}
		*/
	case *api.WorkspacePolicyDefaultsAndRulesV2:
		if policy != nil {
			rulesMap = loadFromJson(policy.Rules)
			defaultsMap = loadFromJson(policy.Defaults)
			imposedAssets.Worker = policy.ImposedAssets
		}
	case *api.TrainingPolicyDefaultsAndRulesV2:
		if policy != nil {
			rulesMap = loadFromJson(policy.Rules)
			defaultsMap = loadFromJson(policy.Defaults)
			imposedAssets.Worker = policy.ImposedAssets
		}
	case *api.InferencePolicyDefaultsAndRulesV2:
		if policy != nil {
			rulesMap = loadFromJson(policy.Rules)
			defaultsMap = loadFromJson(policy.Defaults)
			imposedAssets.Worker = policy.ImposedAssets
		}
	case *api.DistributedInferencePolicyDefaultsAndRulesV2:
		if policy != nil {
			rulesMap = loadFromJson(policy.Rules)
			defaultsMap = loadFromJson(policy.Defaults)
			if policy.ImposedAssets != nil {
				imposedAssets.Worker = policy.ImposedAssets.Worker
				imposedAssets.Master = policy.ImposedAssets.Leader
			}
		}
	case *api.NimServicePolicyDefaultsAndRules:
		if policy != nil {
			rulesMap = loadFromJson(policy.Rules)
			defaultsMap = loadFromJson(policy.Defaults)
			/*WAIT_FOR_OFER imposed assets not supported for nim service policy
			if policy.ImposedAssets != nil {
				imposedAssets.Worker = policy.ImposedAssets.Worker
				imposedAssets.Master = policy.ImposedAssets.Leader
			}
			*/
		}
	case *api.FallbackPolicyDefaultsAndRules:
		if policy != nil {
			rulesMap = loadFromJson(policy.Rules)
			defaultsMap = loadFromJson(policy.Defaults)
			imposedAssets.Worker = policy.ImposedAssets
		}
	default:
		return nil, api.NewError(http.StatusInternalServerError, "unsupported policy type", nil)
	}

	result, applyError := applyPolicy(ctx, &tree.ApplyPolicyRequest{
		GenericPolicyRequest: tree.GenericPolicyRequest{
			WorkloadType:           workloadType,
			PolicyMeta:             meta,
			ApproveClusterDeletion: options.ApproveClusterDeletion,
		},
		Defaults:      defaultsMap,
		Rules:         rulesMap,
		ImposedAssets: imposedAssets,
		FieldsToReset: resetFields,
	}, options)
	if applyError != nil {
		return nil, applyError
	}

	return result, nil
}

func jsonBMapToResponse[T any](dest **T, src dbmodels.JsonBMap) {
	if src != nil {
		if (*dest) == nil {
			(*dest) = new(T)
		}
		projected := dbmodels.JsonBMap(gpucompat.LegacyCompatibleGPUPolicy(src))
		projected.Unmarshal(*dest)
	}
}

func loadFromJson(src any) map[string]interface{} {
	if src != nil {
		return convutils.Must(gojsonq.LoadJson(src))
	}
	return make(map[string]interface{})
}

func validateResetFields(resetFields []string, defaults map[string]interface{}, rules map[string]interface{}) *api.Error {
	defaultsFields := datastructs.NewSet[string]()
	for _, field := range policy_compliance.FieldsOfDefaults(defaults) {
		defaultsFields.Add(gpucompat.CanonicalGPUFieldPath(field))
	}
	rulesFields := datastructs.NewSet[string]()
	for _, field := range policy_compliance.FieldsWithRules(rules) {
		rulesFields.Add(gpucompat.CanonicalGPUFieldPath(field))
	}

	for _, resetField := range resetFields {
		resetField = gpucompat.CanonicalGPUFieldPath(resetField)
		if defaultsFields.Has(resetField) {
			return api.NewValidationError(fmt.Sprintf("%s: reset and defaults are mutually excluded", resetField))
		}
		if rulesFields.Has(resetField) {
			return api.NewValidationError(fmt.Sprintf("%s: reset and rules are mutually excluded", resetField))
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//           Nim Service
////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (pa *PolicyApiService) GetNimServicePolicy(ctx context.Context, scope string, departmentId string, projectId string, clusterId string, noCreationMode bool) (api.ImplResponse, error) {
	projectIdNumeric, err := projectIdToNum(projectId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	getPolicyRequest := tree.NewGenericPolicyRequest(tree.WorkloadTypeNimService, scope, departmentId, projectIdNumeric, clusterId, noCreationMode, noNeedClusterApproval)

	result, getError := getPolicy(ctx, getPolicyRequest, getOptions{})
	if getError != nil {
		return NewErrorResponse(getError), nil
	}

	return toNimServiceResponse(result), nil
}

func (pa *PolicyApiService) OverwriteNimServicePolicy(ctx context.Context, validateOnly bool, request api.NimServicePolicyOverwriteRequest) (api.ImplResponse, error) {
	return pa.ApplyNimServicePolicy(ctx, request.Meta, request.Policy, []string{}, tree.ApplyPolicyOptions{
		ValidateOnly: validateOnly,
		Overwrite:    true,

		ApproveClusterDeletion: noNeedClusterApproval,
	})
}

func (pa *PolicyApiService) UpdateNimServicePolicy(ctx context.Context, validateOnly bool, request api.NimServicePolicyChangeRequest) (api.ImplResponse, error) {
	return pa.ApplyNimServicePolicy(ctx, request.Meta, request.Policy, request.Reset, tree.ApplyPolicyOptions{
		ValidateOnly: validateOnly,
		Overwrite:    false,

		ApproveClusterDeletion: noNeedClusterApproval,
	})
}

func (pa *PolicyApiService) ApplyNimServicePolicy(ctx context.Context, meta api.PolicyCreationFields, policy *api.NimServicePolicyDefaultsAndRules,
	resetFields []string, options tree.ApplyPolicyOptions,
) (api.ImplResponse, error) {
	result, err := pa.applyWorkloadPolicyV2(ctx, meta, policy, resetFields, options, tree.WorkloadTypeNimService)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	if options.ValidateOnly {
		return api.Response(http.StatusNoContent, nil), nil
	}

	return toNimServiceResponse(result), nil
}

func (pa *PolicyApiService) DeleteNimServicePolicy(ctx context.Context, scope string, departmentId string, projectId string, clusterId string) (api.ImplResponse, error) {
	projectIdNumeric, errConv := projectIdToNum(projectId)
	if errConv != nil {
		return NewErrorResponse(errConv), nil
	}
	deletePolicyRequest := tree.NewGenericPolicyRequest(tree.WorkloadTypeNimService, scope, departmentId, projectIdNumeric, clusterId, false, noNeedClusterApproval)

	return deletePolicy(ctx, deletePolicyRequest), nil
}

func toNimServicePolicy(record *dbmodel.PolicyWithOptionalEffective) api.NimServicePolicy {
	response := api.NimServicePolicy{
		Meta:               toPolicyMeta(&record.Policy),
		Policy:             &api.NimServicePolicyDefaultsAndRules{},
		Effective:          &api.NimServicePolicyDefaultsAndRules{},
		EffectiveUpdatedAt: record.UpdatedAt,
	}

	jsonBMapToResponse(&response.Policy.Defaults, record.Policy.DefaultsV2)
	jsonBMapToResponse(&response.Policy.Rules, record.Policy.RulesV2)

	jsonBMapToResponse(&response.Effective.Defaults, record.EffectivePolicy.DefaultsV2)
	jsonBMapToResponse(&response.Effective.Rules, record.EffectivePolicy.RulesV2)

	/*WAIT_FOR_OFER no imposed assets for nim service policy
	if record.Policy.ImposedAssets != nil && record.Policy.ImposedAssets.Worker != nil && len(*record.Policy.ImposedAssets.Worker) != 0 {
		response.Policy.ImposedAssets = record.Policy.ImposedAssets.Worker
		response.Effective.ImposedAssets = record.Policy.ImposedAssets.Worker
	}
	*/

	response.Effective.DefaultFromOrder = toResponse(record.DefaultFromOrder)
	response.Effective.Status = toResponseStatus(record.ValidationError)

	return response
}

func toNimServiceResponse(record *dbmodel.PolicyWithOptionalEffective) api.ImplResponse {
	return api.Response(http.StatusOK, toNimServicePolicy(record))
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//           Fallback
////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (pa *PolicyApiService) ApplyFallbackPolicy(ctx context.Context, meta api.PolicyCreationFields, policy *api.FallbackPolicyDefaultsAndRules,
	resetFields []string, options tree.ApplyPolicyOptions,
) (api.ImplResponse, error) {
	result, err := pa.applyWorkloadPolicyV2(ctx, meta, policy, resetFields, options, tree.WorkloadTypeFallback)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	if options.ValidateOnly {
		return api.Response(http.StatusNoContent, nil), nil
	}

	return toFallbackPolicyResponse(result), nil
}

func (pa *PolicyApiService) UpdateFallbackPolicy(ctx context.Context, validateOnly bool, approveClusterDeletion bool, request api.FallbackPolicyChangeRequest) (api.ImplResponse, error) {
	return pa.ApplyFallbackPolicy(ctx, request.Meta, request.Policy, request.Reset, tree.ApplyPolicyOptions{
		ValidateOnly: validateOnly,
		Overwrite:    false,

		ApproveClusterDeletion: approveClusterDeletion,
	})
}

func (pa *PolicyApiService) OverwriteFallbackPolicy(ctx context.Context, validateOnly bool, approveClusterDeletion bool, request api.FallbackPolicyOverwriteRequest) (api.ImplResponse, error) {
	return pa.ApplyFallbackPolicy(ctx, request.Meta, request.Policy, []string{}, tree.ApplyPolicyOptions{
		ValidateOnly: validateOnly,
		Overwrite:    true,

		ApproveClusterDeletion: approveClusterDeletion,
	})
}

func (pa *PolicyApiService) GetFallbackPolicy(ctx context.Context, scope string, departmentId string, projectId string, clusterId string, noCreationMode bool) (api.ImplResponse, error) {
	projectIdNumeric, err := projectIdToNum(projectId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	getPolicyRequest := tree.NewGenericPolicyRequest(tree.WorkloadTypeFallback, scope, departmentId, int32(projectIdNumeric), clusterId, noCreationMode, noNeedClusterApproval)

	result, getError := getPolicy(ctx, getPolicyRequest, getOptions{})
	if getError != nil {
		return NewErrorResponse(getError), nil
	}

	return toFallbackPolicyResponse(result), nil
}

func (pa *PolicyApiService) DeleteFallbackPolicy(ctx context.Context, scope string, departmentId string, projectId string, clusterId string, approveClusterDeletion bool) (api.ImplResponse, error) {
	projectIdNumeric, errConv := projectIdToNum(projectId)
	if errConv != nil {
		return NewErrorResponse(errConv), nil
	}

	deletePolicyRequest := tree.NewGenericPolicyRequest(tree.WorkloadTypeFallback, scope, departmentId, int32(projectIdNumeric), clusterId, false, approveClusterDeletion)

	return deletePolicy(ctx, deletePolicyRequest), nil
}

func toFallbackPolicyResponse(record *dbmodel.PolicyWithOptionalEffective) api.ImplResponse {
	response := api.FallbackPolicy{
		Meta:               toPolicyMeta(&record.Policy),
		Policy:             &api.FallbackPolicyDefaultsAndRules{},
		Effective:          &api.FallbackPolicyDefaultsAndRules{},
		EffectiveUpdatedAt: record.UpdatedAt,
	}

	jsonBMapToResponse(&response.Policy.Defaults, record.Policy.DefaultsV2)
	jsonBMapToResponse(&response.Policy.Rules, record.Policy.RulesV2)

	jsonBMapToResponse(&response.Effective.Defaults, record.EffectivePolicy.DefaultsV2)
	jsonBMapToResponse(&response.Effective.Rules, record.EffectivePolicy.RulesV2)

	if record.Policy.ImposedAssets != nil && record.Policy.ImposedAssets.Worker != nil && len(*record.Policy.ImposedAssets.Worker) != 0 {
		response.Policy.ImposedAssets = record.Policy.ImposedAssets.Worker
		response.Effective.ImposedAssets = record.Policy.ImposedAssets.Worker
	}

	response.Effective.DefaultFromOrder = toResponse(record.DefaultFromOrder)
	response.Effective.Status = toResponseStatus(record.ValidationError)

	return api.Response(http.StatusOK, response)
}

var workloadTypes = []tree.WorkloadType{
	tree.WorkloadTypeWorkspace,
	tree.WorkloadTypeTraining,
	tree.WorkloadTypeDistributed,
	tree.WorkloadTypeInference,
	tree.WorkloadTypeDistributedInference,
	tree.WorkloadTypeNimService,
}

func (pa *PolicyApiService) GetPolicyCrossType(ctx context.Context, scope string, departmentId string, projectId string, clusterId string, noCreationMode bool) (api.ImplResponse, error) {
	response := api.PolicyCrossType{}

	projectIdNumeric, err := projectIdToNum(projectId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	for _, workloadType := range workloadTypes {
		getPolicyRequest := tree.NewGenericPolicyRequest(workloadType, scope, departmentId, projectIdNumeric, clusterId, noCreationMode, noNeedClusterApproval)

		result, getError := getPolicy(ctx, getPolicyRequest, getOptions{bypassAuth: true})
		if getError != nil {
			if getError.Code == http.StatusNotFound {
				continue
			}
			return NewErrorResponse(getError), nil
		}

		switch workloadType {
		case tree.WorkloadTypeWorkspace:
			response.Workspace = convutils.Ref(toWorkspaceV2(result))
		case tree.WorkloadTypeTraining:
			response.Training = convutils.Ref(toTrainingV2(result))
		case tree.WorkloadTypeDistributed:
			response.Distributed = convutils.Ref(toDistributed(result))
		case tree.WorkloadTypeInference:
			response.Inference = convutils.Ref(toInferenceV2(result))
		case tree.WorkloadTypeDistributedInference:
			response.DistributedInference = convutils.Ref(toDistributedInferenceV2(result))
		case tree.WorkloadTypeNimService:
			response.NimService = convutils.Ref(toNimServicePolicy(result))
		}
	}

	return api.Response(http.StatusOK, response), nil
}

func (pa *PolicyApiService) GetSyncConfig(ctx context.Context, clusterId string) (api.ImplResponse, error) {
	config, err := dao.GetApprovedDeletionsInfoOfCluster(ctx, clusterId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	if config == nil {
		return NewErrorResponse(api.NewError(http.StatusNotFound, "syncConfig information is not available for this cluster", nil)), nil
	}

	return api.Response(http.StatusOK, config), nil
}

func (pa *PolicyApiService) UpdateSyncConfig(ctx context.Context, config api.SyncConfig) (api.ImplResponse, error) {
	result, err := dao.UpdateApprovedDeletionOfCluster(ctx, config)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	if result == nil {
		return NewErrorResponse(api.NewError(http.StatusNotFound, "sync config information is not available for this cluster", nil)), nil
	}

	return api.Response(http.StatusOK, result), nil
}
