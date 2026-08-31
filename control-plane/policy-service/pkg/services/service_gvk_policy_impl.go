/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	authApi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/scopes"
	tree2 "github.com/run-ai/runai/control-plane/policy-service/pkg/tree"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"gorm.io/gorm"
)

// findGvkPolicyOptions is the standard FindPolicyByID options used by every GVK
// policy operation: the given policy must still be live, but its matching
// effective row is always loaded - even when soft-deleted - so recompute and
// update flows can observe and overwrite it in place rather than inserting a
// duplicate row.
var findGvkPolicyOptions = dao.FindPolicyOptions{
	IncludeDeletedGivenPolicies:     dbmodel.NotToIncludeDeletedPolicies,
	IncludeDeletedEffectivePolicies: dbmodel.IncludeDeletedPolicies,
}

const (
	addBreakdown     = true
	dontAddBreakdown = false
)

func (pa *PolicyApiService) createGvkPolicy(ctx context.Context, req api.CreateGvkPolicyReq) (*api.Policy, *api.Error) {
	committed := false
	tx := dao_utils.DB().WithContext(ctx).Begin()
	defer database.FinalizeTransaction(tx, &committed)

	tenantId := auth.TenantIdOf(ctx)

	appliedOrgUnit, scopeParams, scopeErr := NewScopeInfo(ctx, req.Scope, req.ClusterId, req.DepartmentId, req.ProjectId)
	if scopeErr != nil {
		return nil, scopeErr
	}

	rbacErr := auth.ValidateAccessByResourceTypeAndAction(ctx, scopeParams, authApi.RESOURCETYPE_POLICIES, authApi.ACTION_CREATE, nonPrivate)
	if rbacErr != nil {
		return nil, api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))
	}

	// build array containing the relevant org units (the one that the policy is applied on, and all its ancestors).
	// the first org unit is the one that the given policy is being applied on
	relevantOrgUnits, err := appliedOrgUnit.Ancestors(ctx)
	if err != nil {
		if err.Code == http.StatusNotFound {
			return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("%s does not exist", appliedOrgUnit.Scope), nil)
		}
		return nil, err
	}

	workloadTypeKey, gvkErr := encodeGvkWorkloadType(req.WorkloadType)
	if gvkErr != nil {
		return nil, gvkErr
	}

	if len(req.Rules) == 0 {
		return nil, api.NewValidationError("a policy must have at least one rule")
	}

	// POST is for creating a new policy. updating an existing policy at the
	// same (scope, workloadType) must go through PUT / PATCH so the caller
	// is explicit about touching an existing resource. without this check
	// the multi-scope refactor's makeSureAllPoliciesExist conflict-swallow
	// turns POST into a silent upsert, which also resets CreatedAt /
	// CreatedBy on the existing row.
	var projectIdNum *int32
	if req.ProjectId != nil && *req.ProjectId != "" {
		n, projErr := projectIdToNum(*req.ProjectId)
		if projErr != nil {
			return nil, projErr
		}
		projectIdNum = &n
	}
	existing, findErr := dao.FindPolicyOfOrgUnit(ctx, dao_utils.DB().WithContext(ctx),
		req.Scope, req.DepartmentId, projectIdNum, req.ClusterId, workloadTypeKey,
		dbmodel.NotToIncludeDeletedPolicies)
	if findErr != nil {
		return nil, findErr
	}
	if existing != nil {
		return nil, api.NewError(http.StatusConflict,
			"a policy already exists for this scope and workload type",
			nil)
	}

	rulesV2, rulesErr := dbmodel.ToGvkRulesApi[api.PolicyRule](req.Rules, nil)
	if rulesErr != nil {
		return nil, rulesErr
	}

	// create returns on the first user-facing finding, so the collector is
	// capped at 1: both GeneratePolicyBreakdown and validateRules append to
	// the same instance, the cap ensures they short-circuit as soon as
	// anything is wrong.
	issues := newRuleIssues(1)

	breakdownResult, err := pa.GeneratePolicyBreakdown(ctx, tx, req.WorkloadType, req.Rules, issues)
	if err != nil {
		return nil, err
	}
	if e := issues.toError(); e != nil {
		return nil, e
	}

	var breakdown *api.BreakdownBreakdown
	if breakdownResult != nil {
		breakdown = breakdownResult.Breakdown
	}

	if e := pa.validateRules(ctx, breakdown, req.Rules, issues).toError(); e != nil {
		return nil, e
	}

	db := dao_utils.DB().WithContext(ctx)

	updated := false

	var appliedPolicy *dbmodel.PolicyWithOptionalEffective

	// we may need to re-calculate the effective in case one of the given policies has changed during the calculation
	for attempt := 0; attempt < tree2.MaxEffectiveCalcAttempts; attempt++ {
		err = makeSureAllPoliciesExist(ctx, db, workloadTypeKey, relevantOrgUnits)
		if err != nil {
			return nil, err
		}

		// now we're ready to update the database
		committed := false
		tx := db.Begin()
		defer database.FinalizeTransaction(tx, &committed)

		// list the policy of all the relevant nodes (applied node + all its ancestors)
		relevantPolicies, err := dao.ListPoliciesByOrgUnits(ctx, tx, tenantId, workloadTypeKey, relevantOrgUnits,
			dao.ListByOrgUnitOptions{
				IncludeDeleted:  dbmodel.IncludeDeletedPolicies,
				IncludeFallback: false,
				IncludeSystem:   false,
			})
		if err != nil {
			return nil, err
		}

		appliedPolicy = tree2.FindRelevantPolicy(relevantPolicies, *appliedOrgUnit, workloadTypeKey)

		err = applyCreateReqOnPolicy(ctx, appliedPolicy, req, workloadTypeKey, rulesV2)
		if err != nil {
			return nil, err
		}

		appliedPolicy.Policy.DeletedAt = gorm.DeletedAt{}

		err = dao.UpdatePolicySpec(ctx, tx, &appliedPolicy.Policy)
		if err != nil {
			return nil, err
		}

		// calculate the effective policy of the org unit for which we apply the policy
		effectiveCtx := tree2.NewEffectiveContext(ctx, tree2.WorkloadType(workloadTypeKey), relevantOrgUnits, relevantPolicies)

		effectivePolicy, effErr := tree2.CalculateEffectiveOfGvkPolicy(ctx, tx, effectiveCtx, appliedPolicy.ID.String())
		if effErr != nil {
			return nil, effErr
		}

		//  validation errors should reject the apply request. this is different from get, that
		//  can have defaultFrom issues which will appear in the payload but will not fail the get
		if effectivePolicy.ValidationError != nil {
			return nil, api.NewError(http.StatusBadRequest, *effectivePolicy.ValidationError, nil)
		}

		appliedPolicy.EffectivePolicy = *effectivePolicy

		// SoftDeleteEffectivePoliciesOfSubtree returns the descendant policy
		// ids; we forward them to DeleteBreakdownCachesOfSubtree so the
		// cache invalidation step doesn't have to re-traverse `policies`.
		// no-op for project scope (no subtree).
		descendantIds, err := dao.SoftDeleteEffectivePoliciesOfSubtree(ctx, tx, *appliedOrgUnit, workloadTypeKey)
		if err != nil {
			return nil, err
		}

		if err := dao.DeleteBreakdownCachesOfSubtree(ctx, tx, descendantIds); err != nil {
			return nil, err
		}

		result := tx.Commit()
		if result.Error != nil {
			return nil, database.Error(database.UPSERT, database.Policies).WithDetailsFromDb(result)
		}
		committed = true

		updatedPolicies, err := dao.ListPoliciesByOrgUnits(ctx, db, tenantId, workloadTypeKey, relevantOrgUnits,
			dao.ListByOrgUnitOptions{
				IncludeDeleted:  dbmodel.IncludeDeletedPolicies,
				IncludeFallback: false,
				IncludeSystem:   false,
			})
		if err != nil {
			return nil, err
		}

		if !tree2.IsPolicyChangedDuringCalculation(ctx, appliedPolicy.ID.String(), relevantPolicies, updatedPolicies) {
			updated = true
			break
		}
	}

	if !updated {
		return nil, api.NewError(http.StatusConflict, "policies are currently being updated, try again later", nil)
	}

	//
	//  cache the breakdown. at this point we need it mainly for the dependency in other components.
	//  share tx with the surrounding create transaction so the cache row and its dependency edges
	//  commit/roll back atomically with the policy row written above.
	//
	if _, cacheErr := dao.CacheBreakdown(ctx, tx, appliedPolicy.ID.String(), breakdownResult.Breakdown, breakdownResult.Dependencies); cacheErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to cache breakdown for policy %s: %v", appliedPolicy.ID.String(), cacheErr)
		return nil, cacheErr
	}

	return appliedPolicy.ToGvkPolicyApi()
}

