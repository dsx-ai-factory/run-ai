/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"k8s.io/utils/ptr"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
)

type OvertimeMetricRange string

const (
	RangeDay   OvertimeMetricRange = "day"
	RangeWeek  OvertimeMetricRange = "week"
	RangeMonth OvertimeMetricRange = "month"

	IdColumn                     = "id"
	NameColumn                   = "name"
	ParentNameColumn             = "parent_name"
	TotalGpuQuotaColumn          = "total_gpu_quota"
	AllocationValueColumn        = "allocation_value"
	UtilizationValueColumn       = "utilization_value"
	MemoryUtilizationValueColumn = "memory_utilization_value"
	GpuQuotaColumn               = "gpu_quota"
	EmptyOrClause                = "(false"

	AvgGpuAllocation24Column         = "avg_gpu_allocation24h"
	AvgGpuUtilization24Column        = "avg_gpu_utilization24h"
	AvgGpuMemoryUtilization24Column  = "avg_gpu_memory_utilization24h"
	AvgGpuAllocation7dColumn         = "avg_gpu_allocation7d"
	AvgGpuUtilization7dColumn        = "avg_gpu_utilization7d"
	AvgGpuMemoryUtilization7dColumn  = "avg_gpu_memory_utilization7d"
	AvgGpuAllocation30dColumn        = "avg_gpu_allocation30d"
	AvgGpuUtilization30dColumn       = "avg_gpu_utilization30d"
	AvgGpuMemoryUtilization30dColumn = "avg_gpu_memory_utilization30d"

	GpuAllocationRatioProjectExpression    = "CASE WHEN total_gpu_quota > 0 THEN gpu_allocated / total_gpu_quota ELSE 0 END"
	GpuAllocationRatioDepartmentExpression = "CASE WHEN total_gpu_quota > 0 THEN (SELECT COALESCE(SUM(pv.gpu_allocated), 0) " +
		"FROM project_view pv WHERE pv.parent_id = department_view.id::bigint AND pv.deleted_at IS NULL) / total_gpu_quota ELSE 0 END"
)

var metricTypeToOrderBy = map[api.RankingMetricType]string{
	api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H:         AvgGpuAllocation24Column,
	api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION24H:        AvgGpuUtilization24Column,
	api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION24H: AvgGpuMemoryUtilization24Column,
	api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION7D:          AvgGpuAllocation7dColumn,
	api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION7D:         AvgGpuUtilization7dColumn,
	api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION7D:  AvgGpuMemoryUtilization7dColumn,
	api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION30D:         AvgGpuAllocation30dColumn,
	api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION30D:        AvgGpuUtilization30dColumn,
	api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION30D: AvgGpuMemoryUtilization30dColumn,
	api.RANKINGMETRICTYPE_GPU_ALLOCATION_RATIO:          GpuAllocationRatioProjectExpression,
}

var validOrderBy = map[string]bool{
	AvgGpuAllocation24Column:               true,
	AvgGpuUtilization24Column:              true,
	AvgGpuMemoryUtilization24Column:        true,
	AvgGpuAllocation7dColumn:               true,
	AvgGpuUtilization7dColumn:              true,
	AvgGpuMemoryUtilization7dColumn:        true,
	AvgGpuAllocation30dColumn:              true,
	AvgGpuUtilization30dColumn:             true,
	AvgGpuMemoryUtilization30dColumn:       true,
	GpuAllocationRatioProjectExpression:    true,
	GpuAllocationRatioDepartmentExpression: true,
}

type RankingColumnNames struct {
	AllocationColumn        string
	UtilizationColumn       string
	MemoryUtilizationColumn string
}

// SetProjectsOvertimeDataForTenant
// tenantProjectsOvertimeData - map of clusterId -> projectName -> overtimeData
func SetProjectsOvertimeDataForTenant(ctx context.Context, tenantId string, tenantProjectsOvertimeData map[string]map[string]*models.OvertimeData,
	overtimeMetricRange OvertimeMetricRange,
) error {
	var err error = nil
	successfulClusters := []string{}
	for clusterId, projectsOvertimeData := range tenantProjectsOvertimeData {
		innerErr := setProjectsOvertimeDataForCluster(ctx, tenantId, clusterId, projectsOvertimeData, overtimeMetricRange)
		if innerErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to set projects overtime data for cluster %s in tenant %s with range %v, error: %s",
				clusterId, tenantId, overtimeMetricRange, innerErr.Error())
			err = multierror.Append(err, innerErr)
			continue
		}

		successfulClusters = append(successfulClusters, clusterId)
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully set projects overtime data for tenant %s with range %v, for clusters: %v",
		tenantId, overtimeMetricRange, successfulClusters)

	return err
}

