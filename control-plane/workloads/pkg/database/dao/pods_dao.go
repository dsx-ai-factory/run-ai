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
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils/validators"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"

	ext_api "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/api"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	PodsTableName           = "pods"
	ContainersTableName     = "containers"
	podElementsJoinClause   = "LEFT JOIN pod_elements ON pod_elements.pod_id = pods.id"
	podElementsColumnPrefix = "pod_elements."
	completedParamTrue      = "true"
	completedParamFalse     = "false"
)

func applyPodsCompletedFilter(query *gorm.DB, completed string) *gorm.DB {
	switch strings.ToLower(completed) {
	case completedParamTrue:
		return query.Where("pods.completed_at IS NOT NULL")
	case completedParamFalse:
		return query.Where("pods.completed_at IS NULL")
	default:
		return query
	}
}

func DbCountPods(ctx context.Context, tenantId uint, deleted bool, filterBy []validators.ParameterizedFilter, search string, completed string, permittedScopes *openapi.PermittedScopes) (int64, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to count pods", nil)

	tx := dao_utils.DB().WithContext(ctx)

	isTenantPermitted := true
	if permittedScopes != nil {
		isSystemPermitted := permittedScopes.System != nil && *permittedScopes.System
		isTenantPermitted = isSystemPermitted || slices.Contains(permittedScopes.Tenants, strconv.Itoa(int(tenantId)))
	}
	hasDepartmentPermittedScopes := !isTenantPermitted && len(permittedScopes.Departments) > 0

	var query *gorm.DB
	if hasDepartmentPermittedScopes {
		query = tx.Table(PodsTableName).
			Joins("LEFT JOIN workloads as w on w.tenant_id = pods.tenant_id AND w.cluster_id = pods.cluster_id  AND w.id = pods.workload_id").
			Where("pods.tenant_id = ?", tenantId)
	} else {
		query = tx.Table(PodsTableName).
			Where("tenant_id = ?", tenantId)
	}

	needsElementJoin := hasElementIdFilter(filterBy)
	if needsElementJoin {
		query = query.Joins(podElementsJoinClause)
	}

	if !isTenantPermitted {
		query = query.Where(getPermittedScopesPodsFilter(*permittedScopes, hasDepartmentPermittedScopes))
	}

	if search != "" {
		query = dao_utils.ApplyTSSearchParam(query, search, "search_vector")
	}

	if deleted {
		query = query.Unscoped().Where("pods.deleted_at IS NOT NULL")
	} else {
		query = query.Where("pods.deleted_at IS NULL")
	}

	query = applyPodsCompletedFilter(query, completed)

	query = dao_utils.ApplyParameterizedFilters(query, filterBy)

	if needsElementJoin {
		query = query.Distinct("pods.id")
	}

	var count int64
	result := query.
		Count(&count)

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count pods, err: %v", result.Error)
		return 0, dbErr.WithDetails(result.Error)
	}

	return count, nil
}

func getPermittedScopesPodsFilter(permittedScopes openapi.PermittedScopes, addDepartment bool) string {
	if addDepartment {
		return authorization_utils.GetPermittedScopesDbFilterWithColumnNames(
			permittedScopes,
			"pods.tenant_id",
			"pods.cluster_id",
			"w.department_id",
			"pods.project_id",
		)
	}
	orClause := "(false"
	if len(permittedScopes.Clusters) > 0 {
		orClause = orClause + " OR " + "pods.cluster_id" + " IN " + authorization_utils.ListItemsForQuery(permittedScopes.Clusters) //nolint:staticcheck // ListItemsForQuery is deprecated; migration tracked separately
	}
	if len(permittedScopes.Projects) > 0 {
		orClause = orClause + " OR " + "pods.project_id" + " IN " + authorization_utils.ListItemsForQuery(permittedScopes.Projects) //nolint:staticcheck // ListItemsForQuery is deprecated; migration tracked separately
	}

	orClause += ")"

	return orClause
}

