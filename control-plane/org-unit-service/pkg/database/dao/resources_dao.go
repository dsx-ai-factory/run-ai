package dao

import (
	"context"
	"fmt"
	"strings"

	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"gorm.io/gorm"
)

const (
	updateResourcesQuery = `UPDATE resources SET %s 
                 FROM %s WHERE %s.org_unit_data_id = resources.org_unit_data_id 
                             AND %s AND %s.tenant_id = ? AND %s.cluster_id = ? 
                             AND %s.deleted_at is NULL and resources.deleted_at is NULL;`

	disableCpuQuotaQueryCondition = `(cpu_deserved IS NOT NULL OR cpu_limit IS NOT NULL OR cpu_over_quota_weight IS NOT NULL 
                     OR memory_deserved IS NOT NULL OR memory_limit IS NOT NULL OR memory_over_quota_weight IS NOT NULL)`

	disableCpuQuotaQuerySet = `cpu_deserved = NULL, cpu_limit = NULL, cpu_over_quota_weight = NULL, 
                     memory_deserved = NULL, memory_limit = NULL, memory_over_quota_weight = NULL, updated_at=now()`

	hasQuotaQueryCondition = `(gpu_deserved <> 0 OR gpu_limit <> -1 OR 
					 cpu_deserved <> -1 OR cpu_limit <> -1 OR 
					 memory_deserved <> -1 OR memory_limit <> -1)`
)

var nonDefaultRankQueryCondition = fmt.Sprintf(`(resources.priority <> '%s')`, models.DefaultResourceRank)

type ResetResourcesAction string

const (
	QuotasToDefault   = ResetResourcesAction(api.RESOURCESRESETACTION_QUOTAS_TO_DEFAULT)
	RanksToDefault    = ResetResourcesAction(api.RESOURCESRESETACTION_RANKS_TO_DEFAULT)
	WeightsToDefault  = ResetResourcesAction(api.RESOURCESRESETACTION_WEIGHTS_TO_DEFAULT)
	WeightsToDeserved = ResetResourcesAction(api.RESOURCESRESETACTION_WEIGHTS_TO_DESERVED)

	CpuQuotasToDefault = ResetResourcesAction(api.NODEPOOLRESOURCESRESETACTION_CPU_QUOTAS_TO_DEFAULT)
)

func getUpdateResourcesQuery(setQuery, conditionQuery string, projects bool) string {
	table := models.ProjectTable
	if !projects {
		table = models.DepartmentTable
	}

	return fmt.Sprintf(updateResourcesQuery, setQuery, table, table, conditionQuery, table, table, table)
}

func getDisableCpuQuotaResourcesQuery(projects bool) string {
	return getUpdateResourcesQuery(disableCpuQuotaQuerySet, disableCpuQuotaQueryCondition, projects)
}

func DeleteResources(ctx context.Context, resources []models.Resources) error {
	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to db session start: %v", tx.Error)
		return tx.Error
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)
	result := tx.Delete(resources)

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete resources: %v", result.Error)
		return result.Error
	}
	tx.SavePoint("sp")

	result = tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create resources: %v", result.Error)
		return result.Error
	}
	committed = true

	return nil
}

func AddResources(ctx context.Context, resources []*models.Resources) error {
	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to db session start: %v", tx.Error)
		return tx.Error
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	result := tx.Create(resources)

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create resources: %v", result.Error)
		return result.Error
	}
	tx.SavePoint("sp")

	result = tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create resources: %v", result.Error)
		return result.Error
	}
	committed = true

	return nil
}

