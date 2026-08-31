/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package configprovider_test

import (
	"fmt"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/advanced/http_listener/encryption"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/basic/route"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/listen"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/postgres/configprovider"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/sinks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xyproto/jpath"
	"gorm.io/datatypes"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	decrypt            = encryption.NewEncrypterDecrypter([]byte("0123456789abcdefghwertol"))
	secretFieldHandler = encryption.NewSecretFieldHandler(decrypt, route.BasicRouteConfiguration{}, listen.ListeningConfig{}, sinks.EmailSinkConfig{}, sinks.SlackSinkConfig{})

	tenantId, userId, listenerId, routeId, sinkId, pwd = "tenant", "user", "listener-http", "route-id", "sink-email", "abc!23"
	encrypted, _                                       = decrypt.Encrypt(pwd)
	configWithEncryptedPassword                        = datatypes.JSON(
		fmt.Sprintf(`{"listener_id":"listener1","notify_ids":{"email":{"from":"tenant@mail.com","password":"%s","subject":"some subject","user":"user@mail.com"}}}`, encrypted),
	)
)

const testSchema = "test"

func newTestGormDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })

	gormDB, err := gorm.Open(gormPostgres.New(gormPostgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	require.NoError(t, err)
	return gormDB, mock
}

func newTestDao(gormDB *gorm.DB) configprovider.Dao {
	return configprovider.NewTestDao(gormDB, testSchema, secretFieldHandler)
}

func TestFetchListeners_ShouldReturnDecrypted(t *testing.T) {
	gormDB, mock := newTestGormDB(t)

	rows := sqlmock.NewRows([]string{"listener_id", "listener_type", "listener_config"}).
		AddRow("someId", "http", nil)
	mock.ExpectQuery(`SELECT \* FROM "test"\."listeners_config"`).
		WillReturnRows(rows)

	dao := newTestDao(gormDB)

	listeners, err := dao.FetchListeners()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(listeners))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFetchRoute_NoRoute(t *testing.T) {
	gormDB, mock := newTestGormDB(t)

	mock.ExpectQuery(`SELECT \* FROM "test"\."route_global_config"`).
		WillReturnRows(sqlmock.NewRows([]string{"tenant_id"}))

	dao := newTestDao(gormDB)

	r, _ := dao.FetchRoute(listenerId, tenantId, routeId)
	assert.True(t, nil == r)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateRoute_ShouldInsertEncrypted(t *testing.T) {
	gormDB, mock := newTestGormDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "test"\."route_global_config"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`INSERT INTO "test"\."route_global_config"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	dao := newTestDao(gormDB)

	jconf, _ := jpath.New(configWithEncryptedPassword)
	jconf.GetNode("notify.email").Set("password", pwd)
	jstr, _ := jconf.JSON()

	err := dao.UpdateRoute(tenantId, routeId, listenerId, jstr)
	assert.Equal(t, nil, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFetchUserRoute_NoRoute(t *testing.T) {
	gormDB, mock := newTestGormDB(t)

	mock.ExpectQuery(`SELECT \* FROM "test"\."route_user_config"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}))

	dao := newTestDao(gormDB)

	r, _ := dao.FetchUserRoute(listenerId, tenantId, userId, routeId)
	assert.True(t, r == nil)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateUserRoute_ShouldInsertEncrypted(t *testing.T) {
	gormDB, mock := newTestGormDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "test"\."route_user_config"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`INSERT INTO "test"\."route_user_config"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	dao := newTestDao(gormDB)

	jconf, _ := jpath.New(configWithEncryptedPassword)
	jconf.GetNode("notify.email").Set("password", pwd)
	jstr, _ := jconf.JSON()

	err := dao.UpdateUserRoute(tenantId, userId, routeId, listenerId, jstr)
	assert.Equal(t, nil, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateListener_ShouldInsertEncrypted(t *testing.T) {
	gormDB, mock := newTestGormDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "test"\."listeners_config"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	dao := newTestDao(gormDB)

	jconf, _ := jpath.New(configWithEncryptedPassword)
	jconf.GetNode("notify.email").Set("password", pwd)
	jstr, _ := jconf.JSON()

	err := dao.CreateListener(listenerId, "type", jstr)
	assert.Equal(t, nil, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFetchSink_ShouldReturnDecrypted(t *testing.T) {
	gormDB, mock := newTestGormDB(t)

	mock.ExpectQuery(`SELECT \* FROM "test"\."sink_config"`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"sink_id", "type", "config"}).
			AddRow(sinkId, "email", datatypes.JSON(fmt.Sprintf(`{"type":"email","password":"%s"}`, encrypted))))

	dao := newTestDao(gormDB)

	sink, _ := dao.FetchSink(sinkId)
	jconf, _ := jpath.New(sink.Config)
	assert.Equal(t, pwd, jconf.GetNode("password").Interface().(string))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateSink_ShouldCallEncrypted(t *testing.T) {
	gormDB, mock := newTestGormDB(t)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "test"\."sink_config"`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	dao := newTestDao(gormDB)

	err := dao.UpdateSink(sinkId, "email", datatypes.JSON(fmt.Sprintf(`{"type":"email","password":"%s"}`, pwd)))
	assert.Nil(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateSink_ShouldInsertEncrypted(t *testing.T) {
	gormDB, mock := newTestGormDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "test"\."sink_config"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	dao := newTestDao(gormDB)

	err := dao.CreateSink(sinkId, "email", datatypes.JSON(fmt.Sprintf(`{"type":"email","password":"%s"}`, pwd)))
	assert.Nil(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateSink_SinksDaoError(t *testing.T) {
	gormDB, mock := newTestGormDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "test"\."sink_config"`).
		WillReturnError(fmt.Errorf("create failed"))
	mock.ExpectRollback()

	dao := newTestDao(gormDB)

	err := dao.CreateSink(sinkId, "email", datatypes.JSON(fmt.Sprintf(`{"type":"email","password":"%s"}`, pwd)))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFetchListener_NoListener(t *testing.T) {
	gormDB, mock := newTestGormDB(t)

	mock.ExpectQuery(`SELECT \* FROM "test"\."listeners_config"`).
		WillReturnRows(sqlmock.NewRows([]string{"listener_id"}))

	dao := newTestDao(gormDB)

	listener, err := dao.FetchListener(listenerId)
	assert.Nil(t, err)
	assert.Nil(t, listener)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFetchListener_ListenerExist(t *testing.T) {
	gormDB, mock := newTestGormDB(t)

	mock.ExpectQuery(`SELECT \* FROM "test"\."listeners_config"`).
		WillReturnRows(sqlmock.NewRows([]string{"listener_id", "listener_type"}).
			AddRow(listenerId, "http"))

	dao := newTestDao(gormDB)

	listener, err := dao.FetchListener(listenerId)
	assert.Nil(t, err)
	assert.Equal(t, listenerId, listener.ListenerId)
	assert.NoError(t, mock.ExpectationsWereMet())
}
