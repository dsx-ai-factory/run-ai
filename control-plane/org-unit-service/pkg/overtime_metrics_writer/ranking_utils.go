package overtime_metrics_writer

import (
	"context"
	"fmt"
	"strconv"

	metrics_client "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/runaimetrics"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

func queryOvertimeMetricDataForRanking(
	ctx context.Context, tenantId string, overtimeMetricRange dao.OvertimeMetricRange, overtimeMetricType OvertimeMetricType,
	clusterFilter []string, nodePoolFilter []string, departmentFilter []string, projectFilter []string, queryProjects bool, result map[string]map[string]*models.OvertimeData,
) error {
	metric, err := getOvertimeMetricForRanking(overtimeMetricType, clusterFilter, nodePoolFilter, departmentFilter, projectFilter)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get metric for ranking metric %v, error: %v", overtimeMetricRange, err)
		return err
	}

	queryResults, err := queryMetric(ctx, tenantId, metric, overtimeMetricRange)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(
			"Failed to get metric results for tenant %s, overtimeMetricRange: %v, overtimeMetricType: %v, queryProjects: %t, error: %v",
			tenantId, overtimeMetricRange, overtimeMetricType, queryProjects, err)
		return err
	}

	avgMetricType, err := mapMetricToAvgMetric(overtimeMetricType)
	if err != nil {
		return err
	}
	parseTenantOvertimeData(ctx, queryResults, overtimeMetricRange, avgMetricType, queryProjects, result)
	return nil
}

func mapOvertimeDataToProjectGetRanking(ctx context.Context, tenantId string, data map[string]map[string]*models.OvertimeData, timeRange dao.OvertimeMetricRange, nodePoolName string) ([]api.ProjectGetRanking, error) {
	var rankings []api.ProjectGetRanking

	for clusterId, projects := range data {
		for projectName, overtimeData := range projects {
			ranking := api.ProjectGetRanking{
				Name:                   projectName,
				Id:                     clusterId,
				ParentName:             "",
				AllocationValue:        getAllocationValue(overtimeData, timeRange),
				UtilizationValue:       getUtilizationValue(overtimeData, timeRange),
				MemoryUtilizationValue: getMemoryUtilizationValue(overtimeData, timeRange),
			}

			err := enrichProjectDetails(ctx, tenantId, clusterId, &ranking, nodePoolName)
			if err != nil {
				return nil, err
			}

			rankings = append(rankings, ranking)
		}
	}

	return rankings, nil
}

func enrichProjectDetails(ctx context.Context, tenantId string, clusterId string, result *api.ProjectGetRanking, nodePoolName string) error {
	filters, filterErr := models.Project{}.FieldsConfiguration().ValidateAndParameterizeFilters([]string{fmt.Sprintf("clusterId==%s", clusterId), fmt.Sprintf("name==%s", result.Name)})
	if filterErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to validate filter for project %s, error: %v", result.Name, filterErr)
		return filterErr
	}

	projects, err := dao.GetProjects(ctx, tenantId, filters, "", "", 0, 1, "", nil, false, models.SlimVerbosity)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get project for tenant %s, clusterId %s, project name %s, error: %v", tenantId, clusterId, result.Name, err)
		return err
	}
	if len(*projects) > 0 {
		result.ParentName = *(*projects)[0].ParentName
		result.Id = strconv.FormatInt((*projects)[0].ID, 10)

		if nodePoolName != "" {
			result.GpuQuota = extractGpuQuotaFromResources((*projects)[0].OrgUnitData.Resources, nodePoolName)
		} else {
			result.GpuQuota = (*projects)[0].OrgUnitData.TotalGpuQuota
		}
	}
	return nil
}

