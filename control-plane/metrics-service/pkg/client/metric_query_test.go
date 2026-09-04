/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testMetric struct {
	Filters        testMetricFilters
	ResponseLabels testMetricResponseLabels
	Quantile       *float32
}

type testMetricFilters struct {
	TestFilter []string `mapstructure:"test_filter,omitempty"`
	AnotherOne []string `mapstructure:"AnotherOne,omitempty"`
}

type testMetricResponseLabels struct {
	TestFilter bool `mapstructure:"test_filter,omitempty"`
	AnotherOne bool `mapstructure:"AnotherOne,omitempty"`
}

func (m testMetric) GetName() string {
	return "testMetric"
}

func (m testMetric) GetFilters() interface{} {
	return m.Filters
}

func (m testMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m testMetric) IsAggregation() bool {
	return false
}

func (m testMetric) IsQuantile() bool {
	return false
}

func (m testMetric) GetQuantile() *float32 {
	return m.Quantile
}

type testQuantileMetric struct {
	Filters        testMetricFilters
	ResponseLabels testMetricResponseLabels
	Quantile       *float32
}

func (m testQuantileMetric) GetName() string {
	return "testQuantileMetric"
}

func (m testQuantileMetric) GetFilters() interface{} {
	return m.Filters
}

func (m testQuantileMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m testQuantileMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m testQuantileMetric) IsAggregation() bool {
	return false
}

func (m testQuantileMetric) IsQuantile() bool {
	return true
}

var (
	metricName = "test"
	metric1    = testMetric{
		Filters: testMetricFilters{
			TestFilter: []string{"test"},
		},
		ResponseLabels: testMetricResponseLabels{
			TestFilter: true,
		},
	}
)

var metric2 = testMetric{
	Filters: testMetricFilters{
		AnotherOne: []string{"test", "anotherOne"},
	},
	ResponseLabels: testMetricResponseLabels{
		TestFilter: false,
	},
}

var metricNoFiltersAndLabels = testMetric{}

var quantileMetric = testQuantileMetric{}

func TestMetricQuery_metric1(t *testing.T) {
	query, err := NewMetricQuery(metric1)
	assert.Nil(t, err)
	assert.Equal(t, "testMetric", query.name)
	assert.Equal(t, 1, len(query.filters))
	assert.Equal(t, 1, len(query.filters["test_filter"]))
	assert.Equal(t, "test", query.filters["test_filter"][0])
	assert.Equal(t, 1, len(query.responseLabels))
	assert.Equal(t, "test_filter", query.responseLabels[0])
	assert.False(t, query.isQuantile)
}

func TestMetricQuery_metric2(t *testing.T) {
	query, err := NewMetricQuery(metric2)
	assert.Nil(t, err)
	assert.Equal(t, "testMetric", query.name)
	assert.Equal(t, 1, len(query.filters))
	assert.Equal(t, 2, len(query.filters["AnotherOne"]))
	assert.Equal(t, "test", query.filters["AnotherOne"][0])
	assert.Equal(t, "anotherOne", query.filters["AnotherOne"][1])
	assert.Equal(t, 0, len(query.responseLabels))
	assert.False(t, query.isQuantile)
}

func TestNamedMetricQuery_Empty(t *testing.T) {
	query := NewNamedMetricQuery(metricName)
	assert.Equal(t, metricName, query.name)
	assert.Equal(t, 0, len(query.filters))
	assert.Equal(t, 0, len(query.responseLabels))
}

func TestMetricQuery_SetFilters(t *testing.T) {
	query := NewNamedMetricQuery(metricName)
	query.SetFilters(map[string][]string{
		"test_filter": {"test"},
	})
	assert.Equal(t, 1, len(query.filters))
	assert.Equal(t, 1, len(query.filters["test_filter"]))
}

func TestMetricQuery_AddFilterWhenEmpty(t *testing.T) {
	query := NewNamedMetricQuery(metricName)
	query.AddFilter("test_filter", "test")
	assert.Equal(t, 1, len(query.filters))
	assert.Equal(t, 1, len(query.filters["test_filter"]))
}

