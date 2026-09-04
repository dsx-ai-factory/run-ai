package testing

import (
	am "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/internal/registry/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/runaimetrics"
	metricsmocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/runaimetrics/mocks"
)

// -------------- SUMMARY ----------------
type mMSummaryAccessor struct {
	t   TestingT
	msa *am.MockSummaryAccessor
}

func (d *mMSummaryAccessor) Register(description string, labels map[string]string, objectives map[float64]float64) runaimetrics.MetricSummary {
	m := metricsmocks.NewMockMetricSummary(d.t)
	d.msa.On("Register", description, labels, objectives).Return(m).Once()
	return &mMMetricSummary{mms: m}
}

func (d *mMSummaryAccessor) Get() runaimetrics.MetricSummary {
	m := metricsmocks.NewMockMetricSummary(d.t)
	d.msa.On("Get").Return(m).Once()
	return &mMMetricSummary{mms: m}
}

type mMMetricSummary struct {
	mms *metricsmocks.MockMetricSummary
}

func (d *mMMetricSummary) Observe(value float64) {
	d.mms.On("Observe", value).Return(nil).Once()
}

// -------------- SUMMARY VEC ----------------
type mMSummaryVecAccessor struct {
	t    TestingT
	msva *am.MockSummaryVecAccessor
}

func (d *mMSummaryVecAccessor) Register(description string, labels []string, objectives map[float64]float64) runaimetrics.SummaryVector {
	m := metricsmocks.NewMockSummaryVector(d.t)
	d.msva.On("Register", description, labels, objectives).Return(m).Once()
	return &mMMetricSummaryVec{t: d.t, mmsv: m}
}

func (d *mMSummaryVecAccessor) Get() runaimetrics.SummaryVector {
	m := metricsmocks.NewMockSummaryVector(d.t)
	d.msva.On("Get").Return(m).Once()
	return &mMMetricSummaryVec{t: d.t, mmsv: m}
}

type mMMetricSummaryVec struct {
	t    TestingT
	mmsv *metricsmocks.MockSummaryVector
}

func (d *mMMetricSummaryVec) WithLabels(labels map[string]string) runaimetrics.MetricSummary {
	m := metricsmocks.NewMockMetricSummary(d.t)
	d.mmsv.On("WithLabels", labels).Return(m).Once()
	return &mMMetricSummary{mms: m}
}

func (d *mMMetricSummaryVec) Reset() {
	d.mmsv.On("Reset").Return(nil).Once()
}