func mapOvertimeDataToDepartmentGetRanking(ctx context.Context, tenantId string,
	data map[string]map[string]*models.OvertimeData, timeRange dao.OvertimeMetricRange, nodePoolName string,
) ([]api.DepartmentGetRanking, error) {
	var rankings []api.DepartmentGetRanking

	for clusterId, departments := range data {
		for departmentName, overtimeData := range departments {
			ranking := api.DepartmentGetRanking{
				Name:                   departmentName,
				Id:                     clusterId,
				AllocationValue:        getAllocationValue(overtimeData, timeRange),
				UtilizationValue:       getUtilizationValue(overtimeData, timeRange),
				MemoryUtilizationValue: getMemoryUtilizationValue(overtimeData, timeRange),
			}
			err := enrichDepartmentDetails(ctx, tenantId, clusterId, &ranking, nodePoolName)
			if err != nil {
				return nil, err
			}

			rankings = append(rankings, ranking)

		}
	}

	return rankings, nil
}

func enrichDepartmentDetails(ctx context.Context, tenantId string, clusterId string, result *api.DepartmentGetRanking, nodePoolName string) error {
	filters, filterErr := models.Department{}.FieldsConfiguration().ValidateAndParameterizeFilters([]string{fmt.Sprintf("clusterId==%s", clusterId), fmt.Sprintf("name==%s", result.Name)})
	if filterErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to validate filter for department %s, error: %v", result.Name, filterErr)
		return filterErr
	}

	departments, err := dao.GetDepartments(ctx, tenantId, filters, "", "", 0, 1, nil, false, models.SlimVerbosity)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get departments for tenant %s, clusterId %s, department name %s, error: %v", tenantId, clusterId, result.Name, err)
		return err
	}
	if len(*departments) > 0 {
		result.Id = strconv.FormatInt((*departments)[0].ID, 10)

		if nodePoolName != "" {
			result.GpuQuota = extractGpuQuotaFromResources((*departments)[0].OrgUnitData.Resources, nodePoolName)
		} else {
			result.GpuQuota = (*departments)[0].OrgUnitData.TotalGpuQuota
		}
	}
	return nil
}

func getAllocationValue(overtimeData *models.OvertimeData, timeRange dao.OvertimeMetricRange) *float64 {
	switch timeRange {
	case dao.RangeDay:
		return overtimeData.AvgGpuAllocation24h
	case dao.RangeWeek:
		return overtimeData.AvgGpuAllocation7d
	case dao.RangeMonth:
		return overtimeData.AvgGpuAllocation30d
	default:
		return nil
	}
}

func getUtilizationValue(overtimeData *models.OvertimeData, timeRange dao.OvertimeMetricRange) *float64 {
	switch timeRange {
	case dao.RangeDay:
		return overtimeData.AvgGpuUtilization24h
	case dao.RangeWeek:
		return overtimeData.AvgGpuUtilization7d
	case dao.RangeMonth:
		return overtimeData.AvgGpuUtilization30d
	default:
		return nil
	}
}

func getMemoryUtilizationValue(overtimeData *models.OvertimeData, timeRange dao.OvertimeMetricRange) *float64 {
	switch timeRange {
	case dao.RangeDay:
		return overtimeData.AvgGpuMemoryUtilization24h
	case dao.RangeWeek:
		return overtimeData.AvgGpuMemoryUtilization7d
	case dao.RangeMonth:
		return overtimeData.AvgGpuMemoryUtilization30d
	default:
		return nil
	}
}

func extractGpuQuotaFromResources(resources []models.Resources, nodePoolName string) *float64 {
	for _, resource := range resources {
		if resource.NodePoolName == nodePoolName {
			return resource.GpuDeserved
		}
	}
	return nil
}