func (pa *PolicyApiService) getEffectivePolicy(ctx context.Context, req api.WorkloadTypeAndScope) (*api.EffectiveRules, *api.Error) {
	_, scopeParams, err := NewScopeInfo(ctx, req.Scope, req.ClusterId, req.DepartmentId, req.ProjectId)
	if err != nil {
		return nil, err
	}

	rbacErr := auth.ValidateAccessByResourceTypeAndAction(ctx, scopeParams, authApi.RESOURCETYPE_POLICIES, authApi.ACTION_READ, nonPrivate)
	if rbacErr != nil {
		return nil, api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))
	}

	result, err := pa.getGvkPolicyByTypeAndScope(ctx, req, dontAddBreakdown)
	if err != nil {
		return nil, err
	}

	return &api.EffectiveRules{EffectiveRules: result.EffectiveRules}, nil
}

func (pa *PolicyApiService) getGvkPolicyByTypeAndScope(ctx context.Context, req api.WorkloadTypeAndScope, toAddBreakdown bool) (*api.GetGvkPolicyResp, *api.Error) {
	workloadType, gvkErr := encodeGvkWorkloadType(req.WorkloadType)
	if gvkErr != nil {
		return nil, gvkErr
	}

	var projectId *int32
	if req.ProjectId != nil {
		pid, parseErr := projectIdToNum(*req.ProjectId)
		if parseErr != nil {
			return nil, parseErr
		}
		projectId = &pid
	}

	db := dao_utils.DB().WithContext(ctx)

	stored, findErr := dao.FindPolicyOfOrgUnit(ctx, db, req.Scope, nil, projectId, nil, workloadType, dbmodel.IncludeDeletedPolicies)
	if findErr != nil {
		return nil, findErr
	}

	if stored == nil || !stored.GvkWorkload {
		// No GVK policy is stored for this org unit yet. Insert an empty,
		// soft-deleted placeholder so getGvkPolicy's recalculation path
		// can derive an effective policy from ancestor policies.
		appliedOrgUnit, _, scopeErr := NewScopeInfo(ctx, req.Scope, req.ClusterId, req.DepartmentId, req.ProjectId)
		if scopeErr != nil {
			return nil, scopeErr
		}

		if err := makeSureAllPoliciesExist(ctx, db, workloadType, []scopes.OrgUnit{*appliedOrgUnit}); err != nil {
			return nil, err
		}

		stored, findErr = dao.FindPolicyOfOrgUnit(ctx, db, req.Scope, nil, projectId, nil, workloadType, dbmodel.IncludeDeletedPolicies)
		if findErr != nil {
			return nil, findErr
		}
		if stored == nil {
			return nil, api.NewError(http.StatusInternalServerError, "failed to create placeholder gvk policy", nil)
		}
	}

	return pa.getGvkPolicy(ctx, db, stored, toAddBreakdown)
}

