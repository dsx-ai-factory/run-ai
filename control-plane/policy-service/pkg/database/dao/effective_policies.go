package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/scopes"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/config"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"gorm.io/gorm"
)

// UpsertEffectivePolicy creates-or-updates an effective policy record by id.
// --- NOTE ---
// This function was written in sweat, blood and tears. You may look at it and say
// the logic can be replaced with a single call to FirstAndCreate, well - don't.
// The Updating logic is very sensitive, and in this specific function we prefer to do
// any step ourselves, with full control of what we're doing and not to trust GORM to
// understand what we want it to do, because it typically does not understand. At least
// not in this function. Go find some other places in the code to improve.
func UpsertEffectivePolicy(ctx context.Context, tx *gorm.DB, effectivePolicy *dbmodel.EffectivePolicy) *api.Error {
	//
	// prepare the effective policy and insert it
	//
	effectivePolicy.DeletedAt = nil
	effectivePolicy.UpdatedAt = time.Now()

	effectivePolicy.DefaultsV2 = database.CleanMap(effectivePolicy.DefaultsV2)
	effectivePolicy.RulesV2 = database.CleanMap(effectivePolicy.RulesV2)

	//
	//   first step - see if we need to create or update. yes, it reminds you of firstOrCreate,
	//   see my note above the function.
	//
	var existingRecord dbmodel.EffectivePolicy

	result := tx.Table(dbtables.EffectivePoliciesTable).
		Where(fmt.Sprintf("%s.%s = ?", dbtables.EffectivePoliciesTable, dbtables.EffectivePoliciesColumnPolicyId), effectivePolicy.PolicyId).
		First(&existingRecord)

	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		dbErr := database.Error(database.UPSERT, database.EffectivePolicy).WithDetailsFromDb(result)
		logging.Errorf("failed to upsert effective policy: %s: %v", dbErr.Message, result.Error)
		return dbErr
	}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		//
		//  if the record does not exist, meaning - hard deleted, create a new one
		//
		result = tx.Table(dbtables.EffectivePoliciesTable).Create(&effectivePolicy)
		if result.Error != nil {
			dbErr := database.Error(database.CREATE, database.EffectivePolicy).WithDetailsFromDb(result)
			logging.Errorf("failed to create effective policy: %s: %v", dbErr.Message, result.Error)
			return dbErr
		}
	} else {
		//
		//   if it is soft deleted - wake it up, don't trust GORM to do it for you
		//
		result = tx.Table(dbtables.EffectivePoliciesTable).
			Where(fmt.Sprintf("%s.%s = ?", dbtables.EffectivePoliciesTable, dbtables.EffectivePoliciesColumnPolicyId), effectivePolicy.PolicyId).
			UpdateColumns(map[string]interface{}{
				dbtables.DeletedAtColumn:             nil,
				dbtables.PoliciesColumnDefaults:      effectivePolicy.DefaultsV2,
				dbtables.PoliciesColumnRules:         effectivePolicy.RulesV2,
				dbtables.PoliciesColumnImposedAssets: effectivePolicy.ImposedAssets,

				dbtables.EffectivePoliciesColumnDefaultFromOrder: effectivePolicy.DefaultFromOrder,
				dbtables.EffectivePoliciesColumnValidationError:  effectivePolicy.ValidationError,
			})
		if result.Error != nil {
			dbErr := database.Error(database.UPDATE, database.EffectivePolicy).WithDetailsFromDb(result)
			logging.Errorf("failed to update effective policy: %s: %v", dbErr.Message, result.Error)
			return dbErr
		}
	}

	return nil
}

func GetEffectivePolicyById(ctx context.Context, id uuid.UUID, includeDeleted bool) (*dbmodel.EffectivePolicy, *api.Error) {
	effectivePolicy := dbmodel.EffectivePolicy{
		TenantId: auth.TenantIdOf(ctx),
		PolicyId: id,
	}

	var record dbmodel.EffectivePolicy

	result := database.DB().Table(dbtables.EffectivePoliciesTable).Where(&effectivePolicy)
	if !includeDeleted {
		result.Where(fmt.Sprintf("%s.%s IS NULL", dbtables.EffectivePoliciesTable, dbtables.DeletedAtColumn))
	} else {
		result.Unscoped()
	}

	result = result.Scan(&record)
	if result.Error != nil {
		dbErr := database.Error(database.GET, database.EffectivePolicy).WithDetailsFromDb(result)
		logging.Errorf("failed to get effective policy by id: %s: %v", dbErr.Message, result.Error)
		return nil, dbErr
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}

	record.DefaultsV2 = database.CleanMap(record.DefaultsV2)
	record.RulesV2 = database.CleanMap(record.RulesV2)

	return &record, nil
}

