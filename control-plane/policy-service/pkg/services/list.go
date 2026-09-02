/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/cluster"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	authApi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/tree"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const PolicyStatusReady = "Ready"

func (pa *PolicyApiService) ListPolicies(ctx context.Context, workloadType string, scopeType string, departmentId string,
	projectId string, clusterId string, includeFallbackPolicies bool, includeGvkPolicies bool,
) (api.ImplResponse, error) {
	projectIdNumeric, errConv := projectIdToNum(projectId)
	if errConv != nil {
		return NewErrorResponse(errConv), nil
	}

	permittedScopes, authErr := authorization.Get().Client().CheckPermittedScopes(ctx, authApi.RESOURCETYPE_POLICIES, authApi.ACTION_READ)
	if authErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("check permitted scopes failed: %s: %v", authErr.Message, authErr.Details)
		return NewErrorResponse(api.NewError(int(authErr.Code), authErr.Message, nil)), nil
	}

	entries, err := dao.ListGivenPoliciesOfTenant(ctx, workloadType, scopeType, departmentId, projectIdNumeric, clusterId, permittedScopes, dao.ListOptions{
		IncludeDeleted:          dbmodel.NotToIncludeDeletedPolicies,
		IncludeFallbackPolicies: includeFallbackPolicies,
		IncludeNativePolicies:   true,
		IncludeGvkPolicies:      includeGvkPolicies,
	})
	if err != nil {
		return NewErrorResponse(err), nil
	}

	clusterIds, err := cluster.GetClustersInfo(ctx, nil)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	var result api.PolicyListResponse

	for _, entry := range entries {
		var validationStatus *api.PolicyValidationStatusValidation
		if entry.ValidationError != nil {
			validationStatus = &api.PolicyValidationStatusValidation{
				ErrorMessage: *entry.ValidationError,
			}
		}

		policyEntry := api.PolicyListEntry{
			Meta: toPolicyMeta(&entry.Policy),
			Status: api.PolicyListEntryStatus{
				Clusters:   make([]api.PolicySyncStatusOfCluster, 0, len(clusterIds)),
				Validation: validationStatus,
			},
		}

		if entry.GvkWorkload {
			policyEntry.Type = api.PolicyType(entry.WorkloadType)
			if entry.ValidationStatus != nil {
				policyEntry.Availability = &api.PolicyStatus{
					Phase:  *entry.ValidationStatus,
					Issues: entry.ValidationIssues,
				}
			}
		} else {
			policyEntry.Type = policyTypeOf(ctx, entry.WorkloadType)
		}

		for clusterId := range clusterIds {
			if policyEntry.Meta.Scope == api.SCOPETYPE_TENANT || convutils.Deref(policyEntry.Meta.ClusterId, "") == clusterId {
				policyEntry.Status.Clusters = append(policyEntry.Status.Clusters, api.PolicySyncStatusOfCluster{
					ClusterId: clusterId,
					Status:    PolicyStatusReady,
				})
			}
		}

		result.Policies = append(result.Policies, policyEntry)
	}

	return api.Response(http.StatusOK, result), nil
}

func policyTypeOf(ctx context.Context, workloadType string) api.PolicyType {
	switch tree.WorkloadType(workloadType) {
	case tree.WorkloadTypeTraining:
		return api.POLICYTYPE_TRAINING
	case tree.WorkloadTypeWorkspace:
		return api.POLICYTYPE_WORKSPACE
	case tree.WorkloadTypeDistributed:
		return api.POLICYTYPE_DISTRIBUTED
	case tree.WorkloadTypeInference:
		return api.POLICYTYPE_INFERENCE
	case tree.WorkloadTypeDistributedInference:
		return api.POLICYTYPE_DISTRIBUTED_INFERENCE
	case tree.WorkloadTypeNimService:
		return api.POLICYTYPE_NIM_SERVICE
	case tree.WorkloadTypeFallback:
		return api.POLICYTYPE_FALLBACK
	default:
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("unsupported workload type %s", workloadType)
		return ""
	}
}
