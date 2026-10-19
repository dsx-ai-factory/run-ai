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

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	querydbenrich "github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi/dbenrich"

	telemetry_utils "github.com/dsx-ai-factory/run-ai/runai-application-packages/telemetry"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"

	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/dbenrich"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/google/uuid"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
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

	_, err := runaidb.DB().WithContext(ctx).
		Table(WorkloadsTableName).
		Where("tenant_id = ?", tenantId).
		Delete(&models.Workload{}, workloadId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete workload with id %s, err: %v", workloadId, err)
		return dbErr.WithDetails(err)
	}

	return nil
}

func (w *WorkloadsDao) DbUpsertWorkload(ctx context.Context, workload models.Workload) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to upsert workload %s", workload.Name), nil)

	tx := runaidb.DB().WithContext(ctx)

	query := tx.Table(WorkloadsTableName)

	query = omitWorkloadsEmptyStaticFields(query, workload)

	_, err := query.Save(&workload)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to save workload %s, err %v", workload.Name, err)
		return dbErr.WithDetails(err)
	}

	if err := resetSupportSuspendOfV1Workloads(tx, workload.TenantID, notSuspendableIds(workload)); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to reset support_suspend of workload %s: %v", workload.Name, err)
		return dbErr.WithDetails(err)
	}

	return nil
}

// DbUpsertWorkloadFull upserts a workload, its children, and its elements in a single transaction.
// This avoids the overhead of 3 separate transactions (each with its own COMMIT + WAL fsync).
func (w *WorkloadsDao) DbUpsertWorkloadFull(ctx context.Context, tenantID uint, clusterID string,
	workload models.Workload, children []models.WorkloadChild, elements []models.WorkloadElement) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to upsert workload %s", workload.Name), nil)

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	// 1. Upsert the workload row
	query := tx.Table(WorkloadsTableName)
	query = omitWorkloadsEmptyStaticFields(query, workload)
	if _, saveErr := query.Save(&workload); saveErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to save workload %s, err %v", workload.Name, saveErr)
		return dbErr.WithDetails(saveErr)
	}
	if resetErr := resetSupportSuspendOfV1Workloads(tx, tenantID, notSuspendableIds(workload)); resetErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to reset support_suspend of workload %s: %v", workload.Name, resetErr)
		return dbErr.WithDetails(resetErr)
	}

	// 2. Upsert children
	// if children fail, the workload row is still committed but an error is returned.
	tx.SavePoint("workload_children")
	var err *api.Error
	for _, child := range children {
		_, createErr := tx.Table(WorkloadsChildrenTableName).
			Clauses(clause.OnConflict{
				DoNothing: true,
			}).
			Create(&child)
		if createErr != nil {
			tx.RollbackTo("workload_children")
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create child for workload %s, err %v", workload.Name, createErr)
			err = dbErr.WithDetails(createErr)
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

	commitErr := tx.Commit()
	if commitErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit workload upsert for %s: %v", workload.Name, commitErr)
		return dbErr.WithDetails(commitErr)
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

	tx := runaidb.DB().WithContext(ctx)

	var deletedAt gorm.DeletedAt
	if workloadPhase.SoftDelete {
		deletedAt = gorm.DeletedAt{
			Time:  time.Now(),
			Valid: true,
		}
	}

	rowsAffected, err := tx.
		Table(WorkloadsTableName).
		Where("tenant_id = ?", tenantId).
		Where("id = ?", workloadId).
		Where("phase_updated_at < ?", updateOnlyIfLastUpdateSmallerThan).
		Updates(map[string]interface{}{"phase": workloadPhase.Phase, "phase_message": workloadPhase.PhaseMessage, "deleted_at": deletedAt})
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update workload %s, err %v", workloadId, err)
		return workloadPhaseNotUpdated, dbErr.WithDetails(err)
	} else if rowsAffected == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Skipped phase update of %s to %s (possibly this update is outdated)", workloadId, workloadPhase.Phase)
		return workloadPhaseNotUpdated, nil
	}

	return workloadPhaseUpdated, nil
}

