package postgres

import (
	"fmt"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/aws/smithy-go/ptr"
	"github.com/golang-migrate/migrate/v4"
	dbconfig "github.com/run-ai/runai/runai-common-packages/runaidb/config"
	"github.com/run-ai/runai/runai-common-packages/runaidb/internal/logger"
	logmocks "github.com/run-ai/runai/runai-common-packages/runaidb/internal/logger/mocks"
	"github.com/run-ai/runai/runai-common-packages/runaidb/internal/postgres/mocks"
	ormmocks "github.com/run-ai/runai/runai-common-packages/runaidb/orm/mocks"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func defaultMigrationConfig(direction dbconfig.MigrationDirection) *dbconfig.Config {
	return &dbconfig.Config{
		Connection: dbconfig.ConnectionConfig{Schema: "public", Name: "mydb"},
		Migration:  dbconfig.MigrationConfig{Direction: direction, Files: "file:///tmp/migrations"},
	}
}

const (
	createSchemaQuery  = `CREATE SCHEMA IF NOT EXISTS "public";`
	currentDatabaseQry = "SELECT CURRENT_DATABASE()"
	publicFailures     = `"public"."migration_failures"`
	publicVersions     = `"public"."schema_migrations"`
)

var createFailuresTableQry = fmt.Sprintf(createFailuresTableQuery, publicFailures)

// expectLockAcquired plays out schema creation and a granted lock, then fails the
// migration driver setup so the test lands directly on the release path.
func expectLockAcquired(mock sqlmock.Sqlmock) {
	mock.ExpectExec(createSchemaQuery).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(tryLockQuery).WithArgs("public").WillReturnRows(boolRows(true, true))
	mock.ExpectExec(createFailuresTableQry).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(currentDatabaseQry).WillReturnError(fmt.Errorf("driver setup failed"))
}

func TestMigrate_SQLDBError(t *testing.T) {
	db := ormmocks.NewDB(t)
	db.On("Actual").Return(&gorm.DB{Config: &gorm.Config{ConnPool: nil}})

	m := &Migrator{db: db}
	err := m.Migrate(defaultMigrationConfig(dbconfig.MigrationDirectionUp))
	require.ErrorContains(t, err, "failed to get underlying sql.DB")
}

func TestMigrate_EnsureSchemaError(t *testing.T) {
	conn, mock := newMockConn(t)
	mock.ExpectExec(createSchemaQuery).WillReturnError(fmt.Errorf("schema error"))

	err := (&Migrator{}).migrateOnConn(conn, conn, defaultMigrationConfig(dbconfig.MigrationDirectionUp))
	require.ErrorContains(t, err, `failed to ensure schema "public" exists`)
}

func TestMigrate_AcquireError(t *testing.T) {
	conn, mock := newMockConn(t)
	mock.ExpectExec(createSchemaQuery).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(tryLockQuery).WithArgs("public").WillReturnError(fmt.Errorf("db error"))

	err := (&Migrator{}).migrateOnConn(conn, conn, defaultMigrationConfig(dbconfig.MigrationDirectionUp))
	require.ErrorContains(t, err, `failed to acquire migration lock for schema "public"`)
	require.ErrorContains(t, err, "db error")
}

func TestMigrate_LockAlreadyHeld(t *testing.T) {
	conn, mock := newMockConn(t)
	mock.ExpectExec(createSchemaQuery).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(tryLockQuery).WithArgs("public").WillReturnRows(boolRows(true, false))

	err := (&Migrator{}).migrateOnConn(conn, conn, defaultMigrationConfig(dbconfig.MigrationDirectionUp))
	require.EqualError(t, err, "migration lock for schema \"public\" is already held")
}

func TestMigrate_ReleaseError_IsLogged(t *testing.T) {
	log := logmocks.NewLogger(t)
	log.On("Errorf", "failed to release migration lock for schema %q: %v", "public", fmt.Errorf("release failed")).Return()
	logger.SetLogger(log)

	conn, sqlMock := newMockConn(t)
	expectLockAcquired(sqlMock)
	sqlMock.ExpectQuery(unlockQuery).WithArgs("public").WillReturnError(fmt.Errorf("release failed"))

	err := (&Migrator{}).migrateOnConn(conn, conn, defaultMigrationConfig(dbconfig.MigrationDirectionUp))
	require.ErrorContains(t, err, "failed to create postgres migration driver")
}

// A false return from pg_advisory_unlock means the lock was taken on a different
// session and is still held - the exact failure that stalls other replicas.
func TestMigrate_LockNotReleased_IsLogged(t *testing.T) {
	log := logmocks.NewLogger(t)
	log.On("Errorf", "migration lock for schema %q was not held by the releasing session", "public").Return()
	logger.SetLogger(log)

	conn, sqlMock := newMockConn(t)
	expectLockAcquired(sqlMock)
	sqlMock.ExpectQuery(unlockQuery).WithArgs("public").WillReturnRows(boolRows(true, false))

	err := (&Migrator{}).migrateOnConn(conn, conn, defaultMigrationConfig(dbconfig.MigrationDirectionUp))
	require.ErrorContains(t, err, "failed to create postgres migration driver")
}

func TestExecuteMigration_InvalidVersion(t *testing.T) {
	m := &Migrator{}
	conf := defaultMigrationConfig(dbconfig.MigrationDirectionUp)
	conf.Migration.TargetVersion = ptr.String("bad")

	err := m.executeMigration(nil, nil, conf)
	require.ErrorContains(t, err, `invalid target version "bad"`)
}

func TestExecuteMigration_UnknownDirection(t *testing.T) {
	m := &Migrator{}
	conf := defaultMigrationConfig("unknown")

	err := m.executeMigration(nil, nil, conf)
	require.ErrorContains(t, err, `unknown migration direction "unknown"`)
}

func TestExecuteMigration_ErrNoChange(t *testing.T) {
	mig := mocks.NewMigrationRunner(t)
	mig.On("Up").Return(migrate.ErrNoChange)

	m := &Migrator{}
	err := m.executeMigration(mig, nil, defaultMigrationConfig(dbconfig.MigrationDirectionUp))
	require.NoError(t, err)
}

func TestExecuteMigration_StrategyError(t *testing.T) {
	mig := mocks.NewMigrationRunner(t)
	mig.On("Up").Return(fmt.Errorf("up failed"))

	conn, mock := newMockConn(t)
	mock.ExpectQuery(fmt.Sprintf(selectVersionQuery, publicVersions)).
		WillReturnRows(sqlmock.NewRows([]string{"version", "dirty"}).AddRow(7, true))
	mock.ExpectExec(fmt.Sprintf(recordFailureQuery, publicFailures)).
		WithArgs(int64(7), "up failed").WillReturnResult(sqlmock.NewResult(0, 1))

	m := &Migrator{}
	err := m.executeMigration(mig, newMigrationFailuresStore(conn, "public"), defaultMigrationConfig(dbconfig.MigrationDirectionUp))
	require.ErrorContains(t, err, "migration failed: up failed")
}

func TestExecuteMigration_Success(t *testing.T) {
	mig := mocks.NewMigrationRunner(t)
	mig.On("Up").Return(nil)

	m := &Migrator{}
	err := m.executeMigration(mig, nil, defaultMigrationConfig(dbconfig.MigrationDirectionUp))
	require.NoError(t, err)
}

func TestEvaluateVersion_Nil(t *testing.T) {
	v, err := evaluateVersion(nil)
	require.NoError(t, err)
	require.Nil(t, v)
}

func TestEvaluateVersion_Valid(t *testing.T) {
	v, err := evaluateVersion(ptr.String("42"))
	require.NoError(t, err)
	require.NotNil(t, v)
	require.Equal(t, uint(42), *v)
}

func TestEvaluateVersion_Invalid(t *testing.T) {
	_, err := evaluateVersion(ptr.String("not-a-number"))
	require.ErrorContains(t, err, `invalid target version "not-a-number"`)
}

func TestGetMigrationFilesPath_Up(t *testing.T) {
	path := getMigrationFilesPath(defaultMigrationConfig(dbconfig.MigrationDirectionUp))
	require.Equal(t, "file:///tmp/migrations/up", path)
}

func TestGetMigrationFilesPath_Down(t *testing.T) {
	path := getMigrationFilesPath(defaultMigrationConfig(dbconfig.MigrationDirectionDown))
	require.Equal(t, "file:///tmp/migrations/down", path)
}
