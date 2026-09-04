/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/settings"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
)

const (
	maxAllowedMetricsInRequest = 10
	gpuMetricLabel             = "gpu"
)

var (
	groupByToLabel = map[string]string{
		string(client.GroupByCategory):   "workload_category",
		string(client.GroupByProject):    "project",
		string(client.GroupByDepartment): "department",
		string(client.GroupByNodePool):   "nodepool",
		string(client.GroupByNode):       "node",
		string(client.GroupByCluster):    "clusterId",
	}
	allowedResponseLabelsForPodMetricTypes = map[string]map[string]bool{
		string(api.NODEMETRICTYPE_GPU_UTILIZATION_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_GPU_MEMORY_UTILIZATION_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_GPU_MEMORY_USAGE_BYTES_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_GPU_ALLOCATION_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_GPU_MEMORY_ALLOCATION_BYTES_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_GPU_IDLE_TIME_SECONDS_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_GPU_OOMKILL_BURST_COUNT_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_GPU_OOMKILL_IDLE_COUNT_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_GPU_OOMKILL_SWAP_OUT_OF_RAM_COUNT_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_GPU_GRAPHICS_ENGINE_ACTIVITY_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_GPU_SM_ACTIVITY_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_GPU_SM_OCCUPANCY_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_GPU_TENSOR_ACTIVITY_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_GPU_FP64_ENGINE_ACTIVITY_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_GPU_FP32_ENGINE_ACTIVITY_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_GPU_FP16_ENGINE_ACTIVITY_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_GPU_MEMORY_BANDWIDTH_UTILIZATION_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_GPU_NVLINK_TRANSMITTED_BANDWIDTH_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_GPU_NVLINK_RECEIVED_BANDWIDTH_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_GPU_PCIE_TRANSMITTED_BANDWIDTH_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_GPU_PCIE_RECEIVED_BANDWIDTH_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.NODEMETRICTYPE_NVLINK_BANDWIDTH_TOTAL_PER_GPU): {
			gpuMetricLabel: true,
		},
	}

	advancedGpuMetrics = map[api.NodeMetricType]bool{
		api.NODEMETRICTYPE_GPU_GRAPHICS_ENGINE_ACTIVITY_PER_GPU:     true,
		api.NODEMETRICTYPE_GPU_SM_ACTIVITY_PER_GPU:                  true,
		api.NODEMETRICTYPE_GPU_SM_OCCUPANCY_PER_GPU:                 true,
		api.NODEMETRICTYPE_GPU_TENSOR_ACTIVITY_PER_GPU:              true,
		api.NODEMETRICTYPE_GPU_FP64_ENGINE_ACTIVITY_PER_GPU:         true,
		api.NODEMETRICTYPE_GPU_FP32_ENGINE_ACTIVITY_PER_GPU:         true,
		api.NODEMETRICTYPE_GPU_FP16_ENGINE_ACTIVITY_PER_GPU:         true,
		api.NODEMETRICTYPE_GPU_MEMORY_BANDWIDTH_UTILIZATION_PER_GPU: true,
		api.NODEMETRICTYPE_GPU_NVLINK_TRANSMITTED_BANDWIDTH_PER_GPU: true,
		api.NODEMETRICTYPE_GPU_NVLINK_RECEIVED_BANDWIDTH_PER_GPU:    true,
		api.NODEMETRICTYPE_GPU_PCIE_TRANSMITTED_BANDWIDTH_PER_GPU:   true,
		api.NODEMETRICTYPE_GPU_PCIE_RECEIVED_BANDWIDTH_PER_GPU:      true,
	}
)

type ListResultMetric struct {
	indexLabel          string
	numOfItems          int
	indexToDisplayLabel map[string]string
}

var listResultMetrics = map[string]ListResultMetric{
	string(api.METRICSTYPE_GPU_UTILIZATION_DISTRIBUTION): {
		indexLabel: "range",
		numOfItems: 4,
		indexToDisplayLabel: map[string]string{
			"0": "0%-25%",
			"1": "25%-50%",
			"2": "50%-75%",
			"3": "75%-100%",
		},
	},
}

type AdvancedGpuMetricsSettingCacheValue struct {
	IsOn bool
}

func ParseMetricResultToMeasurement(ctx context.Context, metricType string, response openapi.GetMetrics200Response) (api.MeasurementResponse, error) {
	listResultMetric, found := listResultMetrics[metricType]
	if found {
		return parseMetricResultToMeasurementList(ctx, metricType, listResultMetric, response)
	}

	measurement := api.MeasurementResponse{
		Type:   metricType,
		Values: &[]api.GroupedMeasurementResponseValuesInner{},
	}

	results := response.GetData()
	if len(results) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("No result return for metric type %s", metricType)
		return measurement, nil
	}

	measurement.Values = parseMetricSingleResultToMeasurement(results[0])

	return measurement, nil
}

