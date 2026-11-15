/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tree

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

// calculateEffectivePolicy normalizes sources once before calculating inheritance.
func (ht *HierarchyTree) calculateEffectivePolicy(ctx context.Context, tx orm.DB,
	appliedPolicyId string, idxToCalculate int,
	imposedAssets map[string]interface{}, masterImposedAssets map[string]interface{},
	imposedAssetsRules map[string]interface{}, masterImposedAssetsRules map[string]interface{}) (
	*dbmodel.EffectivePolicy, *api.Error,
) {
	effectiveContext := ctx.Value(EffectiveContextKey).(*EffectiveContext)

	if !effectiveContext.GvkWorkload {
		for _, policy := range effectiveContext.PolicyById {
			if err := policy.NormalizeGPUMemory(); err != nil {
				return nil, api.NewValidationError(err.Error())
			}
		}
	}

	return ht.calculateEffectivePolicyAt(ctx, tx, appliedPolicyId, idxToCalculate,
		imposedAssets, masterImposedAssets, imposedAssetsRules, masterImposedAssetsRules)
}

func normalizeGPUPolicyRequest(request *ApplyPolicyRequest) *api.Error {
	var err error
	request.Defaults, err = gpucompat.NormalizeLegacyGPUFields(request.Defaults)
	if err != nil {
		return api.NewValidationError(err.Error())
	}
	request.Rules, err = gpucompat.NormalizeLegacyGPUPolicy(request.Rules)
	if err != nil {
		return api.NewValidationError(err.Error())
	}
	fields := make([]string, len(request.FieldsToReset))
	for i, field := range request.FieldsToReset {
		fields[i] = gpucompat.CanonicalGPUFieldPath(field)
	}
	request.FieldsToReset = fields
	return nil
}