func SoftDeleteEffectivePoliciesByOrgUnit(ctx context.Context, tx *gorm.DB, appliedOrgUnit scopes.OrgUnit, workloadType string) *api.Error {
	result := tx.Table(dbtables.EffectivePoliciesTable)

	sql := fmt.Sprintf("%s.%s in (select %s from %s where %s.%s = %s.%s",
		dbtables.EffectivePoliciesTable, dbtables.EffectivePoliciesColumnPolicyId,
		dbtables.PoliciesColumnId, dbtables.PoliciesTable,
		dbtables.PoliciesTable, dbtables.PoliciesColumnId,
		dbtables.EffectivePoliciesTable, dbtables.EffectivePoliciesColumnPolicyId)

	params := make([]interface{}, 0)

	if workloadType != FallbackPolicyAsType {
		sql += fmt.Sprintf(" and %s.%s = ?", dbtables.PoliciesTable, dbtables.PoliciesColumnWorkloadType)
		params = append(params, workloadType)
	}

	switch appliedOrgUnit.Scope {
	case api.SCOPETYPE_PROJECT:
		sql += fmt.Sprintf(" and %s.%s = ? and %s.%s = ?",
			dbtables.PoliciesTable, dbtables.PoliciesColumnScope,
			dbtables.PoliciesTable, dbtables.PoliciesColumnProjectId)

		params = append(params, api.SCOPETYPE_PROJECT)
		params = append(params, appliedOrgUnit.Id)

	case api.SCOPETYPE_DEPARTMENT:
		sql += fmt.Sprintf(" and %s.%s = ? and %s.%s = ?",
			dbtables.PoliciesTable, dbtables.PoliciesColumnScope,
			dbtables.PoliciesTable, dbtables.PoliciesColumnDepartmentId)

		params = append(params, api.SCOPETYPE_DEPARTMENT)
		params = append(params, appliedOrgUnit.Id)

	case api.SCOPETYPE_CLUSTER:
		sql += fmt.Sprintf(" and %s.%s = ? and %s.%s = ?",
			dbtables.PoliciesTable, dbtables.PoliciesColumnScope,
			dbtables.PoliciesTable, dbtables.PoliciesColumnClusterId)

		params = append(params, api.SCOPETYPE_CLUSTER)
		params = append(params, appliedOrgUnit.Id)

	case api.SCOPETYPE_TENANT:
		sql += fmt.Sprintf(" and %s.%s = ? and %s.%s = ?",
			dbtables.PoliciesTable, dbtables.PoliciesColumnScope,
			dbtables.PoliciesTable, dbtables.PoliciesColumnTenantId)

		params = append(params, api.SCOPETYPE_TENANT)
		params = append(params, appliedOrgUnit.Id)
	}

	sql += ")"

	result = result.Where(sql, params...)

	result = result.Update(dbtables.DeletedAtColumn, time.Now())
	if result.Error != nil {
		dbErr := database.Error(database.DELETE, database.EffectivePolicy).WithDetailsFromDb(result)
		logging.Errorf("failed to soft delete effective policies by org unit: %s: %v", dbErr.Message, result.Error)
		return dbErr
	}

	return nil
}

// SoftDeleteEffectivePoliciesOfSubtree soft-deletes effective_policies rows
// of every descendant policy of appliedOrgUnit (for the given workloadType),
// and returns the descendant policy ids that participated in the delete.
//
// project scope has no subtree — the function returns (nil, nil) there.
//
// The returned ids can be handed directly to
// dao.DeleteBreakdownCachesOfSubtree so the cache invalidation step doesn't
// have to re-traverse the policies table.
func SoftDeleteEffectivePoliciesOfSubtree(ctx context.Context, tx *gorm.DB, appliedOrgUnit scopes.OrgUnit, workloadType string) ([]string, *api.Error) {
	// no subtree under project
	if appliedOrgUnit.Scope == api.SCOPETYPE_PROJECT {
		return nil, nil
	}

	// step 1: collect descendant policy ids in one pass on `policies`. we
	// keep the same shape as the previous sub-SELECT so behavior is
	// preserved: SYSTEM scope reaches every non-GVK policy (legacy fallback
	// cleanup), other scopes restrict to the relevant subtree.
	descendantIds, err := listSubtreePolicyIds(ctx, tx, appliedOrgUnit, workloadType)
	if err != nil {
		return nil, err
	}
	if len(descendantIds) == 0 {
		return nil, nil
	}

	// step 2: stamp deleted_at on the matching effective rows.
	result := tx.Table(dbtables.EffectivePoliciesTable).
		Where(fmt.Sprintf("%s IN ?", dbtables.EffectivePoliciesColumnPolicyId), descendantIds).
		Update(dbtables.DeletedAtColumn, time.Now())
	if result.Error != nil {
		dbErr := database.Error(database.DELETE, database.EffectivePolicy).WithDetailsFromDb(result)
		logging.Errorf("failed to soft delete effective policies of subtree: %s: %v", dbErr.Message, result.Error)
		return nil, dbErr
	}

	return descendantIds, nil
}

