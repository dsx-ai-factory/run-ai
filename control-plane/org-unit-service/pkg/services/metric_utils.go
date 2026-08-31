/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"

	cluster_service_client "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
	metrics_client "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
	openapi "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/runaimetrics"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/versions"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	GpuMemoryUtilizationPerOrgUnitClusterMinVersion = "2.21"

	maxAllowedMetricsInRequest = 10
)

var metricsOnlySupportedFromVersion = map[api.OrgUnitMetricType]string{
	api.ORGUNITMETRICTYPE_GPU_MEMORY_UTILIZATION: GpuMemoryUtilizationPerOrgUnitClusterMinVersion,
}

var rankingMetricsOnlySupportedFromVersion = map[api.RankingMetricType]string{
	api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION24H: GpuMemoryUtilizationPerOrgUnitClusterMinVersion,
	api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION7D:  GpuMemoryUtilizationPerOrgUnitClusterMinVersion,
	api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION30D: GpuMemoryUtilizationPerOrgUnitClusterMinVersion,
}

func ValidateRequestedMetrics(metricTypesLen int) *api.Error {
	if metricTypesLen > maxAllowedMetricsInRequest {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("Too many metrics requested. Max allowed is %d", maxAllowedMetricsInRequest), nil)
	}

	return nil
}

func QueryProjectMetric(ctx context.Context,
	clusterId string, projectName, departmentName, nodePoolName *string,
	metricType api.OrgUnitMetricType, start time.Time, end time.Time, numberOfSamples int32,
	groupBy string,
) *openapi.GetMetrics200Response {
	var metric metrics_client.Metric

	clusterFilter := []string{clusterId}
	projectsFilter := getProjectFilterForMetric(projectName, metricType)
	departmentsFilter := getFilterForString(departmentName)
	nodePoolFilter := getFilterForString(nodePoolName)

	switch metricType {
	case api.ORGUNITMETRICTYPE_GPU_QUOTA:
		metric = runaimetrics.RunaiDeservedGpusCountPerQueuesMetric{
			Filters: runaimetrics.RunaiDeservedGpusCountPerQueuesFilters{
				ClusterId:      clusterFilter,
				Project:        projectsFilter,
				NodePool:       nodePoolFilter,
				Department:     departmentsFilter,
				DepartmentName: departmentsFilter,
			},
		}
	case api.ORGUNITMETRICTYPE_CPU_QUOTA_MILLICORES:
		metric = runaimetrics.RunaiDeservedCpuMillicoresPerQueuesMetric{
			Filters: runaimetrics.RunaiDeservedCpuMillicoresPerQueuesFilters{
				ClusterId:      clusterFilter,
				Project:        projectsFilter,
				NodePool:       nodePoolFilter,
				Department:     departmentsFilter,
				DepartmentName: departmentsFilter,
			},
		}
	case api.ORGUNITMETRICTYPE_CPU_MEMORY_QUOTA_MB:
		metric = runaimetrics.RunaiDeservedMemoryMbPerQueuesMetric{
			Filters: runaimetrics.RunaiDeservedMemoryMbPerQueuesFilters{
				ClusterId:      clusterFilter,
				Project:        projectsFilter,
				NodePool:       nodePoolFilter,
				Department:     departmentsFilter,
				DepartmentName: departmentsFilter,
			},
		}
	case api.ORGUNITMETRICTYPE_GPU_ALLOCATION:
		metric = runaimetrics.RunaiAllocatedGpuCountPerQueuesMetric{
			Filters: runaimetrics.RunaiAllocatedGpuCountPerQueuesFilters{
				ClusterId:  clusterFilter,
				Project:    projectsFilter,
				NodePool:   nodePoolFilter,
				Department: departmentsFilter,
			},
		}
	case api.ORGUNITMETRICTYPE_CPU_ALLOCATION_MILLICORES:
		metric = runaimetrics.RunaiAllocatedCpuCountPerQueuesMetric{
			Filters: runaimetrics.RunaiAllocatedCpuCountPerQueuesFilters{
				ClusterId:  clusterFilter,
				Project:    projectsFilter,
				NodePool:   nodePoolFilter,
				Department: departmentsFilter,
			},
		}
	case api.ORGUNITMETRICTYPE_CPU_MEMORY_ALLOCATION_MB:
		metric = runaimetrics.RunaiAllocatedMemoryMbPerQueuesMetric{
			Filters: runaimetrics.RunaiAllocatedMemoryMbPerQueuesFilters{
				ClusterId:  clusterFilter,
				Project:    projectsFilter,
				NodePool:   nodePoolFilter,
				Department: departmentsFilter,
			},
		}
	case api.ORGUNITMETRICTYPE_GPU_MEMORY_UTILIZATION:
		metric = runaimetrics.RunaiGpuMemoryUtilizationPerQueuesMetric{
			Filters: runaimetrics.RunaiGpuMemoryUtilizationPerQueuesFilters{
				ClusterId:  clusterFilter,
				Project:    projectsFilter,
				NodePool:   nodePoolFilter,
				Department: departmentsFilter,
			},
		}
	case api.ORGUNITMETRICTYPE_GPU_UTILIZATION:
		metric = runaimetrics.RunaiGpuUtilizationPerQueuesMetric{
			Filters: runaimetrics.RunaiGpuUtilizationPerQueuesFilters{
				ClusterId:  clusterFilter,
				Project:    projectsFilter,
				NodePool:   nodePoolFilter,
				Department: departmentsFilter,
			},
		}
	case api.ORGUNITMETRICTYPE_CPU_UTILIZATION:
		metric = runaimetrics.RunaiCpuUtilizationPerQueuesMetric{
			Filters: runaimetrics.RunaiCpuUtilizationPerQueuesFilters{
				ClusterId:  clusterFilter,
				Project:    projectsFilter,
				NodePool:   nodePoolFilter,
				Department: departmentsFilter,
			},
		}
	case api.ORGUNITMETRICTYPE_CPU_MEMORY_UTILIZATION:
		metric = runaimetrics.RunaiMemoryUtilizationPerQueuesMetric{
			Filters: runaimetrics.RunaiMemoryUtilizationPerQueuesFilters{
				ClusterId:  clusterFilter,
				Project:    projectsFilter,
				NodePool:   nodePoolFilter,
				Department: departmentsFilter,
			},
		}
	case api.ORGUNITMETRICTYPE_WORKLOADS_COUNT:
		metricGroupBy := metrics_client.GroupByDepartment
		if groupBy == string(metrics_client.GroupByCategory) {
			metricGroupBy = metrics_client.GroupByCategory
		} else if projectName != nil && *projectName != "" {
			metricGroupBy = metrics_client.GroupByProject
		}
		metric = runaimetrics.RunaiWorkloadsCountPerQueuesMetric{
			Filters: runaimetrics.RunaiWorkloadsCountPerQueuesFilters{
				ClusterId:  clusterFilter,
				Project:    projectsFilter,
				NodePool:   nodePoolFilter,
				Department: departmentsFilter,
			},
			GroupBy: metricGroupBy,
		}
	default:
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Unknown metric type %s.", metricType)
		return nil
	}

	return submitMetric(ctx, metric, string(metricType), start, end, numberOfSamples)
}

