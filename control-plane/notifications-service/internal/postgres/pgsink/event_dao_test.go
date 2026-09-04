/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pgsink

import (
	"fmt"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const eventTestTable = `"test"."events"`

func newEventDaoTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(gormPostgres.New(gormPostgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	require.NoError(t, err)
	return gormDB, mock
}

func validEvent() events.Event {
	return events.Event{
		UID:        "evt-1",
		ListenerId: "listener",
		Properties: map[string]interface{}{
			"tenant_id":   "tenant",
			"user_id":     "user",
			"description": "something happened",
		},
	}
}

func TestEventDaoStore_Success(t *testing.T) {
	gormDB, mock := newEventDaoTestDB(t)

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "test"\."events"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ed := &EventDao{db: gormDB, table: eventTestTable}
	err := ed.Store(validEvent())
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEventDaoStore_Failed(t *testing.T) {
	gormDB, mock := newEventDaoTestDB(t)

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "test"\."events"`).
		WillReturnError(fmt.Errorf("insert failed"))
	mock.ExpectRollback()

	ed := &EventDao{db: gormDB, table: eventTestTable}
	err := ed.Store(validEvent())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insert failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEventDaoStore_TenantIDMissing(t *testing.T) {
	gormDB, mock := newEventDaoTestDB(t)
	ed := &EventDao{db: gormDB, table: eventTestTable}

	ev := validEvent()
	delete(ev.Properties, "tenant_id")

	require.NoError(t, ed.Store(ev))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEventDaoStore_TenantIDWrongType(t *testing.T) {
	gormDB, mock := newEventDaoTestDB(t)
	ed := &EventDao{db: gormDB, table: eventTestTable}

	ev := validEvent()
	ev.Properties["tenant_id"] = 42

	require.NoError(t, ed.Store(ev))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEventDaoStore_UserIDMissing(t *testing.T) {
	gormDB, mock := newEventDaoTestDB(t)
	ed := &EventDao{db: gormDB, table: eventTestTable}

	ev := validEvent()
	delete(ev.Properties, "user_id")

	require.NoError(t, ed.Store(ev))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEventDaoStore_DescriptionMissing(t *testing.T) {
	gormDB, mock := newEventDaoTestDB(t)
	ed := &EventDao{db: gormDB, table: eventTestTable}

	ev := validEvent()
	delete(ev.Properties, "description")

	require.NoError(t, ed.Store(ev))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEventDaoIsHealthy(t *testing.T) {
	gormDB, mock := newEventDaoTestDB(t)

	mock.ExpectExec(`select count\(\*\) from "test"\."events"`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	ed := &EventDao{db: gormDB, table: eventTestTable}
	ok, err := ed.IsHealthy()
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEventDaoIsHealthy_Failed(t *testing.T) {
	gormDB, mock := newEventDaoTestDB(t)

	mock.ExpectExec(`select count\(\*\) from "test"\."events"`).
		WillReturnError(fmt.Errorf("table missing"))

	ed := &EventDao{db: gormDB, table: eventTestTable}
	ok, err := ed.IsHealthy()
	require.Error(t, err)
	assert.False(t, ok)
	assert.Contains(t, err.Error(), "table missing")
	assert.NoError(t, mock.ExpectationsWereMet())
}
