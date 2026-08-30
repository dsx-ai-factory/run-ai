package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	health = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "metrics_health_total",
			Help: "Health metric - temporary",
		},
	)
)

func InitMetrics() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(health)
	metrics.Registry.MustRegister(podGpuInfo)

	health.Inc()
}
