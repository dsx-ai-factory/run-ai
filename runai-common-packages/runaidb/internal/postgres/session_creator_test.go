/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package postgres_test

import (
	"errors"
	"fmt"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/internal/logger"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/internal/postgres"
	pgerrors "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/internal/postgres/errors"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/stretchr/testify/require"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	testUser     = "testuser"
	testPassword = "testpass"
	testSchema   = "testschema"

	tryLockQuery = "SELECT pg_try_advisory_lock(to_regnamespace($1)::integer)"
	unlockQuery  = "SELECT pg_advisory_unlock(to_regnamespace($1)::integer)"
)

var (
	sqlCreateUser     = fmt.Sprintf(`CREATE USER "testuser" WITH PASSWORD '%s';`, testPassword)
	sqlAlterUser      = fmt.Sprintf(`ALTER USER "testuser" WITH LOGIN PASSWORD '%s';`, testPassword)
	sqlGrantUsage     = `GRANT USAGE ON SCHEMA "testschema" TO "testuser";`
	sqlGrantTables    = `GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA "testschema" TO "testuser";`
	sqlGrantSequences = `GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA "testschema" TO "testuser";`
	sqlGrantFunctions = `GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA "testschema" TO "testuser";`
)

func newMockDB(t *testing.T) (orm.DB, sqlmock.Sqlmock) {
	t.Helper()

	// other tests in this binary install a strict logger mock globally
	logger.SetLogger(logger.NewDummyLogger())

	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	gormDB, err := gorm.Open(gormpostgres.New(gormpostgres.Config{Conn: sqlDB}), &gorm.Config{PrepareStmt: false})
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, mock.ExpectationsWereMet())
		_ = sqlDB.Close()
	})

	return orm.NewDB(gormDB, nil), mock
}

func lockRow(held bool) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"lock"}).AddRow(held)
}

func expectLockAcquired(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(tryLockQuery).WithArgs(testSchema).WillReturnRows(lockRow(true))
}

func expectLockReleased(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(unlockQuery).WithArgs(testSchema).WillReturnRows(lockRow(true))
}

func expectGrants(mock sqlmock.Sqlmock) {
	mock.ExpectExec(sqlGrantUsage).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(sqlGrantTables).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(sqlGrantSequences).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(sqlGrantFunctions).WillReturnResult(sqlmock.NewResult(0, 0))
}

func userAlreadyExistsErr() error {
	return &pgconn.PgError{Code: pgerrors.ErrCodeDuplicateObject}
}

func TestCreate(t *testing.T) {
	db, mock := newMockDB(t)
	expectLockAcquired(mock)
	mock.ExpectExec(sqlCreateUser).WillReturnResult(sqlmock.NewResult(0, 0))
	expectGrants(mock)
	expectLockReleased(mock)

	err := postgres.NewRWSessionCreator(db).Create(testUser, testPassword, testSchema)
	require.NoError(t, err)
}

// Service-user DDL must never run unserialized: it is either under the lock or
// it does not run at all.
func TestCreate_LockHeldElsewhere_SkipsDDL(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectQuery(tryLockQuery).WithArgs(testSchema).WillReturnRows(lockRow(false))

	err := postgres.NewRWSessionCreator(db).Create(testUser, testPassword, testSchema)
	require.ErrorContains(t, err, `service user lock for schema "testschema" is already held`)
}

func TestCreate_LockQueryFails_SkipsDDL(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectQuery(tryLockQuery).WithArgs(testSchema).WillReturnError(errors.New("lock query failed"))

	err := postgres.NewRWSessionCreator(db).Create(testUser, testPassword, testSchema)
	require.ErrorContains(t, err, "failed to acquire service user lock")
}

func TestCreate_UserAlreadyExists_PasswordUpdated(t *testing.T) {
	db, mock := newMockDB(t)
	expectLockAcquired(mock)
	mock.ExpectExec(sqlCreateUser).WillReturnError(userAlreadyExistsErr())
	mock.ExpectExec(sqlAlterUser).WillReturnResult(sqlmock.NewResult(0, 0))
	expectGrants(mock)
	expectLockReleased(mock)

	err := postgres.NewRWSessionCreator(db).Create(testUser, testPassword, testSchema)
	require.NoError(t, err)
}

func TestCreate_CreateUserFails(t *testing.T) {
	db, mock := newMockDB(t)
	expectLockAcquired(mock)
	mock.ExpectExec(sqlCreateUser).WillReturnError(errors.New("no permission"))
	expectLockReleased(mock)

	err := postgres.NewRWSessionCreator(db).Create(testUser, testPassword, testSchema)
	require.EqualError(t, err, "no permission")
}

func TestCreate_AlterUserFails(t *testing.T) {
	db, mock := newMockDB(t)
	expectLockAcquired(mock)
	mock.ExpectExec(sqlCreateUser).WillReturnError(userAlreadyExistsErr())
	mock.ExpectExec(sqlAlterUser).WillReturnError(errors.New("alter failed"))
	expectLockReleased(mock)

	err := postgres.NewRWSessionCreator(db).Create(testUser, testPassword, testSchema)
	require.EqualError(t, err, "alter failed")
}

func TestCreate_GrantUsageFails(t *testing.T) {
	db, mock := newMockDB(t)
	expectLockAcquired(mock)
	mock.ExpectExec(sqlCreateUser).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(sqlGrantUsage).WillReturnError(errors.New("grant usage failed"))
	expectLockReleased(mock)

	err := postgres.NewRWSessionCreator(db).Create(testUser, testPassword, testSchema)
	require.EqualError(t, err, "grant usage failed")
}

func TestCreate_GrantTablesFails(t *testing.T) {
	db, mock := newMockDB(t)
	expectLockAcquired(mock)
	mock.ExpectExec(sqlCreateUser).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(sqlGrantUsage).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(sqlGrantTables).WillReturnError(errors.New("grant tables failed"))
	expectLockReleased(mock)

	err := postgres.NewRWSessionCreator(db).Create(testUser, testPassword, testSchema)
	require.EqualError(t, err, "grant tables failed")
}

func TestCreate_GrantSequencesFails(t *testing.T) {
	db, mock := newMockDB(t)
	expectLockAcquired(mock)
	mock.ExpectExec(sqlCreateUser).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(sqlGrantUsage).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(sqlGrantTables).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(sqlGrantSequences).WillReturnError(errors.New("grant sequences failed"))
	expectLockReleased(mock)

	err := postgres.NewRWSessionCreator(db).Create(testUser, testPassword, testSchema)
	require.EqualError(t, err, "grant sequences failed")
}

func TestCreate_GrantFunctionsFails(t *testing.T) {
	db, mock := newMockDB(t)
	expectLockAcquired(mock)
	mock.ExpectExec(sqlCreateUser).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(sqlGrantUsage).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(sqlGrantTables).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(sqlGrantSequences).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(sqlGrantFunctions).WillReturnError(errors.New("grant functions failed"))
	expectLockReleased(mock)

	err := postgres.NewRWSessionCreator(db).Create(testUser, testPassword, testSchema)
	require.EqualError(t, err, "grant functions failed")
}

func TestCreate_PasswordWithSingleQuote(t *testing.T) {
	db, mock := newMockDB(t)
	expectLockAcquired(mock)
	mock.ExpectExec(`CREATE USER "testuser" WITH PASSWORD 'it''s a secret';`).WillReturnResult(sqlmock.NewResult(0, 0))
	expectGrants(mock)
	expectLockReleased(mock)

	err := postgres.NewRWSessionCreator(db).Create(testUser, "it's a secret", testSchema)
	require.NoError(t, err)
}
