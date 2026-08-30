package registry

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	runaimetricsimpl "github.com/run-ai/runai/runai-common-packages/metrics/internal/runaimetrics"
	"github.com/run-ai/runai/runai-common-packages/metrics/runaimetrics"
)

type GaugeVecRegisterer interface {
	Register(name, description string, labels []string) runaimetrics.GaugeVector
	Get(name string) runaimetrics.GaugeVector
}

func NewGaugeVecRegistry(promRegistry prometheus.Registerer) GaugeVecRegisterer {
	return &GaugeVecRegistry{
		promRegistry: promRegistry,
		vecs:         map[string]runaimetrics.GaugeVector{},
	}
}

// GaugeVecRegistry is a struct that implements the GaugeVector interface.
// It contains a map of registered gauge vectors and a reference to the Prometheus registry.
type GaugeVecRegistry struct {
	vecs         map[string]runaimetrics.GaugeVector
	promRegistry prometheus.Registerer
	mu           sync.Mutex
}

func (r *GaugeVecRegistry) Register(name, description string, labels []string) runaimetrics.GaugeVector {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.vecs[name]; ok {
		return m
	}
	vec := runaimetricsimpl.NewPromGaugeVec(name, description, labels)
	r.promRegistry.MustRegister(vec.PromGaugeVec)
	r.vecs[name] = vec
	return vec
}

func (r *GaugeVecRegistry) Get(name string) runaimetrics.GaugeVector {
	if _, ok := r.vecs[name]; !ok {
		return &DummyGaugeVector{}
	}
	return r.vecs[name]
}

type GaugeVecAccessor interface {
	Register(description string, labels []string) runaimetrics.GaugeVector
	Get() runaimetrics.GaugeVector
}

// GaugeVecAccess is a struct that implements the GaugeVecAccessor interface.
// It contains a reference to the gauge vec registry and the name of the vec in the registry.
type GaugeVecAccess struct {
	Registry GaugeVecRegisterer
	VecName  string
}

func (r *GaugeVecAccess) Register(description string, labels []string) runaimetrics.GaugeVector {
	return r.Registry.Register(r.VecName, description, labels)
}

func (r *GaugeVecAccess) Get() runaimetrics.GaugeVector {
	return r.Registry.Get(r.VecName)
}
