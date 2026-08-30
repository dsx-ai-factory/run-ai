package testing

import (
	"github.com/run-ai/runai/runai-common-packages/metrics"
	"github.com/run-ai/runai/runai-common-packages/metrics/mocks"
	"github.com/stretchr/testify/mock"
)

type TestingT interface {
	mock.TestingT
	Cleanup(func())
}

func SetMock(t TestingT) *mockManager {
	m := mocks.NewMockMetricsRegistrar(t)
	mm := newMockManager(m, t)
	metrics.SetMock(m)
	return mm
}
