/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tree

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"

	"gorm.io/gorm"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/compliance"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/scopes"
	policyCompliance "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/policy_compliance"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/conjungo"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
)

const (
	maxMissingIdsPerErrorMessage = 5
	errorDelimiter               = ", "

	MaxEffectiveCalcAttempts = 5
)

type ApplyPolicyRequest struct {
	GenericPolicyRequest
	Defaults      map[string]interface{}
	Rules         map[string]interface{}
	ImposedAssets *api.DistributedImposedAssets
	FieldsToReset []string
}

func (apr *ApplyPolicyRequest) HasImposedAssets() bool {
	if apr.ImposedAssets == nil {
		return false
	}
	if apr.ImposedAssets.Master != nil && len(*apr.ImposedAssets.Master) > 0 {
		return true
	}
	if apr.ImposedAssets.Worker != nil && len(*apr.ImposedAssets.Worker) > 0 {
		return true
	}
	return false
}

type ApplyPolicyOptions struct {
	ValidateOnly bool
	Overwrite    bool

	// old cluster 2.20 and below require cluster deletion approval of corresponding policies
	// this is done to avoid conflicts b/w control pland and cluster policies
	ApproveClusterDeletion bool
}

func (ht *HierarchyTree) HandleApplyRequest(ctx context.Context, request *ApplyPolicyRequest, options ApplyPolicyOptions) (*dbmodel.PolicyWithOptionalEffective, *api.Error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("[%s] apply %s policy started validate only? [%v] overwrite? [%v]", ht.treeId.ToString(), request.WorkloadType, options.ValidateOnly, options.Overwrite)

	if err := gpucompat.ValidateMemoryRequestInput(request.Defaults); err != nil {
		return nil, api.NewValidationError(err.Error())
	}
	if aliasErr := normalizeGPUPolicyRequest(request); aliasErr != nil {
		return nil, aliasErr
	}

	request.Defaults = database.CleanMap(request.Defaults)
	request.Rules = database.CleanMap(request.Rules)

	err := handleClusterDeletionApproval(ctx, &request.GenericPolicyRequest)
	if err != nil {
		return nil, err
	}

	appliedOrgUnit, err := scopes.MakeOrgUnit(request.PolicyMeta.Scope, request.TenantId, request.PolicyMeta.ClusterId, request.PolicyMeta.DepartmentId, request.PolicyMeta.ProjectId)
	if err != nil {
		return nil, err
	}

	// build array containing the relevant org units (the one that the policy is applied on, and all its ancestors).
	// the first org unit is the one that the given policy is being applied on
	relevantOrgUnits, err := appliedOrgUnit.Ancestors(ctx)
	if err != nil {
		return nil, err
	}

	var imposedAssetsWorker *[]string
	var imposedAssetsMaster *[]string

	if request.ImposedAssets != nil {
		imposedAssetsWorker = request.ImposedAssets.Worker
		imposedAssetsMaster = request.ImposedAssets.Master
	}

	imposedAssets, imposedAssetsRules, imposedErr := ht.validateImposedAssets(ctx, appliedOrgUnit, imposedAssetsWorker)
	if imposedErr != nil {
		return nil, imposedErr
	}

	masterImposedAssets, masterImposedAssetsRules, imposedErr := ht.validateImposedAssets(ctx, appliedOrgUnit, imposedAssetsMaster)
	if imposedErr != nil {
		return nil, imposedErr
	}

	hierarchyErr := ht.makeSureAllPoliciesExist(ctx, &request.GenericPolicyRequest, relevantOrgUnits)
	if hierarchyErr != nil {
		return nil, hierarchyErr
	}

	db := runaidb.DB().WithContext(ctx)

	// we may need to re-calculate the effective in case one of the given policies has changed during the calculation
	for attempt := 0; attempt < MaxEffectiveCalcAttempts; attempt++ {
		// now we're ready to update the database
		committed := false
		tx := db.Begin()
		defer database.FinalizeTransaction(tx, &committed)

		// list the policy of all the relevant nodes (applied node + all its ancestors)
		relevantPolicies, err := dao.ListPoliciesByOrgUnits(ctx, tx, request.TenantId, string(request.WorkloadType), relevantOrgUnits,
			dao.ListByOrgUnitOptions{
				IncludeDeleted:  dbmodel.IncludeDeletedPolicies,
				IncludeFallback: true,
				IncludeSystem:   true,
			})
		if err != nil {
			return nil, err
		}

		appliedPolicy := FindRelevantPolicy(relevantPolicies, appliedOrgUnit, string(request.WorkloadType))

		validationErrors := compliance.ValidateSubmittedPolicy(request.Rules, request.Defaults)
		if len(validationErrors) != 0 {
			return nil, api.NewError(http.StatusBadRequest, validationErrors.String(), nil)
		}

		// if this is the first time we apply a policy to this org unit, initialize it. otherwise
		// we have to update the defaults and rules according to the request
		if appliedPolicy == nil {
			appliedPolicy = &dbmodel.PolicyWithOptionalEffective{
				Policy: dbmodel.Policy{
					BaseModel: dbmodels.BaseModel{
						CreatedBy: request.Requester,
						UpdatedBy: request.Requester,
					},
					Name:          request.PolicyMeta.Name,
					WorkloadType:  string(request.WorkloadType),
					Scope:         request.PolicyMeta.Scope,
					DefaultsV2:    request.Defaults,
					RulesV2:       request.Rules,
					ImposedAssets: request.ImposedAssets,
				},
			}

			patchOrgUnits(ctx, &appliedPolicy.Policy, relevantOrgUnits)

			if len(convutils.MapMe(appliedPolicy.Policy.DefaultsV2.AsMAp())) == 0 && len(convutils.MapMe(appliedPolicy.Policy.RulesV2.AsMAp())) == 0 && !request.HasImposedAssets() {
				return nil, api.NewError(http.StatusBadRequest, "To apply a policy, there must be at least one default or rule", nil)
			}

			addError := dao.AddPolicy(ctx, tx, &appliedPolicy.Policy)
			if addError != nil {
				return nil, addError
			}

			relevantPolicies[appliedPolicy.ID.String()] = appliedPolicy
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("[%s] new policy created (id = %s)", ht.treeId.ToString(), appliedPolicy.ID)
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("[%s] policy already exists (id = %s), update it", ht.treeId.ToString(), appliedPolicy.ID)

			isPlaceholderPolicy := !appliedPolicy.HasContent()
			// if the existing policy is a placeholder policy OR the policy is flagged as deleted, we treat it as a first time creation
			if isPlaceholderPolicy || appliedPolicy.Policy.IsDeleted() {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("[%s] existing policy (id = %s) is an empty policy placeholder, updating creator and timestamp", ht.treeId.ToString(), appliedPolicy.ID.String())
				appliedPolicy.CreatedBy = request.Requester
				appliedPolicy.UpdatedBy = request.Requester
				appliedPolicy.CreatedAt = time.Now()
				appliedPolicy.Policy.UpdatedAt = time.Now()
			}

			appliedPolicy.Name = request.PolicyMeta.Name

			// revert the policy deletion flag, even if the policy is deleted we are now re-creating it
			appliedPolicy.Policy.DeletedAt = gorm.DeletedAt{}

			if appliedPolicy.PolicyDeleted.Valid {
				curTime := time.Now()
				appliedPolicy.CreatedAt = curTime
				appliedPolicy.Policy.UpdatedAt = curTime
			}

			if options.Overwrite {
				appliedPolicy.UpdatedBy = request.Requester
				appliedPolicy.Policy.DefaultsV2 = request.Defaults
				appliedPolicy.Policy.RulesV2 = request.Rules
				appliedPolicy.Policy.ImposedAssets = request.ImposedAssets
			} else {
				err := applyRequestedChangesOnGivenPolicy(appliedPolicy, request)
				if err != nil {
					return nil, err
				}

				// in patch, we only update the imposedAssets list if it was implicitly sent
				// in the request. otherwise it remains unchanged
				if request.ImposedAssets != nil {
					appliedPolicy.Policy.ImposedAssets = request.ImposedAssets
				}
			}

			if len(convutils.MapMe(appliedPolicy.Policy.DefaultsV2.AsMAp())) == 0 && len(convutils.MapMe(appliedPolicy.Policy.RulesV2.AsMAp())) == 0 && !request.HasImposedAssets() {
				return nil, api.NewError(http.StatusBadRequest,
					"To apply a policy, there must be at least one default or rule", nil)
			}

			err = dao.UpdatePolicySpec(ctx, tx, &appliedPolicy.Policy)
			if err != nil {
				return nil, err
			}
		}

		// calculate the effective policy of the org unit for which we apply the policy
		effectiveContext := context.WithValue(ctx, EffectiveContextKey, NewEffectiveContext(ctx, request.WorkloadType, relevantOrgUnits, relevantPolicies))

		effectivePolicy, err := ht.calculateEffectivePolicy(effectiveContext, tx, appliedPolicy.ID.String(), 0, imposedAssets, masterImposedAssets, imposedAssetsRules, masterImposedAssetsRules)
		if err != nil {
			return nil, err
		}

		//  validation errors should reject the apply request. this is different from get, that
		//  can have defaultFrom issues which will appear in the payload but will not fail the get
		if effectivePolicy.ValidationError != nil {
			return nil, api.NewError(http.StatusBadRequest, *effectivePolicy.ValidationError, nil)
		}

		translationErr := compliance.ValidateEffectivePolicy(ctx, string(request.WorkloadType), effectivePolicy.RulesV2, effectivePolicy.DefaultsV2)
		if translationErr != nil {
			return nil, api.NewError(int(translationErr.Code), translationErr.Message, nil)
		}

		appliedPolicy.EffectivePolicy = *effectivePolicy

		_, complianceErrors := compliance.ValidateDefaults(appliedPolicy)
		if complianceErrors != nil {
			return nil, api.NewError(http.StatusBadRequest, complianceErrors.String(), nil)
		}

		// no more validations from this point, so if the request was to validate only - time to quit
		if options.ValidateOnly {
			return nil, nil
		}

		// legacy v1/v2 path: discard the descendant ids — only the GVK
		// flows use them for breakdown-cache invalidation.
		_, err = dao.SoftDeleteEffectivePoliciesOfSubtree(ctx, tx, appliedOrgUnit, string(request.WorkloadType))
		if err != nil {
			return nil, err
		}

		commitErr := tx.Commit()
		if commitErr != nil {
			return nil, database.Error(database.UPSERT, database.Policies).WithDetails(commitErr)
		}
		committed = true

		updatedPolicies, err := dao.ListPoliciesByOrgUnits(ctx, db, request.TenantId, string(request.WorkloadType), relevantOrgUnits,
			dao.ListByOrgUnitOptions{
				IncludeDeleted:  dbmodel.IncludeDeletedPolicies,
				IncludeFallback: true,
				IncludeSystem:   true,
			})

		if err != nil {
			return nil, err
		}

		if !IsPolicyChangedDuringCalculation(ctx, appliedPolicy.ID.String(), relevantPolicies, updatedPolicies) {
			return appliedPolicy, nil
		}
	}

	return nil, api.NewError(http.StatusServiceUnavailable, "policy is currently being updated, retry your request later", nil)
}