func (pa *PolicyApiService) getGvkPolicyById(ctx context.Context, policyId string, toAddBreakdown bool) (*api.GetGvkPolicyResp, *api.Error) {
	db := dao_utils.DB().WithContext(ctx)

	stored, findErr := dao.FindPolicyByID(ctx, db, policyId, findGvkPolicyOptions)
	if findErr != nil {
		return nil, findErr
	}

	if stored == nil || !stored.GvkWorkload {
		return nil, api.NewError(http.StatusNotFound, "policy not found", nil)
	}

	return pa.getGvkPolicy(ctx, db, stored, toAddBreakdown)
}

func (pa *PolicyApiService) getGvkPolicy(ctx context.Context, db *gorm.DB, stored *dbmodel.PolicyWithOptionalEffective, toAddBreakdown bool) (*api.GetGvkPolicyResp, *api.Error) {
	_, scopeParams, scopeErr := NewScopeInfoFromPolicy(ctx, &stored.Policy)
	if scopeErr != nil {
		return nil, scopeErr
	}

	tenantId := auth.TenantIdOf(ctx)

	rbacErr := auth.ValidateAccessByResourceTypeAndAction(ctx, scopeParams, authApi.RESOURCETYPE_POLICIES, authApi.ACTION_READ, nonPrivate)
	if rbacErr != nil {
		return nil, api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))
	}

	if stored.EffectivePolicy.IsDeleted() {
		appliedOrgUnit, _, scopeErr := NewScopeInfoFromPolicy(ctx, &stored.Policy)
		if scopeErr != nil {
			return nil, scopeErr
		}

		// build array containing the relevant org units (the one that the policy is applied on, and all its ancestors).
		// the first org unit is the one that the given policy is being applied on
		relevantOrgUnits, err := appliedOrgUnit.Ancestors(ctx)
		if err != nil {
			if err.Code == http.StatusNotFound {
				return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("%s does not exist", appliedOrgUnit.Scope), nil)
			}
			return nil, err
		}

		updated := false

		for attempt := 0; attempt < tree2.MaxEffectiveCalcAttempts; attempt++ {
			err := makeSureAllPoliciesExist(ctx, db, stored.WorkloadType, relevantOrgUnits)
			if err != nil {
				return nil, err
			}

			// now we're ready to update the database
			committed := false
			tx := db.Begin()
			defer database.FinalizeTransaction(tx, &committed)

			// list the policy of all the relevant nodes (applied node + all its ancestors)
			relevantPolicies, err := dao.ListPoliciesByOrgUnits(ctx, tx, tenantId, stored.WorkloadType, relevantOrgUnits,
				dao.ListByOrgUnitOptions{
					IncludeDeleted:  dbmodel.IncludeDeletedPolicies,
					IncludeFallback: false,
					IncludeSystem:   false,
				})
			if err != nil {
				return nil, err
			}

			// calculate the effective policy of the org unit for which we apply the policy
			effectiveCtx := tree2.NewEffectiveContext(ctx, tree2.WorkloadType(stored.WorkloadType), relevantOrgUnits, relevantPolicies)

			effectivePolicy, effErr := tree2.CalculateEffectiveOfGvkPolicy(ctx, tx, effectiveCtx, stored.ID.String())
			if effErr != nil {
				return nil, effErr
			}

			//  validation errors should reject the apply request. this is different from get, that
			//  can have defaultFrom issues which will appear in the payload but will not fail the get
			if effectivePolicy.ValidationError != nil {
				return nil, api.NewError(http.StatusBadRequest, *effectivePolicy.ValidationError, nil)
			}

			stored.EffectivePolicy = *effectivePolicy

			result := tx.Commit()
			if result.Error != nil {
				return nil, database.Error(database.UPSERT, database.Policies).WithDetailsFromDb(result)
			}
			committed = true

			updatedPolicies, err := dao.ListPoliciesByOrgUnits(ctx, db, tenantId, stored.WorkloadType, relevantOrgUnits,
				dao.ListByOrgUnitOptions{
					IncludeDeleted:  dbmodel.IncludeDeletedPolicies,
					IncludeFallback: false,
					IncludeSystem:   false,
				})
			if err != nil {
				return nil, err
			}

			if !tree2.IsPolicyChangedDuringCalculation(ctx, stored.ID.String(), relevantPolicies, updatedPolicies) {
				updated = true
				break
			}
		}

		if !updated {
			return nil, api.NewError(http.StatusConflict, "policies are currently being updated, try again later", nil)
		}
	}

	out, convErr := stored.ToGvkPolicyApi()
	if convErr != nil {
		return nil, convErr
	}

	var gvkPolicyResp api.GetGvkPolicyResp

	copyErr := convutils.CopySpecs(&gvkPolicyResp, out)
	if copyErr != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to copy policy spec", copyErr)
	}

	// Always run breakdown + validate to check policy health and update status.
	// The breakdown is only added to the response if toAddBreakdown is true.
	var breakdown *api.BreakdownBreakdown
	var breakdownResult *BreakdownResult

	cachedBreakdown, err := dao.GetCachedBreakdown(ctx, stored.ID.String())
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get cached breakdown for policy %s: %v", stored.ID.String(), err)
	}

	gvk, err := dbmodel.DecodeGvkWorkloadType(stored.WorkloadType)
	if err != nil {
		return nil, err
	}

	rules, err := dbmodel.FromGvkRulesApi[api.PolicyRule](stored.EffectivePolicy.RulesV2)
	if err != nil {
		return nil, err
	}

	// collect all issues (unlimited) for status reporting
	issues := newRuleIssues(0)

	if cachedBreakdown != nil {
		// Cache invariant: OnComponentDeleted clears the cache and resets the policy
		// to PendingUpdate whenever any dependency changes. So a cache hit means the
		// stored ValidationStatus/Issues still describe this exact breakdown, and the
		// response Status (already populated by CopySpecs from stored) is up-to-date.
		breakdown = cachedBreakdown
	} else {
		committed := false
		tx := db.Begin()
		defer database.FinalizeTransaction(tx, &committed)

		// generate breakdown if not cached - issues are collected for status reporting
		breakdownResult, err = pa.GeneratePolicyBreakdown(ctx, tx, gvk, rules, issues)
		if err != nil {
			return nil, err
		}

		if breakdownResult != nil {
			breakdown = breakdownResult.Breakdown

			if _, cacheErr := dao.CacheBreakdown(ctx, tx, stored.ID.String(), breakdownResult.Breakdown, breakdownResult.Dependencies); cacheErr != nil {
				return nil, cacheErr
			}

			pa.validateRules(ctx, breakdown, rules, issues)
		}

		// determine the new status based on validation results
		var newStatus *api.PolicyStatus

		collectedIssues := issues.all()
		if len(collectedIssues) > 0 {
			issueMessages := make([]string, len(collectedIssues))
			for i, issue := range collectedIssues {
				issueMessages[i] = issue.Message
			}

			newStatus = &api.PolicyStatus{
				Phase:  api.POLICYPHASE_DEGRADED,
				Issues: issueMessages,
			}
		} else {
			newStatus = &api.PolicyStatus{
				Phase:  api.POLICYPHASE_READY,
				Issues: []string{},
			}
		}

		// update status in DB if it changed
		statusChanged := stored.ValidationStatus == nil ||
			*stored.ValidationStatus != newStatus.Phase ||
			!stringSlicesEqual(stored.ValidationIssues, newStatus.Issues)

		if statusChanged {
			if updateErr := dao.UpdatePoliciesStatus(ctx, tx, []string{stored.ID.String()}, newStatus); updateErr != nil {
				return nil, updateErr
			}
			gvkPolicyResp.Status = newStatus
		}

		if res := tx.Commit(); res.Error != nil {
			return nil, database.Error(database.UPDATE, database.Policies).WithDetailsFromDb(res)
		}
		committed = true
	}

	// only add breakdown to response if requested
	if toAddBreakdown && breakdown != nil {
		gvkPolicyResp.Breakdown = breakdown
	}

	return &gvkPolicyResp, nil
}

