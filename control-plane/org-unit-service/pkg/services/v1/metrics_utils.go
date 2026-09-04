/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	metrics_client "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	metrics_api "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/runaimetrics"
)

const (
	departmentLabel                  = "department"
	oldDepartmentLabel               = "department_name"
	projectLabel                     = "project"
	nodePoolLabel                    = "nodepool"
	CurrentMetricPendingWorkloads    = "pendingWorkloadsCurrent"
	CurrentMetricAllocatedGpu        = "allocatedGpuCurrent"
	CurrentMetricAllocatedCpu        = "allocatedCpuCurrent"
	CurrentMetricAllocatedMemory     = "allocatedMemoryCurrent"
	TimerangeMetricPendingWorkloads  = "pendingWorkloadsTimerange"
	TimerangeMetricAllocatedGpu      = "allocatedGpuTimerange"
	TimerangeMetricAllocatedCpu      = "allocatedCpuTimerange"
	TimerangeMetricAllocatedMemory   = "allocatedMemoryTimerange"
	TimerangeMetricGpuQuota          = "gpuQuotaTimerange"
	TimerangeMetricCpuQuota          = "cpuQuotaTimerange"
	TimerangeMetricMemoryQuota       = "memoryQuotaTimerange"
	TimerangeMetricGpuUtilization    = "gpuUtilizationTimerange"
	TimerangeMetricCpuUtilization    = "cpuUtilizationTimerange"
	TimerangeMetricMemoryUtilization = "memoryUtilizationTimerange"
	MaxNumberOfSamples               = 100
	DefaultNumberOfSamples           = 20
)