func DisableCpuQuota(ctx context.Context, tenantId string, clusters []openapi.ClustersMinimalData) error {
	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to db session start: %v", tx.Error)
		return tx.Error
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	var result *gorm.DB
	for _, cluster := range clusters {
		result = tx.Exec(getDisableCpuQuotaResourcesQuery(true), tenantId, cluster.Uuid)
		if result.Error != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update cpu resources for projects in tenant %s and cluster %s. err: %+v", tenantId, cluster.Uuid, tx.Error)
			return result.Error
		}
		result = tx.Exec(getDisableCpuQuotaResourcesQuery(false), tenantId, cluster.Uuid)
		if result.Error != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update cpu resources for departments in tenant %s and cluster %s. err: %+v", tenantId, cluster.Uuid, tx.Error)
			return result.Error
		}
	}

	result = tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed commit update cpu resources: %v", result.Error)
		return result.Error
	}
	committed = true

	for _, cluster := range clusters {
		InvalidateAllCache(ctx, tenantId, cluster.Uuid)
	}
	return nil
}

func getOrgUnitsNodepoolSummaryInternal(ctx context.Context, tenantId, nodepoolId string, nodepoolSummaryType api.OrgUnitsNodepoolSummaryType, isProject bool) ([]api.OrgUnitsNodepoolSummaryInner, error) {
	var tableName string
	if isProject {
		tableName = models.ProjectTable
	} else {
		tableName = models.DepartmentTable
	}

	query := dao_utils.DB().WithContext(ctx).
		Table(fmt.Sprintf("%s AS o", tableName)).
		Joins("JOIN resources on o.org_unit_data_id = resources.org_unit_data_id").
		Select("o.id, o.name").
		Where("o.tenant_id = ?", tenantId).
		Where("o.deleted_at IS NULL").
		Where("resources.deleted_at IS NULL").
		Where("resources.node_pool_id = ?", nodepoolId).
		Order("o.id ASC").
		Limit(1000)

	switch nodepoolSummaryType {
	case api.ORGUNITSNODEPOOLSUMMARYTYPE_HAS_QUOTA:
		query = query.Where(hasQuotaQueryCondition)
	case api.ORGUNITSNODEPOOLSUMMARYTYPE_NON_DEFAULT_RANK:
		query = query.Where(nonDefaultRankQueryCondition)
	default:
		return nil, fmt.Errorf("invalid nodepool summary type: %s", nodepoolSummaryType)
	}

	var results []api.OrgUnitsNodepoolSummaryInner
	result := query.Find(&results)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get org units nodepool summary for nodepool: %s, nodepoolSummaryType: %s, isProject: %t, error: %v",
			nodepoolId, nodepoolSummaryType, isProject, result.Error)
		return nil, result.Error
	}

	for i := range results {
		results[i].Type = tableName
	}

	return results, nil
}

func getResetResourcesResourcesQuery(projects bool, action ResetResourcesAction, extraConditions ...string) string {
	setQuery := ""
	conditionQuery := ""

	switch action {
	case QuotasToDefault:
		setQuery = `gpu_deserved = 0, gpu_limit = NULL,
                     cpu_deserved = NULL, cpu_limit = NULL,
                     memory_deserved = NULL, memory_limit = NULL, updated_at=now()`
		conditionQuery = hasQuotaQueryCondition
	case RanksToDefault:
		setQuery = fmt.Sprintf(`priority = '%s', updated_at=now()`, models.DefaultResourceRank)
		conditionQuery = nonDefaultRankQueryCondition
	case WeightsToDefault:
		// set gpu_over_quota_weight to default always, set cpu_over_quota_weight and memory_over_quota_weight to default only if not null
		setQuery = fmt.Sprintf(`gpu_over_quota_weight = %f, 
					 cpu_over_quota_weight = CASE WHEN cpu_over_quota_weight IS NOT NULL THEN %f ELSE NULL END, 
					 memory_over_quota_weight = CASE WHEN memory_over_quota_weight IS NOT NULL THEN %f ELSE NULL END, 
					 updated_at=now()`, models.DefaultOverQuotaWeight, models.DefaultOverQuotaWeight, models.DefaultOverQuotaWeight)
		conditionQuery = fmt.Sprintf(`(gpu_over_quota_weight IS DISTINCT FROM %f OR 
					 cpu_over_quota_weight <> %f OR 
					 memory_over_quota_weight <> %f)`,
			models.DefaultOverQuotaWeight, models.DefaultOverQuotaWeight, models.DefaultOverQuotaWeight)
	case WeightsToDeserved:
		// set gpu_over_quota_weight to gpu_deserved always, set cpu_over_quota_weight and memory_over_quota_weight to cpu_deserved/memory_deserved only if not null
		setQuery = `gpu_over_quota_weight = gpu_deserved, 
					 cpu_over_quota_weight = cpu_deserved, 
					 memory_over_quota_weight = memory_deserved, 
					 updated_at=now()`
		conditionQuery = `(gpu_over_quota_weight IS DISTINCT FROM gpu_deserved OR
					 cpu_over_quota_weight IS DISTINCT FROM cpu_deserved OR
					 memory_over_quota_weight IS DISTINCT FROM memory_deserved)`
	case CpuQuotasToDefault:
		setQuery = disableCpuQuotaQuerySet
		conditionQuery = disableCpuQuotaQueryCondition
	default:
		return ""
	}

	if len(extraConditions) > 0 {
		conditionQuery = fmt.Sprintf("(%s AND %s)", conditionQuery, strings.Join(extraConditions, " AND "))
	}

	return getUpdateResourcesQuery(setQuery, conditionQuery, projects)
}

