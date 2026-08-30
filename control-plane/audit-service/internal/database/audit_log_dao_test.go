package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/audit-service/internal/database/models"
	"github.com/run-ai/runai/runai-application-packages/queryapi"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

const (
	testTenantID = 1

	sqlInsertAuditLog     = `INSERT INTO audit_logs (created_at,updated_at,deleted_at,log_id,tenant_id,cluster_id,cluster_name,request_id,timestamp,subject,subject_type,source_ip,action,http_method,url,entity_type,entity_name,entity_id,http_status_code,result,origin,metadata) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) RETURNING id`
	sqlSelectAuditLog     = `SELECT * FROM audit_logs WHERE audit_logs.tenant_id = ? AND audit_logs.deleted_at IS NULL`
	sqlCountAuditLog      = `SELECT count(*) FROM audit_logs WHERE audit_logs.tenant_id = ?`
	sqlHardDeleteAuditLog = `DELETE FROM audit_logs WHERE tenant_id = ?`

	sqlSelectFilter = `SELECT * FROM audit_logs WHERE audit_logs.tenant_id = ? AND subject = ? AND audit_logs.deleted_at IS NULL`
)

type insertedID struct {
	ID uint
}

type countResult struct {
	Count int64
}

func newTestAuditLog() *models.AuditLog {
	return &models.AuditLog{
		LogId:          uuid.New(),
		TenantId:       testTenantID,
		ClusterId:      uuid.New(),
		ClusterName:    "test-cluster",
		RequestId:      uuid.New(),
		Timestamp:      time.Now(),
		Subject:        "user@example.com",
		SubjectType:    "user",
		SourceIp:       "10.0.0.1",
		Action:         "CREATE",
		HttpMethod:     "POST",
		Url:            "/api/v1/projects",
		EntityType:     "project",
		EntityName:     "my-project",
		EntityId:       "123",
		HttpStatusCode: 200,
		Result:         "success",
		Origin:         "ui",
	}
}

func TestStore(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectQuery(sqlInsertAuditLog).Returns([]insertedID{{ID: 1}})
	mocker.ExpectTransaction().Commit()

	err := NewAuditLogsDao().Store(context.Background(), []*models.AuditLog{newTestAuditLog()})
	require.NoError(t, err)
}

func TestStore_Error(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectQuery(sqlInsertAuditLog).ReturnsError(errors.New("insert failed"))
	mocker.ExpectTransaction().Rollback()

	err := NewAuditLogsDao().Store(context.Background(), []*models.AuditLog{newTestAuditLog()})
	require.EqualError(t, err, "insert failed")
}

func TestRead(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	expected := newTestAuditLog()
	mocker.ExpectQuery(sqlSelectAuditLog, testTenantID).Returns([]models.AuditLog{*expected})

	logs, count, err := NewAuditLogsDao().Read(context.Background(), testTenantID, QueryParams{})
	require.NoError(t, err)
	require.Equal(t, 0, count)
	require.Len(t, logs, 1)
	require.Equal(t, expected.Subject, logs[0].Subject)
	require.Equal(t, expected.Action, logs[0].Action)
}

func TestRead_WithTotalCount(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	expected := newTestAuditLog()
	mocker.ExpectQuery(sqlCountAuditLog, testTenantID).Returns([]countResult{{Count: 1}})
	mocker.ExpectQuery(sqlSelectAuditLog, testTenantID).Returns([]models.AuditLog{*expected})

	logs, count, err := NewAuditLogsDao().Read(context.Background(), testTenantID, QueryParams{WithTotalCount: true})
	require.NoError(t, err)
	require.Equal(t, 1, count)
	require.Len(t, logs, 1)
}

func TestRead_Empty(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectQuery(sqlSelectAuditLog, testTenantID).Returns([]models.AuditLog{})

	logs, count, err := NewAuditLogsDao().Read(context.Background(), testTenantID, QueryParams{})
	require.NoError(t, err)
	require.Equal(t, 0, count)
	require.Empty(t, logs)
}

func TestRead_Error(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectQuery(sqlSelectAuditLog, testTenantID).ReturnsError(errors.New("db error"))

	logs, count, err := NewAuditLogsDao().Read(context.Background(), testTenantID, QueryParams{})
	require.EqualError(t, err, "db error")
	require.Empty(t, logs)
	require.Equal(t, 0, count)
}

