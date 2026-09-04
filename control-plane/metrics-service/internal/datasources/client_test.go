/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datasources

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/connections"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/datasources/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/services"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func initDatasourcesClientForTesting() (testSubject *Client, connectionManager *mocks.ConnectionGetter) {
	connectionManager = &mocks.ConnectionGetter{}
	cfg := &DatasourcesConfiguration{
		DatasourceUrl: "url",
	}
	config.Set(cfg)
	testSubject = NewClient(connectionManager)
	httpmock.ActivateNonDefault(testSubject.client.GetClient())
	return
}

func TestCalculateStepSizeBetweenSamples_Success(t *testing.T) {
	now := time.Now()
	step, err := calculateStepSizeBetweenSamples(now.Add(-1*time.Hour), now, 10)
	assert.NoError(t, err)
	assert.Equal(t, 400.0, step)
}

func TestCalculateStepSizeBetweenSamples_FailEndBeforeStart(t *testing.T) {
	now := time.Now()
	_, err := calculateStepSizeBetweenSamples(now, now.Add(-1*time.Hour), 10)
	assert.Error(t, err)
}

func TestCalculateStepSizeBetweenSamples_FailNoSamples(t *testing.T) {
	now := time.Now()
	_, err := calculateStepSizeBetweenSamples(now.Add(-1*time.Hour), now, 0)
	assert.Error(t, err)
}

func TestCalculateStepSizeBetweenSamples_FailNoStart(t *testing.T) {
	now := time.Now()
	_, err := calculateStepSizeBetweenSamples(time.Time{}, now, 10)
	assert.Error(t, err)
}

func TestCalculateStepSizeBetweenSamples_FailNoEnd(t *testing.T) {
	now := time.Now()
	_, err := calculateStepSizeBetweenSamples(now, time.Time{}, 10)
	assert.Error(t, err)
}

func TestBuildFullUrl_Range(t *testing.T) {
	testSubject, _ := initDatasourcesClientForTesting()
	now := time.Now()
	connection := &connections.Connection{
		Url:              "",
		Headers:          map[string]string{"key": "value"},
		AuthConfigurator: &connections.EmptyAuthConfigurator{},
	}
	timeRange := &services.MetricInformationRange{
		Start:        now.Add(-1 * time.Hour),
		End:          now,
		NumOfSamples: 10,
	}
	fullUrl := testSubject.buildFullUrl(connection, timeRange)
	assert.Equal(t, "url"+queryRangeEndpoint, fullUrl)
}

func TestBuildFullUrl_Now(t *testing.T) {
	testSubject, _ := initDatasourcesClientForTesting()
	connection := &connections.Connection{
		Url:              "",
		Headers:          map[string]string{"key": "value"},
		AuthConfigurator: &connections.EmptyAuthConfigurator{},
	}
	timeRange := (*services.MetricInformationRange)(nil)
	fullUrl := testSubject.buildFullUrl(connection, timeRange)
	assert.Equal(t, "url"+queryEndpoint, fullUrl)
}

func TestBuildQueryParams_Range(t *testing.T) {
	now := time.Now()
	timeRange := &services.MetricInformationRange{
		Start:        now.Add(-1 * time.Hour),
		End:          now,
		NumOfSamples: 10,
	}
	params, err := buildQueryParams("query", timeRange)
	assert.NoError(t, err)
	assert.Equal(t, "query", params[queryParam])
	assert.Equal(t, fmt.Sprintf("%v", now.Add(-1*time.Hour).Unix()), params[startParam])
	assert.Equal(t, fmt.Sprintf("%v", now.Unix()), params[endParam])
	assert.Equal(t, "400", params[stepParam])
}

func TestBuildQueryParams_Aggregate(t *testing.T) {
	now := time.Now()
	timeRange := &services.MetricInformationRange{
		Start:       now.Add(-1 * time.Hour),
		End:         now,
		Aggregation: true,
	}
	params, err := buildQueryParams("query", timeRange)
	assert.NoError(t, err)
	assert.Equal(t, "query", params[queryParam])
	assert.Equal(t, fmt.Sprintf("%v", now.Unix()), params[timeParam])
}