func (w *WorkloadsDao) DbCreateInternalWorkload(ctx context.Context, workload models.Workload, workloadInternalFields models.InternalCreatedWorkload) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to upsert workload %s", workload.Name), nil)

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	_, err := tx.
		Table(WorkloadsTableName).
		Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(&workload)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to save workload %s, err %v", workload.Name, err)
		return dbErr.WithDetails(err)
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

	_, err = tx.Table(WorkloadsTableName).
		Where("id = ? AND tenant_id = ?", workload.ID, workload.TenantID).
		Updates(updates)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to save internal creation workload data %s, err %v", workload.Name, err)
		return dbErr.WithDetails(err)
	}

	err = tx.Commit()
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit workload upsert: %v", err)
		return dbErr.WithDetails(err)
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
	rowsAffected, err := runaidb.DB().
		WithContext(ctx).
		Table(WorkloadsTableName).
		Where("id = ?", workloadId).
		Where("tenant_id = ?", tenantId).
		Updates(updates)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update workload %s connection metadata, err %v", workloadId.String(), err)
		return dbErr.WithDetails(err)
	}
	if rowsAffected == 0 {
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

	err := runaidb.DB().
		WithContext(ctx).
		Table(WorkloadsTableName).
		Select(WorkloadsColumnAdditionalFields).
		Where(fmt.Sprintf("%s = ?", WorkloadsColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s = ?", WorkloadsColumnId), workloadId).
		Unscoped().
		First(workload)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get additional fields of workload with id %s, err: %v", workloadId, err)
		}
		return nil, dbErr.WithDetails(err)
	}

	return workload.AdditionalFields, nil
}

func (w *WorkloadsDao) DbCountWorkloads(ctx context.Context, tenantId uint, permittedScopes *openapi.PermittedScopes,
	deleted bool, filterBy []queryapi.Filter, search string, precise bool, countFromTable bool) (int64, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to count workloads", nil)

	tx := runaidb.DB().WithContext(ctx)

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

	query = query.WithEnricher(querydbenrich.Filters(filterBy))

	if search != "" {
		query = query.WithEnricher(dbenrich.TSSearch(search, "search_vector"))
	}

	var count int64
	err := query.Count(&count)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count workloads, err: %v", err)
		return 0, dbErr.WithDetails(err)
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

	query := runaidb.DB().WithContext(ctx).Raw(sql, params...)

	if orgUnitType == "department" {
		return w.executeRankingCountQueryForDepartments(ctx, query, dbErr)
	}
	return w.executeRankingCountQueryForProjects(ctx, query, dbErr)
}

