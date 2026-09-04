/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reports

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"time"

	auth_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	metrics_client "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	metrics_swagger "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/runaimetrics"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
)

const (
	ProjectNameFilter    = "projectName"
	DepartmentNameFilter = "departmentName"
	ClusterIdFilter      = "clusterId"
	NodepoolNameFilter   = "nodepool"

	gpuAllocationHoursTitle       = "GPU allocation hours"
	gpuIdleHoursTitle             = "GPU Idle hours"
	inQuotaGpuHoursTitle          = "GPU deserved quota hours"
	overQuotaGpuHoursTitle        = "GPU over quota hours"
	cpuAllocationHoursTitle       = "CPU cores allocation hours"
	cpuMemoryAllocationHoursTitle = "CPU Memory bytes allocation hours"
	cpuUsageHoursTitle            = "CPU cores usage hours"
	cpuMemoryUsageHoursTitle      = "CPU Memory bytes usage hours"
)

var (
	filterRegex      = regexp.MustCompile(`^(?P<key>[^\s=]+)==(?P<value>[^\s=]+)$`)
	supportedFilters = map[string]bool{
		ProjectNameFilter:    true,
		DepartmentNameFilter: true,
		ClusterIdFilter:      true,
		NodepoolNameFilter:   true,
	}
)

func SetPermissionFilters(ctx context.Context, tenantId string, filtersMap *map[string][]string) error {
	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		return err
	}

	if permittedScopes == nil ||
		permittedScopes.GetSystem() ||
		slices.Contains(permittedScopes.Tenants, tenantId) {
		return nil
	}

	var ok bool
	var clusterPermissions []string
	if clusterPermissions, ok = permittedScopes.GetClustersOk(); ok && len(clusterPermissions) > 0 {
		var clusterFilter []string
		if clusterFilter, ok = (*filtersMap)[ClusterIdFilter]; !ok {
			clusterFilter = []string{}
		}

		(*filtersMap)[ClusterIdFilter] = append(clusterFilter, clusterPermissions...)
	}

	filters := []string{}
	for _, permission := range clusterPermissions {
		filters = append(filters, fmt.Sprintf("clusterId!=%s", permission))
	}
	dbFilter, _ := queryapi.NewParser(models.Department{}.SortAndFilterFieldsToColumn()).Filters(filters)

	var departmentPermissions []string
	if departmentPermissions, ok = permittedScopes.GetDepartmentsOk(); ok && len(departmentPermissions) > 0 {
		departments, err := dao.GetDepartments(
			ctx,
			tenantId,
			dbFilter,
			"",
			"",
			0,
			0,
			permittedScopes,
			false,
			models.MetaDataVerbosity,
		)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get departments. err: %v", err)
			return err
		}

		if departments != nil && len(*departments) > 0 {
			for _, department := range *departments {
				(*filtersMap)[DepartmentNameFilter] = append((*filtersMap)[DepartmentNameFilter], department.Name)
			}
		}
	}

	if len(departmentPermissions) > 0 {
		filters = []string{}
		for _, permission := range departmentPermissions {
			filters = append(filters, fmt.Sprintf("parentId!=%s", permission))
		}
		dbFilterParent, _ := queryapi.NewParser(models.Project{}.SortAndFilterFieldsToColumn()).Filters(filters)
		dbFilter = append(dbFilter, dbFilterParent...)
	}
	if projectPermissions, ok := permittedScopes.GetProjectsOk(); ok && len(projectPermissions) > 0 {
		projects, err := dao.GetProjects(ctx, tenantId, dbFilter, "", "", 0, 0, "", permittedScopes, false, models.MetaDataVerbosity)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get projects. err: %v", err)
			return err
		}

		if projects != nil && len(*projects) > 0 {
			for _, project := range *projects {
				(*filtersMap)[ProjectNameFilter] = append((*filtersMap)[ProjectNameFilter], project.Name)
			}
		}
	}

	return nil
}

