package registry

import (
	"testing"

	"github.com/run-ai/runai/runai-common-packages/metrics/internal/registry/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCounterVec_RegistryTwiceReturnsSame(t *testing.T) {
	prm := mocks.NewRegisterer(t)
	prm.On("MustRegister", mock.Anything)

	registry := NewCounterVecRegistry(prm)
	m1 := registry.Register("test_Counter", "description", make([]string, 0))
	m2 := registry.Register("test_Counter", "description", make([]string, 0))
	assert.Same(t, m1, m2)
}

func TestCounterVec_GetFromRegistryNotNilWhenRegistered(t *testing.T) {
	prm := mocks.NewRegisterer(t)
	prm.On("MustRegister", mock.Anything)

	registry := NewCounterVecRegistry(prm)
	registry.Register("test_Counter", "description", make([]string, 0))
	metric := registry.Get("test_Counter")
	assert.NotNil(t, metric)
}

func TestCounterVec_GetFromRegistryNotNilWithoutRegister(t *testing.T) {
	prm := mocks.NewRegisterer(t)

	registry := NewCounterVecRegistry(prm)
	metric := registry.Get("test_Counter")
	assert.Equal(t, &DummyCounterVector{}, metric)
}