func (w *WorkloadsDao) executeRankingCountQueryForProjects(
	ctx context.Context,
	query orm.DB,
	dbErr *api.Error,
) ([]*OrgUnitRankingCounts, *api.Error) {
	type ProjectCountResult struct {
		OrgUnitId    string `gorm:"column:project_id"`
		RunningCount int64  `gorm:"column:running_count"`
		PendingCount int64  `gorm:"column:pending_count"`
	}

	var results []ProjectCountResult
	if err := query.Find(&results); err != nil {
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
	query orm.DB,
	dbErr *api.Error,
) ([]*OrgUnitRankingCounts, *api.Error) {
	type DeptCountResult struct {
		OrgUnitId    string `gorm:"column:department_id"`
		RunningCount int64  `gorm:"column:running_count"`
		PendingCount int64  `gorm:"column:pending_count"`
	}

	var results []DeptCountResult
	if err := query.Find(&results); err != nil {
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
	filterBy []queryapi.Filter,
	search string,
) (*[]models.WorkloadView, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to list workloads", nil)

	tx := runaidb.DB().WithContext(ctx)

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

	query = query.WithEnricher(querydbenrich.Filters(filterBy))

	if sortBy != "" {
		query = query.Order(querydbenrich.FormatSortClause(sortBy, sortOrder))
	}
	query = query.Order("id DESC")

	if offset != 0 {
		query = query.Offset(int(offset))
	}

	if limit != 0 {
		query = query.Limit(int(limit))
	}

	if search != "" {
		query = query.WithEnricher(dbenrich.TSSearch(search, "search_vector"))
	}

	err := query.Find(workloads)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list workloads, err: %v", err)
		return &[]models.WorkloadView{}, dbErr.WithDetails(err)
	}

	// Redact sensitive environment variables before returning
	redactWorkloadsEnvironmentVariables(ctx, workloads)

	return workloads, nil
}

// DbGetWorkloadsSlim lists the active workloads from the workloads table instead of the view, sorted by id ascending:
// only the WorkloadSlim columns, without pod aggregates, and including workloads created after the last materialized
// view refresh.
func (w *WorkloadsDao) DbGetWorkloadsSlim(
	ctx context.Context,
	tenantId uint,
	permittedScopes *openapi.PermittedScopes,
	offset int32,
	limit int32,
	filterBy []queryapi.Filter,
) (*[]models.WorkloadSlim, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to list workloads", nil)

	tx := runaidb.DB().WithContext(ctx)

	workloads := &[]models.WorkloadSlim{}

	query := tx.
		Table(WorkloadsTableName).
		Select("id, name, type, project_name, cluster_id, distributed_framework").
		Where("tenant_id = ?", tenantId).
		Where("deleted_at IS NULL")

	if permittedScopes != nil {
		if permittedScopes.System == nil || !*permittedScopes.System {
			query = query.Where(getPermittedScopesFilter(*permittedScopes))
		}
	}

	query = query.WithEnricher(querydbenrich.Filters(filterBy))

	// a plain "id ASC" is served by the primary key index; FormatSortClause adds NULLS FIRST, which forces a full sort
	query = query.Order("id ASC")

	if offset != 0 {
		query = query.Offset(int(offset))
	}

	if limit != 0 {
		query = query.Limit(int(limit))
	}

	err := query.Find(workloads)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list slim workloads, err: %v", err)
		return &[]models.WorkloadSlim{}, dbErr.WithDetails(err)
	}

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
	deletionFilters []queryapi.Filter,
) ([]string, []string, []string, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to sync workloads", nil)

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)
	tx.SavePoint("sp")

	syncedWorkloads := []string{}
	failedWorkloads := []string{}
	notSuspendableWorkloads := []string{}
	for _, workload := range workloads {
		query := tx.Table(WorkloadsTableName)

		query = omitWorkloadsEmptyStaticFields(query, workload)

		_, saveErr := query.Save(&workload)
		if saveErr != nil {
			tx.RollbackTo("sp") // Rollback to last saved point
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to save workload %s in project %s, err %v", workload.Name, workload.ProjectName, saveErr)
			failedWorkloads = append(failedWorkloads, workload.ID.String())
		} else {
			tx.SavePoint("sp")
			syncedWorkloads = append(syncedWorkloads, workload.ID.String())
			if !workload.SupportSuspend {
				notSuspendableWorkloads = append(notSuspendableWorkloads, workload.ID.String())
			}
		}

		excludingIds = append(excludingIds, workload.ID.String())
	}

	if err := resetSupportSuspendOfV1Workloads(tx, tenantId, notSuspendableWorkloads); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to reset support_suspend of synced workloads: %v", err)
		return nil, nil, nil, dbErr.WithDetails(err)
	}

	if projectId == "" && clusterId == "" {
		return nil, nil, nil, api.NewError(http.StatusBadRequest, "Workloads sync without scope is not supported", nil)
	}

	// if one of projectId or clusterId is empty, we delete by the other one only
	deleteQuery := ConstructDeleteQueryForProjectCluster(tx, WorkloadsTableName, tenantId, clusterId, projectId, from, to, excludingIds, "")
	if len(deletionFilters) > 0 {
		deleteQuery = deleteQuery.WithEnricher(querydbenrich.Filters(deletionFilters))
	}

	deletedWorkloads := []models.Workload{}
	_, err := deleteQuery.
		Clauses(clause.Returning{Columns: []clause.Column{{Name: WorkloadsColumnId}}}).
		Delete(&deletedWorkloads)
	if err != nil {
		return nil, nil, nil, dbErr.WithDetails(err)
	}

	deletedWorkloadIds := make([]string, 0, len(deletedWorkloads))
	for _, workload := range deletedWorkloads {
		deletedWorkloadIds = append(deletedWorkloadIds, workload.ID.String())
	}

	err = tx.Commit()
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit the workloads sync: %v", err)
		return nil, nil, nil, dbErr.WithDetails(err)
	}
	committed = true

	return syncedWorkloads, failedWorkloads, deletedWorkloadIds, nil
}

