/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/scopes"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
)

func (pa *PolicyApiService) GetGvkPolicyByTypeAndScope(ctx context.Context, req api.WorkloadTypeAndScope) (api.ImplResponse, error) {
	// this is an internal command used by the workload manager for applying policy.
	// it needs both the effective policy and the breakdown.
	result, err := pa.getGvkPolicyByTypeAndScope(ctx, req, addBreakdown)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	return api.Response(http.StatusOK, result), nil
}

func (pa *PolicyApiService) GetEffectivePolicy(ctx context.Context, req api.WorkloadTypeAndScope) (api.ImplResponse, error) {
	// this is an external API that needs the effective policy. it is used by
	// the frontend when view-policy is requested for a given project
	result, err := pa.getEffectivePolicy(ctx, req)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, result), nil
}

func (pa *PolicyApiService) GetMappingMinimal(ctx context.Context, req api.GetMappingMinimalInfoReq) (api.ImplResponse, error) {
	result, err := pa.getMappingMinimal(ctx, req.WorkloadType)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	return api.Response(http.StatusOK, result), nil
}

func (pa *PolicyApiService) GetMappingFieldInfo(ctx context.Context, req api.GetMappingFieldInfoReq) (api.ImplResponse, error) {
	result, err := pa.getMappingFieldInfo(ctx, req.WorkloadType, req.FieldsRef)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	return api.Response(http.StatusOK, result), nil
}

func (pa *PolicyApiService) GetGvkPolicy(ctx context.Context, policyId string, toAddBreakdown bool) (api.ImplResponse, error) {
	result, err := pa.getGvkPolicyById(ctx, policyId, toAddBreakdown)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, result), nil
}

func (pa *PolicyApiService) CreateGvkPolicy(ctx context.Context, req api.CreateGvkPolicyReq) (api.ImplResponse, error) {
	result, err := pa.createGvkPolicy(ctx, req)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, result), nil
}

func (pa *PolicyApiService) ListGvkPolicies(ctx context.Context, filterBy []string, sortBy string,
	offset int32, limit int32, sortOrder string,
) (api.ImplResponse, error) {
	res, err := pa.listGvkPolicies(ctx, filterBy, sortBy, offset, limit, sortOrder)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, res), nil
}

type updateOptions struct {
	overwrite bool // put = overwrite, patch = non-overwrite
}

func (pa *PolicyApiService) ApplyGvkPolicy(ctx context.Context, policyId string, req api.UpdateGvkPolicyReq) (api.ImplResponse, error) {
	result, err := pa.updateGvkPolicy(ctx, policyId, req, updateOptions{overwrite: true})
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, result), nil
}

func (pa *PolicyApiService) PatchGvkPolicy(ctx context.Context, policyId string, req api.UpdateGvkPolicyReq) (api.ImplResponse, error) {
	result, err := pa.updateGvkPolicy(ctx, policyId, req, updateOptions{overwrite: false})
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, result), nil
}

func (pa *PolicyApiService) DeleteGvkPolicy(ctx context.Context, policyId string) (api.ImplResponse, error) {
	if err := pa.deleteGvkPolicy(ctx, policyId); err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusNoContent, nil), nil
}

func (pa *PolicyApiService) DeleteGvkPolicyRules(ctx context.Context, policyId string, req api.DeleteGvkPolicyRulesReq) (api.ImplResponse, error) {
	result, err := pa.deleteGvkPolicyRules(ctx, policyId, req)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, result), nil
}

func applyCreateReqOnPolicy(ctx context.Context, result *dbmodel.PolicyWithOptionalEffective, req api.CreateGvkPolicyReq,
	workloadTypeKey string, rulesV2 dao_utils.JsonBMap,
) *api.Error {
	tenantId := auth.TenantIdOf(ctx)
	now := time.Now()
	subject := auth.SubjectIdOf(ctx)

	var projectId *int32
	if req.ProjectId != nil && *req.ProjectId != "" {
		n, err := projectIdToNum(*req.ProjectId)
		if err != nil {
			return err
		}
		projectId = &n
	}

	appliedOrgUnit, mkErr := scopes.MakeOrgUnit(req.Scope, tenantId, req.ClusterId, req.DepartmentId, projectId)
	if mkErr != nil {
		return mkErr
	}

	// derive ancestors (department + cluster) so the policy row carries the full org-unit lineage,
	// which is required for visibility/permission filters that look up policies by cluster or department.
	var departmentId, clusterId *string
	ancestors, ancErr := appliedOrgUnit.Ancestors(ctx)
	if ancErr != nil {
		return ancErr
	}
	for _, ou := range ancestors {
		switch ou.Scope {
		case api.SCOPETYPE_DEPARTMENT:
			departmentId = convutils.Ref(ou.Id)
		case api.SCOPETYPE_CLUSTER:
			clusterId = convutils.Ref(ou.Id)
		}
	}

	var name string
	if req.Description != nil && *req.Description != "" {
		name = *req.Description
	} else {
		name = autoGeneratedName(workloadTypeKey, req.Scope, req.ClusterId, req.DepartmentId, req.ProjectId)
	}

	readyPhase := api.POLICYPHASE_READY

	var policyId uuid.UUID
	if result.ID == uuid.Nil {
		policyId = uuid.New()
	} else {
		policyId = result.ID
	}

	(*result).Policy = dbmodel.Policy{
		BaseModel: dao_utils.BaseModel{
			ID:        policyId,
			CreatedAt: now,
			UpdatedAt: now,
			CreatedBy: subject,
			UpdatedBy: subject,
		},
		Name:             name,
		WorkloadType:     workloadTypeKey,
		Scope:            req.Scope,
		TenantId:         tenantId,
		ClusterId:        clusterId,
		DepartmentId:     departmentId,
		ProjectId:        projectId,
		RulesV2:          rulesV2,
		GvkWorkload:      true,
		ValidationStatus: &readyPhase,
		ValidationIssues: []string{},
	}

	(*result).EffectivePolicy = dbmodel.EffectivePolicy{
		PolicyId:  policyId,
		TenantId:  tenantId,
		UpdatedAt: time.Now(),
	}

	return nil
}

func encodeGvkWorkloadType(gvk api.Gvk) (string, *api.Error) {
	group := convutils.Deref(gvk.Group, "")
	if gvk.Version == "" || gvk.Kind == "" {
		return "", api.NewValidationError("invalid GVK: version, kind are required")
	}

	return fmt.Sprintf("%s/%s/%s", group, gvk.Kind, gvk.Version), nil
}

func autoGeneratedName(workloadType string, scope api.ScopeType, clusterId *string, departmentId *string, projectId *string) string {
	result := fmt.Sprintf("%s-%s", workloadType, scope)

	if scope == api.SCOPETYPE_CLUSTER && clusterId != nil {
		result += "-" + *clusterId
	}
	if scope == api.SCOPETYPE_DEPARTMENT && departmentId != nil {
		result += "-" + *departmentId
	}
	if scope == api.SCOPETYPE_PROJECT && projectId != nil {
		result += "-" + *projectId
	}

	return result
}
