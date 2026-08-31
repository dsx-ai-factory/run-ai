/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policies

import (
	"context"
	"net/http"
	"time"

	policyService "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
)

type ProjectAndType struct {
	Group     string
	Kind      string
	Version   string
	ProjectId string
}

func gvkToKey(gvk policyService.GVK) (group, kind, version string) {
	if groupPtr := gvk.Group.Get(); groupPtr != nil {
		group = *groupPtr
	}
	return group, gvk.Kind, gvk.Version
}

// SimpleMockClient only supports fixed policy. it will always return the same policy and fixed update time
type SimpleMockClient struct {
	EffectiveUpdatedAt                    time.Time
	EffectiveWorkspacePolicyV2            *policyService.WorkspacePolicyDefaultsAndRulesV2
	EffectiveTrainingPolicyV2             *policyService.TrainingPolicyDefaultsAndRulesV2
	EffectiveDistributedPolicyV2          *policyService.DistributedPolicyDefaultsAndRulesV2
	EffectiveInferencePolicyV2            *policyService.InferencePolicyDefaultsAndRulesV2
	EffectiveDistributedInferencePolicyV2 *policyService.DistributedInferencePolicyDefaultsAndRulesV2
	EffectiveNimServicePolicy             *policyService.NimServicePolicyDefaultsAndRules

	GvkPolicyOfProject map[ProjectAndType]*policyService.GetGvkPolicyResp
}

func (smc *SimpleMockClient) GetWorkspacePolicyOfProjectV2(_ context.Context, _ /*projectId*/ int32, _ /*options*/ *PolicyClientOptions) (*policyService.WorkspacePolicyV2, *policyService.Error) {
	result := policyService.WorkspacePolicyV2{
		EffectiveUpdatedAt: &smc.EffectiveUpdatedAt,
	}
	if smc.EffectiveWorkspacePolicyV2 != nil {
		result.SetEffective(*smc.EffectiveWorkspacePolicyV2)
	}
	return &result, nil
}

func (smc *SimpleMockClient) GetTrainingPolicyOfProjectV2(_ context.Context, _ /*projectId*/ int32, _ /*options*/ *PolicyClientOptions) (*policyService.TrainingPolicyV2, *policyService.Error) {
	result := policyService.TrainingPolicyV2{
		EffectiveUpdatedAt: &smc.EffectiveUpdatedAt,
	}
	if smc.EffectiveTrainingPolicyV2 != nil {
		result.SetEffective(*smc.EffectiveTrainingPolicyV2)
	}
	return &result, nil
}

func (smc *SimpleMockClient) GetDistributedPolicyOfProjectV2(_ context.Context, _ /*projectId*/ int32, _ /*options*/ *PolicyClientOptions) (*policyService.DistributedPolicyV2, *policyService.Error) {
	result := policyService.DistributedPolicyV2{
		EffectiveUpdatedAt: &smc.EffectiveUpdatedAt,
	}
	if smc.EffectiveDistributedPolicyV2 != nil {
		result.SetEffective(*smc.EffectiveDistributedPolicyV2)
	}
	return &result, nil
}

func (smc *SimpleMockClient) GetInferencePolicyOfProjectV2(_ context.Context, _ /*projectId*/ int32, _ /*options*/ *PolicyClientOptions) (*policyService.InferencePolicyV2, *policyService.Error) {
	result := policyService.InferencePolicyV2{
		EffectiveUpdatedAt: &smc.EffectiveUpdatedAt,
	}
	if smc.EffectiveInferencePolicyV2 != nil {
		result.SetEffective(*smc.EffectiveInferencePolicyV2)
	}
	return &result, nil
}

func (smc *SimpleMockClient) GetDistributedInferencePolicyOfProjectV2(ctx context.Context, projectId int32, options *PolicyClientOptions) (*policyService.DistributedInferencePolicyV2, *policyService.Error) {
	result := policyService.DistributedInferencePolicyV2{
		EffectiveUpdatedAt: &smc.EffectiveUpdatedAt,
	}
	if smc.EffectiveDistributedInferencePolicyV2 != nil {
		result.SetEffective(*smc.EffectiveDistributedInferencePolicyV2)
	}
	return &result, nil
}

func (smc *SimpleMockClient) GetNimServicePolicyOfProject(ctx context.Context, projectId int32, options *PolicyClientOptions) (*policyService.NimServicePolicy, *policyService.Error) {
	result := policyService.NimServicePolicy{
		EffectiveUpdatedAt: &smc.EffectiveUpdatedAt,
	}
	if smc.EffectiveNimServicePolicy != nil {
		result.SetEffective(*smc.EffectiveNimServicePolicy)
	}
	return &result, nil
}

func (smc *SimpleMockClient) NotifyAssetModified(ctx context.Context, assetId string, scopeType policyService.ScopeType, scopeId string) *policyService.Error {
	return nil
}

func (smc *SimpleMockClient) NotifyAssetDeleted(ctx context.Context, assetId string, scopeType policyService.ScopeType, scopeId string) *policyService.Error {
	return nil
}

func (smc *SimpleMockClient) GetGvkPolicyOfProject(ctx context.Context, projectId string, workloadType policyService.GVK) (*policyService.GetGvkPolicyResp, *policyService.Error) {
	group, kind, version := gvkToKey(workloadType)
	result, ok := smc.GvkPolicyOfProject[ProjectAndType{
		Group:     group,
		Kind:      kind,
		Version:   version,
		ProjectId: projectId,
	}]
	if !ok {
		return nil, policyService.NewError(http.StatusNotFound, "no policy found")
	}

	return result, nil
}

func (smc *SimpleMockClient) WithGvkPolicy(projectId string, workloadType policyService.GVK, policy *policyService.GetGvkPolicyResp) *SimpleMockClient {
	if smc.GvkPolicyOfProject == nil {
		smc.GvkPolicyOfProject = make(map[ProjectAndType]*policyService.GetGvkPolicyResp)
	}

	group, kind, version := gvkToKey(workloadType)
	smc.GvkPolicyOfProject[ProjectAndType{
		Group:     group,
		Kind:      kind,
		Version:   version,
		ProjectId: projectId,
	}] = policy

	return smc
}

type SimpleMockClientFactory struct {
	SimpleMockClient
}

func (smc SimpleMockClientFactory) Client() ClientInterface {
	return &smc.SimpleMockClient
}
