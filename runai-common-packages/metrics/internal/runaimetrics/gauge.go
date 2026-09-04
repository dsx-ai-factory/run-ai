/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Gauge struct {
	PromGauge prometheus.Gauge
}

func NewPromGauge(name string, description string, labels map[string]string) *Gauge {
	gaugeMetric := Gauge{
		PromGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        METRICS_PREFIX + name,
			Help:        description,
			ConstLabels: labels,
		}),
	}
	return &gaugeMetric
}

func (gm *Gauge) Set(value float64) {
	gm.PromGauge.Set(value)
}

func (gm *Gauge) Inc() {
	gm.PromGauge.Inc()
}

func (gm *Gauge) Dec() {
	gm.PromGauge.Dec()
}

func (gm *Gauge) Add(value float64) {
	gm.PromGauge.Add(value)
}

func (gm *Gauge) Sub(value float64) {
	gm.PromGauge.Sub(value)
}
