/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/run-ai/runai/runai-common-packages/dao-utils/validators"

	telemetry_utils "github.com/run-ai/runai/runai-application-packages/telemetry"

	"github.com/run-ai/runai/control-plane/workloads/pkg/services_utils"

	"github.com/google/uuid"
	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	WorkloadsTableName                 = "workloads"
	WorkloadsViewName                  = "workloads_view"
	WorkloadsMaterializedViewName      = "workloads_materialized_view"
	WorkloadsApiViewName               = "workloads_materialized_wrapper_view"
	WorkloadsChildrenTableName         = "workloads_children"
	WorkloadsColumnId                  = "id"
	WorkloadsDeletedAtColumnName       = "deleted_at"
	RunaiJobType                       = "RunaiJob"
	RevisionType                       = "Revision"
	WorkloadsColumnTenantId            = "tenant_id"
	WorkloadsColumnClusterId           = "cluster_id"
	WorkloadsColumnProjectId           = "project_id"
	WorkloadsColumnDepartmentId        = "department_id"
	WorkloadsColumnDatasources         = "datasources"
	WorkloadsColumnEnvironments        = "environments"
	WorkloadsColumnExternalConnections = "external_connections"
	WorkloadsColumnAdditionalFields    = "additional_fields"
	WorkloadsColumnSourceApi           = "source_api"
	WorkloadsColumnPreemptibility      = "preemptibility"
	WorkloadsColumnSupportDelete       = "support_delete"
	WorkloadsColumnSupportSuspend      = "support_suspend"
)

type WorkloadsDao struct {
	view string
}

func NewWorkloadsDao(conf *config.ServerConfig) *WorkloadsDao {
	view := WorkloadsViewName
	if conf.ViewRefreshRateMilliseconds > 0 {
		if conf.UsePartialMaterializedView {
			view = WorkloadsApiViewName
		} else {
			view = WorkloadsMaterializedViewName
		}
	}

	return &WorkloadsDao{
		view: view,
	}
}

func (w *WorkloadsDao) DbDeleteWorkload(ctx context.Context, tenantId uint, workloadId uuid.UUID) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to delete workload %s", workloadId), nil)

	result := dao_utils.DB().WithContext(ctx).
		Table(WorkloadsTableName).
		Where("tenant_id = ?", tenantId).
		Delete(&models.Workload{}, workloadId)

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete workload with id %s, err: %v", workloadId, result.Error)
		return dbErr.WithDetails(result.Error)
	}

	return nil
}

func (w *WorkloadsDao) DbUpsertWorkload(ctx context.Context, workload models.Workload) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to upsert workload %s", workload.Name), nil)

	tx := dao_utils.DB().WithContext(ctx)

	query := tx.Table(WorkloadsTableName)

	query = omitWorkloadsEmptyStaticFields(query, workload)

	result := query.Save(&workload)

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to save workload %s, err %v", workload.Name, result.Error)
		return dbErr.WithDetails(result.Error)
	}

	return nil
}

// DbUpsertWorkloadFull upserts a workload, its children, and its elements in a single transaction.
// This avoids the overhead of 3 separate transactions (each with its own COMMIT + WAL fsync).
func (w *WorkloadsDao) DbUpsertWorkloadFull(ctx context.Context, tenantID uint, clusterID string,
	workload models.Workload, children []models.WorkloadChild, elements []models.WorkloadElement) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to upsert workload %s", workload.Name), nil)

	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to begin transaction for workload upsert: %v", tx.Error)
		return dbErr.WithDetails(tx.Error)
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	// 1. Upsert the workload row
	query := tx.Table(WorkloadsTableName)
	query = omitWorkloadsEmptyStaticFields(query, workload)
	if result := query.Save(&workload); result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to save workload %s, err %v", workload.Name, result.Error)
		return dbErr.WithDetails(result.Error)
	}

	// 2. Upsert children
	// if children fail, the workload row is still committed but an error is returned.
	tx.SavePoint("workload_children")
	var err *api.Error
	for _, child := range children {
		result := tx.Table(WorkloadsChildrenTableName).
			Clauses(clause.OnConflict{
				DoNothing: true,
			}).
			Create(&child)
		if result.Error != nil {
			tx.RollbackTo("workload_children")
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create child for workload %s, err %v", workload.Name, result.Error)
			err = dbErr.WithDetails(result.Error)
			break
		}
	}

	// 3. Upsert elements — skip if children failed (old behavior: elements were not attempted on children failure)
	if err == nil {
		tx.SavePoint("workload_elements")
		if elemErr := dbUpsertWorkloadElements(ctx, tx, tenantID, clusterID, workload.ID, elements); elemErr != nil {
			tx.RollbackTo("workload_elements")
			// Match existing UpsertWorkload behavior: log warning but don't fail the whole upsert
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to sync structure elements for workload %s (in DbUpsertWorkloadFull), error: %s",
				workload.ID, elemErr.Message)
		}
	}

	result := tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit workload upsert for %s: %v", workload.Name, result.Error)
		return dbErr.WithDetails(result.Error)
	}
	committed = true

	return err
}

