/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tree

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/itemized"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/scopes"
	policyCompliance "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/policy_compliance"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

var (
	computeJsonTracker = jsonutils.NewJsonCracker(api.SupersetDefaultsAllOfCompute{})
	computeFields      = datastructs.NewSet(computeJsonTracker.AllNames()...)
)

type RulesFromParentPerField map[string]bool

type contextField string

const EffectiveContextKey contextField = "calculate-effective-context"

var (
	MasterPrefix = masterSection + "."
	WorkerPrefix = workerSection + "."
)

type FallbackAndRealPolicies struct {
	Policy   *dbmodel.PolicyWithOptionalEffective
	Fallback *dbmodel.PolicyWithOptionalEffective
}

type EffectiveContext struct {
	WorkloadType WorkloadType
	OrgUnits     []scopes.OrgUnit
	PolicyById   map[string]*dbmodel.PolicyWithOptionalEffective
	Hierarchy    []FallbackAndRealPolicies
	GvkWorkload  bool // true if we are calculating effective of GVK workload rather than native workload.
}

func (ec EffectiveContext) findInheritedPolicy(calculatedIndex int) int {
	for curIndex := calculatedIndex + 1; curIndex < len(ec.Hierarchy); curIndex++ {
		policy := ec.Hierarchy[curIndex].Policy
		if policy != nil && !policy.Policy.IsDeleted() && !policy.IsEmpty() {
			return curIndex
		}
	}

	return -1
}

func NewEffectiveContext(ctx context.Context, workloadType WorkloadType, orgUnits []scopes.OrgUnit, policies map[string]*dbmodel.PolicyWithOptionalEffective) *EffectiveContext {
	result := &EffectiveContext{
		WorkloadType: workloadType,
		PolicyById:   policies,
		OrgUnits:     orgUnits,
		Hierarchy:    make([]FallbackAndRealPolicies, len(orgUnits)),
	}

	orgUnitToIndex := make(map[scopes.OrgUnit]int)

	for index, orgUnit := range orgUnits {
		orgUnitToIndex[orgUnit] = index
	}

	for _, policy := range policies {
		orgUnit := policy.OrgUnit()

		hierarchyIndex, found := orgUnitToIndex[orgUnit]
		if !found {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to prepare effexctive context: record for scope %s does not exist", orgUnit.Scope)
			continue
		}

		// if we are calculating effective of fallback itself, then treat fallback policy as "regular" policy
		// otherwise, store it aside as a fallback policy
		if workloadType != WorkloadTypeFallback && policy.WorkloadType == WorkloadTypeFallback {
			result.Hierarchy[hierarchyIndex].Fallback = policy
		} else {
			result.Hierarchy[hierarchyIndex].Policy = policy
		}
	}

	return result
}

func CalculateEffectiveOfGvkPolicy(ctx context.Context, tx orm.DB,
	effectiveContext *EffectiveContext,
	appliedPolicyId string,
) (*dbmodel.EffectivePolicy, *api.Error) {
	ht := NewHierarchyTree(auth.TenantIdOf(ctx), effectiveContext.WorkloadType)

	effectiveContext.GvkWorkload = true

	withEffective := context.WithValue(ctx, EffectiveContextKey, effectiveContext)

	return ht.calculateEffectivePolicy(withEffective, tx, appliedPolicyId, 0, nil, nil, nil, nil)
}