// GeneratePolicyBreakdown builds the breakdown for the given policy rules.
// Any rule whose field/specSelector cannot be resolved against the mapping
// is reported into the supplied issues collector (so the caller — create,
// update, or get — surfaces the same finding that validateRules would have
// reported on the resulting incomplete breakdown). The collector is also
// what bounds the work via maxErrors. Pass nil for issues when running in
// a context where there are no rules to attribute findings to.
func (pa *PolicyApiService) GeneratePolicyBreakdown(ctx context.Context, tx *gorm.DB, gvk api.Gvk, rules []api.PolicyRule, issues *ruleIssues) (*BreakdownResult, *api.Error) {
	bg, bgErr := NewBreakdownGenerator(ctx, tx, gvk)
	if bgErr != nil {
		return nil, bgErr
	}

	//
	//  collect all the fields that take part in the policy
	//
	refs := make([]FieldRef, len(rules))

	for i, r := range rules {
		itemSel := ""
		if r.ItemSelector != nil {
			itemSel = *r.ItemSelector
		}

		refs[i] = FieldRef{
			Field:        r.Field,
			SpecSelector: r.SpecSelector,
			ItemSelector: itemSel,
			RuleIndex:    i,
		}
	}

	return bg.Generate(ctx, refs, issues)
}

// getMappingMinimal returns minimal information about all fields in a mapping.
func (pa *PolicyApiService) getMappingMinimal(ctx context.Context, gvk api.Gvk) (*api.GetMappingMinimalResp, *api.Error) {
	tx := dao_utils.DB().WithContext(ctx)

	bg, bgErr := NewBreakdownGenerator(ctx, tx, gvk)
	if bgErr != nil {
		return nil, bgErr
	}

	fgNamesSet := datastructs.NewSet[string]()
	for _, sel := range bg.mapping.SpecSelectors {
		for _, fgName := range sel.FieldGroups {
			fgNamesSet.Add(fgName)
		}
	}

	fgByNameScope, fgErr := dao.GetFieldGroupsByNames(ctx, tx, fgNamesSet.Keys())
	if fgErr != nil {
		return nil, fgErr
	}

	mergedFGs, mergeErr := bg.mergeFieldGroups(fgByNameScope)
	if mergeErr != nil {
		return nil, mergeErr
	}

	fields := make(map[string]map[string]map[string]string)
	specSelectors := make(map[string]string)

	for selName, sel := range bg.mapping.SpecSelectors {
		if sel.Description != nil {
			specSelectors[selName] = *sel.Description
		} else {
			specSelectors[selName] = ""
		}

		for _, fgName := range sel.FieldGroups {
			fg, ok := mergedFGs[fgName]
			if !ok {
				continue
			}
			for fieldName, fieldDef := range fg.Fields {
				if fields[fieldName] == nil {
					fields[fieldName] = make(map[string]map[string]string)
				}

				itemSels := make(map[string]string, len(fieldDef.ItemSelectors))
				for _, isName := range fieldDef.ItemSelectors {
					desc := ""
					if isSel, found := fg.ItemSelectors[isName]; found && isSel.Description != nil {
						desc = *isSel.Description
					}
					itemSels[isName] = desc
				}

				fields[fieldName][selName] = itemSels
			}
		}
	}

	return &api.GetMappingMinimalResp{
		Fields:        fields,
		SpecSelectors: specSelectors,
	}, nil
}

