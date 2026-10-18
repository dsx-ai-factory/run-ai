//go:build integration
// +build integration

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/gocarina/gocsv"
	"github.com/minio/minio-go/v7"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/integration-tests/prometheus-with-agent/client"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	metrics_client "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	s3_client "github.com/dsx-ai-factory/run-ai/runai-common-packages/s3"
	"github.com/stretchr/testify/require"
)

const (
	clusterId1 = "cluster_1"
	clusterId2 = "cluster_2"
	clusterId3 = "cluster_3"
)

type CsvReportRow struct {
	ClusterId string `csv:"ClusterId"`
}

type ReportTestSuite struct {
	suite.Suite
	startTime time.Time
}

func TestReportsTestSuite(t *testing.T) {
	t.Setenv("PUBSUB_TYPE", getEnvOrDefault("PUBSUB_TYPE", "redis"))
	t.Setenv("PUBSUB_URL", getEnvOrDefault("PUBSUB_URL", "localhost:6379"))
	t.Setenv("SERVICE_NAME", getEnvOrDefault("SERVICE_NAME", "metrics-service"))
	t.Setenv("DB_HOST", getEnvOrDefault("DB_HOST", "localhost"))
	t.Setenv("DB_PORT", getEnvOrDefault("DB_PORT", "5432"))
	t.Setenv("DB_NAME", getEnvOrDefault("DB_NAME", "runai"))
	t.Setenv("SECRETS_MOUNT_PATH", getEnvOrDefault("SECRETS_MOUNT_PATH", "../secrets/"))
	t.Setenv("METRICS_SERVER_HOST", getEnvOrDefault("METRICS_SERVER_HOST", "localhost:8080"))
	t.Setenv("S3_ENDPOINT", getEnvOrDefault("S3_ENDPOINT", "localhost:9000"))
	t.Setenv("DB_SCHEMA", getEnvOrDefault("DB_SCHEMA", "metrics"))
	if err := runaidb.Init(nil); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}
	suite.Run(t, new(ReportTestSuite))
}

func (suite *ReportTestSuite) SetupSuite() {
	fmt.Println("Setting up report integration suite")
	suite.startTime = time.Now()
	metricCollection := client.NewMetricCollection()
	metricCollection.AddMetric(client.NewGaugeMetric("example_metric_old", 2).WithTime(suite.startTime.Add(-23*time.Hour)).AddLabel("clusterId", clusterId1))
	metricCollection.AddMetric(client.NewGaugeMetric("example_metric_old", 1).WithTime(suite.startTime.Add(-22*time.Hour)).AddLabel("clusterId", clusterId1))
	metricCollection.AddMetric(client.NewGaugeMetric("example_metric_new", 1).WithTime(suite.startTime.Add(-15*time.Minute)).AddLabel("clusterId", clusterId1))
	metricCollection.AddMetric(client.NewGaugeMetric("example_metric_new", 1).AddLabel("clusterId", clusterId1))
	metricCollection.AddMetric(client.NewGaugeMetric("example_metric_new", 1).AddLabel("clusterId", clusterId2))

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
	client.WaitForMetrics(metricsList, 5, 120)
	require.NoError(suite.T(), err)
}

func (suite *ReportTestSuite) TearDownSuite() {
	client.ClearMetrics(metricsList)
}

func (suite *ReportTestSuite) TestGetReport_StatusShouldBePending() {
	reportId := uuid.New().String()
	metric := createMetric([]string{clusterId3})
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	report, err := createReport(ctx, reportId, metric)

	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), report)
	require.Equal(suite.T(), openapi.REPORTSTATUS_PENDING, report.Status)
}

func (suite *ReportTestSuite) TestGetReport_StatusShouldBeSucceedAfterTime() {
	client, err := initializeS3Client()
	require.NoError(suite.T(), err)

	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	reportId := uuid.New().String()
	metric := createMetric([]string{clusterId1})
	report, err := createReport(ctx, reportId, metric)

	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), report)
	require.Equal(suite.T(), openapi.REPORTSTATUS_PENDING, report.Status)

	err = waitForReportStatus(ctx, report.Id, openapi.REPORTSTATUS_READY, 10*time.Second, 1*time.Second)
	require.NoError(suite.T(), err)
	requireReportStatus(suite, ctx, report.Id, openapi.REPORTSTATUS_READY)

	require.NoError(suite.T(), downloadReport(client, "metrics-service", report.Id))
	defer os.Remove(report.Id)
	headers := readCSVHeaders(report.Id, []string{"ClusterId"})
	require.NotNil(suite.T(), headers)
	require.Equal(suite.T(), headers, []string{"example_metric_aggregation", "an_example_metric_aggregation", "weekly_example_metric_aggregation", "quick_example_metric_aggregation"})
}

