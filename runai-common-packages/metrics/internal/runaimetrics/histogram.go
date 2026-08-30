package runaimetrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Histogram struct {
	PromHistogram prometheus.Histogram
}

func NewPromHistogram(name string, description string, labels map[string]string, buckets []float64) *Histogram {
	histogramMetric := Histogram{
		PromHistogram: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:        METRICS_PREFIX + name,
			Help:        description,
			ConstLabels: labels,
			Buckets:     buckets,
		}),
	}
	return &histogramMetric
}

func (hm *Histogram) Observe(value float64) {
	hm.PromHistogram.Observe(value)
}
