/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package registry

import (
	"sync"

	runaimetricsimpl "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/internal/runaimetrics"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/runaimetrics"
	"github.com/prometheus/client_golang/prometheus"
)

type SummaryRegisterer interface {
	Register(name, description string, labels map[string]string, objectives map[float64]float64) runaimetrics.MetricSummary
	Get(name string) runaimetrics.MetricSummary
}

func NewSummaryRegistry(promRegistry prometheus.Registerer) SummaryRegisterer {
	return &SummaryRegistry{
		promRegistry: promRegistry,
		summarys:     map[string]runaimetrics.MetricSummary{},
	}
}

// SummaryRegistry is a struct that implements the Summary interface.
// It contains a map of registered summary metrics and a reference to the Prometheus registry.
type SummaryRegistry struct {
	summarys     map[string]runaimetrics.MetricSummary
	promRegistry prometheus.Registerer
	mu           sync.Mutex
}

func (r *SummaryRegistry) Register(name, description string, labels map[string]string, objectives map[float64]float64) runaimetrics.MetricSummary {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.summarys[name]; ok {
		return m
	}
	metric := runaimetricsimpl.NewPromSummary(name, description, labels, objectives)
	r.promRegistry.MustRegister(metric.PromSummary)
	r.summarys[name] = metric
	return metric
}

func (r *SummaryRegistry) Get(name string) runaimetrics.MetricSummary {
	if _, ok := r.summarys[name]; !ok {
		return &DummyMetricSummary{}
	}
	return r.summarys[name]
}

type SummaryAccessor interface {
	Register(description string, labels map[string]string, objectives map[float64]float64) runaimetrics.MetricSummary
	Get() runaimetrics.MetricSummary
}

// SummaryAccess is a struct that implements the SummaryAccessor interface.
// It contains a reference to the summary registry and the name of the metric in the registry.
type SummaryAccess struct {
	Registry   SummaryRegisterer
	MetricName string
}

func (r *SummaryAccess) Register(description string, labels map[string]string, objectives map[float64]float64) runaimetrics.MetricSummary {
	return r.Registry.Register(r.MetricName, description, labels, objectives)
}

func (r *SummaryAccess) Get() runaimetrics.MetricSummary {
	return r.Registry.Get(r.MetricName)
}
