package models

type MetricRecord struct {
	Labels map[string]string
	Value  float64
}
