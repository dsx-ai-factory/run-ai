/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/runaimetrics"
	"github.com/prometheus/client_golang/prometheus"
)

type GaugeVec struct {
	PromGaugeVec *prometheus.GaugeVec
}

func NewPromGaugeVec(name string, description string, labels []string) *GaugeVec {
	gaugeVec := GaugeVec{
		PromGaugeVec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: METRICS_PREFIX + name,
			Help: description,
		}, labels),
	}
	return &gaugeVec
}

func (cm *GaugeVec) WithLabels(labels map[string]string) runaimetrics.MetricGauge {
	return &Gauge{
		PromGauge: cm.PromGaugeVec.With(labels),
	}
}

func (cm *GaugeVec) Reset() {
	cm.PromGaugeVec.Reset()
}
