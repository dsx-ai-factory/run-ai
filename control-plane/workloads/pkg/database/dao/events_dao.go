package dao

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	EventsTableName               = "events"
	EventsViewName                = "workloads_events_view"
	RevisionEventsViewName        = "revisions_events_view"
	PhaseHistoryTableName         = "workloads_phase_history"
	RevisionPhaseHistoryTableName = "revisions_phase_history"
	WorkloadTypeStr               = "workload"
	RevisionTypeStr               = "revision"

	SortOrderAsc  = "asc"
	SortOrderDesc = "desc"
)

func DbSyncEvents(ctx context.Context, events []models.Event) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to sync events", nil)

	tx := dao_utils.DB().WithContext(ctx)

	var result *gorm.DB
	if len(events) > 0 {
		result = tx.Table(EventsTableName).Clauses(clause.OnConflict{DoNothing: true}).Create(&events)
		if result.Error != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create the events: %v", result.Error)
			return dbErr.WithDetails(result.Error)
		}
	}

	return nil
}

func dbGetWorkloadOrRevisionEvents[T any](ctx context.Context, tenantId uint, workloadOrRevisionType string, workloadOrRevisionId string,
	tableName string, offset int32, limit int32, sortOrder string) (*[]T, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to list %s events", workloadOrRevisionType), nil)

	if workloadOrRevisionType != WorkloadTypeStr && workloadOrRevisionType != RevisionTypeStr {
		// not expected to happen, this is not a user input
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Invalid workloadOrRevisionType: %s", workloadOrRevisionType)
		return &[]T{}, dbErr
	}

	tx := dao_utils.DB().WithContext(ctx)

	events := &[]T{}

	query := tx.
		Table(tableName).
		Where("tenant_id = ?", tenantId).
		Where(fmt.Sprintf("%s_id = ?", workloadOrRevisionType), workloadOrRevisionId)

	switch strings.ToLower(sortOrder) {
	case SortOrderAsc, SortOrderDesc:
		query = query.Order(fmt.Sprintf("created_at %s", sortOrder))
	default:
		return &[]T{}, api.NewError(http.StatusBadRequest, "Invalid sort order", nil)
	}

	if offset != 0 {
		query = query.Offset(int(offset))
	}

	if limit != 0 {
		query = query.Limit(int(limit))
	}

	result := query.Find(events)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(fmt.Sprintf("Failed to list %s events", workloadOrRevisionType), result.Error)
		return &[]T{}, dbErr.WithDetails(result.Error)
	}

	return events, nil
}

func DbGetWorkloadEvents(ctx context.Context, tenantId uint, workloadId string, offset int32, limit int32, sortOrder string) (*[]models.WorkloadEventView, *api.Error) {
	return dbGetWorkloadOrRevisionEvents[models.WorkloadEventView](
		ctx, tenantId, WorkloadTypeStr, workloadId, EventsViewName, offset, limit, sortOrder,
	)
}

func DbGetRevisionEvents(ctx context.Context, tenantId uint, revisionId string, offset int32, limit int32, sortOrder string) (*[]models.RevisionEventView, *api.Error) {
	return dbGetWorkloadOrRevisionEvents[models.RevisionEventView](
		ctx, tenantId, RevisionTypeStr, revisionId, RevisionEventsViewName, offset, limit, sortOrder,
	)
}

func dbGetWorkloadOrRevisionHistoryRecords[T any](ctx context.Context, tenantId uint, workloadOrRevisionType string, workloadOrRevisionId string,
	eventsViewName string, phaseHistoryTableName string, offset int32, limit int32) (*[]T, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to list %s history records", workloadOrRevisionType), nil)

	if workloadOrRevisionType != WorkloadTypeStr && workloadOrRevisionType != RevisionTypeStr {
		// not expected to happen, this is not a user input
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Invalid workloadOrRevisionType: %s", workloadOrRevisionType)
		return &[]T{}, dbErr
	}

	tx := dao_utils.DB().WithContext(ctx)

	historyRecords := &[]T{}

	idColumn := fmt.Sprintf("%s_id", workloadOrRevisionType)

	query := tx.Table(fmt.Sprintf("(SELECT %s,", idColumn)+
		" created_at,"+
		" cluster_id,"+
		" tenant_id,"+
		" event_id,"+
		" type,"+
		" message,"+
		" reason,"+
		" source,"+
		" involved_object_uid,"+
		" involved_object_kind,"+
		" involved_object_name,"+
		fmt.Sprintf(" involved_object_namespace FROM %s WHERE tenant_id = ? AND %s = ?) as e", eventsViewName, idColumn), tenantId, workloadOrRevisionId).
		Joins(fmt.Sprintf("FULL OUTER JOIN (SELECT %s, created_at, cluster_id, tenant_id, phase, phase_message FROM %s WHERE tenant_id = ? AND %s = ?) as p on 1=0",
			idColumn, phaseHistoryTableName, idColumn), tenantId, workloadOrRevisionId)

	if offset != 0 {
		query = query.Offset(int(offset))
	}

	if limit != 0 {
		query = query.Limit(int(limit))
	}

	query = query.Select(
		fmt.Sprintf("coalesce(e.%s, p.%s) as %s,", idColumn, idColumn, idColumn) +
			"coalesce(e.created_at, p.created_at) as created_at," +
			"coalesce(e.cluster_id, p.cluster_id) as cluster_id," +
			"coalesce(e.tenant_id, p.tenant_id) as tenant_id," +
			"e.event_id," +
			"e.type," +
			"e.message," +
			"e.reason," +
			"e.source," +
			"e.involved_object_uid," +
			"e.involved_object_kind," +
			"e.involved_object_name," +
			"e.involved_object_namespace," +
			"p.phase," +
			"p.phase_message").
		Order("created_at desc")

	result := query.
		Find(historyRecords)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(fmt.Sprintf("Failed to list %s history records", workloadOrRevisionType), result.Error)
		return historyRecords, dbErr.WithDetails(result.Error)
	}

	return historyRecords, nil
}

