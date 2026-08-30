package registry

import (
	"testing"

	"github.com/run-ai/runai/runai-common-packages/metrics/internal/registry/mocks"
	"github.com/run-ai/runai/runai-common-packages/metrics/internal/runaimetrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHistogram_RegistryTwiceReturnsSame(t *testing.T) {
	prm := mocks.NewRegisterer(t)
	prm.On("MustRegister", mock.Anything)

	registry := NewHistogramRegistry(prm)
	m1 := registry.Register("test_histogram", "description", make(map[string]string), []float64{0.1, 0.2, 0.5, 1.0})
	m2 := registry.Register("test_histogram", "description", make(map[string]string), []float64{0.1, 0.2, 0.5, 1.0})
	assert.Same(t, m1, m2)
}

func TestHistogram_GetFromRegistryNotNil(t *testing.T) {
	prm := mocks.NewRegisterer(t)
	prm.On("MustRegister", mock.Anything)

	registry := NewHistogramRegistry(prm)
	registry.Register("test_histogram", "description", make(map[string]string), []float64{0.1, 0.2, 0.5, 1.0})
	metric := registry.Get("test_histogram")
	assert.NotNil(t, metric)
}

func TestHistogram_GetFromRegistryNotNilWhenNotRegistered(t *testing.T) {
	prm := mocks.NewRegisterer(t)

	registry := NewHistogramRegistry(prm)
	metric := registry.Get("test_histogram")
	assert.Equal(t, &DummyMetricHistogram{}, metric)
}

func TestHistogram_GetFromRegistrySame(t *testing.T) {
	prm := mocks.NewRegisterer(t)
	prm.On("MustRegister", mock.Anything)

	registry := NewHistogramRegistry(prm)
	registry.Register("test_histogram2", "description", make(map[string]string), []float64{0.1, 0.2, 0.5, 1.0})
	metric := registry.Get("test_histogram2")
	promMetric := metric.(*runaimetrics.Histogram)

	assert.NotNil(t, metric)
	desc := promMetric.PromHistogram.Desc().String()
	assert.Contains(t, desc, `fqName: "runai_test_histogram2"`)
	assert.Contains(t, desc, `help: "description"`)
}

func TestHistogram_GetFromRegistrySameLabels(t *testing.T) {
	prm := mocks.NewRegisterer(t)
	prm.On("MustRegister", mock.Anything)

	registry := NewHistogramRegistry(prm)
	registry.Register("test_histogram3", "description", map[string]string{"bla": "bla"}, []float64{0.1, 0.2, 0.5, 1.0})
	metric := registry.Get("test_histogram3")
	promMetric := metric.(*runaimetrics.Histogram)

	assert.NotNil(t, metric)
	desc := promMetric.PromHistogram.Desc().String()
	assert.Contains(t, desc, `fqName: "runai_test_histogram3"`)
	assert.Contains(t, desc, `help: "description"`)
	assert.Contains(t, desc, `constLabels: {bla="bla"}`)
}
