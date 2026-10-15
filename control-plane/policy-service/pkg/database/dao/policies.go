/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	authApi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/google/uuid"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/scopes"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	querydbenrich "github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi/dbenrich"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"gorm.io/gorm"
)

const FallbackPolicyAsType = "Fallback"

// AddPolicy creates a new policy record. The ID of the policy is returned in the policy object
// it also returns the ids of non-relevant policies of the same org unit which were deleted
// (in normal condition - one record at the most).
//
// Persisted columns include gvk_workload and validation_issues when set on policy (GORM Save).
func AddPolicy(ctx context.Context, tx orm.DB, policy *dbmodel.Policy) *api.Error {
	//
	//    now add the new policy
	//
	policy.DefaultsV2 = database.CleanMap(policy.DefaultsV2)
	policy.RulesV2 = database.CleanMap(policy.RulesV2)

	stored, projectionErr := policy.ForDB()
	if projectionErr != nil {
		return database.Error(database.CREATE, database.Policies).WithDetails(projectionErr)
	}
	if _, err := tx.Table(dbtables.PoliciesTable).Save(stored); err != nil {
		if policy.GvkWorkload && isUniqueViolation(err) {
			return api.NewError(http.StatusConflict, "policy for similar scope and workload type already exists", nil)
		}
		dbErr := database.Error(database.CREATE, database.Policies).WithDetails(err)
		logging.Errorf("failed to add policy: %s: %v", dbErr.Message, err)
		return dbErr
	}

	policy.BaseModel = stored.BaseModel
	return nil
}

func UpdatePolicySpec(_ context.Context, tx orm.DB, policy *dbmodel.Policy) *api.Error {
	policy.DefaultsV2 = database.CleanMap(policy.DefaultsV2)
	policy.RulesV2 = database.CleanMap(policy.RulesV2)

	stored, projectionErr := policy.ForDB()
	if projectionErr != nil {
		return database.Error(database.UPDATE, database.Policies).WithDetails(projectionErr)
	}

	updates := map[string]interface{}{
		dbtables.DeletedAtColumn:             nil,
		dbtables.CreatedAtColumn:             policy.CreatedAt,
		dbtables.UpdatedAtColumn:             policy.UpdatedAt,
		dbtables.CreatedByColumn:             policy.CreatedBy,
		dbtables.UpdatedByColumn:             policy.UpdatedBy,
		dbtables.PoliciesColumnName:          policy.Name,
		dbtables.PoliciesColumnDefaults:      stored.DefaultsV2,
		dbtables.PoliciesColumnRules:         stored.RulesV2,
		dbtables.PoliciesColumnImposedAssets: policy.ImposedAssets,
		dbtables.PoliciesColumnGvkWorkload:   policy.GvkWorkload,
	}

	// Handle validation status and issues as separate columns
	if policy.ValidationStatus != nil {
		updates[dbtables.PoliciesColumnValidationStatus] = string(*policy.ValidationStatus)
	} else {
		updates[dbtables.PoliciesColumnValidationStatus] = nil
	}

	if len(policy.ValidationIssues) > 0 {
		issuesJSON, err := json.Marshal(policy.ValidationIssues)
		if err != nil {
			return api.NewError(http.StatusInternalServerError, "failed to marshal validation issues", err)
		}
		updates[dbtables.PoliciesColumnValidationIssues] = string(issuesJSON)
	} else {
		updates[dbtables.PoliciesColumnValidationIssues] = nil
	}

	if _, err := tx.Table(dbtables.PoliciesTable).
		Where(fmt.Sprintf("%s.%s = ?", dbtables.PoliciesTable, dbtables.PoliciesColumnId), policy.ID).
		UpdateColumns(updates); err != nil {
		dbErr := database.Error(database.UPDATE, database.Policies).WithDetails(err)
		logging.Errorf("failed to update policy spec: %s: %v", dbErr.Message, err)
		return dbErr
	}

	return nil
}

