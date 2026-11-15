/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policies

import (
	"context"
	"fmt"
	"net/http"

	policy "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

var errUnsupportedLegacyWorkloadType = fmt.Errorf("unsupported workloadType: %w", runaimcp.ErrValidation)

func (in manageWorkloadPolicyInput) validateLegacyPolicyInput() error {
	if in.PolicyID != "" {
		return fmt.Errorf("policyId is only used for GVK policies (policyVersion=gvk)")
	}
	if in.ApproveClusterDeletion && policy.PolicyType(in.WorkloadType) == policy.POLICYTYPE_NIM_SERVICE {
		return fmt.Errorf("nimService does not support approveClusterDeletion")
	}
	if in.Action == actionDelete {
		if err := common.ValidatePolicyScope(in.Scope, in.ClusterID, in.DepartmentID, in.ProjectID); err != nil {
			return err
		}
		_, err := policy.NewPolicyTypeFromValue(in.WorkloadType)
		return err
	}
	if in.Scope != "" || in.ClusterID != "" || in.DepartmentID != "" || in.ProjectID != "" {
		return fmt.Errorf("legacy create/update use request.meta for scope and scope IDs; omit top-level scope, clusterId, departmentId, and projectId")
	}
	_, err := decodeLegacyPolicyRequest(in.WorkloadType, in.Action, in.Request)
	return err
}

func decodeLegacyPolicyRequest(kind, action string, raw map[string]any) (any, error) {
	var request any
	switch policy.PolicyType(kind) {
	case policy.POLICYTYPE_TRAINING:
		if action == actionCreate {
			request = &policy.TrainingPolicyOverwriteRequestV2{}
		} else {
			request = &policy.TrainingPolicyChangeRequestV2{}
		}
	case policy.POLICYTYPE_WORKSPACE:
		if action == actionCreate {
			request = &policy.WorkspacePolicyOverwriteRequestV2{}
		} else {
			request = &policy.WorkspacePolicyChangeRequestV2{}
		}
	case policy.POLICYTYPE_DISTRIBUTED:
		if action == actionCreate {
			request = &policy.DistributedPolicyOverwriteRequestV2{}
		} else {
			request = &policy.DistributedPolicyChangeRequestV2{}
		}
	case policy.POLICYTYPE_INFERENCE:
		if action == actionCreate {
			request = &policy.InferencePolicyOverwriteRequestV2{}
		} else {
			request = &policy.InferencePolicyChangeRequestV2{}
		}
	case policy.POLICYTYPE_DISTRIBUTED_INFERENCE:
		if action == actionCreate {
			request = &policy.DistributedInferencePolicyOverwriteRequestV2{}
		} else {
			request = &policy.DistributedInferencePolicyChangeRequestV2{}
		}
	case policy.POLICYTYPE_NIM_SERVICE:
		if action == actionCreate {
			request = &policy.NimServicePolicyOverwriteRequest{}
		} else {
			request = &policy.NimServicePolicyChangeRequest{}
		}
	case policy.POLICYTYPE_FALLBACK:
		if action == actionCreate {
			request = &policy.FallbackPolicyOverwriteRequest{}
		} else {
			request = &policy.FallbackPolicyChangeRequest{}
		}
	default:
		return nil, fmt.Errorf("unsupported workloadType %q", kind)
	}
	if err := common.DecodeTypedRequest(raw, request); err != nil {
		return nil, err
	}
	return request, nil
}

func executeLegacyPolicyOperation(ctx context.Context, client *runaiclient.Client, in manageWorkloadPolicyInput) (map[string]any, error) {
	var request any
	var resp *http.Response
	var err error
	if in.Action == actionCreate || in.Action == actionUpdate {
		request, err = decodeLegacyPolicyRequest(in.WorkloadType, in.Action, in.Request)
		if err != nil {
			return nil, fmt.Errorf("policy request: %w: %w", err, runaimcp.ErrValidation)
		}
	}

	switch in.Action {
	case actionCreate:
		resp, err = overwriteLegacyPolicyResponse(ctx, client, in, request)
	case actionUpdate:
		resp, err = patchLegacyPolicyResponse(ctx, client, in, request)
	case actionDelete:
		resp, err = deleteLegacyPolicyResponse(ctx, client, in)
	default:
		return nil, fmt.Errorf("unsupported operation %q: %w", in.Action, runaimcp.ErrValidation)
	}
	return common.ReadAPIResponse(resp, err, in.Action == actionDelete)
}

func overwriteLegacyPolicyResponse(ctx context.Context, client *runaiclient.Client, in manageWorkloadPolicyInput, request any) (*http.Response, error) {
	api := client.Policy.PolicyAPI
	var resp *http.Response
	var err error
	switch policy.PolicyType(in.WorkloadType) {
	case policy.POLICYTYPE_TRAINING:
		_, resp, err = api.OverwriteTrainingPolicyV2(ctx).TrainingPolicyOverwriteRequestV2(*request.(*policy.TrainingPolicyOverwriteRequestV2)).ApproveClusterDeletion(in.ApproveClusterDeletion).Execute()
	case policy.POLICYTYPE_WORKSPACE:
		_, resp, err = api.OverwriteWorkspacePolicyV2(ctx).WorkspacePolicyOverwriteRequestV2(*request.(*policy.WorkspacePolicyOverwriteRequestV2)).ApproveClusterDeletion(in.ApproveClusterDeletion).Execute()
	case policy.POLICYTYPE_DISTRIBUTED:
		_, resp, err = api.OverwriteDistributedPolicyV2(ctx).DistributedPolicyOverwriteRequestV2(*request.(*policy.DistributedPolicyOverwriteRequestV2)).ApproveClusterDeletion(in.ApproveClusterDeletion).Execute()
	case policy.POLICYTYPE_INFERENCE:
		_, resp, err = api.OverwriteInferencePolicyV2(ctx).InferencePolicyOverwriteRequestV2(*request.(*policy.InferencePolicyOverwriteRequestV2)).ApproveClusterDeletion(in.ApproveClusterDeletion).Execute()
	case policy.POLICYTYPE_DISTRIBUTED_INFERENCE:
		_, resp, err = api.OverwriteDistributedInferencePolicyV2(ctx).
			DistributedInferencePolicyOverwriteRequestV2(*request.(*policy.DistributedInferencePolicyOverwriteRequestV2)).
			ApproveClusterDeletion(in.ApproveClusterDeletion).
			Execute()
	case policy.POLICYTYPE_NIM_SERVICE:
		_, resp, err = api.OverwriteNimServicePolicy(ctx).NimServicePolicyOverwriteRequest(*request.(*policy.NimServicePolicyOverwriteRequest)).Execute()
	case policy.POLICYTYPE_FALLBACK:
		_, resp, err = api.OverwriteFallbackPolicy(ctx).FallbackPolicyOverwriteRequest(*request.(*policy.FallbackPolicyOverwriteRequest)).ApproveClusterDeletion(in.ApproveClusterDeletion).Execute()
	default:
		return nil, errUnsupportedLegacyWorkloadType
	}
	return resp, err
}

func patchLegacyPolicyResponse(ctx context.Context, client *runaiclient.Client, in manageWorkloadPolicyInput, request any) (*http.Response, error) {
	api := client.Policy.PolicyAPI
	var resp *http.Response
	var err error
	switch policy.PolicyType(in.WorkloadType) {
	case policy.POLICYTYPE_TRAINING:
		_, resp, err = api.UpdateTrainingPolicyV2(ctx).TrainingPolicyChangeRequestV2(*request.(*policy.TrainingPolicyChangeRequestV2)).ApproveClusterDeletion(in.ApproveClusterDeletion).Execute()
	case policy.POLICYTYPE_WORKSPACE:
		_, resp, err = api.UpdateWorkspacePolicyV2(ctx).WorkspacePolicyChangeRequestV2(*request.(*policy.WorkspacePolicyChangeRequestV2)).ApproveClusterDeletion(in.ApproveClusterDeletion).Execute()
	case policy.POLICYTYPE_DISTRIBUTED:
		_, resp, err = api.UpdateDistributedPolicyV2(ctx).DistributedPolicyChangeRequestV2(*request.(*policy.DistributedPolicyChangeRequestV2)).ApproveClusterDeletion(in.ApproveClusterDeletion).Execute()
	case policy.POLICYTYPE_INFERENCE:
		_, resp, err = api.UpdateInferencePolicyV2(ctx).InferencePolicyChangeRequestV2(*request.(*policy.InferencePolicyChangeRequestV2)).ApproveClusterDeletion(in.ApproveClusterDeletion).Execute()
	case policy.POLICYTYPE_DISTRIBUTED_INFERENCE:
		_, resp, err = api.UpdateDistributedInferencePolicyV2(ctx).
			DistributedInferencePolicyChangeRequestV2(*request.(*policy.DistributedInferencePolicyChangeRequestV2)).
			ApproveClusterDeletion(in.ApproveClusterDeletion).
			Execute()
	case policy.POLICYTYPE_NIM_SERVICE:
		_, resp, err = api.UpdateNimServicePolicy(ctx).NimServicePolicyChangeRequest(*request.(*policy.NimServicePolicyChangeRequest)).Execute()
	case policy.POLICYTYPE_FALLBACK:
		_, resp, err = api.UpdateFallbackPolicy(ctx).FallbackPolicyChangeRequest(*request.(*policy.FallbackPolicyChangeRequest)).ApproveClusterDeletion(in.ApproveClusterDeletion).Execute()
	default:
		return nil, errUnsupportedLegacyWorkloadType
	}
	return resp, err
}

func deleteLegacyPolicyResponse(ctx context.Context, client *runaiclient.Client, in manageWorkloadPolicyInput) (*http.Response, error) {
	api := client.Policy.PolicyAPI
	var resp *http.Response
	var err error
	switch policy.PolicyType(in.WorkloadType) {
	case policy.POLICYTYPE_TRAINING:
		resp, err = api.DeleteTrainingPolicy(ctx).Scope(in.Scope).ClusterId(in.ClusterID).DepartmentId(in.DepartmentID).ProjectId(in.ProjectID).ApproveClusterDeletion(in.ApproveClusterDeletion).Execute()
	case policy.POLICYTYPE_WORKSPACE:
		resp, err = api.DeleteWorkspacePolicy(ctx).Scope(in.Scope).ClusterId(in.ClusterID).DepartmentId(in.DepartmentID).ProjectId(in.ProjectID).ApproveClusterDeletion(in.ApproveClusterDeletion).Execute()
	case policy.POLICYTYPE_DISTRIBUTED:
		resp, err = api.DeleteDistributedPolicy(ctx).Scope(in.Scope).ClusterId(in.ClusterID).DepartmentId(in.DepartmentID).ProjectId(in.ProjectID).ApproveClusterDeletion(in.ApproveClusterDeletion).Execute()
	case policy.POLICYTYPE_INFERENCE:
		resp, err = api.DeleteInferencePolicy(ctx).Scope(in.Scope).ClusterId(in.ClusterID).DepartmentId(in.DepartmentID).ProjectId(in.ProjectID).ApproveClusterDeletion(in.ApproveClusterDeletion).Execute()
	case policy.POLICYTYPE_DISTRIBUTED_INFERENCE:
		resp, err = api.DeleteDistributedInferencePolicy(ctx).
			Scope(in.Scope).
			ClusterId(in.ClusterID).
			DepartmentId(in.DepartmentID).
			ProjectId(in.ProjectID).
			ApproveClusterDeletion(in.ApproveClusterDeletion).
			Execute()
	case policy.POLICYTYPE_NIM_SERVICE:
		resp, err = api.DeleteNimServicePolicy(ctx).Scope(in.Scope).ClusterId(in.ClusterID).DepartmentId(in.DepartmentID).ProjectId(in.ProjectID).Execute()
	case policy.POLICYTYPE_FALLBACK:
		resp, err = api.DeleteFallbackPolicy(ctx).Scope(in.Scope).ClusterId(in.ClusterID).DepartmentId(in.DepartmentID).ProjectId(in.ProjectID).ApproveClusterDeletion(in.ApproveClusterDeletion).Execute()
	default:
		return nil, errUnsupportedLegacyWorkloadType
	}
	return resp, err
}
