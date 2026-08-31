//go:build integration
// +build integration


/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/run-ai/runai/control-plane/metrics-service/integration-tests/prometheus-with-agent/client"
	metrics_client "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
	openapi "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"k8s.io/utils/ptr"
)

var metricsList = []string{
	"example_metric_old",
	"example_metric_new",
	"example_metric_histogram_bucket",
	"example_metric_histogram_sum",
	"example_metric_histogram_count",
}

type MetricTestSuite struct {
	suite.Suite
	startTime time.Time
}

func TestMetricTestSuite(t *testing.T) {
	t.Setenv("PUBSUB_TYPE", getEnvOrDefault("PUBSUB_TYPE", "redis"))
	t.Setenv("PUBSUB_URL", getEnvOrDefault("PUBSUB_URL", "localhost:6379"))
	t.Setenv("SERVICE_NAME", getEnvOrDefault("SERVICE_NAME", "metrics-service"))
	t.Setenv("DB_HOST", getEnvOrDefault("DB_HOST", "localhost"))
	t.Setenv("DB_PORT", getEnvOrDefault("DB_PORT", "5432"))
	t.Setenv("DB_NAME", getEnvOrDefault("DB_NAME", "runai"))
	t.Setenv("SECRETS_MOUNT_PATH", getEnvOrDefault("SECRETS_MOUNT_PATH", "../secrets/"))
	t.Setenv("METRICS_SERVER_HOST", getEnvOrDefault("METRICS_SERVER_HOST", "localhost:8080"))

	suite.Run(t, new(MetricTestSuite))
}

func (suite *MetricTestSuite) SetupSuite() {
	fmt.Println("Setting up metrics integration suite")
	suite.startTime = time.Now()
	metricCollection := client.NewMetricCollection()
	metricCollection.AddMetric(client.NewGaugeMetric("example_metric_old", 2).WithTime(suite.startTime.Add(-23*time.Hour)).AddLabel("clusterId", "cluster_1"))
	metricCollection.AddMetric(client.NewGaugeMetric("example_metric_old", 1).WithTime(suite.startTime.Add(-22*time.Hour)).AddLabel("clusterId", "cluster_1"))
	metricCollection.AddMetric(client.NewGaugeMetric("example_metric_new", 1).WithTime(suite.startTime.Add(-15*time.Minute)).AddLabel("clusterId", "cluster_1"))
	metricCollection.AddMetric(client.NewGaugeMetric("example_metric_new", 1).AddLabel("clusterId", "cluster_1"))
	metricCollection.AddMetric(client.NewGaugeMetric("example_metric_new", 1).AddLabel("clusterId", "cluster_2"))

	metricCollection.AddMetric(client.NewHistogramMetric("example_metric_histogram", []float64{0.5, 1, 2, 5}).
		AddLabel("clusterId", clusterId1).
		WithTime(suite.startTime.Add(-12 * time.Hour)).
		Observe(0.15).Observe(1.5).Observe(3))
	metricCollection.AddMetric(client.NewHistogramMetric("example_metric_histogram", []float64{0.5, 1, 2, 5}).
		AddLabel("clusterId", clusterId1).
		WithTime(suite.startTime.Add(-6 * time.Hour)).
		Observe(0.25).Observe(2.5).Observe(4))
	metricCollection.AddMetric(client.NewHistogramMetric("example_metric_histogram", []float64{0.5, 1, 2, 5}).
		AddLabel("clusterId", clusterId1).
		Observe(0.35).Observe(3.5).Observe(5))

	err := client.PostMetrics(metricCollection.ToOpenMetrics())
	assert.NoError(suite.T(), err)
	err = client.WaitForMetrics(metricsList, 5, 120)
	assert.NoError(suite.T(), err)
}

func (suite *MetricTestSuite) TearDownSuite() {
	client.ClearMetrics(metricsList)
}

func (suite *MetricTestSuite) TestGetMetrics_Now() {
	metric := &ExampleMetric{
		Filters: ExampleMetricFilters{
			ClusterId: []string{"cluster_1"},
		},
		ResponseLabels: ExampleMetricResponseLabels{
			ClusterId: true,
		},
	}
	metricQuery, err := metrics_client.NewMetricQuery(metric)
	assert.NoError(suite.T(), err)
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	metrics, err := metricQuery.Query(ctx)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), metrics)
	assert.Contains(suite.T(), metrics.Query, "example_metric_new")
	assert.NotContains(suite.T(), metrics.Query, "example_metric_old")
	assert.Len(suite.T(), metrics.Data, 1)
	assert.Len(suite.T(), metrics.Data[0].Measurements, 1)
}

