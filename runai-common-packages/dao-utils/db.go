/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_utils

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/golang-migrate/migrate/v4"
	"gorm.io/gorm"
)

const (
	PostgresDriverName = "postgres"
)

var (
	conf              *DataBaseConfig
	onceWrite         sync.Once
	onceRead          sync.Once
	initOnce          sync.Once
	globalWriterDb    *gorm.DB
	globalReaderDb    *gorm.DB
	useServiceSession bool
	pgLock            *PGLock
)

func Init(dataBaseConfig *DataBaseConfig) *gorm.DB {
	initOnce.Do(func() {
		logging.Debugf("initializing database")
		conf = dataBaseConfig
		pgLock = NewPGLock(conf.Schema)

		Migrate(dataBaseConfig)

		useServiceSession = dataBaseConfig.UseServiceSession
		err := createServiceDbUser(dataBaseConfig)
		if err != nil {
			logging.Warnf("Failed to create service user: %v, defaulting to configured user", err)
			useServiceSession = false
		}
	})

	return DB()
}

// DB initializes (once) a database object base on env variables
// and returns a pointer to the DB object
func DB() *gorm.DB {
	onceWrite.Do(ReconnectDb)
	return globalWriterDb
}

// usually for testing purposes, if we want to reconnect the db.
// for example, after changing the schema in runtime during a test.
func ReconnectDb() {
	factory := &ConnectionFactory{}
	globalWriterDb = factory.Get(CTWriter, conf)
}

func DBRead() *gorm.DB {
	onceRead.Do(func() {
		factory := &ConnectionFactory{}
		globalReaderDb = factory.Get(CTReader, conf)
		if globalReaderDb == nil {
			globalReaderDb = DB()
		}
	})
	return globalReaderDb
}

type MigrateLogger struct{}

func (l *MigrateLogger) Printf(format string, v ...interface{}) {
	logging.Infof(format, v...)
}

func (l *MigrateLogger) Verbose() bool {
	return true
}

func Migrate(conf *DataBaseConfig) {
	logging.Infof("Start migration of schema %s with command %s", conf.Schema, conf.MigrationCommand)

	acquiredLock, err := pgLock.TryAcquireSchemaLock(DB(), conf.MigrationLockRetries)
	if err != nil {
		logging.Fatalf("Got an error trying to acquire schema %s lock. error: %v", conf.Schema, err)
	}
	if !acquiredLock {
		logging.Fatalf("Lock for schema %s already taken", conf.Schema)
	}
	defer pgLock.ReleaseSchemaLock(DB())

	// TODO - work with gorm factory
	dsn := createDsnNew(conf.Name, conf.User, conf.Password, conf.Schema, conf.Host, conf.Port, conf.SslMode, conf.StatementTimeout)

	migrationCommands := strings.Fields(conf.MigrationCommand)
	if len(migrationCommands) == 0 {
		logging.Infof("Migrating up")
		m, migrationFilesLocation, err := getMigrate(dsn, conf.Schema, conf.Name, conf.MigrationFiles, false)
		if conf.VerboseLogging {
			m.Log = &MigrateLogger{}
		}
		if err != nil {
			logging.Fatalf("failed retrieving postgres migrations: %v", err)
			return
		}

		err = m.Up()
		if err != nil {
			if strings.Contains(err.Error(), "no migration found for version") {
				logging.Warnf("Unknown current migration version: %v", err)
				err = forceMigrateOnUnknownVersion(m, migrationFilesLocation)
			}

			if err != nil {
				if err != migrate.ErrNoChange {
					logging.Fatalf("failed running migrations: %v", err)
				} else {
					logging.Debugf("migration ended without changes")
				}
			}
		}
	} else if len(migrationCommands) == 1 {
		command := migrationCommands[0]
		var m *migrate.Migrate
		switch command {
		case "downmost":
			logging.Infof("Migrating downmost")
			m, _, err = getMigrate(dsn, conf.Schema, conf.Name, conf.MigrationFiles, true)
			if err != nil {
				logging.Fatalf("failed retrieving postgres migrations: %v", err)
			}
			err = m.Down()
		case "upmost":
			logging.Infof("Migrating upmost")
			m, _, err = getMigrate(dsn, conf.Schema, conf.Name, conf.MigrationFiles, false)
			if err != nil {
				logging.Fatalf("failed retrieving postgres migrations: %v", err)
			}
			err = m.Up()
		}
	} else if len(migrationCommands) == 2 {
		command := migrationCommands[0]
		versionCommand := migrationCommands[1]
		var m *migrate.Migrate
		switch command {
		case "down":
			logging.Infof("Migrating down to version %s", versionCommand)
			m, _, err = getMigrate(dsn, conf.Schema, conf.Name, conf.MigrationFiles, true)
			if err != nil {
				logging.Fatalf("failed retrieving postgres migrations: %v", err)
			}
			var targetVersion uint
			targetVersion, err = getVersion(m, versionCommand, true)
			if err != nil {
				logging.Fatalf("failed running migrations: %v", err)
			}
			err = m.Migrate(targetVersion)
		case "up":
			logging.Infof("Migrating up to version %s", versionCommand)
			m, _, err = getMigrate(dsn, conf.Schema, conf.Name, conf.MigrationFiles, false)
			if err != nil {
				logging.Fatalf("failed retrieving postgres migrations: %v", err)
			}
			var targetVersion uint
			targetVersion, err = getVersion(m, versionCommand, false)
			if err != nil {
				logging.Fatalf("failed running migrations: %v", err)
			}
			err = m.Migrate(targetVersion)
		case "force":
			logging.Infof("Migration force version to %s", versionCommand)
			m, _, err = getMigrate(dsn, conf.Schema, conf.Name, conf.MigrationFiles, false)
			if err != nil {
				logging.Fatalf("failed retrieving postgres migrations: %v", err)
			}
			var targetVersion int
			targetVersion, err = strconv.Atoi(migrationCommands[1])
			if err != nil {
				logging.Fatalf("failed running migrations: %v", err)
			}
			err = m.Force(targetVersion)
		}
	}
	if err != nil {
		if err != migrate.ErrNoChange {
			logging.Fatalf("failed running migrations: %v", err)
		}
		logging.Debugf("Migration ended without changes")
	}

	logging.Infof("Migrations ended successfully")
}

