/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/run-ai/runai/runai-common-packages/metrics/runaimetrics"
)

type CounterVec struct {
	PromCounterVec *prometheus.CounterVec
}

func NewPromCounterVec(name string, description string, labels []string) *CounterVec {
	counterVecMetric := CounterVec{
		PromCounterVec: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: METRICS_PREFIX + name,
			Help: description,
		}, labels),
	}
	return &counterVecMetric
}

func (cm *CounterVec) WithLabels(labels map[string]string) runaimetrics.MetricCounter {
	return &Counter{
		PromCounter: cm.PromCounterVec.With(labels),
	}
}

func (cm *CounterVec) Reset() {
	cm.PromCounterVec.Reset()
}
