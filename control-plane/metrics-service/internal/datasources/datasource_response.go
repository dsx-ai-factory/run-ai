/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datasources

import (
	"time"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
)

type DatasourceResponse struct {
	Status string `json:"status"`
	Data   Data   `json:"data"`
}

type Data struct {
	ResultType string   `json:"resultType"`
	Result     []Result `json:"result"`
}

type Result struct {
	Metric map[string]string `json:"metric"`
	Values [][]interface{}   `json:"values"`
	Value  []interface{}     `json:"value"`
}

func ApiResponseFromDatasourceResponse(datasourceResponse DatasourceResponse, query string) *api.GetMetrics200Response {
	apiResponse := &api.GetMetrics200Response{
		Status:  datasourceResponse.Status,
		Query:   query,
		IsRange: datasourceResponse.Data.ResultType == "matrix",
	}
	for _, result := range datasourceResponse.Data.Result {
		apiResponse.Data = append(apiResponse.Data, api.Result{
			MetricLabels: result.Metric,
			Measurements: buildMeasurements(result, apiResponse.IsRange),
		})
	}
	return apiResponse
}

func buildMeasurements(r Result, isRange bool) []api.Measurement {
	if isRange {
		return buildRangeMeasurements(r)
	}

	return buildNowMeasurements(r)
}

func buildRangeMeasurements(r Result) []api.Measurement {
	values := make([]api.Measurement, 0, len(r.Values))
	for _, value := range r.Values {
		values = append(values, api.Measurement{
			Timestamp: time.UnixMilli(int64(value[0].(float64) * 1000)),
			Value:     value[1].(string),
		})
	}
	return values
}

func buildNowMeasurements(r Result) []api.Measurement {
	return []api.Measurement{
		{
			Timestamp: time.UnixMilli(int64(r.Value[0].(float64) * 1000)),
			Value:     r.Value[1].(string),
		},
	}
}
