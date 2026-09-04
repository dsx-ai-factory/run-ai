/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"fmt"
	"testing"
	"time"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

var (
	end   = time.Now()
	start = end.Add(-time.Hour)
)

type aggTestMetric struct {
	Filters        aggTestMetricFilters
	ResponseLabels aggTestMetricResponseLabels
	Quantile       *float32
}

type aggTestMetricFilters struct {
	TestFilter []string `mapstructure:"test_filter,omitempty"`
	AnotherOne []string `mapstructure:"AnotherOne,omitempty"`
}

type aggTestMetricResponseLabels struct {
	TestFilter bool `mapstructure:"test_filter,omitempty"`
	AnotherOne bool `mapstructure:"AnotherOne,omitempty"`
}

func (m aggTestMetric) GetName() string {
	return "aggTestMetric"
}

func (m aggTestMetric) GetFilters() interface{} {
	return m.Filters
}

func (m aggTestMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m aggTestMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m aggTestMetric) IsAggregation() bool {
	return true
}

func (m aggTestMetric) IsQuantile() bool {
	return false
}

var aggMetric1 = aggTestMetric{
	Filters: aggTestMetricFilters{
		TestFilter: []string{"test"},
	},
	ResponseLabels: aggTestMetricResponseLabels{
		TestFilter: true,
	},
}

var aggMetric2 = aggTestMetric{
	Filters: aggTestMetricFilters{
		AnotherOne: []string{"test", "anotherOne"},
	},
	ResponseLabels: aggTestMetricResponseLabels{
		TestFilter: false,
	},
}

func TestReportCreator_SetMetrics(t *testing.T) {
	r := NewReportCreator("test", &start, &end, "")
	r.SetMetrics([]ReportColumn{
		{DisplayName: "test", Metric: aggMetric1},
	})
	assert.Equal(t, 1, len(r.metrics))
	assert.Equal(t, "test", r.metrics[0].DisplayName)
	assert.Equal(t, "aggTestMetric", r.metrics[0].Metric.GetName())
}

func TestReportCreator_AddMetric(t *testing.T) {
	r := NewReportCreator("test", &start, &end, "")
	r.AddMetric(aggMetric1, "test")
	r.AddMetric(aggMetric2, "test2")
	assert.Equal(t, 2, len(r.metrics))
	assert.Equal(t, "test", r.metrics[0].DisplayName)
	assert.Equal(t, "aggTestMetric", r.metrics[0].Metric.GetName())
	assert.Equal(t, "test2", r.metrics[1].DisplayName)
	assert.Equal(t, "aggTestMetric", r.metrics[1].Metric.GetName())
}

func TestReportCreator_SetLabels(t *testing.T) {
	r := NewReportCreator("test", &start, &end, "")
	r.SetLabels([]openapi.Label{
		{
			Name:        "test",
			DisplayName: "Test",
		},
		{
			Name:        "test2",
			DisplayName: "Test2",
		},
	})
	assert.Equal(t, 2, len(r.labels))
	assert.Equal(t, "test", r.labels[0].Name)
	assert.Equal(t, "Test", r.labels[0].DisplayName)
}

func TestReportCreator_AddLabel(t *testing.T) {
	r := NewReportCreator("test", &start, &end, "")
	r.AddLabel(openapi.Label{
		Name:        "test",
		DisplayName: "Test",
	})
	r.AddLabel(openapi.Label{
		Name:        "test2",
		DisplayName: "Test2",
	})
	assert.Equal(t, 2, len(r.labels))
	assert.Equal(t, "test", r.labels[0].Name)
	assert.Equal(t, "Test", r.labels[0].DisplayName)
}

func TestReportCreator_SetAdditionalFields(t *testing.T) {
	r := NewReportCreator("test", &start, &end, "")
	r.SetAdditionalFields(map[string]interface{}{
		"test": "test",
	})
	assert.Equal(t, 1, len(r.additionalFields))
	assert.Equal(t, "test", r.additionalFields["test"])
}

func TestReportCreator_AddAdditionalField(t *testing.T) {
	r := NewReportCreator("test", &start, &end, "")
	r.AddAdditionalField("test", "test")
	r.AddAdditionalField("test2", "test2")
	assert.Equal(t, 2, len(r.additionalFields))
	assert.Equal(t, "test", r.additionalFields["test"])
	assert.Equal(t, "test2", r.additionalFields["test2"])
}