func (suite *ReportTestSuite) TestGetReport_ClusterIdLabelSetToNameIfClusterExists() {
	client, err := initializeS3Client()
	require.NoError(suite.T(), err)

	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")

	testCases := []struct {
		id            string
		expectedLabel string
	}{
		{clusterId1, getClusterNameFromId(clusterId1)},
		{clusterId2, clusterId2}, // there is no mock cluster for clusterId2, so label is unchanged
	}

	for _, tc := range testCases {
		reportId := uuid.New().String()
		metric := createMetric([]string{tc.id})
		report, err := createReport(ctx, reportId, metric)

		require.NoError(suite.T(), err)
		require.NotNil(suite.T(), report)
		require.Equal(suite.T(), openapi.REPORTSTATUS_PENDING, report.Status)

		err = waitForReportStatus(ctx, report.Id, openapi.REPORTSTATUS_READY, 10*time.Second, 1*time.Second)
		require.NoError(suite.T(), err)
		requireReportStatus(suite, ctx, report.Id, openapi.REPORTSTATUS_READY)

		require.NoError(suite.T(), downloadReport(client, "metrics-service", report.Id))
		defer os.Remove(report.Id)

		file, err := os.Open(report.Id)
		require.Nil(suite.T(), err)
		defer file.Close()

		rows := []CsvReportRow{}
		err = gocsv.Unmarshal(file, &rows)
		require.Nil(suite.T(), err)

		require.Equal(suite.T(), tc.expectedLabel, rows[0].ClusterId)
	}
}

func (suite *ReportTestSuite) TestGetReport_WithQuantileMetric() {
	client, err := initializeS3Client()
	require.NoError(suite.T(), err)

	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	reportId := uuid.New().String()

	metric := createQuantileMetric([]string{clusterId1})
	report, err := createReportWithQuantileMetrics(ctx, reportId, metric)

	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), report)
	require.Equal(suite.T(), openapi.REPORTSTATUS_PENDING, report.Status)

	err = waitForReportStatus(ctx, report.Id, openapi.REPORTSTATUS_READY, 10*time.Second, 1*time.Second)
	require.NoError(suite.T(), err)
	requireReportStatus(suite, ctx, report.Id, openapi.REPORTSTATUS_READY)

	require.NoError(suite.T(), downloadReport(client, "metrics-service", report.Id))
	defer os.Remove(report.Id)

	file, err := os.Open(report.Id)
	require.Nil(suite.T(), err)
	defer file.Close()

	rows := []struct {
		CsvReportRow
		ExampleMetricQuantileAggregation string `csv:"example_metric_quantile_aggregation"`
		SecondExampleQuantileAggregation string `csv:"second_example_metric_quantile_aggregation"`
	}{}
	err = gocsv.Unmarshal(file, &rows)
	require.Nil(suite.T(), err)

	require.Equal(suite.T(), getClusterNameFromId(clusterId1), rows[0].ClusterId)
	require.Equal(suite.T(), "1.5", rows[0].ExampleMetricQuantileAggregation)
	require.Equal(suite.T(), "1.5", rows[0].SecondExampleQuantileAggregation)
}

func (suite *ReportTestSuite) TestRetryFailedReportsCronJob() {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "2") // tenant 2 is not mocked, therefore report will fail

	reportId := uuid.New().String()
	metric := createMetric([]string{clusterId2})
	report, err := createReport(ctx, reportId, metric)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), report)

	require.Equal(suite.T(), openapi.REPORTSTATUS_PENDING, report.Status)
	err = waitForReportStatus(ctx, report.Id, openapi.REPORTSTATUS_FAILED, 10*time.Second, 1*time.Second)
	require.NoError(suite.T(), err)

	client := runaidb.DB()
	require.Eventually(suite.T(), func() bool {
		row := client.Table("metrics.reports").Where("id = ?", report.Id).Select("retry_count").Row()
		var retryCount int
		row.Scan(&retryCount)
		return retryCount == 1
	}, 10*time.Second, 1*time.Second)
}