// notSuspendableIds returns the workload's id when the cluster reports it as not suspendable, so single-row
// upserts can share resetSupportSuspendOfV1Workloads with the batch sync.
func notSuspendableIds(workload models.Workload) []string {
	if workload.SupportSuspend {
		return nil
	}
	return []string{workload.ID.String()}
}

// resetSupportSuspendOfV1Workloads lets the cluster's "not suspendable" report win for workloads that
// are not managed through the V2 API. omitWorkloadsEmptyStaticFields drops a false support_suspend from
// the save so a control-plane-set true on V2 workloads survives the sync, but for V1 workloads the
// cluster is the source of truth: a legacy InferenceWorkload is created with support_suspend=true and
// must be corrected once the cluster reports it cannot be suspended.
func resetSupportSuspendOfV1Workloads(tx orm.DB, tenantId uint, workloadIds []string) error {
	if len(workloadIds) == 0 {
		return nil
	}

	_, err := tx.Table(WorkloadsTableName).
		Where(fmt.Sprintf("%s = ?", WorkloadsColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s IN ?", WorkloadsColumnId), workloadIds).
		Where(fmt.Sprintf("%s = ?", WorkloadsColumnSupportSuspend), true).
		Where(fmt.Sprintf("%s IS NULL OR %s <> ?", WorkloadsColumnSourceApi, WorkloadsColumnSourceApi), string(api.SOURCEAPI_WORKLOADS_V2)).
		Update(WorkloadsColumnSupportSuspend, false)
	if err != nil {
		return fmt.Errorf("reset support_suspend for tenant %d: %w", tenantId, err)
	}

	return nil
}

func omitWorkloadsEmptyStaticFields(query orm.DB, workload models.Workload) orm.DB {
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

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	failedWorkloadsMap := createMapFromSlice(failedWorkloads)

	for _, child := range workloadsChildren {
		if _, ok := failedWorkloadsMap[child.WorkloadID.String()]; ok {
			continue
		}

		_, err := tx.Table(WorkloadsChildrenTableName).
			Clauses(clause.OnConflict{
				DoNothing: true,
			}).
			Create(&child)
		if err != nil {
			return dbErr.WithDetails(err)
		}
	}

	err := tx.Commit()
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit the workloads children: %v", err)
		return dbErr.WithDetails(err)
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
func applyRequestedNodepoolFilter(query orm.DB, nodepoolName string) orm.DB {
	return query.Where("array_length(requested_nodepools, 1) = 1").
		Where("? = ANY(requested_nodepools)", nodepoolName)
}

// applyCurrentNodepoolFilter filters workloads by current_nodepools array.
// Includes workloads that are currently running in the specified nodepool.
func applyCurrentNodepoolFilter(query orm.DB, nodepoolName string) orm.DB {
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
	filterBy []queryapi.Filter,
) orm.DB {
	tx := runaidb.DB().WithContext(ctx)
	query := tx.Table(w.view).
		Where("tenant_id = ?", tenantId).
		Where("deleted_at IS NULL").
		WithEnricher(querydbenrich.Filters(filterBy))

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
	filterBy []queryapi.Filter,
) (*[]telemetry_utils.TelemetryGroupValue, error) {
	telemetryGroupValues := []telemetry_utils.TelemetryGroupValue{}

	selectPhrase := fmt.Sprintf("%s as data", dataField)
	for i, groups := range groupBy {
		selectPhrase = fmt.Sprintf("%s, %s as group%d_main_value", selectPhrase, groups[0], i+1)
		if len(groups) > 1 {
			selectPhrase = fmt.Sprintf("%s, %s as group%d_secondary_value", selectPhrase, groups[1], i+1)
		}
	}

	query := w.buildTelemetryQuery(ctx, tenantId, clusterId, nodepoolName, departmentId, projectId, permittedScopes, groupBy, phases, filterBy).
		Select(selectPhrase)

	_, err := query.Scan(&telemetryGroupValues)
	if err != nil {
		return &telemetryGroupValues, err
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
	filterBy []queryapi.Filter,
) (*[]telemetry_utils.TelemetryGroupValue, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get workload count", nil)

	workloadTelemetry, err := w.dbGetWorkloadTelemetry(ctx, tenantId, clusterId, nodepoolName, departmentId, projectId, permittedScopes, groupBy, "COUNT(id)", phases, filterBy)
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
	filterBy []queryapi.Filter,
) (*[]telemetry_utils.TelemetryGroupValue, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get gpu allocation", nil)

	workloadTelemetry, err := w.dbGetWorkloadTelemetry(ctx, tenantId, clusterId, nodepoolName, departmentId, projectId, permittedScopes, groupBy, "COALESCE(SUM(allocated_gpu), 0)", phases, filterBy)
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
	filterBy []queryapi.Filter,
) (*[]telemetry_utils.TelemetryGroupValue, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get idle allocated gpus", nil)

	workloadTelemetry, err := w.dbGetWorkloadTelemetry(ctx, tenantId, clusterId, nodepoolName, departmentId, projectId, permittedScopes, groupBy, "COALESCE(SUM(idle_allocated_gpus), 0)", phases, filterBy)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get idle allocated gpus, err: %v", err)
		return workloadTelemetry, dbErr.WithDetails(err)
	}

	return workloadTelemetry, nil
}

func (w *WorkloadsDao) DbGetWorkloadPhase(ctx context.Context, tenantId uint, clusterId *string, workloadId string) (api.Phase, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get workload %s phase", workloadId), nil)

	tx := runaidb.DB().WithContext(ctx)

	workload := models.Workload{}

	query := tx.Model(workload).
		Unscoped().
		Where("tenant_id = ?", tenantId).
		Where("id = ?", workloadId)

	if clusterId != nil {
		query = query.Where("cluster_id = ?", *clusterId)
	}

	err := query.First(&workload)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get workload with id %s phase, err: %v", workloadId, err)
		return workload.Phase, dbErr.WithDetails(err)
	}

	return workload.Phase, nil
}

func (w *WorkloadsDao) DbGetChildren(ctx context.Context, workloadId uuid.UUID, childTypes *[]string) (*[]models.WorkloadChild, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get children of workload %s with childTypes %v", workloadId, childTypes), nil)
	tx := runaidb.DB().WithContext(ctx)

	children := &[]models.WorkloadChild{}
	query := tx.
		Table(WorkloadsChildrenTableName).
		Where("workload_id = ?", workloadId)

	if childTypes != nil {
		query = query.Where("child_type IN ?", *childTypes)
	}

	err := query.Find(children)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get children of workload with id %s with childTypes %v, err: %v",
			workloadId, childTypes, err)
		return &[]models.WorkloadChild{}, dbErr.WithDetails(err)
	}
	return children, nil
}

