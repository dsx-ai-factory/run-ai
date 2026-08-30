package testing

import (
	am "github.com/run-ai/runai/runai-common-packages/metrics/internal/registry/mocks"
	"github.com/run-ai/runai/runai-common-packages/metrics/runaimetrics"
	metricsmocks "github.com/run-ai/runai/runai-common-packages/metrics/runaimetrics/mocks"
)

// -------------- GAUGE ----------------
type mMGaugeAccessor struct {
	t   TestingT
	mga *am.MockGaugeAccessor
}

func (d *mMGaugeAccessor) Register(description string, labels map[string]string) runaimetrics.MetricGauge {
	m := metricsmocks.NewMockMetricGauge(d.t)
	d.mga.On("Register", description, labels).Return(m).Once()
	return &mMMetricGauge{mmg: m}
}

func (d *mMGaugeAccessor) Get() runaimetrics.MetricGauge {
	m := metricsmocks.NewMockMetricGauge(d.t)
	d.mga.On("Get").Return(m).Once()
	return &mMMetricGauge{mmg: m}
}

type mMMetricGauge struct {
	mmg *metricsmocks.MockMetricGauge
}

func (d *mMMetricGauge) Set(value float64) {
	d.mmg.On("Set", value).Return(nil).Once()
}

func (d *mMMetricGauge) Inc() {
	d.mmg.On("Inc").Return(nil).Once()
}

func (d *mMMetricGauge) Dec() {
	d.mmg.On("Dec").Return(nil).Once()
}

func (d *mMMetricGauge) Add(value float64) {
	d.mmg.On("Add", value).Return(nil).Once()
}

func (d *mMMetricGauge) Sub(value float64) {
	d.mmg.On("Sub", value).Return(nil).Once()
}

// -------------- GAUGE VEC ----------------
type mMGaugeVecAccessor struct {
	t    TestingT
	mgva *am.MockGaugeVecAccessor
}

func (d *mMGaugeVecAccessor) Register(description string, labels []string) runaimetrics.GaugeVector {
	m := metricsmocks.NewMockGaugeVector(d.t)
	d.mgva.On("Register", description, labels).Return(m).Once()
	return &mMMetricGaugeVec{t: d.t, mmgv: m}
}

func (d *mMGaugeVecAccessor) Get() runaimetrics.GaugeVector {
	m := metricsmocks.NewMockGaugeVector(d.t)
	d.mgva.On("Get").Return(m).Once()
	return &mMMetricGaugeVec{t: d.t, mmgv: m}
}

type mMMetricGaugeVec struct {
	t    TestingT
	mmgv *metricsmocks.MockGaugeVector
}

func (d *mMMetricGaugeVec) WithLabels(labels map[string]string) runaimetrics.MetricGauge {
	m := metricsmocks.NewMockMetricGauge(d.t)
	d.mmgv.On("WithLabels", labels).Return(m).Once()
	return &mMMetricGauge{mmg: m}
}

func (d *mMMetricGaugeVec) Reset() {
	d.mmgv.On("Reset").Return(nil).Once()
}