// SetDepartmentsOvertimeDataForTenant
// tenantDepartmentsOvertimeData - map of clusterId -> departmentName -> overtimeData
func SetDepartmentsOvertimeDataForTenant(ctx context.Context, tenantId string, tenantDepartmentsOvertimeData map[string]map[string]*models.OvertimeData,
	overtimeMetricRange OvertimeMetricRange,
) error {
	var err error = nil
	successfulClusters := []string{}
	for clusterId, departmentsOvertimeData := range tenantDepartmentsOvertimeData {
		innerErr := setDepartmentsOvertimeDataForCluster(ctx, tenantId, clusterId, departmentsOvertimeData, overtimeMetricRange)
		if innerErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to set departments overtime data for cluster %s with range %v, error: %s",
				clusterId, overtimeMetricRange, innerErr.Error())
			err = multierror.Append(err, innerErr)
			continue
		}

		successfulClusters = append(successfulClusters, clusterId)
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully set departments overtime data for tenant %s with range %v, for clusters: %v",
		tenantId, overtimeMetricRange, successfulClusters)

	return err
}

func setProjectsOvertimeDataForCluster(ctx context.Context, tenantId string, clusterId string, projectsOvertimeData map[string]*models.OvertimeData,
	overtimeMetricRange OvertimeMetricRange,
) error {
	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to parse cluster id %s, error: %v", clusterId, err)
		return err
	}

	projects, err := GetProjectsByCluster(ctx, tenantId, clusterUuid, nil, models.SlimVerbosity)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get projects for setting the overtime data for cluster id %s, error: %s",
			clusterId, err.Error())
		return err
	}

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	for _, dbProject := range projects {
		_ = setOrgUnitOvertimeData(ctx, tx,
			dbProject.Name, dbProject.ID, dbProject.OrgUnitData,
			dbProject.OrgUnitData.OvertimeData.ID, "project",
			projectsOvertimeData, overtimeMetricRange)
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit transaction for setProjectsOvertimeDataForCluster: %v", err)
		return err
	}
	committed = true
	return nil
}

func setDepartmentsOvertimeDataForCluster(ctx context.Context, tenantId string, clusterId string, departmentsOvertimeData map[string]*models.OvertimeData,
	overtimeMetricRange OvertimeMetricRange,
) error {
	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to parse cluster id %s, error: %v", clusterId, err)
		return err
	}

	departments, err := GetDepartmentsByCluster(ctx, tenantId, clusterUuid, nil, models.SlimVerbosity)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get departments for setting the overtime data for cluster id %s, error: %s",
			clusterId, err.Error())
		return err
	}

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	for _, dbDepartment := range *departments {
		_ = setOrgUnitOvertimeData(ctx, tx,
			dbDepartment.Name, dbDepartment.ID, dbDepartment.OrgUnitData,
			dbDepartment.OrgUnitData.OvertimeData.ID, "department",
			departmentsOvertimeData, overtimeMetricRange)
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit transaction for setDepartmentsOvertimeDataForCluster: %v", err)
		return err
	}
	committed = true
	return nil
}

func setOrgUnitOvertimeData(
	ctx context.Context,
	tx orm.DB,
	orgUnitName string,
	orgUnitID int64,
	orgUnitData models.OrgUnitData,
	overtimeDataId uuid.UUID,
	orgUnitType string,
	orgUnitsOvertimeData map[string]*models.OvertimeData,
	overtimeMetricRange OvertimeMetricRange,
) error {
	overtimeData, found := orgUnitsOvertimeData[orgUnitName]
	if !found {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Missing overtime data for %s %s (id: %d)",
			orgUnitType, orgUnitName, orgUnitID)
		return nil
	}

	emptyId := uuid.UUID{}
	var err error
	if overtimeDataId == emptyId {
		// Create new overtime data
		setOvertimeDataUpdatedAtIfNeeded(overtimeData, overtimeMetricRange)
		overtimeData.OrgUnitDataID = orgUnitData.ID
		_, err = tx.Table(models.OvertimeDataTable).Create(overtimeData)
	} else {
		queryUpdates, updateErr := getOvertimeDataQueryUpdates(ctx, overtimeData, overtimeMetricRange, orgUnitData)
		if updateErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get overtime data query updates for %s %s (id: %d), error: %s",
				orgUnitType, orgUnitName, orgUnitID, updateErr.Error())
			return nil
		}
		if queryUpdates == nil {
			return nil
		}

		_, err = tx.Table(models.OvertimeDataTable).Where("org_unit_data_id = ?", orgUnitData.ID).Updates(queryUpdates)
	}

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update %s %s (id: %d) overtime data, error: %s",
			orgUnitType, orgUnitName, orgUnitID, err.Error())
		return err
	}

	return nil
}