func ParseMetricResultToGroupedMeasurement(ctx context.Context, metricType, groupBy string, response openapi.GetMetrics200Response) ([]api.GroupedMeasurementResponse, error) {
	measurements := []api.GroupedMeasurementResponse{}
	listResultMetric, found := listResultMetrics[metricType]
	if found {
		measurementsResponse, err := parseMetricResultToMeasurementList(ctx, metricType, listResultMetric, response)
		if err != nil {
			return nil, err
		}

		measurements = append(measurements, api.GroupedMeasurementResponse{
			Type:   measurementsResponse.Type,
			Values: measurementsResponse.Values,
		})

		return measurements, nil
	}

	results := response.GetData()
	if len(results) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("No result return for metric type %s", metricType)
		return measurements, nil
	}
	for _, result := range results {
		measurement := api.GroupedMeasurementResponse{
			Type:   metricType,
			Values: parseMetricSingleResultToMeasurement(result),
			Groups: parseGroupByValue(groupBy, result),
		}
		measurements = append(measurements, measurement)
	}

	return measurements, nil
}

func ParseNodeMetricResultToMeasurements(ctx context.Context, metricType string, response openapi.GetMetrics200Response) []api.MeasurementResponse {
	measurement := api.MeasurementResponse{
		Type:   metricType,
		Values: &[]api.GroupedMeasurementResponseValuesInner{},
	}

	results := response.GetData()
	if len(results) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("No result return for metric type %s", metricType)
		return []api.MeasurementResponse{measurement}
	}

	measurementsResponse := []api.MeasurementResponse{}
	for _, result := range results {
		measurementsResponse = append(measurementsResponse, api.MeasurementResponse{
			Type:   metricType,
			Values: parseMetricSingleResultToMeasurement(result),
			Labels: getLabelsForMetricResult(result, metricType, allowedResponseLabelsForPodMetricTypes),
		})
	}
	return measurementsResponse
}

// parseMetricResultToMeasurementList parse results with list (distribution) listLabelIndex is the label indicating the index of the value, numOfItems the expected length of the values
func parseMetricResultToMeasurementList(ctx context.Context, metricType string, listResultMetric ListResultMetric, response openapi.GetMetrics200Response) (api.MeasurementResponse, error) {
	measurement := api.MeasurementResponse{
		Type:   metricType,
		Values: &[]api.GroupedMeasurementResponseValuesInner{},
	}

	results := response.GetData()
	if len(results) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("No result return for metric type %s", metricType)
		return measurement, nil
	}

	resultList := make(map[time.Time][]string)
	for _, result := range results {
		label, found := result.MetricLabels[listResultMetric.indexLabel]
		if !found {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("No label in result of metric %s", metricType)
		}
		index, err := strconv.Atoi(label)
		if err != nil {
			return measurement, err
		}
		if index >= listResultMetric.numOfItems {
			return measurement, errors.New("index is larger than number of expected items")
		}
		for _, m := range result.Measurements {
			timestamp := m.Timestamp
			list, found := resultList[timestamp]
			if !found {
				list = make([]string, listResultMetric.numOfItems)
				resultList[timestamp] = list
			}
			displayLabel := listResultMetric.indexToDisplayLabel[label]
			list[index] = fmt.Sprintf("{\"%s\":\"%s\"}", displayLabel, m.Value)
		}
	}

	for timestamp, valueList := range resultList {
		*measurement.Values = append(*measurement.Values, api.GroupedMeasurementResponseValuesInner{
			Value:     "[" + strings.Join(valueList, ",") + "]",
			Timestamp: convutils.Ref(timestamp),
		})
	}

	return measurement, nil
}

func ParseNodeMetricResultToTelemetryResponse(ctx context.Context, nodeTelemetryType api.NodeTelemetryType,
	response openapi.GetMetrics200Response, timestamp time.Time, groupBy []string, modelFieldToMetricLabelMap map[string]string,
) *api.TelemetryResponse {
	telemetryResponse := &api.TelemetryResponse{
		Type:      string(nodeTelemetryType),
		Timestamp: timestamp,
		Values:    []api.TelemetryResponseValuesInner{},
	}

	results := response.GetData()
	if len(results) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("No result return for node telemetry metric type %s", string(nodeTelemetryType))
		return telemetryResponse
	}

	var telemetryGroups []api.TelemetryResponseValuesInner
	for _, result := range results {
		measurements, ok := result.GetMeasurementsOk()
		if !ok {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Cannot parse result of metric type %s to measurements", string(nodeTelemetryType))
			continue
		}
		if len(measurements) != 1 {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Measurements of metric type %s has length that is different than 1 (%d): %+v", string(nodeTelemetryType), len(measurements), measurements)
			continue
		}
		sampleValue, ok := measurements[0].GetValueOk()
		if !ok {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Cannot parse measurement of metric type %s to sample value %+v", string(nodeTelemetryType), measurements)
			continue
		}

		metricLabels, ok := result.GetMetricLabelsOk()
		if !ok {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Cannot parse result of metric type %s to metric labels %+v", string(nodeTelemetryType), result)
			continue
		}

		telemetryGroups = append(telemetryGroups, api.TelemetryResponseValuesInner{
			Value:  *sampleValue,
			Groups: MetricValuesToTelemetryGroups(ctx, nodeTelemetryType, *metricLabels, groupBy, modelFieldToMetricLabelMap),
		})
	}

	telemetryResponse.Values = telemetryGroups
	return telemetryResponse
}

