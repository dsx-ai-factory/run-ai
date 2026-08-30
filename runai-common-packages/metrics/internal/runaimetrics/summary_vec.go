package runaimetrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/run-ai/runai/runai-common-packages/metrics/runaimetrics"
)

type SummaryVec struct {
	PromSummaryVec *prometheus.SummaryVec
}

func NewPromSummaryVec(name string, description string, labels []string, objectives map[float64]float64) *SummaryVec {
	summaryVec := SummaryVec{
		PromSummaryVec: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name:       METRICS_PREFIX + name,
			Help:       description,
			Objectives: objectives,
		}, labels),
	}
	return &summaryVec
}

func (cm *SummaryVec) WithLabels(labels map[string]string) runaimetrics.MetricSummary {
	return &Summary{
		PromSummary: cm.PromSummaryVec.With(labels).(prometheus.Summary),
	}
}

func (cm *SummaryVec) Reset() {
	cm.PromSummaryVec.Reset()
}
