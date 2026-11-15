/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/generated/pkg/api"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

const (
	sqlSelectReportMetrics  = `SELECT * FROM report_metrics_queries WHERE report_metrics_queries.report_id = ?`
	sqlSelectReport         = `SELECT * FROM reports WHERE tenant_id = ? AND id = ? ORDER BY reports.id LIMIT ?`
	sqlSelectReportsByIds   = `SELECT * FROM reports WHERE tenant_id = ? AND id IN (?,?)`
	sqlDeleteReport         = `DELETE FROM reports WHERE tenant_id = ? AND id = ?`
	sqlUpdateStatus         = `UPDATE reports SET status=? WHERE tenant_id = ? AND id = ?`
	sqlUpdateStatusGuarded  = `UPDATE reports SET status=? WHERE tenant_id = ? AND id = ? AND status = ?`
	sqlUpdateStatusMessage  = `UPDATE reports SET status=?,status_message=? WHERE tenant_id = ? AND id = ?`
	sqlUpdatePendingDelete  = `UPDATE reports SET status=?,status_updated_at=? WHERE tenant_id = ? AND id IN (?,?) AND status NOT IN (?,?) RETURNING id`
	sqlSelectPending        = `SELECT * FROM reports WHERE status = ? ORDER BY created_at ASC NULLS FIRST LIMIT ?`
	sqlSelectPendingDelete  = `SELECT * FROM reports WHERE status = ? ORDER BY status_updated_at ASC NULLS FIRST LIMIT ?`
	sqlTotalReportsByStatus = `SELECT status, COUNT(*) AS count FROM reports GROUP BY status`
	sqlSelectReportsPage    = `SELECT * FROM reports WHERE tenant_id = ? ORDER BY created_by DESC NULLS LAST,id LIMIT ? OFFSET ?`
	sqlSelectReportsDefault = `SELECT * FROM reports WHERE tenant_id = ? ORDER BY id`

	testTenantID = int32(7)
	testReportID = "4f8f5f0e-0b8c-4c5b-9d3a-1d2f3a4b5c6d"
)

var errDB = errors.New("db down")

// expectExecInTx wraps the exec in the transaction the ORM opens for writes.
func expectExecInTx(mocker dbtesting.Mocker, rowsAffected int64, query string, args ...any) {
	tx := mocker.ExpectTransaction()
	tx.Begin()
	mocker.ExpectExec(query, args...).Returns(rowsAffected)
	tx.Commit()
}

func expectExecErrorInTx(mocker dbtesting.Mocker, query string, args ...any) {
	tx := mocker.ExpectTransaction()
	tx.Begin()
	mocker.ExpectExec(query, args...).ReturnsError(errDB)
	tx.Rollback()
}

// reportRow omits the Metrics association, which sqlmock cannot serialize, and
// carries the jsonb columns as raw bytes, which is what their Scan methods expect.
type reportRow struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt        time.Time
	CreatedBy        string
	Name             string
	Description      string
	TenantId         int32
	Status           string
	StatusUpdatedAt  time.Time
	StatusMessage    string
	Start            time.Time
	End              time.Time
	Labels           []byte
	AdditionalFields []byte
}

type idRow struct {
	ID uuid.UUID
}

func newReportRow(status api.ReportStatus) reportRow {
	return reportRow{
		ID:               uuid.New(),
		TenantId:         testTenantID,
		Name:             "report",
		Status:           string(status),
		Labels:           []byte("[]"),
		AdditionalFields: []byte("{}"),
	}
}