// applyRequestedChangesOnGivenPolicy receives an existing policy and a request to change it. the request can hold
// defaults and rules to add/update, and also list of fields to remove from the policy. the function updates the existing
// policy with those changes
func applyRequestedChangesOnGivenPolicy(existingPolicy *dbmodel.PolicyWithOptionalEffective, request *ApplyPolicyRequest) *api.Error {
	if err := existingPolicy.NormalizeGPUMemory(); err != nil {
		return api.NewError(http.StatusBadRequest, err.Error(), err)
	}
	if err := normalizeGPUPolicyRequest(request); err != nil {
		return err
	}
	existingPolicy.UpdatedBy = request.Requester

	// Before merging - clean all the fields that we intend to apply. this is needed only for itemized fields
	// in order to clear all the items from the policy and replace them with a fresh copy from the request
	removeFields(existingPolicy.Policy.DefaultsV2, policyCompliance.FieldsOfDefaults(request.Defaults))
	removeFields(existingPolicy.Policy.RulesV2, policyCompliance.FieldsWithRules(request.Rules))

	err := conjungo.Merge(&existingPolicy.Policy.DefaultsV2, dbmodels.JsonBMap(request.Defaults), conjungo.NewOptions())
	if err != nil {
		return api.NewError(http.StatusServiceUnavailable, "failed to apply the new defaults", err)
	}

	err = conjungo.Merge(&existingPolicy.Policy.RulesV2, dbmodels.JsonBMap(request.Rules), conjungo.NewOptions())
	if err != nil {
		return api.NewError(http.StatusServiceUnavailable, "failed to apply the new rules", err)
	}

	removeFields(existingPolicy.Policy.DefaultsV2, request.FieldsToReset)
	removeFields(existingPolicy.Policy.RulesV2, request.FieldsToReset)

	existingPolicy.Policy.DefaultsV2 = database.CleanMap(existingPolicy.Policy.DefaultsV2)
	existingPolicy.Policy.RulesV2 = database.CleanMap(existingPolicy.Policy.RulesV2)

	return nil
}