// calculateEffectivePolicyAt calculates the effective policy of the node at idxToCalculate.
// Missing parent policies are calculated recursively before merging their rules and defaults.
func (ht *HierarchyTree) calculateEffectivePolicyAt(ctx context.Context, tx orm.DB,
	appliedPolicyId string, idxToCalculate int,
	imposedAssets map[string]interface{}, masterImposedAssets map[string]interface{},
	imposedAssetsRules map[string]interface{}, masterImposedAssetsRules map[string]interface{}) (
	*dbmodel.EffectivePolicy, *api.Error,
) {
	effectiveContext := ctx.Value(EffectiveContextKey).(*EffectiveContext)

	currentPolicyId := effectiveContext.Hierarchy[idxToCalculate].Policy.ID.String()
	currentGivenPolicy := effectiveContext.Hierarchy[idxToCalculate].Policy.Policy
	currentOrgUnit := effectiveContext.OrgUnits[idxToCalculate]

	currentSourceOfRule, orgUnitErr := currentOrgUnit.AsSourceOfRule(string(effectiveContext.WorkloadType))
	if orgUnitErr != nil {
		return nil, orgUnitErr
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Calculating effective policy given %s for %s [idx %d]", appliedPolicyId, currentPolicyId, idxToCalculate)

	// Find our which node inherits its policy to us.
	inheritFrom := effectiveContext.findInheritedPolicy(idxToCalculate)
	if inheritFrom < 0 {
		// if there is no node above us which has a policy, then our effective policy is equal to our given policy
		currentEffectivePolicy := &effectiveContext.PolicyById[currentPolicyId].EffectivePolicy
		currentEffectivePolicy.PolicyId = currentGivenPolicy.ID
		currentEffectivePolicy.TenantId = currentGivenPolicy.TenantId
		currentEffectivePolicy.ImposedAssets = currentGivenPolicy.ImposedAssets

		if !currentGivenPolicy.IsDeleted() {
			currentEffectivePolicy.RulesV2 = database.CleanMap(currentGivenPolicy.RulesV2.Clone()) // in effective, we add sourceOfRule, so we can't do it on the original map
			currentEffectivePolicy.DefaultsV2 = database.CleanMap(currentGivenPolicy.DefaultsV2)
		} else {
			currentEffectivePolicy.RulesV2 = nil
			currentEffectivePolicy.DefaultsV2 = nil
		}

		ht.InitSourceOfRules(currentEffectivePolicy.RulesV2, currentSourceOfRule, effectiveContext.GvkWorkload)

		if effectiveContext.GvkWorkload {
			ht.NormalizeAllowOverrideInGvkRules(currentEffectivePolicy.RulesV2)
		}

		// if we have fallback policy, we need to merge its rules and values to the effective policy of this policy.
		if effectiveContext.Hierarchy[idxToCalculate].Fallback != nil {
			fallbackPolicy := effectiveContext.Hierarchy[idxToCalculate].Fallback
			ht.mergeFallbackPolicy(effectiveContext, currentEffectivePolicy, &fallbackPolicy.EffectivePolicy, currentOrgUnit, currentSourceOfRule)
		}

		var enrichErr *api.Error
		currentEffectivePolicy.RulesV2, enrichErr = ht.enrichRulesWithImposed(currentEffectivePolicy.RulesV2, imposedAssetsRules, masterImposedAssetsRules)
		if enrichErr != nil {
			return nil, enrichErr
		}

		currentEffectivePolicy.DefaultFromOrder, currentEffectivePolicy.ValidationError = calculateDefaultFromOrder(currentEffectivePolicy.RulesV2)

		currentEffectivePolicy.DefaultsV2, enrichErr = ht.enrichDefaultsWithImposed(currentEffectivePolicy.DefaultsV2, imposedAssets, masterImposedAssets)
		if enrichErr != nil {
			return nil, enrichErr
		}

		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("[%s] effective policy of %s copied from its given policy", ht.treeId.ToString(), effectiveContext.OrgUnits[idxToCalculate].String())

		return currentEffectivePolicy, dao.UpsertEffectivePolicy(ctx, tx, currentEffectivePolicy)
	}

	// if the effective policy of the parent scope is missing, calculate it.
	if effectiveContext.Hierarchy[inheritFrom].Policy.EffectivePolicy.IsDeleted() {
		if _, err := ht.calculateEffectivePolicyAt(ctx, tx, appliedPolicyId, inheritFrom, nil, nil, nil, nil); err != nil {
			return nil, err
		}
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("[%s] effective policy of %s based on %s", ht.treeId.ToString(),
		effectiveContext.OrgUnits[idxToCalculate].String(), effectiveContext.OrgUnits[inheritFrom].String())

	inheritEffectivePolicy := effectiveContext.Hierarchy[inheritFrom].Policy.EffectivePolicy

	currentEffectivePolicy := effectiveContext.Hierarchy[idxToCalculate].Policy.EffectivePolicy
	currentEffectivePolicy.PolicyId = effectiveContext.Hierarchy[idxToCalculate].Policy.ID
	currentEffectivePolicy.TenantId = effectiveContext.Hierarchy[idxToCalculate].Policy.Policy.TenantId
	currentEffectivePolicy.RulesV2 = database.CleanMap(inheritEffectivePolicy.RulesV2.Clone())
	currentEffectivePolicy.ImposedAssets = effectiveContext.Hierarchy[idxToCalculate].Policy.Policy.ImposedAssets

	if !effectiveContext.GvkWorkload {
		currentEffectivePolicy.DefaultsV2 = database.CleanMap(inheritEffectivePolicy.DefaultsV2.Clone())
	}

	// Only merge rules and defaults if the current policy is not deleted
	if !currentGivenPolicy.IsDeleted() {
		if effectiveContext.GvkWorkload {
			var err error

			currentEffectivePolicy.RulesV2, err = ht.MergeGvkRules(inheritEffectivePolicy.RulesV2.Clone(), currentGivenPolicy.RulesV2, currentSourceOfRule)
			if err != nil {
				return nil, api.NewError(http.StatusBadRequest, "failed to merge given -> effective rules", err)
			}
		} else {
			//
			//   need to determine for each field, if the rules from the parent override rules of the current scope.
			//   note that system level is exception, its rules are never restrictive; they are only recommendations that
			//   lower level policies can override.
			//
			rulesFromParent := make(RulesFromParentPerField)

			if effectiveContext.Hierarchy[inheritFrom].Policy.Scope != api.SCOPETYPE_SYSTEM {
				rulesFromParent = ht.DetermineSourceOfRules(currentGivenPolicy.RulesV2, inheritEffectivePolicy.RulesV2)
			}

			var err error

			currentEffectivePolicy.RulesV2, err = ht.MergeRules(currentEffectivePolicy.RulesV2, currentGivenPolicy.RulesV2, rulesFromParent, currentSourceOfRule)
			if err != nil {
				return nil, api.NewError(http.StatusBadRequest, "failed to merge given -> effective rules", err)
			}

			currentEffectivePolicy.DefaultsV2, err = ht.MergeDefaults(currentEffectivePolicy.DefaultsV2, currentGivenPolicy.DefaultsV2, currentEffectivePolicy.RulesV2,
				currentOrgUnit, MergeDefaultsOptions{})
			if err != nil {
				return nil, api.NewError(http.StatusInternalServerError, "failed to merge given -> effective defaults", err)
			}
		}
	}

	if effectiveContext.Hierarchy[idxToCalculate].Fallback != nil {
		fallbackPolicy := effectiveContext.Hierarchy[idxToCalculate].Fallback
		ht.mergeFallbackPolicy(effectiveContext, &currentEffectivePolicy, &fallbackPolicy.EffectivePolicy, currentOrgUnit, currentSourceOfRule)
	}

	if !effectiveContext.GvkWorkload {
		var enrichErr *api.Error
		currentEffectivePolicy.RulesV2, enrichErr = ht.enrichRulesWithImposed(currentEffectivePolicy.RulesV2, imposedAssetsRules, masterImposedAssetsRules)
		if enrichErr != nil {
			return nil, enrichErr
		}

		currentEffectivePolicy.DefaultFromOrder, currentEffectivePolicy.ValidationError = calculateDefaultFromOrder(currentEffectivePolicy.RulesV2)

		currentEffectivePolicy.DefaultsV2, enrichErr = ht.enrichDefaultsWithImposed(currentEffectivePolicy.DefaultsV2, imposedAssets, masterImposedAssets)
		if enrichErr != nil {
			return nil, enrichErr
		}
	}

	return &currentEffectivePolicy, dao.UpsertEffectivePolicy(ctx, tx, &currentEffectivePolicy)
}

// NormalizeAllowOverrideInGvkRules ensures every GVK rule in the given rules
// map carries an explicit allowOverride value. When the field is missing
// (nil / omitted in the payload) it is set to false, matching the merge
// semantic where unspecified means "locked". This makes the flag always
// visible to API consumers of the effective rules — they no longer have to
// compensate for the "omitempty" JSON serialization when a rule was
// authored without an allowOverride value.
func (ht *HierarchyTree) NormalizeAllowOverrideInGvkRules(rulesMap map[string]interface{}) {
	if rulesMap == nil {
		return
	}
	rulesArr := convutils.ArrayMe(rulesMap[dbmodel.GvkRulesJsonKey])
	for idx := 0; idx < len(rulesArr); idx++ {
		rule := convutils.MapMe(rulesArr[idx])
		if _, ok := rule["allowOverride"]; !ok {
			rule["allowOverride"] = false
		}
	}
}

// InitSourceOfRules initializes the sourceOfRules for all the rules in the given policy to
// indicate the org unit of the given policy. during mergeRules, those rules that we decide
// to apply from the current given policy will thus have the correct source of rules.
func (ht *HierarchyTree) InitSourceOfRules(currentRulesMap map[string]interface{}, sourceOfRulesAsMap map[string]interface{}, gvkWorkload bool) {
	if gvkWorkload {
		// the structure of the rules in gvk-workloads is different, need to patch it accordingly
		// the rules are stores as an array, under "rules" key.
		rulesArr := convutils.ArrayMe(currentRulesMap[dbmodel.GvkRulesJsonKey])

		for idx := 0; idx < len(rulesArr); idx++ {
			rule := convutils.MapMe(rulesArr[idx])
			rule[policyCompliance.PolicySourceOfRule] = sourceOfRulesAsMap
		}

		return
	}

	for _, rule := range policyCompliance.GetRulesForSourceOfRules(currentRulesMap) {
		fieldRulesAsMap := convutils.MapMe(rule)
		if fieldRulesAsMap[policyCompliance.PolicySourceOfRule] == nil {
			fieldRulesAsMap[policyCompliance.PolicySourceOfRule] = sourceOfRulesAsMap
		}
	}
}

// DetermineSourceOfRules decides for each field that has rules in both the parent effective
// and the current given policy, if we want to prefer the rules from the parent or from the
// given policy.
func (ht *HierarchyTree) DetermineSourceOfRules(currentRulesMap map[string]interface{},
	parentSourceOfRules map[string]interface{},
) RulesFromParentPerField {
	fieldsWithRulesCurrent := policyCompliance.GetRulesForSourceOfRules(currentRulesMap)
	fieldsWithRulesParent := policyCompliance.GetRulesForSourceOfRules(parentSourceOfRules)

	rulesFromParentPerField := make(RulesFromParentPerField)

	// for each field in the given policy...
	for fieldWithRules := range fieldsWithRulesCurrent {
		field := fieldWithRules // we play with it so copy it aside
		fromParent := false

		// The current policy has a rule for this field.
		// If the parent's effective policy also has one, check whether it came from the system.
		// System rules stay overridable. For rules defined by tenant, cluster, or department policies,
		// use the checks below to decide whether the parent or current policy wins.
		if fieldsWithRulesParent[field] != nil && !isRuleFromSystem(fieldsWithRulesParent[field]) {
			// we do not distinguish master/worker fields

			field = strings.TrimPrefix(field, MasterPrefix)
			field = strings.TrimPrefix(field, WorkerPrefix)

			if strings.HasPrefix(field, securitySection+".") {
				// all the security fields are parent-takes-over mode
				fromParent = true
			} else if strings.HasPrefix(field, computeSection+".") && len(strings.Split(field, ".")) == 2 {
				// the instance rules of compute fields is parent-takes-over (this is actually only the
				// locked and canAdd rules of extendedResources)
				fromParent = true
			} else if strings.HasPrefix(field, computeSection+".") {
				_, isItemized := itemized.GetItemizedFieldInfo(field)

				// all the flat fields of the compute sections are parent-takes-over mode
				field = field[len(computeSection)+1:]
				fromParent = computeFields.Has(field) && !isItemized
			} else if strings.HasPrefix(field, storageSection+".") && len(strings.Split(field, ".")) == 2 {
				// the instance rules of storage (e.g storage.nfs, storage.pvc, etc) are from parent
				// to allow the parent to provide canAdd=false in upper level without lower scope bypassing it
				fromParent = true
			}
		}

		rulesFromParentPerField[fieldWithRules] = fromParent
	}

	return rulesFromParentPerField
}

func isRuleFromSystem(rule interface{}) bool {
	ruleAsMap := convutils.MapMe(rule)
	sourceOfRule := convutils.MapMe(ruleAsMap[policyCompliance.PolicySourceOfRule])
	return api.ScopeType(fmt.Sprintf("%v", sourceOfRule["scope"])) == api.SCOPETYPE_SYSTEM
}

func calculateDefaultFromOrder(rules map[string]interface{}) (*api.DefaultFromOrder, *string) {
	var result api.DefaultFromOrder
	var err error

	masterRules := rules[masterSection]
	if masterRules == nil {
		masterRules = rules[leaderSection]
	}
	workerRules := rules[workerSection]

	if masterRules != nil || workerRules != nil {
		result.Worker, err = policyCompliance.NewDefaultsFromOrder(convutils.MapMe(workerRules))
		if err != nil {
			return nil, convutils.Ref(err.Error())
		}

		result.Master, err = policyCompliance.NewDefaultsFromOrder(convutils.MapMe(masterRules))
		if err != nil {
			return nil, convutils.Ref(err.Error())
		}
	} else {
		result.Worker, err = policyCompliance.NewDefaultsFromOrder(rules)
		if err != nil {
			return nil, convutils.Ref(err.Error())
		}
	}

	return &result, nil
}

func (ht *HierarchyTree) mergeFallbackPolicy(effectiveContext *EffectiveContext, currentEffectivePolicy *dbmodel.EffectivePolicy, fallbackPolicy *dbmodel.EffectivePolicy,
	currentOrgUnit scopes.OrgUnit, currentSourceOfRule map[string]interface{},
) *api.Error {
	var err error

	if currentEffectivePolicy.RulesV2 == nil {
		currentEffectivePolicy.RulesV2 = make(map[string]interface{})
	}
	if currentEffectivePolicy.DefaultsV2 == nil {
		currentEffectivePolicy.DefaultsV2 = make(map[string]interface{})
	}

	rulesFromParent := make(RulesFromParentPerField)

	switch effectiveContext.WorkloadType {
	case WorkloadTypeDistributed, WorkloadTypeDistributedInference:
		return ht.mergeFallbackPolicyForDistributedWorkload(effectiveContext.WorkloadType, currentEffectivePolicy, fallbackPolicy, currentOrgUnit, currentSourceOfRule, rulesFromParent)
	default:
		currentEffectivePolicy.RulesV2, err = ht.MergeRules(fallbackPolicy.RulesV2, currentEffectivePolicy.RulesV2, rulesFromParent, currentSourceOfRule)
		if err != nil {
			return api.NewError(http.StatusBadRequest, "failed to merge fallback -> effective rules", err)
		}

		currentEffectivePolicy.DefaultsV2, err = ht.MergeDefaults(fallbackPolicy.DefaultsV2, currentEffectivePolicy.DefaultsV2, currentEffectivePolicy.RulesV2,
			currentOrgUnit, MergeDefaultsOptions{mergingWithFallback: true})
		if err != nil {
			return api.NewError(http.StatusInternalServerError, "failed to merge fallback -> effective defaults", err)
		}

		return nil
	}
}

// mergeFallbackPolicyForDistributedWorkload handles the merging logic for distributed (distributed inference and distributed training) workloads
func (ht *HierarchyTree) mergeFallbackPolicyForDistributedWorkload(workloadType WorkloadType, currentEffectivePolicy *dbmodel.EffectivePolicy, fallbackPolicy *dbmodel.EffectivePolicy,
	currentOrgUnit scopes.OrgUnit, currentSourceOfRule map[string]interface{}, rulesFromParent RulesFromParentPerField,
) *api.Error {

	// Determine the primary section name based on workload type
	var primarySection string
	switch workloadType {
	case WorkloadTypeDistributed:
		primarySection = masterSection
	case WorkloadTypeDistributedInference:
		primarySection = leaderSection
	default:
		return api.NewError(http.StatusBadRequest, "unsupported workload type for distributed merging", nil)
	}

	// fallback does not have sections for primary and worker, its defaults and rules must be applied
	// to both worker and primary section
	ht.expandFallbackPolicySections(fallbackPolicy, primarySection)

	// Get or create sections for current policy
	primaryRules := jsonutils.AddSection(currentEffectivePolicy.RulesV2, primarySection)
	workerRules := jsonutils.AddSection(currentEffectivePolicy.RulesV2, workerSection)

	primaryDefaults := jsonutils.AddSection(currentEffectivePolicy.DefaultsV2, primarySection)
	workerDefaults := jsonutils.AddSection(currentEffectivePolicy.DefaultsV2, workerSection)

	// Merge rules
	if fallbackPolicy.RulesV2 != nil {
		if err := ht.mergeRulesForSection(currentEffectivePolicy.RulesV2, fallbackPolicy.RulesV2, primarySection, primaryRules, rulesFromParent, currentSourceOfRule); err != nil {
			return api.NewError(http.StatusBadRequest, fmt.Sprintf("failed to merge fallback -> %s rules", primarySection), err)
		}
		if err := ht.mergeRulesForSection(currentEffectivePolicy.RulesV2, fallbackPolicy.RulesV2, workerSection, workerRules, rulesFromParent, currentSourceOfRule); err != nil {
			return api.NewError(http.StatusBadRequest, "failed to merge fallback -> worker rules", err)
		}

		// Clean up empty rules
		if len(convutils.MapMe(currentEffectivePolicy.RulesV2[primarySection])) == 0 && len(convutils.MapMe(currentEffectivePolicy.RulesV2[workerSection])) == 0 {
			currentEffectivePolicy.RulesV2 = nil
		}
	}

	// Merge defaults
	if fallbackPolicy.DefaultsV2 != nil {
		if err := ht.mergeDefaultsForSection(currentEffectivePolicy.DefaultsV2, fallbackPolicy.DefaultsV2, primarySection, primaryDefaults, primaryRules, currentOrgUnit); err != nil {
			return api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to merge fallback -> %s defaults", primarySection), err)
		}
		if err := ht.mergeDefaultsForSection(currentEffectivePolicy.DefaultsV2, fallbackPolicy.DefaultsV2, workerSection, workerDefaults, workerRules, currentOrgUnit); err != nil {
			return api.NewError(http.StatusInternalServerError, "failed to merge fallback -> worker defaults", err)
		}

		// Clean up empty defaults
		if len(convutils.MapMe(currentEffectivePolicy.DefaultsV2[primarySection])) == 0 && len(convutils.MapMe(currentEffectivePolicy.DefaultsV2[workerSection])) == 0 {
			currentEffectivePolicy.DefaultsV2 = nil
		}
	}

	return nil
}

// expandFallbackPolicySections expands fallback policy to include both primary and worker sections
func (ht *HierarchyTree) expandFallbackPolicySections(fallbackPolicy *dbmodel.EffectivePolicy, primarySection string) {
	if fallbackPolicy.DefaultsV2 != nil {
		fallbackDefaults := fallbackPolicy.DefaultsV2.Clone()
		fallbackPolicy.DefaultsV2 = map[string]interface{}{
			primarySection: map[string]interface{}(fallbackDefaults.Clone()),
			workerSection:  map[string]interface{}(fallbackDefaults),
		}
	}

	if fallbackPolicy.RulesV2 != nil {
		fallbackRules := fallbackPolicy.RulesV2.Clone()
		fallbackPolicy.RulesV2 = map[string]interface{}{
			primarySection: map[string]interface{}(fallbackRules.Clone()),
			workerSection:  map[string]interface{}(fallbackRules),
		}
	}
}

// mergeRulesForSection merges rules for a specific section
func (ht *HierarchyTree) mergeRulesForSection(currentRules map[string]interface{}, fallbackRules map[string]interface{}, section string, sectionRules map[string]interface{},
	rulesFromParent RulesFromParentPerField, currentSourceOfRule map[string]interface{}) error {
	var err error
	currentRules[section], err = ht.MergeRules(convutils.MapMe(fallbackRules[section]), sectionRules, rulesFromParent, currentSourceOfRule)
	return err
}

// mergeDefaultsForSection merges defaults for a specific section
func (ht *HierarchyTree) mergeDefaultsForSection(currentDefaults map[string]interface{}, fallbackDefaults map[string]interface{}, section string, sectionDefaults map[string]interface{},
	sectionRules map[string]interface{}, currentOrgUnit scopes.OrgUnit) error {
	var err error
	currentDefaults[section], err = ht.MergeDefaults(convutils.MapMe(fallbackDefaults[section]), sectionDefaults, sectionRules,
		currentOrgUnit, MergeDefaultsOptions{mergingWithFallback: true})
	return err
}

func (ht *HierarchyTree) MergeGvkRules(effectiveRules dbmodels.JsonBMap, givenRules dbmodels.JsonBMap,
	currentSourceOfRule map[string]interface{},
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
	ht.InitSourceOfRules(rulesOfGivenPolicy, currentSourceOfRule, true)
	ht.NormalizeAllowOverrideInGvkRules(rulesOfGivenPolicy)

	//
	//   if there are no rules from the parent, just copy the given rules to the result
	//
	if len(effectiveRules) == 0 {
		return database.CleanMap(rulesOfGivenPolicy), nil
	}

	//
	//   both parent-effective and given have rules. merge them by rule identity
	//   (field + specSelector + itemSelector + selectorParams).
	//
	effectiveArr := convutils.ArrayMe(effectiveRules[dbmodel.GvkRulesJsonKey])
	givenArr := convutils.ArrayMe(rulesOfGivenPolicy[dbmodel.GvkRulesJsonKey])

	keyToIdx := make(map[string]int, len(effectiveArr))
	for idx, rule := range effectiveArr {
		key := dbmodel.PolicyRuleRefKey(dbmodel.PolicyRuleRefFromMap(convutils.MapMe(rule)))
		keyToIdx[key] = idx
	}

	for _, rule := range givenArr {
		key := dbmodel.PolicyRuleRefKey(dbmodel.PolicyRuleRefFromMap(convutils.MapMe(rule)))
		if idx, found := keyToIdx[key]; found {
			parentRule := convutils.MapMe(effectiveArr[idx])

			// note that the default for allowOverride is false, so only if "true" is explicitly
			// sent in the json, the override=true takes place.
			allowOverride, allowOverrideSpecified := parentRule["allowOverride"].(bool)
			if allowOverrideSpecified && allowOverride {
				// when override is allowed, take the rule from the child scope. otherwise keep
				// the rule from the parent scope
				effectiveArr[idx] = rule
			}
			continue
		}

		keyToIdx[key] = len(effectiveArr)
		effectiveArr = append(effectiveArr, rule)
	}

	effectiveRules[dbmodel.GvkRulesJsonKey] = effectiveArr

	return database.CleanMap(effectiveRules), nil
}