func (suite *ReportTestSuite) TestFailStuckProcessingReportsCronJob() {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")

	reportId := uuid.New().String()
	metric := createMetric([]string{clusterId2})
	report, err := createReport(ctx, reportId, metric)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), report)

	err = waitForReportStatus(ctx, report.Id, openapi.REPORTSTATUS_READY, 10*time.Second, 1*time.Second)
	require.NoError(suite.T(), err)

	var status string
	client := runaidb.DB()
	_, err = client.
		Raw("UPDATE metrics.reports SET created_at = NOW() - INTERVAL '31 minutes', status = ? WHERE id = ? RETURNING status", openapi.REPORTSTATUS_PROCESSING, report.Id).
		Scan(&status)

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), status, string(openapi.REPORTSTATUS_PROCESSING))

	err = waitForReportStatus(ctx, report.Id, openapi.REPORTSTATUS_FAILED, 10*time.Second, 1*time.Second)
	require.NoError(suite.T(), err)

	report, _, err = metrics_client.GetClientInstance().GetReport(ctx, report.Id)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), *(report.StatusMessage), "timeout error")
}

func (suite *ReportTestSuite) TestBatchDeleteReports_PendingSucceedsAndProcessingFails() {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	pendingReportId := createReportInStatus(suite, openapi.REPORTSTATUS_PENDING)
	processingReportId := createReportInStatus(suite, openapi.REPORTSTATUS_PROCESSING)

	response, statusCode, err := metrics_client.GetClientInstance().
		BatchDeleteReports(ctx, []string{pendingReportId, processingReportId})

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusAccepted, statusCode)
	require.ElementsMatch(suite.T(), []string{pendingReportId}, response.Accepted)
	require.Len(suite.T(), response.Failed, 1)
	require.Equal(suite.T(), processingReportId, response.Failed[0].Id)
	require.Equal(suite.T(), int32(http.StatusProcessing), response.Failed[0].Code)

	requireReportStatus(suite, ctx, pendingReportId, openapi.REPORTSTATUS_PENDING_DELETE)
	requireReportStatus(suite, ctx, processingReportId, openapi.REPORTSTATUS_PROCESSING)
}

func (suite *ReportTestSuite) TestBatchDeleteReports_ReportsDoNotExist() {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	reportIds := []string{uuid.NewString(), uuid.NewString()}

	response, statusCode, err := metrics_client.GetClientInstance().BatchDeleteReports(ctx, reportIds)

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusAccepted, statusCode)
	require.Empty(suite.T(), response.Accepted)
	require.Len(suite.T(), response.Failed, len(reportIds))
	require.ElementsMatch(suite.T(), reportIds, []string{response.Failed[0].Id, response.Failed[1].Id})
	for _, failed := range response.Failed {
		require.Equal(suite.T(), int32(http.StatusNotFound), failed.Code)
	}
}

func (suite *ReportTestSuite) TestBatchDeleteReports_ProcessingReportFails() {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	reportId := createReportInStatus(suite, openapi.REPORTSTATUS_PROCESSING)

	response, statusCode, err := metrics_client.GetClientInstance().BatchDeleteReports(ctx, []string{reportId})

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusAccepted, statusCode)
	require.Empty(suite.T(), response.Accepted)
	require.Len(suite.T(), response.Failed, 1)
	require.Equal(suite.T(), reportId, response.Failed[0].Id)
	require.Equal(suite.T(), int32(http.StatusProcessing), response.Failed[0].Code)
	requireReportStatus(suite, ctx, reportId, openapi.REPORTSTATUS_PROCESSING)
}

