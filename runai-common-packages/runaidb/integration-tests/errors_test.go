/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package integrationtests

import (
	"testing"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/stretchr/testify/require"
)

const missingTableQuery = "SELECT * FROM test_schema.no_such_table"

func TestUndefinedTable_Select(t *testing.T) {
	err := runaidb.DB().Actual().Exec(missingTableQuery).Error
	require.ErrorIs(t, err, orm.ErrUndefinedTable)
}

func TestUndefinedTable_SelectRaw(t *testing.T) {
	rows, err := runaidb.DB().Raw(missingTableQuery).Rows()
	require.Error(t, err)
	require.Nil(t, rows)
	fmt.Println(err)
	require.ErrorIs(t, err, orm.ErrUndefinedTable)
}

func TestUndefinedTable_Insert(t *testing.T) {
	err := runaidb.DB().Actual().Exec(
		"INSERT INTO test_schema.no_such_table (name) VALUES (?)", "value",
	).Error
	require.ErrorIs(t, err, orm.ErrUndefinedTable)
}

func TestDuplicatedKey_RawInsert(t *testing.T) {
	var id int
	err := runaidb.DB().Raw(
		"INSERT INTO test_schema.test_secrets (name, secret) VALUES (?, ?) RETURNING id",
		"dup-key-raw", "raw-secret",
	).Row().Scan(&id)
	require.NoError(t, err)
	require.NotZero(t, id)

	rows, err := runaidb.DB().Raw(
		"INSERT INTO test_schema.test_secrets (id, name, secret) VALUES (?, ?, ?)",
		id, "dup-key-raw", "raw-secret",
	).Rows()
	require.ErrorIs(t, err, orm.ErrDuplicatedKey)
	require.Nil(t, rows)
}

func TestUndefinedTable_PreservesPgError(t *testing.T) {
	err := runaidb.DB().Actual().Exec(missingTableQuery).Error

	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, "42P01", pgErr.Code)
	require.Contains(t, pgErr.Message, "no_such_table")
}

func TestUndefinedTable_IsNotOtherSentinels(t *testing.T) {
	err := runaidb.DB().Actual().Exec(missingTableQuery).Error

	require.NotErrorIs(t, err, orm.ErrReadOnlyConnection)
	require.NotErrorIs(t, err, orm.ErrRecordNotFound)
}

func TestReadOnlyConnection_PreservesPgError(t *testing.T) {
	err := requireReadDB(t).Actual().Exec(
		"INSERT INTO test_schema.test_secrets (name, secret) VALUES (?, ?)",
		"read-only-pgerror", "value",
	).Error

	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, "25006", pgErr.Code)
}
