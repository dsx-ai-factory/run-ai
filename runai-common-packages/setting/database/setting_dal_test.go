/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	testKey      = "test-key"
	testTenantID = "tenant-1"

	sqlSelectByKey    = `SELECT * FROM settings WHERE key = ? and tenant_id = ? ORDER BY settings.id LIMIT ?`
	sqlSelectByTenant = `SELECT * FROM settings WHERE tenant_id = ?`
	sqlSelectKeyOnly  = `SELECT * FROM settings WHERE key = ? ORDER BY settings.id LIMIT ?`
	sqlInsert         = `INSERT INTO settings (id,key,value,tenant_id,updated_at,updated_by_customer) VALUES (?,?,?,?,?,?)`
	sqlBulkInsert     = `INSERT INTO settings (id,key,value,tenant_id,updated_at,updated_by_customer) VALUES (?,?,?,?,?,?) ON CONFLICT DO NOTHING`
	sqlUpdate         = `UPDATE settings SET key=?,value=?,tenant_id=?,updated_at=?,updated_by_customer=? WHERE (key = ? AND tenant_id = ?) AND id = ?`
)

func newTestSetting() Setting {
	return Setting{
		ID:                uuid.New(),
		Key:               testKey,
		Value:             []byte(`{"a":1}`),
		TenantID:          testTenantID,
		UpdatedByCustomer: true,
	}
}

func TestGetSettingByKey(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	expected := newTestSetting()
	mocker.ExpectQuery(sqlSelectByKey, testKey, testTenantID, 1).Returns([]Setting{expected})

	got, err := NewSettingDAL().GetSettingByKey(testKey, testTenantID)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, expected.ID, got.ID)
	require.Equal(t, expected.Key, got.Key)
	require.Equal(t, expected.Value, got.Value)
	require.Equal(t, expected.TenantID, got.TenantID)
	require.Equal(t, expected.UpdatedByCustomer, got.UpdatedByCustomer)
}

func TestGetSettingByKey_NotFound(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectQuery(sqlSelectByKey, testKey, testTenantID, 1).Returns([]Setting{})

	got, err := NewSettingDAL().GetSettingByKey(testKey, testTenantID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	require.Nil(t, got)
}

func TestGetSettingByKey_Error(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectQuery(sqlSelectByKey, testKey, testTenantID, 1).ReturnsError(errors.New("db error"))

	got, err := NewSettingDAL().GetSettingByKey(testKey, testTenantID)
	require.EqualError(t, err, "db error")
	require.Nil(t, got)
}

func TestGetSettingsByTenantId(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	first, second := newTestSetting(), newTestSetting()
	mocker.ExpectQuery(sqlSelectByTenant, testTenantID).Returns([]Setting{first, second})

	got, err := NewSettingDAL().GetSettingsByTenantId(testTenantID)
	require.NoError(t, err)
	require.Len(t, *got, 2)
	require.Equal(t, first.ID, (*got)[0].ID)
	require.Equal(t, second.ID, (*got)[1].ID)
}

func TestGetSettingsByTenantId_Empty(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectQuery(sqlSelectByTenant, testTenantID).Returns([]Setting{})

	got, err := NewSettingDAL().GetSettingsByTenantId(testTenantID)
	require.NoError(t, err)
	require.Empty(t, *got)
}

func TestGetSettingsByTenantId_Error(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectQuery(sqlSelectByTenant, testTenantID).ReturnsError(errors.New("db error"))

	got, err := NewSettingDAL().GetSettingsByTenantId(testTenantID)
	require.EqualError(t, err, "db error")
	require.Nil(t, got)
}

func TestShouldMigrateSetting_NotFound(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectQuery(sqlSelectKeyOnly, testKey, 1).Returns([]Setting{})

	should, err := NewSettingDAL().ShouldMigrateSetting(testKey)
	require.NoError(t, err)
	require.True(t, should)
}

func TestShouldMigrateSetting_Exists(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectQuery(sqlSelectKeyOnly, testKey, 1).Returns([]Setting{newTestSetting()})

	should, err := NewSettingDAL().ShouldMigrateSetting(testKey)
	require.NoError(t, err)
	require.False(t, should)
}

func TestShouldMigrateSetting_Error(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectQuery(sqlSelectKeyOnly, testKey, 1).ReturnsError(errors.New("db error"))

	should, err := NewSettingDAL().ShouldMigrateSetting(testKey)
	require.EqualError(t, err, "db error")
	require.False(t, should)
}

func TestSaveSetting_Insert(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectQuery(sqlSelectByKey, testKey, testTenantID, 1).Returns([]Setting{})
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlInsert, dbtesting.AnyArg(), testKey, dbtesting.AnyArg(), testTenantID, dbtesting.AnyTime(), true).Returns(1)
	mocker.ExpectTransaction().Commit()

	err := NewSettingDAL().SaveSetting(testKey, testTenantID, true, []byte(`{"a":1}`))
	require.NoError(t, err)
}

