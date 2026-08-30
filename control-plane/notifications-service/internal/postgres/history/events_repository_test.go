package history

import (
	"errors"
	"fmt"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	history "github.com/run-ai/runai/control-plane/notifications-service/internal/history"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/postgres/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	eventsTestSchema = "test"
	eventsTestTable  = `"test"."events"`
)

func newEventsRepositoryTest(t *testing.T) (*EventsRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	gormDB, err := gorm.Open(gormPostgres.New(gormPostgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	return &EventsRepository{db: gormDB, schema: eventsTestSchema}, mock
}

func TestGetEvents_omitsIsReadFilterWhenUnset(t *testing.T) {
	repo, mock := newEventsRepositoryTest(t)
	createdAt := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`SELECT \* FROM "test"\."events" WHERE .*tenant_id = \$1.*user_id = \$2`).
		WithArgs("t1", "u1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "tenant_id", "user_id", "listener_id", "description", "properties", "is_read", "created_at", "updated_at",
		}).AddRow("evt-1", "t1", "u1", "listener-1", "desc", `{}`, true, createdAt, createdAt))

	events, err := repo.GetEvents(history.EventFilter{TenantId: "t1", UserId: "u1"})
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.True(t, events[0].Read)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetEvents_All(t *testing.T) {
	repo, mock := newEventsRepositoryTest(t)
	createdAt := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`SELECT \* FROM "test"\."events"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "tenant_id", "user_id", "listener_id", "description", "properties", "is_read", "created_at", "updated_at",
		}).AddRow("evt-1", "t1", "u1", "listener-1", "desc", `{}`, true, createdAt, createdAt))

	events, err := repo.GetEvents(history.EventFilter{})
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.True(t, events[0].Read)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetEvents_filtersByIsRead(t *testing.T) {
	repo, mock := newEventsRepositoryTest(t)
	createdAt := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	read := false

	mock.ExpectQuery(`SELECT \* FROM "test"\."events" WHERE .*tenant_id = \$1.*user_id = \$2.*is_read = \$3`).
		WithArgs("t1", "u1", false).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "tenant_id", "user_id", "listener_id", "description", "properties", "is_read", "created_at", "updated_at",
		}).AddRow("evt-1", "t1", "u1", "listener-1", "desc", `{}`, false, createdAt, createdAt))

	events, err := repo.GetEvents(history.EventFilter{TenantId: "t1", UserId: "u1", Read: &read})
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.False(t, events[0].Read)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestToServiceEvents(t *testing.T) {
	createdAt := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	events := toServiceEvents([]*models.Event{{
		Id:          "evt-1",
		ListenerId:  "listener-1",
		Properties:  map[string]interface{}{"k": "v"},
		CreatedAt:   createdAt,
		IsRead:      true,
		Description: "desc",
	}})

	require.Len(t, events, 1)
	assert.Equal(t, "evt-1", events[0].Id)
	assert.Equal(t, "listener-1", events[0].SourceListener)
	assert.Equal(t, map[string]interface{}{"k": "v"}, events[0].Properties)
	assert.Equal(t, createdAt, events[0].CreatedAt)
	assert.True(t, events[0].Read)
}

func TestMarkRead_UpdatesIsReadAndUpdatedAt(t *testing.T) {
	repo, mock := newEventsRepositoryTest(t)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "test"\."events" SET "is_read"=\$1,"updated_at"=\$2 WHERE id = \$3`).
		WithArgs(true, sqlmock.AnyArg(), "evt-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.MarkRead("evt-1"))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMarkUnread_UpdatesIsReadFalse(t *testing.T) {
	repo, mock := newEventsRepositoryTest(t)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "test"\."events" SET "is_read"=\$1,"updated_at"=\$2 WHERE id = \$3`).
		WithArgs(false, sqlmock.AnyArg(), "evt-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.MarkUnread("evt-1"))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMarkRead_NotFound(t *testing.T) {
	repo, mock := newEventsRepositoryTest(t)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "test"\."events" SET "is_read"=\$1,"updated_at"=\$2 WHERE id = \$3`).
		WithArgs(true, sqlmock.AnyArg(), "missing").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := repo.MarkRead("missing")
	require.ErrorIs(t, err, history.ErrEventNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteEvent_Success(t *testing.T) {
	repo, mock := newEventsRepositoryTest(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "test"\."events" WHERE .*id`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.DeleteEvent("evt-1"))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteEvent_NotFound(t *testing.T) {
	repo, mock := newEventsRepositoryTest(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "test"\."events" WHERE .*id`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := repo.DeleteEvent("missing")
	require.ErrorIs(t, err, history.ErrEventNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteEvents_Success(t *testing.T) {
	repo, mock := newEventsRepositoryTest(t)
	read := false

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "test"\."events" WHERE .*tenant_id = \$1.*user_id = \$2.*is_read = \$3`).
		WithArgs("t1", "u1", false).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	err := repo.DeleteEvents(history.EventFilter{TenantId: "t1", UserId: "u1", Read: &read})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteEvents_NoRowsDeleted_Succeeds(t *testing.T) {
	repo, mock := newEventsRepositoryTest(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "test"\."events" WHERE .*tenant_id = \$1.*user_id = \$2`).
		WithArgs("t1", "u1").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.DeleteEvents(history.EventFilter{TenantId: "t1", UserId: "u1"})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCleanup_NoFilters(t *testing.T) {
	repo, mock := newEventsRepositoryTest(t)

	require.NoError(t, repo.Cleanup(history.CleanupFilter{}))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCleanup_MaxAge(t *testing.T) {
	repo, mock := newEventsRepositoryTest(t)
	maxAge := 24 * time.Hour

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "test"\."events" WHERE created_at < \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	require.NoError(t, repo.Cleanup(history.CleanupFilter{MaxAge: &maxAge}))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCleanup_MaxEventsPerUser(t *testing.T) {
	repo, mock := newEventsRepositoryTest(t)
	maxPerUser := 100

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "test"\."events" WHERE id IN`).
		WithArgs(maxPerUser).
		WillReturnResult(sqlmock.NewResult(0, 5))
	mock.ExpectCommit()

	require.NoError(t, repo.Cleanup(history.CleanupFilter{MaxEventsPerUser: &maxPerUser}))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCleanup_ErrorRollsBack(t *testing.T) {
	repo, mock := newEventsRepositoryTest(t)
	maxAge := 12 * time.Hour
	maxPerUser := 10

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "test"\."events" WHERE created_at < \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM "test"\."events" WHERE id IN`).
		WithArgs(maxPerUser).
		WillReturnError(errors.New("cleanup failed"))
	mock.ExpectRollback()

	err := repo.Cleanup(history.CleanupFilter{MaxAge: &maxAge, MaxEventsPerUser: &maxPerUser})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cleanup failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIsHealthy(t *testing.T) {
	repo, mock := newEventsRepositoryTest(t)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(eventsTestSchema, models.EventsTable).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	ok, err := repo.IsHealthy()
	require.NoError(t, err)
	assert.True(t, ok)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIsHealthy_TableMissing(t *testing.T) {
	repo, mock := newEventsRepositoryTest(t)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(eventsTestSchema, models.EventsTable).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	ok, err := repo.IsHealthy()
	require.NoError(t, err)
	assert.False(t, ok)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIsHealthy_Error(t *testing.T) {
	repo, mock := newEventsRepositoryTest(t)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(eventsTestSchema, models.EventsTable).
		WillReturnError(fmt.Errorf("db down"))

	ok, err := repo.IsHealthy()
	require.Error(t, err)
	assert.False(t, ok)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWhereIdInIdsToBeDeletedByLimitPerUser(t *testing.T) {
	clause := whereIdInIdsToBeDeletedByLimitPerUser(`"test"."events"`)
	assert.Contains(t, clause, `FROM "test"."events"`)
	assert.Contains(t, clause, "PARTITION BY tenant_id, user_id")
	assert.Contains(t, clause, "counter > ?")
}