// listSubtreePolicyIds returns every policy id under the subtree rooted at
// appliedOrgUnit for the given workloadType. it mirrors the filter shape
// previously embedded as a sub-SELECT in SoftDeleteEffectivePoliciesOfSubtree:
//
//   - project scope has no subtree (handled by the caller).
//   - department: scope=project, department_id=appliedOrgUnit.Id, tenant_id=ctx tenant.
//   - cluster: scope in (project, department), cluster_id=appliedOrgUnit.Id, tenant_id=ctx tenant.
//   - tenant: scope in (project, department, cluster), tenant_id=ctx tenant.
//   - system: every non-GVK policy across all tenants (legacy fallback
//     cleanup; gvk_workload IS NOT TRUE === gvk_workload = false).
//
// Every non-system branch carries a tenant_id filter taken from the caller's
// context. Without it, child ids that collide across tenants (e.g.,
// department "111" exists under every tenant in the mock layout) would
// match unrelated tenants' rows. Even unique-looking ids like cluster UUIDs
// get the defensive guard so a leaked ancestor id from one tenant can never
// reach into another tenant's subtree. The TENANT branch no longer relies
// on appliedOrgUnit.Id alone for tenant scoping: it shares the same
// tenant_id = ctx tenant filter as the other branches, which is the only
// tenant a caller is ever allowed to operate on.
//
// soft-deleted rows are included via Unscoped() so placeholder policies that
// were just created still appear.
func listSubtreePolicyIds(ctx context.Context, tx *gorm.DB, appliedOrgUnit scopes.OrgUnit, workloadType string) ([]string, *api.Error) {
	q := tx.Table(dbtables.PoliciesTable).Unscoped()

	if appliedOrgUnit.Scope == api.SCOPETYPE_SYSTEM {
		// legacy fallback cleanup is explicitly cross-tenant: every non-GVK
		// policy in every tenant gets wiped together.
		q = q.Where(fmt.Sprintf("%s IS NOT TRUE", dbtables.PoliciesColumnGvkWorkload))
	} else {
		// pin the lookup to the caller's tenant before applying the
		// per-scope filters; this is what stops descendant ids from
		// matching rows in other tenants.
		q = q.Where(fmt.Sprintf("%s = ?", dbtables.PoliciesColumnTenantId), auth.TenantIdOf(ctx))

		if workloadType != FallbackPolicyAsType {
			q = q.Where(fmt.Sprintf("%s = ?", dbtables.PoliciesColumnWorkloadType), workloadType)
		}

		switch appliedOrgUnit.Scope {
		case api.SCOPETYPE_DEPARTMENT:
			q = q.
				Where(fmt.Sprintf("%s = ?", dbtables.PoliciesColumnScope), api.SCOPETYPE_PROJECT).
				Where(fmt.Sprintf("%s = ?", dbtables.PoliciesColumnDepartmentId), appliedOrgUnit.Id)
		case api.SCOPETYPE_CLUSTER:
			q = q.
				Where(fmt.Sprintf("%s IN ?", dbtables.PoliciesColumnScope),
					[]api.ScopeType{api.SCOPETYPE_PROJECT, api.SCOPETYPE_DEPARTMENT}).
				Where(fmt.Sprintf("%s = ?", dbtables.PoliciesColumnClusterId), appliedOrgUnit.Id)
		case api.SCOPETYPE_TENANT:
			// tenant_id is already pinned above; the per-row scope filter
			// is all that's needed to restrict to the subtree.
			q = q.Where(fmt.Sprintf("%s IN ?", dbtables.PoliciesColumnScope),
				[]api.ScopeType{api.SCOPETYPE_PROJECT, api.SCOPETYPE_DEPARTMENT, api.SCOPETYPE_CLUSTER})
		default:
			return nil, nil
		}
	}

	var ids []string
	res := q.Pluck(dbtables.PoliciesColumnId, &ids)
	if res.Error != nil {
		dbErr := database.Error(database.LIST, database.Policies).WithDetailsFromDb(res)
		logging.Errorf("failed to list subtree policy ids for %s/%s: %s: %v",
			appliedOrgUnit.Scope, appliedOrgUnit.Id, dbErr.Message, res.Error)
		return nil, dbErr
	}

	return ids, nil
}