func setOvertimeDataUpdatedAtIfNeeded(overtimeData *models.OvertimeData, overtimeMetricRange OvertimeMetricRange) {
	switch overtimeMetricRange {
	case RangeDay:
		overtimeData.Avg24hRangeUpdatedAt = ptr.To(time.Now())
	case RangeWeek:
		overtimeData.Avg7dRangeUpdatedAt = ptr.To(time.Now())
	case RangeMonth:
		overtimeData.Avg30dRangeUpdatedAt = ptr.To(time.Now())
	}
}

func getOvertimeDataQueryUpdates(ctx context.Context, overtimeData *models.OvertimeData, overtimeMetricRange OvertimeMetricRange, orgUnitData models.OrgUnitData) (map[string]interface{}, error) {
	if overtimeData == nil {
		return nil, errors.New("missing overtime data")
	}

	result := map[string]interface{}{}

	switch overtimeMetricRange {
	case RangeDay:
		hasAllocation := overtimeData.AvgGpuAllocation24h != nil
		hasUtilization := overtimeData.AvgGpuUtilization24h != nil
		hasMemoryUtilization := overtimeData.AvgGpuMemoryUtilization24h != nil

		if !hasAllocation && !hasUtilization && !hasMemoryUtilization {
			if orgUnitData.CreatedAt.After(time.Now().Add(-1 * 24 * time.Hour)) {
				return nil, errors.New("missing 24h metric data")
			} else {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf(
					"No 24h metrics available for org unit (created at %v) - metrics collection may be failing",
					orgUnitData.CreatedAt)
				return nil, nil
			}
		}

		if hasAllocation {
			result[AvgGpuAllocation24Column] = *overtimeData.AvgGpuAllocation24h
		}
		if hasUtilization {
			result[AvgGpuUtilization24Column] = *overtimeData.AvgGpuUtilization24h
		}
		if hasMemoryUtilization {
			result[AvgGpuMemoryUtilization24Column] = *overtimeData.AvgGpuMemoryUtilization24h
		}
		result["avg24h_range_updated_at"] = orm.Expr("NOW()")
	case RangeWeek:
		hasAllocation := overtimeData.AvgGpuAllocation7d != nil
		hasUtilization := overtimeData.AvgGpuUtilization7d != nil
		hasMemoryUtilization := overtimeData.AvgGpuMemoryUtilization7d != nil

		if !hasAllocation && !hasUtilization && !hasMemoryUtilization {
			if orgUnitData.CreatedAt.After(time.Now().Add(-7 * 24 * time.Hour)) {
				return nil, errors.New("missing 7d metric data")
			} else {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf(
					"No 7d metrics available for org unit (created at %v) - metrics collection may be failing",
					orgUnitData.CreatedAt)
				return nil, nil
			}
		}

		if hasAllocation {
			result[AvgGpuAllocation7dColumn] = *overtimeData.AvgGpuAllocation7d
		}
		if hasUtilization {
			result[AvgGpuUtilization7dColumn] = *overtimeData.AvgGpuUtilization7d
		}
		if hasMemoryUtilization {
			result[AvgGpuMemoryUtilization7dColumn] = *overtimeData.AvgGpuMemoryUtilization7d
		}
		result["avg7d_range_updated_at"] = orm.Expr("NOW()")
	case RangeMonth:
		hasAllocation := overtimeData.AvgGpuAllocation30d != nil
		hasUtilization := overtimeData.AvgGpuUtilization30d != nil
		hasMemoryUtilization := overtimeData.AvgGpuMemoryUtilization30d != nil

		if !hasAllocation && !hasUtilization && !hasMemoryUtilization {
			if orgUnitData.CreatedAt.After(time.Now().Add(-30 * 24 * time.Hour)) {
				return nil, errors.New("missing 30d metric data")
			} else {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf(
					"No 30d metrics available for org unit (created at %v) - metrics collection may be failing",
					orgUnitData.CreatedAt)
				return nil, nil
			}
		}

		if hasAllocation {
			result[AvgGpuAllocation30dColumn] = *overtimeData.AvgGpuAllocation30d
		}
		if hasUtilization {
			result[AvgGpuUtilization30dColumn] = *overtimeData.AvgGpuUtilization30d
		}
		if hasMemoryUtilization {
			result[AvgGpuMemoryUtilization30dColumn] = *overtimeData.AvgGpuMemoryUtilization30d
		}
		result["avg30d_range_updated_at"] = orm.Expr("NOW()")
	default:
		return nil, fmt.Errorf("invalid overtime metric range: %s", overtimeMetricRange)
	}

	return result, nil
}