// ParseGroupedMetricResultToMeasurements creates one measurement per group (e.g. per category).
// Each Result in the response has MetricLabels identifying the group value.
func ParseGroupedMetricResultToMeasurements(ctx context.Context, response openapi.GetMetrics200Response, groupBy string) []api.MeasurementResponse {
	groupByLabelKey := groupByToLabelKey(groupBy)
	results := response.GetData()
	measurements := []api.MeasurementResponse{}

	for _, result := range results {
		groupValue := result.MetricLabels[groupByLabelKey]
		if groupValue == "" {
			continue
		}
		values := parseMetricSingleResultToMeasurement(result)
		measurements = append(measurements, api.MeasurementResponse{
			Type:   groupValue,
			Values: values,
		})
	}

	if len(measurements) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("No grouped results for groupBy %s", groupBy)
	}
	return measurements
}

func groupByToLabelKey(groupBy string) string {
	switch groupBy {
	case string(metrics_client.GroupByCategory):
		return "workload_category"
	default:
		return groupBy
	}
}

func ParseOrgUnitMetricResultToMeasurement(ctx context.Context, metricType string, response openapi.GetMetrics200Response) api.MeasurementResponse {
	measurement := api.MeasurementResponse{
		Type:   metricType,
		Values: &[]api.MeasurementResponseValuesInner{},
	}

	results := response.GetData()
	if len(results) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("No result return for metric type %s", metricType)
		return measurement
	}

	*measurement.Values = parseMetricResultsToMeasurement(results)
	return measurement
}

func parseMetricResultsToMeasurement(results []openapi.Result) []api.MeasurementResponseValuesInner {
	values := []api.MeasurementResponseValuesInner{}
	for _, result := range results {
		values = append(values, *parseMetricSingleResultToMeasurement(result)...)
	}

	return values
}

func parseMetricSingleResultToMeasurement(result openapi.Result) *[]api.MeasurementResponseValuesInner {
	values := &[]api.MeasurementResponseValuesInner{}
	for _, m := range result.Measurements {
		*values = append(*values, api.MeasurementResponseValuesInner{
			Value:     m.Value,
			Timestamp: convutils.Ref(m.Timestamp),
		})
	}

	return values
}

type StringLike interface {
	~string
}

func GetMetricsCsv[T StringLike](response *api.MetricsResponse, metricTypes []T) [][]string {
	var csvData [][]string

	timeToMeasurements := getTimeMeasurementsMap(response)
	orderedTimestamps := getOrderedTimestamps(timeToMeasurements)

	titles := []string{"Timestamp"}
	for _, metricType := range metricTypes {
		titles = append(titles, string(metricType))
	}

	csvData = append(csvData, titles)

	for _, timestamp := range orderedTimestamps {
		row := []string{}
		row = append(row, timestamp.String())
		timestampVal := timeToMeasurements[timestamp]
		for _, metricType := range metricTypes {
			metricValue := timestampVal[string(metricType)]
			row = append(row, metricValue)
		}
		csvData = append(csvData, row)
	}

	return csvData
}