func getOvertimeMetricForRanking(overtimeMetricType OvertimeMetricType, clusterIdFilter []string, nodePoolFilter []string, departmentFilter []string, projectFilter []string) (metrics_client.Metric, error) {
	var metric metrics_client.Metric
	switch overtimeMetricType {
	case GpuUtilizationPerProjectNodepoolTop15:
		metric = runaimetrics.RunaiAvgGpuUtilizationPerProjectNodePoolTop15{
			Filters: runaimetrics.RunaiAvgGpuUtilizationPerProjectNodePoolTop15Filters{
				ClusterId:  clusterIdFilter,
				Department: departmentFilter,
				NodePool:   nodePoolFilter,
				Project:    projectFilter,
			},
		}
	case GpuAllocationPerProjectNodepoolTop15:
		metric = runaimetrics.RunaiAvgAllocatedGpuCountPerProjectNodePoolTop15{
			Filters: runaimetrics.RunaiAvgAllocatedGpuCountPerProjectNodePoolTop15Filters{
				ClusterId:  clusterIdFilter,
				Department: departmentFilter,
				NodePool:   nodePoolFilter,
				Project:    projectFilter,
			},
		}
	case GpuMemoryUtilizationPerProjectNodepoolTop15:
		metric = runaimetrics.RunaiAvgGpuMemoryUtilizationPerProjectNodePoolTop15{
			Filters: runaimetrics.RunaiAvgGpuMemoryUtilizationPerProjectNodePoolTop15Filters{
				ClusterId:  clusterIdFilter,
				Department: departmentFilter,
				NodePool:   nodePoolFilter,
				Project:    projectFilter,
			},
		}
	case GpuUtilizationPerProjectNodepoolBottom15:
		metric = runaimetrics.RunaiAvgGpuUtilizationPerProjectNodePoolBottom15{
			Filters: runaimetrics.RunaiAvgGpuUtilizationPerProjectNodePoolBottom15Filters{
				ClusterId:  clusterIdFilter,
				Department: departmentFilter,
				NodePool:   nodePoolFilter,
				Project:    projectFilter,
			},
		}
	case GpuAllocationPerProjectNodepoolBottom15:
		metric = runaimetrics.RunaiAvgAllocatedGpuCountPerProjectNodePoolBottom15{
			Filters: runaimetrics.RunaiAvgAllocatedGpuCountPerProjectNodePoolBottom15Filters{
				ClusterId:  clusterIdFilter,
				Department: departmentFilter,
				NodePool:   nodePoolFilter,
				Project:    projectFilter,
			},
		}
	case GpuMemoryUtilizationPerProjectNodepoolBottom15:
		metric = runaimetrics.RunaiAvgGpuMemoryUtilizationPerProjectNodePoolBottom15{
			Filters: runaimetrics.RunaiAvgGpuMemoryUtilizationPerProjectNodePoolBottom15Filters{
				ClusterId:  clusterIdFilter,
				Department: departmentFilter,
				NodePool:   nodePoolFilter,
				Project:    projectFilter,
			},
		}
	case GpuUtilizationPerDepartmentNodepoolTop15:
		metric = runaimetrics.RunaiAvgGpuUtilizationPerDepartmentNodePoolTop15{
			Filters: runaimetrics.RunaiAvgGpuUtilizationPerDepartmentNodePoolTop15Filters{
				ClusterId:  clusterIdFilter,
				NodePool:   nodePoolFilter,
				Department: departmentFilter,
			},
		}
	case GpuAllocationPerDepartmentNodepoolTop15:
		metric = runaimetrics.RunaiAvgAllocatedGpuCountPerDepartmentNodePoolTop15{
			Filters: runaimetrics.RunaiAvgAllocatedGpuCountPerDepartmentNodePoolTop15Filters{
				ClusterId:  clusterIdFilter,
				NodePool:   nodePoolFilter,
				Department: departmentFilter,
			},
		}
	case GpuMemoryUtilizationPerDepartmentNodepoolTop15:
		metric = runaimetrics.RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolTop15{
			Filters: runaimetrics.RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolTop15Filters{
				ClusterId:  clusterIdFilter,
				NodePool:   nodePoolFilter,
				Department: departmentFilter,
			},
		}
	case GpuUtilizationPerDepartmentNodepoolBottom15:
		metric = runaimetrics.RunaiAvgGpuUtilizationPerDepartmentNodePoolBottom15{
			Filters: runaimetrics.RunaiAvgGpuUtilizationPerDepartmentNodePoolBottom15Filters{
				ClusterId:  clusterIdFilter,
				NodePool:   nodePoolFilter,
				Department: departmentFilter,
			},
		}
	case GpuAllocationPerDepartmentNodepoolBottom15:
		metric = runaimetrics.RunaiAvgAllocatedGpuCountPerDepartmentNodePoolBottom15{
			Filters: runaimetrics.RunaiAvgAllocatedGpuCountPerDepartmentNodePoolBottom15Filters{
				ClusterId:  clusterIdFilter,
				NodePool:   nodePoolFilter,
				Department: departmentFilter,
			},
		}
	case GpuMemoryUtilizationPerDepartmentNodepoolBottom15:
		metric = runaimetrics.RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolBottom15{
			Filters: runaimetrics.RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolBottom15Filters{
				ClusterId:  clusterIdFilter,
				NodePool:   nodePoolFilter,
				Department: departmentFilter,
			},
		}
	default:
		return nil, fmt.Errorf("unknown metric type: %s", overtimeMetricType)
	}
	return metric, nil
}

