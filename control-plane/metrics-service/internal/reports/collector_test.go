package reports_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/reports"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/database/models"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/reports/mocks"
	service_mocks "github.com/run-ai/runai/control-plane/metrics-service/internal/services/mocks"
	"github.com/stretchr/testify/mock"
)

var report = models.Report{
	ID:       uuid.New(),
	TenantId: 1,
	Start:    time.Now(),
	End:      time.Now().Add(14 * time.Hour),
	Metrics: []*models.Metric{{
		ReportID:    uuid.New(),
		DisplayName: "cpu",
		Query:       "cpu_query",
	}, {
		ReportID:    uuid.New(),
		DisplayName: "gpu",
		Query:       "gpu_query",
	}},
	Labels: []models.Label{{
		Name:        "l1",
		DisplayName: "label1",
	}, {
		Name:        "l2",
		DisplayName: "label2",
	}},
}

func TestCollect_QueryFailed(t *testing.T) {
	ru := mocks.NewReportsUpdater(t)
	mq := service_mocks.NewMetricQuerier(t)
	s3Saver := mocks.NewS3Saver(t)
	rmb := mocks.NewResultMapBuilder(t)
	collector := &reports.Collector{
		ReportsUpdater:   ru,
		MetricQuerier:    mq,
		S3Saver:          s3Saver,
		ResultMapBuilder: rmb,
	}

	ru.On("UpdateReportStatus", int32(1), report.ID.String(), api.REPORTSTATUS_PROCESSING, mock.Anything).Return(nil).Once()
	mq.On("Query", mock.Anything, int32(1), "cpu_query", mock.Anything).
		Return(nil, fmt.Errorf("failed")).Once()
	ru.On("UpdateReportStatusWithMessage", int32(1), report.ID.String(), api.REPORTSTATUS_FAILED, mock.Anything, (*api.ReportStatus)(nil)).Return(nil).Once()

	collector.Collect(report)
}

func TestCollect_S3Failed(t *testing.T) {
	ru := mocks.NewReportsUpdater(t)
	mq := service_mocks.NewMetricQuerier(t)
	s3Saver := mocks.NewS3Saver(t)
	rmb := mocks.NewResultMapBuilder(t)
	collector := &reports.Collector{
		ReportsUpdater:   ru,
		MetricQuerier:    mq,
		S3Saver:          s3Saver,
		ResultMapBuilder: rmb,
	}

	ru.On("UpdateReportStatus", int32(1), report.ID.String(), api.REPORTSTATUS_PROCESSING, mock.Anything).Return(nil).Once()
	mq.On("Query", mock.Anything, int32(1), "cpu_query", mock.Anything).
		Return(&api.GetMetrics200Response{}, nil).Once()
	mq.On("Query", mock.Anything, int32(1), "gpu_query", mock.Anything).
		Return(&api.GetMetrics200Response{}, nil).Once()
	rmb.On("Build", mock.Anything, mock.Anything, mock.Anything, int32(1)).Return(map[string][]string{}).Once()
	s3Saver.On("Save", int32(1), report.ID.String(), mock.AnythingOfType("reports.ResultMapIterator"), []string{"label1", "label2"}, mock.Anything).Return(fmt.Errorf("failed")).Once()
	ru.On("UpdateReportStatusWithMessage", int32(1), report.ID.String(), api.REPORTSTATUS_FAILED, mock.Anything, (*api.ReportStatus)(nil)).Return(nil).Once()

	collector.Collect(report)
}

func TestCollect_Succeed(t *testing.T) {
	ru := mocks.NewReportsUpdater(t)
	mq := service_mocks.NewMetricQuerier(t)
	s3Saver := mocks.NewS3Saver(t)
	rmb := mocks.NewResultMapBuilder(t)
	collector := &reports.Collector{
		ReportsUpdater:   ru,
		MetricQuerier:    mq,
		S3Saver:          s3Saver,
		ResultMapBuilder: rmb,
	}

	ru.On("UpdateReportStatus", int32(1), report.ID.String(), api.REPORTSTATUS_PROCESSING, mock.Anything).Return(nil).Once()
	mq.On("Query", mock.Anything, int32(1), "cpu_query", mock.Anything).
		Return(&api.GetMetrics200Response{}, nil).Once()
	mq.On("Query", mock.Anything, int32(1), "gpu_query", mock.Anything).
		Return(&api.GetMetrics200Response{}, nil).Once()
	rmb.On("Build", mock.Anything, mock.Anything, mock.Anything, int32(1)).Return(map[string][]string{}).Once()
	s3Saver.On("Save", int32(1), report.ID.String(), mock.AnythingOfType("reports.ResultMapIterator"), []string{"label1", "label2"}, mock.Anything).Return(nil).Once()
	ru.On("UpdateReportStatus", int32(1), report.ID.String(), api.REPORTSTATUS_READY, mock.Anything).Return(nil).Once()

	collector.Collect(report)
}