func (w *WorkloadsDao) DbGetWorkloadClusterId(ctx context.Context, tenantId uint, workloadId string) (uuid.UUID, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get clusterId of workload with id %s", workloadId), nil)
	tx := runaidb.DB().WithContext(ctx)

	workload := &models.Workload{}
	query := tx.Table(WorkloadsTableName).
		Select("cluster_id").
		Where("tenant_id = ?", tenantId).
		Where("id = ?", workloadId).
		Unscoped()

	err := query.First(workload)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get clusterId of workload with id %s, err: %v", workloadId, err)
		return uuid.UUID{}, dbErr.WithDetails(err)
	}

	return workload.ClusterID, nil
}

// DbGetWorkloadScope returns only the cluster, project, department and namespace of a live
// workload, reading the workloads table rather than workloads_view to skip its pod aggregates.
// A deleted workload is reported as not found.
func (w *WorkloadsDao) DbGetWorkloadScope(ctx context.Context, tenantId uint, workloadId uuid.UUID) (*models.Workload, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get workload %s", workloadId), nil)

	workload := &models.Workload{}
	err := runaidb.DB().WithContext(ctx).
		Table(WorkloadsTableName).
		Select("cluster_id, project_id, department_id, namespace").
		Where("tenant_id = ?", tenantId).
		Where("id = ?", workloadId).
		First(workload)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get workload %s, err: %v", workloadId, err)
		}
		return nil, dbErr.WithDetails(err)
	}

	return workload, nil
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

	tx := runaidb.DB().WithContext(ctx)

	workload := &models.WorkloadView{}
	err := tx.
		Table(WorkloadsViewName).
		Unscoped().
		Where("tenant_id = ?", tenantId).
		Where("id = ?", workloadId).
		First(workload)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get workload with id %s, err: %v", workloadId, err)
		}
		return nil, dbErr.WithDetails(err)
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

	tx := runaidb.DB().WithContext(ctx)

	workloads := []*models.WorkloadView{}
	query := tx.
		Table(WorkloadsViewName).
		Unscoped().
		Where("tenant_id = ?", tenantId).
		Where("id IN ?", workloadIds)

	if excludeDeleted != nil && *excludeDeleted {
		query = query.Where(fmt.Sprintf("%s IS NULL", WorkloadsDeletedAtColumnName))
	}

	_, err := query.Scan(&workloads)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get workload with id %v, err: %v", workloadIds, err)
		return nil, dbErr.WithDetails(err)
	}

	// Redact sensitive environment variables before returning
	redactWorkloadsEnvironmentVariablesPointers(ctx, workloads)

	return workloads, nil
}

