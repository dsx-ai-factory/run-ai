package testing

import (
	"github.com/run-ai/runai/runai-common-packages/metrics/internal/registry"
	am "github.com/run-ai/runai/runai-common-packages/metrics/internal/registry/mocks"
	"github.com/run-ai/runai/runai-common-packages/metrics/mocks"
)

type mockManager struct {
	mmr *mocks.MockMetricsRegistrar
	t   TestingT
}

func newMockManager(m *mocks.MockMetricsRegistrar, t TestingT) *mockManager {
	return &mockManager{
		mmr: m,
		t:   t,
	}
}

func (mm *mockManager) Gauge(name string) registry.GaugeAccessor {
	m := am.NewMockGaugeAccessor(mm.t)
	mm.mmr.On("Gauge", name).Return(m).Once()
	return &mMGaugeAccessor{mga: m, t: mm.t}
}

func (mm *mockManager) Counter(name string) registry.CounterAccessor {
	m := am.NewMockCounterAccessor(mm.t)
	mm.mmr.On("Counter", name).Return(m).Once()
	return &mMCounterAccessor{mca: m, t: mm.t}
}

func (mm *mockManager) Histogram(name string) registry.HistogramAccessor {
	m := am.NewMockHistogramAccessor(mm.t)
	mm.mmr.On("Histogram", name).Return(m).Once()
	return &mMHistogramAccessor{mha: m, t: mm.t}
}

func (mm *mockManager) Summary(name string) registry.SummaryAccessor {
	m := am.NewMockSummaryAccessor(mm.t)
	mm.mmr.On("Summary", name).Return(m).Once()
	return &mMSummaryAccessor{msa: m, t: mm.t}
}

func (mm *mockManager) HistogramVec(name string) registry.HistogramVecAccessor {
	m := am.NewMockHistogramVecAccessor(mm.t)
	mm.mmr.On("HistogramVec", name).Return(m).Once()
	return &mMHistogramVecAccessor{mhva: m, t: mm.t}
}

func (mm *mockManager) CounterVec(name string) registry.CounterVecAccessor {
	m := am.NewMockCounterVecAccessor(mm.t)
	mm.mmr.On("CounterVec", name).Return(m).Once()
	return &mMCounterVecAccessor{mcva: m, t: mm.t}
}

func (mm *mockManager) GaugeVec(name string) registry.GaugeVecAccessor {
	m := am.NewMockGaugeVecAccessor(mm.t)
	mm.mmr.On("GaugeVec", name).Return(m).Once()
	return &mMGaugeVecAccessor{mgva: m, t: mm.t}
}

func (mm *mockManager) SummaryVec(name string) registry.SummaryVecAccessor {
	m := am.NewMockSummaryVecAccessor(mm.t)
	mm.mmr.On("SummaryVec", name).Return(m).Once()
	return &mMSummaryVecAccessor{msva: m, t: mm.t}
}

func (mm *mockManager) GetMetricsRegistrar() *mocks.MockMetricsRegistrar {
	return mm.mmr
}
