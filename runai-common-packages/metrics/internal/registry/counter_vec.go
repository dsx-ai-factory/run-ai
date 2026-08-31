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

type CounterVecRegisterer interface {
	Register(name, description string, labels []string) runaimetrics.CounterVector
	Get(name string) runaimetrics.CounterVector
}

func NewCounterVecRegistry(promRegistry prometheus.Registerer) CounterVecRegisterer {
	return &CounterVecRegistry{
		promRegistry: promRegistry,
		vecs:         map[string]runaimetrics.CounterVector{},
	}
}

// CounterVecRegistry is a struct that implements the CounterVector interface.
// It contains a map of registered counter vectors and a reference to the Prometheus registry.
type CounterVecRegistry struct {
	vecs         map[string]runaimetrics.CounterVector
	promRegistry prometheus.Registerer
	mu           sync.Mutex
}

func (r *CounterVecRegistry) Register(name, description string, labels []string) runaimetrics.CounterVector {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.vecs[name]; ok {
		return m
	}
	vec := runaimetricsimpl.NewPromCounterVec(name, description, labels)
	r.promRegistry.MustRegister(vec.PromCounterVec)
	r.vecs[name] = vec
	return vec
}

func (r *CounterVecRegistry) Get(name string) runaimetrics.CounterVector {
	if _, ok := r.vecs[name]; !ok {
		return &DummyCounterVector{}
	}
	return r.vecs[name]
}

type CounterVecAccessor interface {
	Register(description string, labels []string) runaimetrics.CounterVector
	Get() runaimetrics.CounterVector
}

// CounterVecAccess is a struct that implements the CounterVecAccessor interface.
// It contains a reference to the counter vec registry and the name of the vec in the registry.
type CounterVecAccess struct {
	Registry CounterVecRegisterer
	VecName  string
}

func (r *CounterVecAccess) Register(description string, labels []string) runaimetrics.CounterVector {
	return r.Registry.Register(r.VecName, description, labels)
}

func (r *CounterVecAccess) Get() runaimetrics.CounterVector {
	return r.Registry.Get(r.VecName)
}