// UpdatePoliciesStatus updates the validation_status and validation_issues columns for multiple policies by their IDs.
// This is used when mappings or other components change and we need to mark dependent policies
// as PendingUpdate.
func UpdatePoliciesStatus(ctx context.Context, tx orm.DB, policyIds []string, status *api.PolicyStatus) *api.Error {
	if len(policyIds) == 0 {
		return nil
	}

	tenantId := auth.TenantIdOf(ctx)

	var updates map[string]interface{}
	if status != nil {
		updates = map[string]interface{}{
			dbtables.PoliciesColumnValidationStatus: string(status.Phase),
		}
		if len(status.Issues) > 0 {
			issuesJSON, err := json.Marshal(status.Issues)
			if err != nil {
				return api.NewError(http.StatusInternalServerError, "failed to marshal policy issues", err)
			}
			updates[dbtables.PoliciesColumnValidationIssues] = string(issuesJSON)
		} else {
			updates[dbtables.PoliciesColumnValidationIssues] = nil
		}
	} else {
		updates = map[string]interface{}{
			dbtables.PoliciesColumnValidationStatus: nil,
			dbtables.PoliciesColumnValidationIssues: nil,
		}
	}

	if _, err := tx.Table(dbtables.PoliciesTable).
		Where(fmt.Sprintf("%s.%s = ?", dbtables.PoliciesTable, dbtables.PoliciesColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s.%s IN ?", dbtables.PoliciesTable, dbtables.PoliciesColumnId), policyIds).
		UpdateColumns(updates); err != nil {
		dbErr := database.Error(database.UPDATE, database.Policies).WithDetails(err)
		logging.Errorf("failed to update policies status: %s: %v", dbErr.Message, err)
		return dbErr
	}

	return nil
}

// PatchPolicyScopes is used for migrating policy records that had partial scope to
// having complete scope ids. It is part of our transition from tree based to non-tree based.
func PatchPolicyScopes(_ context.Context, tx orm.DB, policy *dbmodel.Policy) *api.Error {
	if _, err := tx.Table(dbtables.PoliciesTable).
		Where(fmt.Sprintf("%s.%s = ?", dbtables.PoliciesTable, dbtables.PoliciesColumnId), policy.ID).
		UpdateColumns(map[string]interface{}{
			dbtables.PoliciesColumnClusterId:    policy.ClusterId,
			dbtables.PoliciesColumnDepartmentId: policy.DepartmentId,
			dbtables.PoliciesColumnProjectId:    policy.ProjectId,
		}); err != nil {
		dbErr := database.Error(database.UPDATE, database.Policies).WithDetails(err)
		logging.Errorf("failed to patch policy scopes: %s: %v", dbErr.Message, err)
		return dbErr
	}

	return nil
}

// DeleteObsoletePoliciesByScope deletes policies that are under the scope id
// This function is invoked AFTER the service function to delete policies by scope finishes.
// This function removed everything from memory + mark everything as soft delete in the DB.
// This function, DeleteObsoletePoliciesByScope, is then invoked to perform hard-delete on
// all the policies that were soft deleted.
func DeleteObsoletePoliciesByScope(ctx context.Context, db orm.DB, scopeType api.ScopeType, scopeId string) *api.Error {
	_ = db.Transaction(func(tx orm.DB) error {
		tenantId := auth.TenantIdOf(ctx)
		var idsToDelete []string
		scopeIdColumn, err := dbtables.ScopeIdColumnName(scopeType)
		if err != nil {
			return err
		}

		if err := tx.Table(dbtables.PoliciesTable).
			Where(fmt.Sprintf("%s = ?", dbtables.PoliciesColumnTenantId), tenantId).
			Where(fmt.Sprintf("%s = ?", scopeIdColumn), scopeId).
			Pluck(dbtables.PoliciesColumnId, &idsToDelete); err != nil {
			apiErr := database.Error(database.LIST, database.Policies).WithDetails(err)
			logging.Errorf("failed to list policies to delete: %s: %v", apiErr.Message, err)
			return apiErr
		}

		// nothing to delete
		if len(idsToDelete) == 0 {
			return nil
		}

		if _, err := tx.Table(dbtables.PoliciesTable).
			Where(fmt.Sprintf("%s IN (?)", dbtables.PoliciesColumnId), idsToDelete).
			Delete(dbtables.DeleteThisRecord); err != nil {
			apiErr := database.Error(database.DELETE, database.Policies).WithDetails(err)
			logging.Errorf("failed to delete policies: %s: %v", apiErr.Message, err)
			return apiErr
		}

		if _, err := tx.Table(dbtables.EffectivePoliciesTable).
			Where(fmt.Sprintf("%s IN (?)", dbtables.EffectivePoliciesColumnPolicyId), idsToDelete).
			Delete(dbtables.DeleteThisRecord); err != nil {
			apiErr := database.Error(database.DELETE, database.EffectivePolicy).WithDetails(err)
			logging.Errorf("failed to delete effective policies: %s: %v", apiErr.Message, err)
			return apiErr
		}

		return nil
	})

	return nil
}

// SoftDeletePolicyById sets deleted_at on the policy row (legacy / non-GVK-workload policies).
// For gvk_workload policies, use HardDeletePolicyById instead.
func SoftDeletePolicyById(ctx context.Context, tx orm.DB, policy *dbmodel.Policy) *api.Error {
	searchQuery := dbmodel.Policy{
		BaseModel: dbmodels.BaseModel{
			ID: policy.ID,
		},
		TenantId: auth.TenantIdOf(ctx),
	}

	if _, err := tx.
		Model(policy).
		Where(searchQuery).
		UpdateColumns(map[string]interface{}{
			dbtables.DeletedAtColumn: time.Now(),
		}); err != nil {
		dbErr := database.Error(database.DELETE, database.Policies).WithDetails(err)
		logging.Errorf("failed to soft delete policy: %s: %v", dbErr.Message, err)
		return dbErr
	}

	// we don't trust gorm to fill in the deletedAt time, so we fill it ourselves. the exact time
	// is not important, we just want the field to have valid value, so when asked if the policy
	// is deleted the result will be correct.
	policy.DeletedAt.Time = time.Now()
	policy.DeletedAt.Valid = true

	return nil
}

// HardDeletePolicyById permanently removes the policy row and its effective_policies rows.
// Use this for GVK-workload policies (gvk_workload = true), which do not use soft delete.
func HardDeletePolicyById(ctx context.Context, tx orm.DB, policy *dbmodel.Policy) *api.Error {
	searchQuery := dbmodel.Policy{
		BaseModel: dbmodels.BaseModel{
			ID: policy.ID,
		},
		TenantId: auth.TenantIdOf(ctx),
	}

	rowsAffected, err := tx.Unscoped().Table(dbtables.EffectivePoliciesTable).
		Where(fmt.Sprintf("%s = ?", dbtables.EffectivePoliciesColumnPolicyId), policy.ID).
		Where(fmt.Sprintf("%s = ?", dbtables.EffectivePoliciesColumnTenantId), auth.TenantIdOf(ctx)).
		Delete(dbtables.DeleteThisRecord)
	if err != nil {
		dbErr := database.Error(database.DELETE, database.EffectivePolicy).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("failed to delete effective policies: %s: %v", dbErr.Message, err)
		return dbErr
	}

	rowsAffected, err = tx.Unscoped().Model(searchQuery).Where(&searchQuery).Delete(dbtables.DeleteThisRecord)
	if err != nil {
		dbErr := database.Error(database.DELETE, database.Policies).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("failed to delete policy: %s: %v", dbErr.Message, err)
		return dbErr
	}

	return nil
}

// FindPolicyOptions controls which rows FindPolicyByID will surface. The given
// policy and its effective policy are filtered independently: the given side
// follows the policies row's soft-delete state, while the effective side
// follows the effective_policies row's soft-delete state. For most callers
// the two flags will be identical; GVK operations are an exception and ask
// for the effective row even when it has been soft-deleted (so that recompute
// flows can still observe and update it in place).
type FindPolicyOptions struct {
	IncludeDeletedGivenPolicies     bool
	IncludeDeletedEffectivePolicies bool
}

// FindPolicyByID loads a policy by id for the current tenant, including gvk_workload and validation_issues.
// When IncludeDeletedGivenPolicies is false, soft-deleted policies rows are
// excluded; when true, Unscoped() is used. The effective row is filtered
// independently via IncludeDeletedEffectivePolicies (applied as part of the
// LEFT JOIN condition so a missing effective row never hides the given one).
func FindPolicyByID(ctx context.Context, tx orm.DB, policyID string, opts FindPolicyOptions) (*dbmodel.PolicyWithOptionalEffective, *api.Error) {
	id, parseErr := uuid.Parse(policyID)
	if parseErr != nil {
		return nil, api.NewValidationError("invalid policy id")
	}

	searchQ := dbmodel.Policy{
		BaseModel: dbmodels.BaseModel{
			ID: id,
		},
		TenantId: auth.TenantIdOf(ctx),
	}

	var record dbmodel.PolicyWithOptionalEffective

	result := tx.Table(dbtables.PoliciesTable)

	// workaround for postgres deletedAt issue
	result = selectForOptionalEffective(result)

	result = result.Where(&searchQ)

	if !opts.IncludeDeletedGivenPolicies {
		result = result.Where(fmt.Sprintf("%s.%s IS NULL", dbtables.PoliciesTable, dbtables.DeletedAtColumn))
	} else {
		result = result.Unscoped()
	}

	rowsAffected, err := result.
		Joins(searchQ.LeftJoinEffectivePolicy(opts.IncludeDeletedEffectivePolicies)).
		Scan(&record)
	if err != nil {
		dbErr := database.Error(database.LIST, database.Policies).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("failed to find policy: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}
	if rowsAffected == 0 {
		return nil, nil
	}

	record.Policy.DeletedAt = gorm.DeletedAt(record.PolicyDeleted)
	if record.EffectiveDeleted.Valid {
		record.DeletedAt = convutils.Ref(record.EffectiveDeleted.Time)
	}

	record.Policy.DefaultsV2 = database.CleanMap(record.Policy.DefaultsV2)
	record.Policy.RulesV2 = database.CleanMap(record.Policy.RulesV2)

	record.EffectivePolicy.DefaultsV2 = database.CleanMap(record.EffectivePolicy.DefaultsV2)
	record.EffectivePolicy.RulesV2 = database.CleanMap(record.EffectivePolicy.RulesV2)

	return &record, nil
}

type ListByOrgUnitOptions struct {
	IncludeDeleted  bool
	IncludeFallback bool
	IncludeSystem   bool
}

func ListPoliciesByOrgUnits(ctx context.Context, tx orm.DB, tenantId int32, workloadType string,
	relevantOrgUnits []scopes.OrgUnit, options ListByOrgUnitOptions,
) (map[string]*dbmodel.PolicyWithOptionalEffective, *api.Error) {
	records := make([]*dbmodel.PolicyWithOptionalEffective, 0)

	result := tx.Table(dbtables.PoliciesTable)

	// workaround for postgres deletedAt issue
	result = selectForOptionalEffective(result)

	//
	//   add to the query where statement that covers all the hierarchy of the policy
	//
	params := make([]interface{}, 0)
	scopesQuery := "false"

	for _, orgUnit := range relevantOrgUnits {
		columnToFilter := ""

		switch orgUnit.Scope {
		case api.SCOPETYPE_PROJECT:
			columnToFilter = dbtables.PoliciesColumnProjectId
		case api.SCOPETYPE_DEPARTMENT:
			columnToFilter = dbtables.PoliciesColumnDepartmentId
		case api.SCOPETYPE_CLUSTER:
			columnToFilter = dbtables.PoliciesColumnClusterId
		case api.SCOPETYPE_TENANT:
			columnToFilter = dbtables.PoliciesColumnTenantId
		default:
			continue
		}

		scopesQuery += fmt.Sprintf(" or (%s.%s = ? and %s.%s = ?)",
			dbtables.PoliciesTable, dbtables.PoliciesColumnScope, dbtables.PoliciesTable, columnToFilter)

		params = append(params, orgUnit.Scope)
		params = append(params, orgUnit.Id)
	}

	result = result.Where(scopesQuery, params...)

	if options.IncludeFallback {
		result = result.Where(fmt.Sprintf("(%s.%s = ? or %s.%s = ?)", dbtables.PoliciesTable, dbtables.PoliciesColumnWorkloadType, dbtables.PoliciesTable, dbtables.PoliciesColumnWorkloadType),
			workloadType, FallbackPolicyAsType)
	} else {
		result = result.Where(fmt.Sprintf("(%s.%s = ?)", dbtables.PoliciesTable, dbtables.PoliciesColumnWorkloadType), workloadType)
	}

	if !options.IncludeDeleted {
		result = result.Where(fmt.Sprintf("%s.%s IS NULL", dbtables.PoliciesTable, dbtables.DeletedAtColumn))
	} else {
		result = result.Unscoped()
	}

	var policy dbmodel.Policy
	if _, err := result.
		Joins(policy.LeftJoinEffectivePolicy(options.IncludeDeleted)).
		Scan(&records); err != nil {
		dbErr := database.Error(database.LIST, database.Policies).WithDetails(err)
		logging.Errorf("failed to list policies: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}

	//
	//   add the system policy
	//
	if options.IncludeSystem {
		records = append(records, &systemPolicyDB)
	}

	return listToMap(records), nil
}

func listToMap(records []*dbmodel.PolicyWithOptionalEffective) map[string]*dbmodel.PolicyWithOptionalEffective {
	var recordsAsMap = make(map[string]*dbmodel.PolicyWithOptionalEffective)

	// arrange the records in a map that is more comfortable to work with
	for _, record := range records {
		record.Policy.DeletedAt = gorm.DeletedAt(record.PolicyDeleted)
		if record.EffectiveDeleted.Valid {
			record.DeletedAt = convutils.Ref(record.EffectiveDeleted.Time)
		}

		record.Policy.DefaultsV2 = database.CleanMap(record.Policy.DefaultsV2)
		record.Policy.RulesV2 = database.CleanMap(record.Policy.RulesV2)

		record.EffectivePolicy.DefaultsV2 = database.CleanMap(record.EffectivePolicy.DefaultsV2)
		record.EffectivePolicy.RulesV2 = database.CleanMap(record.EffectivePolicy.RulesV2)

		recordsAsMap[record.ID.String()] = record
	}

	return recordsAsMap
}

func CountGivenPoliciesOfTenant(ctx context.Context) (int64, *api.Error) {
	tx := runaidb.DB().WithContext(ctx)

	whereClause := dbmodel.Policy{
		TenantId: auth.TenantIdOf(ctx),
	}

	count := int64(0)

	if err := tx.Model(&whereClause).Where(whereClause).Unscoped().Count(&count); err != nil {
		dbErr := database.Error(database.COUNT, database.Policies).WithDetails(err)
		logging.Errorf("failed to count policies: %s: %v", dbErr.Message, err)
		return 0, dbErr
	}

	return count, nil
}

type ListOptions struct {
	IncludeDeleted          bool
	IncludeFallbackPolicies bool
	IncludeGvkPolicies      bool
	IncludeNativePolicies   bool
}

// ListGivenPoliciesOfTenant returns all the given policies of a tenant. as opposed to ListPoliciesOfTenant,
// it returns policies of all types, but it does not return the defaults/rules and nor the effective policy,
// only general info about the given policy
func ListGivenPoliciesOfTenant(ctx context.Context, workloadType string, scope string, departmentId string,
	projectId int32, clusterId string, permittedScopes *authApi.PermittedScopes, options ListOptions,
) ([]dbmodel.PolicyWithPartialEffective, *api.Error) {
	tx := runaidb.DB().WithContext(ctx)

	var policy dbmodel.Policy

	records := make([]dbmodel.PolicyWithPartialEffective, 0)

	whereClause := dbmodel.Policy{
		TenantId: auth.TenantIdOf(ctx),
	}

	whereClause.ApplyScopeQueryParams(workloadType, scope, departmentId, projectId, clusterId)

	result := tx.Table(dbtables.PoliciesTable).
		Select(fmt.Sprintf("%s.*, %s.%s, %s.%s", dbtables.PoliciesTable,
			dbtables.EffectivePoliciesTable, dbtables.EffectivePoliciesColumnUpdateCount,
			dbtables.EffectivePoliciesTable, dbtables.EffectivePoliciesColumnValidationError)).
		Omit(dbtables.PoliciesColumnRules, dbtables.PoliciesColumnDefaults).
		Where(whereClause).
		Where(fmt.Sprintf("(%s.%s is not null or %s.%s is not null or %s.%s is not null)",
			dbtables.PoliciesTable, dbtables.PoliciesColumnDefaults,
			dbtables.PoliciesTable, dbtables.PoliciesColumnRules,
			dbtables.PoliciesTable, dbtables.PoliciesColumnImposedAssets,
		))

	if !options.IncludeDeleted {
		result = result.Where(fmt.Sprintf("%s.%s IS NULL", dbtables.PoliciesTable, dbtables.DeletedAtColumn))
	} else {
		result = result.Unscoped()
	}

	if !options.IncludeFallbackPolicies {
		result = result.Where(fmt.Sprintf("%s.%s <> ?", dbtables.PoliciesTable, dbtables.PoliciesColumnWorkloadType), FallbackPolicyAsType)
	}

	if !options.IncludeNativePolicies && !options.IncludeGvkPolicies {
		return nil, api.NewValidationError("either native or gvk policies must be included")
	}

	if options.IncludeGvkPolicies != options.IncludeNativePolicies {
		result = result.Where(fmt.Sprintf("(%s.%s = ?)", dbtables.PoliciesTable, dbtables.PoliciesColumnGvkWorkload), options.IncludeGvkPolicies)
	}

	result = dbmodel.ApplyPermittedScopes(result, permittedScopes)

	if err := result.
		Joins(policy.LeftJoinEffectivePolicy(options.IncludeDeleted)).
		Find(&records); err != nil {
		dbErr := database.Error(database.LIST, database.Policies).WithDetails(err)
		logging.Errorf("failed to list policies: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}

	return records, nil
}

// ListGvkPoliciesOfTenant returns all non-deleted policies with gvk_workload=true for the current tenant,
func ListGvkPoliciesOfTenant(ctx context.Context, permittedScopes *authApi.PermittedScopes, filters []queryapi.Filter,
	sortColumn string, sortOrder string, offset int32, limit int32,
) ([]dbmodel.PolicyWithOptionalEffective, *api.Error) {
	tx := runaidb.DB().WithContext(ctx)

	records := make([]dbmodel.PolicyWithOptionalEffective, 0)

	// Use Policy only for Where — PolicyWithOptionalEffective also carries PolicyDeleted /
	// EffectiveDeleted (SELECT aliases), and GORM would wrongly emit WHERE policies.policy_deleted ...
	policyWhere := dbmodel.Policy{
		TenantId:    auth.TenantIdOf(ctx),
		GvkWorkload: true,
	}

	res := tx.Table(dbtables.PoliciesTable)

	// workaround for postgres deletedAt issue
	res = selectForOptionalEffective(res)

	res = res.Where(&policyWhere).
		Where(fmt.Sprintf("%s.%s IS NULL", dbtables.PoliciesTable, dbtables.DeletedAtColumn)).
		Joins(policyWhere.LeftJoinEffectivePolicy(dbmodel.NotToIncludeDeletedPolicies))

	res = dbmodel.ApplyPermittedScopes(res, permittedScopes)

	res = res.WithEnricher(querydbenrich.Filters(filters))

	if sortColumn != "" {
		res = res.Order(querydbenrich.FormatSortClause(sortColumn, sortOrder))
	}

	res = res.Order("id DESC")

	if offset != 0 {
		res = res.Offset(int(offset))
	}

	if limit != 0 {
		res = res.Limit(int(limit))
	}

	if err := res.Find(&records); err != nil {
		dbErr := database.Error(database.LIST, database.Policies).WithDetails(err)
		logging.Errorf("error listing policies: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}

	for i := range records {
		records[i].Policy.DeletedAt = gorm.DeletedAt(records[i].PolicyDeleted)
		if records[i].EffectiveDeleted.Valid {
			records[i].DeletedAt = convutils.Ref(records[i].EffectiveDeleted.Time)
		}

		sanitizeRecord(&records[i].Policy)

		if records[i].EffectivePolicy.RulesV2 != nil {
			records[i].EffectivePolicy.RulesV2 = database.CleanMap(records[i].EffectivePolicy.RulesV2)
		}

		if records[i].EffectivePolicy.DefaultsV2 != nil {
			records[i].EffectivePolicy.DefaultsV2 = database.CleanMap(records[i].EffectivePolicy.DefaultsV2)
		}
	}

	return records, nil
}

// ListPoliciesOfTenant returns all policies of a given tenant. it also returns indication for policies
// that have effective ID. This function is "tailor made" for the function which loads from the DB to the
// tenant hierarchy tree,
// note that includeDeleted applies only to policies. deleted effective policy we do not need for loading
// into the tree, unlike deleted policies which we create for scopes that have effective but no given policy.
func ListPoliciesOfTenant(ctx context.Context, tenantId int32, workloadType string, includeDeleted bool) ([]dbmodel.PolicyWithOptionalEffectiveId, *api.Error) {
	tx := runaidb.DB().WithContext(ctx)

	records := make([]dbmodel.PolicyWithOptionalEffectiveId, 0)

	whereClause := dbmodel.PolicyWithOptionalEffectiveId{
		Policy: dbmodel.Policy{
			TenantId:     tenantId,
			WorkloadType: workloadType,
		},
	}

	result := tx.Table(dbtables.PoliciesTable).
		Select(fmt.Sprintf("%s.*, %s.%s",
			dbtables.PoliciesTable,
			dbtables.EffectivePoliciesTable,
			dbtables.EffectivePoliciesColumnPolicyId)).
		Where(whereClause)

	if !includeDeleted {
		result = result.Where(fmt.Sprintf("%s.%s IS NULL", dbtables.PoliciesTable, dbtables.DeletedAtColumn))
	} else {
		result = result.Unscoped()
	}

	if err := result.
		Joins(whereClause.LeftJoinEffectivePolicy(dbmodel.NotToIncludeDeletedPolicies)).
		Find(&records); err != nil {
		dbErr := database.Error(database.LIST, database.Policies).WithDetails(err)
		logging.Errorf("failed to list policies: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}

	for _, record := range records {
		record.DefaultsV2 = database.CleanMap(record.DefaultsV2)
		record.RulesV2 = database.CleanMap(record.RulesV2)
	}

	return records, nil
}

func FindGivenPolicyOfOrgUnit(ctx context.Context, tx orm.DB, scope api.ScopeType, clusterId *string, departmentId *string, projectId *int32,
	workloadType string, includeDeleted bool,
) (*dbmodel.Policy, *api.Error) {
	policy := dbmodel.Policy{
		TenantId:     auth.TenantIdOf(ctx),
		Scope:        scope,
		ClusterId:    clusterId,
		DepartmentId: departmentId,
		ProjectId:    projectId,
		WorkloadType: workloadType,
	}

	var record dbmodel.Policy

	result := tx.Table(dbtables.PoliciesTable).
		Where(&policy)

	if !includeDeleted {
		result = result.Where(fmt.Sprintf("%s.%s IS NULL", dbtables.PoliciesTable, dbtables.DeletedAtColumn))
	} else {
		result = result.Unscoped()
	}

	rowsAffected, err := result.Scan(&record)
	if err != nil {
		dbErr := database.Error(database.LIST, database.Policies).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("error FindGivenPolicyOfOrgUnit: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}
	if rowsAffected == 0 {
		return nil, nil
	}

	sanitizeRecord(&record)
	return &record, nil
}

func FindGivenPoliciesOfOrgUnitAllTypes(ctx context.Context, tx orm.DB, scope api.ScopeType, clusterId *string,
	departmentId *string, projectId *int32, includeDeleted bool,
) ([]dbmodel.Policy, *api.Error) {
	policy := dbmodel.Policy{
		TenantId:     auth.TenantIdOf(ctx),
		Scope:        scope,
		ClusterId:    clusterId,
		DepartmentId: departmentId,
		ProjectId:    projectId,
	}

	records := make([]dbmodel.Policy, 0)

	result := tx.Table(dbtables.PoliciesTable).Where(&policy)

	if !includeDeleted {
		result = result.Where(fmt.Sprintf("%s.%s IS NULL", dbtables.PoliciesTable, dbtables.DeletedAtColumn))
	} else {
		result = result.Unscoped()
	}

	if _, err := result.Scan(&records); err != nil {
		dbErr := database.Error(database.LIST, database.Policies).WithDetails(err)
		logging.Errorf("error FindGivenPoliciesOfOrgUnitAllTypes: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}

	for _, record := range records {
		sanitizeRecord(&record)
	}

	return records, nil
}

func sanitizeRecord(record *dbmodel.Policy) {
	if record.IsDeleted() {
		record.DefaultsV2 = nil
		record.RulesV2 = nil
		record.ImposedAssets = nil
	} else {
		record.DefaultsV2 = database.CleanMap(record.DefaultsV2)
		record.RulesV2 = database.CleanMap(record.RulesV2)
		if record.ImposedAssets != nil {
			// there is no sense keeping the imposed assets section if both worker and master appear as null
			// in this case remove the entire section from the JSON it makes the API response more compact
			if record.ImposedAssets.Worker == nil && record.ImposedAssets.Master == nil {
				record.ImposedAssets = nil
			}
		}
	}
}

func FindPolicyOfOrgUnit(ctx context.Context, tx orm.DB, scope api.ScopeType, departmentId *string, projectId *int32,
	clusterId *string, workloadType string, includeDeleted bool,
) (*dbmodel.PolicyWithOptionalEffective, *api.Error) {
	policy := dbmodel.Policy{
		TenantId:     auth.TenantIdOf(ctx),
		Scope:        scope,
		DepartmentId: departmentId,
		ProjectId:    projectId,
		ClusterId:    clusterId,
		WorkloadType: workloadType,
	}

	var record dbmodel.PolicyWithOptionalEffective

	result := tx.Table(dbtables.PoliciesTable)

	// workaround for postgres deletedAt issue
	result = selectForOptionalEffective(result)

	result = result.Where(&policy)

	if !includeDeleted {
		result = result.Where(fmt.Sprintf("%s.%s IS NULL", dbtables.PoliciesTable, dbtables.DeletedAtColumn))
	} else {
		result = result.Unscoped()
	}

	rowsAffected, err := result.
		Joins(policy.LeftJoinEffectivePolicy(includeDeleted)).
		Scan(&record)
	if err != nil {
		dbErr := database.Error(database.LIST, database.Policies).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("error FindPolicyOfOrgUnit: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}
	if rowsAffected == 0 {
		return nil, nil
	}

	record.Policy.DeletedAt = gorm.DeletedAt(record.PolicyDeleted)
	if record.EffectiveDeleted.Valid {
		record.DeletedAt = convutils.Ref(record.EffectiveDeleted.Time)
	}

	record.Policy.DefaultsV2 = database.CleanMap(record.Policy.DefaultsV2)
	record.Policy.RulesV2 = database.CleanMap(record.Policy.RulesV2)

	record.EffectivePolicy.DefaultsV2 = database.CleanMap(record.EffectivePolicy.DefaultsV2)
	record.EffectivePolicy.RulesV2 = database.CleanMap(record.EffectivePolicy.RulesV2)

	return &record, nil
}

func FindAllIdsOfOrgUnit(_ context.Context, tx orm.DB, policy *dbmodel.Policy, includeDeleted bool) ([]string, *api.Error) {
	searchKey := dbmodel.Policy{
		TenantId:     policy.TenantId,
		Scope:        policy.Scope,
		DepartmentId: policy.DepartmentId,
		ProjectId:    policy.ProjectId,
		WorkloadType: policy.WorkloadType,
		ClusterId:    policy.ClusterId,
	}

	var ids []string

	result := tx.Table(dbtables.PoliciesTable).
		Select(fmt.Sprintf("%s.%s", dbtables.PoliciesTable, dbtables.PoliciesColumnId))

	if !includeDeleted {
		result = result.Where(fmt.Sprintf("%s.%s IS NULL", dbtables.PoliciesTable, dbtables.DeletedAtColumn))
	} else {
		result = result.Unscoped()
	}

	if _, err := result.Where(&searchKey).Scan(&ids); err != nil {
		dbErr := database.Error(database.LIST, database.Policies).WithDetails(err)
		logging.Errorf("error FindAllIdsOfOrgUnit: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}

	return ids, nil
}

// workaround for postgres issue not filling policies.DeletedAt.
// we request the two deletion flags specifically and patch the result with their value
func selectForOptionalEffective(tx orm.DB) orm.DB {
	return tx.Select(fmt.Sprintf("%s.*, %s.*, %s.%s as policy_deleted, %s.%s as effective_deleted",
		dbtables.PoliciesTable,
		dbtables.EffectivePoliciesTable,
		dbtables.PoliciesTable, dbtables.DeletedAtColumn,
		dbtables.EffectivePoliciesTable, dbtables.DeletedAtColumn))
}
