/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"time"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	"github.com/mitchellh/mapstructure"
)

type Metric interface {
	GetName() string
	GetFilters() interface{}
	GetResponseLabels() interface{}
	GetQuantile() *float32
	IsAggregation() bool
	IsQuantile() bool
}

type MetricGroupBy string

const (
	GroupByCluster    MetricGroupBy = "Cluster"
	GroupByNodePool   MetricGroupBy = "NodePool"
	GroupByNode       MetricGroupBy = "Node"
	GroupByCategory   MetricGroupBy = "Category"
	GroupByProject    MetricGroupBy = "Project"
	GroupByDepartment MetricGroupBy = "Department"
	GroupByGpu        MetricGroupBy = "Gpu"
)

type GroupedMetric interface {
	Metric
	GetGroupBy() MetricGroupBy
}

type MetricQuerier interface {
	SetFilters(filters map[string][]string) MetricQuerier
	AddFilter(key string, value string) MetricQuerier
	SetResponseLabels(responseLabels []string) MetricQuerier
	AddResponseLabel(responseLabel string) MetricQuerier
	SetRange(start time.Time, end time.Time) MetricQuerier
	SetDefaultNumOfSamples(numOfSamples int32) MetricQuerier
	SetIsAggregation(isAggregation bool) MetricQuerier
	SetQuantile(quantile *float32) MetricQuerier
	Query(ctx context.Context) (*openapi.GetMetrics200Response, error)
}

type metricQuery struct {
	name              string
	filters           map[string][]string
	responseLabels    []string
	start             *time.Time
	end               *time.Time
	numOfSamples      *int32
	quantile          *float32
	isQuantile        bool
	isAggregation     bool
	overrideTenantId  int32
	getClientInstance GetClientInstancer
}

func NewMetricQuery(metric Metric) (*metricQuery, error) {
	filterMap, err := filtersToMap(metric)
	if err != nil {
		return nil, err
	}
	responseLabelsArray, err := responseLabelsToArray(metric)
	if err != nil {
		return nil, err
	}

	return &metricQuery{
		name:              metric.GetName(),
		filters:           filterMap,
		responseLabels:    responseLabelsArray,
		isAggregation:     metric.IsAggregation(),
		isQuantile:        metric.IsQuantile(),
		getClientInstance: GetClientInstance,
	}, nil
}

func NewNamedMetricQuery(metricName string) *metricQuery {
	return &metricQuery{
		name:              metricName,
		getClientInstance: GetClientInstance,
	}
}

func (q *metricQuery) SetFilters(filters map[string][]string) MetricQuerier {
	q.filters = filters
	return q
}

func (q *metricQuery) AddFilter(key string, value string) MetricQuerier {
	if q.filters == nil {
		q.filters = make(map[string][]string)
	}
	q.filters[key] = append(q.filters[key], value)
	return q
}

func (q *metricQuery) SetResponseLabels(responseLabels []string) MetricQuerier {
	q.responseLabels = responseLabels
	return q
}

func (q *metricQuery) AddResponseLabel(responseLabel string) MetricQuerier {
	q.responseLabels = append(q.responseLabels, responseLabel)
	return q
}

func (q *metricQuery) SetRange(start time.Time, end time.Time) MetricQuerier {
	q.start = &start
	q.end = &end
	return q
}

func (q *metricQuery) SetQuantile(quantile *float32) MetricQuerier {
	q.quantile = quantile
	return q
}

func (q *metricQuery) SetDefaultNumOfSamples(numOfSamples int32) MetricQuerier {
	if numOfSamples == 0 {
		numOfSamples = readConfig().DefaultNumOfSamples
	}
	numOfSamples = int32(math.Min(float64(numOfSamples), float64(readConfig().MaxNumOfSamples)))
	numOfSamples = int32(math.Max(float64(numOfSamples), 1))

	q.numOfSamples = &numOfSamples
	return q
}

func (q *metricQuery) SetIsAggregation(isAggregation bool) MetricQuerier {
	q.isAggregation = isAggregation
	return q
}

func (q *metricQuery) SetOverrideTenantId(overrideTenantId int32) MetricQuerier {
	q.overrideTenantId = overrideTenantId
	return q
}

func (q *metricQuery) Query(ctx context.Context) (*openapi.GetMetrics200Response, error) {
	metricInfo, err := metricQueryToMetricInformation(*q)
	if err != nil {
		return nil, err
	}

	resp, _, err := q.getClientInstance().GetMetrics(ctx, metricInfo, q.overrideTenantId)
	return resp, err
}

func filtersToMap(m Metric) (map[string][]string, error) {
	mp := make(map[string]interface{})
	ans := make(map[string][]string)
	err := mapstructure.Decode(m.GetFilters(), &mp)
	if err != nil {
		return ans, fmt.Errorf("error decoding filters: %v", err)
	}
	for key, val := range mp {
		rv := reflect.ValueOf(val)
		ans[key] = arrayToStringArray(rv)
	}
	return ans, nil
}

func arrayToStringArray(slice reflect.Value) []string {
	ans := make([]string, 0)
	if slice.Kind() != reflect.Slice && slice.Kind() != reflect.Array {
		return ans
	}
	for i := 0; i < slice.Len(); i++ {
		ans = append(ans, fmt.Sprintf("%v", slice.Index(i)))
	}
	return ans
}

func responseLabelsToArray(m Metric) ([]string, error) {
	mp := make(map[string]bool)
	ans := make([]string, 0)
	err := mapstructure.Decode(m.GetResponseLabels(), &mp)
	if err != nil {
		return ans, fmt.Errorf("error decoding response labels: %v", err)
	}
	for key := range mp {
		ans = append(ans, key)
	}
	return ans, nil
}

func metricQueryToMetricInformation(mc metricQuery) (*openapi.MetricInformation, error) {
	metricInfo := openapi.NewMetricInformation(mc.name)
	metricInfo.SetFilters(mc.filters)
	metricInfo.SetResponseLabels(mc.responseLabels)

	r, err := getMetricsInformationRange(mc.name, mc.isAggregation, mc.start, mc.end, mc.numOfSamples)
	if err != nil {
		return nil, err
	}
	if r != nil {
		metricInfo.SetRange(*r)
	}

	err = validateQuantile(mc.name, mc.isQuantile, mc.quantile)
	if err != nil {
		return nil, err
	}
	if mc.quantile != nil {
		metricInfo.SetQuantile(*mc.quantile)
	}
	return metricInfo, nil
}

func getMetricsInformationRange(metricName string, isAggregation bool, start *time.Time, end *time.Time, numOfSamples *int32) (*openapi.MetricInformationRange, error) {
	r := buildMetricsInformationRange(start, end, numOfSamples)
	if r != nil {
		r.SetAggregation(isAggregation)
		return r, nil
	}

	if isAggregation {
		return nil, fmt.Errorf("aggregation metric %s requires range", metricName)
	}

	return nil, nil
}

func buildMetricsInformationRange(start *time.Time, end *time.Time, numOfSamples *int32) *openapi.MetricInformationRange {
	if start == nil || end == nil {
		return nil
	}
	r := openapi.NewMetricInformationRange(*start, *end)
	if numOfSamples != nil {
		r.NumOfSamples = numOfSamples
	}
	return r
}

func validateQuantile(metricName string, isQuantile bool, quantile *float32) error {
	if !isQuantile {
		return nil
	}
	if quantile == nil {
		return fmt.Errorf("quantile metric %s requires quantile", metricName)
	}
	if *quantile < 0 || *quantile > 1 {
		return fmt.Errorf("quantile metric %s requires quantile between 0 and 1", metricName)
	}
	return nil
}