// DeleteWorkloadsOfClusterId deletes all workloads within a cluster
func DeleteWorkloadsOfClusterId(ctx context.Context, tenantId string, clusterId string) error {
	// By default, GORM's Delete() method only performs a soft delete if the model has gorm.DeletedAt.
	// Using .Unscoped() ensures a hard delete, permanently removing the record from the database.
	_, err := runaidb.DB().WithContext(ctx).Unscoped().
		Table(WorkloadsTableName).
		Where(fmt.Sprintf("%s = ?", WorkloadsColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s = ?", WorkloadsColumnClusterId), clusterId).
		Delete(&models.Workload{})
	if err != nil {
		return fmt.Errorf("failed to delete workload for cluster %s in tenant %s. err: %w", clusterId, tenantId, err)
	}
	return nil
}

// DeleteWorkloadsOfTenantId deletes all workloads within a cluster
func DeleteWorkloadsOfTenantId(ctx context.Context, tenantId string) error {
	// By default, GORM's Delete() method only performs a soft delete if the model has gorm.DeletedAt.
	// Using .Unscoped() ensures a hard delete, permanently removing the record from the database.
	_, err := runaidb.DB().WithContext(ctx).Unscoped().
		Table(WorkloadsTableName).
		Where(fmt.Sprintf("%s = ?", WorkloadsColumnTenantId), tenantId).
		Delete(&models.Workload{})
	if err != nil {
		return fmt.Errorf("failed to delete workload in tenant %s. err: %w", tenantId, err)
	}
	return nil
}

// redactWorkloadEnvironmentVariables clears legacy plaintext access tokens for every caller and removes
// USERCREDENTIAL references when the current user is not the workload submitter.
func redactWorkloadEnvironmentVariables(ctx context.Context, workload *models.WorkloadView) {
	if workload == nil || workload.EnvironmentVariables == nil {
		return
	}

	redactLegacyPlaintextTokenValues(workload.EnvironmentVariables)

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

func redactLegacyPlaintextTokenValues(environmentVariables models.EnvironmentVariables) {
	for name, value := range environmentVariables {
		if !isLegacyPlaintextTokenEnvironmentVariable(name) {
			continue
		}
		if strings.HasPrefix(value, "USERCREDENTIAL:") {
			continue
		}
		environmentVariables[name] = ""
	}
}

func isLegacyPlaintextTokenEnvironmentVariable(name string) bool {
	return name == "HF_TOKEN" || name == "NGC_API_KEY"
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