func TestMetricQuery_AddFilterWhenNotEmpty(t *testing.T) {
	query := metricQuery{
		name: metricName,
		filters: map[string][]string{
			"test_filter": {"test"},
		},
	}
	query.AddFilter("test_filter", "anotherOne")
	assert.Equal(t, 1, len(query.filters))
	assert.Equal(t, 2, len(query.filters["test_filter"]))
	assert.Equal(t, "test", query.filters["test_filter"][0])
	assert.Equal(t, "anotherOne", query.filters["test_filter"][1])
}

func TestMetricQuery_SetResponseLabels(t *testing.T) {
	query := NewNamedMetricQuery(metricName)
	query.SetResponseLabels([]string{"test_filter"})
	assert.Equal(t, 1, len(query.responseLabels))
	assert.Equal(t, "test_filter", query.responseLabels[0])
}

func TestMetricQuery_AddResponseLabelWhenEmpty(t *testing.T) {
	query := NewNamedMetricQuery(metricName)
	query.AddResponseLabel("test_filter")
	assert.Equal(t, 1, len(query.responseLabels))
	assert.Equal(t, "test_filter", query.responseLabels[0])
}

func TestMetricQuery_AddResponseLabelWhenNotEmpty(t *testing.T) {
	query := metricQuery{
		name: metricName,
		responseLabels: []string{
			"test_filter",
		},
	}
	query.AddResponseLabel("anotherOne")
	assert.Equal(t, 2, len(query.responseLabels))
	assert.Equal(t, "test_filter", query.responseLabels[0])
	assert.Equal(t, "anotherOne", query.responseLabels[1])
}

func TestNewNamedMetricQuery_SetRange(t *testing.T) {
	query := NewNamedMetricQuery(metricName)
	testEndTime := time.Now()
	testStartTime := testEndTime.Add(-time.Hour)
	query.SetRange(testStartTime, testEndTime)
	assert.Equal(t, testStartTime, *query.start)
	assert.Equal(t, testEndTime, *query.end)
}

func TestMetricQuery_SetNumOfSamples(t *testing.T) {
	query := NewNamedMetricQuery(metricName)
	testNumOfSamples := int32(100)
	query.SetDefaultNumOfSamples(testNumOfSamples)
	assert.Equal(t, testNumOfSamples, *query.numOfSamples)
}

func TestMetricQueryToMetricInformation(t *testing.T) {
	query, err := NewMetricQuery(metric1)
	assert.Nil(t, err)
	query.SetRange(time.Now().Add(-time.Hour), time.Now())
	query.SetDefaultNumOfSamples(100)
	metricInfo, _ := metricQueryToMetricInformation(*query)
	assert.Equal(t, query.name, metricInfo.Name)
	assert.Equal(t, query.filters, *metricInfo.Filters)
	assert.Equal(t, query.responseLabels, metricInfo.ResponseLabels)
	assert.Equal(t, *query.start, metricInfo.GetRange().Start)
	assert.Equal(t, *query.end, metricInfo.GetRange().End)
	assert.Equal(t, query.numOfSamples, metricInfo.GetRange().NumOfSamples)
}

func TestMetricQueryToMetricInformation_NoNumOfSamples(t *testing.T) {
	query, err := NewMetricQuery(metric1)
	assert.Nil(t, err)
	query.SetRange(time.Now().Add(-time.Hour), time.Now())
	metricInfo, _ := metricQueryToMetricInformation(*query)
	assert.Equal(t, *query.start, metricInfo.Range.Get().Start)
	assert.Equal(t, *query.end, metricInfo.Range.Get().End)
	assert.Equal(t, (int32)(20), *metricInfo.GetRange().NumOfSamples)
}

func TestMetricQueryToMetricInformation_NoStart(t *testing.T) {
	query, err := NewMetricQuery(metric1)
	assert.Nil(t, err)
	query.end = &time.Time{}
	metricInfo, _ := metricQueryToMetricInformation(*query)
	assert.Equal(t, openapi.NullableMetricInformationRange{}, metricInfo.Range)
	assert.Equal(t, (*int32)(nil), metricInfo.GetRange().NumOfSamples)
}

func TestFiltersToMap(t *testing.T) {
	filters, err := filtersToMap(metric1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(filters))
	assert.Equal(t, 1, len(filters["test_filter"]))
	assert.Equal(t, "test", filters["test_filter"][0])
}

func TestFiltersToMap_NoFilters(t *testing.T) {
	filters, err := filtersToMap(metricNoFiltersAndLabels)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(filters))
}