func (suite *MetricTestSuite) TestGetMetrics_Now_InternalMetricStore() {
	tenantId := random.Int31()
	fmt.Printf("running TestGetMetrics_Now_InternalMetricStore, tenant [%d]\n", tenantId)
	tenantMetricStore := &openapi.MetricStore{
		TenantId:     tenantId,
		DatasourceId: ptr.To[string](datasourceId),
		MetricStoreDetails: *openapi.NewNullableMetricStoreDetails(&openapi.MetricStoreDetails{
			Type: openapi.METRICSTORETYPE_MIMIR,
			Read: openapi.MetricStoreEndpoint{
				Url: "http://prometheus:9090",
			},
			Write: openapi.MetricStoreEndpoint{
				Url: "http://prometheus:9090",
			},
		}),
	}

	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "test@run.ai")
	res, err := metrics_client.GetMetricStoresClient().CreateMetricStore(ctx, tenantMetricStore)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), res)

	metric := &ExampleMetric{
		Filters: ExampleMetricFilters{
			ClusterId: []string{"cluster_1"},
		},
		ResponseLabels: ExampleMetricResponseLabels{
			ClusterId: true,
		},
	}
	metricQuery, err := metrics_client.NewMetricQuery(metric)
	assert.NoError(suite.T(), err)
	metrics, err := metricQuery.Query(ctx)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), metrics)
	assert.Contains(suite.T(), metrics.Query, "example_metric_new")
	assert.NotContains(suite.T(), metrics.Query, "example_metric_old")
	assert.Len(suite.T(), metrics.Data, 1)
	assert.Len(suite.T(), metrics.Data[0].Measurements, 1)
}

func (suite *MetricTestSuite) TestGetMetrics_Now_OverrideTenantId() {
	metric := &ExampleMetric{
		Filters: ExampleMetricFilters{
			ClusterId: []string{"cluster_1"},
		},
		ResponseLabels: ExampleMetricResponseLabels{
			ClusterId: true,
		},
	}
	metricQuery, err := metrics_client.NewMetricQuery(metric)
	metricQuery.SetOverrideTenantId(1)
	assert.NoError(suite.T(), err)
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "2")
	metrics, err := metricQuery.Query(ctx)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), metrics)
	assert.Contains(suite.T(), metrics.Query, "example_metric_new")
	assert.NotContains(suite.T(), metrics.Query, "example_metric_old")
	assert.Len(suite.T(), metrics.Data, 1)
	assert.Len(suite.T(), metrics.Data[0].Measurements, 1)
}

func (suite *MetricTestSuite) TestGetMetrics_NoTenantId() {
	metric := &ExampleMetric{
		Filters: ExampleMetricFilters{
			ClusterId: []string{"cluster_1"},
		},
		ResponseLabels: ExampleMetricResponseLabels{
			ClusterId: true,
		},
	}
	metricQuery, err := metrics_client.NewMetricQuery(metric)
	assert.NoError(suite.T(), err)
	ctx := context.Background()
	metrics, err := metricQuery.Query(ctx)
	assert.ErrorContains(suite.T(), err, "400 Bad Request")
	assert.Empty(suite.T(), metrics)
}

func (suite *MetricTestSuite) TestGetMetrics_TenantDatasourceNotExist() {
	metric := &ExampleMetric{
		Filters: ExampleMetricFilters{
			ClusterId: []string{"cluster_1"},
		},
		ResponseLabels: ExampleMetricResponseLabels{
			ClusterId: true,
		},
	}
	metricQuery, err := metrics_client.NewMetricQuery(metric)
	assert.NoError(suite.T(), err)
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "123")
	metrics, err := metricQuery.Query(ctx)
	assert.ErrorContains(suite.T(), err, "421 Misdirected Request")
	assert.Empty(suite.T(), metrics)
}

