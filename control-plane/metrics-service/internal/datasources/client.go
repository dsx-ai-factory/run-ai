/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datasources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/connections"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/services"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/tracing"
)

var (
	queryEndpoint      = "/api/v1/query"
	queryRangeEndpoint = "/api/v1/query_range"
	queryParam         = "query"
	startParam         = "start"
	endParam           = "end"
	timeParam          = "time"
	stepParam          = "step"
)

type Client struct {
	connectionManager ConnectionGetter
	client            *resty.Client
}

func NewClient(connectionManager ConnectionGetter) *Client {
	return &Client{
		connectionManager: connectionManager,
		client:            resty.New().SetHeader("Content-Type", "application/json"),
	}
}

func (dc *Client) Query(ctx context.Context, tenantId int32, query string, timeRange *services.MetricInformationRange) (*api.GetMetrics200Response, error) {
	ctx, span := tracing.GetTracer().StartSpan(ctx, "Query")
	defer span.End()
	span.SetAttribute("query", query)
	connection, err := dc.connectionManager.GetConnection(ctx, tenantId)
	if err != nil {
		return nil, handleErrorWithTracing(span, fmt.Errorf("failed to get connection from connection manager: %v", err))
	}

	fullUrl := dc.buildFullUrl(connection, timeRange)
	span.SetAttribute("fullUrl", fullUrl)
	queryParams, err := buildQueryParams(query, timeRange)
	if err != nil {
		return nil, handleErrorWithTracing(span, fmt.Errorf("failed to build query params: %v", err))
	}
	span.SetAttribute("queryParams", fmt.Sprintf("%v", queryParams))
	resp, err := connection.AuthConfigurator.Configure(dc.client.R()).SetHeaders(connection.Headers).SetQueryParams(queryParams).Get(fullUrl)
	if err != nil {
		return nil, handleErrorWithTracing(span, fmt.Errorf("failed to query datasource %s, res: %+v: %v", fullUrl, resp, err))
	}
	span.SetAttribute("status_code", fmt.Sprintf("%v", resp.StatusCode()))
	if resp.StatusCode() != 200 {
		return nil, handleErrorWithTracing(span, fmt.Errorf("failed to query datasource %s, status code: %v, body: %v", fullUrl, resp.StatusCode(), resp.String()))
	}
	metricsResponse, err := buildResponse(resp.Body(), query)
	if err != nil {
		return nil, handleErrorWithTracing(span, fmt.Errorf("failed to build metrics response: %v", err))
	}
	return metricsResponse, nil
}

func (dc *Client) buildFullUrl(conn *connections.Connection, timeRange *services.MetricInformationRange) string {
	conf := config.Get[*DatasourcesConfiguration]()

	if isRangeQuery(timeRange) && !timeRange.Aggregation {
		return conn.UrlOrDefault(conf.DatasourceUrl) + queryRangeEndpoint
	}
	return conn.UrlOrDefault(conf.DatasourceUrl) + queryEndpoint
}

func buildQueryParams(query string, timeRange *services.MetricInformationRange) (map[string]string, error) {
	params := map[string]string{
		queryParam: query,
	}

	err := addAggregationQueryParams(timeRange, params)
	if err != nil {
		return params, err
	}

	err = addRangeQueryParams(timeRange, params)
	if err != nil {
		return params, err
	}

	return params, nil
}

func addAggregationQueryParams(timeRange *services.MetricInformationRange, params map[string]string) error {
	if !isRangeQuery(timeRange) || !timeRange.Aggregation {
		return nil
	}

	params[timeParam] = fmt.Sprintf("%v", timeRange.End.Unix())
	return nil
}

func addRangeQueryParams(timeRange *services.MetricInformationRange, params map[string]string) error {
	if !isRangeQuery(timeRange) || timeRange.Aggregation {
		return nil
	}

	params[startParam] = fmt.Sprintf("%v", timeRange.Start.Unix())
	params[endParam] = fmt.Sprintf("%v", timeRange.End.Unix())
	step, err := calculateStepSizeBetweenSamples(timeRange.Start, timeRange.End, timeRange.NumOfSamples)
	if err != nil {
		return fmt.Errorf("failed to calculate step: %v", err)
	}

	params[stepParam] = fmt.Sprint(step)
	return nil
}

func isRangeQuery(timeRange *services.MetricInformationRange) bool {
	return timeRange != nil
}

func calculateStepSizeBetweenSamples(start time.Time, end time.Time, numOfSamples int32) (float64, error) {
	if numOfSamples <= 0 {
		return 0, errors.New("number of samples must be greater than 0")
	}
	if start.IsZero() {
		return 0, errors.New("start time must be provided")
	}
	if end.IsZero() {
		return 0, errors.New("end time must be provided")
	}
	timeDiff := end.Sub(start).Seconds()
	if timeDiff <= 0 {
		return 0, errors.New("end time must be greater than start time")
	}
	// we reduce 1 from num of samples, since when asking for n samples - we get n+1
	return timeDiff / math.Max(1.0, float64(numOfSamples-1)), nil
}

func buildResponse(body []byte, query string) (*api.GetMetrics200Response, error) {
	res := DatasourceResponse{}
	err := json.Unmarshal(body, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to parse datasource response: %v", err)
	}
	return ApiResponseFromDatasourceResponse(res, query), nil
}

func handleErrorWithTracing(span tracing.Span, err error) error {
	span.RecordError(err)
	span.SetStatus(tracing.SpanStatusError, err.Error())
	return err
}
