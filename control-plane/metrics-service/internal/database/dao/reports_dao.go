package dao

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/database/models"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/queryapi"
	querydbenrich "github.com/run-ai/runai/runai-application-packages/queryapi/dbenrich"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/dbenrich"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm/clause"
)

type PostgresReportsDao struct{}

func NewPostgresReportsDao() *PostgresReportsDao {
	return &PostgresReportsDao{}
}

func (p PostgresReportsDao) CreateReport(ctx context.Context, tenantId int32, report models.Report) (models.Report, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to create report", nil)

	preInsert(ctx, tenantId, &report)

	err := runaidb.DB().WithContext(ctx).Transaction(func(tx orm.DB) error {
		if _, err := tx.Clauses(clause.Returning{
			Columns: []clause.Column{{Name: "status_updated_at"}},
		}).Create(&report); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed report creation: %v", err)
			return err
		}

		if _, err := tx.Create(report.Metrics); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed metrics creation: %v", err)
			return err
		}

		return nil
	})
	if err != nil {
		return models.Report{}, dbErr.WithDetails(err)
	}

	return report, nil
}

func preInsert(ctx context.Context, tenantId int32, report *models.Report) {
	report.ID = uuid.New()
	report.TenantId = tenantId
	report.CreatedBy = convutils.Deref(runaictx.GetString(ctx, runaictx.CfSubjectID), "")
	report.Status = string(api.REPORTSTATUS_PENDING)

	for _, metric := range report.Metrics {
		metric.ReportID = report.ID
	}
}

func (p PostgresReportsDao) GetReportsByIds(
	ctx context.Context,
	tenantId int32,
	reportIds []string,
	filterBy ...queryapi.Filter,
) ([]models.Report, *api.Error) {
	if len(reportIds) == 0 {
		return []models.Report{}, nil
	}

	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get reports", nil)

	reports := []models.Report{}
	query := runaidb.DB().WithContext(ctx).
		Preload("Metrics").
		Where("tenant_id = ?", tenantId).
		Where("id IN ?", reportIds)

	if len(filterBy) > 0 {
		query = query.WithEnricher(querydbenrich.Filters(filterBy))
	}

	if err := query.Find(&reports); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get reports by ids: %v", err)
		return nil, dbErr.WithDetails(err)
	}

	return reports, nil
}

func (p PostgresReportsDao) GetReport(ctx context.Context, tenantId int32, reportId string) (models.Report, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get report", nil)

	report := models.Report{}
	err := runaidb.DB().WithContext(ctx).
		Preload("Metrics").
		Where("tenant_id = ?", tenantId).
		Where("id = ?", reportId).
		First(&report)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get report with id %s: %v", reportId, err)
		return models.Report{}, dbErr.WithDetails(err)
	}

	return report, nil
}

func (p PostgresReportsDao) GetReports(
	ctx context.Context, tenantId int32, filterBy []queryapi.Filter, sortBy string, sortOrder string, offset int32, limit int32, search string,
) ([]models.Report, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get report", nil)

	reports := []models.Report{}
	query := runaidb.DB().WithContext(ctx).
		Preload("Metrics").
		Where("tenant_id = ?", tenantId).
		WithEnricher(querydbenrich.Filters(filterBy))

	if sortBy != "" {
		query = query.WithEnricher(querydbenrich.SortBy(sortBy, sortOrder))
	}
	query = query.Order("id")

	if offset != 0 {
		query = query.Offset(int(offset))
	}
	if limit != 0 {
		query = query.Limit(int(limit))
	}
	if search != "" {
		query = query.WithEnricher(dbenrich.TSSearch(search, "search_vector"))
	}

	if err := query.Find(&reports); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get reports: %v", err)
		return nil, dbErr.WithDetails(err)
	}

	return reports, nil
}

func (p PostgresReportsDao) CountReports(ctx context.Context, tenantId int32, filterBy []queryapi.Filter, search string) (int64, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to count report", nil)

	var count int64
	query := runaidb.DB().WithContext(ctx).
		Model(&models.Report{}).
		Where("tenant_id = ?", tenantId).
		WithEnricher(querydbenrich.Filters(filterBy))

	if search != "" {
		query = query.WithEnricher(dbenrich.TSSearch(search, "search_vector"))
	}

	if err := query.Count(&count); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count reports: %v", err)
		return 0, dbErr.WithDetails(err)
	}

	return count, nil
}

func (p PostgresReportsDao) DeleteReport(ctx context.Context, tenantId int32, reportId string) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to delete report", nil)

	report := models.Report{}

	if _, err := runaidb.DB().WithContext(ctx).
		Where("tenant_id = ?", tenantId).
		Where("id = ?", reportId).
		Delete(&report); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(fmt.Sprintf("Failed to delete report with id %s", reportId), err)
		return dbErr.WithDetails(err)
	}

	return nil
}