// DbUpdateWorkloadPhase updateOnlyIfLastUpdateSmallerThan will update the phase if and only if the incoming request is newer than what we currently has in the DB
func DbUpdateWorkloadPhase(ctx context.Context, tenantId uint, workloadId string, workloadPhase api.WorkloadPatchFields, updateOnlyIfLastUpdateSmallerThan time.Time) *api.Error {
	_, dbErr := DbUpdateWorkloadPhaseWithResult(ctx, tenantId, workloadId, workloadPhase, updateOnlyIfLastUpdateSmallerThan)
	return dbErr
}

func DbUpdateWorkloadPhaseWithResult(ctx context.Context, tenantId uint, workloadId string, workloadPhase api.WorkloadPatchFields, updateOnlyIfLastUpdateSmallerThan time.Time) (bool, *api.Error) {
	const (
		workloadPhaseUpdated    = true
		workloadPhaseNotUpdated = false
	)

	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to update workload %s phase", workloadId), nil)

	tx := dao_utils.DB().WithContext(ctx)

	var deletedAt gorm.DeletedAt
	if workloadPhase.SoftDelete {
		deletedAt = gorm.DeletedAt{
			Time:  time.Now(),
			Valid: true,
		}
	}

	result := tx.
		Table(WorkloadsTableName).
		Where("tenant_id = ?", tenantId).
		Where("id = ?", workloadId).
		Where("phase_updated_at < ?", updateOnlyIfLastUpdateSmallerThan).
		Updates(map[string]interface{}{"phase": workloadPhase.Phase, "phase_message": workloadPhase.PhaseMessage, "deleted_at": deletedAt})
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update workload %s, err %v", workloadId, result.Error)
		return workloadPhaseNotUpdated, dbErr.WithDetails(result.Error)
	} else if result.RowsAffected == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Skipped phase update of %s to %s (possibly this update is outdated)", workloadId, workloadPhase.Phase)
		return workloadPhaseNotUpdated, nil
	}

	return workloadPhaseUpdated, nil
}

func (w *WorkloadsDao) DbCreateInternalWorkload(ctx context.Context, workload models.Workload, workloadInternalFields models.InternalCreatedWorkload) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to upsert workload %s", workload.Name), nil)

	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to begin a session with the db: %v", tx.Error)
		return dbErr.WithDetails(tx.Error)
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	result := tx.
		Table(WorkloadsTableName).
		Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(&workload)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to save workload %s, err %v", workload.Name, result.Error)
		return dbErr.WithDetails(result.Error)
	}

	updates := map[string]interface{}{
		WorkloadsColumnSourceApi:      workloadInternalFields.SourceApi,
		WorkloadsColumnPreemptibility: workloadInternalFields.Preemptibility,
	}
	if workloadInternalFields.Datasources != nil {
		updates[WorkloadsColumnDatasources] = workloadInternalFields.Datasources
	}
	if workloadInternalFields.Environments != nil {
		updates[WorkloadsColumnEnvironments] = workloadInternalFields.Environments
	}
	if workloadInternalFields.ExternalConnections != nil {
		updates[WorkloadsColumnExternalConnections] = workloadInternalFields.ExternalConnections
	}
	if workloadInternalFields.SourceApi != nil && *workloadInternalFields.SourceApi == api.SOURCEAPI_WORKLOADS_V2 {
		updates[WorkloadsColumnSupportDelete] = workload.SupportDelete
		updates[WorkloadsColumnSupportSuspend] = workload.SupportSuspend
	}

	result = tx.Table(WorkloadsTableName).
		Where("id = ? AND tenant_id = ?", workload.ID, workload.TenantID).
		Updates(updates)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to save internal creation workload data %s, err %v", workload.Name, result.Error)
		return dbErr.WithDetails(result.Error)
	}

	result = tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit workload upsert: %v", result.Error)
		return dbErr.WithDetails(result.Error)
	}
	committed = true

	return nil
}

func (w *WorkloadsDao) DbUpdateWorkloadConnectionMetadata(
	ctx context.Context,
	tenantId uint,
	workloadId uuid.UUID,
	environments models.EnvironmentArray,
	externalConnections models.ExternalConnectionArray,
) *api.Error {
	// Workloads-manager owns the full creation notification for WorkloadsV2
	// submissions. This update only adds workloads-service-specific connection
	// metadata that is known by the native request translator.
	updates := map[string]interface{}{}
	if len(environments) > 0 {
		updates[WorkloadsColumnEnvironments] = environments
	}
	if len(externalConnections) > 0 {
		updates[WorkloadsColumnExternalConnections] = externalConnections
	}
	if len(updates) == 0 {
		return nil
	}

	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to update workload %s connection metadata", workloadId.String()), nil)
	result := dao_utils.DB().
		WithContext(ctx).
		Table(WorkloadsTableName).
		Where("id = ?", workloadId).
		Where("tenant_id = ?", tenantId).
		Updates(updates)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update workload %s connection metadata, err %v", workloadId.String(), result.Error)
		return dbErr.WithDetails(result.Error)
	}
	if result.RowsAffected == 0 {
		return api.NewError(http.StatusNotFound, fmt.Sprintf("Workload <%s> not found", workloadId.String()), nil)
	}

	return nil
}