func GetColumnNames(metricType api.RankingMetricType) (RankingColumnNames, error) {
	switch metricType {
	case api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H,
		api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION24H,
		api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION24H:
		return RankingColumnNames{
			AllocationColumn:        AvgGpuAllocation24Column,
			UtilizationColumn:       AvgGpuUtilization24Column,
			MemoryUtilizationColumn: AvgGpuMemoryUtilization24Column,
		}, nil
	case api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION7D,
		api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION7D,
		api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION7D:
		return RankingColumnNames{
			AllocationColumn:        AvgGpuAllocation7dColumn,
			UtilizationColumn:       AvgGpuUtilization7dColumn,
			MemoryUtilizationColumn: AvgGpuMemoryUtilization7dColumn,
		}, nil
	case api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION30D,
		api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION30D,
		api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION30D:
		return RankingColumnNames{
			AllocationColumn:        AvgGpuAllocation30dColumn,
			UtilizationColumn:       AvgGpuUtilization30dColumn,
			MemoryUtilizationColumn: AvgGpuMemoryUtilization30dColumn,
		}, nil
	case api.RANKINGMETRICTYPE_GPU_ALLOCATION_RATIO:
		// Handled specially in GetOrgUnitsRanking — returns project expression as default
		return RankingColumnNames{
			AllocationColumn:        GpuAllocationRatioProjectExpression,
			UtilizationColumn:       GpuAllocationRatioProjectExpression,
			MemoryUtilizationColumn: GpuAllocationRatioProjectExpression,
		}, nil
	default:
		return RankingColumnNames{}, fmt.Errorf("invalid metric type %s", metricType)
	}
}

func GetOrderBy(metricType api.RankingMetricType, isDepartments bool) (string, error) {
	// Department ratio needs a dedicated subquery (child project allocations / dept quota)
	// because department_view doesn't expose gpu_allocated directly.
	if metricType == api.RANKINGMETRICTYPE_GPU_ALLOCATION_RATIO && isDepartments {
		return GpuAllocationRatioDepartmentExpression, nil
	}
	orderBy, ok := metricTypeToOrderBy[metricType]
	if !ok {
		return "", fmt.Errorf("invalid metric type %s", metricType)
	}
	return orderBy, nil
}

func GetColumnNamesForRanking(metricType api.RankingMetricType, isDepartments bool) (RankingColumnNames, error) {
	if metricType == api.RANKINGMETRICTYPE_GPU_ALLOCATION_RATIO && isDepartments {
		return RankingColumnNames{
			AllocationColumn:        GpuAllocationRatioDepartmentExpression,
			UtilizationColumn:       GpuAllocationRatioDepartmentExpression,
			MemoryUtilizationColumn: GpuAllocationRatioDepartmentExpression,
		}, nil
	}
	return GetColumnNames(metricType)
}

