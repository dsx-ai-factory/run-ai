package testing

import (
	am "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/internal/registry/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/runaimetrics"
	metricsmocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/runaimetrics/mocks"
)

// -------------- COUNTER ----------------
type mMCounterAccessor struct {
	t   TestingT
	mca *am.MockCounterAccessor
}

func (d *mMCounterAccessor) Register(description string, labels map[string]string) runaimetrics.MetricCounter {
	m := metricsmocks.NewMockMetricCounter(d.t)
	d.mca.On("Register", description, labels).Return(m).Once()
	return &mMMetricCounter{mmc: m}
}

func (d *mMCounterAccessor) Get() runaimetrics.MetricCounter {
	m := metricsmocks.NewMockMetricCounter(d.t)
	d.mca.On("Get").Return(m).Once()
	return &mMMetricCounter{mmc: m}
}

type mMMetricCounter struct {
	mmc *metricsmocks.MockMetricCounter
}

func (d *mMMetricCounter) Inc() {
	d.mmc.On("Inc").Return(nil).Once()
}

func (d *mMMetricCounter) Add(value float64) {
	d.mmc.On("Add", value).Return(nil).Once()
}

// -------------- COUNTER VEC ----------------
type mMCounterVecAccessor struct {
	t    TestingT
	mcva *am.MockCounterVecAccessor
}

func (d *mMCounterVecAccessor) Register(description string, labels []string) runaimetrics.CounterVector {
	m := metricsmocks.NewMockCounterVector(d.t)
	d.mcva.On("Register", description, labels).Return(m).Once()
	return &mMMetricCounterVec{t: d.t, mmcv: m}
}

func (d *mMCounterVecAccessor) Get() runaimetrics.CounterVector {
	m := metricsmocks.NewMockCounterVector(d.t)
	d.mcva.On("Get").Return(m).Once()
	return &mMMetricCounterVec{t: d.t, mmcv: m}
}

type mMMetricCounterVec struct {
	t    TestingT
	mmcv *metricsmocks.MockCounterVector
}

func (d *mMMetricCounterVec) WithLabels(labels map[string]string) runaimetrics.MetricCounter {
	m := metricsmocks.NewMockMetricCounter(d.t)
	d.mmcv.On("WithLabels", labels).Return(m).Once()
	return &mMMetricCounter{mmc: m}
}

func (d *mMMetricCounterVec) Reset() {
	d.mmcv.On("Reset").Return(nil).Once()
}