func forceMigrateOnUnknownVersion(m *migrate.Migrate, migrationFilesLocation string) error {
	_, migrationFilesLocation, _ = strings.Cut(migrationFilesLocation, "file://")

	curVersion, _, err := m.Version()
	if err != nil {
		return err
	}

	prev, next, err := getClosestNumbers(migrationFilesLocation, int(curVersion)) //nolint:gosec // G115: value is bounded/non-negative
	if err != nil {
		return err
	}

	if next == -1 {
		logging.Infof("No existing migration higher than %d. Nothing to do.", curVersion)
		return nil
	}

	if prev == -1 {
		return fmt.Errorf("no existing migration lower than %d. Cant fix", curVersion)
	}

	logging.Infof("Found migrations %d and %d that are the closest to %d", prev, next, curVersion)

	err = m.Force(prev)
	if err != nil {
		logging.Errorf("Faield to force migrate to %d", prev)
		return err
	}

	return m.Up()
}

func createServiceDbUser(config *DataBaseConfig) error {
	if config.ServiceName == "" || !useServiceSession {
		return nil
	}

	formattedServiceName := getUserFromConfig(config)
	factory := &ConnectionFactory{}
	db := factory.Get(CTAdmin, config)

	// create a user in the service name
	result := db.Exec(fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s';", formattedServiceName, config.Password))
	if result.Error != nil && !strings.Contains(result.Error.Error(), "already exists") {
		return result.Error
	}
	// give the user permissions to current schema
	result = db.Exec(fmt.Sprintf("GRANT USAGE ON SCHEMA %s TO %s;", config.Schema, formattedServiceName))
	if result.Error != nil {
		return result.Error
	}
	result = db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA %s TO %s;", config.Schema, formattedServiceName))
	if result.Error != nil {
		return result.Error
	}
	result = db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA %s TO %s;", config.Schema, formattedServiceName))
	if result.Error != nil {
		return result.Error
	}
	result = db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA %s TO %s;", config.Schema, formattedServiceName))
	if result.Error != nil {
		return result.Error
	}
	logging.Infof("db service session user was created successfully")
	return nil
}

func SetMock(mock *gorm.DB) {
	globalWriterDb = mock
	globalReaderDb = mock
	onceWrite.Do(func() {})
	onceRead.Do(func() {})
}

func ResetMock() {
	globalWriterDb = nil
	globalReaderDb = nil
	onceWrite = sync.Once{}
	onceRead = sync.Once{}
	initOnce = sync.Once{}
}