// ChoiceInfo helps us to choose between type-specific and fallback for the same scope
// see below for further details
type ChoiceInfo struct {
	RecordIndex  int
	WorkloadType string
}

func FindPoliciesForSync(ctx context.Context, workloadType string, clusterId string, afterUpdate int64) ([]dbmodel.PolicyAndEffective, *api.Error) {
	tx := dao_utils.DB().WithContext(ctx)

	tenantId := auth.TenantIdOf(ctx)

	effectivePolicy := dbmodel.EffectivePolicy{}

	// prepare room for 2 x the bulk size, because we select also fallback candidates
	records := make([]dbmodel.PolicyAndEffective, 0, runai_config.Get[*config.ServerConfig]().PoliciesPerBulk*2)

	// join effective and policies
	query := tx.Table(dbtables.EffectivePoliciesTable).
		Select("*").
		Joins(effectivePolicy.JoinWithPolicy())

	// filter by tenant id
	query.Where(fmt.Sprintf("%s.%s = ?", dbtables.EffectivePoliciesTable, dbtables.PoliciesColumnTenantId), tenantId)

	// filter out records without given rules/defaults, they were only created as placeholders
	// as part of the "get"
	query.Where(fmt.Sprintf("(%s.%s is not null or %s.%s is not null or %s.%s is not null)",
		dbtables.PoliciesTable, dbtables.PoliciesColumnDefaults,
		dbtables.PoliciesTable, dbtables.PoliciesColumnRules,
		dbtables.PoliciesTable, dbtables.PoliciesColumnImposedAssets))

	// filter by cluster ID (and include also tenant-policies which are populated to all the clusters)
	query.Where(fmt.Sprintf("%s.%s IS NULL OR %s.%s = ?",
		dbtables.PoliciesTable, dbtables.PoliciesColumnClusterId,
		dbtables.PoliciesTable, dbtables.PoliciesColumnClusterId),
		clusterId)

	// filter by workload type
	query.Where(fmt.Sprintf("(%s = ? or %s = ?)", dbtables.PoliciesColumnWorkloadType, dbtables.PoliciesColumnWorkloadType), workloadType, FallbackPolicyAsType)

	// start from given update count
	query.Order(dbtables.EffectivePoliciesColumnUpdateCount)

	if afterUpdate != 0 {
		query = query.Where(fmt.Sprintf("%s > ?", dbtables.EffectivePoliciesColumnUpdateCount), afterUpdate)
	}

	query = query.Limit(runai_config.Get[*config.ServerConfig]().PoliciesPerBulk * 2)

	// execute the query unscoped (=include soft deleted records)
	result := query.Unscoped().Find(&records)
	if result.Error != nil {
		dbErr := database.Error(database.LIST, database.EffectivePolicy).WithDetailsFromDb(result)
		logging.Errorf("failed to list effective policies for sync: %s: %v", dbErr.Message, result.Error)
		return nil, dbErr
	}

	//
	//   if there is both fallback and type-specific record for the same scope, the type-specific takes over
	//   we need to sync the fallback policy only if no type-specific is synced
	//
	chosenRecordPerOrgUnit := make(map[string]ChoiceInfo)

	for idx, record := range records {
		orgUnit := record.OrgUnit().String()

		_, found := chosenRecordPerOrgUnit[orgUnit]
		if !found || record.WorkloadType != FallbackPolicyAsType {
			chosenRecordPerOrgUnit[orgUnit] = ChoiceInfo{
				WorkloadType: record.WorkloadType,
				RecordIndex:  idx,
			}
		}
	}

	chosenRecords := make([]dbmodel.PolicyAndEffective, 0, runai_config.Get[*config.ServerConfig]().PoliciesPerBulk)
	for idx := range records {
		if len(chosenRecords) == runai_config.Get[*config.ServerConfig]().PoliciesPerBulk {
			break
		}

		record := &records[idx]

		choiceInfo := chosenRecordPerOrgUnit[record.OrgUnit().String()]
		if choiceInfo.RecordIndex != idx {
			continue
		}

		// we no longer care about the content of the policy bcause we only send "delete" to the agent.
		record.WorkloadType = workloadType // make sure fallback is changed to the real workload type
		chosenRecords = append(chosenRecords, *record)
	}

	return chosenRecords, nil
}