func getTimeMeasurementsMap(response *api.MetricsResponse) map[time.Time]map[string]string {
	timestampMap := make(map[time.Time]map[string]string)

	for _, m := range response.Measurements {
		if m.Values != nil {
			for _, v := range *m.Values {
				if _, ok := timestampMap[*v.Timestamp]; !ok {
					timestampMap[*v.Timestamp] = make(map[string]string)
				}
				timestampMap[*v.Timestamp][m.Type] = v.Value
			}
		}
	}

	return timestampMap
}

func getOrderedTimestamps(timeToMeasurements map[time.Time]map[string]string) []time.Time {
	var orderedTimestamps []time.Time
	for k := range timeToMeasurements {
		orderedTimestamps = append(orderedTimestamps, k)
	}

	sort.Slice(orderedTimestamps, func(i, j int) bool {
		return orderedTimestamps[i].Before(orderedTimestamps[j])
	})

	return orderedTimestamps
}

func submitMetric(ctx context.Context, metric metrics_client.Metric, metricType string, start time.Time, end time.Time, numberOfSamples int32) *openapi.GetMetrics200Response {
	q, err := metrics_client.NewMetricQuery(metric)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to construct query for metric type %s. err: %v", metricType, err)
		return nil
	}

	q.SetRange(start, end)
	q.SetDefaultNumOfSamples(numberOfSamples)
	res, err := q.Query(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to query metric type %s. err: %v", metricType, err)
		return nil
	}

	return res
}

func getProjectFilterForMetric(projectName *string, metricType api.OrgUnitMetricType) []string {
	if projectName != nil && *projectName != "" {
		return []string{*projectName}
	}

	switch metricType {
	case api.ORGUNITMETRICTYPE_GPU_QUOTA,
		api.ORGUNITMETRICTYPE_CPU_QUOTA_MILLICORES,
		api.ORGUNITMETRICTYPE_CPU_MEMORY_QUOTA_MB:
		// for quota, the metrics are complicated with backwards compatibility, which requires all
		// the filters to have at least 1 value. but for the department quota we want project=""
		// so we send an empty string in the array.
		return []string{""}
	case api.ORGUNITMETRICTYPE_GPU_ALLOCATION,
		api.ORGUNITMETRICTYPE_CPU_ALLOCATION_MILLICORES,
		api.ORGUNITMETRICTYPE_CPU_MEMORY_ALLOCATION_MB,
		api.ORGUNITMETRICTYPE_GPU_MEMORY_UTILIZATION,
		api.ORGUNITMETRICTYPE_GPU_UTILIZATION,
		api.ORGUNITMETRICTYPE_CPU_UTILIZATION,
		api.ORGUNITMETRICTYPE_CPU_MEMORY_UTILIZATION,
		api.ORGUNITMETRICTYPE_WORKLOADS_COUNT:
		// for the allocation/utilization/count metrics - if project name is not specified,
		// don't use the filter at all.
		return []string{}
	}

	return []string{}
}

func getFilterForString(str *string) []string {
	if str == nil || *str == "" {
		// some of the metrics require at least 1 value in the filter
		// because of specific filters for backwards compatibility,
		// so if not specified - return a wildcard.
		return []string{".*"}
	}

	return []string{*str}
}

func ValidateMinClusterVersionForMetrics(ctx context.Context, clusterId uuid.UUID, metricTypes []api.OrgUnitMetricType,
	clusterClient cluster_service_client.IClusterExternalClient,
) error {
	var minVersionToValidate string
	var metricOfInterest string
	for _, metricType := range metricTypes {
		minVersion, ok := metricsOnlySupportedFromVersion[metricType]
		if !ok {
			continue
		}

		if minVersionToValidate == "" || minVersion < minVersionToValidate {
			minVersionToValidate = minVersion
			metricOfInterest = string(metricType)
		}
	}

	if minVersionToValidate == "" {
		return nil
	}

	clusterVersion, err := GetClusterVersion(ctx, clusterId, clusterClient)
	if err != nil {
		return err
	}

	return validateMetricsSupported(clusterVersion, minVersionToValidate, metricOfInterest)
}

func ValidateMinClusterVersionForRankingMetric(
	ctx context.Context,
	clusterId uuid.UUID,
	metricType api.RankingMetricType,
	clusterClient cluster_service_client.IClusterExternalClient,
) error {
	minVersion, ok := rankingMetricsOnlySupportedFromVersion[metricType]
	if !ok {
		return nil
	}

	clusterVersion, err := GetClusterVersion(ctx, clusterId, clusterClient)
	if err != nil {
		return err
	}

	return validateMetricsSupported(clusterVersion, minVersion, string(metricType))
}

func validateMetricsSupported(clusterVersion string, minSupportedVersion, metricType string) error {
	supported, err := versions.IsVersionSupported(clusterVersion, minSupportedVersion)
	if err != nil {
		return err
	}
	if !supported {
		return fmt.Errorf("%s metric is not supported for cluster version %s", metricType, clusterVersion)
	}
	return nil
}