func getMetric(metricType string, clusterFilter []string, departmentsFilter []string, projectsFilter []string, nodePoolFilter []string) (metrics_client.Metric, error) {
	var metric metrics_client.Metric
	switch metricType {
	case CurrentMetricPendingWorkloads:
		metric = runaimetrics.RunaiPendingWorkloadsPerQueueMetric{
			Filters: runaimetrics.RunaiPendingWorkloadsPerQueueFilters{
				ClusterId:  clusterFilter,
				Project:    projectsFilter,
				NodePool:   nodePoolFilter,
				Department: departmentsFilter,
			},
		}
	case CurrentMetricAllocatedGpu:
		metric = runaimetrics.RunaiAllocatedGpuPerQueueMetric{
			Filters: runaimetrics.RunaiAllocatedGpuPerQueueFilters{
				ClusterId:  clusterFilter,
				Project:    projectsFilter,
				NodePool:   nodePoolFilter,
				Department: departmentsFilter,
			},
		}
	case CurrentMetricAllocatedCpu:
		metric = runaimetrics.RunaiAllocatedCpuPerQueueMetric{
			Filters: runaimetrics.RunaiAllocatedCpuPerQueueFilters{
				ClusterId:  clusterFilter,
				Project:    projectsFilter,
				NodePool:   nodePoolFilter,
				Department: departmentsFilter,
			},
		}
	case CurrentMetricAllocatedMemory:
		metric = runaimetrics.RunaiAllocatedMemoryMBPerQueueMetric{
			Filters: runaimetrics.RunaiAllocatedMemoryMBPerQueueFilters{
				ClusterId:  clusterFilter,
				Project:    projectsFilter,
				NodePool:   nodePoolFilter,
				Department: departmentsFilter,
			},
		}
	case TimerangeMetricPendingWorkloads:
		metric = runaimetrics.RunaiPendingWorkloadsCountPerQueuesMetric{
			Filters: runaimetrics.RunaiPendingWorkloadsCountPerQueuesFilters{
				ClusterId:      clusterFilter,
				Project:        projectsFilter,
				NodePool:       nodePoolFilter,
				Department:     departmentsFilter,
				DepartmentName: departmentsFilter,
			},
		}
	case TimerangeMetricAllocatedGpu:
		metric = runaimetrics.RunaiAllocatedGpuCountPerQueuesMetric{
			Filters: runaimetrics.RunaiAllocatedGpuCountPerQueuesFilters{
				ClusterId:  clusterFilter,
				Project:    projectsFilter,
				NodePool:   nodePoolFilter,
				Department: departmentsFilter,
			},
		}
	case TimerangeMetricGpuQuota:
		metric = runaimetrics.RunaiDeservedGpusCountPerQueuesMetric{
			Filters: runaimetrics.RunaiDeservedGpusCountPerQueuesFilters{
				ClusterId:      clusterFilter,
				Project:        projectsFilter,
				NodePool:       nodePoolFilter,
				Department:     departmentsFilter,
				DepartmentName: departmentsFilter,
			},
		}
	case TimerangeMetricGpuUtilization:
		metric = runaimetrics.RunaiGpuUtilizationPerNodepoolsMetric{
			Filters: runaimetrics.RunaiGpuUtilizationPerNodepoolsFilters{
				ClusterId: clusterFilter,
				NodePool:  nodePoolFilter,
			},
		}
	case TimerangeMetricAllocatedCpu:
		metric = runaimetrics.RunaiAllocatedCpuCountPerQueuesMetric{
			Filters: runaimetrics.RunaiAllocatedCpuCountPerQueuesFilters{
				ClusterId:  clusterFilter,
				Project:    projectsFilter,
				NodePool:   nodePoolFilter,
				Department: departmentsFilter,
			},
		}
	case TimerangeMetricCpuQuota:
		metric = runaimetrics.RunaiDeservedCpuMetric{
			Filters: runaimetrics.RunaiDeservedCpuPerQueuesFilters{
				ClusterId:      clusterFilter,
				Project:        projectsFilter,
				NodePool:       nodePoolFilter,
				Department:     departmentsFilter,
				DepartmentName: departmentsFilter,
			},
		}
	case TimerangeMetricCpuUtilization:
		metric = runaimetrics.RunaiCpuUtilizationPerNodepoolsMetric{
			Filters: runaimetrics.RunaiCpuUtilizationPerNodepoolsFilters{
				ClusterId: clusterFilter,
				NodePool:  nodePoolFilter,
			},
		}
	case TimerangeMetricAllocatedMemory:
		metric = runaimetrics.RunaiAllocatedMemoryMbPerQueuesMetric{
			Filters: runaimetrics.RunaiAllocatedMemoryMbPerQueuesFilters{
				ClusterId:  clusterFilter,
				Project:    projectsFilter,
				NodePool:   nodePoolFilter,
				Department: departmentsFilter,
			},
		}
	case TimerangeMetricMemoryQuota:
		metric = runaimetrics.RunaiDeservedMemoryMbPerQueuesMetric{
			Filters: runaimetrics.RunaiDeservedMemoryMbPerQueuesFilters{
				ClusterId:      clusterFilter,
				Project:        projectsFilter,
				NodePool:       nodePoolFilter,
				Department:     departmentsFilter,
				DepartmentName: departmentsFilter,
			},
		}
	case TimerangeMetricMemoryUtilization:
		metric = runaimetrics.RunaiMemoryUtilizationPerNodepoolsMetric{
			Filters: runaimetrics.RunaiMemoryUtilizationPerNodepoolsFilters{
				ClusterId: clusterFilter,
				NodePool:  nodePoolFilter,
			},
		}
	default:
		return nil, fmt.Errorf("metric %s not supported", metricType)
	}
	return metric, nil
}

func getCurrentMetricsResults(ctx context.Context, clusterId string, nodePools []string, departments []string, projects []string, metricType string) ([]metrics_api.Result, error) {
	clusterFilter := []string{clusterId}
	projectsFilter := projects[:]
	departmentsFilter := departments[:]
	nodePoolFilter := nodePools[:]
	metric, err := getMetric(metricType, clusterFilter, departmentsFilter, projectsFilter, nodePoolFilter)
	if err != nil {
		return nil, err
	}
	query, err := metrics_client.NewMetricQuery(metric)
	if err != nil {
		return nil, err
	}
	res, err := query.Query(ctx)
	return res.GetData(), err
}