func DbListPods(ctx context.Context, tenantId uint, deleted bool, offset int32, limit int32, sortOrder string, sortBy string,
	filterBy []validators.ParameterizedFilter, verbosity ext_api.Verbosity, completed string, search string,
	permittedScopes *openapi.PermittedScopes,
) (interface{}, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to list pods", nil)

	tx := dao_utils.DB().WithContext(ctx)

	isTenantPermitted := true
	if permittedScopes != nil {
		isSystemPermitted := permittedScopes.System != nil && *permittedScopes.System
		isTenantPermitted = isSystemPermitted || slices.Contains(permittedScopes.Tenants, strconv.Itoa(int(tenantId)))
	}
	hasDepartmentPermittedScopes := !isTenantPermitted && len(permittedScopes.Departments) > 0

	var query *gorm.DB
	if verbosity == ext_api.VerbosityFull || hasDepartmentPermittedScopes {
		query = tx.Table(PodsTableName).
			Joins("LEFT JOIN workloads as w on w.tenant_id = pods.tenant_id AND w.cluster_id = pods.cluster_id  AND w.id = pods.workload_id")
	} else {
		query = tx.Table(PodsTableName)
	}
	needsElementJoin := hasElementIdFilter(filterBy)
	if needsElementJoin {
		query = query.Joins(podElementsJoinClause)
	}
	query = query.Where("pods.tenant_id = ?", tenantId)
	if !isTenantPermitted {
		query = query.Where(getPermittedScopesPodsFilter(*permittedScopes, hasDepartmentPermittedScopes))
	}

	if deleted {
		query = query.Unscoped().Where("deleted_at IS NOT NULL")
	}

	query = applyPodsCompletedFilter(query, completed)

	query = query.Preload("Containers").Preload("Elements")

	var err error
	query = dao_utils.ApplyParameterizedFilters(query, filterBy)

	if sortBy != "" {
		query, err = dao_utils.ApplySortByParam(query, sortBy, sortOrder)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sort pods, err: %v", err)
			return nil, dbErr.WithDetails(err)
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

	if needsElementJoin {
		query = query.Distinct("pods.*")
	}

	if verbosity == ext_api.VerbosityFull {
		var podsFull []models.PodFull
		query = query.Select("pods.*, w.name as workload_name, w.project_name as project_name")

		result := query.Find(&podsFull)
		if result.Error != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list pods with verbosity %v: %v", verbosity, result.Error)
			return nil, dbErr.WithDetails(result.Error)
		}
		return &podsFull, nil
	} else {
		var pods []models.Pod
		result := query.Find(&pods)
		if result.Error != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list pods with verbosity %v: %v", verbosity, result.Error)
			return nil, dbErr.WithDetails(result.Error)
		}
		return &pods, nil
	}
}

func DbGetRevisionPods(ctx context.Context, tenantId uint, revisionId string, deleted bool, offset int32, limit int32, search string) (*[]models.Pod, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to list revision pods", nil)

	tx := dao_utils.DB().WithContext(ctx)

	pods := &[]models.Pod{}

	query := tx.
		Table(PodsTableName).
		Where("tenant_id = ?", tenantId).
		Where("revision_id = ?", revisionId)

	if deleted {
		query = query.Unscoped().Where("deleted_at IS NOT NULL")
	}

	query = query.Preload("Containers").Preload("Elements")

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
		Find(pods)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list revision pods: %v", result.Error)
		return &[]models.Pod{}, dbErr.WithDetails(result.Error)
	}

	return pods, nil
}

// DbGetPod returns a single pod by id within the given workload. Caller-level
// authorization is enforced upstream by the workloadId-scoped authz middleware,
// so this DAO only carries the tenant + workload + pod scoping needed to
// resolve the row.
func DbGetPod(ctx context.Context, tenantId uint, workloadId uuid.UUID, podId uuid.UUID) (*models.Pod, *api.Error) {
	var pod models.Pod
	result := dao_utils.DB().WithContext(ctx).
		Table(PodsTableName).
		Where("pods.tenant_id = ? AND pods.workload_id = ? AND pods.id = ?", tenantId, workloadId, podId).
		Preload("Containers").
		Preload("Elements").
		First(&pod)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get pod %s, err: %v", podId, result.Error)
		dbErr := api.NewError(http.StatusInternalServerError, "Failed to get pod", nil) // 500
		return nil, dbErr.WithDetails(result.Error)
	}
	return &pod, nil
}

func DbDeletePod(ctx context.Context, tenantId uint, podId uuid.UUID) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to delete pod %s", podId), nil)

	result := dao_utils.DB().WithContext(ctx).
		Table(PodsTableName).
		Where("tenant_id = ?", tenantId).
		Delete(&models.Pod{}, podId)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to delete pod %s: not found", podId)
			return dbErr.WithDetails(result.Error)
		}

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(fmt.Sprintf("Failed to delete pod %s", podId), result.Error)
		return dbErr.WithDetails(result.Error)
	}

	return nil
}

