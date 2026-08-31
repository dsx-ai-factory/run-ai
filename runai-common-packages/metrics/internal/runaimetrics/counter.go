/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Counter struct {
	PromCounter prometheus.Counter
}

func NewPromCounter(name string, description string, labels map[string]string) *Counter {
	counterMetric := Counter{
		PromCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        METRICS_PREFIX + name,
			Help:        description,
			ConstLabels: labels,
		}),
	}
	return &counterMetric
}

func (cm *Counter) Inc() {
	cm.PromCounter.Inc()
}

func (cm *Counter) Add(value float64) {
	cm.PromCounter.Add(value)
}