func TestArrayToStringArray(t *testing.T) {
	uid := uuid.New()
	array := arrayToStringArray(reflect.ValueOf([]interface{}{1, "anotherOne", true, uid}))
	assert.Equal(t, 4, len(array))
	assert.Equal(t, "1", array[0])
	assert.Equal(t, "anotherOne", array[1])
	assert.Equal(t, "true", array[2])
	assert.Equal(t, uid.String(), array[3])
}

func TestArrayToStringArray_NotArray(t *testing.T) {
	array := arrayToStringArray(reflect.ValueOf(true))
	assert.Equal(t, 0, len(array))
}

func TestResponseLabelsToArray(t *testing.T) {
	responseLabels, err := responseLabelsToArray(metric1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(responseLabels))
	assert.Equal(t, "test_filter", responseLabels[0])
}

func TestResponseLabelsToArray_NoResponseLabels(t *testing.T) {
	responseLabels, err := responseLabelsToArray(metricNoFiltersAndLabels)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(responseLabels))
}

func TestNewNamedMetricQuery_Query(t *testing.T) {
	clientMock := mocks.NewMetricsClientContract(t)
	getClientInstanceMock := func() MetricsClientContract {
		return clientMock
	}
	ctx := context.Background()
	clientMock.On("GetMetrics", ctx, mock.Anything, mock.Anything).Return(nil, 200, nil)
	query := metricQuery{
		name:              metricName,
		getClientInstance: getClientInstanceMock,
	}
	query.AddResponseLabel("test_filter")
	query.AddFilter("test_filter", "anotherOne")
	query.SetRange(time.Now().Add(-time.Hour), time.Now())
	query.SetDefaultNumOfSamples(100)
	query.Query(ctx)
	clientMock.AssertCalled(t, "GetMetrics", ctx, mock.Anything, mock.Anything)
	clientMock.AssertNumberOfCalls(t, "GetMetrics", 1)
}

func TestQuantileMetricQuery(t *testing.T) {
	query, err := NewMetricQuery(quantileMetric)
	assert.Nil(t, err)
	assert.Equal(t, "testQuantileMetric", query.name)
	assert.Equal(t, 0, len(query.filters))
	assert.Equal(t, 0, len(query.responseLabels))
	assert.True(t, query.isQuantile)
}

func TestQuantileMetricQuery_Query(t *testing.T) {
	clientMock := mocks.NewMetricsClientContract(t)
	getClientInstanceMock := func() MetricsClientContract {
		return clientMock
	}
	ctx := context.Background()
	clientMock.On("GetMetrics", ctx, mock.Anything, mock.Anything).Return(nil, 200, nil)

	query, err := NewMetricQuery(quantileMetric)
	assert.Nil(t, err)

	query.getClientInstance = getClientInstanceMock
	var quantile float32 = 0.5
	query.SetQuantile(&quantile)

	_, err = query.Query(ctx)
	assert.Nil(t, err)
}

func TestQuantileMetricQuery_Query_WithoutQuantile(t *testing.T) {
	ctx := context.Background()
	query, err := NewMetricQuery(quantileMetric)
	assert.Nil(t, err)

	_, err = query.Query(ctx)
	assert.Equal(t, "quantile metric testQuantileMetric requires quantile", err.Error())
}

func TestQuantileMetricQuery_Query_InvalidQuantile(t *testing.T) {
	ctx := context.Background()
	query, err := NewMetricQuery(quantileMetric)
	assert.Nil(t, err)

	var quantile float32 = -1
	query.SetQuantile(&quantile)

	_, err = query.Query(ctx)
	assert.Equal(t, "quantile metric testQuantileMetric requires quantile between 0 and 1", err.Error())
}

func TestQuantileMetricQuery_Query_ApiError(t *testing.T) {
	clientMock := mocks.NewMetricsClientContract(t)
	getClientInstanceMock := func() MetricsClientContract {
		return clientMock
	}
	ctx := context.Background()
	clientMock.On("GetMetrics", ctx, mock.Anything, mock.Anything).Return(nil, 500, fmt.Errorf("failed to get metrics"))
	query, err := NewMetricQuery(quantileMetric)
	assert.Nil(t, err)

	query.getClientInstance = getClientInstanceMock
	var quantile float32 = 0.5
	query.SetQuantile(&quantile)

	_, err = query.Query(ctx)
	assert.Equal(t, "failed to get metrics", err.Error())
}