func (suite *ReportTestSuite) TestBatchGetReports_BeforeAndAfterReportIsReady() {
	s3Client, err := initializeS3Client()
	require.NoError(suite.T(), err)

	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	metric := createMetric([]string{clusterId1})
	report, err := createReport(ctx, uuid.NewString(), metric)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), report)

	response, statusCode, err := metrics_client.GetClientInstance().
		BatchGetReportsByIds(ctx, []string{report.Id})

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusOK, statusCode)
	require.Empty(suite.T(), response.Failed)
	require.Len(suite.T(), response.Succeeded, 1)
	require.Equal(suite.T(), report.Id, response.Succeeded[0].Id)
	requireReportStatusIn(suite, ctx, report.Id, openapi.REPORTSTATUS_PENDING, openapi.REPORTSTATUS_PROCESSING)

	_, err = s3Client.StatObject(ctx, "metrics-service", report.Id, minio.StatObjectOptions{})
	require.Error(suite.T(), err)
	require.Equal(suite.T(), "NoSuchKey", minio.ToErrorResponse(err).Code)

	err = waitForReportStatus(ctx, report.Id, openapi.REPORTSTATUS_READY, 10*time.Second, time.Second)
	require.NoError(suite.T(), err)

	response, statusCode, err = metrics_client.GetClientInstance().
		BatchGetReportsByIds(ctx, []string{report.Id})

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusOK, statusCode)
	require.Empty(suite.T(), response.Failed)
	require.Len(suite.T(), response.Succeeded, 1)
	require.Equal(suite.T(), report.Id, response.Succeeded[0].Id)
	require.Equal(suite.T(), openapi.REPORTSTATUS_READY, response.Succeeded[0].Status)
	requireReportStatus(suite, ctx, report.Id, openapi.REPORTSTATUS_READY)

	_, err = s3Client.StatObject(ctx, "metrics-service", report.Id, minio.StatObjectOptions{})
	require.NoError(suite.T(), err)
}

func (suite *ReportTestSuite) TestBatchGetAndDeleteReport_EventuallyDeleted() {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	report, err := createReport(ctx, uuid.NewString(), createMetric([]string{clusterId1}))
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), report)

	suite.T().Cleanup(func() {
		runaidb.DB().Exec("DELETE FROM metrics.reports WHERE id = ?", report.Id)
	})

	response, statusCode, err := metrics_client.GetClientInstance().
		BatchGetReportsByIds(ctx, []string{report.Id})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusOK, statusCode)
	require.Empty(suite.T(), response.Failed)
	require.Len(suite.T(), response.Succeeded, 1)
	require.Equal(suite.T(), report.Id, response.Succeeded[0].Id)

	err = waitForReportStatus(ctx, report.Id, openapi.REPORTSTATUS_READY, 10*time.Second, time.Second)
	require.NoError(suite.T(), err)

	deleteResponse, statusCode, err := metrics_client.GetClientInstance().
		BatchDeleteReports(ctx, []string{report.Id})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusAccepted, statusCode)
	require.Empty(suite.T(), deleteResponse.Failed)
	require.Equal(suite.T(), []string{report.Id}, deleteResponse.Accepted)

	require.Eventually(suite.T(), func() bool {
		var count int64
		db := runaidb.DB().Table("metrics.reports").Where("id = ?", report.Id)
		err := db.Count(&count)
		return err == nil && count == 0
	}, 15*time.Second, time.Second)
}

func createReportInStatus(suite *ReportTestSuite, status openapi.ReportStatus) string {
	now := time.Now()
	report := models.Report{
		ID:               uuid.New(),
		CreatedAt:        now,
		CreatedBy:        "integration-test",
		Name:             fmt.Sprintf("batch-report-%s-%s", status, uuid.NewString()),
		TenantId:         1,
		Status:           string(status),
		StatusUpdatedAt:  now,
		Start:            now.Add(-time.Hour),
		End:              now,
		Labels:           models.LabelArray{},
		AdditionalFields: models.AdditionalFields{},
	}
	_, err := runaidb.DB().Create(&report)
	require.NoError(suite.T(), err)
	suite.T().Cleanup(func() {
		runaidb.DB().Exec("DELETE FROM metrics.reports WHERE id = ?", report.ID)
	})
	return report.ID.String()
}

func requireReportStatus(suite *ReportTestSuite, ctx context.Context, reportId string, expectedStatus openapi.ReportStatus) {
	report, _, err := metrics_client.GetClientInstance().GetReport(ctx, reportId)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), report)
	require.Equal(suite.T(), expectedStatus, report.Status)
}

func requireReportStatusIn(suite *ReportTestSuite, ctx context.Context, reportId string, expectedStatuses ...openapi.ReportStatus) {
	report, _, err := metrics_client.GetClientInstance().GetReport(ctx, reportId)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), report)
	require.Contains(suite.T(), expectedStatuses, report.Status)
}

func waitForReportStatus(ctx context.Context, reportId string, expectedStatus openapi.ReportStatus, timeout, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for report %s to reach status %s", reportId, expectedStatus)
		case <-ticker.C:
			report, _, err := metrics_client.GetClientInstance().GetReport(ctx, reportId)
			if err != nil {
				return fmt.Errorf("error fetching report %s: %v", reportId, err)
			}

			if report.Status == expectedStatus {
				return nil
			}
		}
	}
}

