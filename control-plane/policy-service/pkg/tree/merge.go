/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tree

import (
	"reflect"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/itemized"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/scopes"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	policyCompliance "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/policy_compliance"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
)

const isFallbackField = "isFallback"

// MergeRules calculates current effective based on inherit effective and current given rules
// it is also used to merge effective with fallback effective, in this case the givenRules are the
// rules of the fallback policy
func (ht *HierarchyTree) MergeRules(effectiveRules dbmodels.JsonBMap, givenRules dbmodels.JsonBMap,
	rulesFromParent map[string]bool, currentSourceOfTrue map[string]interface{},
) (map[string]interface{}, error) {
	//
	//   no given rules - the result already contains the effective of the parent so we are done
	//
	if len(givenRules) == 0 {
		return effectiveRules, nil
	}

	//
	//  set the sourceOfRules for all the rules of the given policy to the current scope
	//
	rulesOfGivenPolicy := givenRules.Clone()
	ht.InitSourceOfRules(rulesOfGivenPolicy, currentSourceOfTrue, false)

	//
	//   if there are no rules from the parent, just copy the given rules to the result
	//
	if len(effectiveRules) == 0 {
		return database.CleanMap(rulesOfGivenPolicy), nil
	}

	//
	//   if there are rules from the parent, need to merge the two based on the rulesFromParent map
	//
	return effectiveRules, mergeRules("", effectiveRules, rulesOfGivenPolicy, rulesFromParent)
}

// toMap containing effective policy from parent
// fromMap containing given policy
// --> toMap should contain the merged rules b/w then
func mergeRules(fieldPrefix string, toMap dbmodels.JsonBMap, fromMap dbmodels.JsonBMap, rulesFromParent map[string]bool) error {
	for field, rulesFromGivenPolicy := range fromMap {
		//
		//   parent does not have rules for this field so just use the rules from the given policy
		//
		if _, parentHasRulesForThisField := toMap[field]; !parentHasRulesForThisField {
			toMap[field] = rulesFromGivenPolicy
			continue
		}

		//
		//   compose the full name of the field. we do not include the instances/attributes in the field name
		//   because this is how they appear in the rulesFromParent map
		//
		fullFieldName := fieldPrefix

		if field != itemized.AttributesSuffix[1:] && field != itemized.InstancesSuffix[1:] {
			if fullFieldName != "" {
				fullFieldName += "." + field
			} else {
				fullFieldName = field
			}
		}

		//
		//   if this rules it current-override-parent then copy the rule from the current policy to the
		//   toMap result
		//
		sourceOfRulesFromGivenPolicy := convutils.MapMe(rulesFromGivenPolicy)[policyCompliance.PolicySourceOfRule]
		if sourceOfRulesFromGivenPolicy != nil {
			fromParent := rulesFromParent[fullFieldName]
			if !fromParent {
				toMap[field] = rulesFromGivenPolicy
			} else {
				//
				//  since we are merging into the fallback policy, isFallback=true for all the fields.
				//  when we have a field in the given policy, and the given policy itself is not fallback,
				//  we have to clear the isFallback from the target to indicate that the value was driven
				//  from a non-fallback policy
				//
				fromFallback, foundIsFallback := convutils.MapMe(sourceOfRulesFromGivenPolicy)[isFallbackField]
				if !foundIsFallback || !fromFallback.(bool) {
					sourceOfRulesToPolicy := convutils.MapMe(convutils.MapMe(toMap[field])[policyCompliance.PolicySourceOfRule])
					delete(sourceOfRulesToPolicy, isFallbackField)
				}
			}
			continue
		}

		err := mergeRules(fullFieldName, convutils.MapMe(toMap[field]), convutils.MapMe(fromMap[field]), rulesFromParent)
		if err != nil {
			return err
		}
	}

	return nil
}

type MergeDefaultsOptions struct {
	mergingWithFallback bool
}

// MergeDefaults calculates current effective defaults based on inherit effective and current given rules
// it is also used to merge effective with fallback effective, in this case the givenDefaults are the
// rules of the fallback policy
func (ht *HierarchyTree) MergeDefaults(effectiveDefaults dbmodels.JsonBMap, givenDefaults dbmodels.JsonBMap,
	effectiveRules dbmodels.JsonBMap, currentOrgUnit scopes.OrgUnit,
	options MergeDefaultsOptions,
) (map[string]interface{}, error) {
	//
	//   no given defaults - the result already contains the effective defaults of the parent so we are done
	//
	if len(givenDefaults) == 0 {
		return effectiveDefaults, nil
	}

	//
	//   if there are no defaults from the parent, just copy the given defaults to the result
	//
	if len(effectiveDefaults) == 0 {
		return database.CleanMap(givenDefaults), nil
	}

	policyCompliance.PrepareDefaultsForMerge(effectiveDefaults, nil)

	err := mergeDefaults("", effectiveDefaults, givenDefaults, effectiveRules, currentOrgUnit, options)
	if err != nil {
		return nil, err
	}

	policyCompliance.RemoveFieldNameInstance(effectiveDefaults)
	return effectiveDefaults, nil
}

// toMap containing effective policy from parent
// fromMap containing given policy
// --> toMap should contain the merged defaults b/w then
func mergeDefaults(fieldPrefix string, toMap dbmodels.JsonBMap, fromMap dbmodels.JsonBMap,
	effectiveRules dbmodels.JsonBMap, currentOrgUnit scopes.OrgUnit, options MergeDefaultsOptions,
) error {
	for field, defaultsFromGivenPolicy := range fromMap {
		//
		//  no parent defaults just take the defaults from the given policy as is
		//
		if _, parentHasDefaultsForThisField := toMap[field]; !parentHasDefaultsForThisField {
			toMap[field] = defaultsFromGivenPolicy
			continue
		}

		fullFieldName := fieldPrefix

		if fullFieldName != "" {
			fullFieldName += "." + field
		} else {
			fullFieldName = field
		}

		//
		//   if the effective policy has canEdit rule = false, which we inherit from the parent,
		//   then we also leave the default from the parent. this only applies when we merge
		//   policy with parent, it does not relevant for merging specific workload type with fallback
		//
		if !options.mergingWithFallback {
			rules := convutils.MapMe(effectiveRules[field])
			if rules[string(policyCompliance.PolicyCanEdit)] != nil && rules[string(policyCompliance.PolicyCanEdit)] == false {
				sourceOfRulesThisField := convutils.MapMe(rules[policyCompliance.PolicySourceOfRule])
				orgUnitOfThisRule := scopes.NewFromJson(sourceOfRulesThisField)
				if !orgUnitOfThisRule.IsSameScopeAs(currentOrgUnit) { // which means that we inherit this rule
					continue
				}
			}
		}

		switch defaultsFromGivenPolicy.(type) {
		case []interface{}:
			// itemized fields are handled using the slices-merge function which knows how to merge
			// two itemized fields according to the keys of the items
			resultValue, err := policyCompliance.WorkloadSlicesMergeFunc(reflect.ValueOf(toMap[field]), reflect.ValueOf(fromMap[field]), nil)
			if err != nil {
				return err
			}
			toMap[field] = resultValue.Interface()
		case map[string]interface{}:
			// for compound fields we recursively call the mergeDefaults to merge their own fields
			// for example compute section we need to merge the fields in the compute section
			err := mergeDefaults(fullFieldName, convutils.MapMe(toMap[field]), convutils.MapMe(fromMap[field]),
				convutils.MapMe(effectiveRules[field]), currentOrgUnit, options)
			if err != nil {
				return err
			}
		default:
			toMap[field] = defaultsFromGivenPolicy
		}
	}

	return nil
}
