package runaimetrics

type GaugeVector interface {
	WithLabels(labels map[string]string) MetricGauge
	Reset()
}

type CounterVector interface {
	WithLabels(labels map[string]string) MetricCounter
	Reset()
}

type HistogramVector interface {
	WithLabels(labels map[string]string) MetricHistogram
	Reset()
}

type SummaryVector interface {
	WithLabels(labels map[string]string) MetricSummary
	Reset()
}
