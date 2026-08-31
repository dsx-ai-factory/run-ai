/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"k8s.io/utils/ptr"

	metrics_api "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services"
)

const (
	AffinityTypeNoLimit      = "no_limit"
	AffinityTypeOnlySelected = "only_selected"
	DefaultOverQuotaWeight   = float64(2)
	MaxTimestampsDiff        = 5 * 1000 // 5 seconds
	UnlimitedValue           = float32(-1)
)

// RawStatus - this struct is used to unmarshal the status sent as a json string from the cluster-sync
type RawStatus struct {
	Namespace  string             `json:"namespace,omitempty" protobuf:"bytes,3,opt,name=namespace"`
	Phase      string             `json:"phase,omitempty" protobuf:"bytes,3,opt,name=phase,casttype=ProjectPhase"`
	Message    string             `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`
	Conditions []models.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

func MapSingleNodePool(ctx context.Context, id int32, nodePoolId float32, nodePoolModelResource models.Resources, conversionFn services.MemConvertFunc, cpuResourcesFlag bool) api.NodePoolAssignedResourcesV1Response {
	var unlimitedLimitValue *float32 = nil
	if cpuResourcesFlag {
		unlimitedLimitValue = ptr.To(UnlimitedValue)
	}
	return api.NodePoolAssignedResourcesV1Response{
		Id: float32(id),
		NodePool: api.NodePoolAssignedResourcesV1ResponseNodePool{
			Id:   nodePoolId,
			Name: nodePoolModelResource.NodePoolName,
		},
		Gpu: api.ResourceV1Response{
			Deserved:        ptrToFloat32(nodePoolModelResource.GpuDeserved, nil),
			MaxAllowed:      ptrToFloat32(nodePoolModelResource.GpuLimit, ptr.To(UnlimitedValue)),
			OverQuotaWeight: ptrToFloat32(nodePoolModelResource.GpuOverQuotaWeight, nil),
		},
		Cpu: api.ResourceV1Response{
			Deserved:        ptrToFloat32(nodePoolModelResource.CpuDeserved, nil),
			MaxAllowed:      ptrToFloat32(nodePoolModelResource.CpuLimit, unlimitedLimitValue),
			OverQuotaWeight: ptrToFloat32(nodePoolModelResource.CpuOverQuotaWeight, nil),
		},
		Memory: api.ResourceV1Response{
			Deserved:        ptrToFloat32(services.ApplyMemoryConversion(nodePoolModelResource.MemoryDeserved, conversionFn), nil),
			MaxAllowed:      ptrToFloat32(services.ApplyMemoryConversion(nodePoolModelResource.MemoryLimit, conversionFn), unlimitedLimitValue),
			OverQuotaWeight: ptrToFloat32(nodePoolModelResource.MemoryOverQuotaWeight, nil),
		},
	}
}

func ptrToFloat32(val *float64, defaultValue *float32) *float32 {
	if val == nil {
		return defaultValue
	}
	newVal := float32(*val)
	return ptr.To(newVal)
}

func getNodePoolId(strId string) (float32, error) {
	nodePoolId64, err := strconv.ParseFloat(strId, 32)
	if err != nil {
		return 0, err
	}
	nodePoolId := float32(nodePoolId64)
	if float64(nodePoolId) != nodePoolId64 {
		return 0, fmt.Errorf("failed to map department resources. nodepool id %f is not float32", nodePoolId64)
	}
	return nodePoolId, nil
}

func mapNodePoolToResources(resources []models.Resources) map[string]models.Resources {
	resourcesMap := make(map[string]models.Resources)
	for _, resource := range resources {
		resourcesMap[resource.NodePoolName] = resource
	}
	return resourcesMap
}

func mapCurrentResourceInner(resources models.Resources, nodePool string, currMetricResult []CurrentMetricMeasurement) (api.DepartmentMetricsV1CurrentResourcesInner, error) {
	resourcesInner := api.DepartmentMetricsV1CurrentResourcesInner{
		NodepoolName: nodePool,
	}
	nodePoolResources := resources
	for _, results := range currMetricResult {
		if len(results.Measurements) == 0 {
			continue
		}
		if results.MetricName == CurrentMetricPendingWorkloads {
			value, err := strconv.Atoi(results.Measurements[0].Value)
			if err != nil {
				return resourcesInner, err
			}
			resourcesInner.NumberOfPendingWorkloads = int32(value)
			continue
		}
		parsed, err := strconv.ParseFloat(results.Measurements[0].Value, 32)
		if err != nil {
			return resourcesInner, err
		}
		value := float32(parsed)
		if results.MetricName == CurrentMetricAllocatedGpu {
			quota := float32(0)
			if nodePoolResources.GpuDeserved != nil {
				quota = float32(*nodePoolResources.GpuDeserved)
			}
			resourcesInner.Gpu = api.MetricsResourceDataV1{
				Allocated: value,
				Quota:     quota,
			}
		}
		if results.MetricName == CurrentMetricAllocatedCpu {
			quota := float32(0)
			if nodePoolResources.CpuDeserved != nil {
				cpuDeserved := float32(*nodePoolResources.CpuDeserved)
				quota = cpuDeserved
			}
			resourcesInner.Cpu = api.MetricsResourceDataV1{
				Allocated: value,
				Quota:     quota,
			}
		}
		if results.MetricName == CurrentMetricAllocatedMemory {
			quota := float32(0)
			if nodePoolResources.MemoryDeserved != nil {
				memoryDeserved := float32(*nodePoolResources.MemoryDeserved)
				quota = memoryDeserved
			}
			resourcesInner.Memory = api.MetricsResourceDataV1{
				Allocated: value,
				Quota:     quota,
			}
		}
	}
	return resourcesInner, nil
}

func MapCurrentMetricsToApi(ctx context.Context, orgUnitData models.OrgUnitData, nodePoolsMetrics map[string][]CurrentMetricMeasurement) ([]api.DepartmentMetricsV1CurrentResourcesInner, error) {
	currentResources := []api.DepartmentMetricsV1CurrentResourcesInner{}
	resources := mapNodePoolToResources(orgUnitData.Resources)
	for nodePool, currMetricResult := range nodePoolsMetrics {
		resourcesInner, err := mapCurrentResourceInner(resources[nodePool], nodePool, currMetricResult)
		if err != nil {
			return currentResources, err
		}
		currentResources = append(currentResources, resourcesInner)
	}
	return currentResources, nil
}

func getShortestMetric(timeRangeMetrics map[string][]metrics_api.Result) []metrics_api.Result {
	if len(timeRangeMetrics) == 0 {
		return []metrics_api.Result{}
	}
	length := int(^uint(0) >> 1)
	var metric []metrics_api.Result
	for _, results := range timeRangeMetrics {
		if len(results) == 0 {
			continue
		}
		if results[0].Measurements == nil {
			length = 0
			metric = results
		} else if len(results[0].Measurements) < length {
			length = len(results[0].Measurements)
			metric = results
		}
	}
	return metric
}

func getTimerangeMeasurement(timeRangeMetrics map[string][]metrics_api.Result, metricName string, timerangeIndex int, currentTimestamp time.Time) (metrics_api.Measurement, error) {
	results, found := timeRangeMetrics[metricName]
	if !found {
		return metrics_api.Measurement{}, fmt.Errorf("no such metric %s", metricName)
	}
	if len(results) == 0 {
		return metrics_api.Measurement{
			Value: "0",
		}, nil
	}

	if len(results[0].Measurements) < timerangeIndex+1 {
		return metrics_api.Measurement{}, fmt.Errorf("no such result in index %d", timerangeIndex)
	}
	if math.Abs(float64(results[0].Measurements[timerangeIndex].Timestamp.UnixMilli()-currentTimestamp.UnixMilli())) <= MaxTimestampsDiff {
		return results[0].Measurements[timerangeIndex], nil
	}

	// Avoid going over all metric values since in most cases it will have the same index as the shortest metric
	for _, measurement := range results[0].Measurements {
		if math.Abs(float64(currentTimestamp.UnixMilli()-measurement.Timestamp.UnixMilli())) <= MaxTimestampsDiff {
			return measurement, nil
		}
	}

	return metrics_api.Measurement{}, fmt.Errorf("failed to match metrics timestamps for metric %s", metricName)
}

func getValueFromTimerangeMetric(timeRangeMetrics map[string][]metrics_api.Result, metricName string, timerangeIndex int, currentTimeStamp time.Time) (float32, error) {
	measurement, err := getTimerangeMeasurement(timeRangeMetrics, metricName, timerangeIndex, currentTimeStamp)
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseFloat(measurement.Value, 32)
	if err != nil {
		return 0, err
	}
	return float32(value), nil
}

func createResourceResults(
	timeRangeMetrics map[string][]metrics_api.Result,
	timerangeIndex int,
	quotaMetric string,
	allocationMetric string,
	utilizationMetric string,
	fetchUtilization bool,
	currentTimeStamp time.Time,
) (*api.MetricsResourceDataV1, error) {
	quota, err := getValueFromTimerangeMetric(timeRangeMetrics, quotaMetric, timerangeIndex, currentTimeStamp)
	if err != nil {
		return nil, err
	}
	allocation, err := getValueFromTimerangeMetric(timeRangeMetrics, allocationMetric, timerangeIndex, currentTimeStamp)
	if err != nil {
		return nil, err
	}
	data := &api.MetricsResourceDataV1{
		Quota:     quota,
		Allocated: allocation,
	}
	if fetchUtilization {
		utilization, err := getValueFromTimerangeMetric(timeRangeMetrics, utilizationMetric, timerangeIndex, currentTimeStamp)
		if err != nil {
			return nil, err
		}
		data.Utilization = utilization
	}
	return data, nil
}

func MapTimeRangeMetricsToApi(timeRangeMetrics map[string][]metrics_api.Result, fetchUtilization bool) ([]api.ResourcesDataOvertimeV1, error) {
	resources := []api.ResourcesDataOvertimeV1{}
	shortestMetric := getShortestMetric(timeRangeMetrics)
	if len(shortestMetric) == 0 || shortestMetric[0].Measurements == nil {
		return resources, nil
	}
	length := len(shortestMetric[0].Measurements)

	for i := 0; i < length; i++ {
		currentTimeStamp := shortestMetric[0].Measurements[i].Timestamp
		pendingWorkloads, err := getTimerangeMeasurement(timeRangeMetrics, TimerangeMetricPendingWorkloads, i, currentTimeStamp)
		if err != nil {
			return nil, err
		}
		pendingWorkloadsVal, err := strconv.ParseInt(pendingWorkloads.Value, 10, 32)
		if err != nil {
			return nil, err
		}
		gpuResults, err := createResourceResults(timeRangeMetrics, i, TimerangeMetricGpuQuota, TimerangeMetricAllocatedGpu, TimerangeMetricGpuUtilization, fetchUtilization, currentTimeStamp)
		if err != nil {
			return nil, err
		}
		// for backwards compatibility, old metrics used to return the data / 100
		gpuResults.Utilization = gpuResults.Utilization / 100
		cpuResults, err := createResourceResults(timeRangeMetrics, i, TimerangeMetricCpuQuota, TimerangeMetricAllocatedCpu, TimerangeMetricCpuUtilization, fetchUtilization, currentTimeStamp)
		if err != nil {
			return nil, err
		}
		// for backwards compatibility, we need to return, old metrics used to return the data * 1000
		cpuResults.Quota = cpuResults.Quota * 1000
		memoryResults, err := createResourceResults(timeRangeMetrics, i, TimerangeMetricMemoryQuota, TimerangeMetricAllocatedMemory, TimerangeMetricMemoryUtilization, fetchUtilization, currentTimeStamp)
		if err != nil {
			return nil, err
		}
		data := api.ResourcesDataOvertimeV1{
			NumberOfPendingWorkloads: int32(pendingWorkloadsVal),
			Gpu:                      *gpuResults,
			Cpu:                      *cpuResults,
			Memory:                   *memoryResults,
			Timestamp:                currentTimeStamp,
		}
		resources = append(resources, data)
	}
	return resources, nil
}
