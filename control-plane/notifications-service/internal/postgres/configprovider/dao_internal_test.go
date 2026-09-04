/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package configprovider

import (
	"fmt"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/advanced/http_listener/encryption"
	edmock "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/advanced/http_listener/encryption/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres/configprovider/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres/models"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var dao = Dao{}

func NewTestDao(db *gorm.DB, schema string, sfh encryption.SecretFieldHandlerer) Dao {
	return Dao{
		db:                db,
		listenerDao:       newListenerDao(db, schema),
		tenantDao:         newTenantDao(db, schema),
		userDao:           newUserDao(db, schema),
		sinksDao:          newSinksDao(db, schema),
		secretDataHandler: sfh,
	}
}

func TestInit_ValidationSucceed(t *testing.T) {
	vd := mocks.NewValidationDao(t)
	ed := edmock.NewEncrypterDecrypterer(t)

	ed.On("Encrypt", ValidationText).Return("ValidationTextEncrypted", nil)
	vd.On("InsertValidationIfNotExist", "ValidationTextEncrypted").Return(&models.Validation{
		Data: "ValidationTextEncrypted",
	}, nil)
	ed.On("Decrypt", "ValidationTextEncrypted").Return(ValidationText, nil)

	require.Nil(t, dao.init(vd, ed))
}

func TestInit_FailedToGetRecord(t *testing.T) {
	vd := mocks.NewValidationDao(t)
	ed := edmock.NewEncrypterDecrypterer(t)

	ed.On("Encrypt", ValidationText).Return("ValidationTextEncrypted", nil)
	vd.On("InsertValidationIfNotExist", "ValidationTextEncrypted").Return(nil, fmt.Errorf("failed to get"))

	require.NotNil(t, dao.init(vd, ed))
}

func TestInit_FailedToDecrypt(t *testing.T) {
	vd := mocks.NewValidationDao(t)
	ed := edmock.NewEncrypterDecrypterer(t)

	ed.On("Encrypt", ValidationText).Return("ValidationTextEncrypted", nil)
	vd.On("InsertValidationIfNotExist", "ValidationTextEncrypted").Return(&models.Validation{
		Data: "ValidationTextEncrypted",
	}, nil)
	ed.On("Decrypt", "ValidationTextEncrypted").Return(ValidationText, fmt.Errorf("failed to decrypt"))

	require.NotNil(t, dao.init(vd, ed))
}

func TestApplyListenerId(t *testing.T) {
	routes := []*models.RouteGlobalConfig{
		{ListenerId: "http-*"},
		{ListenerId: "http*"},
		{ListenerId: "runai*"},
		{ListenerId: "runai-http"},
		{ListenerId: "*-runai"},
		{ListenerId: "http-runai"},
		{ListenerId: "http.*"},
		{ListenerId: "runai"},
		{ListenerId: ".*runai"},
	}

	final := filterRoutesByListenerId("http-runai", routes)
	require.Equal(t, []*models.RouteGlobalConfig{
		{ListenerId: "http-*"},
		{ListenerId: "http*"},
		{ListenerId: "*-runai"},
		{ListenerId: "http-runai"},
	}, final)
}

func TestSoftCreateUniqueIndex(t *testing.T) {
	table := migrateTestSinkTable()
	db, mock := newMigrateTestDB(t)

	expectSinkIndexAbsent(mock)
	mock.ExpectExec(`CREATE UNIQUE INDEX IF NOT EXISTS "idx_sink_config_unique_sink_id" ON .+ WHERE deleted_at IS NULL`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	softCreateUniqueIndex(db, table, "sink_id", sinkIdUniqueIndex, &models.SinkConfig{})

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSoftCreateUniqueIndex_createFailsWhenDuplicatesExist(t *testing.T) {
	table := migrateTestSinkTable()
	db, mock := newMigrateTestDB(t)

	expectSinkIndexAbsent(mock)
	mock.ExpectExec(`CREATE UNIQUE INDEX IF NOT EXISTS "idx_sink_config_unique_sink_id" ON .+ WHERE deleted_at IS NULL`).
		WillReturnError(fmt.Errorf(`ERROR: could not create unique index "idx_sink_config_unique_sink_id" (SQLSTATE 23505)`))

	require.NotPanics(t, func() {
		softCreateUniqueIndex(db, table, "sink_id", sinkIdUniqueIndex, &models.SinkConfig{})
	})
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSoftCreateUniqueIndex_IndexExists(t *testing.T) {
	table := migrateTestSinkTable()
	db, mock := newMigrateTestDB(t)

	expectSinkIndexPresent(mock)
	softCreateUniqueIndex(db, table, "sink_id", sinkIdUniqueIndex, &models.SinkConfig{})

	expectSinkIndexPresent(mock)
	require.NotPanics(t, func() {
		softCreateUniqueIndex(db, table, "sink_id", sinkIdUniqueIndex, &models.SinkConfig{})
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

func migrateTestSinkTable() string {
	return fmt.Sprintf("%s.%s", pq.QuoteIdentifier("migrate_test"), pq.QuoteIdentifier(models.SinkTable))
}

func newMigrateTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	db, err := gorm.Open(gormPostgres.New(gormPostgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	require.NoError(t, err)
	return db, mock
}

func expectSinkIndexAbsent(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT count\(\*\) FROM pg_indexes`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
}

func expectSinkIndexPresent(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT count\(\*\) FROM pg_indexes`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
}
