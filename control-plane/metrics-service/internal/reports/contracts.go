package reports

import (
	"context"
	"iter"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/database/models"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
)

type ResultMapIterator iter.Seq2[string, []string]

type Collectorer interface {
	Collect(report models.Report)
}

type PendingReportsGetter interface {
	GetPendingReports(stopCtx context.Context, limit int) ([]models.Report, *api.Error)
	GetTotalReportsByStatus(stopCtx context.Context) (map[string]int, *api.Error)
}

type ReportsUpdater interface {
	UpdateReportStatus(tenantId int32, reportId string, status api.ReportStatus, previousStatus *api.ReportStatus) error
	UpdateReportStatusWithMessage(tenantId int32, reportId string, status api.ReportStatus, message string, previousStatus *api.ReportStatus) error
}

type ReportsDeleter interface {
	DeleteReport(ctx context.Context, tenantId int32, reportId string) *api.Error
	GetPendingDeleteReports(stopCtx context.Context, limit int) ([]models.Report, *api.Error)
}

type S3Saver interface {
	// Save
	// data is a map of labelVal1,...,labelValN => []string{MetricsVal1, ..., MetricsValM}
	// labels length should equal N
	// metrics length should equal M
	Save(tenantId int32, id string, iterator ResultMapIterator, labels []string, metrics []string) error
}

type S3Deleter interface {
	Delete(id string) error
}

type ResultMapBuilder interface {
	Build(results map[string]*api.GetMetrics200Response, labels []string, metricDisplayNames []string, tenantId int32) map[string][]string
}

type ClusterClienter interface {
	GetClusterName(ctx context.Context, clusterId string) (string, error)
}

type WaitGrouper interface {
	Add(int)
	Done()
	Wait()
}