func initializeS3Client() (*minio.Client, error) {
	endpoint := getEnvOrDefault("S3_ENDPOINT", "localhost:9000")
	s3_client.Init(&s3_client.S3Config{Endpoint: endpoint, AccessKey: "minioadmin", AccessKeyId: "minioadmin"})
	client, err := s3_client.Client()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize S3 client: %v", err)
	}

	exists, err := client.BucketExists(context.Background(), "metrics-service")
	if err != nil || !exists {
		return nil, fmt.Errorf("bucket does not exist: %v", err)
	}

	return client, nil
}

func createMetric(clusterIdsToFilter []string) *ExampleMetricAggregation {
	return &ExampleMetricAggregation{
		Filters: ExampleMetricAggregationFilters{
			ClusterId: clusterIdsToFilter,
		},
		ResponseLabels: ExampleMetricAggregationResponseLabels{
			ClusterId: true,
		},
	}
}

func createReport(ctx context.Context, reportId string, metric *ExampleMetricAggregation) (*openapi.Report, error) {
	start, end := time.Now().Add(-12*time.Hour), time.Now()
	report, _, err := metrics_client.NewReportCreator(reportId, &start, &end, "").
		SetLabels([]openapi.Label{{Name: "clusterId", DisplayName: "ClusterId"}}).
		AddMetric(metric, "example_metric_aggregation").
		AddMetric(metric, "an_example_metric_aggregation").
		AddMetric(metric, "weekly_example_metric_aggregation").
		AddMetric(metric, "quick_example_metric_aggregation").
		Create(ctx)
	return report, err
}

func createQuantileMetric(clusterIdsToFilter []string) *ExampleMetricQuantile {
	return &ExampleMetricQuantile{
		Filters: ExampleMetricQuantileFilters{
			ClusterId: clusterIdsToFilter,
		},
		ResponseLabels: ExampleMetricQuantileResponseLabels{
			ClusterId: true,
		},
		Quantile: ptr.To[float32](0.5),
	}
}

func createReportWithQuantileMetrics(ctx context.Context, reportId string, metric *ExampleMetricQuantile) (*openapi.Report, error) {
	start, end := time.Now().Add(-24*time.Hour), time.Now()
	report, _, err := metrics_client.NewReportCreator(reportId, &start, &end, "").
		SetLabels([]openapi.Label{{Name: "clusterId", DisplayName: "ClusterId"}}).
		AddMetric(metric, "example_metric_quantile_aggregation").
		AddMetric(metric, "second_example_metric_quantile_aggregation").
		Create(ctx)
	return report, err
}

func downloadReport(client *minio.Client, bucket, reportId string) error {
	object, err := client.GetObject(context.Background(), bucket, reportId, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer object.Close()

	localFile, err := os.Create(reportId)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer localFile.Close()

	_, err = io.Copy(localFile, object)
	if err != nil {
		return fmt.Errorf("failed to save object locally: %w", err)
	}

	return nil
}

func readCSVHeaders(fileName string, exclude []string) []string {
	file, err := os.Open(fileName)
	if err != nil {
		return nil
	}
	defer file.Close()

	reader := csv.NewReader(file)

	headers, err := reader.Read()
	if err != nil {
		return nil
	}

	var result []string
	for _, h := range headers {
		excluded := false
		for _, e := range exclude {
			excluded = excluded || (e == h)
		}
		if !excluded {
			result = append(result, h)
		}
	}
	return result
}

func getClusterNameFromId(id string) string {
	return fmt.Sprintf("%s_name", id)
}

func reportCtx(tenantId, subject string) context.Context {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	return context.WithValue(ctx, runaictx.CfSubjectID, subject)
}

// createReportWithCleanup creates a report through the API and deletes it when the test ends.
func createReportWithCleanup(suite *ReportTestSuite, ctx context.Context, name string) *openapi.Report {
	report, err := createReport(ctx, name, createMetric([]string{clusterId1}))
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), report)
	suite.T().Cleanup(func() {
		if _, err := metrics_client.GetClientInstance().DeleteReport(ctx, report.Id); err != nil {
			suite.T().Errorf("failed to delete report %s: %v", report.Id, err)
		}
	})
	return report
}