func GetQuotaMetricResults(ctx context.Context, clusterId string, projects []string) ([]metrics_api.Result, error) {
	clusterFilter := []string{clusterId}
	projectsFilter := projects[:]
	metric := runaimetrics.RunaiAllocatedGpuCountPerProjectMetric{
		Filters: runaimetrics.RunaiAllocatedGpuCountPerProjectFilters{
			ClusterId: clusterFilter,
			Project:   projectsFilter,
		},
	}
	query, err := metrics_client.NewMetricQuery(metric)
	if err != nil {
		return nil, err
	}
	res, err := query.Query(ctx)
	return res.GetData(), err
}

func getTimerangeMetricsResults(
	ctx context.Context,
	clusterId string,
	nodePools []string,
	departments []string,
	projects []string,
	metricType string,
	start time.Time,
	end time.Time,
	numberOfSamples int32,
) ([]metrics_api.Result, error) {
	clusterFilter := []string{clusterId}
	projectsFilter := projects[:]
	nodePoolFilter := nodePools[:]
	departmentsFilter := departments[:]
	metric, err := getMetric(metricType, clusterFilter, departmentsFilter, projectsFilter, nodePoolFilter)
	if err != nil {
		return nil, err
	}
	query, err := metrics_client.NewMetricQuery(metric)
	if err != nil {
		return nil, err
	}
	query.SetRange(start, end)
	query.SetDefaultNumOfSamples(numberOfSamples)
	res, err := query.Query(ctx)
	if err != nil {
		return nil, err
	}

	return res.GetData(), nil
}

type CurrentMetricMeasurement struct {
	MetricName   string
	Measurements []metrics_api.Measurement
}

func handleSingleProjectCurrentMetric(result metrics_api.Result, metricName string, project string, nodePool string, projAggregationResponse map[string]map[string][]CurrentMetricMeasurement) {
	projResponse, exists := projAggregationResponse[project]
	if !exists {
		projResponse = make(map[string][]CurrentMetricMeasurement)
	}
	measurements, exists := projResponse[nodePool]
	if !exists {
		measurements = []CurrentMetricMeasurement{}
	}

	// Deep copy the Measurements slice to avoid reference sharing with department aggregation
	measurementsCopy := make([]metrics_api.Measurement, len(result.Measurements))
	copy(measurementsCopy, result.Measurements)

	measurements = append(measurements, CurrentMetricMeasurement{
		MetricName:   metricName,
		Measurements: measurementsCopy,
	})
	projResponse[nodePool] = measurements
	projAggregationResponse[project] = projResponse
}

func handleSingleDepartmentCurrentMetric(result metrics_api.Result, metricName string, department string, nodePool string, departmentAggregationResponse map[string]map[string][]CurrentMetricMeasurement) {
	departmentResponse, exists := departmentAggregationResponse[department]
	if !exists {
		departmentResponse = make(map[string][]CurrentMetricMeasurement)
	}
	measurements, exists := departmentResponse[nodePool]
	if !exists {
		measurements = []CurrentMetricMeasurement{}
	}
	measurementExists := false
	for i := range measurements {
		if measurements[i].MetricName == metricName {
			measurementExists = true
			oldVal, _ := strconv.ParseFloat(measurements[i].Measurements[0].Value, 32)
			newVal, _ := strconv.ParseFloat(result.Measurements[0].Value, 32)
			newVal = newVal + oldVal
			if metricName == CurrentMetricPendingWorkloads {
				measurements[i].Measurements[0].Value = fmt.Sprint(int32(newVal))
			} else {
				measurements[i].Measurements[0].Value = fmt.Sprint(newVal)
			}
		}
	}
	if !measurementExists {
		// Deep copy the Measurements slice to avoid reference sharing
		measurementsCopy := make([]metrics_api.Measurement, len(result.Measurements))
		copy(measurementsCopy, result.Measurements)

		currentMetricMeasurement := CurrentMetricMeasurement{
			MetricName:   metricName,
			Measurements: measurementsCopy,
		}
		measurements = append(measurements, currentMetricMeasurement)
	}
	departmentResponse[nodePool] = measurements
	departmentAggregationResponse[department] = departmentResponse
}

