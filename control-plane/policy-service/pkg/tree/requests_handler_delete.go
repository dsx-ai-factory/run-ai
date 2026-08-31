/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tree

import (
	"context"
	"net/http"

	_ "github.com/google/uuid"

	"github.com/run-ai/runai/control-plane/policy-service/pkg/database"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/scopes"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"gorm.io/gorm"
)

// PolicyNotFoundErrMsg is the error message for a request to delete a policy that doesn't exist
const PolicyNotFoundErrMsg = "requested policy does not exist"

// DeleteRequestOpt holds options for delete policy request
type DeleteRequestOpt struct {
	// DeleteSubTree indicates that sibling policies should be deleted as well
	DeleteSubTree bool
	// SkipAuthorization indicates whether to skip authorization checks during deletion
	SkipAuthorization bool
}

func (ht *HierarchyTree) HandleDeleteRequest(ctx context.Context, request *GenericPolicyRequest, opt DeleteRequestOpt) *api.Error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("[%s] delete %s policy started", ht.treeId.ToString(), request.WorkloadType)

	db := dao_utils.DB().WithContext(ctx)

	// attempt to find the policy in the DB
	policy, err := dao.FindPolicyOfOrgUnit(ctx, db, request.PolicyMeta.Scope, request.PolicyMeta.DepartmentId,
		request.PolicyMeta.ProjectId, request.PolicyMeta.ClusterId, string(request.WorkloadType), dbmodel.NotToIncludeDeletedPolicies)
	if err != nil {
		return err
	}
	if policy == nil {
		return api.NewError(http.StatusNotFound, PolicyNotFoundErrMsg, nil)
	}

	orgUnit, err := scopes.MakeOrgUnit(request.PolicyMeta.Scope, request.TenantId, request.PolicyMeta.ClusterId, request.PolicyMeta.DepartmentId, request.PolicyMeta.ProjectId)
	if err != nil {
		return err
	}

	// build array containing the relevant org units (the one that the policy is applied on, and all its ancestors).
	// the first org unit is the one that the given policy is being applied on
	relevantOrgUnits, err := orgUnit.Ancestors(ctx)
	if err != nil {
		if err.Code == http.StatusNotFound {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof(
				"[%s] scope entity no longer exists, performing orphaned policy cleanup for policy %s",
				ht.treeId.ToString(), policy.ID.String())
			return ht.handleOrphanedPolicyDelete(ctx, db, policy, orgUnit, request.WorkloadType)
		}
		return err
	}

	hierarchyErr := ht.makeSureAllPoliciesExist(ctx, request, relevantOrgUnits)
	if hierarchyErr != nil {
		return hierarchyErr
	}

	// we may need to re-calculate the effective in case one of the given policies has changed during the calculation
	for attempt := 0; attempt < MaxEffectiveCalcAttempts; attempt++ {

		// we have multiple updates in the db, we want them as single transaction
		committed := false
		tx := db.Begin()
		defer database.FinalizeTransaction(tx, &committed)

		//
		//   soft delete the policies
		//
		err = dao.SoftDeleteEffectivePoliciesByOrgUnit(ctx, tx, orgUnit, string(request.WorkloadType))
		if err != nil {
			return err
		}

		err = dao.SoftDeletePolicyById(ctx, tx, &policy.Policy)
		if err != nil {
			return err
		}

		//
		//   list the relevant policies, both specific workload type and fallback
		//
		relevantPolicies, err := dao.ListPoliciesByOrgUnits(ctx, tx, request.TenantId, string(request.WorkloadType), relevantOrgUnits,
			dao.ListByOrgUnitOptions{
				IncludeDeleted:  dbmodel.IncludeDeletedPolicies,
				IncludeFallback: true,
				IncludeSystem:   true,
			})
		if err != nil {
			return err
		}

		// legacy v1/v2 path: discard the descendant ids — only the GVK
		// flows use them for breakdown-cache invalidation.
		_, err = dao.SoftDeleteEffectivePoliciesOfSubtree(ctx, tx, orgUnit, string(request.WorkloadType))
		if err != nil {
			return err
		}

		// in the past, we used to calculate the effective policy of the current org unit.
		// the purpose was that the sync mechanism will take it.
		// now we no longer sync, and the calculation just upsets the sync mock that think
		// that the effective policy was updated. so no need to calculate it now, it will be
		// re-calculated on the next update/get

		result := tx.Commit()
		if result.Error != nil {
			return database.Error(database.DELETE, database.Policies).WithDetailsFromDb(result)
		}
		committed = true

		updatedPolicies, err := dao.ListPoliciesByOrgUnits(ctx, db, request.TenantId, string(request.WorkloadType), relevantOrgUnits,
			dao.ListByOrgUnitOptions{
				IncludeDeleted:  dbmodel.IncludeDeletedPolicies,
				IncludeFallback: true,
				IncludeSystem:   true,
			})
		if err != nil {
			return err
		}

		if !IsPolicyChangedDuringCalculation(ctx, policy.ID.String(), relevantPolicies, updatedPolicies) {
			return nil
		}
	}

	return api.NewError(http.StatusServiceUnavailable, "policy is currently being updated, retry your request later", nil)
}

// handleOrphanedPolicyDelete performs a simplified delete for policies whose scope entity
// (e.g. department, project) has been deleted from the backend. Since the scope entity no
// longer exists, we cannot resolve ancestors or recalculate the hierarchy — we just
// soft-delete the policy and its effective policies directly.
func (ht *HierarchyTree) handleOrphanedPolicyDelete(ctx context.Context, db *gorm.DB, policy *dbmodel.PolicyWithOptionalEffective, orgUnit scopes.OrgUnit, workloadType WorkloadType) *api.Error {
	committed := false
	tx := db.Begin()
	defer database.FinalizeTransaction(tx, &committed)

	err := dao.SoftDeleteEffectivePoliciesByOrgUnit(ctx, tx, orgUnit, string(workloadType))
	if err != nil {
		return err
	}

	err = dao.SoftDeletePolicyById(ctx, tx, &policy.Policy)
	if err != nil {
		return err
	}

	// legacy v1/v2 path: discard the descendant ids — only the GVK
	// flows use them for breakdown-cache invalidation.
	_, err = dao.SoftDeleteEffectivePoliciesOfSubtree(ctx, tx, orgUnit, string(workloadType))
	if err != nil {
		return err
	}

	result := tx.Commit()
	if result.Error != nil {
		return database.Error(database.DELETE, database.Policies).WithDetailsFromDb(result)
	}
	committed = true

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof(
		"[%s] orphaned policy %s successfully deleted", ht.treeId.ToString(), policy.ID.String())
	return nil
}