func (suite *ReportTestSuite) TestGetReport_ReturnsTheCreatedReport() {
	ctx := reportCtx("1", "creator@example.com")
	name := "get-fields-" + uuid.NewString()
	created := createReportWithCleanup(suite, ctx, name)

	got, statusCode, err := metrics_client.GetClientInstance().GetReport(ctx, created.Id)

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusOK, statusCode)
	require.Equal(suite.T(), created.Id, got.Id)
	require.Equal(suite.T(), name, got.Name)
	require.Equal(suite.T(), "creator@example.com", got.CreatedBy)
	require.Equal(suite.T(), int32(1), got.TenantId)
	require.Equal(suite.T(), []openapi.Label{{Name: "clusterId", DisplayName: "ClusterId"}}, got.Labels)
	require.False(suite.T(), got.CreatedAt.IsZero())
	require.False(suite.T(), got.StatusUpdatedAt.Before(got.CreatedAt))
}

func (suite *ReportTestSuite) TestDeleteReport() {
	ctx := reportCtx("1", "creator@example.com")
	report, err := createReport(ctx, "delete-"+uuid.NewString(), createMetric([]string{clusterId1}))
	require.NoError(suite.T(), err)

	statusCode, err := metrics_client.GetClientInstance().DeleteReport(ctx, report.Id)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusNoContent, statusCode)

	_, statusCode, err = metrics_client.GetClientInstance().GetReport(ctx, report.Id)
	require.Error(suite.T(), err)
	require.Equal(suite.T(), http.StatusNotFound, statusCode)
}

func (suite *ReportTestSuite) TestBatchGetReports_OtherTenantsReportIsNotReturned() {
	owner := reportCtx("1", "creator@example.com")
	created := createReportWithCleanup(suite, owner, "tenant-scope-"+uuid.NewString())

	response, statusCode, err := metrics_client.GetClientInstance().
		BatchGetReportsByIds(reportCtx("3", "other@example.com"), []string{created.Id})

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusOK, statusCode)
	require.Empty(suite.T(), response.Succeeded)
	require.Len(suite.T(), response.Failed, 1)
	require.Equal(suite.T(), created.Id, response.Failed[0].Id)
	require.Equal(suite.T(), int32(http.StatusNotFound), response.Failed[0].Code)
}

func (suite *ReportTestSuite) TestListReports_SortingAndPaging() {
	// A tenant of its own keeps the listing to exactly these reports.
	tenantId := fmt.Sprintf("%d", random.Int31())
	ctxFor := func(subject string) context.Context { return reportCtx(tenantId, subject) }

	// Created in this order, so name, createdAt and createdBy each order them differently.
	names := []string{"b", "a", "c"}
	subjects := []string{"user-2", "user-3", "user-1"}
	prefix := uuid.NewString()
	for i, name := range names {
		createReportWithCleanup(suite, ctxFor(subjects[i]), prefix+"-"+name)
	}
	listCtx := ctxFor("lister")
	listNames := func(sortBy openapi.ReportFilterAndSortFields, sortOrder string, offset, limit int32) []string {
		reports, statusCode, err := metrics_client.GetClientInstance().ListReports(listCtx, nil, sortBy, sortOrder, offset, limit, "")
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), http.StatusOK, statusCode)
		got := make([]string, 0, len(reports))
		for _, report := range reports {
			got = append(got, strings.TrimPrefix(report.Name, prefix+"-"))
		}
		return got
	}

	require.Equal(suite.T(), []string{"a", "b", "c"}, listNames(openapi.REPORTFILTERANDSORTFIELDS_NAME, "asc", 0, 100))
	require.Equal(suite.T(), []string{"c", "b", "a"}, listNames(openapi.REPORTFILTERANDSORTFIELDS_NAME, "desc", 0, 100))
	require.Equal(suite.T(), []string{"b", "a", "c"}, listNames(openapi.REPORTFILTERANDSORTFIELDS_CREATED_AT, "asc", 0, 100))
	require.Equal(suite.T(), []string{"c", "b", "a"}, listNames(openapi.REPORTFILTERANDSORTFIELDS_CREATED_BY, "asc", 0, 100))

	require.Equal(suite.T(), []string{"a", "b"}, listNames(openapi.REPORTFILTERANDSORTFIELDS_NAME, "asc", 0, 2))
	require.Equal(suite.T(), []string{"c"}, listNames(openapi.REPORTFILTERANDSORTFIELDS_NAME, "asc", 2, 2))
}
