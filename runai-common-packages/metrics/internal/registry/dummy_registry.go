package registry

import (
	"github.com/run-ai/runai/runai-common-packages/metrics/runaimetrics"
)

// -------------- HISTOGRAM ----------------
type DummyHistogramAccessor struct{}

func (d *DummyHistogramAccessor) Register(description string, labels map[string]string, buckets []float64) runaimetrics.MetricHistogram {
	return &DummyMetricHistogram{}
}

func (d *DummyHistogramAccessor) Get() runaimetrics.MetricHistogram {
	return &DummyMetricHistogram{}
}

type DummyMetricHistogram struct{}

func (d *DummyMetricHistogram) Observe(value float64) {
	// Do nothing
}

type DummyHistogramVecAccessor struct{}

func (d *DummyHistogramVecAccessor) Register(description string, labels []string, buckets []float64) runaimetrics.HistogramVector {
	return &DummyHistogramVector{}
}

func (d *DummyHistogramVecAccessor) Get() runaimetrics.HistogramVector {
	return &DummyHistogramVector{}
}

type DummyHistogramVector struct{}

func (d *DummyHistogramVector) WithLabels(labels map[string]string) runaimetrics.MetricHistogram {
	return &DummyMetricHistogram{}
}

func (d *DummyHistogramVector) Reset() {
}

// -------------- GAUGE ----------------
type DummyGaugeAccessor struct{}

func (d *DummyGaugeAccessor) Register(description string, labels map[string]string) runaimetrics.MetricGauge {
	return &DummyMetricGauge{}
}

func (d *DummyGaugeAccessor) Get() runaimetrics.MetricGauge {
	return &DummyMetricGauge{}
}

type DummyMetricGauge struct{}

func (d *DummyMetricGauge) Set(value float64) {
	// Do nothing
}

func (d *DummyMetricGauge) Inc() {
	// Do nothing
}

func (d *DummyMetricGauge) Dec() {
	// Do nothing
}

func (d *DummyMetricGauge) Add(value float64) {
	// Do nothing
}

func (d *DummyMetricGauge) Sub(value float64) {
	// Do nothing
}

type DummyGaugeVecAccessor struct{}

func (d *DummyGaugeVecAccessor) Register(description string, labels []string) runaimetrics.GaugeVector {
	return &DummyGaugeVector{}
}

func (d *DummyGaugeVecAccessor) Get() runaimetrics.GaugeVector {
	return &DummyGaugeVector{}
}

type DummyGaugeVector struct{}

func (d *DummyGaugeVector) WithLabels(labels map[string]string) runaimetrics.MetricGauge {
	return &DummyMetricGauge{}
}

func (d *DummyGaugeVector) Reset() {
}

// -------------- COUNTER ----------------
type DummyCounterAccessor struct{}

func (d *DummyCounterAccessor) Register(description string, labels map[string]string) runaimetrics.MetricCounter {
	return &DummyMetricCounter{}
}

func (d *DummyCounterAccessor) Get() runaimetrics.MetricCounter {
	return &DummyMetricCounter{}
}

type DummyMetricCounter struct{}

func (d *DummyMetricCounter) Inc() {
	// Do nothing
}

func (d *DummyMetricCounter) Add(value float64) {
	// Do nothing
}

type DummyCounterVecAccessor struct{}

func (d *DummyCounterVecAccessor) Register(description string, labels []string) runaimetrics.CounterVector {
	return &DummyCounterVector{}
}

func (d *DummyCounterVecAccessor) Get() runaimetrics.CounterVector {
	return &DummyCounterVector{}
}

type DummyCounterVector struct{}

func (d *DummyCounterVector) WithLabels(labels map[string]string) runaimetrics.MetricCounter {
	return &DummyMetricCounter{}
}

func (d *DummyCounterVector) Reset() {
}

// -------------- SUMMARY ----------------
type DummySummaryAccessor struct{}

func (d *DummySummaryAccessor) Register(description string, labels map[string]string, objectives map[float64]float64) runaimetrics.MetricSummary {
	return &DummyMetricSummary{}
}

func (d *DummySummaryAccessor) Get() runaimetrics.MetricSummary {
	return &DummyMetricSummary{}
}

type DummyMetricSummary struct{}

func (d *DummyMetricSummary) Observe(value float64) {
	// Do nothing
}

type DummySummaryVecAccessor struct{}

func (d *DummySummaryVecAccessor) Register(description string, labels []string, objectives map[float64]float64) runaimetrics.SummaryVector {
	return &dummySummaryVector{}
}

func (d *DummySummaryVecAccessor) Get() runaimetrics.SummaryVector {
	return &dummySummaryVector{}
}

type dummySummaryVector struct{}

func (d *dummySummaryVector) WithLabels(labels map[string]string) runaimetrics.MetricSummary {
	return &DummyMetricSummary{}
}

func (d *dummySummaryVector) Reset() {
}