func removeFields(map1 map[string]interface{}, fieldsToRemove []string) {
	for _, fld := range fieldsToRemove {
		removeField(map1, fld)
	}
}

func removeField(map1 map[string]interface{}, fieldToRemove string) {
	fieldParts := strings.Split(fieldToRemove, ".")
	for index, fieldPart := range fieldParts {
		if index == len(fieldParts)-1 {
			delete(map1, fieldPart)
			return
		}
		var ok bool
		if map1, ok = map1[fieldPart].(map[string]interface{}); !ok {
			return
		}
	}
}

func FindRelevantPolicy(relevantPolicies map[string]*dbmodel.PolicyWithOptionalEffective, orgUnit scopes.OrgUnit, workloadType string) *dbmodel.PolicyWithOptionalEffective {
	for _, policy := range relevantPolicies {
		if policy.WorkloadType == workloadType && policy.OrgUnit() == orgUnit {
			return policy
		}
	}

	return nil
}

func patchOrgUnits(ctx context.Context, policy *dbmodel.Policy, orgUnit []scopes.OrgUnit) {
	for _, orgUnit := range orgUnit {
		switch orgUnit.Scope {
		case api.SCOPETYPE_TENANT:
			tenantId, err := strconv.ParseInt(orgUnit.Id, 10, 64)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to parse tenant id %s", orgUnit.Id)
			} else {
				policy.TenantId = int32(tenantId)
			}

		case api.SCOPETYPE_CLUSTER:
			policy.ClusterId = convutils.Ref(orgUnit.Id)

		case api.SCOPETYPE_DEPARTMENT:
			policy.DepartmentId = convutils.Ref(orgUnit.Id)

		case api.SCOPETYPE_PROJECT:
			projectId, err := strconv.ParseInt(orgUnit.Id, 10, 64)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to parse project id %s", orgUnit.Id)
			} else {
				policy.ProjectId = convutils.Ref(int32(projectId))
			}
		}
	}
}

func IsPolicyChangedDuringCalculation(ctx context.Context, appliedPolicyId string, originalSet map[string]*dbmodel.PolicyWithOptionalEffective, newSet map[string]*dbmodel.PolicyWithOptionalEffective) bool {
	if len(originalSet) != len(newSet) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("recalc: number of relevant policies changed from %d to %d", len(originalSet), len(newSet))
		return true
	}

	curTime := time.Now()

	for id, origPolicy := range originalSet {
		// no sense to check the applied policy, because naturaly it has changed because my own calculation
		if id == appliedPolicyId {
			continue
		}

		newPolicy, ok := newSet[id]
		if !ok {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("recalc: policy %s is no longer found", id)
			return true
		}

		if convutils.Deref(newPolicy.DeletedAt, curTime) != convutils.Deref(origPolicy.DeletedAt, curTime) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("recalc: policy %s deletion date updated", id)
			return true
		}

		if newPolicy.UpdatedAt != origPolicy.UpdatedAt {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("recalc: policy %s updated", id)
			return true
		}
	}

	return false
}
