package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

const (
	tryLockQuery = "SELECT pg_try_advisory_lock(to_regnamespace($1)::integer)"
	unlockQuery  = "SELECT pg_advisory_unlock(to_regnamespace($1)::integer)"
)

func newMockConn(t *testing.T) (*sql.Conn, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	conn, err := db.Conn(context.Background())
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, mock.ExpectationsWereMet())
		_ = conn.Close()
		_ = db.Close()
	})

	return conn, mock
}

func boolRows(valid, val bool) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"lock"})
	if !valid {
		return rows.AddRow(nil)
	}
	return rows.AddRow(val)
}

func TestNewLocker(t *testing.T) {
	conn, _ := newMockConn(t)
	require.NotNil(t, NewLocker(conn))
}

func TestAcquire_Success(t *testing.T) {
	conn, mock := newMockConn(t)
	mock.ExpectQuery(tryLockQuery).WithArgs("some-schema").WillReturnRows(boolRows(true, true))

	acquired, err := (&Locker{conn: conn}).Acquire("some-schema")

	require.NoError(t, err)
	require.True(t, acquired)
}

func TestAcquire_SuccessAfterFailed(t *testing.T) {
	conn, mock := newMockConn(t)
	mock.ExpectQuery(tryLockQuery).WithArgs("some-schema").WillReturnRows(boolRows(true, false))
	mock.ExpectQuery(tryLockQuery).WithArgs("some-schema").WillReturnRows(boolRows(true, true))

	acquired, err := (&Locker{conn: conn, retries: 3}).Acquire("some-schema")

	require.NoError(t, err)
	require.True(t, acquired)
}

func TestAcquire_NotAcquired(t *testing.T) {
	conn, mock := newMockConn(t)
	mock.ExpectQuery(tryLockQuery).WithArgs("some-schema").WillReturnRows(boolRows(true, false))

	acquired, err := (&Locker{conn: conn}).Acquire("some-schema")

	require.NoError(t, err)
	require.False(t, acquired)
}

func TestAcquire_NullResult(t *testing.T) {
	conn, mock := newMockConn(t)
	mock.ExpectQuery(tryLockQuery).WithArgs("some-schema").WillReturnRows(boolRows(false, false))

	acquired, err := (&Locker{conn: conn}).Acquire("some-schema")

	require.NoError(t, err)
	require.False(t, acquired)
}

func TestAcquire_Failed(t *testing.T) {
	conn, mock := newMockConn(t)
	mock.ExpectQuery(tryLockQuery).WithArgs("some-schema").WillReturnError(fmt.Errorf("failed to acquire lock"))

	acquired, err := (&Locker{conn: conn}).Acquire("some-schema")

	require.ErrorContains(t, err, "failed to acquire lock")
	require.False(t, acquired)
}

// The lock is session scoped, so retries and the release have to stay on the
// connection the locker was handed rather than fanning out across the pool.
func TestAcquireAndRelease_StayOnSameConnection(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	db.SetMaxOpenConns(10)

	conn, err := db.Conn(context.Background())
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	mock.ExpectQuery(tryLockQuery).WithArgs("some-schema").WillReturnRows(boolRows(true, false))
	mock.ExpectQuery(tryLockQuery).WithArgs("some-schema").WillReturnRows(boolRows(true, true))
	mock.ExpectQuery(unlockQuery).WithArgs("some-schema").WillReturnRows(boolRows(true, true))

	locker := &Locker{conn: conn, retries: 3}
	acquired, err := locker.Acquire("some-schema")
	require.NoError(t, err)
	require.True(t, acquired)

	released, err := locker.Release("some-schema")
	require.NoError(t, err)
	require.True(t, released)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRelease_Success(t *testing.T) {
	conn, mock := newMockConn(t)
	mock.ExpectQuery(unlockQuery).WithArgs("some-schema").WillReturnRows(boolRows(true, true))

	released, err := (&Locker{conn: conn}).Release("some-schema")

	require.NoError(t, err)
	require.True(t, released)
}

func TestRelease_NotReleased(t *testing.T) {
	conn, mock := newMockConn(t)
	mock.ExpectQuery(unlockQuery).WithArgs("some-schema").WillReturnRows(boolRows(true, false))

	released, err := (&Locker{conn: conn}).Release("some-schema")

	require.NoError(t, err)
	require.False(t, released)
}

func TestRelease_Failed(t *testing.T) {
	conn, mock := newMockConn(t)
	mock.ExpectQuery(unlockQuery).WithArgs("some-schema").WillReturnError(fmt.Errorf("failed to release lock"))

	released, err := (&Locker{conn: conn}).Release("some-schema")

	require.Error(t, err)
	require.False(t, released)
}
