/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/internal/logger"
	logmocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/internal/logger/mocks"
	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/require"
)

func TestFailureStore_Migrate(t *testing.T) {
	conn, mock := newMockConn(t)
	mock.ExpectExec(createFailuresTableQry).WillReturnResult(sqlmock.NewResult(0, 0))

	require.NoError(t, newMigrationFailuresStore(conn, "public").Migrate(context.Background()))
}

func TestFailureStore_MigrateError(t *testing.T) {
	conn, mock := newMockConn(t)
	mock.ExpectExec(createFailuresTableQry).WillReturnError(fmt.Errorf("no permission"))

	err := newMigrationFailuresStore(conn, "public").Migrate(context.Background())
	require.ErrorContains(t, err, "no permission")
}

func TestFailureStore_RecordsFailureAgainstDirtyVersion(t *testing.T) {
	conn, mock := newMockConn(t)
	mock.ExpectQuery(fmt.Sprintf(selectVersionQuery, publicVersions)).
		WillReturnRows(sqlmock.NewRows([]string{"version", "dirty"}).AddRow(42, true))
	mock.ExpectExec(fmt.Sprintf(recordFailureQuery, publicFailures)).
		WithArgs(int64(42), "syntax error at or near \"CREAT\"").
		WillReturnResult(sqlmock.NewResult(0, 1))

	cause := fmt.Errorf("syntax error at or near %q", "CREAT")
	err := newMigrationFailuresStore(conn, "public").RecordFailure(context.Background(), cause)
	require.Equal(t, cause, err)
}

func TestFailureStore_RecordFailureIsBestEffort(t *testing.T) {
	log := logmocks.NewLogger(t)
	log.On("Errorf", "failed to store the failure of migration %d: %v", int64(42), fmt.Errorf("write failed")).Return()
	logger.SetLogger(log)

	conn, mock := newMockConn(t)
	mock.ExpectQuery(fmt.Sprintf(selectVersionQuery, publicVersions)).
		WillReturnRows(sqlmock.NewRows([]string{"version", "dirty"}).AddRow(42, true))
	mock.ExpectExec(fmt.Sprintf(recordFailureQuery, publicFailures)).
		WillReturnError(fmt.Errorf("write failed"))

	cause := fmt.Errorf("boom")
	err := newMigrationFailuresStore(conn, "public").RecordFailure(context.Background(), cause)
	require.Equal(t, cause, err)
}

func TestFailureStore_UnknownVersionIsBestEffort(t *testing.T) {
	log := logmocks.NewLogger(t)
	log.On("Errorf", "failed to resolve the failing migration version: %v", fmt.Errorf("read failed")).Return()
	logger.SetLogger(log)

	conn, mock := newMockConn(t)
	mock.ExpectQuery(fmt.Sprintf(selectVersionQuery, publicVersions)).
		WillReturnError(fmt.Errorf("read failed"))

	cause := fmt.Errorf("boom")
	err := newMigrationFailuresStore(conn, "public").RecordFailure(context.Background(), cause)
	require.Equal(t, cause, err)
}

func TestFailureStore_DirtyErrorCarriesStoredReason(t *testing.T) {
	failedAt := time.Date(2026, 8, 3, 14, 5, 0, 0, time.UTC)

	conn, mock := newMockConn(t)
	mock.ExpectQuery(fmt.Sprintf(selectFailureQuery, publicFailures)).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"failed_at", "details"}).
			AddRow(failedAt, "column \"foo\" already exists"))

	err := newMigrationFailuresStore(conn, "public").RecordFailure(context.Background(), migrate.ErrDirty{Version: 42})
	require.ErrorContains(t, err, "Dirty database version 42")
	require.ErrorContains(t, err, "migration 42 last failed at 2026-08-03T14:05:00Z")
	require.ErrorContains(t, err, `column "foo" already exists`)
}

func TestFailureStore_DirtyErrorWithoutStoredReason(t *testing.T) {
	conn, mock := newMockConn(t)
	mock.ExpectQuery(fmt.Sprintf(selectFailureQuery, publicFailures)).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"failed_at", "details"}))

	cause := migrate.ErrDirty{Version: 42}
	err := newMigrationFailuresStore(conn, "public").RecordFailure(context.Background(), cause)
	require.Equal(t, cause, err)
}

func TestFailureStore_DirtyErrorLookupFailureIsLogged(t *testing.T) {
	log := logmocks.NewLogger(t)
	log.On("Errorf", "failed to read the stored failure of migration %d: %v", int64(42), fmt.Errorf("read failed")).Return()
	logger.SetLogger(log)

	conn, mock := newMockConn(t)
	mock.ExpectQuery(fmt.Sprintf(selectFailureQuery, publicFailures)).
		WithArgs(int64(42)).WillReturnError(fmt.Errorf("read failed"))

	cause := migrate.ErrDirty{Version: 42}
	err := newMigrationFailuresStore(conn, "public").RecordFailure(context.Background(), cause)
	require.Equal(t, cause, err)
}

type execCounter struct {
	dbExecutor
	execs int
}

func (e *execCounter) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	e.execs++
	return e.dbExecutor.ExecContext(ctx, query, args...)
}

func TestFailureStore_DirtyErrorDoesNotOverwriteStoredReason(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery(fmt.Sprintf(selectVersionQuery, publicVersions)).
		WillReturnRows(sqlmock.NewRows([]string{"version", "dirty"}).AddRow(42, true))
	mock.ExpectQuery(fmt.Sprintf(selectFailureQuery, publicFailures)).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"failed_at", "details"}).
			AddRow(time.Date(2026, 8, 3, 14, 5, 0, 0, time.UTC), "the original reason"))
	mock.ExpectExec(fmt.Sprintf(recordFailureQuery, publicFailures)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	counter := &execCounter{dbExecutor: db}
	annotated := newMigrationFailuresStore(counter, "public").
		RecordFailure(context.Background(), migrate.ErrDirty{Version: 42})

	require.ErrorContains(t, annotated, "the original reason")
	require.Zero(t, counter.execs)
}

func TestFailureStore_SkipsRecordingWhenVersionIsClean(t *testing.T) {
	conn, mock := newMockConn(t)
	mock.ExpectQuery(fmt.Sprintf(selectVersionQuery, publicVersions)).
		WillReturnRows(sqlmock.NewRows([]string{"version", "dirty"}).AddRow(3, false))

	counter := &execCounter{dbExecutor: conn}
	cause := fmt.Errorf("failed to read migration files")
	err := newMigrationFailuresStore(counter, "public").RecordFailure(context.Background(), cause)

	require.Equal(t, cause, err)
	require.Zero(t, counter.execs)
}

func TestFailureStore_SanitizesSchemaName(t *testing.T) {
	store := newMigrationFailuresStore(nil, `we"ird`)
	require.Equal(t, `"we""ird"."migration_failures"`, store.failuresTableName)
	require.Equal(t, `"we""ird"."schema_migrations"`, store.migrationsTableName)
	store = newMigrationFailuresStore(nil, `schema_name`)
	require.Equal(t, `"schema_name"."migration_failures"`, store.failuresTableName)
	require.Equal(t, `"schema_name"."schema_migrations"`, store.migrationsTableName)
}