func (w *WorkloadsDao) DbGetWorkload(ctx context.Context, tenantId uint, workloadId uuid.UUID) (*models.WorkloadView, *api.Error) {
	return dbGetWorkload(ctx, tenantId, workloadId)
}

func (w *WorkloadsDao) DbGetWorkloadAdditionalFields(ctx context.Context, tenantId uint, workloadId uuid.UUID) (models.AdditionalFields, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get additional fields of workload %s", workloadId), nil)
	workload := &models.Workload{}

	result := dao_utils.DB().
		WithContext(ctx).
		Table(WorkloadsTableName).
		Select(WorkloadsColumnAdditionalFields).
		Where(fmt.Sprintf("%s = ?", WorkloadsColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s = ?", WorkloadsColumnId), workloadId).
		Unscoped().
		First(workload)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get additional fields of workload with id %s, err: %v", workloadId, result.Error)
		}
		return nil, dbErr.WithDetails(result.Error)
	}

	return workload.AdditionalFields, nil
}

func (w *WorkloadsDao) DbCountWorkloads(ctx context.Context, tenantId uint, permittedScopes *openapi.PermittedScopes,
	deleted bool, filterBy []validators.ParameterizedFilter, search string, precise bool, countFromTable bool) (int64, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to count workloads", nil)

	tx := dao_utils.DB().WithContext(ctx)

	viewNameToFetchFrom := w.view
	if countFromTable {
		viewNameToFetchFrom = WorkloadsTableName
	} else if precise {
		viewNameToFetchFrom = WorkloadsViewName
	}

	query := tx.
		Table(viewNameToFetchFrom).
		Where("tenant_id = ?", tenantId)

	if permittedScopes != nil {
		if permittedScopes.System == nil || !*permittedScopes.System {
			query = query.Where(getPermittedScopesFilter(*permittedScopes))
		}
	}

	if deleted {
		query = query.Unscoped().Where("deleted_at IS NOT NULL")
	} else {
		query = query.Where("deleted_at IS NULL")
	}

	query = dao_utils.ApplyParameterizedFilters(query, filterBy)

	if search != "" {
		query = dao_utils.ApplyTSSearchParam(query, search, "search_vector")
	}

	var count int64
	result := query.
		Count(&count)

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count workloads, err: %v", result.Error)
		return 0, dbErr.WithDetails(result.Error)
	}

	return count, nil
}

type OrgUnitRankingCounts struct {
	OrgUnitId    string
	RunningCount int64
	PendingCount int64
}

func (w *WorkloadsDao) DbCountWorkloadsForRanking(
	ctx context.Context,
	tenantId uint,
	clusterId string,
	orgUnitIds []string,
	orgUnitType string,
	nodePoolName *string,
) (map[string]*OrgUnitRankingCounts, *api.Error) {
	if len(orgUnitIds) == 0 {
		return make(map[string]*OrgUnitRankingCounts), nil
	}

	orgUnitColumn := w.getRankingOrgUnitColumn(orgUnitType)
	results, err := w.buildAndExecuteRankingCountQuery(ctx, tenantId, clusterId, orgUnitIds, orgUnitColumn, orgUnitType, nodePoolName)
	if err != nil {
		return nil, err
	}

	return w.convertRankingCountsToMap(results), nil
}

func (w *WorkloadsDao) getRankingOrgUnitColumn(orgUnitType string) string {
	if orgUnitType == "department" {
		return WorkloadsColumnDepartmentId
	}
	return WorkloadsColumnProjectId
}

// buildAndExecuteRankingCountQuery builds and executes a query for counting workloads by phase for ranking
// Handles both with and without node pool filtering in a single implementation
// Running workloads: nodePoolName is in current_nodepools (if provided)
// Pending workloads: nodePoolName is in requested_nodepools AND requested_nodepools length is 1 (if provided)
func (w *WorkloadsDao) buildAndExecuteRankingCountQuery(
	ctx context.Context,
	tenantId uint,
	clusterId string,
	orgUnitIds []string,
	orgUnitColumn string,
	orgUnitType string,
	nodePoolName *string,
) ([]*OrgUnitRankingCounts, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to count workloads for ranking", nil)

	var runningFilter, pendingFilter string
	var params []interface{}

	if nodePoolName != nil && *nodePoolName != "" {
		runningFilter = "phase = 'Running' AND ?::text = ANY(current_nodepools)"
		pendingFilter = "phase = 'Pending' AND cardinality(requested_nodepools) = 1 AND ?::text = ANY(requested_nodepools)"
		params = append(params, *nodePoolName, *nodePoolName)
	} else {
		runningFilter = "phase = 'Running'"
		pendingFilter = "phase = 'Pending'"
	}

	params = append(params, tenantId, clusterId)

	inClause := ""
	for _, orgId := range orgUnitIds {
		if inClause != "" {
			inClause += ","
		}
		inClause += "?"
		params = append(params, orgId)
	}

	sql := fmt.Sprintf(`
		SELECT 
			%s,
			COUNT(*) FILTER (WHERE %s) as running_count,
			COUNT(*) FILTER (WHERE %s) as pending_count
		FROM %s
		WHERE tenant_id = ?
			AND cluster_id = ?
			AND phase IN ('Running', 'Pending')
			AND deleted_at IS NULL
			AND %s IN (%s)
		GROUP BY %s
	`, orgUnitColumn, runningFilter, pendingFilter, w.view, orgUnitColumn, inClause, orgUnitColumn)

	query := dao_utils.DB().WithContext(ctx).Raw(sql, params...)

	if orgUnitType == "department" {
		return w.executeRankingCountQueryForDepartments(ctx, query, dbErr)
	}
	return w.executeRankingCountQueryForProjects(ctx, query, dbErr)
}