// getMappingFieldInfo returns breakdown information for specific field refs in a mapping.
// When a ref's specSelector is empty, the breakdown includes all spec selectors
// that expose the field.
func (pa *PolicyApiService) getMappingFieldInfo(ctx context.Context, gvk api.Gvk, fieldsRef []api.BreakdownFieldRef) (*api.Breakdown, *api.Error) {
	if len(fieldsRef) == 0 {
		return nil, api.NewError(http.StatusBadRequest, "fieldsRef is required", nil)
	}

	tx := dao_utils.DB().WithContext(ctx)

	bg, bgErr := NewBreakdownGenerator(ctx, tx, gvk)
	if bgErr != nil {
		return nil, bgErr
	}

	// Pre-compute spec selectors for fields that have no specSelector specified.
	// This avoids repeated DB lookups for the same field name.
	fieldsNeedingExpansion := datastructs.NewSet[string]()
	for _, r := range fieldsRef {
		if r.SpecSelector == nil || *r.SpecSelector == "" {
			fieldsNeedingExpansion.Add(r.FieldName)
		}
	}

	selectorsPerField := make(map[string][]string, len(fieldsNeedingExpansion))
	for _, fieldName := range fieldsNeedingExpansion.Keys() {
		selectors, findErr := bg.FindSpecSelectorsForField(ctx, fieldName)
		if findErr != nil {
			return nil, findErr
		}
		selectorsPerField[fieldName] = selectors
	}

	var refs []FieldRef
	for i, r := range fieldsRef {
		itemSel := ""
		if r.ItemSelector != nil {
			itemSel = *r.ItemSelector
		}

		specSel := ""
		if r.SpecSelector != nil {
			specSel = *r.SpecSelector
		}

		if specSel == "" {
			for _, sel := range selectorsPerField[r.FieldName] {
				refs = append(refs, FieldRef{
					Field:        r.FieldName,
					SpecSelector: sel,
					ItemSelector: itemSel,
					RuleIndex:    i,
				})
			}
		} else {
			refs = append(refs, FieldRef{
				Field:        r.FieldName,
				SpecSelector: specSel,
				ItemSelector: itemSel,
				RuleIndex:    i,
			})
		}
	}

	// no policy context here, so refs aren't really "rules"; pass a fresh
	// collector that we discard. Generate requires a non-nil collector —
	// we never inspect what lands in it because there's no rule to
	// attribute findings to.
	result, err := bg.Generate(ctx, refs, newRuleIssues(0))
	if err != nil {
		return nil, err
	}

	if result == nil {
		return &api.Breakdown{}, nil
	}

	return &api.Breakdown{Breakdown: result.Breakdown}, nil
}

func (pa *PolicyApiService) listGvkPolicies(ctx context.Context, filterBy []string,
	sortBy string, offset int32, limit int32, sortOrder string,
) (*api.ListGvkPoliciesResp, *api.Error) {
	permittedScopes, authErr := authorization.Get().Client().CheckPermittedScopes(ctx, authApi.RESOURCETYPE_POLICIES, authApi.ACTION_READ)
	if authErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("check permitted scopes failed: %s: %v", authErr.Message, authErr.Details)
		return nil, api.NewError(int(authErr.Code), authErr.Message, nil)
	}

	filters, filterErr := dao_utils.ValidateAndParameterizeApiFilters(filterBy, dbmodel.Policy{}.FilterFieldsToColumn())
	if filterErr != nil {
		return nil, api.NewValidationError(filterErr.Error())
	}

	sortColumn, sortError := dao_utils.ValidateSortByField(sortBy, dbmodel.Policy{}.SortFieldsToColumn())
	if sortError != nil {
		return nil, api.NewValidationError(sortError.Error())
	}

	records, err := dao.ListGvkPoliciesOfTenant(ctx, permittedScopes, filters, sortColumn, sortOrder, offset, limit)
	if err != nil {
		return nil, err
	}

	items := make([]api.Policy, 0, len(records))
	for i := range records {
		p, convErr := records[i].ToGvkPolicyApi()
		if convErr != nil {
			return nil, convErr
		}

		items = append(items, *p)
	}

	return &api.ListGvkPoliciesResp{
		Items: items,
	}, nil
}

func (pa *PolicyApiService) deleteGvkPolicy(ctx context.Context, policyId string) *api.Error {
	committed := false
	tx := dao_utils.DB().WithContext(ctx).Begin()
	defer database.FinalizeTransaction(tx, &committed)

	stored, findErr := dao.FindPolicyByID(ctx, tx, policyId, findGvkPolicyOptions)
	if findErr != nil {
		return findErr
	}
	if stored == nil || !stored.GvkWorkload {
		return api.NewError(http.StatusNotFound, "policy not found", nil)
	}

	appliedOrgUnit, scopeParams, scopeErr := NewScopeInfoFromPolicy(ctx, &stored.Policy)
	if scopeErr != nil {
		return scopeErr
	}

	rbacErr := auth.ValidateAccessByResourceTypeAndAction(ctx, scopeParams, authApi.RESOURCETYPE_POLICIES, authApi.ACTION_DELETE, nonPrivate)
	if rbacErr != nil {
		return api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))
	}

	if delErr := dao.HardDeletePolicyById(ctx, tx, &stored.Policy); delErr != nil {
		return delErr
	}

	// mark the effective policies of the subtree as deleted, they will be
	// recalculated when needed (lazily). the returned ids feed straight
	// into the breakdown-cache cleanup below, so the cache step doesn't
	// have to re-traverse `policies`.
	descendantIds, err := dao.SoftDeleteEffectivePoliciesOfSubtree(ctx, tx, *appliedOrgUnit, stored.WorkloadType)
	if err != nil {
		return err
	}

	if err := dao.OnEntityDeleted(ctx, tx, dbmodel.EntityTypeBreakdown, policyId); err != nil {
		return err
	}

	if err := dao.DeleteBreakdownCachesOfSubtree(ctx, tx, descendantIds); err != nil {
		return err
	}

	if res := tx.Commit(); res.Error != nil {
		return database.Error(database.DELETE, database.Policies).WithDetailsFromDb(res)
	}
	committed = true

	return nil
}