func getTimeRangeForRankingMetric(metricType api.RankingMetricType) (dao.OvertimeMetricRange, error) {
	var overtimeMetricRange dao.OvertimeMetricRange

	switch metricType {
	case api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H,
		api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION24H,
		api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION24H:
		overtimeMetricRange = dao.RangeDay
	case api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION7D,
		api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION7D,
		api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION7D:
		overtimeMetricRange = dao.RangeWeek
	case api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION30D,
		api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION30D,
		api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION30D:
		overtimeMetricRange = dao.RangeMonth
	default:
		return "", fmt.Errorf("unknown ranking metric type: %s", metricType)
	}

	return overtimeMetricRange, nil
}

func getRankingMetricsList(metricType api.RankingMetricType, orderType api.RankingOrderType, queryProjects bool) ([]OvertimeMetricType, error) {
	if !queryProjects {
		return getDepartmentRankingMetricsList(metricType, orderType)
	}
	return getProjectRankingMetricsList(metricType, orderType)
}

func getDepartmentRankingMetricsList(metricType api.RankingMetricType, orderType api.RankingOrderType) ([]OvertimeMetricType, error) {
	if orderType == api.RANKINGORDERTYPE_TOP {
		switch metricType {
		case api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION7D, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION30D:
			return []OvertimeMetricType{GpuAllocationPerDepartmentNodepoolTop15, GpuUtilizationPerDepartmentNodepoolTop15, GpuMemoryUtilizationPerDepartmentNodepoolTop15}, nil
		case api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION24H, api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION7D, api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION30D:
			return []OvertimeMetricType{GpuUtilizationPerDepartmentNodepoolTop15, GpuAllocationPerDepartmentNodepoolTop15, GpuMemoryUtilizationPerDepartmentNodepoolTop15}, nil
		case api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION24H, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION7D, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION30D:
			return []OvertimeMetricType{GpuMemoryUtilizationPerDepartmentNodepoolTop15, GpuUtilizationPerDepartmentNodepoolTop15, GpuAllocationPerDepartmentNodepoolTop15}, nil
		default:
			return nil, fmt.Errorf("invalid metric type: %s", metricType)
		}
	} else {
		switch metricType {
		case api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION7D, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION30D:
			return []OvertimeMetricType{GpuAllocationPerDepartmentNodepoolBottom15, GpuUtilizationPerDepartmentNodepoolBottom15, GpuMemoryUtilizationPerDepartmentNodepoolBottom15}, nil
		case api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION24H, api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION7D, api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION30D:
			return []OvertimeMetricType{GpuUtilizationPerDepartmentNodepoolBottom15, GpuAllocationPerDepartmentNodepoolBottom15, GpuMemoryUtilizationPerDepartmentNodepoolBottom15}, nil
		case api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION24H, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION7D, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION30D:
			return []OvertimeMetricType{GpuMemoryUtilizationPerDepartmentNodepoolBottom15, GpuUtilizationPerDepartmentNodepoolBottom15, GpuAllocationPerDepartmentNodepoolBottom15}, nil
		default:
			return nil, fmt.Errorf("invalid metric type: %s", metricType)
		}
	}
}