func (w *WorkloadsDao) executeRankingCountQueryForProjects(
	ctx context.Context,
	query *gorm.DB,
	dbErr *api.Error,
) ([]*OrgUnitRankingCounts, *api.Error) {
	type ProjectCountResult struct {
		OrgUnitId    string `gorm:"column:project_id"`
		RunningCount int64  `gorm:"column:running_count"`
		PendingCount int64  `gorm:"column:pending_count"`
	}

	var results []ProjectCountResult
	if err := query.Find(&results).Error; err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count workloads for project ranking, err: %v", err)
		return nil, dbErr.WithDetails(err)
	}

	counts := make([]*OrgUnitRankingCounts, len(results))
	for i, r := range results {
		counts[i] = &OrgUnitRankingCounts{
			OrgUnitId:    r.OrgUnitId,
			RunningCount: r.RunningCount,
			PendingCount: r.PendingCount,
		}
	}
	return counts, nil
}

func (w *WorkloadsDao) executeRankingCountQueryForDepartments(
	ctx context.Context,
	query *gorm.DB,
	dbErr *api.Error,
) ([]*OrgUnitRankingCounts, *api.Error) {
	type DeptCountResult struct {
		OrgUnitId    string `gorm:"column:department_id"`
		RunningCount int64  `gorm:"column:running_count"`
		PendingCount int64  `gorm:"column:pending_count"`
	}

	var results []DeptCountResult
	if err := query.Find(&results).Error; err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count workloads for department ranking, err: %v", err)
		return nil, dbErr.WithDetails(err)
	}

	counts := make([]*OrgUnitRankingCounts, len(results))
	for i, r := range results {
		counts[i] = &OrgUnitRankingCounts{
			OrgUnitId:    r.OrgUnitId,
			RunningCount: r.RunningCount,
			PendingCount: r.PendingCount,
		}
	}
	return counts, nil
}

func (w *WorkloadsDao) convertRankingCountsToMap(results []*OrgUnitRankingCounts) map[string]*OrgUnitRankingCounts {
	countsMap := make(map[string]*OrgUnitRankingCounts, len(results))
	for _, result := range results {
		countsMap[result.OrgUnitId] = result
	}
	return countsMap
}

func (w *WorkloadsDao) DbGetWorkloads(
	ctx context.Context,
	tenantId uint,
	permittedScopes *openapi.PermittedScopes,
	deleted bool,
	offset int32,
	limit int32,
	sortOrder string,
	sortBy string,
	filterBy []validators.ParameterizedFilter,
	search string,
) (*[]models.WorkloadView, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to list workloads", nil)

	tx := dao_utils.DB().WithContext(ctx)

	workloads := &[]models.WorkloadView{}

	query := tx.
		Table(w.view).
		Where("tenant_id = ?", tenantId)

	if permittedScopes != nil {
		if permittedScopes.System == nil || !*permittedScopes.System {
			query = query.Where(getPermittedScopesFilter(*permittedScopes))
		}
	}

	if deleted {
		query = query.Unscoped().Where("deleted_at IS NOT NULL")
	}

	query = dao_utils.ApplyParameterizedFilters(query, filterBy)

	var err error
	if sortBy != "" {
		query, err = dao_utils.ApplySortByParam(query, sortBy, sortOrder)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sort workloads, err: %v", err)
			return &[]models.WorkloadView{}, dbErr.WithDetails(err)
		}
	}
	query = query.Order("id DESC")

	if offset != 0 {
		query = query.Offset(int(offset))
	}

	if limit != 0 {
		query = query.Limit(int(limit))
	}

	if search != "" {
		query = dao_utils.ApplyTSSearchParam(query, search, "search_vector")
	}

	result := query.
		Find(workloads)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list workloads, err: %v", result.Error)
		return &[]models.WorkloadView{}, dbErr.WithDetails(result.Error)
	}

	// Redact sensitive environment variables before returning
	redactWorkloadsEnvironmentVariables(ctx, workloads)

	return workloads, nil
}