func (p PostgresReportsDao) UpdateReportStatus(tenantId int32, reportId string, status api.ReportStatus, previousStatus *api.ReportStatus) error {
	return p.UpdateReportStatusWithMessage(tenantId, reportId, status, "", previousStatus)
}

func (p PostgresReportsDao) UpdateReportStatusWithMessage(tenantId int32, reportId string, status api.ReportStatus, message string, previousStatus *api.ReportStatus) error {
	logging.Debugf("marking report as %s, tenant_id [%d], report_id: [%s]. report status message: %s", status, tenantId, reportId, message)

	report := models.Report{
		Status: string(status),
	}
	if message != "" {
		report.StatusMessage = message
	}

	query := runaidb.DB().
		Model(&report).
		Where("tenant_id = ?", tenantId).
		Where("id = ?", reportId)

	if previousStatus != nil {
		logging.Debugf("updating report status of tenant [%d] with id [%s] from %s to %s", tenantId, reportId, *previousStatus, status)
		query = query.
			Where("status = ?", *previousStatus)
	}

	rowsAffected, err := query.Updates(report)
	if err != nil {
		logging.Errorf("Failed to update report %s status to %s. err: %v", reportId, status, err)
		return err
	}

	if rowsAffected == 0 {
		logging.Warnf("Failed to update report %s status to %s. No rows affected", reportId, status)
		return fmt.Errorf("failed to update report %s status to %s. No rows affected", reportId, status)
	}
	logging.Debugf("marked report as %s, tenant_id [%d], report_id: [%s]", status, tenantId, reportId)

	return nil
}

func (p PostgresReportsDao) UpdatePendingDelete(ctx context.Context, tenantId int32, reportIds []string) (updated []string, apiErr *api.Error) {
	if len(reportIds) == 0 {
		return []string{}, nil
	}

	updatedReports := []models.Report{}
	query := runaidb.DB().WithContext(ctx).
		Model(&updatedReports).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
		Where("tenant_id = ?", tenantId).
		Where("id IN ?", reportIds).
		Where("status NOT IN ?", []api.ReportStatus{
			api.REPORTSTATUS_PROCESSING,
			api.REPORTSTATUS_DELETING,
		})

	if _, err := query.Updates(models.Report{
		Status:          string(api.REPORTSTATUS_PENDING_DELETE),
		StatusUpdatedAt: time.Now(),
	}); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update reports status to %s: %v", api.REPORTSTATUS_PENDING_DELETE, err)
		return nil, api.NewError(http.StatusInternalServerError, "Failed to update reports status", err)
	}

	updated = make([]string, 0, len(updatedReports))
	for _, report := range updatedReports {
		updated = append(updated, report.ID.String())
	}

	return updated, nil
}

func (p PostgresReportsDao) GetPendingReports(stopCtx context.Context, limit int) ([]models.Report, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get pending reports", nil)

	reports := []models.Report{}
	err := runaidb.DB().WithContext(stopCtx).
		Preload("Metrics").
		Where("status = ?", api.REPORTSTATUS_PENDING).
		Limit(limit).
		WithEnricher(querydbenrich.SortBy("created_at", "asc")).
		Find(&reports)
	if err != nil {
		logging.Errorf("Failed to get pending reports. err: %v", err)
		return nil, dbErr.WithDetails(err)
	}

	return reports, nil
}

func (p PostgresReportsDao) GetPendingDeleteReports(stopCtx context.Context, limit int) ([]models.Report, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get pending-delete reports", nil)

	reports := []models.Report{}
	err := runaidb.DB().WithContext(stopCtx).
		Where("status = ?", api.REPORTSTATUS_PENDING_DELETE).
		Limit(limit).
		WithEnricher(querydbenrich.SortBy("status_updated_at", "asc")).
		Find(&reports)
	if err != nil {
		logging.Errorf("Failed to get pending-delete reports. err: %v", err)
		return nil, dbErr.WithDetails(err)
	}

	return reports, nil
}

type statusCount struct {
	Status string
	Count  int
}

func (p PostgresReportsDao) GetTotalReportsByStatus(ctx context.Context) (map[string]int, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get total reports by status", nil)

	var statusCounts []statusCount

	if _, err := runaidb.DB().WithContext(ctx).Table("reports").
		Select("status, COUNT(*) AS count").
		Group("status").
		Scan(&statusCounts); err != nil {
		logging.Errorf("Failed to get total reports by status. err: %v", err)
		return nil, dbErr.WithDetails(err)
	}

	return convertToStatusMap(statusCounts), nil
}

func convertToStatusMap(statusCounts []statusCount) map[string]int {
	statusMap := make(map[string]int)
	for _, sc := range statusCounts {
		statusMap[sc.Status] = sc.Count
	}

	return statusMap
}
