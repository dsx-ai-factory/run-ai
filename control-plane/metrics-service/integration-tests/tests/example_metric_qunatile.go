package tests

type ExampleMetricQuantileFilters struct {
	ClusterId []string `mapstructure:"clusterId,omitempty"`
}

type ExampleMetricQuantileResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type ExampleMetricQuantile struct {
	Filters        ExampleMetricQuantileFilters
	ResponseLabels ExampleMetricQuantileResponseLabels
	Quantile       *float32
}

func (m ExampleMetricQuantile) GetName() string {
	return "example_metric_quantile"
}

func (m ExampleMetricQuantile) GetFilters() interface{} {
	return m.Filters
}

func (m ExampleMetricQuantile) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m ExampleMetricQuantile) GetQuantile() *float32 {
	return m.Quantile
}

func (m ExampleMetricQuantile) IsAggregation() bool {
	return true
}

func (m ExampleMetricQuantile) IsQuantile() bool {
	return true
}
