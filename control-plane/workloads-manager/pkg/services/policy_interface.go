/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	policiesClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	policyCompliance "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/policy_compliance"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	policyApi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const WorkloadTypeNimService = "NimService"

// if tenant is not working with policies, do not attempt to create them
var nonCreationMode = policiesClient.PolicyClientOptions{
	NonCreationMode: true,
}

func validateMutatedSpecVsPolicy(rules map[string]interface{}, defaults map[string]interface{},
	mutatedSpec map[string]interface{},
) (map[string]interface{}, *api.SubmissionErrorV2) {
	defaults, complianceErrors, err := policyCompliance.ValidateWorkloadCompliance(rules, defaults, mutatedSpec, WorkloadTypeNimService)
	if err != nil {
		return nil, NewSubmissionError(http.StatusBadRequest, err.Error(), err)
	}

	if len(complianceErrors) != 0 {
		issues := complianceErrors.ForWorkloadsV2()

		complianceErr := api.SubmissionErrorV2{
			Code:             http.StatusBadRequest,
			Message:          "The workload request does not comply with the policy set by the administrator",
			ComplianceIssues: make([]api.ComplianceIssuesV2ComplianceIssuesInner, 0, len(issues)),
		}

		for _, issue := range issues {
			err := api.ComplianceIssuesV2ComplianceIssuesInner{
				Field:   issue.Field,
				Details: issue.Details,
			}

			if convutils.HasValue(issue.Rule) {
				err.Rule = convutils.Ref(api.PolicyRuleEnum(*issue.Rule.Get()))
			}

			complianceErr.ComplianceIssues = append(complianceErr.ComplianceIssues, err)
		}

		return nil, &complianceErr
	}

	return defaults, nil
}

func GetPolicyOfProject(ctx context.Context, workloadType string, projectId string) (map[string]interface{}, map[string]interface{}, *policyApi.DefaultFromOrder, *api.Error) {
	//
	// currently project ID in policy service is still numeric, this needs to be changed to string in separate PR
	//
	projectIdInt, err := strconv.Atoi(projectId)
	if err != nil {
		return nil, nil, nil, api.NewError(http.StatusBadRequest, "project id is not numeric", nil)
	}

	rules := make(map[string]interface{})
	defaults := make(map[string]interface{})

	defaultFromOrder := &policyApi.DefaultFromOrder{
		Master: &map[string][]string{},
		Worker: &map[string][]string{},
	}

	switch workloadType {
	case WorkloadTypeNimService:
		nimServicePolicy, policyErr := policiesClient.Get().Client().GetNimServicePolicyOfProject(ctx, int32(projectIdInt), &nonCreationMode)
		if policyErr != nil && policyErr.Code != http.StatusNotFound {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get %s policy %s: %v", workloadType, policyErr.Message, policyErr.Details)
			return nil, nil, nil, api.NewError(int(policyErr.Code), policyErr.Message, nil)
		}

		if nimServicePolicy != nil && nimServicePolicy.Effective.IsSet() {
			var loadErr error

			effectivePolicy := nimServicePolicy.Effective.Get()
			if effectivePolicy.Rules.IsSet() {
				rules, loadErr = gojsonq.LoadJson(effectivePolicy.Rules.Get())
				if loadErr != nil {
					return nil, nil, nil, api.NewError(http.StatusServiceUnavailable, "failed to load rules", loadErr)
				}
			}
			if effectivePolicy.Defaults.IsSet() {
				defaults, loadErr = gojsonq.LoadJson(effectivePolicy.Defaults.Get())
				if loadErr != nil {
					return nil, nil, nil, api.NewError(http.StatusServiceUnavailable, "failed to load defaults", loadErr)
				}
			}

			if effectivePolicy.DefaultFromOrder.IsSet() {
				defaultFromOrder = effectivePolicy.DefaultFromOrder.Get()
			}
		}

	default:
		return nil, nil, nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf("unsupported policy type: %s", workloadType), nil)
	}

	return rules, defaults, defaultFromOrder, nil
}

type WorkloadV2Spec interface {
	api.NimServiceSpec
}

// MutateWorkload applies policy defaults to a workload spec.
//
// Parameters:
//   - policyDefaults: OUTPUT parameter - the caller's map that will be filled with the final mutated spec.
//     This map may initially contain policy defaults (if they exist) which get merged with the workload.
//
// Returns:
//   - mutated spec as map
//   - error: SubmissionError if mutation fails
func MutateWorkload[S WorkloadV2Spec](
	inputSpec *S,
	templateSpec map[string]interface{},
	policyDefaults map[string]interface{},
) (map[string]interface{}, *api.SubmissionErrorV2) {

	var inputSpecAsMap map[string]interface{}
	if inputSpec == nil {
		inputSpecAsMap = make(map[string]interface{})
	} else {
		// Use gojsonq.LoadJson (not CopySpecs) to preserve numeric types via decoder.UseNumber()
		inputSpecAsMap, _ = gojsonq.LoadJson(inputSpec)
	}

	if err := gpucompat.ValidateMemoryRequestInput(inputSpecAsMap); err != nil {
		return nil, NewSubmissionError(http.StatusBadRequest, err.Error(), err)
	}

	for _, source := range []*map[string]interface{}{&inputSpecAsMap, &templateSpec, &policyDefaults} {
		canonical, err := gpucompat.NormalizeLegacyGPUFields(*source)
		if err != nil {
			return nil, NewSubmissionError(http.StatusBadRequest, err.Error(), err)
		}
		*source = canonical
	}

	// Phase 1: Template = Spec -> Template
	if len(templateSpec) == 0 {
		templateSpec = inputSpecAsMap
	} else {
		// Merge user input with template: inputSpec gets merged INTO templateSpec
		policyCompliance.PrepareTemplateForMutation(templateSpec)

		err := policyCompliance.ApplyDefaults(inputSpecAsMap, templateSpec, map[policyCompliance.PerReplicaField]interface{}{})
		if err != nil {
			return nil, NewSubmissionError(int(err.Code), err.Message, nil)
		}
	}

	// Phase 2: Mutated = (Spec -> Template) -> Policy
	var mutatedSpec map[string]interface{}

	if len(policyDefaults) == 0 {
		mutatedSpec = templateSpec
	} else {
		// Use gojsonq.LoadJson (not CopySpecs) to preserve numeric types via decoder.UseNumber()
		mutatedSpec, _ = gojsonq.LoadJson(policyDefaults)

		attributesDefaults := policyCompliance.PrepareDefaultsForMutation(mutatedSpec)

		err := policyCompliance.ApplyDefaults(templateSpec, mutatedSpec, attributesDefaults)
		if err != nil {
			return nil, NewSubmissionError(int(err.Code), err.Message, nil)
		}
	}

	cleanEmptyArraysMap(mutatedSpec)

	return mutatedSpec, nil
}

func cleanEmptyArraysMap(aMap map[string]interface{}) {
	for key, value := range aMap {
		if arrayInMap, ok := value.([]interface{}); ok {
			if len(arrayInMap) == 0 {
				delete(aMap, key)
				continue
			}
		}

		//  handle only entries which are maps
		if nestedMap, ok := value.(map[string]interface{}); ok {
			// recursively cleans any nested map
			cleanEmptyArraysMap(nestedMap)

			// remove nil entries
			for innerKey, innerValue := range nestedMap {
				if innerValue == nil {
					delete(nestedMap, innerKey)
				}
			}

			// if nested map is empty remove it
			if len(nestedMap) == 0 {
				delete(aMap, key)
			}
		}
	}
}