func GetOrgUnitsRanking(
	ctx context.Context,
	tenantId string,
	clusterId string,
	permittedScopes *openapi.PermittedScopes,
	columnNames RankingColumnNames,
	orderBy string,
	orderType api.RankingOrderType,
	limit int32,
	isDepartments bool,
	departmentId string,
) ([]api.ProjectGetRanking, error) {
	if columnNames.AllocationColumn == "" ||
		columnNames.UtilizationColumn == "" ||
		columnNames.MemoryUtilizationColumn == "" ||
		orderBy == "" {
		return nil, fmt.Errorf("invalid column names provided")
	}

	valid, present := validOrderBy[orderBy]
	if !valid || !present {
		return nil, fmt.Errorf("invalid orderBy %s", orderBy)
	}
	order := "DESC"
	if orderType == api.RANKINGORDERTYPE_BOTTOM {
		order = "ASC"
	}

	var viewName string
	var selectFields string

	if isDepartments {
		viewName = models.DepartmentViewName
		selectFields = fmt.Sprintf("%s, %s, %s as %s, %s as %s, %s as %s, %s as %s",
			IdColumn, NameColumn,
			columnNames.AllocationColumn, AllocationValueColumn,
			columnNames.UtilizationColumn, UtilizationValueColumn,
			columnNames.MemoryUtilizationColumn, MemoryUtilizationValueColumn,
			TotalGpuQuotaColumn, GpuQuotaColumn)
	} else {
		viewName = models.ProjectViewName
		// The only difference from department select fields is the addition of parent_name for projects
		selectFields = fmt.Sprintf("%s, %s, %s, %s as %s, %s as %s, %s as %s, %s as %s",
			IdColumn, NameColumn, ParentNameColumn,
			columnNames.AllocationColumn, AllocationValueColumn,
			columnNames.UtilizationColumn, UtilizationValueColumn,
			columnNames.MemoryUtilizationColumn, MemoryUtilizationValueColumn,
			TotalGpuQuotaColumn, GpuQuotaColumn)
	}

	query := runaidb.DB().WithContext(ctx).Table(viewName).
		Select(selectFields).
		Where("tenant_id = ?", tenantId).
		Where("cluster_id = ?", clusterId).
		Where("deleted_at IS NULL")

	if departmentId != "" && !isDepartments {
		query = query.Where("parent_id = ?", departmentId)
	}

	if permittedScopes != nil {
		tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
		if !tenantPermitted {
			if isDepartments {
				query = query.Where(getPermittedScopesForDepartments(*permittedScopes))
			} else {
				query = query.Where(getPermittedScopesForProjects(*permittedScopes))
			}
		}
	}

	// Order by the requested column with NULLS LAST to ensure items with data appear first,
	// then by name for consistent ordering when values are equal or NULL
	query.Order(fmt.Sprintf("%s %s NULLS LAST, %s ASC", orderBy, order, NameColumn)).
		Limit(int(limit))

	var results []api.ProjectGetRanking
	if _, err := query.Scan(&results); err != nil {
		return nil, err
	}

	return results, nil
}