func TestReportsDaoSQLGetReport(t *testing.T) {
	t.Run("returns the report", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		row := newReportRow(api.REPORTSTATUS_PENDING)
		mocker.ExpectQuery(sqlSelectReport, testTenantID, testReportID, 1).Returns([]reportRow{row})
		mocker.ExpectQuery(sqlSelectReportMetrics, row.ID).Returns([]models.Metric{})

		got, err := NewPostgresReportsDao().GetReport(context.Background(), testTenantID, testReportID)

		require.Nil(t, err)
		require.Equal(t, row.ID, got.ID)
		require.Equal(t, string(api.REPORTSTATUS_PENDING), got.Status)
	})

	t.Run("no rows is a 404", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		mocker.ExpectQuery(sqlSelectReport, testTenantID, testReportID, 1).Returns([]reportRow{})

		_, err := NewPostgresReportsDao().GetReport(context.Background(), testTenantID, testReportID)

		require.NotNil(t, err)
		require.Equal(t, http.StatusNotFound, int(err.Code))
		require.Equal(t, "requested resource does not exist", err.Details)
	})

	t.Run("db failure is a 500", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		mocker.ExpectQuery(sqlSelectReport, testTenantID, testReportID, 1).ReturnsError(errDB)

		_, err := NewPostgresReportsDao().GetReport(context.Background(), testTenantID, testReportID)

		require.NotNil(t, err)
		require.Equal(t, http.StatusInternalServerError, int(err.Code))
		require.Equal(t, "Failed to get report", err.Message)
	})
}

func TestReportsDaoSQLGetReportsByIds(t *testing.T) {
	t.Run("empty ids runs no query", func(t *testing.T) {
		dbtesting.SetSQLMock(t)

		got, err := NewPostgresReportsDao().GetReportsByIds(context.Background(), testTenantID, nil)

		require.Nil(t, err)
		require.Empty(t, got)
	})

	t.Run("db failure is a 500", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		mocker.ExpectQuery(sqlSelectReportsByIds, testTenantID, testReportID, testReportID).ReturnsError(errDB)

		_, err := NewPostgresReportsDao().GetReportsByIds(context.Background(), testTenantID, []string{testReportID, testReportID})

		require.NotNil(t, err)
		require.Equal(t, http.StatusInternalServerError, int(err.Code))
	})
}

func TestReportsDaoSQLGetReports(t *testing.T) {
	t.Run("defaults to id order", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		mocker.ExpectQuery(sqlSelectReportsDefault, testTenantID).Returns([]reportRow{})

		_, err := NewPostgresReportsDao().GetReports(context.Background(), testTenantID, nil, "", "", 0, 0, "")

		require.Nil(t, err)
	})

	for _, tc := range []struct {
		column    string
		sortOrder string
		orderBy   string
	}{
		{column: "name", sortOrder: "asc", orderBy: "name ASC NULLS FIRST"},
		{column: "name", sortOrder: "desc", orderBy: "name DESC NULLS LAST"},
		{column: "created_at", sortOrder: "asc", orderBy: "created_at ASC NULLS FIRST"},
		{column: "created_at", sortOrder: "desc", orderBy: "created_at DESC NULLS LAST"},
		{column: "created_by", sortOrder: "asc", orderBy: "created_by ASC NULLS FIRST"},
		{column: "created_by", sortOrder: "desc", orderBy: "created_by DESC NULLS LAST"},
		{column: "status", sortOrder: "asc", orderBy: "status ASC NULLS FIRST"},
		{column: "status", sortOrder: "desc", orderBy: "status DESC NULLS LAST"},
	} {
		t.Run("sorts by "+tc.column+" "+tc.sortOrder+" with id as the tiebreaker", func(t *testing.T) {
			mocker := dbtesting.SetSQLMock(t)
			mocker.ExpectQuery(`SELECT * FROM reports WHERE tenant_id = ? ORDER BY `+tc.orderBy+`,id`, testTenantID).Returns([]reportRow{})

			_, err := NewPostgresReportsDao().GetReports(context.Background(), testTenantID, nil, tc.column, tc.sortOrder, 0, 0, "")

			require.Nil(t, err)
		})
	}

	t.Run("pages with limit and offset", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		mocker.ExpectQuery(sqlSelectReportsPage, testTenantID, 10, 20).Returns([]reportRow{})

		_, err := NewPostgresReportsDao().GetReports(context.Background(), testTenantID, nil, "created_by", "desc", 20, 10, "")

		require.Nil(t, err)
	})

	t.Run("db failure is a 500", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		mocker.ExpectQuery(sqlSelectReportsDefault, testTenantID).ReturnsError(errDB)

		_, err := NewPostgresReportsDao().GetReports(context.Background(), testTenantID, nil, "", "", 0, 0, "")

		require.NotNil(t, err)
		require.Equal(t, http.StatusInternalServerError, int(err.Code))
	})
}

