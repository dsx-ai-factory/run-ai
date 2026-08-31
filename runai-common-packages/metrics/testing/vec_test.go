/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"testing"

	"github.com/run-ai/runai/runai-common-packages/metrics"
)

func TestHistogramVecWithLabels(t *testing.T) {
	manager := SetMock(t)
	manager.HistogramVec("histogram_vec_test").Get().WithLabels(map[string]string{"label1": "value1"}).Observe(1.23)
	manager.HistogramVec("histogram_vec_test").Register("histogram_vec_test", []string{"label1"}, []float64{0.1, 1.0, 10.0}).WithLabels(map[string]string{"label1": "value1"}).Observe(1.23)

	metrics.HistogramVec("histogram_vec_test").Get().WithLabels(map[string]string{"label1": "value1"}).Observe(1.23)
	metrics.HistogramVec("histogram_vec_test").Register("histogram_vec_test", []string{"label1"}, []float64{0.1, 1.0, 10.0}).WithLabels(map[string]string{"label1": "value1"}).Observe(1.23)
}

func TestHistogramVecReset(t *testing.T) {
	manager := SetMock(t)
	manager.HistogramVec("histogram_vec_test").Get().Reset()
	metrics.HistogramVec("histogram_vec_test").Get().Reset()
}

func TestCounterVecWithLabels(t *testing.T) {
	manager := SetMock(t)
	manager.CounterVec("counter_vec_test").Get().WithLabels(map[string]string{"label1": "value1"}).Inc()
	manager.CounterVec("counter_vec_test").Register("counter_vec_test", []string{"label1"}).WithLabels(map[string]string{"label1": "value1"}).Inc()

	metrics.CounterVec("counter_vec_test").Get().WithLabels(map[string]string{"label1": "value1"}).Inc()
	metrics.CounterVec("counter_vec_test").Register("counter_vec_test", []string{"label1"}).WithLabels(map[string]string{"label1": "value1"}).Inc()
}

func TestCounterVecReset(t *testing.T) {
	manager := SetMock(t)
	manager.CounterVec("counter_vec_test").Get().Reset()
	metrics.CounterVec("counter_vec_test").Get().Reset()
}

func TestGaugeVecWithLabels(t *testing.T) {
	manager := SetMock(t)
	manager.GaugeVec("gauge_vec_test").Get().WithLabels(map[string]string{"label1": "value1"}).Set(42)
	manager.GaugeVec("gauge_vec_test").Register("gauge_vec_test", []string{"label1"}).WithLabels(map[string]string{"label1": "value1"}).Set(42)

	metrics.GaugeVec("gauge_vec_test").Get().WithLabels(map[string]string{"label1": "value1"}).Set(42)
	metrics.GaugeVec("gauge_vec_test").Register("gauge_vec_test", []string{"label1"}).WithLabels(map[string]string{"label1": "value1"}).Set(42)
}

func TestGaugeVecReset(t *testing.T) {
	manager := SetMock(t)
	manager.GaugeVec("gauge_vec_test").Get().Reset()
	metrics.GaugeVec("gauge_vec_test").Get().Reset()
}

func TestSummaryVecWithLabels(t *testing.T) {
	manager := SetMock(t)
	manager.SummaryVec("summary_vec_test").Get().WithLabels(map[string]string{"label1": "value1"}).Observe(3.14)
	manager.SummaryVec("summary_vec_test").Register("summary_vec_test", []string{"label1"}, map[float64]float64{0.5: 0.05, 0.9: 0.01}).WithLabels(map[string]string{"label1": "value1"}).Observe(3.14)

	metrics.SummaryVec("summary_vec_test").Get().WithLabels(map[string]string{"label1": "value1"}).Observe(3.14)
	metrics.SummaryVec("summary_vec_test").Register("summary_vec_test", []string{"label1"}, map[float64]float64{0.5: 0.05, 0.9: 0.01}).WithLabels(map[string]string{"label1": "value1"}).Observe(3.14)
}

func TestSummaryVecReset(t *testing.T) {
	manager := SetMock(t)
	manager.SummaryVec("summary_vec_test").Get().Reset()
	metrics.SummaryVec("summary_vec_test").Get().Reset()
}
