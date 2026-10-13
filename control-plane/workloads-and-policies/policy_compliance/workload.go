/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy_compliance

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils"
)

var workloadJsonCracker = jsonutils.NewJsonCracker(workloadsApi.DistributedWorkload{})
var masterSpecField = workloadJsonCracker.NameOf(&workloadJsonCracker.Struct().MasterSpec)

// ValidateWorkloadCompliance validates a spec after source defaults have been merged.
// It normalizes aliases and resolves derived fields in mutatedSpec, returning a copy
// of normalized policy defaults.
// Distributed training validates worker and master policies separately.
func ValidateWorkloadCompliance(rules, defaults, mutatedSpec map[string]interface{}, distributed bool) (map[string]interface{}, ComplianceErrors, error) {
	var aliasErr error
	rules, aliasErr = gpucompat.NormalizeLegacyGPUPolicy(rules)
	if aliasErr != nil {
		return nil, nil, aliasErr
	}
	defaults, aliasErr = gpucompat.NormalizeLegacyGPUFields(defaults)
	if aliasErr != nil {
		return nil, nil, aliasErr
	}
	canonical, aliasErr := gpucompat.NormalizeLegacyGPUFields(mutatedSpec)
	if aliasErr != nil {
		return nil, nil, aliasErr
	}
	for key := range mutatedSpec {
		delete(mutatedSpec, key)
	}
	for key, value := range canonical {
		mutatedSpec[key] = value
	}

	// Alias normalization can reveal dependencies or cycles that used different spellings.
	workerRules := rules
	if distributed {
		workerRules = convutils.MapMe(rules[workerDefaultsSection])
	}
	workerOrder, orderErr := NewDefaultsFromOrder(workerRules)
	if orderErr != nil {
		return nil, nil, orderErr
	}
	masterOrder, orderErr := NewDefaultsFromOrder(convutils.MapMe(rules[masterDefaultsSection]))
	if orderErr != nil {
		return nil, nil, orderErr
	}

	fieldsWithCantEditRule := make([]string, 0)
	complianceErrors := make(ComplianceErrors, 0)

	if !distributed {
		ValidateCompliance("", rules, defaults, mutatedSpec, workerOrder,
			&fieldsWithCantEditRule, nil, &complianceErrors)
	} else {
		// check the workers spec VS the workers policy
		ValidateCompliance("",
			convutils.MapMe(rules[workerDefaultsSection]),
			convutils.MapMe(defaults[workerDefaultsSection]),
			convutils.MapMe(mutatedSpec[workerDefaultsSection]),
			workerOrder,
			&fieldsWithCantEditRule, nil, &complianceErrors)

		// if there is a master policy, we need to validate the master spec VS the master policy
		// no need to validate master if no master is provided
		if rules[masterDefaultsSection] != nil && mutatedSpec[masterDefaultsSection] != nil {
			masterComplianceErrors := make(ComplianceErrors, 0)

			ValidateCompliance("",
				convutils.MapMe(rules[masterDefaultsSection]),
				convutils.MapMe(defaults[masterDefaultsSection]),
				convutils.MapMe(mutatedSpec[masterDefaultsSection]),
				masterOrder,
				&fieldsWithCantEditRule, nil, &masterComplianceErrors)

			for _, err := range masterComplianceErrors {
				complianceErrors.Add(err.Rule, err.Details,
					convutils.DotUs(masterSpecField, err.WorkloadField),
					convutils.DotUs(masterSpecField, err.PolicyField))
			}
		}
	}

	return defaults, complianceErrors, nil
}
