/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reports

import (
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/generated/pkg/api"
)

var metric1 = &api.GetMetrics200Response{
	Status: successStatus,
	Data: []api.Result{
		{
			MetricLabels: map[string]string{
				"app":  "app1",
				"job":  "job1",
				"junk": "junk1",
			},
			Measurements: []api.Measurement{
				{
					Value: "1",
				},
			},
		},
		{
			MetricLabels: map[string]string{
				"app":  "app2",
				"job":  "job1",
				"junk": "junk2",
			},
			Measurements: []api.Measurement{
				{
					Value: "2",
				},
			},
		},
	},
}

var metric2 = &api.GetMetrics200Response{
	Status: successStatus,
	Data: []api.Result{
		{
			MetricLabels: map[string]string{
				"app":  "app1",
				"job":  "job1",
				"junk": "junk1",
			},
			Measurements: []api.Measurement{
				{
					Value: "3",
				},
			},
		},
		{
			MetricLabels: map[string]string{
				"app":  "app2",
				"job":  "job1",
				"junk": "junk2",
			},
			Measurements: []api.Measurement{
				{
					Value: "4",
				},
			},
		},
	},
}

var metricWithMissingSample = &api.GetMetrics200Response{
	Status: successStatus,
	Data: []api.Result{
		{
			MetricLabels: map[string]string{
				"app":  "app2",
				"job":  "job1",
				"junk": "junk2",
			},
			Measurements: []api.Measurement{
				{
					Value: "5",
				},
			},
		},
	},
}

var metricWithFailedStatus = &api.GetMetrics200Response{
	Status: "failed",
}

var metricValuesWithDecimalOverPrecision = &api.GetMetrics200Response{
	Status: successStatus,
	Data: []api.Result{
		{
			MetricLabels: map[string]string{
				"app": "app1",
			},
			Measurements: []api.Measurement{
				{
					Value: "1.12345",
				},
			},
		},
	},
}

func TestBuild_Success_TwoMetricsTwoSamples(t *testing.T) {
	results := map[string]*api.GetMetrics200Response{
		"metric1": metric1,
		"metric2": metric2,
	}
	labels := []string{"app", "job"}
	rmb := &resultMapBuilder{}
	data := rmb.Build(results, labels, []string{"metric1", "metric2"}, int32(1))

	expectedData := map[string][]string{
		"app1,job1": {"1", "3"},
		"app2,job1": {"2", "4"},
	}
	assert.Equal(t, 2, len(data))
	assert.ElementsMatch(t, expectedData["app1,job1"], data["app1,job1"])
	assert.ElementsMatch(t, expectedData["app2,job1"], data["app2,job1"])
}

func TestBuild_Success_MissingSample(t *testing.T) {
	results := map[string]*api.GetMetrics200Response{
		"metric1": metric1,
		"metric2": metricWithMissingSample,
	}
	labels := []string{"app", "job"}
	rmb := &resultMapBuilder{}
	data := rmb.Build(results, labels, []string{"metric1", "metric2"}, int32(1))

	expectedData := map[string][]string{
		"app1,job1": {"1", ""},
		"app2,job1": {"2", "5"},
	}
	assert.Equal(t, 2, len(data))
	assert.ElementsMatch(t, expectedData["app1,job1"], data["app1,job1"])
	assert.ElementsMatch(t, expectedData["app2,job1"], data["app2,job1"])
}

func TestBuild_Success_FailedMetric(t *testing.T) {
	results := map[string]*api.GetMetrics200Response{
		"metric1": metric1,
		"metric2": metricWithFailedStatus,
	}
	labels := []string{"app", "job"}
	rmb := &resultMapBuilder{}
	data := rmb.Build(results, labels, []string{"metric1", "metric2"}, int32(1))

	expectedData := map[string][]string{
		"app1,job1": {"1", ""},
		"app2,job1": {"2", ""},
	}
	assert.Equal(t, 2, len(data))
	assert.ElementsMatch(t, expectedData["app1,job1"], data["app1,job1"])
}

func TestBuild_Success_FloatNumberIsRoundedCorrectly(t *testing.T) {
	results := map[string]*api.GetMetrics200Response{
		"metric1": metricValuesWithDecimalOverPrecision,
	}
	labels := []string{"app"}
	rmb := &resultMapBuilder{}
	data := rmb.Build(results, labels, []string{"metric1"}, int32(1))

	expectedData := map[string][]string{
		"app1": {"1.1235"},
	}
	assert.Equal(t, 1, len(data))
	assert.ElementsMatch(t, expectedData["app1"], data["app1"])
}

