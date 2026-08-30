package metricsctr

import (
	"github.com/run-ai/runai/runai-common-packages/metrics"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

func Init() {
	ctrlmetrics.Registry = metrics.GetPrometheusRegistry()
}
