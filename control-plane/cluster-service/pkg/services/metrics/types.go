package metrics

import "github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"

// MetricsResponseInterface defines a common interface for both GroupedMetricsResponse and MetricsResponse
type MetricsResponseInterface interface {
	GetMeasurements() []MetricsMeasurementInterface
}

// MetricsMeasurementInterface defines a common interface for both GroupedMeasurementResponse and MeasurementResponse
type MetricsMeasurementInterface interface {
	GetType() string
	GetValues() *[]api.GroupedMeasurementResponseValuesInner
}

// GroupedMetricsResponseWrapper wraps api.GroupedMetricsResponse to implement MetricsResponseInterface
type GroupedMetricsResponseWrapper struct {
	*api.GroupedMetricsResponse
}

// GetMeasurements implements MetricsResponseInterface for GroupedMetricsResponseWrapper
func (g *GroupedMetricsResponseWrapper) GetMeasurements() []MetricsMeasurementInterface {
	measurements := make([]MetricsMeasurementInterface, len(g.Measurements))
	for i, m := range g.Measurements {
		measurements[i] = &GroupedMeasurementResponseWrapper{&m}
	}
	return measurements
}

// GroupedMeasurementResponseWrapper wraps api.GroupedMeasurementResponse to implement MetricsMeasurementInterface
type GroupedMeasurementResponseWrapper struct {
	*api.GroupedMeasurementResponse
}

// GetType implements MetricsMeasurementInterface for GroupedMeasurementResponseWrapper
func (g *GroupedMeasurementResponseWrapper) GetType() string {
	return g.Type
}

// GetValues implements MetricsMeasurementInterface for GroupedMeasurementResponseWrapper
func (g *GroupedMeasurementResponseWrapper) GetValues() *[]api.GroupedMeasurementResponseValuesInner {
	return g.Values
}

// MetricsResponseWrapper wraps api.MetricsResponse to implement MetricsResponseInterface
type MetricsResponseWrapper struct {
	*api.MetricsResponse
}

// GetMeasurements implements MetricsResponseInterface for MetricsResponseWrapper
func (m *MetricsResponseWrapper) GetMeasurements() []MetricsMeasurementInterface {
	measurements := make([]MetricsMeasurementInterface, len(m.Measurements))
	for i, measurement := range m.Measurements {
		measurements[i] = &MeasurementResponseWrapper{&measurement}
	}
	return measurements
}

// MeasurementResponseWrapper wraps api.MeasurementResponse to implement MetricsMeasurementInterface
type MeasurementResponseWrapper struct {
	*api.MeasurementResponse
}

// GetType implements MetricsMeasurementInterface for MeasurementResponseWrapper
func (m *MeasurementResponseWrapper) GetType() string {
	return m.Type
}

// GetValues implements MetricsMeasurementInterface for MeasurementResponseWrapper
func (m *MeasurementResponseWrapper) GetValues() *[]api.GroupedMeasurementResponseValuesInner {
	return m.Values
}
