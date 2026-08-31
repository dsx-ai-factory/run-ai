/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	dbconfig "github.com/run-ai/runai/runai-common-packages/runaidb/config"
	"github.com/run-ai/runai/runai-common-packages/runaidb/internal/logger"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

type Migrator struct {
	db orm.DB
}

func NewMigrator(db orm.DB) *Migrator {
	return &Migrator{db: db}
}

func (m *Migrator) Migrate(conf *dbconfig.Config) error {
	sqlDB, err := m.db.Actual().DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	conn, err := sqlDB.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("failed to reserve a migration connection: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			logger.GetLogger().Errorf("failed to return migration connection to the pool: %v", err)
		}
	}()

	return m.migrateOnConn(conn, sqlDB, conf)
}

func (m *Migrator) migrateOnConn(conn *sql.Conn, audit dbExecutor, conf *dbconfig.Config) error {
	if err := m.ensureSchemaExists(conn, conf.Connection.Schema); err != nil {
		return fmt.Errorf("failed to ensure schema %q exists: %w", conf.Connection.Schema, err)
	}

	lock := NewLocker(conn)
	acquired, err := lock.Acquire(conf.Connection.Schema)
	if err != nil {
		return fmt.Errorf("failed to acquire migration lock for schema %q: %w", conf.Connection.Schema, err)
	}
	if !acquired {
		return fmt.Errorf("migration lock for schema %q is already held", conf.Connection.Schema)
	}
	defer func() {
		released, err := lock.Release(conf.Connection.Schema)
		if err != nil {
			logger.GetLogger().Errorf("failed to release migration lock for schema %q: %v", conf.Connection.Schema, err)
			return
		}
		if !released {
			logger.GetLogger().Errorf("migration lock for schema %q was not held by the releasing session", conf.Connection.Schema)
		}
	}()

	store := newMigrationFailuresStore(audit, conf.Connection.Schema)
	if err := store.Migrate(context.Background()); err != nil {
		logger.GetLogger().Errorf("failed to migrate the %s table: %v", migrationFailuresTable, err)
	}

	migRunner, err := m.getMigrationRunner(conn, conf)
	if err != nil {
		return err
	}
	return m.executeMigration(migRunner, store, conf)
}

func (m *Migrator) getMigrationRunner(conn *sql.Conn, conf *dbconfig.Config) (migrationRunner, error) {
	driver, err := migratepostgres.WithConnection(context.Background(), conn, &migratepostgres.Config{
		SchemaName: conf.Connection.Schema,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres migration driver: %w", err)
	}

	mig, err := migrate.NewWithDatabaseInstance(getMigrationFilesPath(conf), conf.Connection.Name, driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration instance: %w", err)
	}
	mig.Log = &migrationLogger{}

	return mig, nil
}

func (m *Migrator) executeMigration(mig migrationRunner, store *migrationFailuresStore, conf *dbconfig.Config) error {
	version, err := evaluateVersion(conf.Migration.TargetVersion)
	if err != nil {
		return err
	}

	strategy, err := newMigrationStrategy(mig, conf.Migration.Direction)
	if err != nil {
		return err
	}
	if err := strategy.execute(version); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", store.RecordFailure(context.Background(), err))
	}

	return nil
}

func evaluateVersion(targetVersion *string) (*uint, error) {
	if targetVersion == nil {
		return nil, nil
	}
	v, err := strconv.ParseUint(*targetVersion, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid target version %q: %w", *targetVersion, err)
	}
	version := uint(v)
	return &version, nil
}

func getMigrationFilesPath(conf *dbconfig.Config) string {
	if conf.Migration.Direction == dbconfig.MigrationDirectionDown {
		return conf.Migration.Files + "/down"
	}
	return conf.Migration.Files + "/up"
}

func (m *Migrator) ensureSchemaExists(conn *sql.Conn, schema string) error {
	sanitizedSchema := pgx.Identifier{schema}.Sanitize()
	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s;", sanitizedSchema)
	_, err := conn.ExecContext(context.Background(), query)
	return err
}