// GetCurrentMetrics returns map of project to nodepool to metrics measurements
func GetCurrentMetrics(
	ctx context.Context,
	clusterId string,
	nodePools []string,
	departments []string,
	projects []string,
	projectsAggregationOnly bool,
) (map[string]map[string][]CurrentMetricMeasurement, map[string]map[string][]CurrentMetricMeasurement, error) {
	departmentsAggregationResponse := make(map[string]map[string][]CurrentMetricMeasurement)
	projAggregationResponse := make(map[string]map[string][]CurrentMetricMeasurement)

	metricNames := []string{CurrentMetricPendingWorkloads, CurrentMetricAllocatedGpu, CurrentMetricAllocatedCpu, CurrentMetricAllocatedMemory}
	for _, metricName := range metricNames {
		metricResults, err := getCurrentMetricsResults(ctx, clusterId, nodePools, departments, projects, metricName)
		if err != nil {
			return nil, nil, err
		}
		for _, result := range metricResults {
			department := result.GetMetricLabels()[departmentLabel]
			if department == "" {
				department = result.GetMetricLabels()[oldDepartmentLabel]
			}
			project := result.GetMetricLabels()[projectLabel]
			nodePool := result.GetMetricLabels()[nodePoolLabel]
			if nodePool == "" || project == "" || department == "" {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("warning: failed to get project/department metrics (projectsAggregationOnly: %t). "+
					"metric '%s', nodepool label '%s', project label '%s' or department label '%s' are invalid.",
					projectsAggregationOnly, metricName, nodePool, project, department)
				continue
			}

			handleSingleProjectCurrentMetric(result, metricName, project, nodePool, projAggregationResponse)
			if !projectsAggregationOnly {
				handleSingleDepartmentCurrentMetric(result, metricName, department, nodePool, departmentsAggregationResponse)
			}
		}
	}
	if projectsAggregationOnly {
		return nil, projAggregationResponse, nil
	}
	return departmentsAggregationResponse, projAggregationResponse, nil
}

func GetTimeRangeMetrics(
	ctx context.Context, clusterId string, nodePools []string, departments []string, projects *[]string, start time.Time, end time.Time, numberOfSamples int32, fetchUtilization bool,
) (map[string][]metrics_api.Result, error) {
	result := make(map[string][]metrics_api.Result)
	if start.IsZero() || end.IsZero() {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Cannot get timeranges metrics, start/end time are empty")
		return result, nil
	}
	metricNames := []string{
		TimerangeMetricPendingWorkloads,
		TimerangeMetricAllocatedGpu,
		TimerangeMetricAllocatedCpu,
		TimerangeMetricAllocatedMemory,
		TimerangeMetricGpuQuota,
		TimerangeMetricCpuQuota,
		TimerangeMetricMemoryQuota,
	}
	if fetchUtilization {
		metricNames = append(metricNames, TimerangeMetricGpuUtilization, TimerangeMetricCpuUtilization, TimerangeMetricMemoryUtilization)
	}

	projectsFilter := []string{}
	if projects != nil {
		projectsFilter = *projects
	}

	for _, metricName := range metricNames {
		if projects == nil {
			projectsFilter = getProjectsFilterForDepartmentMetrics(metricName)
		}

		metricResults, err := getTimerangeMetricsResults(ctx, clusterId, nodePools, departments, projectsFilter, metricName, start, end, numberOfSamples)
		if err != nil {
			return nil, err
		}
		result[metricName] = metricResults
	}
	return result, nil
}

func getProjectsFilterForDepartmentMetrics(metricName string) []string {
	switch metricName {
	case TimerangeMetricGpuQuota, TimerangeMetricCpuQuota, TimerangeMetricMemoryQuota:
		// for the department quota we want project="" - to only take into account the department quota
		return []string{""}
	default:
		// otherwise - metric will be filtered by department names, project can be anything
		return []string{".*"}
	}
}
