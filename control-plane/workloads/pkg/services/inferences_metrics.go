/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
)

func GetInferenceWorkloadMetrics(
	ctx context.Context,
	workloadsDao *dao.WorkloadsDao,
	workloadId string,
	metricTypes []api.InferenceWorkloadMetricType,
	start time.Time,
	end time.Time,
	numberOfSamples int32,
) (api.ImplResponse, error) {
	apiErr := validateRequestedMetrics(len(metricTypes))
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	clusterId, dbErr := workloadsDao.DbGetWorkloadClusterId(ctx, tenantId, workloadId)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	apiErr = validateRequestedInferenceWorkloadMetrics(ctx, clusterId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	measurements := []api.MeasurementResponse{}
	responseCode := http.StatusOK

	for _, metricType := range metricTypes {
		result := queryInferenceWorkloadMetricType(ctx, workloadId, clusterId, start, end, metricType, numberOfSamples)
		if result == nil {
			responseCode = http.StatusMultiStatus
			measurements = append(measurements, api.MeasurementResponse{
				Type: string(metricType),
			})

			continue
		}

		measurement := parseWorkloadMetricResultToMeasurement(ctx, string(metricType), *result)
		measurements = append(measurements, measurement)
	}

	responseStruct := api.MetricsResponse{Measurements: measurements}
	accept := runaictx.GetString(ctx, runaictx.CfAccept)
	if accept != nil && *accept == CsvAccept {
		var types []string
		for _, metricType := range metricTypes {
			types = append(types, string(metricType))
		}
		data := GetMetricsCsv(&responseStruct, types)
		response := runaihttp.CsvFile{Data: data, FileName: "workload-metrics.csv"}
		return api.Response(http.StatusOK, response), nil
	}

	return api.Response(responseCode, responseStruct), nil
}

func GetInferenceWorkloadPodMetrics(
	ctx context.Context,
	workloadId string,
	podId string,
	metricTypes []api.InferencePodMetricType,
	start time.Time,
	end time.Time,
	numberOfSamples int32,
) (api.ImplResponse, error) {
	apiErr := validateRequestedMetrics(len(metricTypes))
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	clusterId, dbErr := dao.DbGetPodClusterId(ctx, tenantId, workloadId, podId)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	apiErr = validateRequestedInferenceWorkloadMetrics(ctx, clusterId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	measurements := []api.MeasurementResponse{}
	responseCode := http.StatusOK

	for _, metricType := range metricTypes {
		result := queryInferencePodMetricType(ctx, podId, clusterId, start, end, metricType, numberOfSamples)
		if result == nil {
			responseCode = http.StatusMultiStatus
			measurements = append(measurements, api.MeasurementResponse{
				Type: string(metricType),
			})

			continue
		}

		measurement := parsePodMetricResultToMeasurements(ctx, string(metricType), *result)
		measurements = append(measurements, measurement...)
	}

	responseStruct := api.MetricsResponse{Measurements: measurements}
	accept := runaictx.GetString(ctx, runaictx.CfAccept)
	if accept != nil && *accept == CsvAccept {
		var types []string
		for _, metricType := range metricTypes {
			types = append(types, string(metricType))
		}
		data := GetMetricsCsv(&responseStruct, types)
		response := runaihttp.CsvFile{Data: data, FileName: "pods-metrics.csv"}
		return api.Response(http.StatusOK, response), nil
	}
	return api.Response(responseCode, responseStruct), nil
}