func TestReportsDaoSQLDeleteReport(t *testing.T) {
	t.Run("scopes the delete to tenant and id", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		expectExecInTx(mocker, 1, sqlDeleteReport, testTenantID, testReportID)

		require.Nil(t, NewPostgresReportsDao().DeleteReport(context.Background(), testTenantID, testReportID))
	})

	t.Run("db failure is a 500", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		expectExecErrorInTx(mocker, sqlDeleteReport, testTenantID, testReportID)

		err := NewPostgresReportsDao().DeleteReport(context.Background(), testTenantID, testReportID)

		require.NotNil(t, err)
		require.Equal(t, http.StatusInternalServerError, int(err.Code))
	})
}

func TestReportsDaoSQLUpdateReportStatus(t *testing.T) {
	failed := api.REPORTSTATUS_FAILED

	t.Run("without a previous status does not guard on it", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		expectExecInTx(mocker, 1, sqlUpdateStatus, string(failed), testTenantID, testReportID)

		require.NoError(t, NewPostgresReportsDao().UpdateReportStatus(testTenantID, testReportID, failed, nil))
	})

	t.Run("with a previous status only updates from that status", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		expectExecInTx(mocker, 1, sqlUpdateStatusGuarded, string(api.REPORTSTATUS_PROCESSING), testTenantID, testReportID, string(failed))

		require.NoError(t, NewPostgresReportsDao().UpdateReportStatus(testTenantID, testReportID, api.REPORTSTATUS_PROCESSING, ptr.To(failed)))
	})

	t.Run("a message is written only when given", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		expectExecInTx(mocker, 1, sqlUpdateStatusMessage, string(failed), "timeout error", testTenantID, testReportID)

		require.NoError(t, NewPostgresReportsDao().UpdateReportStatusWithMessage(testTenantID, testReportID, failed, "timeout error", nil))
	})

	t.Run("no rows affected is an error", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		expectExecInTx(mocker, 0, sqlUpdateStatusGuarded, string(api.REPORTSTATUS_PROCESSING), testTenantID, testReportID, string(api.REPORTSTATUS_PENDING))

		err := NewPostgresReportsDao().UpdateReportStatus(testTenantID, testReportID, api.REPORTSTATUS_PROCESSING, ptr.To(api.REPORTSTATUS_PENDING))

		require.ErrorContains(t, err, "No rows affected")
	})

	t.Run("db failure is returned", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		expectExecErrorInTx(mocker, sqlUpdateStatus, string(failed), testTenantID, testReportID)

		require.ErrorIs(t, NewPostgresReportsDao().UpdateReportStatus(testTenantID, testReportID, failed, nil), errDB)
	})
}