func TestSaveSetting_Update(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectQuery(sqlSelectByKey, testKey, testTenantID, 1).Returns([]Setting{newTestSetting()})
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlUpdate, testKey, dbtesting.AnyArg(), testTenantID, dbtesting.AnyTime(), true, testKey, testTenantID, dbtesting.AnyArg()).Returns(1)
	mocker.ExpectTransaction().Commit()

	err := NewSettingDAL().SaveSetting(testKey, testTenantID, true, []byte(`{"a":1}`))
	require.NoError(t, err)
}

func TestSaveSetting_InsertError(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectQuery(sqlSelectByKey, testKey, testTenantID, 1).Returns([]Setting{})
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlInsert, dbtesting.AnyArg(), testKey, dbtesting.AnyArg(), testTenantID, dbtesting.AnyTime(), true).ReturnsError(errors.New("insert failed"))
	mocker.ExpectTransaction().Rollback()

	err := NewSettingDAL().SaveSetting(testKey, testTenantID, true, []byte(`{"a":1}`))
	require.EqualError(t, err, "insert failed")
}

func TestSaveSettingByKeyForTenantsList(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectTransaction().Begin()
	sqlBulkInsertTwo := `INSERT INTO settings (id,key,value,tenant_id,updated_at,updated_by_customer) VALUES (?,?,?,?,?,?),(?,?,?,?,?,?) ON CONFLICT DO NOTHING`
	mocker.ExpectExec(sqlBulkInsertTwo,
		dbtesting.AnyArg(), testKey, dbtesting.AnyArg(), "tenant-a", dbtesting.AnyTime(), true,
		dbtesting.AnyArg(), testKey, dbtesting.AnyArg(), "tenant-b", dbtesting.AnyTime(), true,
	).Returns(2)
	mocker.ExpectTransaction().Commit()

	err := NewSettingDAL().SaveSettingByKeyForTenantsList(testKey, []TenantSettingInput{
		{TenantID: "tenant-a", Value: []byte(`{"a":1}`)},
		{TenantID: "tenant-b", Value: []byte(`{"b":2}`)},
	})
	require.NoError(t, err)
}

func TestSaveSettingByKeyForTenantsList_Empty(t *testing.T) {
	dbtesting.SetSQLMock(t)

	err := NewSettingDAL().SaveSettingByKeyForTenantsList(testKey, nil)
	require.NoError(t, err)
}

func TestSaveSettingByKeyForTenantsList_Error(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlBulkInsert, dbtesting.AnyArg(), testKey, dbtesting.AnyArg(), "tenant-a", dbtesting.AnyTime(), true).ReturnsError(errors.New("bulk failed"))
	mocker.ExpectTransaction().Rollback()

	err := NewSettingDAL().SaveSettingByKeyForTenantsList(testKey, []TenantSettingInput{
		{TenantID: "tenant-a", Value: []byte(`{"a":1}`)},
	})
	require.EqualError(t, err, "bulk failed")
}