func parseMetricSingleResultToMeasurement(result openapi.Result) *[]api.GroupedMeasurementResponseValuesInner {
	values := &[]api.GroupedMeasurementResponseValuesInner{}
	for _, m := range result.Measurements {
		*values = append(*values, api.GroupedMeasurementResponseValuesInner{
			Value:     m.Value,
			Timestamp: convutils.Ref(m.Timestamp),
		})
	}

	return values
}

func parseGroupByValue(groupBy string, result openapi.Result) []api.GroupedMeasurementGroup {
	if groupBy == "" {
		return []api.GroupedMeasurementGroup{}
	}

	labelName := groupByToLabel[groupBy]
	if labelName == "" {
		return []api.GroupedMeasurementGroup{}
	}

	groupValue, ok := result.MetricLabels[labelName]
	if !ok || groupValue == "" {
		return []api.GroupedMeasurementGroup{}
	}
	return []api.GroupedMeasurementGroup{{
		Value: groupValue,
		Key:   groupBy,
	}}
}

func getLabelsForMetricResult(result openapi.Result, metricType string, allowedResponseLabelsForMetricTypes map[string]map[string]bool) *map[string]string {
	allowedResponseLabels, found := allowedResponseLabelsForMetricTypes[metricType]
	if !found {
		return nil
	}

	metricResponseLabels := result.MetricLabels
	for label := range metricResponseLabels {
		if _, found = allowedResponseLabels[label]; !found {
			delete(metricResponseLabels, label)
		}
	}
	return &metricResponseLabels
}

func MetricValuesToTelemetryGroups(ctx context.Context, nodeTelemetryType api.NodeTelemetryType,
	metricLabels map[string]string, groupBy []string, groupByKeyToMetricLabel map[string]string,
) []api.TelemetryResponseValuesInnerGroupsInner {
	var apiGroups []api.TelemetryResponseValuesInnerGroupsInner
	for _, group := range groupBy {
		apiGroup := MetricValueToTelemetryGroup(metricLabels, group, groupByKeyToMetricLabel)
		if apiGroup == nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to convert metric labels of telemetry %s to telemetry label value, group: %s, metric labels: %+v",
				string(nodeTelemetryType), group, metricLabels)
			continue
		}
		apiGroups = append(apiGroups, *apiGroup)
	}

	return apiGroups
}

func MetricValueToTelemetryGroup(metricLabels map[string]string, group string, groupByKeyToMetricLabel map[string]string) *api.TelemetryResponseValuesInnerGroupsInner {
	metricLabel := groupByKeyToMetricLabel[group]

	metricValue, ok := metricLabels[metricLabel]
	if !ok {
		return nil
	}

	return &api.TelemetryResponseValuesInnerGroupsInner{
		Key:   group,
		Value: metricValue,
	}
}

func ValidateRequestedMetrics(metricTypesLen int) *api.Error {
	if metricTypesLen > maxAllowedMetricsInRequest {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("Too many metrics requested. Max allowed is %d", maxAllowedMetricsInRequest), nil)
	}

	return nil
}

func ValidateRequestedAdvancedMetrics(ctx context.Context, metricTypes []api.NodeMetricType,
) *api.Error {
	if !WasAdvancedMetricRequested(metricTypes) {
		return nil
	}

	if !IsAdvancedGpuMetricsFFOn(ctx) {
		return api.NewError(http.StatusBadRequest, "Advanced GPU metrics are not enabled", nil)
	}

	return nil
}

func WasAdvancedMetricRequested(metricTypes []api.NodeMetricType) bool {
	for _, metricType := range metricTypes {
		if _, found := advancedGpuMetrics[metricType]; found {
			return true
		}
	}

	return false
}

func IsAdvancedGpuMetricsFFOn(ctx context.Context) bool {
	tenantId := runaictx.GetString(ctx, runaictx.CfTenantId)
	if tenantId == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("missing tenant id in context")
		return false
	}
	s, err := setting_pkg.Get[*settings.EnableAdvancedGPUMetrics](*tenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get EnableAdvancedGPUMetrics setting: %v", err)
		return false
	}
	return s.Enabled
}
