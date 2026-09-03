/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tree

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/scopes"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/jinzhu/copier"
)

const EmptyPolicyCreator = "runai-internal"

type GenericPolicyRequest struct {
	Requester      string
	TenantId       int32
	WorkloadType   WorkloadType
	PolicyMeta     api.PolicyCreationFields
	NoCreationMode bool

	// old cluster 2.20 and below require cluster deletion approval of corresponding policies
	// this is done to avoid conflicts b/w control pland and cluster policies
	ApproveClusterDeletion bool
}

func (gpr *GenericPolicyRequest) Clone() *GenericPolicyRequest {
	result := GenericPolicyRequest{
		Requester:      gpr.Requester,
		TenantId:       gpr.TenantId,
		WorkloadType:   gpr.WorkloadType,
		NoCreationMode: gpr.NoCreationMode,

		ApproveClusterDeletion: gpr.ApproveClusterDeletion,
	}

	copier.Copy(&result.PolicyMeta, &gpr.PolicyMeta)

	return &result
}

func (gpr *GenericPolicyRequest) ToFallback() *GenericPolicyRequest {
	result := gpr.Clone()
	result.WorkloadType = WorkloadTypeFallback
	return result
}

func NewGenericPolicyRequest(workloadType WorkloadType, scope string, departmentId string, projectId int32, clusterId string,
	noCreationMode bool, approveClusterDeletion bool,
) *GenericPolicyRequest {
	request := GenericPolicyRequest{
		WorkloadType: workloadType,
		PolicyMeta: api.PolicyCreationFields{
			Scope: api.ScopeType(scope),
		},
		NoCreationMode:         noCreationMode,
		ApproveClusterDeletion: approveClusterDeletion,
	}

	if departmentId != "" {
		request.PolicyMeta.DepartmentId = convutils.Ref(departmentId)
	}

	if projectId != 0 {
		request.PolicyMeta.ProjectId = convutils.Ref(projectId)
	}

	if clusterId != "" {
		request.PolicyMeta.ClusterId = convutils.Ref(clusterId)
	}

	return &request
}

func NewEmptyPolicy(ctx context.Context, request *GenericPolicyRequest, orgUnits []scopes.OrgUnit) *dbmodel.Policy {
	name := fmt.Sprintf("empty-policy-%s-%s", request.PolicyMeta.Scope, orgUnits[0].Id)

	record := dbmodel.Policy{
		BaseModel: dbmodels.BaseModel{
			CreatedBy: EmptyPolicyCreator,
			UpdatedBy: EmptyPolicyCreator,
		},
		Name:         name,
		WorkloadType: string(request.WorkloadType),
		TenantId:     request.TenantId,
		Scope:        request.PolicyMeta.Scope,
		DefaultsV2:   nil,
		RulesV2:      nil,
	}

	patchOrgUnits(ctx, &record, orgUnits)

	return &record
}
