package tree

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/scopes"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

func (ht *HierarchyTree) HandleGetRequest(ctx context.Context, request *GenericPolicyRequest) (*dbmodel.PolicyWithOptionalEffective, *api.Error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("[%s] get %s policy started", ht.treeId.ToString(), request.WorkloadType)

	db := dao_utils.DB().WithContext(ctx)

	orgUnit, err := scopes.MakeOrgUnit(request.PolicyMeta.Scope, request.TenantId, request.PolicyMeta.ClusterId, request.PolicyMeta.DepartmentId, request.PolicyMeta.ProjectId)
	if err != nil {
		return nil, err
	}

	var policyAndEffective dbmodel.PolicyWithOptionalEffective

	// given policy can be deleted or not exist
	givenPolicy, err := dao.FindGivenPolicyOfOrgUnit(ctx, db, request.PolicyMeta.Scope, request.PolicyMeta.ClusterId,
		request.PolicyMeta.DepartmentId, request.PolicyMeta.ProjectId, string(request.WorkloadType), dbmodel.IncludeDeletedPolicies)
	if err != nil {
		return nil, err
	}

	if givenPolicy != nil {
		policyAndEffective.Policy = *givenPolicy

		// if the effective policy was already calculated, load it from the db. no need for locking or anything
		effectivePolicy, err := dao.GetEffectivePolicyById(ctx, givenPolicy.ID, dbmodel.IncludeDeletedPolicies)
		if err != nil {
			return nil, err
		}

		// since defaultFrom is a new feature, it is possible that effective records of old records is missing from
		// in the table, and we have to calculate it here. this will continue until the policy is updated, unfortunately.
		if effectivePolicy != nil && effectivePolicy.DefaultFromOrder == nil {
			effectivePolicy.DefaultFromOrder, effectivePolicy.ValidationError = calculateDefaultFromOrder(effectivePolicy.RulesV2)
		}

		if effectivePolicy != nil && !effectivePolicy.IsDeleted() {
			policyAndEffective.EffectivePolicy = *effectivePolicy
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("[%s] found effective policy in db, policy id %s", ht.treeId.ToString(), givenPolicy.ID.String())
			return &policyAndEffective, nil
		}
	}

	// build array containing the relevant org units (the one that the policy is applied on, and all its ancestors).
	// the first org unit is the one that the given policy is being applied on
	relevantOrgUnits, err := orgUnit.Ancestors(ctx)
	if err != nil {
		return nil, err
	}

	hierarchyErr := ht.makeSureAllPoliciesExist(ctx, request, relevantOrgUnits)
	if hierarchyErr != nil {
		return nil, hierarchyErr
	}

	// we may need to re-calculate the effective in case one of the given policies has changed during the calculation
	for attempt := 0; attempt < MaxEffectiveCalcAttempts; attempt++ {
		// during the derivation of the effective policy we update the DB, so we have to start transaction
		committed := false
		tx := db.Begin()
		defer database.FinalizeTransaction(tx, &committed)

		if givenPolicy == nil {
			// the policy is now expected to exist due to MakeSureAllPoliciesExist just load it again
			givenPolicy, err = dao.FindGivenPolicyOfOrgUnit(ctx, db, request.PolicyMeta.Scope, request.PolicyMeta.ClusterId,
				request.PolicyMeta.DepartmentId, request.PolicyMeta.ProjectId, string(request.WorkloadType), dbmodel.IncludeDeletedPolicies)
			if err != nil {
				return nil, api.NewError(http.StatusInternalServerError, "failed to initialize policy", fmt.Errorf("%s: %v", err.Message, err.Details))
			}

			// this is not likely to happen, but just to be on the safe side check it...
			if givenPolicy == nil {
				return nil, api.NewError(http.StatusInternalServerError, "failed to locate policy", fmt.Errorf("%s: %v", err.Message, err.Details))
			}

			policyAndEffective.Policy = *givenPolicy
		}

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

		// process imposed assets if the policy has them
		var imposedAssetsWorker *[]string
		var imposedAssetsMaster *[]string
		if givenPolicy.ImposedAssets != nil {
			imposedAssetsWorker = givenPolicy.ImposedAssets.Worker
			imposedAssetsMaster = givenPolicy.ImposedAssets.Master
		}

		imposedAssets, imposedAssetsRules, imposedErr := ht.validateImposedAssets(ctx, orgUnit, imposedAssetsWorker)
		if imposedErr != nil {
			return nil, imposedErr
		}

		masterImposedAssets, masterImposedAssetsRules, imposedErr := ht.validateImposedAssets(ctx, orgUnit, imposedAssetsMaster)
		if imposedErr != nil {
			return nil, imposedErr
		}

		// calculate the effective policy of the relevant org unit
		effectiveContext := context.WithValue(ctx, EffectiveContextKey, NewEffectiveContext(ctx, request.WorkloadType, relevantOrgUnits, relevantPolicies))

		effectivePolicy, err := ht.calculateEffectivePolicy(effectiveContext, tx, givenPolicy.ID.String(), 0, imposedAssets, masterImposedAssets, imposedAssetsRules, masterImposedAssetsRules)
		if err != nil {
			return nil, err
		}
		policyAndEffective.EffectivePolicy = *effectivePolicy

		result := tx.Commit()
		if result.Error != nil {
			return nil, database.Error(database.UPSERT, database.Policies).WithDetailsFromDb(result)
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

		if !IsPolicyChangedDuringCalculation(ctx, givenPolicy.ID.String(), relevantPolicies, updatedPolicies) {
			return &policyAndEffective, nil
		}
	}

	return nil, api.NewError(http.StatusServiceUnavailable, "policy is currently being updated, retry your request later", nil)
}
