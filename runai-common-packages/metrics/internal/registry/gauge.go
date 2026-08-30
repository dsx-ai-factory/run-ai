package registry

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	runaimetricsimpl "github.com/run-ai/runai/runai-common-packages/metrics/internal/runaimetrics"
	"github.com/run-ai/runai/runai-common-packages/metrics/runaimetrics"
)

type GaugeRegisterer interface {
	Register(name, description string, labels map[string]string) runaimetrics.MetricGauge
	Get(name string) runaimetrics.MetricGauge
}

func NewGaugeRegistry(promRegistry prometheus.Registerer) GaugeRegisterer {
	return &GaugeRegistry{
		promRegistry: promRegistry,
		gauges:       map[string]runaimetrics.MetricGauge{},
	}
}

// GaugeRegistry is a struct that manages the registration and retrieval of gauge metrics.
// It implements the Gauge interface and interacts with a Prometheus registry.
type GaugeRegistry struct {
	gauges       map[string]runaimetrics.MetricGauge
	promRegistry prometheus.Registerer
	mu           sync.Mutex
}

func (r *GaugeRegistry) Register(name, description string, labels map[string]string) runaimetrics.MetricGauge {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.gauges[name]; ok {
		return m
	}
	metric := runaimetricsimpl.NewPromGauge(name, description, labels)
	r.promRegistry.MustRegister(metric.PromGauge)
	r.gauges[name] = metric
	return metric
}

func (r *GaugeRegistry) Get(name string) runaimetrics.MetricGauge {
	if _, ok := r.gauges[name]; !ok {
		return &DummyMetricGauge{}
	}
	return r.gauges[name]
}

type GaugeAccessor interface {
	Register(description string, labels map[string]string) runaimetrics.MetricGauge

	Get() runaimetrics.MetricGauge
}

// GaugeAccess is a struct that implements the GaugeAccessor interface.
// It contains a reference to the gauge registry and the name of the metric in the registry.
type GaugeAccess struct {
	Registry   GaugeRegisterer
	MetricName string
}

func (r *GaugeAccess) Register(description string, labels map[string]string) runaimetrics.MetricGauge {
	return r.Registry.Register(r.MetricName, description, labels)
}

func (r *GaugeAccess) Get() runaimetrics.MetricGauge {
	return r.Registry.Get(r.MetricName)
}