func TestBuildQueryParams_Now(t *testing.T) {
	timeRange := (*services.MetricInformationRange)(nil)
	params, err := buildQueryParams("query", timeRange)
	assert.NoError(t, err)
	assert.Equal(t, "query", params[queryParam])
}

var body = []byte(`{
	"status": "success",
	"data": {
		"resultType": "matrix",
		"result": [
			{
				"metric": {
					"__name__": "runai_gpu_count_per_cluster",
					"clusterId": "5fe798df-33bf-4816-8cf7-146a84230d1a"
				},
				"values": [
					[
						1706439037,
						"32"
					],
					[
						1706439397,
						"32"
					]
				]
			},
			{
				"metric": {
					"__name__": "runai_gpu_count_per_cluster",
					"clusterId": "872c4fbf-f100-4633-aa7f-dc04bb454e31"
				},
				"values": [
					[
						1706439037,
						"32"
					],
					[
						1706439397,
						"32"
					]
				]
			}
		]
	}
}`)

func TestBuildRespose_Query(t *testing.T) {
	query := "runai_gpu_count_per_cluster"
	res, _ := buildResponse(body, query)
	assert.Equal(t, query, res.Query)
}

func TestBuildRespose_Status(t *testing.T) {
	query := "runai_gpu_count_per_cluster"
	res, _ := buildResponse(body, query)
	assert.Equal(t, "success", res.Status)
}

func TestBuildRespose_Range(t *testing.T) {
	query := "runai_gpu_count_per_cluster"
	res, _ := buildResponse(body, query)
	assert.True(t, res.IsRange)
}

func TestBuildRespose_Data(t *testing.T) {
	query := "runai_gpu_count_per_cluster"
	res, _ := buildResponse(body, query)
	assert.Equal(t, 2, len(res.Data))
	assert.Equal(t, "5fe798df-33bf-4816-8cf7-146a84230d1a", res.Data[0].MetricLabels["clusterId"])
	assert.Equal(t, time.UnixMilli(1706439037000), res.Data[0].Measurements[0].Timestamp)
	assert.Equal(t, "32", res.Data[0].Measurements[1].Value)
	assert.Equal(t, "872c4fbf-f100-4633-aa7f-dc04bb454e31", res.Data[1].MetricLabels["clusterId"])
	assert.Equal(t, "32", res.Data[1].Measurements[1].Value)
}

func TestQuery_Success(t *testing.T) {
	testSubject, connectionManager := initDatasourcesClientForTesting()
	connection := &connections.Connection{
		Url:              "",
		Headers:          map[string]string{"key": "value"},
		AuthConfigurator: &connections.EmptyAuthConfigurator{},
	}
	ctx := context.Background()
	connectionManager.On("GetConnection", ctx, int32(1)).Return(connection, nil)
	httpmock.RegisterResponder("GET", fmt.Sprintf("/url%s", queryEndpoint), httpmock.NewBytesResponder(200, body))
	res, err := testSubject.Query(ctx, int32(1), "query", nil)
	assert.NoError(t, err)
	assert.Equal(t, "query", res.Query)
	assert.Equal(t, 2, len(res.Data))
	assert.Equal(t, "5fe798df-33bf-4816-8cf7-146a84230d1a", res.Data[0].MetricLabels["clusterId"])
	assert.Equal(t, time.UnixMilli(1706439037000), res.Data[0].Measurements[0].Timestamp)
	assert.Equal(t, "32", res.Data[0].Measurements[1].Value)
	assert.Equal(t, "872c4fbf-f100-4633-aa7f-dc04bb454e31", res.Data[1].MetricLabels["clusterId"])
	assert.Equal(t, "32", res.Data[1].Measurements[1].Value)
}

func TestQuery_FailGetConnection(t *testing.T) {
	testSubject, connectionManager := initDatasourcesClientForTesting()
	ctx := context.Background()
	connectionManager.On("GetConnection", ctx, int32(1)).Return(nil, fmt.Errorf("failed"))
	res, err := testSubject.Query(ctx, int32(1), "query", nil)
	assert.Nil(t, res)
	assert.Error(t, err)
}
