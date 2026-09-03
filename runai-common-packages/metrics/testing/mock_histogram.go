package testing

import (
	am "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/internal/registry/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/runaimetrics"
	metricsmocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/runaimetrics/mocks"
)

// -------------- HISTOGRAM ----------------
type mMHistogramAccessor struct {
	t   TestingT
	mha *am.MockHistogramAccessor
}

func (d *mMHistogramAccessor) Register(description string, labels map[string]string, buckets []float64) runaimetrics.MetricHistogram {
	m := metricsmocks.NewMockMetricHistogram(d.t)
	d.mha.On("Register", description, labels, buckets).Return(m).Once()
	return &mMMetricHistogram{mmh: m}
}

func (d *mMHistogramAccessor) Get() runaimetrics.MetricHistogram {
	m := metricsmocks.NewMockMetricHistogram(d.t)
	d.mha.On("Get").Return(m).Once()
	return &mMMetricHistogram{mmh: m}
}

type mMMetricHistogram struct {
	mmh *metricsmocks.MockMetricHistogram
}

func (d *mMMetricHistogram) Observe(value float64) {
	d.mmh.On("Observe", value).Return(nil).Once()
}

// -------------- HISTOGRAM VEC ----------------
type mMHistogramVecAccessor struct {
	t    TestingT
	mhva *am.MockHistogramVecAccessor
}

func (d *mMHistogramVecAccessor) Register(description string, labels []string, buckets []float64) runaimetrics.HistogramVector {
	m := metricsmocks.NewMockHistogramVector(d.t)
	d.mhva.On("Register", description, labels, buckets).Return(m).Once()
	return &mMMetricHistogramVec{t: d.t, mmhv: m}
}

func (d *mMHistogramVecAccessor) Get() runaimetrics.HistogramVector {
	m := metricsmocks.NewMockHistogramVector(d.t)
	d.mhva.On("Get").Return(m).Once()
	return &mMMetricHistogramVec{t: d.t, mmhv: m}
}

type mMMetricHistogramVec struct {
	t    TestingT
	mmhv *metricsmocks.MockHistogramVector
}

func (d *mMMetricHistogramVec) WithLabels(labels map[string]string) runaimetrics.MetricHistogram {
	m := metricsmocks.NewMockMetricHistogram(d.t)
	d.mmhv.On("WithLabels", labels).Return(m).Once()
	return &mMMetricHistogram{mmh: m}
}

func (d *mMMetricHistogramVec) Reset() {
	d.mmhv.On("Reset").Return(nil).Once()
}
