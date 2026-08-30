package client

import (
	"strings"
)

type OpenMetricsStringer interface {
	String() string
}

type metricCollection struct {
	metrics []OpenMetricsStringer
}

func NewMetricCollection() *metricCollection {
	return &metricCollection{metrics: make([]OpenMetricsStringer, 0)}
}

func (b *metricCollection) AddMetric(metric OpenMetricsStringer) *metricCollection {
	b.metrics = append(b.metrics, metric)
	return b
}

func (b *metricCollection) ToOpenMetrics() string {
	sb := strings.Builder{}
	for _, metric := range b.metrics {
		sb.WriteString(metric.String())
	}
	sb.WriteString("# EOF")
	return sb.String()
}
