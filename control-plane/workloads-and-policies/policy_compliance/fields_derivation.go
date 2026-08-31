/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy_compliance

import (
	"github.com/run-ai/runai/control-plane/workloads-and-policies/itemized"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"golang.org/x/exp/maps"
)

// FieldsWithRules return array of field names based on the rules payload
// that are aligned with the sourceOfRules stamping.
// For example:
//
//	 image:
//	   <-- source of rules (1)
//	   required: true
//		storage:
//		  hostPath:
//		    instances:
//	            <-- source of rules (2)
//		        canEAdd: false
//	     attribute:
//	        path:
//	            <-- source of rules (3)
//	            required: true
//
// The result for this case is:
//   - image (1)
//   - storage.hostPath (2)
//   - storage.hostPath.path (3)
func FieldsWithRules(rulesMap map[string]interface{}) []string {
	return maps.Keys(GetRulesForSourceOfRules(rulesMap))
}

type getRulesOptions struct {
	BreakdownInstances  bool
	BreakdownAttributes bool
}

func GetRulesForSourceOfRules(rulesMap map[string]interface{}) map[string]interface{} {
	fields := make(map[string]interface{}, 0)
	getRulesRecursive(fields, rulesMap, "", &getRulesOptions{
		BreakdownInstances:  true,
		BreakdownAttributes: true,
	})
	return fields
}

// GetRulesForCompliance is similar to GetRulesForSourceOfRules, but it does not break down
// the attributes and instances, because this is done by the validate compliance function
// so the result to the above policy will be
//   - image (with the image related rules)
//   - storage.hostPath (with instances and attributes rules)
func GetRulesForCompliance(rulesMap map[string]interface{}) map[string]interface{} {
	fields := make(map[string]interface{}, 0)
	getRulesRecursive(fields, rulesMap, "", &getRulesOptions{
		BreakdownInstances:  false,
		BreakdownAttributes: false,
	})
	return fields
}

func getRulesRecursive(result map[string]interface{}, rulesMap map[string]interface{}, prefix string, options *getRulesOptions) {
	isItemized := false

	if options.BreakdownInstances && rulesMap[itemized.PolicyInstances] != nil {
		result[prefix] = rulesMap[itemized.PolicyInstances]
		isItemized = true
	}
	if options.BreakdownAttributes && rulesMap[itemized.PolicyAttributes] != nil {
		subTree := convutils.MapMe(rulesMap[itemized.PolicyAttributes])
		getRulesRecursive(result, subTree, prefix, options)
		isItemized = true
	}
	if isItemized {
		return
	}

	// we have to collect all the names (environmentVariables, image, ...) but we have to take into account
	// that there are nested fields, in which the rules will require recursive call. it is sufficient
	// to check the first field in the map, though, because it's either only-rules or only-fields
	ruleFound := false
	for name := range rulesMap {
		ruleFound = RuleNames.Has(name)
		break
	}
	if ruleFound {
		result[prefix] = rulesMap
		return
	}

	for name := range rulesMap {
		subTree := convutils.MapMe(rulesMap[name])
		getRulesRecursive(result, subTree, convutils.DotUs(prefix, name), options)
	}
}

// FieldsOfDefaults is similar to FieldsOfRules, except that the payload is simpler, as there are no rules.
//
//		compute:
//		  gpuPortionRequest: 10
//	      extendedResources:
//	        attributes:  <-- this is not part of the field name
//	          resource:  ...
//
// The result for this case is same as for the rules:
//   - compute.extendedResources
//   - compute.gpuPortionRequest
func FieldsOfDefaults(defaultsMap map[string]interface{}) []string {
	result := make([]string, 0)
	if len(defaultsMap) == 0 {
		return result
	}
	getDefaultsRecursive(&result, defaultsMap, "")
	return result
}

func getDefaultsRecursive(result *[]string, defaultsMap map[string]interface{}, prefix string) {
	if len(defaultsMap) == 0 || defaultsMap[itemized.PolicyInstances] != nil || defaultsMap[itemized.PolicyAttributes] != nil {
		*result = append(*result, prefix)
		return
	}

	for name := range defaultsMap {
		subTree := convutils.MapMe(defaultsMap[name])
		getDefaultsRecursive(result, subTree, convutils.DotUs(prefix, name))
	}
}
