/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package integrationtests

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	dbconfig "github.com/run-ai/runai/runai-common-packages/runaidb/config"
	"github.com/stretchr/testify/require"
)

const failingMigrationError = `relation "table_that_does_not_exist" does not exist`

func TestMigrationFailure_ReasonIsRecorded(t *testing.T) {
	db := connectAsAdmin(t)
	useFailingMigrations(t, db, "failing_migration_recorded")

	err := runaidb.Migrate(&runaidb.MigrationParams{Direction: dbconfig.MigrationDirectionUp})
	require.ErrorContains(t, err, failingMigrationError)

	var version int64
	var details string
	err = db.
		QueryRow(context.Background(), fmt.Sprintf("SELECT version, details FROM %s", failuresTable("failing_migration_recorded"))).
		Scan(&version, &details)
	require.NoError(t, err)
	require.EqualValues(t, 1, version)
	require.Contains(t, details, failingMigrationError)
}

func TestMigrationFailure_RerunReportsRecordedReason(t *testing.T) {
	db := connectAsAdmin(t)
	useFailingMigrations(t, db, "failing_migration_rerun")

	err := runaidb.Migrate(&runaidb.MigrationParams{Direction: dbconfig.MigrationDirectionUp})
	require.ErrorContains(t, err, failingMigrationError)

	err = runaidb.Migrate(&runaidb.MigrationParams{Direction: dbconfig.MigrationDirectionUp})
	require.ErrorContains(t, err, "Dirty database version 1")
	require.ErrorContains(t, err, "migration 1 last failed at")
	require.ErrorContains(t, err, failingMigrationError)
}

func useFailingMigrations(t *testing.T, admin *pgx.Conn, schema string) {
	conf := config.Get[*dbconfig.Config]()
	originalMigration, originalSchema := conf.Migration, conf.Connection.Schema

	autoMigrate := false
	conf.Migration.Files = "file://" + failingMigrationsDir()
	conf.Migration.Auto = &autoMigrate
	conf.Connection.Schema = schema

	t.Cleanup(func() {
		conf.Migration, conf.Connection.Schema = originalMigration, originalSchema

		_, err := admin.Exec(context.Background(),
			fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", pgx.Identifier{schema}.Sanitize()))
		require.NoError(t, err)
	})
}

func connectAsAdmin(t *testing.T) *pgx.Conn {
	conn := config.Get[*dbconfig.Config]().Connection
	cfg, err := pgx.ParseConfig("sslmode=disable")
	require.NoError(t, err)

	cfg.User = conn.User
	cfg.Password = conn.Password
	cfg.Host = conn.Host
	cfg.Port = uint16(conn.Port)
	cfg.Database = conn.Name

	admin, err := pgx.ConnectConfig(context.Background(), cfg)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, admin.Close(context.Background()))
	})

	return admin
}

func failuresTable(schema string) string {
	return pgx.Identifier{schema, "migration_failures"}.Sanitize()
}

func failingMigrationsDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "migrations_failure")
}
