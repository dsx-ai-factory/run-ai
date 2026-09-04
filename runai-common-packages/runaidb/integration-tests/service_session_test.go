/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package integrationtests

import (
	"errors"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

const ErrCodeInsufficientPrivilege = "42501"

func TestServiceSession_CurrentUserIsServiceUser(t *testing.T) {
	var currentUser string
	err := runaidb.DB().Raw("SELECT current_user").Row().Scan(&currentUser)
	require.NoError(t, err)
	require.Equal(t, "test-integdb", currentUser)
}

func TestServiceSession_CanInsert(t *testing.T) {
	record := TestSecret{Name: "session-insert-test", Secret: "value"}
	_, err := runaidb.DB().Create(&record)
	require.NoError(t, err)
	require.NotZero(t, record.ID)
}

func TestServiceSession_CanSelect(t *testing.T) {
	record := TestSecret{Name: "session-select-test", Secret: "value"}
	_, err := runaidb.DB().Create(&record)
	require.NoError(t, err)

	var result TestSecret
	err = runaidb.DB().First(&result, record.ID)
	require.NoError(t, err)
	require.Equal(t, record.Name, result.Name)
}

func TestServiceSession_CanUpdate(t *testing.T) {
	record := TestSecret{Name: "session-update-test", Secret: "value"}
	_, err := runaidb.DB().Create(&record)
	require.NoError(t, err)

	_, err = runaidb.DB().Model(&TestSecret{}).Where("id = ?", record.ID).Update("name", "updated-name")
	require.NoError(t, err)

	var result TestSecret
	err = runaidb.DB().First(&result, record.ID)
	require.NoError(t, err)
	require.Equal(t, "updated-name", result.Name)
}

func TestServiceSession_CanDelete(t *testing.T) {
	record := TestSecret{Name: "session-delete-test", Secret: "value"}
	_, err := runaidb.DB().Create(&record)
	require.NoError(t, err)

	_, err = runaidb.DB().Delete(&TestSecret{}, record.ID)
	require.NoError(t, err)

	var count int64
	err = runaidb.DB().Model(&TestSecret{}).Where("id = ?", record.ID).Count(&count)
	require.NoError(t, err)
	require.Equal(t, int64(0), count)
}

func TestServiceSession_CannotCreateTable(t *testing.T) {
	_, err := runaidb.DB().Exec("CREATE TABLE test_schema.unauthorized_table (id SERIAL PRIMARY KEY)")
	require.True(t, isInsufficientPrivilegeErr(err))
}

func TestServiceSession_CannotDropSchema(t *testing.T) {
	_, err := runaidb.DB().Exec("DROP SCHEMA test_schema CASCADE")
	require.True(t, isInsufficientPrivilegeErr(err))
}

func TestServiceSession_CannotSelectFromOtherSchema(t *testing.T) {
	var result string
	err := runaidb.DB().Raw("SELECT * FROM restricted_schema.restricted_table").Row().Scan(&result)
	require.True(t, isInsufficientPrivilegeErr(err))
}

func TestServiceSession_CannotCreateRecordInOtherSchema(t *testing.T) {
	_, err := runaidb.DB().Exec("INSERT INTO restricted_schema.restricted_table (name) VALUES ('test')")
	require.True(t, isInsufficientPrivilegeErr(err))
}

func isInsufficientPrivilegeErr(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == ErrCodeInsufficientPrivilege
}