func TestReportCreator_Create(t *testing.T) {
	clientMock := mocks.NewMetricsClientContract(t)
	clientMock.On("CreateReport", mock.Anything, mock.Anything).Return(nil, 200, nil)
	getClientInstanceMock := func() MetricsClientContract {
		return clientMock
	}
	r := &reportCreator{
		name:              "test",
		start:             &start,
		end:               &end,
		getClientInstance: getClientInstanceMock,
	}
	r.SetMetrics([]ReportColumn{
		{DisplayName: "test", Metric: aggMetric1},
	})
	r.SetLabels([]openapi.Label{
		{
			Name:        "test",
			DisplayName: "Test",
		},
	})
	r.SetAdditionalFields(map[string]interface{}{
		"test": "test",
	})
	ctx := context.Background()
	createdReport, _, err := r.Create(ctx)
	fmt.Println(createdReport)
	assert.Nil(t, err)
	expectedRequestFields := &openapi.ReportRequestFields{
		Name:        "test",
		Description: ptr.To[string](""),
		Range: *openapi.NewNullableRange(&openapi.Range{
			Start: start,
			End:   end,
		}),
		Metrics: []openapi.MetricInformationForReport{
			{
				Name:        "aggTestMetric",
				DisplayName: "test",
				Filters: &map[string][]string{
					"test_filter": {"test"},
				},
				Quantile: *openapi.NewNullableFloat32(nil),
			},
		},
		Labels: []openapi.Label{
			{
				Name:        "test",
				DisplayName: "Test",
			},
		},
		AdditionalFields: map[string]interface{}{
			"test": "test",
		},
	}

	clientMock.AssertCalled(t, "CreateReport", ctx, mock.Anything)
	clientMock.AssertNumberOfCalls(t, "CreateReport", 1)
	arg := clientMock.Calls[0].Arguments[1]
	assert.Equal(t, expectedRequestFields, arg)
}

func TestReportCreator_ValidateMetric_Success(t *testing.T) {
	err := validateMetric(aggMetric1)
	assert.Nil(t, err)
}

func TestReportCreator_ValidateMetric_Fail(t *testing.T) {
	err := validateMetric(metric1)
	assert.NotNil(t, err)
}

func TestReportCreator_ValidateLabels_Success(t *testing.T) {
	err := validateLabels([]openapi.Label{
		{
			Name:        "test",
			DisplayName: "Test",
		},
	})
	assert.Nil(t, err)
}

func TestReportCreator_ValidateLabels_Fail(t *testing.T) {
	err := validateLabels([]openapi.Label{
		{
			Name: "test",
		},
	})
	assert.NotNil(t, err)
}

func TestReportCreator_BuildReportRequestFields(t *testing.T) {
	r := NewReportCreator("test", &start, &end, "")
	r.SetMetrics([]ReportColumn{
		{DisplayName: "test", Metric: aggMetric1},
	})
	r.SetLabels([]openapi.Label{
		{
			Name:        "test",
			DisplayName: "Test",
		},
	})
	r.SetAdditionalFields(map[string]interface{}{
		"test": "test",
	})
	fields, err := buildReportRequestFields(r)
	assert.Nil(t, err)
	assert.Equal(t, "test", fields.Name)
	assert.Equal(t, start, fields.GetRange().Start)
	assert.Equal(t, end, fields.GetRange().End)
	assert.Equal(t, 1, len(fields.Metrics))
	assert.Equal(t, "aggTestMetric", fields.Metrics[0].Name)
	assert.Equal(t, 1, len(fields.Labels))
	assert.Equal(t, "test", fields.Labels[0].Name)
	assert.Equal(t, "Test", fields.Labels[0].DisplayName)
	assert.Equal(t, 1, len(fields.AdditionalFields))
	assert.Equal(t, "test", fields.AdditionalFields["test"])
}

func TestReportCreator_BuildReportMetrics(t *testing.T) {
	metrics := []ReportColumn{
		{DisplayName: "test", Metric: aggMetric1},
		{DisplayName: "test2", Metric: aggMetric2},
	}
	reportMetrics, err := buildReportMetrics(metrics)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(reportMetrics))
	assert.Equal(t, "test", reportMetrics[0].DisplayName)
	assert.Equal(t, "test2", reportMetrics[1].DisplayName)
}
