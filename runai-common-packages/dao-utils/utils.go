/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_utils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4/database"
	"gorm.io/gorm"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/golang-migrate/migrate/v4"
	goMigratePostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func connectToDbMigrator(dsn url.URL, schema string) *sql.DB {
	// TODO - no need use gorm.Exec.. and factory
	migrator, err := sql.Open("postgres", dsn.String())
	if err != nil {
		logging.Fatalf("failed connecting to database: %v", err)
	}
	_, err = migrator.ExecContext(context.Background(), fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS \"%s\"", schema))
	if err != nil {
		logging.Fatalf("failed creating schema: %v", err)
	}
	_, err = migrator.ExecContext(context.Background(), fmt.Sprintf("SET search_path TO %s", schema))
	if err != nil {
		logging.Fatalf("failed connecting to schema: %v", err)
	}
	return migrator
}

type MigratorDsnAndSchema struct {
	dsn    string
	schema string
}

var migrators = map[MigratorDsnAndSchema]database.Driver{}

func getMigrate(dsn url.URL, schema string, databaseName string, migrationFiles string, migrateDown bool) (*migrate.Migrate, string, error) {
	migratorDsnAndSchema := MigratorDsnAndSchema{
		dsn:    dsn.String(),
		schema: schema,
	}

	driver, foundDriver := migrators[migratorDsnAndSchema]
	if !foundDriver {
		var err error
		migrator := connectToDbMigrator(dsn, schema)
		driver, err = goMigratePostgres.WithInstance(migrator, &goMigratePostgres.Config{})
		if err != nil {
			logging.Errorf("failed retrieving postgres driver: %v", err)
			return nil, "", err
		}

		migrators[migratorDsnAndSchema] = driver
	}

	var migrationFilesLocation string
	if migrateDown {
		migrationFilesLocation = migrationFiles + "/down"
	} else {
		migrationFilesLocation = migrationFiles + "/up"
	}
	m, err := migrate.NewWithDatabaseInstance(
		migrationFilesLocation,
		databaseName,
		driver,
	)
	if err != nil {
		logging.Errorf("failed retrieving postgres migrations: %v", err)
		return nil, "", err
	}
	return m, migrationFilesLocation, nil
}

func getVersion(m *migrate.Migrate, versionCommand string, stepDown bool) (uint, error) {
	version, err := strconv.ParseUint(versionCommand, 10, 64)
	if err != nil {
		logging.Errorf("failed running migrations: %v", err)
		return 0, err
	}
	targetVersion := uint(version)
	currVersion, _, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		logging.Errorf("failed running migrations, cannot get current version: %v", err)
		return 0, err
	}
	if (stepDown && targetVersion > currVersion) || (!stepDown && targetVersion < currVersion) {
		errorMsg := fmt.Sprintf("failed running migrations, cannot step up/down from %d to %d", currVersion, targetVersion)
		logging.Errorf("%s", errorMsg)
		return 0, errors.New(errorMsg)
	}
	return targetVersion, nil
}

func FinalizeTransaction(tx *gorm.DB, committed *bool) {
	if !*committed {
		logging.Errorf("transaction was not committed, undoing all changes")
		tx.Rollback()
		return
	}

	if r := recover(); r != nil {
		logging.Errorf("panic in transaction: %v, undoing all changes", r)
		tx.Rollback()
	}
}

func ListItemsForQuery(ids []string) string {
	base := "("
	for _, id := range ids {
		base += fmt.Sprintf(" '%s',", id)
	}
	return strings.TrimSuffix(base, ",") + ")"
}

func GormDeletedAtToNullableTime(time gorm.DeletedAt) *time.Time {
	if !time.Valid {
		return nil
	}

	return &time.Time
}

// Extract numbers from filenames in a given directory and return the closest numbers above and below the input.
func getClosestNumbers(dir string, input int) (int, int, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return 0, 0, err
	}

	var numbers []int
	re := regexp.MustCompile(`^\d+`)

	// Extract numbers from filenames
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		match := re.FindString(file.Name())
		if match != "" {
			num, err := strconv.Atoi(match)
			if err == nil {
				numbers = append(numbers, num)
			}
		}
	}

	// Sort the numbers
	sort.Ints(numbers)

	// Find the closest numbers
	var below, above int
	below, above = -1, -1

	for _, num := range numbers {
		if num < input {
			below = num
		} else if num > input {
			above = num
			break
		}
	}

	return below, above, nil
}
