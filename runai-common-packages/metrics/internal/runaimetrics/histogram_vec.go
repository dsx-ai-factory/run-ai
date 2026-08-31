/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/run-ai/runai/runai-common-packages/metrics/runaimetrics"
)

type HistogramVec struct {
	PromHistogramVec *prometheus.HistogramVec
}

func NewPromHistogramVec(name string, description string, labels []string, buckets []float64) *HistogramVec {
	histogramVec := HistogramVec{
		PromHistogramVec: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    METRICS_PREFIX + name,
			Help:    description,
			Buckets: buckets,
		}, labels),
	}
	return &histogramVec
}

func (cm *HistogramVec) WithLabels(labels map[string]string) runaimetrics.MetricHistogram {
	return &Histogram{
		PromHistogram: cm.PromHistogramVec.With(labels).(prometheus.Histogram),
	}
}

func (cm *HistogramVec) Reset() {
	cm.PromHistogramVec.Reset()
}