func TestBuild_Success_EmptyResults(t *testing.T) {
	results := map[string]*api.GetMetrics200Response{}
	labels := []string{"app", "job"}
	rmb := &resultMapBuilder{}
	data := rmb.Build(results, labels, []string{}, int32(1))
	assert.Equal(t, 0, len(data))
}

func TestBuild_Success_EmptyLabelArray(t *testing.T) {
	results := map[string]*api.GetMetrics200Response{
		"metric1": metricWithMissingSample,
	}
	labels := []string{}
	rmb := &resultMapBuilder{}
	data := rmb.Build(results, labels, []string{"metric1"}, int32(1))

	expectedData := map[string][]string{
		"": {"5"},
	}
	assert.Equal(t, 1, len(data))
	assert.ElementsMatch(t, expectedData[""], data[""])
}

func TestFormatDecimalValue_StripFourDecimalPlaces(t *testing.T) {
	tests := []struct {
		value    string
		expected string
	}{
		{"1.11111111", "1.1111"},
		{"1.42344242", "1.4234"},
		{"1.6666666686868686", "1.6667"},
	}

	for _, test := range tests {
		val := formatDecimalValue(test.value)
		assert.Equal(t, test.expected, val)
	}
}

func TestFormatDecimalValue_NonFloatValuesRemainUnchanged(t *testing.T) {
	values := []string{
		"",
		"something",
	}
	for _, v := range values {
		val := formatDecimalValue(v)
		assert.Equal(t, v, val)
	}
}

func TestFormatDecimalValue_LeavesTrailingZeroes(t *testing.T) {
	values := []string{
		"1",
		"1.6",
		"1.53",
		"2.222",
	}

	for _, v := range values {
		val := formatDecimalValue(v)
		assert.Equal(t, v, val)
	}
}

func TestRound_AsExpected(t *testing.T) {
	tests := []struct {
		value     float64
		expected  float64
		precision int
	}{
		{1.1, 1, 0},
		{1.4123123, 1.4, 1},
		{1.5, 1.5, 2},
		{1.655, 1.66, 2},
		{1.655, 1.655, 3},
		{1.156165, 1.1562, 4},
		{1.4123123, 1.41231, 5},
	}

	for _, test := range tests {
		val := round(test.value, test.precision)
		assert.Equal(t, test.expected, val)
	}
}

func TestRoundToNearestInt_AsExpected(t *testing.T) {
	tests := []struct {
		value    float64
		expected int
	}{
		{1.1, 1},
		{1.4123123, 1},
		{1.5, 2},
		{1.655, 2},
	}

	for _, test := range tests {
		val := roundToNearestInt(test.value)
		assert.Equal(t, test.expected, val)
	}
}

// Prometheus reports NaN for a metric without a value and +Inf/-Inf for
// overflow, and a measurement can be arbitrarily large. None of those has a
// value that survives the int conversion in round, which used to replace them
// with an unrelated number.
func TestFormatDecimalValue_KeepsValuesThatCannotBeRounded(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "NaN", value: "NaN"},
		{name: "positive infinity", value: "+Inf"},
		{name: "negative infinity", value: "-Inf"},
		{name: "above the exactly representable range", value: "1e300"},
		{name: "below the exactly representable range", value: "-1e300"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want, err := strconv.ParseFloat(tc.value, 64)
			require.NoError(t, err)

			got := formatDecimalValue(tc.value)
			parsed, err := strconv.ParseFloat(got, 64)
			require.NoError(t, err, "formatDecimalValue(%q) = %q, which is not a number", tc.value, got)

			if math.IsNaN(want) {
				assert.True(t, math.IsNaN(parsed), "formatDecimalValue(%q) = %q, want NaN", tc.value, got)
				return
			}
			assert.Equal(t, want, parsed, "formatDecimalValue(%q) = %q", tc.value, got)
		})
	}
}

func TestFormatDecimalValue_StillRoundsLargeFiniteValues(t *testing.T) {
	assert.Equal(t, "1234567.8912", formatDecimalValue("1234567.8912345"))
	assert.Equal(t, "-1234567.8912", formatDecimalValue("-1234567.8912345"))
}
