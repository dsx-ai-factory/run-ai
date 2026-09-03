/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"errors"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/metrics/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/services"
	"github.com/stretchr/testify/assert"
)

func Test_Inject_NoAggregation(t *testing.T) {
	query := "some query"
	r := &services.MetricInformation{
		Range: &services.MetricInformationRange{
			Aggregation: false,
		},
	}
	res, err := NewAggregationRangeInjector(nil).Inject(query, r)
	assert.NoError(t, err)
	assert.Equal(t, query, res)
}

func Test_Inject_NoRange(t *testing.T) {
	query := "some query"
	r := &services.MetricInformation{
		Range: nil,
	}
	res, err := NewAggregationRangeInjector(nil).Inject(query, r)
	assert.NoError(t, err)
	assert.Equal(t, query, res)
}

func Test_Inject_BadRange(t *testing.T) {
	query := "some query"
	r := &services.MetricInformation{
		Range: &services.MetricInformationRange{
			Aggregation: true,
		},
	}
	res, err := NewAggregationRangeInjector(nil).Inject(query, r)
	assert.Error(t, err)
	assert.Equal(t, "", res)
}

func Test_Inject(t *testing.T) {
	start := time.Now()
	end := time.Now().Add(1 * time.Hour)
	query := "some query[RANGE]"
	r := &services.MetricInformation{
		Range: &services.MetricInformationRange{
			Start:       start,
			End:         end,
			Aggregation: true,
		},
	}
	res, err := NewAggregationRangeInjector(nil).Inject(query, r)
	assert.NoError(t, err)
	assert.Equal(t, "some query[3600s]", res)
}

func Test_GetTimeDiff(t *testing.T) {
	start := time.Now()
	end := time.Now().Add(1 * time.Hour)
	r := &services.MetricInformationRange{
		Start:       start,
		End:         end,
		Aggregation: true,
	}
	diff, err := getTimeDiff(r)
	assert.NoError(t, err)
	assert.Equal(t, 3600, diff)
}

func Test_GetTimeDiff_Error(t *testing.T) {
	start := time.Now()
	end := time.Now().Add(-1 * time.Hour)
	r := &services.MetricInformationRange{
		Start:       start,
		End:         end,
		Aggregation: true,
	}
	_, err := getTimeDiff(r)
	assert.Error(t, err)
}

func TestInjectNextIsCalled_AggregationRange(t *testing.T) {
	next := mocks.NewQueryInjector(t)
	aggregationRangeInjector := NewAggregationRangeInjector(next)
	query := "some query[RANGE]"
	metricInfo := &services.MetricInformation{
		Range: &services.MetricInformationRange{
			Start:       time.Now(),
			End:         time.Now().Add(1 * time.Hour),
			Aggregation: true,
		},
	}
	expectedQuery := "some query[3600s]"
	next.On("Inject", expectedQuery, metricInfo).Return("final query", nil)
	query, err := aggregationRangeInjector.Inject(query, metricInfo)
	assert.NoError(t, err)
	assert.Equal(t, "final query", query)
}

func TestInjectNextReturnsError_AggregationRange(t *testing.T) {
	next := mocks.NewQueryInjector(t)
	aggregationRangeInjector := NewAggregationRangeInjector(next)
	query := "some query[RANGE]"
	metricInfo := &services.MetricInformation{
		Range: &services.MetricInformationRange{
			Start:       time.Now(),
			End:         time.Now().Add(1 * time.Hour),
			Aggregation: true,
		},
	}
	expectedQuery := "some query[3600s]"
	next.On("Inject", expectedQuery, metricInfo).Return("", errors.New("next error"))
	_, err := aggregationRangeInjector.Inject(query, metricInfo)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "next error")
}
