package metrics

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/run-ai/runai/runai-common-packages/metrics/internal/registry"
)

var (
	metricsRegistry MetricsRegistrar
	extraRegistries sync.Map = sync.Map{}
	metricsPush     MetricsPusher
	syncOnce        sync.Once
)

// Metric is an interface for a metric.
type MetricsRegistrar interface {
	GetPrometheusRegistry() *prometheus.Registry
	Histogram(name string) registry.HistogramAccessor
	Counter(name string) registry.CounterAccessor
	Gauge(name string) registry.GaugeAccessor
	Summary(name string) registry.SummaryAccessor
	Handler() http.Handler
	HistogramVec(name string) registry.HistogramVecAccessor
	CounterVec(name string) registry.CounterVecAccessor
	GaugeVec(name string) registry.GaugeVecAccessor
	SummaryVec(name string) registry.SummaryVecAccessor
	BuildMetricsString() (string, error)
}

type MetricsPusher interface {
	Pusher
	RegistryGetter
}

type RegistryGetter interface {
	GetRegistry() MetricsRegistrar
}

type Pusher interface {
	Push() error
}

// NewMetrics initializes a new Metrics instance with default Go runtime metrics.
func Init(config *Config) (MetricsRegistrar, error) {
	config, err := readConfig(config)
	if err != nil {
		return nil, err
	}
	syncOnce.Do(func() {
		metricsRegistry = newMetricsRegistry(initPromRegistry())
		metricsPush = newMetricsPusher(prometheus.NewRegistry(), config.ServiceName, config.PushGatewayUrl)
		server := &http.Server{
			Addr:              fmt.Sprintf(":%v", config.ExposePort),
			Handler:           metricsRegistry.Handler(),
			ReadHeaderTimeout: 10 * time.Second,
		}
		go server.ListenAndServe()
	})
	return metricsRegistry, nil
}

func initPromRegistry() *prometheus.Registry {
	promRegistry := prometheus.NewRegistry()

	promRegistry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return promRegistry
}

func Histogram(name string) registry.HistogramAccessor {
	if metricsRegistry == nil {
		return &registry.DummyHistogramAccessor{}
	}
	return metricsRegistry.Histogram(name)
}

func Summary(name string) registry.SummaryAccessor {
	if metricsRegistry == nil {
		return &registry.DummySummaryAccessor{}
	}
	return metricsRegistry.Summary(name)
}

func Gauge(name string) registry.GaugeAccessor {
	if metricsRegistry == nil {
		return &registry.DummyGaugeAccessor{}
	}
	return metricsRegistry.Gauge(name)
}

func Counter(name string) registry.CounterAccessor {
	if metricsRegistry == nil {
		return &registry.DummyCounterAccessor{}
	}
	return metricsRegistry.Counter(name)
}

func Handler() http.Handler {
	if metricsRegistry == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This handler does nothing
		})
	}
	return metricsRegistry.Handler()
}

func HistogramVec(name string) registry.HistogramVecAccessor {
	if metricsRegistry == nil {
		return &registry.DummyHistogramVecAccessor{}
	}
	return metricsRegistry.HistogramVec(name)
}

func SummaryVec(name string) registry.SummaryVecAccessor {
	if metricsRegistry == nil {
		return &registry.DummySummaryVecAccessor{}
	}
	return metricsRegistry.SummaryVec(name)
}

func GaugeVec(name string) registry.GaugeVecAccessor {
	if metricsRegistry == nil {
		return &registry.DummyGaugeVecAccessor{}
	}
	return metricsRegistry.GaugeVec(name)
}

func CounterVec(name string) registry.CounterVecAccessor {
	if metricsRegistry == nil {
		return &registry.DummyCounterVecAccessor{}
	}
	return metricsRegistry.CounterVec(name)
}

func Pushable() MetricsRegistrar {
	if metricsPush == nil {
		return &DummyMetricsRegistrar{}
	}
	return metricsPush.GetRegistry()
}

func Push() error {
	if metricsPush == nil {
		return errors.New("metrics pusher is not initialized")
	}
	return metricsPush.Push()
}

func RegisterCollector(c prometheus.Collector) error {
	if metricsRegistry == nil {
		return errors.New("metrics registry is not initialized")
	}
	if err := metricsRegistry.(*MetricsRegistry).promRegistry.Register(c); err != nil {
		return fmt.Errorf("error registering collector: %v", err)
	}
	return nil
}

func GetPrometheusRegistry() *prometheus.Registry {
	if metricsRegistry == nil {
		return prometheus.NewRegistry()
	}
	return metricsRegistry.GetPrometheusRegistry()
}

func Registry(name string) MetricsRegistrar {
	registry, _ := extraRegistries.LoadOrStore(name, newMetricsRegistry(prometheus.NewRegistry()))
	return registry.(*MetricsRegistry)
}

func BuildMetricsString() (string, error) {
	if metricsRegistry == nil {
		return "", errors.New("metrics registry is not initialized")
	}
	return metricsRegistry.BuildMetricsString()
}

// SetMock - sets the mock metrics registry for testing purposes.
// For easy testing using prepared mocks use testing.SetMock(t) under the metrics/testing package.
func SetMock(registry MetricsRegistrar) {
	metricsRegistry = registry
}
