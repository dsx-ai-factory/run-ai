package registry

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	runaimetricsimpl "github.com/run-ai/runai/runai-common-packages/metrics/internal/runaimetrics"
	"github.com/run-ai/runai/runai-common-packages/metrics/runaimetrics"
)

type SummaryVecRegisterer interface {
	Register(name, description string, labels []string, objectives map[float64]float64) runaimetrics.SummaryVector
	Get(name string) runaimetrics.SummaryVector
}

func NewSummaryVecRegistry(promRegistry prometheus.Registerer) SummaryVecRegisterer {
	return &SummaryVecRegistry{
		promRegistry: promRegistry,
		vecs:         map[string]runaimetrics.SummaryVector{},
	}
}

// SummaryVecRegistry is a struct that implements the SummaryVector interface.
// It contains a map of registered summary vectors and a reference to the Prometheus registry.
type SummaryVecRegistry struct {
	vecs         map[string]runaimetrics.SummaryVector
	promRegistry prometheus.Registerer
	mu           sync.Mutex
}

func (r *SummaryVecRegistry) Register(name, description string, labels []string, objectives map[float64]float64) runaimetrics.SummaryVector {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.vecs[name]; ok {
		return m
	}
	vec := runaimetricsimpl.NewPromSummaryVec(name, description, labels, objectives)
	r.promRegistry.MustRegister(vec.PromSummaryVec)
	r.vecs[name] = vec
	return vec
}

func (r *SummaryVecRegistry) Get(name string) runaimetrics.SummaryVector {
	if _, ok := r.vecs[name]; !ok {
		return &dummySummaryVector{}
	}
	return r.vecs[name]
}

type SummaryVecAccessor interface {
	Register(description string, labels []string, objectives map[float64]float64) runaimetrics.SummaryVector
	Get() runaimetrics.SummaryVector
}

// SummaryVecAccess is a struct that implements the SummaryVecAccessor interface.
// It contains a reference to the summary vec registry and the name of the vec in the registry.
type SummaryVecAccess struct {
	Registry SummaryVecRegisterer
	VecName  string
}

func (r *SummaryVecAccess) Register(description string, labels []string, objectives map[float64]float64) runaimetrics.SummaryVector {
	return r.Registry.Register(r.VecName, description, labels, objectives)
}

func (r *SummaryVecAccess) Get() runaimetrics.SummaryVector {
	return r.Registry.Get(r.VecName)
}
