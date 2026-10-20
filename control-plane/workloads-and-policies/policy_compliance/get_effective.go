/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy_compliance

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	policyApi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

var nonCreationMode = policies.PolicyClientOptions{
	NonCreationMode: true,
}

var (
	policyJsonCracker = jsonutils.NewJsonCracker(policyApi.SupersetPolicy{})
	defaultsSection   = policyJsonCracker.NameOf(&policyJsonCracker.Struct().Defaults)
	rulesSection      = policyJsonCracker.NameOf(&policyJsonCracker.Struct().Rules)
)

type EffectivePolicyOfProject struct {
	DefaultsV2 dbmodels.JsonBMap
	RulesV2    dbmodels.JsonBMap

	DefaultFromOrder policyApi.DefaultFromOrder

	UpdatedAt time.Time
}

func NewEffectivePolicyOfProject() *EffectivePolicyOfProject {
	return &EffectivePolicyOfProject{
		DefaultsV2: make(map[string]interface{}),
		RulesV2:    make(map[string]interface{}),
	}
}

func GetEffectivePolicyOfProject(ctx context.Context, policyClient policies.ClientInterface, projectId string, workloadType workloadsApi.WorkloadType) (*EffectivePolicyOfProject, *policyApi.Error) {
	projectIdNumeric, err := strconv.Atoi(projectId)
	if err != nil {
		return nil, policyApi.NewError(http.StatusBadRequest, "invalid project id")
	}

	var effectivePolicyAsJson map[string]interface{}
	var policyUpdateTime *time.Time
	var jsonError error
	var defaultFromOrder *policyApi.DefaultFromOrder

	// in case policy not found, report it as empty policy
	emptyPolicy := NewEffectivePolicyOfProject()
	emptyPolicy.DefaultFromOrder = policyApi.DefaultFromOrder{
		Worker: &map[string][]string{},
		Master: &map[string][]string{},
	}
	emptyPolicy.UpdatedAt = time.Now()

	switch workloadType {
	case workloadsApi.WORKLOADTYPE_TRAINING:
		response, err := policyClient.GetTrainingPolicyOfProjectV2(ctx, int32(projectIdNumeric), &nonCreationMode)
		if err != nil {
			if err.Code != http.StatusNotFound {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get training policy of project %s: %s (%v)", projectId, err.Message, err.Details)
				return nil, policyApi.NewError(err.Code, err.Message)
			}
			return emptyPolicy, nil
		}

		effectivePolicyAsJson, jsonError = gojsonq.LoadJson(response.Effective.Get())
		policyUpdateTime = response.EffectiveUpdatedAt

		if response.Effective.Get() != nil {
			defaultFromOrder = response.Effective.Get().DefaultFromOrder.Get()
		}

	case workloadsApi.WORKLOADTYPE_WORKSPACE:
		response, err := policyClient.GetWorkspacePolicyOfProjectV2(ctx, int32(projectIdNumeric), &nonCreationMode)
		if err != nil {
			if err.Code != http.StatusNotFound {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get workspace policy of project %s: %s (%v)", projectId, err.Message, err.Details)
				return nil, policyApi.NewError(err.Code, err.Message)
			}
			return emptyPolicy, nil
		}

		effectivePolicyAsJson, jsonError = gojsonq.LoadJson(response.Effective.Get())
		policyUpdateTime = response.EffectiveUpdatedAt

		if response.Effective.Get() != nil {
			defaultFromOrder = response.Effective.Get().DefaultFromOrder.Get()
		}

	case workloadsApi.WORKLOADTYPE_DISTRIBUTED:
		response, err := policyClient.GetDistributedPolicyOfProjectV2(ctx, int32(projectIdNumeric), &nonCreationMode)
		if err != nil {
			if err.Code != http.StatusNotFound {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get distributed policy of project %s: %s (%v)", projectId, err.Message, err.Details)
				return nil, policyApi.NewError(err.Code, err.Message)
			}
			return emptyPolicy, nil
		}

		effectivePolicyAsJson, jsonError = gojsonq.LoadJson(response.Effective.Get())
		policyUpdateTime = response.EffectiveUpdatedAt

		if response.Effective.Get() != nil {
			defaultFromOrder = response.Effective.Get().DefaultFromOrder.Get()
		}

	case workloadsApi.WORKLOADTYPE_INFERENCE:
		response, err := policyClient.GetInferencePolicyOfProjectV2(ctx, int32(projectIdNumeric), &nonCreationMode)
		if err != nil {
			if err.Code != http.StatusNotFound {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get inference policy of project %s: %s (%v)", projectId, err.Message, err.Details)
				return nil, policyApi.NewError(err.Code, err.Message)
			}
			return emptyPolicy, nil
		}

		effectivePolicyAsJson, jsonError = gojsonq.LoadJson(response.Effective.Get())
		policyUpdateTime = response.EffectiveUpdatedAt

		if response.Effective.Get() != nil {
			defaultFromOrder = response.Effective.Get().DefaultFromOrder.Get()
		}

	case workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE:
		response, err := policyClient.GetDistributedInferencePolicyOfProjectV2(ctx, int32(projectIdNumeric), &nonCreationMode)
		if err != nil {
			if err.Code != http.StatusNotFound {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get distributed inference policy of project %s: %s (%v)", projectId, err.Message, err.Details)
				return nil, policyApi.NewError(err.Code, err.Message)
			}
			return emptyPolicy, nil
		}

		effectivePolicyAsJson, jsonError = gojsonq.LoadJson(response.Effective.Get())
		policyUpdateTime = response.EffectiveUpdatedAt

		if response.Effective.Get() != nil {
			defaultFromOrder = response.Effective.Get().DefaultFromOrder.Get()
		}

	default:
		return nil, policyApi.NewError(http.StatusInternalServerError, fmt.Sprintf("unsupported workload type %s", workloadType))
	}

	if jsonError != nil {
		return nil, policyApi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to load effective %s policy of the project: %v", workloadType, jsonError))
	}

	if defaultFromOrder == nil {
		defaultFromOrder = &emptyPolicy.DefaultFromOrder
	}

	effectivePolicy := EffectivePolicyOfProject{
		DefaultsV2: convutils.MapMe(effectivePolicyAsJson[defaultsSection]),
		RulesV2:    convutils.MapMe(effectivePolicyAsJson[rulesSection]),

		DefaultFromOrder: *defaultFromOrder,
	}

	var aliasErr error
	effectivePolicy.DefaultsV2, aliasErr = gpucompat.NormalizeLegacyGPUFields(effectivePolicy.DefaultsV2)
	if aliasErr != nil {
		return nil, policyApi.NewError(http.StatusBadRequest, aliasErr.Error())
	}
	effectivePolicy.RulesV2, aliasErr = gpucompat.NormalizeLegacyGPUPolicy(effectivePolicy.RulesV2)
	if aliasErr != nil {
		return nil, policyApi.NewError(http.StatusBadRequest, aliasErr.Error())
	}
	if defaultFromOrder.Worker != nil {
		order := gpucompat.CanonicalGPUDefaultFromOrder(*defaultFromOrder.Worker)
		effectivePolicy.DefaultFromOrder.Worker = &order
	}
	if defaultFromOrder.Master != nil {
		order := gpucompat.CanonicalGPUDefaultFromOrder(*defaultFromOrder.Master)
		effectivePolicy.DefaultFromOrder.Master = &order
	}

	if policyUpdateTime != nil {
		effectivePolicy.UpdatedAt = *policyUpdateTime
	} else {
		effectivePolicy.UpdatedAt = time.Now()
	}

	return &effectivePolicy, nil
}