func getProjectRankingMetricsList(metricType api.RankingMetricType, orderType api.RankingOrderType) ([]OvertimeMetricType, error) {
	if orderType == api.RANKINGORDERTYPE_TOP {
		switch metricType {
		case api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION7D, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION30D:
			return []OvertimeMetricType{GpuAllocationPerProjectNodepoolTop15, GpuUtilizationPerProjectNodepoolTop15, GpuMemoryUtilizationPerProjectNodepoolTop15}, nil
		case api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION24H, api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION7D, api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION30D:
			return []OvertimeMetricType{GpuUtilizationPerProjectNodepoolTop15, GpuAllocationPerProjectNodepoolTop15, GpuMemoryUtilizationPerProjectNodepoolTop15}, nil
		case api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION24H, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION7D, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION30D:
			return []OvertimeMetricType{GpuMemoryUtilizationPerProjectNodepoolTop15, GpuUtilizationPerProjectNodepoolTop15, GpuAllocationPerProjectNodepoolTop15}, nil
		default:
			return nil, fmt.Errorf("invalid metric type: %s", metricType)
		}
	} else {
		switch metricType {
		case api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION7D, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION30D:
			return []OvertimeMetricType{GpuAllocationPerProjectNodepoolBottom15, GpuUtilizationPerProjectNodepoolBottom15, GpuMemoryUtilizationPerProjectNodepoolBottom15}, nil
		case api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION24H, api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION7D, api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION30D:
			return []OvertimeMetricType{GpuUtilizationPerProjectNodepoolBottom15, GpuAllocationPerProjectNodepoolBottom15, GpuMemoryUtilizationPerProjectNodepoolBottom15}, nil
		case api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION24H, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION7D, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION30D:
			return []OvertimeMetricType{GpuMemoryUtilizationPerProjectNodepoolBottom15, GpuUtilizationPerProjectNodepoolBottom15, GpuAllocationPerProjectNodepoolBottom15}, nil
		default:
			return nil, fmt.Errorf("invalid metric type: %s", metricType)
		}
	}
}

func mapMetricToAvgMetric(metricType OvertimeMetricType) (OvertimeMetricType, error) {
	switch metricType {
	case GpuUtilizationPerProjectNodepoolTop15, GpuUtilizationPerProjectNodepoolBottom15,
		GpuUtilizationPerDepartmentNodepoolTop15, GpuUtilizationPerDepartmentNodepoolBottom15:
		return GpuUtilizationAvg, nil
	case GpuAllocationPerProjectNodepoolTop15, GpuAllocationPerProjectNodepoolBottom15,
		GpuAllocationPerDepartmentNodepoolTop15, GpuAllocationPerDepartmentNodepoolBottom15:
		return GpuAllocationAvg, nil
	case GpuMemoryUtilizationPerProjectNodepoolTop15, GpuMemoryUtilizationPerProjectNodepoolBottom15,
		GpuMemoryUtilizationPerDepartmentNodepoolTop15, GpuMemoryUtilizationPerDepartmentNodepoolBottom15:
		return GpuMemoryUtilizationAvg, nil
	default:
		return "", fmt.Errorf("invalid metric type: %s", metricType)
	}
}

func extractOrgUnitNamesFromOverTimeData(overTimeData map[string]map[string]*models.OvertimeData) []string {
	var names []string
	for _, orgUnitNames := range overTimeData {
		for name := range orgUnitNames {
			names = append(names, name)
		}
	}
	return names
}