func (w *WorkloadsDao) DbSyncWorkloads(
	ctx context.Context,
	tenantId uint,
	projectId string,
	clusterId string,
	workloads []models.Workload,
	to *uuid.UUID,
	from *uuid.UUID,
	excludingIds []string,
	deletionFilters []validators.ParameterizedFilter,
) ([]string, []string, []string, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to sync workloads", nil)

	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to begin a session with the db: %v", tx.Error)
		return nil, nil, nil, dbErr.WithDetails(tx.Error)
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)
	tx.SavePoint("sp")

	syncedWorkloads := []string{}
	failedWorkloads := []string{}
	for _, workload := range workloads {
		query := tx.Table(WorkloadsTableName)

		query = omitWorkloadsEmptyStaticFields(query, workload)

		result := query.Save(&workload)
		if result.Error != nil {
			tx.RollbackTo("sp") // Rollback to last saved point
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to save workload %s in project %s, err %v", workload.Name, workload.ProjectName, result.Error)
			failedWorkloads = append(failedWorkloads, workload.ID.String())
		} else {
			tx.SavePoint("sp")
			syncedWorkloads = append(syncedWorkloads, workload.ID.String())
		}

		excludingIds = append(excludingIds, workload.ID.String())
	}

	if projectId == "" && clusterId == "" {
		return nil, nil, nil, api.NewError(http.StatusBadRequest, "Workloads sync without scope is not supported", nil)
	}

	// if one of projectId or clusterId is empty, we delete by the other one only
	deleteQuery := ConstructDeleteQueryForProjectCluster(tx, WorkloadsTableName, tenantId, clusterId, projectId, from, to, excludingIds, "")
	if len(deletionFilters) > 0 {
		deleteQuery = dao_utils.ApplyParameterizedFilters(deleteQuery, deletionFilters)
	}

	deletedWorkloads := []models.Workload{}
	result := deleteQuery.
		Clauses(clause.Returning{Columns: []clause.Column{{Name: WorkloadsColumnId}}}).
		Delete(&deletedWorkloads)

	if result.Error != nil {
		return nil, nil, nil, dbErr.WithDetails(result.Error)
	}

	deletedWorkloadIds := make([]string, 0, len(deletedWorkloads))
	for _, workload := range deletedWorkloads {
		deletedWorkloadIds = append(deletedWorkloadIds, workload.ID.String())
	}

	result = tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit the workloads sync: %v", result.Error)
		return nil, nil, nil, dbErr.WithDetails(result.Error)
	}
	committed = true

	return syncedWorkloads, failedWorkloads, deletedWorkloadIds, nil
}

func omitWorkloadsEmptyStaticFields(query *gorm.DB, workload models.Workload) *gorm.DB {
	omitColumns := []string{}

	if workload.ClusterID.String() == "" {
		omitColumns = append(omitColumns, "cluster_id")
	}
	if workload.ProjectID == "" {
		omitColumns = append(omitColumns, "project_id")
	}
	if workload.ProjectName == "" {
		omitColumns = append(omitColumns, "project_name")
	}
	if workload.DepartmentID == "" {
		omitColumns = append(omitColumns, "department_id")
	}
	if workload.DepartmentName == "" {
		omitColumns = append(omitColumns, "department_name")
	}
	if workload.SubmittedBy == "" {
		omitColumns = append(omitColumns, "submitted_by")
	}
	if workload.Preemptibility == nil {
		omitColumns = append(omitColumns, WorkloadsColumnPreemptibility)
	}
	if workload.ExternalConnections == nil {
		omitColumns = append(omitColumns, WorkloadsColumnExternalConnections)
	}
	// support_suspend is set by the control-plane at creation time (V2 workloads) or by
	// the cluster sync for V1 workloads. When the cluster sends false (its zero value), we
	// skip the update so that a CP-set true is never clobbered by a subsequent cluster sync.
	if !workload.SupportSuspend {
		omitColumns = append(omitColumns, WorkloadsColumnSupportSuspend)
	}

	if len(omitColumns) > 0 {
		query = query.Omit(omitColumns...)
	}

	return query
}

func (w *WorkloadsDao) DbUpsertWorkloadsChildren(ctx context.Context, workloadsChildren []models.WorkloadChild, failedWorkloads []string) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to create workloads children", nil)

	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to begin a session with the db: %v", tx.Error)
		return dbErr.WithDetails(tx.Error)
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	failedWorkloadsMap := createMapFromSlice(failedWorkloads)

	for _, child := range workloadsChildren {
		if _, ok := failedWorkloadsMap[child.WorkloadID.String()]; ok {
			continue
		}

		result := tx.Table(WorkloadsChildrenTableName).
			Clauses(clause.OnConflict{
				DoNothing: true,
			}).
			Create(&child)
		if result.Error != nil {
			return dbErr.WithDetails(result.Error)
		}
	}

	result := tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit the workloads children: %v", result.Error)
		return dbErr.WithDetails(result.Error)
	}
	committed = true

	return nil
}

