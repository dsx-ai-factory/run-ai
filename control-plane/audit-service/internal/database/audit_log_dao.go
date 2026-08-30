package database

import (
	"context"
	"time"

	"github.com/run-ai/runai/control-plane/audit-service/internal/database/models"
	"github.com/run-ai/runai/runai-application-packages/queryapi"
	querydbenrich "github.com/run-ai/runai/runai-application-packages/queryapi/dbenrich"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

type QueryParams struct {
	Offset         int
	Samples        int
	Order          string
	SortBy         string
	Start          *time.Time
	End            *time.Time
	Filters        []queryapi.Filter
	WithTotalCount bool
}

type AuditLogDao struct{}

func NewAuditLogsDao() *AuditLogDao {
	return &AuditLogDao{}
}

func (ald *AuditLogDao) Store(ctx context.Context, batch []*models.AuditLog) error {
	_, err := runaidb.DB().WithContext(ctx).Table(models.AuditLogsTable).Create(batch)
	return err
}

func (ald *AuditLogDao) Read(ctx context.Context, tenantId int, qp QueryParams) ([]*models.AuditLog, int, error) {
	query := runaidb.DB().WithContext(ctx).
		Table(models.AuditLogsTable).
		Where(models.AuditLog{TenantId: tenantId})
	query = addFilteringParams(query, qp)
	totalCount := int64(0)
	if qp.WithTotalCount {
		err := query.Count(&totalCount)
		if err != nil {
			return nil, 0, err
		}
	}
	addDisplayingParams(query, qp)
	logs := make([]*models.AuditLog, 0)
	err := query.Find(&logs)
	if err != nil {
		return nil, 0, err
	}
	logging.Infof("count is %v", totalCount)
	return logs, int(totalCount), nil
}

func (ald *AuditLogDao) HardDeleteTenantAuditLogs(ctx context.Context, tenantId int) error {
	_, err := runaidb.DB().Unscoped().WithContext(ctx).Table(models.AuditLogsTable).Where("tenant_id = ?", tenantId).Delete(&models.AuditLog{})
	return err
}

func addFilteringParams(query orm.DB, qp QueryParams) orm.DB {
	query = query.WithEnricher(querydbenrich.Filters(qp.Filters))
	return applyTimestampWhereQuery(query, qp.Start, qp.End)
}

func addDisplayingParams(query orm.DB, qp QueryParams) {
	if qp.Offset > 0 {
		query.Offset(qp.Offset)
	}
	if qp.Samples > 0 {
		query.Limit(qp.Samples)
	}
	if qp.SortBy != "" && qp.Order != "" {
		query.WithEnricher(querydbenrich.SortBy(qp.SortBy, qp.Order))
	}
}

func applyTimestampWhereQuery(query orm.DB, start, end *time.Time) orm.DB {
	if start != nil {
		query = query.Where("timestamp >= ?", start)
	}
	if end != nil {
		query = query.Where("timestamp <= ?", end)
	}
	return query
}