func (suite *MetricTestSuite) TestGetMetrics_Range() {
	metric := &ExampleMetric{
		Filters: ExampleMetricFilters{
			ClusterId: []string{"cluster_1"},
		},
		ResponseLabels: ExampleMetricResponseLabels{
			ClusterId: true,
		},
	}
	metricQuery, err := metrics_client.NewMetricQuery(metric)
	assert.NoError(suite.T(), err)
	metricQuery.SetRange(suite.startTime.Add(-24*time.Hour), suite.startTime)
	metricQuery.SetDefaultNumOfSamples(1000)

	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	metrics, err := metricQuery.Query(ctx)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), metrics)
	assert.Contains(suite.T(), metrics.Query, "example_metric_new")
	assert.Contains(suite.T(), metrics.Query, "example_metric_old")
	assert.Len(suite.T(), metrics.Data, 1)
}

func (suite *MetricTestSuite) TestGetMetrics_Aggregation() {
	metric := &ExampleMetricAggregation{
		Filters: ExampleMetricAggregationFilters{
			ClusterId: []string{"cluster_1"},
		},
		ResponseLabels: ExampleMetricAggregationResponseLabels{
			ClusterId: true,
		},
	}
	metricQuery, err := metrics_client.NewMetricQuery(metric)
	assert.NoError(suite.T(), err)
	metricQuery.SetRange(suite.startTime.Add(-24*time.Hour), suite.startTime)

	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	metrics, err := metricQuery.Query(ctx)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), metrics)
	assert.Equal(suite.T(), metrics.Query, "sum(sum_over_time(example_metric_new{clusterId=~'cluster_1'}[86400s])) by(clusterId)")
	assert.Len(suite.T(), metrics.Data, 1)
	assert.Equal(suite.T(), metrics.Data[0].Measurements[0].Value, "2")
}

func (suite *MetricTestSuite) TestGetMetrics_Quantile() {
	metric := &ExampleMetricQuantile{
		Filters: ExampleMetricQuantileFilters{
			ClusterId: []string{"cluster_1"},
		},
		ResponseLabels: ExampleMetricQuantileResponseLabels{
			ClusterId: true,
		},
	}
	metricQuery, err := metrics_client.NewMetricQuery(metric)
	assert.NoError(suite.T(), err)

	metricQuery.SetQuantile(ptr.To[float32](0.75))
	metricQuery.SetRange(suite.startTime.Add(-24*time.Hour), suite.startTime)

	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	metrics, err := metricQuery.Query(ctx)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), metrics)
	assert.Equal(suite.T(), "histogram_quantile(0.75, sum(rate(example_metric_histogram_bucket{clusterId=~'cluster_1'}[86400s])) by (le, clusterId))", metrics.Query)
	assert.Len(suite.T(), metrics.Data, 1)
	assert.Equal(suite.T(), "1.75", metrics.Data[0].Measurements[0].Value)
}

func (suite *MetricTestSuite) TestGetMetrics_Now_NoClusterVersionHistory() {
	metric := &ExampleMetric{
		Filters: ExampleMetricFilters{
			ClusterId: []string{"cluster_2"},
		},
		ResponseLabels: ExampleMetricResponseLabels{
			ClusterId: true,
		},
	}
	metricQuery, err := metrics_client.NewMetricQuery(metric)
	assert.NoError(suite.T(), err)
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	metrics, err := metricQuery.Query(ctx)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), metrics)
	assert.Contains(suite.T(), metrics.Query, "example_metric_new")
	assert.Contains(suite.T(), metrics.Query, "example_metric_old")
	assert.Len(suite.T(), metrics.Data, 1)
	assert.Len(suite.T(), metrics.Data[0].Measurements, 1)
}

func (suite *MetricTestSuite) TestGetMetrics_Aggregation_NoClusterVersionHistory() {
	metric := &ExampleMetricAggregation{
		Filters: ExampleMetricAggregationFilters{
			ClusterId: []string{"cluster_3"},
		},
		ResponseLabels: ExampleMetricAggregationResponseLabels{
			ClusterId: true,
		},
	}
	metricQuery, err := metrics_client.NewMetricQuery(metric)
	assert.NoError(suite.T(), err)
	metricQuery.SetRange(suite.startTime.Add(-24*time.Hour), suite.startTime)

	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	metrics, err := metricQuery.Query(ctx)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), metrics)
	assert.Equal(suite.T(), metrics.Query, "sum(sum_over_time(example_metric_new{clusterId=~'cluster_3'}[86400s])) by(clusterId)")
	assert.Len(suite.T(), metrics.Data, 0)
}