func (w *WorkloadsDao) DbUpsertWorkloadsPodsChildren(ctx context.Context, workloadsChildren []models.WorkloadChild, filterOutPods []string) *api.Error {
	var childrenToCreate []models.WorkloadChild
	filterOutPodsMap := createMapFromSlice(filterOutPods)

	for _, child := range workloadsChildren {
		if _, ok := filterOutPodsMap[child.ChildID.String()]; ok {
			continue
		}

		childrenToCreate = append(childrenToCreate, child)
	}

	return w.DbUpsertWorkloadsChildren(ctx, childrenToCreate, nil)
}

func getPermittedScopesFilter(permittedScopes openapi.PermittedScopes) string {
	return authorization_utils.GetPermittedScopesDbFilter(permittedScopes)
}

// applyRequestedNodepoolFilter filters workloads by requested_nodepools array.
// Only includes workloads that requested exactly one nodepool matching the given name.
func applyRequestedNodepoolFilter(query *gorm.DB, nodepoolName string) *gorm.DB {
	return query.Where("array_length(requested_nodepools, 1) = 1").
		Where("? = ANY(requested_nodepools)", nodepoolName)
}

// applyCurrentNodepoolFilter filters workloads by current_nodepools array.
// Includes workloads that are currently running in the specified nodepool.
func applyCurrentNodepoolFilter(query *gorm.DB, nodepoolName string) *gorm.DB {
	return query.Where("? = ANY(current_nodepools)", nodepoolName)
}

// buildTelemetryQuery builds a base query with common filters and GROUP BY for telemetry queries.
// Returns the query before SELECT and execution, allowing callers to customize the SELECT clause.
func (w *WorkloadsDao) buildTelemetryQuery(
	ctx context.Context,
	tenantId uint,
	clusterId string,
	nodepoolName string,
	departmentId string,
	projectId string,
	permittedScopes *openapi.PermittedScopes,
	groupBy [][]string,
	phases []api.Phase,
) *gorm.DB {
	tx := dao_utils.DB().WithContext(ctx)
	query := tx.Table(w.view).
		Where("tenant_id = ?", tenantId).
		Where("deleted_at IS NULL")

	if permittedScopes != nil {
		if permittedScopes.System == nil || !*permittedScopes.System {
			query = query.Where(getPermittedScopesFilter(*permittedScopes))
		}
	}
	if clusterId != "" {
		query = query.Where("cluster_id = ?", clusterId)
	}
	if nodepoolName != "" {
		// For pending workloads, filter by requested_nodepools that contain only nodepoolName.
		usePendingNodepoolFilter := len(phases) == 1 && phases[0] == api.PHASE_PENDING
		if usePendingNodepoolFilter {
			query = applyRequestedNodepoolFilter(query, nodepoolName)
		} else {
			query = applyCurrentNodepoolFilter(query, nodepoolName)
		}
	}
	if departmentId != "" {
		query = query.Where("department_id = ?", departmentId)
	}
	if projectId != "" {
		query = query.Where("project_id = ?", projectId)
	}
	if len(phases) > 0 {
		query = query.Where("phase IN ?", phases)
	}
	for _, groups := range groupBy {
		query = query.Group(groups[0])
		if len(groups) > 1 {
			query = query.Group(groups[1])
		}
	}

	return query
}

func (w *WorkloadsDao) dbGetWorkloadTelemetry(
	ctx context.Context,
	tenantId uint,
	clusterId string,
	nodepoolName string,
	departmentId string,
	projectId string,
	permittedScopes *openapi.PermittedScopes,
	groupBy [][]string,
	dataField string,
	phases []api.Phase,
) (*[]telemetry_utils.TelemetryGroupValue, error) {
	telemetryGroupValues := []telemetry_utils.TelemetryGroupValue{}

	selectPhrase := fmt.Sprintf("%s as data", dataField)
	for i, groups := range groupBy {
		selectPhrase = fmt.Sprintf("%s, %s as group%d_main_value", selectPhrase, groups[0], i+1)
		if len(groups) > 1 {
			selectPhrase = fmt.Sprintf("%s, %s as group%d_secondary_value", selectPhrase, groups[1], i+1)
		}
	}

	query := w.buildTelemetryQuery(ctx, tenantId, clusterId, nodepoolName, departmentId, projectId, permittedScopes, groupBy, phases).
		Select(selectPhrase)

	result := query.Scan(&telemetryGroupValues)

	if result.Error != nil {
		return &telemetryGroupValues, result.Error
	}

	return &telemetryGroupValues, nil
}

func (w *WorkloadsDao) DbGetWorkloadCount(
	ctx context.Context,
	tenantId uint,
	clusterId string,
	nodepoolName string,
	departmentId string,
	projectId string,
	permittedScopes *openapi.PermittedScopes,
	groupBy [][]string,
	phases []api.Phase,
) (*[]telemetry_utils.TelemetryGroupValue, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get workload count", nil)

	workloadTelemetry, err := w.dbGetWorkloadTelemetry(ctx, tenantId, clusterId, nodepoolName, departmentId, projectId, permittedScopes, groupBy, "COUNT(id)", phases)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get workload count, err: %v", err)
		return workloadTelemetry, dbErr.WithDetails(err)
	}

	return workloadTelemetry, nil
}

