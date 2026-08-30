package registry

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	runaimetricsimpl "github.com/run-ai/runai/runai-common-packages/metrics/internal/runaimetrics"
	"github.com/run-ai/runai/runai-common-packages/metrics/runaimetrics"
)

type HistogramRegisterer interface {
	Register(name, description string, labels map[string]string, buckets []float64) runaimetrics.MetricHistogram
	Get(name string) runaimetrics.MetricHistogram
}

func NewHistogramRegistry(promRegistry prometheus.Registerer) HistogramRegisterer {
	return &HistogramRegistry{
		promRegistry: promRegistry,
		histograms:   map[string]runaimetrics.MetricHistogram{},
	}
}

// HistogramRegistry is a struct that implements the Histogram interface.
// It contains a map of registered histogram metrics and a reference to the Prometheus registry.
type HistogramRegistry struct {
	histograms   map[string]runaimetrics.MetricHistogram
	promRegistry prometheus.Registerer
	mu           sync.Mutex
}

func (r *HistogramRegistry) Register(name, description string, labels map[string]string, buckets []float64) runaimetrics.MetricHistogram {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.histograms[name]; ok {
		return m
	}
	metric := runaimetricsimpl.NewPromHistogram(name, description, labels, buckets)
	r.promRegistry.MustRegister(metric.PromHistogram)
	r.histograms[name] = metric
	return metric
}

func (r *HistogramRegistry) Get(name string) runaimetrics.MetricHistogram {
	if _, ok := r.histograms[name]; !ok {
		return &DummyMetricHistogram{}
	}
	return r.histograms[name]
}

type HistogramAccessor interface {
	Register(description string, labels map[string]string, buckets []float64) runaimetrics.MetricHistogram
	Get() runaimetrics.MetricHistogram
}

// HistogramAccess is a struct that implements the HistogramAccessor interface.
// It contains a reference to the histogram registry and the name of the metric in the registry.
type HistogramAccess struct {
	Registry   HistogramRegisterer
	MetricName string
}

func (r *HistogramAccess) Register(description string, labels map[string]string, buckets []float64) runaimetrics.MetricHistogram {
	return r.Registry.Register(r.MetricName, description, labels, buckets)
}

func (r *HistogramAccess) Get() runaimetrics.MetricHistogram {
	return r.Registry.Get(r.MetricName)
}