func (pa *PolicyApiService) deleteGvkPolicyRules(ctx context.Context, policyId string, req api.DeleteGvkPolicyRulesReq) (*api.Policy, *api.Error) {
	db := dao_utils.DB().WithContext(ctx)

	stored, findErr := dao.FindPolicyByID(ctx, db, policyId, findGvkPolicyOptions)
	if findErr != nil {
		return nil, findErr
	}
	if stored == nil || !stored.GvkWorkload {
		return nil, api.NewError(http.StatusNotFound, "policy not found", nil)
	}

	appliedOrgUnit, scopeParams, scopeErr := NewScopeInfoFromPolicy(ctx, &stored.Policy)
	if scopeErr != nil {
		return nil, scopeErr
	}

	rbacErr := auth.ValidateAccessByResourceTypeAndAction(ctx, scopeParams, authApi.RESOURCETYPE_POLICIES, authApi.ACTION_UPDATE, nonPrivate)
	if rbacErr != nil {
		return nil, api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))
	}

	// build array containing the relevant org units (the one that the policy is applied on, and all its ancestors).
	// the first org unit is the one that the given policy is being applied on
	relevantOrgUnits, ancestorErr := appliedOrgUnit.Ancestors(ctx)
	if ancestorErr != nil {
		if ancestorErr.Code == http.StatusNotFound {
			return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("%s does not exist", appliedOrgUnit.Scope), nil)
		}
		return nil, ancestorErr
	}

	gvk, gvkErr := dbmodel.DecodeGvkWorkloadType(stored.WorkloadType)
	if gvkErr != nil {
		return nil, gvkErr
	}

	tenantId := auth.TenantIdOf(ctx)

	updated := false

	// we may need to re-calculate the effective in case one of the given
	// policies has changed during the calculation — mirror the retry loop
	// in createGvkPolicy so concurrent ancestor changes don't leave us with
	// a stale effective.
	for attempt := 0; attempt < tree2.MaxEffectiveCalcAttempts; attempt++ {
		if err := makeSureAllPoliciesExist(ctx, db, stored.WorkloadType, relevantOrgUnits); err != nil {
			return nil, err
		}

		committed := false
		tx := db.Begin()
		defer database.FinalizeTransaction(tx, &committed)

		relevantPolicies, listErr := dao.ListPoliciesByOrgUnits(ctx, tx, tenantId, stored.WorkloadType, relevantOrgUnits,
			dao.ListByOrgUnitOptions{
				IncludeDeleted:  dbmodel.IncludeDeletedPolicies,
				IncludeFallback: false,
				IncludeSystem:   false,
			})
		if listErr != nil {
			return nil, listErr
		}

		appliedPolicy := tree2.FindRelevantPolicy(relevantPolicies, *appliedOrgUnit, stored.WorkloadType)
		if appliedPolicy == nil {
			return nil, api.NewError(http.StatusNotFound, "policy not found", nil)
		}

		// re-read the current rule set inside the loop so concurrent edits
		// to other rules survive the removal.
		existing, fromErr := dbmodel.FromGvkRulesApi[api.PolicyRule](appliedPolicy.Policy.RulesV2)
		if fromErr != nil {
			return nil, fromErr
		}

		next, removeErr := removeGvkPolicyRulesByRefs(existing, req.Rules)
		if removeErr != nil {
			return nil, removeErr
		}
		if len(next) == 0 {
			return nil, api.NewValidationError("a policy must have at least one rule")
		}

		rulesV2, convErr := dbmodel.ToGvkRulesApi[api.PolicyRule](next, nil)
		if convErr != nil {
			return nil, convErr
		}

		// regenerate the breakdown for the remaining rules
		issues := newRuleIssues(0)
		breakdownResult, bdErr := pa.GeneratePolicyBreakdown(ctx, tx, gvk, next, issues)
		if bdErr != nil {
			return nil, bdErr
		}
		breakdownHadIssues := len(issues.all()) > 0

		appliedPolicy.Policy.RulesV2 = rulesV2
		appliedPolicy.Policy.UpdatedAt = time.Now()
		appliedPolicy.UpdatedBy = auth.SubjectIdOf(ctx)

		// remaining rules are still valid → status Ready with no issues
		readyPhase := api.POLICYPHASE_READY
		appliedPolicy.ValidationStatus = &readyPhase
		appliedPolicy.ValidationIssues = []string{}

		if upErr := dao.UpdatePolicySpec(ctx, tx, &appliedPolicy.Policy); upErr != nil {
			return nil, upErr
		}

		// calculate the effective policy with the full relevant chain so
		// ancestor rules are honored.
		effectiveCtx := tree2.NewEffectiveContext(ctx, tree2.WorkloadType(stored.WorkloadType), relevantOrgUnits, relevantPolicies)

		effectivePolicy, effErr := tree2.CalculateEffectiveOfGvkPolicy(ctx, tx, effectiveCtx, appliedPolicy.ID.String())
		if effErr != nil {
			return nil, effErr
		}

		appliedPolicy.EffectivePolicy = *effectivePolicy

		descendantIds, sdErr := dao.SoftDeleteEffectivePoliciesOfSubtree(ctx, tx, *appliedOrgUnit, stored.WorkloadType)
		if sdErr != nil {
			return nil, sdErr
		}
		if err := dao.DeleteBreakdownCachesOfSubtree(ctx, tx, descendantIds); err != nil {
			return nil, err
		}

		// refresh this policy's own cache atomically with the spec update
		if err := dao.OnEntityDeleted(ctx, tx, dbmodel.EntityTypeBreakdown, policyId); err != nil {
			return nil, err
		}
		if breakdownResult != nil && !breakdownHadIssues {
			if _, cacheErr := dao.CacheBreakdown(ctx, tx, policyId, breakdownResult.Breakdown, breakdownResult.Dependencies); cacheErr != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to cache breakdown for policy %s: %v", policyId, cacheErr)
				return nil, cacheErr
			}
		}

		result := tx.Commit()
		if result.Error != nil {
			return nil, database.Error(database.UPDATE, database.Policies).WithDetailsFromDb(result)
		}
		committed = true

		updatedPolicies, listErr := dao.ListPoliciesByOrgUnits(ctx, db, tenantId, stored.WorkloadType, relevantOrgUnits,
			dao.ListByOrgUnitOptions{
				IncludeDeleted:  dbmodel.IncludeDeletedPolicies,
				IncludeFallback: false,
				IncludeSystem:   false,
			})
		if listErr != nil {
			return nil, listErr
		}

		if !tree2.IsPolicyChangedDuringCalculation(ctx, appliedPolicy.ID.String(), relevantPolicies, updatedPolicies) {
			updated = true
			break
		}
	}

	if !updated {
		return nil, api.NewError(http.StatusConflict, "policies are currently being updated, try again later", nil)
	}

	reloaded, rerr := dao.FindPolicyByID(ctx, db, policyId, findGvkPolicyOptions)
	if rerr != nil {
		return nil, rerr
	}

	return reloaded.ToGvkPolicyApi()
}