func TestRead_Offset(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	expected := newTestAuditLog()
	mocker.ExpectQuery(`SELECT * FROM audit_logs WHERE audit_logs.tenant_id = ? AND audit_logs.deleted_at IS NULL OFFSET ?`, testTenantID, 10).Returns([]models.AuditLog{*expected})

	logs, _, err := NewAuditLogsDao().Read(context.Background(), testTenantID, QueryParams{Offset: 10})
	require.NoError(t, err)
	require.Len(t, logs, 1)
	requireLogEqual(t, expected, logs[0])
}

func TestRead_Samples(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	first, second := newTestAuditLog(), newTestAuditLog()
	mocker.ExpectQuery(`SELECT * FROM audit_logs WHERE audit_logs.tenant_id = ? AND audit_logs.deleted_at IS NULL LIMIT ?`, testTenantID, 2).Returns([]models.AuditLog{*first, *second})

	logs, _, err := NewAuditLogsDao().Read(context.Background(), testTenantID, QueryParams{Samples: 2})
	require.NoError(t, err)
	require.Len(t, logs, 2)
	requireLogEqual(t, first, logs[0])
	requireLogEqual(t, second, logs[1])
}

func TestRead_OffsetAndSamples(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	expected := newTestAuditLog()
	mocker.ExpectQuery(`SELECT * FROM audit_logs WHERE audit_logs.tenant_id = ? AND audit_logs.deleted_at IS NULL LIMIT ? OFFSET ?`, testTenantID, 5, 10).Returns([]models.AuditLog{*expected})

	logs, _, err := NewAuditLogsDao().Read(context.Background(), testTenantID, QueryParams{Offset: 10, Samples: 5})
	require.NoError(t, err)
	require.Len(t, logs, 1)
	requireLogEqual(t, expected, logs[0])
}

func TestRead_SortByAsc(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	expected := newTestAuditLog()
	mocker.ExpectQuery(`SELECT * FROM audit_logs WHERE audit_logs.tenant_id = ? AND audit_logs.deleted_at IS NULL ORDER BY timestamp ASC NULLS FIRST`, testTenantID).Returns([]models.AuditLog{*expected})

	logs, _, err := NewAuditLogsDao().Read(context.Background(), testTenantID, QueryParams{SortBy: "timestamp", Order: "asc"})
	require.NoError(t, err)
	require.Len(t, logs, 1)
	requireLogEqual(t, expected, logs[0])
}

func TestRead_SortByDesc(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	expected := newTestAuditLog()
	mocker.ExpectQuery(`SELECT * FROM audit_logs WHERE audit_logs.tenant_id = ? AND audit_logs.deleted_at IS NULL ORDER BY timestamp DESC NULLS LAST`, testTenantID).Returns([]models.AuditLog{*expected})

	logs, _, err := NewAuditLogsDao().Read(context.Background(), testTenantID, QueryParams{SortBy: "timestamp", Order: "desc"})
	require.NoError(t, err)
	require.Len(t, logs, 1)
	requireLogEqual(t, expected, logs[0])
}

func TestRead_SingleFilter(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	expected := newTestAuditLog()
	mocker.ExpectQuery(sqlSelectFilter, testTenantID, "user@example.com").Returns([]models.AuditLog{*expected})

	f := queryapi.Filter{Field: "subject", Operator: "=", ValueFormat: "%s", Values: []string{"user@example.com"}}
	logs, _, err := NewAuditLogsDao().Read(context.Background(), testTenantID, QueryParams{Filters: []queryapi.Filter{f}})
	require.NoError(t, err)
	require.Len(t, logs, 1)
	requireLogEqual(t, expected, logs[0])
}

func TestRead_MultiValueFilter(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	expected := newTestAuditLog()
	mocker.ExpectQuery(`SELECT * FROM audit_logs WHERE audit_logs.tenant_id = ? AND (action = ? OR action = ?) AND audit_logs.deleted_at IS NULL`, testTenantID, "CREATE", "DELETE").Returns([]models.AuditLog{*expected})

	f := queryapi.Filter{Field: "action", Operator: "=", ValueFormat: "%s", Values: []string{"CREATE", "DELETE"}}
	logs, _, err := NewAuditLogsDao().Read(context.Background(), testTenantID, QueryParams{Filters: []queryapi.Filter{f}})
	require.NoError(t, err)
	require.Len(t, logs, 1)
	requireLogEqual(t, expected, logs[0])
}