func (w *WorkloadsDao) DbGetGpuAllocation(
	ctx context.Context,
	tenantId uint,
	clusterId string,
	nodepoolName string,
	departmentId string,
	projectId string,
	permittedScopes *openapi.PermittedScopes,
	groupBy [][]string,
	phases []api.Phase,
) (*[]telemetry_utils.TelemetryGroupValue, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get gpu allocation", nil)

	workloadTelemetry, err := w.dbGetWorkloadTelemetry(ctx, tenantId, clusterId, nodepoolName, departmentId, projectId, permittedScopes, groupBy, "COALESCE(SUM(allocated_gpu), 0)", phases)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get gpu allocation, err: %v", err)
		return workloadTelemetry, dbErr.WithDetails(err)
	}

	return workloadTelemetry, nil
}

func (w *WorkloadsDao) DbGetIdleAllocatedGpus(
	ctx context.Context,
	tenantId uint,
	clusterId string,
	nodepoolName string,
	departmentId string,
	projectId string,
	permittedScopes *openapi.PermittedScopes,
	groupBy [][]string,
	phases []api.Phase,
) (*[]telemetry_utils.TelemetryGroupValue, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get idle allocated gpus", nil)

	workloadTelemetry, err := w.dbGetWorkloadTelemetry(ctx, tenantId, clusterId, nodepoolName, departmentId, projectId, permittedScopes, groupBy, "COALESCE(SUM(idle_allocated_gpus), 0)", phases)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get idle allocated gpus, err: %v", err)
		return workloadTelemetry, dbErr.WithDetails(err)
	}

	return workloadTelemetry, nil
}

func (w *WorkloadsDao) DbGetWorkloadPhase(ctx context.Context, tenantId uint, clusterId *string, workloadId string) (api.Phase, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get workload %s phase", workloadId), nil)

	tx := dao_utils.DB().WithContext(ctx)

	workload := models.Workload{}

	result := tx.Model(workload).
		Unscoped().
		Where("tenant_id = ?", tenantId).
		Where("id = ?", workloadId)

	if clusterId != nil {
		result.Where("cluster_id = ?", *clusterId)
	}

	result.First(&workload)

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get workload with id %s phase, err: %v", workloadId, result.Error)
		return workload.Phase, dbErr.WithDetails(result.Error)
	}

	return workload.Phase, nil
}

func (w *WorkloadsDao) DbGetChildren(ctx context.Context, workloadId uuid.UUID, childTypes *[]string) (*[]models.WorkloadChild, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get children of workload %s with childTypes %v", workloadId, childTypes), nil)
	tx := dao_utils.DB().WithContext(ctx)

	children := &[]models.WorkloadChild{}
	query := tx.
		Table(WorkloadsChildrenTableName).
		Where("workload_id = ?", workloadId)

	if childTypes != nil {
		query = query.Where("child_type IN ?", *childTypes)
	}

	result := query.Find(children)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get children of workload with id %s with childTypes %v, err: %v",
			workloadId, childTypes, result.Error)
		return &[]models.WorkloadChild{}, dbErr.WithDetails(result.Error)
	}
	return children, nil
}

func (w *WorkloadsDao) DbGetWorkloadClusterId(ctx context.Context, tenantId uint, workloadId string) (uuid.UUID, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get clusterId of workload with id %s", workloadId), nil)
	tx := dao_utils.DB().WithContext(ctx)

	workload := &models.Workload{}
	query := tx.Table(WorkloadsTableName).
		Select("cluster_id").
		Where("tenant_id = ?", tenantId).
		Where("id = ?", workloadId).
		Unscoped()

	result := query.First(workload)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get clusterId of workload with id %s, err: %v", workloadId, result.Error)
		return uuid.UUID{}, dbErr.WithDetails(result.Error)
	}

	return workload.ClusterID, nil
}

// GetWorkloadFromDb returns the workload from the db
func GetWorkloadsViewByIdsFromDb(ctx context.Context, workloadIds []string, excludeDeleted *bool) ([]*models.WorkloadView, *api.Error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return nil, apiErr
	}

	if len(workloadIds) == 0 {
		return []*models.WorkloadView{}, nil
	}

	ids := make([]uuid.UUID, len(workloadIds))
	for i, id := range workloadIds {
		currUuid, err := uuid.Parse(id)
		if err != nil {
			return nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("Failed to parse workload id %s", id), nil)
		}
		ids[i] = currUuid
	}

	return dbGetWorkloads(ctx, tenantId, ids, excludeDeleted)
}

func GetWorkloadViewFromDb(ctx context.Context, workloadId string) (*models.WorkloadView, *api.Error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return nil, apiErr
	}

	id, err := uuid.Parse(workloadId)
	if err != nil {
		return nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("workload Id must be a valid uuid. <%s>", workloadId), err)
	}

	return dbGetWorkload(ctx, tenantId, id)
}