func (pa *PolicyApiService) updateGvkPolicy(ctx context.Context, policyId string, req api.UpdateGvkPolicyReq, options updateOptions) (*api.Policy, *api.Error) {
	db := dao_utils.DB().WithContext(ctx)

	stored, findErr := dao.FindPolicyByID(ctx, db, policyId, findGvkPolicyOptions)
	if findErr != nil {
		return nil, findErr
	}
	if stored == nil || !stored.GvkWorkload {
		return nil, api.NewError(http.StatusNotFound, "policy not found", nil)
	}

	appliedOrgUnit, scopeParams, scopeErr := NewScopeInfoFromPolicy(ctx, &stored.Policy)
	if scopeErr != nil {
		return nil, scopeErr
	}

	rbacErr := auth.ValidateAccessByResourceTypeAndAction(ctx, scopeParams, authApi.RESOURCETYPE_POLICIES, authApi.ACTION_UPDATE, nonPrivate)
	if rbacErr != nil {
		return nil, api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))
	}

	relevantOrgUnits, ancestorErr := appliedOrgUnit.Ancestors(ctx)
	if ancestorErr != nil {
		if ancestorErr.Code == http.StatusNotFound {
			return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("%s does not exist", appliedOrgUnit.Scope), nil)
		}
		return nil, ancestorErr
	}

	gvk, decodeErr := dbmodel.DecodeGvkWorkloadType(stored.WorkloadType)
	if decodeErr != nil {
		return nil, decodeErr
	}

	//
	//   validate the incoming rules up-front (no DB writes here). the
	//   breakdown depends only on the rules + system mapping, so it's the
	//   same across loop attempts and we don't have to redo it inside.
	//
	// update mirrors create: collect into a single shared bucket capped at
	// the first finding so we return early on the first issue from either
	// the breakdown generator or the rule validator.
	issues := newRuleIssues(1)

	breakdownResult, bdErr := pa.GeneratePolicyBreakdown(ctx, db, gvk, req.Rules, issues)
	if bdErr != nil {
		return nil, bdErr
	}
	if e := issues.toError(); e != nil {
		return nil, e
	}

	var breakdown *api.BreakdownBreakdown
	if breakdownResult != nil {
		breakdown = breakdownResult.Breakdown
	}

	if e := pa.validateRules(ctx, breakdown, req.Rules, issues).toError(); e != nil {
		return nil, e
	}

	tenantId := auth.TenantIdOf(ctx)
	subject := auth.SubjectIdOf(ctx)

	updated := false

	// we may need to re-calculate the effective in case one of the given
	// policies has changed during the calculation — mirror the retry loop
	// in createGvkPolicy so concurrent ancestor changes don't leave us with
	// a stale effective.
	for attempt := 0; attempt < tree2.MaxEffectiveCalcAttempts; attempt++ {
		if err := makeSureAllPoliciesExist(ctx, db, stored.WorkloadType, relevantOrgUnits); err != nil {
			return nil, err
		}

		committed := false
		tx := db.Begin()
		defer database.FinalizeTransaction(tx, &committed)

		relevantPolicies, listErr := dao.ListPoliciesByOrgUnits(ctx, tx, tenantId, stored.WorkloadType, relevantOrgUnits,
			dao.ListByOrgUnitOptions{
				IncludeDeleted:  dbmodel.IncludeDeletedPolicies,
				IncludeFallback: false,
				IncludeSystem:   false,
			})
		if listErr != nil {
			return nil, listErr
		}

		appliedPolicy := tree2.FindRelevantPolicy(relevantPolicies, *appliedOrgUnit, stored.WorkloadType)
		if appliedPolicy == nil {
			return nil, api.NewError(http.StatusNotFound, "policy not found", nil)
		}

		//
		//  patch/overwrite the rules to create the updated set of rules.
		//  re-read existing inside the loop so patch-mode (merge) survives
		//  concurrent rule edits.
		//
		existing, fromErr := dbmodel.FromGvkRulesApi[api.PolicyRule](appliedPolicy.Policy.RulesV2)
		if fromErr != nil {
			return nil, fromErr
		}

		var updatedRules []api.PolicyRule

		if options.overwrite {
			updatedRules = req.Rules
		} else if len(req.Rules) != 0 {
			updatedRules = mergeGvkPolicyRulesByRef(existing, req.Rules)
		} else {
			updatedRules = existing
		}

		if len(updatedRules) == 0 {
			return nil, api.NewValidationError("a policy must have at least one rule")
		}

		rulesV2, convErr := dbmodel.ToGvkRulesApi(updatedRules, nil)
		if convErr != nil {
			return nil, convErr
		}

		appliedPolicy.Policy.RulesV2 = rulesV2
		appliedPolicy.Policy.UpdatedAt = time.Now()
		appliedPolicy.UpdatedBy = subject

		if req.Description != nil {
			appliedPolicy.Name = *req.Description
		}

		// validation passed, set status to Ready with no issues
		readyPhase := api.POLICYPHASE_READY
		appliedPolicy.ValidationStatus = &readyPhase
		appliedPolicy.ValidationIssues = []string{}

		if upErr := dao.UpdatePolicySpec(ctx, tx, &appliedPolicy.Policy); upErr != nil {
			return nil, upErr
		}

		// calculate the effective policy with the full relevant chain so
		// ancestor rules are honored.
		effectiveCtx := tree2.NewEffectiveContext(ctx, tree2.WorkloadType(stored.WorkloadType), relevantOrgUnits, relevantPolicies)

		effectivePolicy, effErr := tree2.CalculateEffectiveOfGvkPolicy(ctx, tx, effectiveCtx, appliedPolicy.ID.String())
		if effErr != nil {
			return nil, effErr
		}

		appliedPolicy.EffectivePolicy = *effectivePolicy

		descendantIds, sdErr := dao.SoftDeleteEffectivePoliciesOfSubtree(ctx, tx, *appliedOrgUnit, stored.WorkloadType)
		if sdErr != nil {
			return nil, sdErr
		}
		if err := dao.DeleteBreakdownCachesOfSubtree(ctx, tx, descendantIds); err != nil {
			return nil, err
		}

		// refresh this policy's own cache atomically with the spec update
		if err := dao.OnEntityDeleted(ctx, tx, dbmodel.EntityTypeBreakdown, policyId); err != nil {
			return nil, err
		}
		if breakdownResult != nil {
			if _, cacheErr := dao.CacheBreakdown(ctx, tx, policyId, breakdownResult.Breakdown, breakdownResult.Dependencies); cacheErr != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to cache breakdown for policy %s: %v", policyId, cacheErr)
				return nil, cacheErr
			}
		}

		result := tx.Commit()
		if result.Error != nil {
			return nil, database.Error(database.UPDATE, database.Policies).WithDetailsFromDb(result)
		}
		committed = true

		updatedPolicies, listErr := dao.ListPoliciesByOrgUnits(ctx, db, tenantId, stored.WorkloadType, relevantOrgUnits,
			dao.ListByOrgUnitOptions{
				IncludeDeleted:  dbmodel.IncludeDeletedPolicies,
				IncludeFallback: false,
				IncludeSystem:   false,
			})
		if listErr != nil {
			return nil, listErr
		}

		if !tree2.IsPolicyChangedDuringCalculation(ctx, appliedPolicy.ID.String(), relevantPolicies, updatedPolicies) {
			updated = true
			break
		}
	}

	if !updated {
		return nil, api.NewError(http.StatusConflict, "policies are currently being updated, try again later", nil)
	}

	reloaded, rerr := dao.FindPolicyByID(ctx, db, policyId, findGvkPolicyOptions)
	if rerr != nil {
		return nil, rerr
	}

	return reloaded.ToGvkPolicyApi()
}