func DbUpdatePodsByNodeName(ctx context.Context, tenantId uint, clusterId uuid.UUID, nodeName string, updateNodePodsRequest api.UpdateNodePodsRequest) (int64, *api.Error) {
	tx := dao_utils.DB().WithContext(ctx)

	updates := map[string]any{}
	if updateNodePodsRequest.NetworkTopologyLabelValues != nil {
		labels := pq.StringArray(*updateNodePodsRequest.NetworkTopologyLabelValues)
		updates["network_topology_label_values"] = labels
	} else {
		updates["network_topology_label_values"] = nil
	}
	// Update all pods matching clusterId and nodeName
	query := tx.Table(PodsTableName).
		Where("tenant_id = ?", tenantId).
		Where("cluster_id = ?", clusterId.String()).
		Where("node_name = ?", nodeName).
		Where("phase = 'Running'").
		Where("deleted_at IS NULL").
		Updates(updates)

	if query.Error != nil {
		return 0, api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to update pods by node name %s", nodeName), query.Error)
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Updated %d pods for node <%s> in cluster <%s>", query.RowsAffected, nodeName, clusterId.String())
	return query.RowsAffected, nil
}

func DbUpsertPod(ctx context.Context, pod models.Pod) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to upsert pod %s", pod.Name), nil)

	tx := dao_utils.DB().WithContext(ctx)

	result := tx.Table(PodsTableName).Save(&pod)
	if result.Error != nil {
		var pgErr *pgconn.PgError
		errIsPg := errors.As(result.Error, &pgErr)
		if (errIsPg && pgErr.Code == runaihttp.ErrForeignKey) || errors.Is(result.Error, gorm.ErrForeignKeyViolated) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to upsert pod %s: FK violated. workloadId: %s, clusterId: %s", pod.ID, pod.WorkloadID, pod.ClusterID)
			return dbErr.WithDetails(result.Error)
		}

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to upsert pod: %v", result.Error)
		return dbErr.WithDetails(result.Error)
	}

	return nil
}

func DbSyncPods(ctx context.Context, tenantId uint, projectId string, clusterId string, pods []models.Pod, to *uuid.UUID, from *uuid.UUID, excludingIds []string, workloadId string) ([]string, []string, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to sync pods", nil)

	sort.Slice(pods, func(i, j int) bool {
		return pods[i].ID.String() < pods[j].ID.String()
	})

	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to begin a session with the db: %v", tx.Error)
		return nil, nil, dbErr.WithDetails(tx.Error)
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)
	tx.SavePoint("sp")

	syncedPods := []string{}
	failedPods := []string{}
	for _, pod := range pods {
		result := tx.Table(PodsTableName).Save(&pod)

		if result.Error != nil {
			tx.RollbackTo("sp") // Rollback to last saved point

			var pgErr *pgconn.PgError
			errIsPg := errors.As(result.Error, &pgErr)
			if (errIsPg && pgErr.Code == runaihttp.ErrForeignKey) || errors.Is(result.Error, gorm.ErrForeignKeyViolated) {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to delete pod %s: FK violated. workloadId: %s, clusterId: %s", pod.ID, pod.WorkloadID, pod.ClusterID)
			} else {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync pod %s. error:  %v", pod.ID, result.Error.Error())
			}

			failedPods = append(failedPods, pod.ID.String())
		} else {
			tx.SavePoint("sp")
			syncedPods = append(syncedPods, pod.ID.String())
		}

		excludingIds = append(excludingIds, pod.ID.String()) // If pod sync failed, do not delete it — it may still be active.
	}

	if projectId == "" && clusterId == "" {
		return nil, nil, api.NewError(http.StatusBadRequest, "Pods sync without scope is not supported", nil)
	}

	// if one of projectId, clusterId or workloadId is empty, we delete by the other one only
	deleteQuery := ConstructDeleteQueryForProjectCluster(tx, PodsTableName, tenantId, clusterId, projectId, from, to, excludingIds, workloadId)

	result := deleteQuery.
		Delete(&models.Pod{})

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete unexisting pods. error:  %v", result.Error.Error())
		return nil, nil, dbErr.WithDetails(result.Error)
	}

	result = tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit the pods sync: %v", result.Error)
		return nil, nil, dbErr.WithDetails(result.Error)
	}
	committed = true

	return syncedPods, failedPods, nil
}

func DbUpsertContainers(ctx context.Context, containers []models.Container, failedPods []string) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to upsert containers", nil)

	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to begin a session with the db: %v", tx.Error)
		return dbErr.WithDetails(tx.Error)
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	failedPodsMap := createMapFromSlice(failedPods)

	for _, container := range containers {
		if _, ok := failedPodsMap[container.PodID.String()]; ok {
			continue
		}

		result := tx.Table(ContainersTableName).
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "cluster_id"}, {Name: "pod_id"}, {Name: "name"}},
				DoUpdates: clause.AssignmentColumns([]string{"started_at", "image"}),
			}).
			Create(&container)
		if result.Error != nil {
			return dbErr.WithDetails(result.Error)
		}
	}

	result := tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit upsert containers: %v", result.Error)
		return dbErr.WithDetails(result.Error)
	}
	committed = true

	return nil
}

