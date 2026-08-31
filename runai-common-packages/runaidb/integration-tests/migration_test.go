/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package integrationtests

import (
	"testing"

	"github.com/run-ai/runai/runai-common-packages/runaidb"
	dbconfig "github.com/run-ai/runai/runai-common-packages/runaidb/config"
	"github.com/stretchr/testify/require"
)

func TestMigration_SchemaMigrationsTableExists(t *testing.T) {
	var count int64
	err := runaidb.DB().Raw(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ?",
		"test_schema", "schema_migrations",
	).Row().Scan(&count)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}

func TestMigration_VersionIsCorrect(t *testing.T) {
	var version int
	var dirty bool
	err := runaidb.DB().Raw("SELECT version, dirty FROM test_schema.schema_migrations").Row().Scan(&version, &dirty)
	require.NoError(t, err)
	require.Equal(t, 3, version)
	require.False(t, dirty)
}

func TestMigration_FailuresTableExists(t *testing.T) {
	var count int64
	err := runaidb.DB().Raw(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ?",
		"test_schema", "migration_failures",
	).Row().Scan(&count)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}

func TestMigration_DirtyStateReportsStoredFailure(t *testing.T) {
	db := runaidb.DB()

	_, err := db.Exec("UPDATE test_schema.schema_migrations SET dirty = true")
	require.NoError(t, err)
	_, err = db.Exec(
		"INSERT INTO test_schema.migration_failures (version, details) VALUES (?, ?)",
		3, `relation "search_docs" does not exist`)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := db.Exec("UPDATE test_schema.schema_migrations SET dirty = false")
		require.NoError(t, err)
		_, err = db.Exec("DELETE FROM test_schema.migration_failures")
		require.NoError(t, err)
	})

	err = runaidb.Migrate(&runaidb.MigrationParams{Direction: dbconfig.MigrationDirectionUp})
	require.ErrorContains(t, err, "Dirty database version 3")
	require.ErrorContains(t, err, "migration 3 last failed at")
	require.ErrorContains(t, err, `relation "search_docs" does not exist`)
}

func TestMigration_AdvisoryLocksAreReleased(t *testing.T) {
	var count int64
	err := runaidb.DB().Raw("SELECT COUNT(*) FROM pg_locks WHERE locktype = 'advisory'").Row().Scan(&count)
	require.NoError(t, err)
	require.Zero(t, count)
}
