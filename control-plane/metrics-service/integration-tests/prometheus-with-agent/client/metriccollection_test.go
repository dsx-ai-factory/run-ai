package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const expected = "my_metric{code=\"200\"} 4.400000 123\n" +
	"my_metric2{code=\"500\",another=\"stam_label\"} 4.400000 457\n" +
	"my_metric_histogram_bucket{clusterId=\"cluster_1\",le=\"2\"} 0.000000 457\n" +
	"my_metric_histogram_bucket{clusterId=\"cluster_1\",le=\"+Inf\"} 1.000000 457\n" +
	"my_metric_histogram_sum{clusterId=\"cluster_1\"} 3.000000 457\n" +
	"my_metric_histogram_count{clusterId=\"cluster_1\"} 1.000000 457\n" +
	"# EOF"

func TestMetricCollection(t *testing.T) {
	mc := NewMetricCollection()
	mc.AddMetric(NewGaugeMetric("my_metric", 4.4).
		WithTime(time.Unix(123, 0)).
		AddLabel("code", "200"))
	mc.AddMetric(NewGaugeMetric("my_metric2", 4.4).
		WithTime(time.Unix(457, 0)).
		AddLabel("code", "500").
		AddLabel("another", "stam_label"))
	mc.AddMetric(NewHistogramMetric("my_metric_histogram", []float64{2}).
		AddLabel("clusterId", "cluster_1").
		WithTime(time.Unix(457, 0)).
		Observe(3))

	metrics := mc.ToOpenMetrics()
	assert.Equal(t, expected, metrics)
}
