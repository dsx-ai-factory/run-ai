/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Summary struct {
	PromSummary prometheus.Summary
}

func NewPromSummary(name string, description string, labels map[string]string, objectives map[float64]float64) *Summary {
	summaryMetric := Summary{
		PromSummary: prometheus.NewSummary(prometheus.SummaryOpts{
			Name:        METRICS_PREFIX + name,
			Help:        description,
			ConstLabels: labels,
			Objectives:  objectives,
		}),
	}
	return &summaryMetric
}

func (sm *Summary) Observe(value float64) {
	sm.PromSummary.Observe(value)
}