func reportValidations(fields api.ReportRequestFields) (map[string][]string, error) {
	if err := rangeValidation(fields.Start, fields.End); err != nil {
		return nil, err
	}

	filtersMap, err := filterValidation(fields.FilterBy)
	if err != nil {
		return nil, err
	}

	if err := groupAndFilterValidation(fields.GroupBy, filtersMap); err != nil {
		return nil, err
	}

	return filtersMap, nil
}

func parseFilter(input string) (string, string, error) {
	match := filterRegex.FindStringSubmatch(input)
	if match == nil {
		return "", "", fmt.Errorf("filter does not match the required format: <key>==<value>")
	}

	key := match[filterRegex.SubexpIndex("key")]
	value := match[filterRegex.SubexpIndex("value")]

	return key, value, nil
}

func filterValidation(filterBy []string) (map[string][]string, error) {
	filtersMap := map[string][]string{}
	for _, filter := range filterBy {
		key, value, err := parseFilter(filter)
		if err != nil {
			return nil, err
		}

		if supported, ok := supportedFilters[key]; !ok || !supported {
			return nil, fmt.Errorf("filter %s is not supported", key)
		}

		values := []string{}
		if exists, ok := filtersMap[key]; ok {
			values = exists
		}
		values = append(values, value)
		filtersMap[key] = values
	}

	return filtersMap, nil
}

// ParseReportFilters exposes the internal filter validation logic for reuse and testing.
func ParseReportFilters(filterBy []string) (map[string][]string, error) {
	return filterValidation(filterBy)
}

func groupAndFilterValidation(groupBy *api.ReportGroupByOptions, filterBy map[string][]string) error {
	if groupBy == nil || len(filterBy) == 0 {
		return nil
	}

	departmentFilterDefined := len(filterBy[DepartmentNameFilter]) != 0
	projectFilterDefined := len(filterBy[ProjectNameFilter]) != 0

	switch *groupBy {
	case api.REPORTGROUPBYOPTIONS_CLUSTER:
		if departmentFilterDefined || projectFilterDefined {
			return fmt.Errorf("can not group by cluster with department or project filters or limitted permissions")
		}
	case api.REPORTGROUPBYOPTIONS_DEPARTMENT:
		if projectFilterDefined {
			return fmt.Errorf("can not group by department with project filter or limitted permissions")
		}
	}

	return nil
}

func rangeValidation(start time.Time, end time.Time) error {
	if !start.Before(end) {
		return fmt.Errorf("start time must be before end time")
	}
	if start.Add(32 * 24 * time.Hour).Before(end) {
		return fmt.Errorf("range exceeds 32 days limit")
	}

	return nil
}

func getReportLabels(groupBy api.ReportGroupByOptions) []metrics_swagger.Label {
	clusterLabel := metrics_swagger.Label{
		Name:        "clusterId",
		DisplayName: "Cluster",
	}
	departmentLabel := metrics_swagger.Label{
		Name:        "department",
		DisplayName: "Department",
	}
	projectLabel := metrics_swagger.Label{
		Name:        "project",
		DisplayName: "Project",
	}
	npLabel := metrics_swagger.Label{
		Name:        "nodepool",
		DisplayName: "Nodepool",
	}

	switch groupBy {
	case api.REPORTGROUPBYOPTIONS_CLUSTER:
		return []metrics_swagger.Label{
			clusterLabel,
		}
	case api.REPORTGROUPBYOPTIONS_DEPARTMENT:
		return []metrics_swagger.Label{
			clusterLabel,
			departmentLabel,
		}
	case api.REPORTGROUPBYOPTIONS_PROJECT:
		return []metrics_swagger.Label{
			clusterLabel,
			departmentLabel,
			projectLabel,
		}
	case api.REPORTGROUPBYOPTIONS_NODEPOOL:
		return []metrics_swagger.Label{
			clusterLabel,
			departmentLabel,
			projectLabel,
			npLabel,
		}
	}

	return []metrics_swagger.Label{
		clusterLabel,
		departmentLabel,
		projectLabel,
		npLabel,
	}
}