func TestRead_ContainsFilter(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	expected := newTestAuditLog()
	mocker.ExpectQuery(sqlSelectFilter, testTenantID, "%adm%").Returns([]models.AuditLog{*expected})

	f := queryapi.Filter{Field: "subject", Operator: "=", ValueFormat: "%%%s%%", Values: []string{"adm"}}
	logs, _, err := NewAuditLogsDao().Read(context.Background(), testTenantID, QueryParams{Filters: []queryapi.Filter{f}})
	require.NoError(t, err)
	require.Len(t, logs, 1)
	requireLogEqual(t, expected, logs[0])
}

func TestRead_EmptyFilterValuesSkipped(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	expected := newTestAuditLog()
	mocker.ExpectQuery(sqlSelectAuditLog, testTenantID).Returns([]models.AuditLog{*expected})

	f := queryapi.Filter{Field: "subject", Operator: "=", ValueFormat: "%s", Values: []string{}}
	logs, _, err := NewAuditLogsDao().Read(context.Background(), testTenantID, QueryParams{Filters: []queryapi.Filter{f}})
	require.NoError(t, err)
	require.Len(t, logs, 1)
	requireLogEqual(t, expected, logs[0])
}

func TestRead_TimeRange(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	expected := newTestAuditLog()
	mocker.ExpectQuery(`SELECT * FROM audit_logs WHERE audit_logs.tenant_id = ? AND timestamp >= ? AND timestamp <= ? AND audit_logs.deleted_at IS NULL`, testTenantID, dbtesting.AnyTime(), dbtesting.AnyTime()).Returns([]models.AuditLog{*expected})

	start := time.Now()
	end := start.Add(time.Hour)
	logs, _, err := NewAuditLogsDao().Read(context.Background(), testTenantID, QueryParams{Start: &start, End: &end})
	require.NoError(t, err)
	require.Len(t, logs, 1)
	requireLogEqual(t, expected, logs[0])
}

func TestRead_AllParams(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	expected := newTestAuditLog()
	countSQL := `SELECT count(*) FROM audit_logs WHERE audit_logs.tenant_id = ? AND timestamp >= ? AND timestamp <= ? AND subject = ?`
	mocker.ExpectQuery(countSQL, testTenantID, dbtesting.AnyTime(), dbtesting.AnyTime(), "user@example.com").Returns([]countResult{{Count: 1}})
	mocker.ExpectQuery(`SELECT * FROM audit_logs WHERE audit_logs.tenant_id = ? AND timestamp >= ? AND timestamp <= ? AND subject = ? AND audit_logs.deleted_at IS NULL ORDER BY timestamp DESC NULLS LAST LIMIT ? OFFSET ?`, testTenantID, dbtesting.AnyTime(), dbtesting.AnyTime(), "user@example.com", 5, 10).Returns([]models.AuditLog{*expected})

	start := time.Now()
	end := start.Add(time.Hour)
	f := queryapi.Filter{Field: "subject", Operator: "=", ValueFormat: "%s", Values: []string{"user@example.com"}}
	logs, count, err := NewAuditLogsDao().Read(context.Background(), testTenantID, QueryParams{
		Offset: 10, Samples: 5, SortBy: "timestamp", Order: "desc",
		Start: &start, End: &end, Filters: []queryapi.Filter{f}, WithTotalCount: true,
	})
	require.NoError(t, err)
	require.Equal(t, 1, count)
	require.Len(t, logs, 1)
	requireLogEqual(t, expected, logs[0])
}

func TestHardDeleteTenantAuditLogs(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlHardDeleteAuditLog, testTenantID).Returns(1)
	mocker.ExpectTransaction().Commit()

	err := NewAuditLogsDao().HardDeleteTenantAuditLogs(context.Background(), testTenantID)
	require.NoError(t, err)
}

func TestHardDeleteTenantAuditLogs_Error(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlHardDeleteAuditLog, testTenantID).ReturnsError(errors.New("delete failed"))
	mocker.ExpectTransaction().Rollback()

	err := NewAuditLogsDao().HardDeleteTenantAuditLogs(context.Background(), testTenantID)
	require.EqualError(t, err, "delete failed")
}

func requireLogEqual(t *testing.T, expected, actual *models.AuditLog) {
	require.Equal(t, expected.LogId, actual.LogId)
	require.Equal(t, expected.Subject, actual.Subject)
	require.Equal(t, expected.Action, actual.Action)
	require.Equal(t, expected.EntityType, actual.EntityType)
	require.Equal(t, expected.HttpStatusCode, actual.HttpStatusCode)
}
