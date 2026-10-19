/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package compliance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v1/runai_workloads"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	dinfwv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	policyCompliance "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/policy_compliance"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/validation"
)

const (
	dummyName = "policy-to-validate"

	workloadTypeDistributed          = "Distributed"
	workloadTypeDistributedInference = "DistributedInference"

	DefaultKeyAttribute = "name"

	// InstanceKeyForTheDefaults is a reserved key that the workload controller uses to identify
	// instance which is used to provide defaults for attributes.
	InstanceKeyForTheDefaults = "DEFAULTS"
)

var (
	distributedJsonCracker = jsonutils.NewJsonCracker(api.DistributedPolicyDefaultsV2{})
	masterReplica          = distributedJsonCracker.NameOf(&distributedJsonCracker.Struct().Master)
	workerReplica          = distributedJsonCracker.NameOf(&distributedJsonCracker.Struct().Worker)
)

func ValidateDefaults(appliedPolicy *dbmodel.PolicyWithOptionalEffective) ([]string, policyCompliance.ComplianceErrors) {
	complianceErrors := make(policyCompliance.ComplianceErrors, 0)
	fieldsWithCantEditRule := make([]string, 0)

	policyCompliance.ValidateCompliance("",
		appliedPolicy.EffectivePolicy.RulesV2,
		nil,
		appliedPolicy.EffectivePolicy.DefaultsV2,
		policyCompliance.DefaultFromOrder{}, // defaultFrom only validated on workloads, not on policies
		&fieldsWithCantEditRule,
		&policyCompliance.ValidateComplianceOptions{
			PolicySelfValidation:           true,
			IgnoreRequiredFields:           true,
			CanEditWithNoDefaultValidation: true,
		},
		&complianceErrors)
	if len(complianceErrors) != 0 {
		return nil, complianceErrors
	}

	return fieldsWithCantEditRule, nil
}

func ValidateWorkload(policy *dbmodel.EffectivePolicy, specValues map[string]interface{}) policyCompliance.ComplianceErrors {
	complianceErrors := make(policyCompliance.ComplianceErrors, 0)
	// if there are no rules - nothing to check
	if len(policy.RulesV2) == 0 {
		return complianceErrors
	}

	fieldsWithCantEditRule := make([]string, 0)

	policyCompliance.ValidateCompliance("",
		policy.RulesV2,
		policy.DefaultsV2,
		specValues,
		policyCompliance.DefaultFromOrder{}, // defaultFrom only validated on workloads, not on policies
		&fieldsWithCantEditRule,
		nil,
		&complianceErrors)

	return complianceErrors
}

func ValidateSubmittedPolicy(policyRules map[string]interface{}, policyDefaults map[string]interface{}) policyCompliance.ComplianceErrors {
	complianceErrors := make(policyCompliance.ComplianceErrors, 0)

	// do some minimal validation on the defaults, specifically - on itemized fields
	// the complete validation is done only on the effective.
	policyCompliance.ValidateGivenPolicyDefaults(policyDefaults, &complianceErrors)

	// validate consistency of the rules
	validateRulesConsistency(policyRules, &complianceErrors)

	return complianceErrors
}

func validateRulesConsistency(policyRules map[string]interface{}, complianceErrors *policyCompliance.ComplianceErrors) {
	submittedRules := convutils.MapMe(policyRules)
	if len(submittedRules) == 0 {
		return
	}

	for fieldName, fieldRules := range policyCompliance.GetRulesForSourceOfRules(policyRules) {
		constraintsMap := convutils.MapMe(fieldRules)

		minPolicy := constraintsMap[string(policyCompliance.PolicyMin)]
		maxPolicy := constraintsMap[string(policyCompliance.PolicyMax)]
		stepPolicy := constraintsMap[string(policyCompliance.PolicyStep)]
		blocked := constraintsMap[string(policyCompliance.PolicyBlocked)]
		canEdit := constraintsMap[string(policyCompliance.PolicyCanEdit)]
		options := constraintsMap[string(policyCompliance.PolicyOptions)]
		required := constraintsMap[string(policyCompliance.PolicyRequired)]

		if minPolicy != nil && maxPolicy != nil {
			if _, err := validation.ToFloat(maxPolicy); err != nil {
				complianceErrors.Add(&policyCompliance.PolicyMax, "invalid value of max field", fieldName, fieldName)
				continue
			}

			if _, err := validation.ToFloat(minPolicy); err != nil {
				complianceErrors.Add(&policyCompliance.PolicyMin, "invalid value of min field", fieldName, fieldName)
				continue
			}
			// Validate that minPolicy is not greater than maxPolicy
			validationResult := validation.Validate(minPolicy, validation.Max(maxPolicy))
			_, err := policyCompliance.IsOk(validationResult)
			if err != nil {
				complianceErrors.Add(&policyCompliance.PolicyMax, "max value must be no less than min value", fieldName, fieldName)
			}
		}

		if stepPolicy != nil && maxPolicy != nil {
			_, err := policyCompliance.IsOk(validation.Validate(stepPolicy, validation.Max(maxPolicy)))
			if err != nil {
				complianceErrors.Add(&policyCompliance.PolicyStep, "step value must not exceed max value", fieldName, fieldName)
			}
			if minPolicy != nil {

				maxVal, maxErr := validation.ToFloat(maxPolicy)
				minVal, minErr := validation.ToFloat(minPolicy)

				if maxErr == nil && minErr == nil {
					minMaxRange := maxVal - minVal
					_, err := policyCompliance.IsOk(validation.Validate(stepPolicy, validation.Max(minMaxRange)))
					if err != nil {
						complianceErrors.Add(&policyCompliance.PolicyStep, "step value must not exceed the calculated range between min and max values", fieldName, fieldName)
					}
				}
			}
		}

		if blocked != nil && blocked.(bool) {
			if canEdit != nil && !canEdit.(bool) {
				complianceErrors.Add(&policyCompliance.PolicyStep, "specify either blocked=true or canEdit=false, not both", fieldName, fieldName)
			}
			if required != nil && required.(bool) {
				complianceErrors.Add(&policyCompliance.PolicyStep, "specify either blocked=true or required=true, not both", fieldName, fieldName)
			}
			if options != nil {
				complianceErrors.Add(&policyCompliance.PolicyStep, "specify either blocked=true or options, not both", fieldName, fieldName)
			}
		}
	}
}