func TestReportsDaoSQLUpdatePendingDelete(t *testing.T) {
	ids := []string{testReportID, uuid.NewString()}

	t.Run("empty ids runs no query", func(t *testing.T) {
		dbtesting.SetSQLMock(t)

		updated, err := NewPostgresReportsDao().UpdatePendingDelete(context.Background(), testTenantID, nil)

		require.Nil(t, err)
		require.Empty(t, updated)
	})

	t.Run("returns the ids that were updated, skipping processing and deleting", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		tx := mocker.ExpectTransaction()
		tx.Begin()
		mocker.ExpectQuery(sqlUpdatePendingDelete, string(api.REPORTSTATUS_PENDING_DELETE), dbtesting.AnyTime(), testTenantID, ids[0], ids[1],
			string(api.REPORTSTATUS_PROCESSING), string(api.REPORTSTATUS_DELETING)).
			Returns([]idRow{{ID: uuid.MustParse(ids[0])}})
		tx.Commit()

		updated, err := NewPostgresReportsDao().UpdatePendingDelete(context.Background(), testTenantID, ids)

		require.Nil(t, err)
		require.Equal(t, []string{ids[0]}, updated)
	})

	t.Run("db failure is a 500", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		tx := mocker.ExpectTransaction()
		tx.Begin()
		mocker.ExpectQuery(sqlUpdatePendingDelete, string(api.REPORTSTATUS_PENDING_DELETE), dbtesting.AnyTime(), testTenantID, ids[0], ids[1],
			string(api.REPORTSTATUS_PROCESSING), string(api.REPORTSTATUS_DELETING)).
			ReturnsError(errDB)
		tx.Rollback()

		_, err := NewPostgresReportsDao().UpdatePendingDelete(context.Background(), testTenantID, ids)

		require.NotNil(t, err)
		require.Equal(t, http.StatusInternalServerError, int(err.Code))
	})
}

func TestReportsDaoSQLGetPendingReports(t *testing.T) {
	t.Run("selects the oldest pending reports up to the limit", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		row := newReportRow(api.REPORTSTATUS_PENDING)
		mocker.ExpectQuery(sqlSelectPending, string(api.REPORTSTATUS_PENDING), 10).Returns([]reportRow{row})
		mocker.ExpectQuery(sqlSelectReportMetrics, row.ID).Returns([]models.Metric{})

		got, err := NewPostgresReportsDao().GetPendingReports(context.Background(), 10)

		require.Nil(t, err)
		require.Len(t, got, 1)
		require.Equal(t, row.ID, got[0].ID)
	})

	t.Run("db failure is a 500", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		mocker.ExpectQuery(sqlSelectPending, string(api.REPORTSTATUS_PENDING), 10).ReturnsError(errDB)

		_, err := NewPostgresReportsDao().GetPendingReports(context.Background(), 10)

		require.NotNil(t, err)
		require.Equal(t, http.StatusInternalServerError, int(err.Code))
	})
}

func TestReportsDaoSQLGetPendingDeleteReports(t *testing.T) {
	t.Run("selects the pending-delete reports ordered by when they were marked", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		row := newReportRow(api.REPORTSTATUS_PENDING_DELETE)
		mocker.ExpectQuery(sqlSelectPendingDelete, string(api.REPORTSTATUS_PENDING_DELETE), 10).Returns([]reportRow{row})

		got, err := NewPostgresReportsDao().GetPendingDeleteReports(context.Background(), 10)

		require.Nil(t, err)
		require.Len(t, got, 1)
		require.Equal(t, row.ID, got[0].ID)
	})

	t.Run("db failure is a 500", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		mocker.ExpectQuery(sqlSelectPendingDelete, string(api.REPORTSTATUS_PENDING_DELETE), 10).ReturnsError(errDB)

		_, err := NewPostgresReportsDao().GetPendingDeleteReports(context.Background(), 10)

		require.NotNil(t, err)
		require.Equal(t, http.StatusInternalServerError, int(err.Code))
	})
}

func TestReportsDaoSQLGetTotalReportsByStatus(t *testing.T) {
	type statusRow struct {
		Status string
		Count  int
	}

	t.Run("maps rows to a status count map", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		mocker.ExpectQuery(sqlTotalReportsByStatus).Returns([]statusRow{{Status: "Pending", Count: 3}, {Status: "Ready", Count: 5}})

		got, err := NewPostgresReportsDao().GetTotalReportsByStatus(context.Background())

		require.Nil(t, err)
		require.Equal(t, map[string]int{"Pending": 3, "Ready": 5}, got)
	})

	t.Run("db failure is a 500", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		mocker.ExpectQuery(sqlTotalReportsByStatus).ReturnsError(errDB)

		_, err := NewPostgresReportsDao().GetTotalReportsByStatus(context.Background())

		require.NotNil(t, err)
		require.Equal(t, http.StatusInternalServerError, int(err.Code))
	})
}