func dbGetWorkload(ctx context.Context, tenantId uint, workloadId uuid.UUID) (*models.WorkloadView, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get workload %s", workloadId), nil)

	tx := dao_utils.DB().WithContext(ctx)

	workload := &models.WorkloadView{}
	result := tx.
		Table(WorkloadsViewName).
		Unscoped().
		Where("tenant_id = ?", tenantId).
		Where("id = ?", workloadId).
		First(workload)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get workload with id %s, err: %v", workloadId, result.Error)
		}
		return nil, dbErr.WithDetails(result.Error)
	}

	// Redact sensitive environment variables before returning
	redactWorkloadEnvironmentVariables(ctx, workload)

	return workload, nil
}

// dbGetWorkload queries the db for a workload
// maximum of 100 ids in one query
func dbGetWorkloads(ctx context.Context, tenantId uint, workloadIds []uuid.UUID, excludeDeleted *bool) ([]*models.WorkloadView, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get workload %v", workloadIds), nil)

	if len(workloadIds) > 100 {
		return nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("Maximum of 100 ids can be queried at once, got %d", len(workloadIds)), nil)
	}

	tx := dao_utils.DB().WithContext(ctx)

	workloads := []*models.WorkloadView{}
	result := tx.
		Table(WorkloadsViewName).
		Unscoped().
		Where("tenant_id = ?", tenantId).
		Where("id IN ?", workloadIds)

	if excludeDeleted != nil && *excludeDeleted {
		result = result.Where(fmt.Sprintf("%s IS NULL", WorkloadsDeletedAtColumnName))
	}

	result.Scan(&workloads)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get workload with id %v, err: %v", workloadIds, result.Error)
		return nil, dbErr.WithDetails(result.Error)
	}

	// Redact sensitive environment variables before returning
	redactWorkloadsEnvironmentVariablesPointers(ctx, workloads)

	return workloads, nil
}

// DeleteWorkloadsOfClusterId deletes all workloads within a cluster
func DeleteWorkloadsOfClusterId(ctx context.Context, tenantId string, clusterId string) error {
	// By default, GORM's Delete() method only performs a soft delete if the model has gorm.DeletedAt.
	// Using .Unscoped() ensures a hard delete, permanently removing the record from the database.
	result := dao_utils.DB().WithContext(ctx).Unscoped().
		Table(WorkloadsTableName).
		Where(fmt.Sprintf("%s = ?", WorkloadsColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s = ?", WorkloadsColumnClusterId), clusterId).
		Delete(&models.Workload{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete workload for cluster %s in tenant %s. err: %w", clusterId, tenantId, result.Error)
	}
	return nil
}

// DeleteWorkloadsOfTenantId deletes all workloads within a cluster
func DeleteWorkloadsOfTenantId(ctx context.Context, tenantId string) error {
	// By default, GORM's Delete() method only performs a soft delete if the model has gorm.DeletedAt.
	// Using .Unscoped() ensures a hard delete, permanently removing the record from the database.
	result := dao_utils.DB().WithContext(ctx).Unscoped().
		Table(WorkloadsTableName).
		Where(fmt.Sprintf("%s = ?", WorkloadsColumnTenantId), tenantId).
		Delete(&models.Workload{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete workload in tenant %s. err: %w", tenantId, result.Error)
	}
	return nil
}

// redactWorkloadEnvironmentVariables redacts environment variables that start with "USERCREDENTIAL:" prefix
// by removing them entirely from the map if the current user is not the workload submitter
func redactWorkloadEnvironmentVariables(ctx context.Context, workload *models.WorkloadView) {
	if workload == nil || workload.EnvironmentVariables == nil {
		return
	}

	// Check if current user is the workload submitter
	currentSubject := runaictx.GetString(ctx, runaictx.CfSubjectID)
	if currentSubject != nil && *currentSubject == workload.SubmittedBy {
		return
	}

	// Only redact if current user is not the submitter
	var toDelete []string
	for key, value := range workload.EnvironmentVariables {
		if strings.HasPrefix(value, "USERCREDENTIAL:") {
			toDelete = append(toDelete, key)
		}
	}

	for _, key := range toDelete {
		delete(workload.EnvironmentVariables, key)
	}
}

// redactWorkloadsEnvironmentVariables redacts environment variables for a slice of workloads
func redactWorkloadsEnvironmentVariables(ctx context.Context, workloads *[]models.WorkloadView) {
	if workloads == nil {
		return
	}

	for i := range *workloads {
		redactWorkloadEnvironmentVariables(ctx, &(*workloads)[i])
	}
}

// redactWorkloadsEnvironmentVariablesPointers redacts environment variables for a slice of pointers to workloads
func redactWorkloadsEnvironmentVariablesPointers(ctx context.Context, workloads []*models.WorkloadView) {
	if workloads == nil {
		return
	}

	for i := range workloads {
		redactWorkloadEnvironmentVariables(ctx, workloads[i])
	}
}