func getReportMetrics(ctx context.Context, groupBy api.ReportGroupByOptions, filterBy map[string][]string) []metrics_client.ReportColumn {
	var clusterFilter []uuid.UUID
	var departmentFilter []string
	var projectFilter []string
	var nodepoolFilter []string

	if values, ok := filterBy[ClusterIdFilter]; ok && len(values) != 0 {
		for _, value := range values {
			uuidValue, err := uuid.Parse(value)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("cluster id %s must be a uuid", value)
			} else {
				clusterFilter = append(clusterFilter, uuidValue)
			}
		}
	}
	if values, ok := filterBy[DepartmentNameFilter]; ok && len(values) != 0 {
		departmentFilter = values
	}
	if values, ok := filterBy[ProjectNameFilter]; ok && len(values) != 0 {
		projectFilter = values
	}
	if values, ok := filterBy[NodepoolNameFilter]; ok && len(values) != 0 {
		nodepoolFilter = values
	}

	switch groupBy {
	case api.REPORTGROUPBYOPTIONS_CLUSTER:
		return []metrics_client.ReportColumn{
			{
				DisplayName: gpuAllocationHoursTitle,
				Metric: runaimetrics.RunaiAllocatedGpuHoursPerClusterMetric{
					Filters: runaimetrics.RunaiAllocatedGpuHoursPerClusterFilters{
						ClusterId: clusterFilter,
					},
				},
			},
			{
				DisplayName: inQuotaGpuHoursTitle,
				Metric: runaimetrics.RunaiInQuotaGpuHoursMetric{
					Filters: runaimetrics.RunaiInQuotaGpuHoursFilters{
						ClusterId: clusterFilter,
					},
					GroupBy: metrics_client.GroupByCluster,
				},
			},
			{
				DisplayName: overQuotaGpuHoursTitle,
				Metric: runaimetrics.RunaiOverQuotaGpuHoursMetric{
					Filters: runaimetrics.RunaiOverQuotaGpuHoursFilters{
						ClusterId: clusterFilter,
					},
					GroupBy: metrics_client.GroupByCluster,
				},
			},
			{
				DisplayName: gpuIdleHoursTitle,
				Metric: runaimetrics.RunaiIdleGpuHoursPerClusterMetric{
					Filters: runaimetrics.RunaiIdleGpuHoursPerClusterFilters{
						ClusterId: clusterFilter,
					},
				},
			},
			{
				DisplayName: cpuAllocationHoursTitle,
				Metric: runaimetrics.RunaiAllocatedCpuCoresHoursPerClusterMetric{
					Filters: runaimetrics.RunaiAllocatedCpuCoresHoursPerClusterFilters{
						ClusterId: clusterFilter,
					},
				},
			},
			{
				DisplayName: cpuMemoryAllocationHoursTitle,
				Metric: runaimetrics.RunaiAllocatedMemoryBytesHoursPerClusterMetric{
					Filters: runaimetrics.RunaiAllocatedMemoryBytesHoursPerClusterFilters{
						ClusterId: clusterFilter,
					},
				},
			},
			{
				DisplayName: cpuUsageHoursTitle,
				Metric: runaimetrics.RunaiUsedCpuCoresHoursPerClusterMetric{
					Filters: runaimetrics.RunaiUsedCpuCoresHoursPerClusterFilters{
						ClusterId: clusterFilter,
					},
				},
			},
			{
				DisplayName: cpuMemoryUsageHoursTitle,
				Metric: runaimetrics.RunaiUsedMemoryBytesHoursPerClusterMetric{
					Filters: runaimetrics.RunaiUsedMemoryBytesHoursPerClusterFilters{
						ClusterId: clusterFilter,
					},
				},
			},
		}
	case api.REPORTGROUPBYOPTIONS_DEPARTMENT:
		return []metrics_client.ReportColumn{
			{
				DisplayName: gpuAllocationHoursTitle,
				Metric: runaimetrics.RunaiAllocatedGpuHoursPerDepartmentMetric{
					Filters: runaimetrics.RunaiAllocatedGpuHoursPerDepartmentFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
					},
				},
			},
			{
				DisplayName: inQuotaGpuHoursTitle,
				Metric: runaimetrics.RunaiInQuotaGpuHoursMetric{
					Filters: runaimetrics.RunaiInQuotaGpuHoursFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
					},
					GroupBy: metrics_client.GroupByDepartment,
				},
			},
			{
				DisplayName: overQuotaGpuHoursTitle,
				Metric: runaimetrics.RunaiOverQuotaGpuHoursMetric{
					Filters: runaimetrics.RunaiOverQuotaGpuHoursFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
					},
					GroupBy: metrics_client.GroupByDepartment,
				},
			},
			{
				DisplayName: gpuIdleHoursTitle,
				Metric: runaimetrics.RunaiIdleGpuHoursPerDepartmentMetric{
					Filters: runaimetrics.RunaiIdleGpuHoursPerDepartmentFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
					},
				},
			},
			{
				DisplayName: cpuAllocationHoursTitle,
				Metric: runaimetrics.RunaiAllocatedCpuCoresHoursPerDepartmentMetric{
					Filters: runaimetrics.RunaiAllocatedCpuCoresHoursPerDepartmentFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
					},
				},
			},
			{
				DisplayName: cpuMemoryAllocationHoursTitle,
				Metric: runaimetrics.RunaiAllocatedMemoryBytesHoursPerDepartmentMetric{
					Filters: runaimetrics.RunaiAllocatedMemoryBytesHoursPerDepartmentFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
					},
				},
			},
			{
				DisplayName: cpuUsageHoursTitle,
				Metric: runaimetrics.RunaiUsedCpuCoresHoursPerDepartmentMetric{
					Filters: runaimetrics.RunaiUsedCpuCoresHoursPerDepartmentFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
					},
				},
			},
			{
				DisplayName: cpuMemoryUsageHoursTitle,
				Metric: runaimetrics.RunaiUsedMemoryBytesHoursPerDepartmentMetric{
					Filters: runaimetrics.RunaiUsedMemoryBytesHoursPerDepartmentFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
					},
				},
			},
		}
	case api.REPORTGROUPBYOPTIONS_PROJECT:
		return []metrics_client.ReportColumn{
			{
				DisplayName: gpuAllocationHoursTitle,
				Metric: runaimetrics.RunaiAllocatedGpuHoursPerProjectMetric{
					Filters: runaimetrics.RunaiAllocatedGpuHoursPerProjectFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
						Project:    projectFilter,
					},
				},
			},
			{
				DisplayName: inQuotaGpuHoursTitle,
				Metric: runaimetrics.RunaiInQuotaGpuHoursMetric{
					Filters: runaimetrics.RunaiInQuotaGpuHoursFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
						Project:    projectFilter,
					},
					GroupBy: metrics_client.GroupByProject,
				},
			},
			{
				DisplayName: overQuotaGpuHoursTitle,
				Metric: runaimetrics.RunaiOverQuotaGpuHoursMetric{
					Filters: runaimetrics.RunaiOverQuotaGpuHoursFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
						Project:    projectFilter,
					},
					GroupBy: metrics_client.GroupByProject,
				},
			},
			{
				DisplayName: gpuIdleHoursTitle,
				Metric: runaimetrics.RunaiIdleGpuHoursPerProjectMetric{
					Filters: runaimetrics.RunaiIdleGpuHoursPerProjectFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
						Project:    projectFilter,
					},
				},
			},
			{
				DisplayName: cpuAllocationHoursTitle,
				Metric: runaimetrics.RunaiAllocatedCpuCoresHoursPerProjectMetric{
					Filters: runaimetrics.RunaiAllocatedCpuCoresHoursPerProjectFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
						Project:    projectFilter,
					},
				},
			},
			{
				DisplayName: cpuMemoryAllocationHoursTitle,
				Metric: runaimetrics.RunaiAllocatedMemoryBytesHoursPerProjectMetric{
					Filters: runaimetrics.RunaiAllocatedMemoryBytesHoursPerProjectFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
						Project:    projectFilter,
					},
				},
			},
			{
				DisplayName: cpuUsageHoursTitle,
				Metric: runaimetrics.RunaiUsedCpuCoresHoursPerProjectMetric{
					Filters: runaimetrics.RunaiUsedCpuCoresHoursPerProjectFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
						Project:    projectFilter,
					},
				},
			},
			{
				DisplayName: cpuMemoryUsageHoursTitle,
				Metric: runaimetrics.RunaiUsedMemoryBytesHoursPerProjectMetric{
					Filters: runaimetrics.RunaiUsedMemoryBytesHoursPerProjectFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
						Project:    projectFilter,
					},
				},
			},
		}
	case api.REPORTGROUPBYOPTIONS_NODEPOOL:
		return []metrics_client.ReportColumn{
			{
				DisplayName: gpuAllocationHoursTitle,
				Metric: runaimetrics.RunaiAllocatedGpuHoursPerQueueMetric{
					Filters: runaimetrics.RunaiAllocatedGpuHoursPerQueueFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
						Project:    projectFilter,
					},
				},
			},
			{
				DisplayName: inQuotaGpuHoursTitle,
				Metric: runaimetrics.RunaiInQuotaGpuHoursMetric{
					Filters: runaimetrics.RunaiInQuotaGpuHoursFilters{
						ClusterId: clusterFilter,
						Nodepool:  nodepoolFilter,
					},
					GroupBy: metrics_client.GroupByNodePool,
				},
			},
			{
				DisplayName: overQuotaGpuHoursTitle,
				Metric: runaimetrics.RunaiOverQuotaGpuHoursMetric{
					Filters: runaimetrics.RunaiOverQuotaGpuHoursFilters{
						ClusterId: clusterFilter,
						Nodepool:  nodepoolFilter,
					},
					GroupBy: metrics_client.GroupByNodePool,
				},
			},
			{
				DisplayName: gpuIdleHoursTitle,
				Metric: runaimetrics.RunaiIdleGpuHoursPerQueueMetric{
					Filters: runaimetrics.RunaiIdleGpuHoursPerQueueFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
						Project:    projectFilter,
					},
				},
			},
			{
				DisplayName: cpuAllocationHoursTitle,
				Metric: runaimetrics.RunaiAllocatedCpuCoresHoursPerQueueMetric{
					Filters: runaimetrics.RunaiAllocatedCpuCoresHoursPerQueueFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
						Project:    projectFilter,
					},
				},
			},
			{
				DisplayName: cpuMemoryAllocationHoursTitle,
				Metric: runaimetrics.RunaiAllocatedMemoryBytesHoursPerQueueMetric{
					Filters: runaimetrics.RunaiAllocatedMemoryBytesHoursPerQueueFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
						Project:    projectFilter,
					},
				},
			},
			{
				DisplayName: cpuUsageHoursTitle,
				Metric: runaimetrics.RunaiUsedCpuCoresHoursPerQueueMetric{
					Filters: runaimetrics.RunaiUsedCpuCoresHoursPerQueueFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
						Project:    projectFilter,
					},
				},
			},
			{
				DisplayName: cpuMemoryUsageHoursTitle,
				Metric: runaimetrics.RunaiUsedMemoryBytesHoursPerQueueMetric{
					Filters: runaimetrics.RunaiUsedMemoryBytesHoursPerQueueFilters{
						ClusterId:  clusterFilter,
						Department: departmentFilter,
						Project:    projectFilter,
					},
				},
			},
		}
	}

	return nil
}

func metricsReportToApi(report *metrics_swagger.Report) api.Report {
	filterBy, groupBy := parseAdditionalFieldsToReport(report.AdditionalFields)
	return api.Report{
		Name:            report.Name,
		Description:     convutils.Deref(report.Description, ""),
		Start:           report.Range.Get().Start,
		End:             report.Range.Get().End,
		GroupBy:         groupBy,
		FilterBy:        filterBy,
		Id:              report.Id,
		CreatedAt:       report.CreatedAt,
		CreatedBy:       report.CreatedBy,
		TenantId:        report.TenantId,
		Status:          api.ReportStatus(report.Status),
		StatusUpdatedAt: report.StatusUpdatedAt,
		StatusMessage:   convutils.Deref(report.StatusMessage, ""),
	}
}