func TestCollect_MetricsOrderPreserved(t *testing.T) {
	// Create a report with 10 metrics in non-alphabetical order
	reportWithMetrics := models.Report{
		ID:       uuid.New(),
		TenantId: 1,
		Start:    time.Now(),
		End:      time.Now().Add(14 * time.Hour),
		Metrics: []*models.Metric{
			{ReportID: uuid.New(), DisplayName: "Z-Metric", Query: "z_query"},
			{ReportID: uuid.New(), DisplayName: "Y-Metric", Query: "y_query"},
			{ReportID: uuid.New(), DisplayName: "M-Metric", Query: "m_query"},
			{ReportID: uuid.New(), DisplayName: "A-Metric", Query: "a_query"},
			{ReportID: uuid.New(), DisplayName: "B-Metric", Query: "b_query"},
			{ReportID: uuid.New(), DisplayName: "X-Metric", Query: "x_query"},
			{ReportID: uuid.New(), DisplayName: "C-Metric", Query: "c_query"},
			{ReportID: uuid.New(), DisplayName: "D-Metric", Query: "d_query"},
			{ReportID: uuid.New(), DisplayName: "E-Metric", Query: "e_query"},
			{ReportID: uuid.New(), DisplayName: "F-Metric", Query: "f_query"},
		},
		Labels: []models.Label{
			{Name: "l1", DisplayName: "label1"},
			{Name: "l2", DisplayName: "label2"},
		},
	}

	// Extract expected order of metrics from report.Metrics
	expectedMetricsOrder := make([]string, len(reportWithMetrics.Metrics))
	for i, metric := range reportWithMetrics.Metrics {
		expectedMetricsOrder[i] = metric.DisplayName
	}

	ru := mocks.NewReportsUpdater(t)
	mq := service_mocks.NewMetricQuerier(t)
	s3Saver := mocks.NewS3Saver(t)
	rmb := mocks.NewResultMapBuilder(t)
	collector := &reports.Collector{
		ReportsUpdater:   ru,
		MetricQuerier:    mq,
		S3Saver:          s3Saver,
		ResultMapBuilder: rmb,
	}

	// Mock ReportsUpdater to return nil error
	ru.On("UpdateReportStatus", int32(1), reportWithMetrics.ID.String(), api.REPORTSTATUS_PROCESSING, mock.Anything).Return(nil).Once()

	// Mock MetricQuerier for all 10 metrics
	for _, metric := range reportWithMetrics.Metrics {
		mq.On("Query", mock.Anything, int32(1), metric.Query, mock.Anything).
			Return(&api.GetMetrics200Response{}, nil).Once()
	}

	// Mock ResultMapBuilder
	rmb.On("Build", mock.Anything, mock.Anything, mock.Anything, int32(1)).Return(map[string][]string{}).Once()

	// Mock S3Saver and capture the displayMetricsNames argument
	var capturedDisplayMetricsNames []string
	s3Saver.On("Save", int32(1), reportWithMetrics.ID.String(), mock.AnythingOfType("reports.ResultMapIterator"), []string{"label1", "label2"}, mock.Anything).
		Run(func(args mock.Arguments) {
			// Capture the 5th argument (displayMetricsNames)
			capturedDisplayMetricsNames = args.Get(4).([]string)
		}).
		Return(nil).Once()

	ru.On("UpdateReportStatus", int32(1), reportWithMetrics.ID.String(), api.REPORTSTATUS_READY, mock.Anything).Return(nil).Once()

	// Execute Collect
	collector.Collect(reportWithMetrics)

	// Verify that the displayMetricsNames passed to S3Saver matches the expected order
	if len(capturedDisplayMetricsNames) != len(expectedMetricsOrder) {
		t.Fatalf("Expected %d metrics, but got %d", len(expectedMetricsOrder), len(capturedDisplayMetricsNames))
	}

	for i, expectedName := range expectedMetricsOrder {
		if capturedDisplayMetricsNames[i] != expectedName {
			t.Errorf("Metric at position %d: expected '%s', but got '%s'", i, expectedName, capturedDisplayMetricsNames[i])
		}
	}
}