// DbGetWorkloadHistoryRecords returns the history records of a workload, ordered by creation time in descending order.
func DbGetWorkloadHistoryRecords(ctx context.Context, tenantId uint, workloadId string, offset int32, limit int32) (*[]models.WorkloadHistoryRecord, *api.Error) {
	return dbGetWorkloadOrRevisionHistoryRecords[models.WorkloadHistoryRecord](ctx, tenantId, WorkloadTypeStr, workloadId, EventsViewName, PhaseHistoryTableName, offset, limit)
}

// DbGetRevisionHistoryRecords returns the history records of a revision, ordered by creation time in descending order.
func DbGetRevisionHistoryRecords(ctx context.Context, tenantId uint, revisionId string, offset int32, limit int32) (*[]models.RevisionHistoryRecord, *api.Error) {
	return dbGetWorkloadOrRevisionHistoryRecords[models.RevisionHistoryRecord](ctx, tenantId, RevisionTypeStr, revisionId, RevisionEventsViewName, RevisionPhaseHistoryTableName, offset, limit)
}

// DbGetTenantWorkloadPhaseHistoryRecords returns the phase history records of all the workloads in the tenant,
// since createdAfter, ordered by creation time in ascending order.
func DbGetTenantWorkloadPhaseHistoryRecords(ctx context.Context, tenantId string, createdAfter time.Time) (*[]models.WorkloadPhasesView, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to list tenant workloads phase history records", nil)

	tx := dao_utils.DB().WithContext(ctx)

	phasesRecords := &[]models.WorkloadPhasesView{}

	result := tx.Table(models.WorkloadPhasesView{}.TableName()).
		Where("tenant_id = ?", tenantId).
		Where("created_at > ?", createdAfter).
		Order("created_at asc").
		Find(phasesRecords)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list tenant workloads phase history records: %v", result.Error)
		return phasesRecords, dbErr.WithDetails(result.Error)
	}

	return phasesRecords, nil
}

// DbGetTimeoutEventsWithReason returns all the workload events in the tenant with the given reason, since createdAfter.
func DbGetTimeoutEventsWithReason(ctx context.Context, tenantId string, reason string, createdAfter time.Time) (*[]models.TimeoutEventsOfWorkloadView, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to list tenant events with reason", nil)

	tx := dao_utils.DB().WithContext(ctx)

	events := &[]models.TimeoutEventsOfWorkloadView{}

	result := tx.Table(models.TimeoutEventsOfWorkloadView{}.TableName()).
		Where("tenant_id = ?", tenantId).
		Where("reason = ?", reason).
		Where("created_at > ?", createdAfter).
		Find(events)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list tenant events with reason: %v", result.Error)
		return events, dbErr.WithDetails(result.Error)
	}

	return events, nil
}

// DbDeleteWorkloadEventsOfClusterId deletes all workload events within a cluster
func DbDeleteWorkloadEventsOfClusterId(ctx context.Context, tenantId string, clusterId string) error {
	// By default, GORM's Delete() method only performs a soft delete if the model has gorm.DeletedAt.
	// Using .Unscoped() ensures a hard delete, permanently removing the record from the database.
	result := dao_utils.DB().WithContext(ctx).Unscoped().
		Table(EventsTableName).
		Where("tenant_id = ?", tenantId).
		Where("cluster_id = ?", clusterId).
		Delete(&models.Event{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete workload event for cluster %s in tenant %s. err: %w", clusterId, tenantId, result.Error)
	}
	return nil
}

// DbDeleteWorkloadEventsOfTenantId deletes all workload events within a tenant
func DbDeleteWorkloadEventsOfTenantId(ctx context.Context, tenantId string) error {
	// By default, GORM's Delete() method only performs a soft delete if the model has gorm.DeletedAt.
	// Using .Unscoped() ensures a hard delete, permanently removing the record from the database.
	result := dao_utils.DB().WithContext(ctx).Unscoped().
		Table(EventsTableName).
		Where("tenant_id = ?", tenantId).
		Delete(&models.Event{})
	if result.Error != nil {
		return fmt.Errorf("failed to workload delete event in tenant %s. err: %w", tenantId, result.Error)
	}
	return nil
}
