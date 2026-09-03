/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"errors"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/metrics/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/services"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

func TestInject(t *testing.T) {
	query := "some query[QUANTILE]"
	metricInfo := &services.MetricInformation{
		Quantile: ptr.To[float32](0.02),
	}
	res, err := NewQuantileInjector(nil).Inject(query, metricInfo)
	assert.NoError(t, err)
	assert.Equal(t, "some query[0.02]", res)
}

func TestInject_NoQuantile(t *testing.T) {
	query := "some query"
	res, err := NewQuantileInjector(nil).Inject(query, &services.MetricInformation{})
	assert.NoError(t, err)
	assert.Equal(t, query, res)
}

func TestInject_NoPlaceholder(t *testing.T) {
	query := "some query"
	metricInfo := &services.MetricInformation{
		Quantile: ptr.To[float32](0.5),
	}
	res, err := NewQuantileInjector(nil).Inject(query, metricInfo)
	assert.Equal(t, "quantile metric query must contain QUANTILE placeholder", err.Error())
	assert.Equal(t, "", res)
}

func TestInjectNextIsCalled_Quantile(t *testing.T) {
	next := mocks.NewQueryInjector(t)
	quantileInjector := NewQuantileInjector(next)
	query := "some query[QUANTILE]"
	metricInfo := &services.MetricInformation{
		Quantile: ptr.To[float32](0.02),
	}
	expectedQuery := "some query[0.02]"
	next.On("Inject", expectedQuery, metricInfo).Return("final query", nil)
	query, err := quantileInjector.Inject(query, metricInfo)
	assert.NoError(t, err)
	assert.Equal(t, "final query", query)
}

func TestInjectNextReturnsError_Quantile(t *testing.T) {
	next := mocks.NewQueryInjector(t)
	quantileInjector := NewQuantileInjector(next)
	query := "some query[QUANTILE]"
	metricInfo := &services.MetricInformation{
		Quantile: ptr.To[float32](0.02),
	}
	expectedQuery := "some query[0.02]"
	next.On("Inject", expectedQuery, metricInfo).Return("", errors.New("next error"))
	_, err := quantileInjector.Inject(query, metricInfo)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "next error")
}