func DbUpsertPodElements(ctx context.Context, podElements []models.PodElement, failedPods []string) *api.Error {
	if len(podElements) == 0 {
		return nil
	}

	dbErr := api.NewError(http.StatusInternalServerError, "Failed to upsert pod elements", nil)

	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to begin a session with the db: %v", tx.Error)
		return dbErr.WithDetails(tx.Error)
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	failedPodsMap := createMapFromSlice(failedPods)

	for _, elem := range podElements {
		if _, ok := failedPodsMap[elem.PodID.String()]; ok {
			continue
		}

		result := tx.Table(models.PodElementsTableName).
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "pod_id"}, {Name: "level"}},
				DoNothing: true,
			}).
			Create(&elem)
		if result.Error != nil {
			return dbErr.WithDetails(result.Error)
		}
	}

	result := tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit upsert pod elements: %v", result.Error)
		return dbErr.WithDetails(result.Error)
	}
	committed = true

	return nil
}

// DbGetWorkloadTopologyLca - topology lca - (lowest common ancestor) across all running pods of a workload.
func DbGetWorkloadTopologyLca(ctx context.Context, tenantId uint, workloadId uuid.UUID) (pq.StringArray, *api.Error) {
	return getRunningPodsTopologyLca(ctx, "", []any{}, tenantId, workloadId)
}

// DbGetElementTopologyLca - topology lca across the running pods of a specific workload element.
func DbGetElementTopologyLca(ctx context.Context, tenantId uint, workloadId uuid.UUID, elementId string) (pq.StringArray, *api.Error) {
	extraWhere := `AND id IN (SELECT pod_id FROM pod_elements WHERE element_id = ?)`
	return getRunningPodsTopologyLca(ctx, extraWhere, []any{elementId}, tenantId, workloadId)
}

// getRunningPodsTopologyLca computes the lowest common ancestor of network topology paths
// across the running pods of a workload, optionally further scoped by extraWhere/extraArgs
// (used to scope to a single element). Only running, non-deleted pods are considered.
func getRunningPodsTopologyLca(ctx context.Context, extraWhere string, extraArgs []any, tenantId uint, workloadId uuid.UUID) (pq.StringArray, *api.Error) {
	sql := fmt.Sprintf(`
		WITH running_pods AS (
			SELECT network_topology_label_values
			FROM pods
			WHERE tenant_id = ?
			AND workload_id = ?
			AND deleted_at IS NULL
			AND phase = 'Running'
			%s
		),
		bounds AS (
			SELECT
				MIN(network_topology_label_values) as min_path,
				MAX(network_topology_label_values) as max_path
			FROM running_pods
		),
		mismatch AS (
			SELECT i
			FROM bounds, generate_series(1, array_length(min_path, 1)) as i
			WHERE min_path[i] IS DISTINCT FROM max_path[i]
			ORDER BY i ASC
			LIMIT 1
		)
		SELECT
			CASE
				WHEN EXISTS(
					SELECT 1 FROM running_pods
					WHERE network_topology_label_values IS NULL
					OR array_length(network_topology_label_values, 1) IS NULL
				) THEN NULL
				WHEN (SELECT i FROM mismatch) IS NULL THEN (SELECT min_path FROM bounds)
				ELSE NULLIF((SELECT min_path[1 : i-1] FROM bounds, mismatch), '{}'::text[])
			END as lca_topology
		FROM bounds;
	`, extraWhere)

	args := append([]any{tenantId, workloadId.String()}, extraArgs...)
	var result pq.StringArray
	row := dao_utils.DB().WithContext(ctx).Raw(sql, args...).Row()
	if err := row.Scan(&result); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get topology LCA for workload %s: %v", workloadId, err)
		return nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get topology LCA for workload %s", workloadId), err)
	}

	return result, nil
}

func DbGetPodClusterId(ctx context.Context, tenantId uint, workloadId string, podId string) (uuid.UUID, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get clusterId of pod with id %s", podId), nil)
	tx := dao_utils.DB().WithContext(ctx)

	pod := &models.Pod{}
	query := tx.Table(PodsTableName).
		Select("cluster_id").
		Where("tenant_id = ?", tenantId).
		Where("workload_id = ?", workloadId).
		Where("id = ?", podId)

	result := query.First(pod)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get clusterId of pod with id %s, err: %v", podId, result.Error)
		return uuid.UUID{}, dbErr.WithDetails(result.Error)
	}

	return pod.ClusterID, nil
}

func hasElementIdFilter(filters []validators.ParameterizedFilter) bool {
	for _, f := range filters {
		if strings.HasPrefix(f.Field, podElementsColumnPrefix) {
			return true
		}
	}
	return false
}
