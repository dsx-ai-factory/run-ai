/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/internal/logger"
)

const (
	migrationFailuresTable = "migration_failures"
	schemaMigrationsTable  = "schema_migrations"

	// Short bound for failure-store DB ops (runs while the migration lock may be held).
	migrationFailuresOpTimeout = 5 * time.Second

	// schema_migrations.version is a primary key
	createFailuresTableQuery = `
		CREATE TABLE IF NOT EXISTS %s 
		(version bigint PRIMARY KEY, 
		created_at timestamptz NOT NULL DEFAULT now(), 
		failed_at timestamptz NOT NULL DEFAULT now(), 
		details text NOT NULL)`
	recordFailureQuery = `INSERT INTO %s (version, details) VALUES ($1, $2) ON CONFLICT (version) DO UPDATE SET failed_at = now(), details = EXCLUDED.details`
	selectFailureQuery = `SELECT failed_at, details FROM %s WHERE version = $1`
	selectVersionQuery = `SELECT version, dirty FROM %s LIMIT 1`
)

type dbExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type migrationFailuresStore struct {
	exec                dbExecutor
	migrationsTableName string
	failuresTableName   string
}

func newMigrationFailuresStore(exec dbExecutor, schema string) *migrationFailuresStore {
	return &migrationFailuresStore{exec: exec, failuresTableName: tableName(schema, migrationFailuresTable), migrationsTableName: tableName(schema, schemaMigrationsTable)}
}

func (s *migrationFailuresStore) Migrate(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, migrationFailuresOpTimeout)
	defer cancel()

	_, err := s.exec.ExecContext(ctx, fmt.Sprintf(createFailuresTableQuery, s.failuresTableName))
	return err
}

func (s *migrationFailuresStore) RecordFailure(ctx context.Context, cause error) error {
	ctx, cancel := context.WithTimeout(ctx, migrationFailuresOpTimeout)
	defer cancel()

	var dirty migrate.ErrDirty
	if errors.As(cause, &dirty) {
		return s.withStoredFailure(ctx, cause, int64(dirty.Version))
	}

	s.record(ctx, cause)
	return cause
}

func (s *migrationFailuresStore) record(ctx context.Context, cause error) {
	version, dirty, err := s.currentVersion(ctx)
	if err != nil {
		logger.GetLogger().Errorf("failed to resolve the failing migration version: %v", err)
		return
	}
	if !dirty {
		return
	}

	query := fmt.Sprintf(recordFailureQuery, s.failuresTableName)
	if _, err := s.exec.ExecContext(ctx, query, version, cause.Error()); err != nil {
		logger.GetLogger().Errorf("failed to store the failure of migration %d: %v", version, err)
	}
}

func (s *migrationFailuresStore) withStoredFailure(ctx context.Context, cause error, version int64) error {
	query := fmt.Sprintf(selectFailureQuery, s.failuresTableName)

	var failedAt time.Time
	var details string
	if err := s.exec.QueryRowContext(ctx, query, version).Scan(&failedAt, &details); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.GetLogger().Errorf("failed to read the stored failure of migration %d: %v", version, err)
		}
		return cause
	}

	return fmt.Errorf("%w migration %d last failed at %s: %s",
		cause, version, failedAt.UTC().Format(time.RFC3339), details)
}

func (s *migrationFailuresStore) currentVersion(ctx context.Context) (int64, bool, error) {
	query := fmt.Sprintf(selectVersionQuery, s.migrationsTableName)

	var version int64
	var dirty bool
	err := s.exec.QueryRowContext(ctx, query).Scan(&version, &dirty)
	return version, dirty, err
}

func tableName(schema, table string) string {
	return pgx.Identifier{schema, table}.Sanitize()
}