func ResetResources(ctx context.Context, tenantId, clusterId, departmentId string, nodepoolIds []string, action api.ResourcesResetAction, projects bool) error {
	tx := dao_utils.DB().WithContext(ctx)

	err := resetResourcesInternal(ctx, tx, tenantId, clusterId, departmentId, nodepoolIds, ResetResourcesAction(action), projects)
	if err != nil {
		return err
	}

	InvalidateAllCache(ctx, tenantId, clusterId)
	return nil
}

func resetResourcesInternal(ctx context.Context, tx *gorm.DB, tenantId, clusterId, departmentId string, nodepoolIds []string, action ResetResourcesAction, projects bool) error {
	// tenant_id and cluster_id are part of the query in the helper func
	// nodepool_id and parent_id - we need to add as extra conditions here

	extraConds := []string{"resources.node_pool_id IN ?"}
	extraArgs := []interface{}{nodepoolIds}

	// departmentId filter is relevant for projects: select only projects with the given parent department
	if projects && departmentId != "" {
		extraConds = append(extraConds, fmt.Sprintf("%s.parent_id = ?", models.ProjectTable))
		extraArgs = append(extraArgs, departmentId)
	}

	// Build final UPDATE ... FROM ... WHERE ... statement using helper funcs
	sql := getResetResourcesResourcesQuery(projects, action, extraConds...)
	if sql == "" {
		return fmt.Errorf("invalid reset resources action: %v", action)
	}

	// Execute with parameters in order: extraArgs (nodepool/department), then tenant and cluster
	result := tx.Exec(sql, append(extraArgs, tenantId, clusterId)...)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to reset resources, action <%v>, cluster <%s>, nodepools <%v> (projects=%t, department <%s>): %v",
			action, clusterId, nodepoolIds, projects, departmentId, result.Error)
		return result.Error
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("ResetResources number of rows affected: %d, action <%v>, cluster <%s>, nodepools <%v> (projects=%t, department <%s>)",
		result.RowsAffected, action, clusterId, nodepoolIds, projects, departmentId)

	return nil
}

// NodepoolResetResources - implementation for the internal API call
func NodepoolResetResources(ctx context.Context, tenantId, clusterId, nodepoolId string, action api.NodepoolResourcesResetAction) error {
	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to db session start: %v", tx.Error)
		return tx.Error
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	// preform reset for both projects and departments
	for _, forProjects := range []bool{true, false} {
		err := resetResourcesInternal(ctx, tx, tenantId, clusterId, "", []string{nodepoolId}, ResetResourcesAction(action), forProjects)
		if err != nil {
			return err
		}
	}

	result := tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit nodepool reset resources, action <%v>, cluster <%s>, nodepool <%s>: %v",
			action, clusterId, nodepoolId, result.Error)
		return result.Error
	}
	committed = true

	InvalidateAllCache(ctx, tenantId, clusterId)
	return nil
}
