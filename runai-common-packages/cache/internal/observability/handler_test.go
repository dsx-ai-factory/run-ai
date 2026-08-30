package observability

import (
	"context"
	"errors"
	"testing"

	cache_config "github.com/run-ai/runai/runai-common-packages/cache/v2/config"
	"github.com/run-ai/runai/runai-common-packages/tracing"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewObservabilityHandler(t *testing.T) {
	handler := NewObservabilityHandler()
	require.NotNil(t, handler)
	require.NotNil(t, handler.metricsWriter)
}

func TestWithObservability_OperationSucceeded(t *testing.T) {
	initTracing()

	mockMetricsWriter := NewMockMetricsWriterer(t)
	mockMetricsWriter.On("RecordOperation", string(cache_config.CTRedis), "storeTest", OTSet, mock.Anything, nil, false).Return()

	handler := observabilityHandler{
		cacheType:     string(cache_config.CTRedis),
		metricsWriter: mockMetricsWriter,
	}
	ctx := context.Background()
	err := handler.WithObservability(ctx, "storeTest", "", OTSet, func(ctx context.Context) (bool, error) {
		return false, nil
	})
	require.NoError(t, err)
}

func TestWithObservability_WithCacheMiss(t *testing.T) {
	initTracing()

	mockMetricsWriter := NewMockMetricsWriterer(t)
	mockMetricsWriter.On("RecordOperation", string(cache_config.CTRedis), "storeTest", OTSet, mock.Anything, nil, true).Return()

	handler := observabilityHandler{
		cacheType:     string(cache_config.CTRedis),
		metricsWriter: mockMetricsWriter,
	}
	ctx := context.Background()
	err := handler.WithObservability(ctx, "storeTest", "", OTSet, func(ctx context.Context) (bool, error) {
		return true, nil
	})
	require.NoError(t, err)
}

func TestWithObservability_OperationFailed(t *testing.T) {
	initTracing()

	mockMetricsWriter := NewMockMetricsWriterer(t)
	mockMetricsWriter.On("RecordOperation", string(cache_config.CTRedis), "storeTest", OTSet, mock.Anything, errors.New("operation failed"), false).Return()

	handler := observabilityHandler{
		cacheType:     string(cache_config.CTRedis),
		metricsWriter: mockMetricsWriter,
	}
	ctx := context.Background()
	err := handler.WithObservability(ctx, "storeTest", "", OTSet, func(ctx context.Context) (bool, error) {
		return false, errors.New("operation failed")
	})
	require.Equal(t, "operation failed", err.Error())
}

func initTracing() {
	tracing.Init(&tracing.Config{
		ServiceName:       "test-service",
		UnderlineTracer:   "otel",
		CollectorEndpoint: "localhost:4317",
		Enabled:           true,
	})
}
