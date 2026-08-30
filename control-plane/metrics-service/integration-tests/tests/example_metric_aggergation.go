package tests

type ExampleMetricAggregationFilters struct {
	ClusterId []string `mapstructure:"clusterId,omitempty"`
}

type ExampleMetricAggregationResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type ExampleMetricAggregation struct {
	Filters        ExampleMetricAggregationFilters
	ResponseLabels ExampleMetricAggregationResponseLabels
	Quantile       *float32
}

func (m ExampleMetricAggregation) GetName() string {
	return "example_metric_aggregation"
}

func (m ExampleMetricAggregation) GetFilters() interface{} {
	return m.Filters
}

func (m ExampleMetricAggregation) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m ExampleMetricAggregation) GetQuantile() *float32 {
	return m.Quantile
}

func (m ExampleMetricAggregation) IsAggregation() bool {
	return true
}

func (m ExampleMetricAggregation) IsQuantile() bool {
	return false
}
