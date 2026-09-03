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

type HistogramVecRegisterer interface {
	Register(name, description string, labels []string, buckets []float64) runaimetrics.HistogramVector
	Get(name string) runaimetrics.HistogramVector
}

func NewHistogramVecRegistry(promRegistry prometheus.Registerer) HistogramVecRegisterer {
	return &HistogramVecRegistry{
		promRegistry: promRegistry,
		vecs:         map[string]runaimetrics.HistogramVector{},
	}
}

// HistogramVecRegistry is a struct that implements the HistogramVector interface.
// It contains a map of registered histogram vectors and a reference to the Prometheus registry.
type HistogramVecRegistry struct {
	vecs         map[string]runaimetrics.HistogramVector
	promRegistry prometheus.Registerer
	mu           sync.Mutex
}

func (r *HistogramVecRegistry) Register(name, description string, labels []string, buckets []float64) runaimetrics.HistogramVector {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.vecs[name]; ok {
		return m
	}
	vec := runaimetricsimpl.NewPromHistogramVec(name, description, labels, buckets)
	r.promRegistry.MustRegister(vec.PromHistogramVec)
	r.vecs[name] = vec
	return vec
}

func (r *HistogramVecRegistry) Get(name string) runaimetrics.HistogramVector {
	if _, ok := r.vecs[name]; !ok {
		return &DummyHistogramVector{}
	}
	return r.vecs[name]
}

type HistogramVecAccessor interface {
	Register(description string, labels []string, buckets []float64) runaimetrics.HistogramVector
	Get() runaimetrics.HistogramVector
}

// HistogramVecAccess is a struct that implements the HistogramVecAccessor interface.
// It contains a reference to the histogram vec registry and the name of the vec in the registry.
type HistogramVecAccess struct {
	Registry HistogramVecRegisterer
	VecName  string
}

func (r *HistogramVecAccess) Register(description string, labels []string, buckets []float64) runaimetrics.HistogramVector {
	return r.Registry.Register(r.VecName, description, labels, buckets)
}

func (r *HistogramVecAccess) Get() runaimetrics.HistogramVector {
	return r.Registry.Get(r.VecName)
}