// ValidateEffectivePolicy validates that there are no issues translating the policy to the cluster CRD
func ValidateEffectivePolicy(ctx context.Context, workloadType string, _ map[string]interface{}, policyDefaults map[string]interface{}) *workloadsApi.Error {
	options := composer.OptionsForPolicy()

	legacyDefaults, err := gpucompat.LegacyGPUFields(policyDefaults)
	if err != nil {
		return workloadsApi.NewError(http.StatusBadRequest, err.Error())
	}
	rearrangedPolicy := utils.MapMe(FlattenInstances(legacyDefaults))

	switch workloadType {
	case workloadTypeDistributedInference:
		var defaultsSpec workloadsApi.SupersetSpec

		jsonErr := json.Unmarshal(gojsonq.ToBytes(rearrangedPolicy), &defaultsSpec)
		if jsonErr != nil {
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to load defaults: %v", jsonErr))
		}

		dInfComposer := runai_workloads.NewDinfWorkloadComposer(
			dummyName, "dummyUser", nil, options)

		composerErr := dInfComposer.Compose(ctx, &defaultsSpec, &dinfwv1.DistributedInferenceWorkload{})
		if composerErr != nil {
			return workloadsApi.NewError(composerErr.Code, composerErr.Message)
		}

	case workloadTypeDistributed:
		var defaultsWorkload workloadsApi.SupersetWorkload

		jsonErr := json.Unmarshal(gojsonq.ToBytes(utils.MapMe(rearrangedPolicy[workerReplica])), &defaultsWorkload.Spec)
		if jsonErr != nil {
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to load worker defaults: %v", jsonErr))
		}

		masterDefaults, masterDefaultsExist := rearrangedPolicy[masterReplica]
		if masterDefaultsExist {
			defaultsWorkload.SetMasterSpec(workloadsApi.MasterSpec{})
			jsonErr := json.Unmarshal(gojsonq.ToBytes(utils.MapMe(masterDefaults)), defaultsWorkload.MasterSpec.Get())
			if jsonErr != nil {
				return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to load master defaults: %v", jsonErr))
			}
		}

		distributedComposer := runai_workloads.NewDistributedWorkloadComposer(
			dummyName, "dummyUser", nil, options)

		composerErr := distributedComposer.Compose(ctx, &defaultsWorkload, &jobs.JobWorkload{})
		if composerErr != nil {
			return workloadsApi.NewError(composerErr.Code, composerErr.Message)
		}

	default:
		var defaultsSpec workloadsApi.SupersetSpec

		jsonErr := json.Unmarshal(gojsonq.ToBytes(rearrangedPolicy), &defaultsSpec)
		if jsonErr != nil {
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to load defaults: %v", jsonErr))
		}

		jobSpecComposer := runai_workloads.NewSupersetToJobSpecComposer(
			workloadType, dummyName, "dummyUser", nil, nil, options)

		composerErr := jobSpecComposer.Compose(ctx, &defaultsSpec, &jobs.JobSpec{})()
		if composerErr != nil {
			return workloadsApi.NewError(composerErr.Code, composerErr.Message)
		}
	}

	return nil
}

func FlattenInstances(v interface{}) interface{} {
	switch node := v.(type) {

	case map[string]interface{}:
		// Case 1: has instances -> replace node
		if inst, ok := node["instances"]; ok {
			if arr, ok := inst.([]interface{}); ok {
				out := make([]interface{}, 0, len(arr))
				for _, item := range arr {
					out = append(out, FlattenInstances(item))
				}
				return out
			}
		}

		// Case 2: map has ONLY "attributes" -> null
		if len(node) == 1 {
			if _, ok := node["attributes"]; ok {
				return nil
			}
		}

		// Normal recursive descent
		out := make(map[string]interface{}, len(node))
		for k, v := range node {
			out[k] = FlattenInstances(v)
		}
		return out

	case []interface{}:
		out := make([]interface{}, 0, len(node))
		for _, item := range node {
			out = append(out, FlattenInstances(item))
		}
		return out

	default:
		return v
	}
}