// stringSlicesEqual compares two string slices for equality.
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// makeSureAllPoliciesExist ensures every org unit in relevantOrgUnits has a
// matching (possibly soft-deleted) policy row for the given workloadType
// before we calculate the effective. Missing rows are inserted as soft-deleted
// placeholders so the effective-calculation path can rely on a stable
// hierarchy.
func makeSureAllPoliciesExist(ctx context.Context, db *gorm.DB, workloadType string, relevantOrgUnits []scopes.OrgUnit) *api.Error {
	emptyRules := dao_utils.JsonBMap{}

	for _, orgUnit := range relevantOrgUnits {
		if orgUnit.Scope == api.SCOPETYPE_SYSTEM {
			continue
		}

		orgUnitRequest := api.CreateGvkPolicyReq{
			Scope: orgUnit.Scope,
		}

		switch orgUnit.Scope {
		case api.SCOPETYPE_CLUSTER:
			orgUnitRequest.ClusterId = convutils.Ref(orgUnit.Id)
		case api.SCOPETYPE_DEPARTMENT:
			orgUnitRequest.DepartmentId = convutils.Ref(orgUnit.Id)
		case api.SCOPETYPE_PROJECT:
			orgUnitRequest.ProjectId = convutils.Ref(orgUnit.Id)
		}

		var policyRow dbmodel.PolicyWithOptionalEffective
		if err := applyCreateReqOnPolicy(ctx, &policyRow, orgUnitRequest, workloadType, emptyRules); err != nil {
			return err
		}

		// we first want to add the policy as deleted
		policyRow.Policy.DeletedAt = gorm.DeletedAt{
			Time:  time.Now(),
			Valid: true,
		}

		// attempt to add and ignore conflicts from concurrent callers that may
		// have just inserted the same placeholder row
		if dbErr := dao.AddPolicy(ctx, db, &policyRow.Policy); dbErr != nil && dbErr.Code != http.StatusConflict {
			return dbErr
		}
	}
	return nil
}
