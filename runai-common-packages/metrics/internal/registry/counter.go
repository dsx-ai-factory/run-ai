/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package registry

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	runaimetricsimpl "github.com/run-ai/runai/runai-common-packages/metrics/internal/runaimetrics"
	"github.com/run-ai/runai/runai-common-packages/metrics/runaimetrics"
)

type CounterRegisterer interface {
	Register(name, description string, labels map[string]string) runaimetrics.MetricCounter
	Get(name string) runaimetrics.MetricCounter
}

func NewCounterRegistry(promRegistry prometheus.Registerer) CounterRegisterer {
	return &CounterRegistry{
		promRegistry: promRegistry,
		counters:     map[string]runaimetrics.MetricCounter{},
	}
}

type CounterRegistry struct {
	counters     map[string]runaimetrics.MetricCounter
	promRegistry prometheus.Registerer
	mu           sync.Mutex
}

func (r *CounterRegistry) Register(name, description string, labels map[string]string) runaimetrics.MetricCounter {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.counters[name]; ok {
		return m
	}
	metric := runaimetricsimpl.NewPromCounter(name, description, labels)
	r.promRegistry.MustRegister(metric.PromCounter)
	r.counters[name] = metric
	return metric
}

func (r *CounterRegistry) Get(name string) runaimetrics.MetricCounter {
	if _, ok := r.counters[name]; !ok {
		return &DummyMetricCounter{}
	}
	return r.counters[name]
}

type CounterAccessor interface {
	Register(description string, labels map[string]string) runaimetrics.MetricCounter
	Get() runaimetrics.MetricCounter
}

// CounterAccess is a struct that implements the CounterAccessor interface.
// It contains a reference to the counter registry and the name of the metric in the registry.
type CounterAccess struct {
	Registry   CounterRegisterer
	MetricName string
}

func (r *CounterAccess) Register(description string, labels map[string]string) runaimetrics.MetricCounter {
	return r.Registry.Register(r.MetricName, description, labels)
}

func (r *CounterAccess) Get() runaimetrics.MetricCounter {
	return r.Registry.Get(r.MetricName)
}
