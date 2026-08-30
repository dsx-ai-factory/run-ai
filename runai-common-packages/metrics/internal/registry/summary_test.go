package registry

import (
	"testing"

	"github.com/run-ai/runai/runai-common-packages/metrics/internal/registry/mocks"
	"github.com/run-ai/runai/runai-common-packages/metrics/internal/runaimetrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSummary_RegistryTwiceReturnsSame(t *testing.T) {
	prm := mocks.NewRegisterer(t)
	prm.On("MustRegister", mock.Anything)

	registry := NewSummaryRegistry(prm)
	m1 := registry.Register("test_summary", "description", make(map[string]string), make(map[float64]float64))
	m2 := registry.Register("test_summary", "description", make(map[string]string), make(map[float64]float64))
	assert.Same(t, m1, m2)
}

func TestSummary_GetFromRegistryNotNil(t *testing.T) {
	prm := mocks.NewRegisterer(t)
	prm.On("MustRegister", mock.Anything)

	registry := NewSummaryRegistry(prm)
	registry.Register("test_summary", "description", make(map[string]string), make(map[float64]float64))
	metric := registry.Get("test_summary")
	assert.NotNil(t, metric)
}

func TestSummary_GetFromRegistryNotNilWhenNotRegistered(t *testing.T) {
	prm := mocks.NewRegisterer(t)

	registry := NewSummaryRegistry(prm)
	metric := registry.Get("test_summary")
	assert.Equal(t, &DummyMetricSummary{}, metric)
}

func TestSummary_GetFromRegistrySame(t *testing.T) {
	prm := mocks.NewRegisterer(t)
	prm.On("MustRegister", mock.Anything)

	registry := NewSummaryRegistry(prm)
	registry.Register("test_summary2", "description", make(map[string]string), make(map[float64]float64))
	metric := registry.Get("test_summary2")
	promMetric := metric.(*runaimetrics.Summary)

	assert.NotNil(t, metric)
	desc := promMetric.PromSummary.Desc().String()
	assert.Contains(t, desc, `fqName: "runai_test_summary2"`)
	assert.Contains(t, desc, `help: "description"`)
}

func TestSummary_GetFromRegistrySameLabels(t *testing.T) {
	prm := mocks.NewRegisterer(t)
	prm.On("MustRegister", mock.Anything)

	registry := NewSummaryRegistry(prm)
	registry.Register("test_summary3", "description", map[string]string{"bla": "bla"}, make(map[float64]float64))
	metric := registry.Get("test_summary3")
	promMetric := metric.(*runaimetrics.Summary)

	assert.NotNil(t, metric)
	desc := promMetric.PromSummary.Desc().String()
	assert.Contains(t, desc, `fqName: "runai_test_summary3"`)
	assert.Contains(t, desc, `help: "description"`)
	assert.Contains(t, desc, `constLabels: {bla="bla"}`)
}