// GetAllocationRatioRankingByNodepool queries the resources table directly for per-nodepool allocation ratio.
// For projects: ratio = allocated_gpu / gpu_deserved per nodepool
// For departments: ratio = SUM(child projects' allocated_gpu) / department's own gpu_deserved per nodepool
func GetAllocationRatioRankingByNodepool(
	ctx context.Context,
	tenantId string,
	clusterId string,
	nodePoolName string,
	permittedScopes *openapi.PermittedScopes,
	orderType api.RankingOrderType,
	limit int32,
	isDepartments bool,
	departmentId string,
) ([]api.ProjectGetRanking, error) {
	order := "DESC"
	if orderType == api.RANKINGORDERTYPE_BOTTOM {
		order = "ASC"
	}

	deptScopeFilter, projScopeFilter := buildNodepoolRankingScopeFilters(tenantId, permittedScopes)

	var query orm.DB

	// allocationRatioResult is a local scan target so the SQL can use the descriptive
	// alias "allocation_ratio" rather than reusing the generic "allocation_value" name.
	type allocationRatioResult struct {
		Name            string   `gorm:"column:name"`
		Id              string   `gorm:"column:id"`
		ParentName      string   `gorm:"column:parent_name"`
		AllocationRatio *float64 `gorm:"column:allocation_ratio"`
		GpuQuota        *float64 `gorm:"column:gpu_quota"`
	}

	if isDepartments {
		// LATERAL subquery aggregates child project allocations per department in one pass,
		// avoiding row multiplication from LEFT JOIN chains + GROUP BY.
		rawSQL := fmt.Sprintf(`SELECT d.name, d.id::text as id,
			CASE WHEN dept_r.gpu_deserved > 0 THEN child_alloc.total_allocated / dept_r.gpu_deserved ELSE 0 END as allocation_ratio,
			dept_r.gpu_deserved as gpu_quota
			FROM department d
			JOIN org_unit_data dept_oud ON d.org_unit_data_id = dept_oud.id
			JOIN resources dept_r ON dept_r.org_unit_data_id = dept_oud.id AND dept_r.deleted_at IS NULL AND dept_r.node_pool_name = ?
			LEFT JOIN LATERAL (
				SELECT COALESCE(SUM(r.allocated_gpu), 0) as total_allocated
				FROM project p
				JOIN org_unit_data oud ON p.org_unit_data_id = oud.id
				JOIN resources r ON r.org_unit_data_id = oud.id AND r.node_pool_name = ? AND r.deleted_at IS NULL
				WHERE p.parent_id = d.id AND p.deleted_at IS NULL
			) child_alloc ON true
			WHERE d.tenant_id = ? AND d.cluster_id = ? AND d.deleted_at IS NULL
			%s
			ORDER BY allocation_ratio %s NULLS LAST, name ASC
			LIMIT %d`, deptScopeFilter, order, limit)

		query = runaidb.DB().WithContext(ctx).Raw(rawSQL, nodePoolName, nodePoolName, tenantId, clusterId)
	} else {
		deptFilter := ""
		args := []interface{}{tenantId, clusterId, nodePoolName}
		if departmentId != "" {
			deptFilter = "AND p.parent_id = ?"
			args = append(args, departmentId)
		}

		rawSQL := fmt.Sprintf(`SELECT p.name, p.id::text as id, p.parent_name,
			CASE WHEN r.gpu_deserved > 0 THEN r.allocated_gpu / r.gpu_deserved ELSE 0 END as allocation_ratio,
			r.gpu_deserved as gpu_quota
			FROM project p
			JOIN org_unit_data oud ON p.org_unit_data_id = oud.id
			JOIN resources r ON r.org_unit_data_id = oud.id AND r.deleted_at IS NULL
			WHERE p.tenant_id = ? AND p.cluster_id = ? AND p.deleted_at IS NULL
			AND r.node_pool_name = ? %s %s
			ORDER BY allocation_ratio %s NULLS LAST, name ASC
			LIMIT %d`, deptFilter, projScopeFilter, order, limit)

		query = runaidb.DB().WithContext(ctx).Raw(rawSQL, args...)
	}

	var rows []allocationRatioResult
	if _, err := query.Scan(&rows); err != nil {
		return nil, err
	}

	results := make([]api.ProjectGetRanking, 0, len(rows))
	for _, r := range rows {
		results = append(results, api.ProjectGetRanking{
			Name:            r.Name,
			Id:              r.Id,
			ParentName:      r.ParentName,
			AllocationValue: r.AllocationRatio,
			GpuQuota:        r.GpuQuota,
		})
	}
	return results, nil
}

// buildNodepoolRankingScopeFilters returns SQL WHERE clause fragments (prefixed with "AND") that
// restrict department and project raw SQL queries to the caller's permitted scopes (RBAC).
// The ranking query itself is always single-cluster (filtered by WHERE cluster_id = ?);
// the Clusters scope check here is for authorization, not multi-cluster querying.
// Returns empty strings when the tenant is fully permitted (no row-level restriction needed).
func buildNodepoolRankingScopeFilters(tenantId string, permittedScopes *openapi.PermittedScopes) (deptFilter string, projFilter string) {
	if permittedScopes == nil || IsTenantPermitted(tenantId, permittedScopes) {
		return "", ""
	}

	dClause := EmptyOrClause
	if len(permittedScopes.Clusters) > 0 {
		dClause += " OR d.cluster_id IN " + ListItemsForQuery(permittedScopes.Clusters)
	}
	if len(permittedScopes.Departments) > 0 {
		dClause += " OR d.id IN " + ListItemsForQuery(permittedScopes.Departments)
	}
	dClause += ")"

	pClause := EmptyOrClause
	if len(permittedScopes.Clusters) > 0 {
		pClause += " OR p.cluster_id IN " + ListItemsForQuery(permittedScopes.Clusters)
	}
	if len(permittedScopes.Departments) > 0 {
		pClause += " OR p.parent_id IN " + ListItemsForQuery(permittedScopes.Departments)
	}
	if len(permittedScopes.Projects) > 0 {
		pClause += " OR p.id IN " + ListItemsForQuery(permittedScopes.Projects)
	}
	pClause += ")"

	return "AND " + dClause, "AND " + pClause
}
