package testing

import (
	"fmt"
	"math"
	"time"

	openapi "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"
)

type MetricsMockValue struct {
	Project    string
	Department string
	NodePool   string
	Value      string
}

func GetMetricCurrentMockResult(values []MetricsMockValue) openapi.GetMetrics200Response {
	data := []openapi.Result{}

	for _, mockValue := range values {
		measurements := []openapi.Measurement{{
			Timestamp: time.Now(),
			Value:     mockValue.Value,
		}}
		data = append(data, openapi.Result{
			Measurements: measurements,
			MetricLabels: map[string]string{"nodepool": mockValue.NodePool, "department": mockValue.Department, "project": mockValue.Project},
		})
	}

	return openapi.GetMetrics200Response{
		Status:  "success",
		Query:   "",
		IsRange: true,
		Data:    data,
	}
}

func GetTimerangeFloatMetricsMock(baseValue float32, baseTimestampUnix float64, numOfSamples int, query string) openapi.GetMetrics200Response {
	value := baseValue
	measurements := []openapi.Measurement{}

	for i := 0; i < numOfSamples; i++ {
		measurements = append(measurements, openapi.Measurement{
			Timestamp: time.UnixMilli(int64(baseTimestampUnix) + int64(i*1000)),
			Value:     fmt.Sprint(value),
		})
		value = float32(math.Round(float64(value+0.1)*100) / 100)
	}

	data := []openapi.Result{{
		Measurements: measurements,
	}}

	return openapi.GetMetrics200Response{
		Status:  "success",
		Query:   query,
		IsRange: true,
		Data:    data,
	}
}

func GetTimerangeIntMetricsMock(value int32, baseTimestampUnix float64, numOfSamples int, query string) openapi.GetMetrics200Response {
	measurements := []openapi.Measurement{}

	for i := 0; i < numOfSamples; i++ {
		measurements = append(measurements, openapi.Measurement{
			Timestamp: time.UnixMilli(int64(baseTimestampUnix) + int64(i*1000)),
			Value:     fmt.Sprint(value),
		})
	}

	data := []openapi.Result{{
		Measurements: measurements,
	}}

	return openapi.GetMetrics200Response{
		Status:  "success",
		Query:   query,
		IsRange: true,
		Data:    data,
	}
}
