/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaidb

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/run-ai/runai/runai-common-packages/config"
	dbconfig "github.com/run-ai/runai/runai-common-packages/runaidb/config"
	"github.com/run-ai/runai/runai-common-packages/runaidb/internal/dbinit"
	"github.com/run-ai/runai/runai-common-packages/runaidb/internal/logger"
	"github.com/run-ai/runai/runai-common-packages/runaidb/mocks"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
	ormmocks "github.com/run-ai/runai/runai-common-packages/runaidb/orm/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	defer cleanup()

	dbClient := orm.NewDB(nil, nil)
	dbInitializerMock := mocks.NewDBInitializer(t)
	dbInitializerMock.On("InitializeDB", mock.Anything).Return(dbClient, nil)

	dbInit = dbInitializerMock

	err := Init(&dbconfig.Config{
		Connection: dbconfig.ConnectionConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			Schema:   "public",
		},
		ServiceName: "test",
	})
	require.NoError(t, err)
	require.Equal(t, dbClient, DB())
	readDB, err := ReadDB()
	require.EqualError(t, err, "read DB is not initialized")
	require.Nil(t, readDB)
}

func TestInit_WithReadReplica(t *testing.T) {
	defer cleanup()

	primaryClient := orm.NewDB(nil, nil)
	readClient := orm.AsReadDB(orm.NewDB(nil, nil))
	dbInitializerMock := mocks.NewDBInitializer(t)
	dbInitializerMock.On("InitializeDB", mock.MatchedBy(func(conf *dbconfig.Config) bool {
		return conf.Connection.Host == "localhost"
	})).Return(primaryClient, nil)
	dbInitializerMock.On("InitializeReadDB", mock.MatchedBy(func(conf *dbconfig.Config) bool {
		return conf.ReadConnection.Host == "read-host" &&
			conf.ReadConnection.User == "read-user" &&
			conf.ReadConnection.Password == "read-password" &&
			conf.ReadConnection.Port == 5433
	})).Return(readClient, nil)

	dbInit = dbInitializerMock

	err := Init(&dbconfig.Config{
		Connection: dbconfig.ConnectionConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			Schema:   "public",
		},
		ReadConnection: dbconfig.ReadConnectionConfig{
			Host:     "read-host",
			Port:     5433,
			User:     "read-user",
			Password: "read-password",
		},
		ServiceName: "test",
	})
	require.NoError(t, err)
	require.Equal(t, primaryClient, DB())
	readDB, err := ReadDB()
	require.NoError(t, err)
	require.Equal(t, readClient, readDB)
}

func TestInit_ReadReplicaInitFailed(t *testing.T) {
	defer cleanup()

	primaryClient := orm.NewDB(nil, nil)
	dbInitializerMock := mocks.NewDBInitializer(t)
	dbInitializerMock.On("InitializeDB", mock.MatchedBy(func(conf *dbconfig.Config) bool {
		return conf.Connection.Host == "localhost"
	})).Return(primaryClient, nil)
	dbInitializerMock.On("InitializeReadDB", mock.MatchedBy(func(conf *dbconfig.Config) bool {
		return conf.ReadConnection.Host == "read-host"
	})).Return(nil, errors.New("read replica connection failed"))

	dbInit = dbInitializerMock

	err := Init(&dbconfig.Config{
		Connection: dbconfig.ConnectionConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			Schema:   "public",
		},
		ReadConnection: dbconfig.ReadConnectionConfig{
			Host:     "read-host",
			Port:     5433,
			User:     "read-user",
			Password: "read-password",
		},
		ServiceName: "test",
	})
	require.EqualError(t, err, "read replica connection failed")
	require.Equal(t, primaryClient, DB())
	require.Nil(t, readerDB)
	readDB, err := ReadDB()
	require.EqualError(t, err, "read DB is not initialized")
	require.Nil(t, readDB)
}

func TestInit_FailedToInitConfig(t *testing.T) {
	defer cleanup()

	t.Setenv("SECRETS_MOUNT_PATH", "non-existent-path")
	err := Init(nil)
	require.ErrorContains(t, err, "failed to set secrets folder")
}

func TestInit_FailedToInitialize(t *testing.T) {
	defer cleanup()

	dbInitializerMock := mocks.NewDBInitializer(t)
	dbInitializerMock.On("InitializeDB", mock.Anything).Return(nil, errors.New("failed to initialize"))
	dbInit = dbInitializerMock

	err := Init(&dbconfig.Config{
		Connection: dbconfig.ConnectionConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			Schema:   "public",
		},
		ServiceName: "test",
	})
	require.EqualError(t, err, "failed to initialize")
	require.Nil(t, DB())
}

func TestDB(t *testing.T) {
	defer cleanup()

	primaryDB = orm.NewDB(nil, nil)
	require.Equal(t, primaryDB, DB())
}

func TestDB_DBNotInitialized(t *testing.T) {
	defer cleanup()

	primaryDB = nil
	readerDB = nil
	require.Nil(t, DB())
}

func TestReadDB_ReturnsErrorWhenNotInitialized(t *testing.T) {
	defer cleanup()

	primaryDB = orm.NewDB(nil, nil)
	readerDB = nil
	readDB, err := ReadDB()
	require.EqualError(t, err, "read DB is not initialized")
	require.Nil(t, readDB)
}

func TestReadDB_ReturnsReadReplica(t *testing.T) {
	defer cleanup()

	primaryDB = orm.NewDB(nil, nil)
	replica := orm.AsReadDB(orm.NewDB(nil, nil))
	readerDB = replica
	readDB, err := ReadDB()
	require.NoError(t, err)
	require.Same(t, replica, readDB)
}

func TestNewDB(t *testing.T) {
	defer cleanup()

	conf := &dbconfig.Config{
		Connection: dbconfig.ConnectionConfig{Schema: "schema1"},
		Migration:  dbconfig.MigrationConfig{Auto: ptr.Bool(true)},
	}
	err := config.Set(conf)
	require.NoError(t, err)

	dbClient := orm.NewDB(nil, nil)
	dbInitializerMock := mocks.NewDBInitializer(t)
	dbInitializerMock.On("InitializeDB", &dbconfig.Config{
		Connection: dbconfig.ConnectionConfig{Schema: "schema2"},
		Migration:  dbconfig.MigrationConfig{Auto: ptr.Bool(false)},
	}).Return(dbClient, nil)
	dbInit = dbInitializerMock

	db, err := NewDB("schema2")
	require.NoError(t, err)
	require.Equal(t, dbClient, db)

	loadedDB, ok := schemaDBs.Load("schema2")
	require.True(t, ok)
	require.Equal(t, dbClient, loadedDB.(orm.DB))
}

func TestNewDB_LoadsFromSchemaDBs(t *testing.T) {
	defer cleanup()

	dbClient := orm.NewDB(nil, nil)
	schemaDBs.Store("schema1", dbClient)

	db, err := NewDB("schema1")
	require.NoError(t, err)
	require.Equal(t, dbClient, db)
}

func TestNewDB_LoadsFromPrimaryDB(t *testing.T) {
	defer cleanup()

	conf := &dbconfig.Config{
		Connection: dbconfig.ConnectionConfig{Schema: "primary"},
	}
	err := config.Set(conf)
	require.NoError(t, err)

	primaryDB = orm.NewDB(nil, nil)
	db, err := NewDB("primary")
	require.NoError(t, err)
	require.Equal(t, primaryDB, db)
}

func TestMigrate(t *testing.T) {
	defer cleanup()

	conf := &dbconfig.Config{
		Migration: dbconfig.MigrationConfig{
			Auto:          ptr.Bool(false),
			Direction:     dbconfig.MigrationDirectionUp,
			TargetVersion: ptr.String("2"),
		},
		UseServiceSession: ptr.Bool(true),
	}
	err := config.Set(conf)
	require.NoError(t, err)

	primaryDB = orm.NewDB(nil, nil)

	dbClient := orm.NewDB(nil, nil)
	dbInitializerMock := mocks.NewDBInitializer(t)
	dbInitializerMock.On("InitializeDB", &dbconfig.Config{
		Migration: dbconfig.MigrationConfig{
			Auto:          ptr.Bool(true),
			Direction:     dbconfig.MigrationDirectionDown,
			TargetVersion: ptr.String("1"),
		},
		UseServiceSession: ptr.Bool(false),
	}).Return(dbClient, nil)
	dbInit = dbInitializerMock

	err = Migrate(&MigrationParams{
		Direction:     dbconfig.MigrationDirectionDown,
		TargetVersion: ptr.String("1"),
	})
	require.NoError(t, err)
}

func TestMigrate_DBNotInitialized(t *testing.T) {
	defer cleanup()

	err := Migrate(&MigrationParams{
		Direction: dbconfig.MigrationDirectionUp,
	})
	require.EqualError(t, err, "DB is not initialized")
}

func TestMigrate_InvalidParams(t *testing.T) {
	defer cleanup()

	primaryDB = orm.NewDB(nil, nil)

	require.EqualError(t, Migrate(nil), "migration parameters are not valid")
	require.EqualError(t, Migrate(&MigrationParams{}), "migration parameters are not valid")
}

func TestMigrate_Failed(t *testing.T) {
	defer cleanup()

	conf := &dbconfig.Config{
		Connection: dbconfig.ConnectionConfig{Schema: "public"},
		Migration: dbconfig.MigrationConfig{
			Auto:      ptr.Bool(true),
			Direction: dbconfig.MigrationDirectionUp,
		},
	}
	err := config.Set(conf)
	require.NoError(t, err)

	primaryDB = orm.NewDB(nil, nil)

	dbInitializerMock := mocks.NewDBInitializer(t)
	dbInitializerMock.On("InitializeDB", mock.Anything).Return(nil, errors.New("migration error"))
	dbInit = dbInitializerMock

	err = Migrate(&MigrationParams{
		Direction: dbconfig.MigrationDirectionUp,
	})
	require.EqualError(t, err, "failed to migrate: migration error")
}

func TestClose(t *testing.T) {
	defer cleanup()

	dbMock := ormmocks.NewDB(t)
	dbMock.On("Close").Return(nil)
	primaryDB = dbMock
	require.NoError(t, Close())
	require.Nil(t, primaryDB)
	require.Nil(t, readerDB)
}

func TestClose_WithReadReplica(t *testing.T) {
	defer cleanup()

	primaryDBMock := ormmocks.NewDB(t)
	primaryDBMock.On("Close").Return(nil)
	primaryDB = primaryDBMock

	readDBMock := ormmocks.NewDB(t)
	readDBMock.On("Close").Return(nil)
	readerDB = orm.AsReadDB(readDBMock)

	require.NoError(t, Close())
}

func TestClose_Failed(t *testing.T) {
	defer cleanup()

	dbMock := ormmocks.NewDB(t)
	dbMock.On("Close").Return(errors.New("failed to close"))
	primaryDB = dbMock
	require.EqualError(t, Close(), "failed to close primary DB: failed to close")
	require.Nil(t, primaryDB)
}

func TestClose_WithSchemaDBs(t *testing.T) {
	defer cleanup()

	primaryDBMock := ormmocks.NewDB(t)
	primaryDBMock.On("Close").Return(nil)
	primaryDB = primaryDBMock

	schemaDBMock1 := ormmocks.NewDB(t)
	schemaDBMock1.On("Close").Return(nil)
	schemaDBs.Store("schema1", schemaDBMock1)

	schemaDBMock2 := ormmocks.NewDB(t)
	schemaDBMock2.On("Close").Return(nil)
	schemaDBs.Store("schema2", schemaDBMock2)

	require.NoError(t, Close())
}

func TestClose_WithSchemaDBs_Failed(t *testing.T) {
	defer cleanup()

	primaryDBMock := ormmocks.NewDB(t)
	primaryDBMock.On("Close").Return(nil)
	primaryDB = primaryDBMock

	schemaDBMock1 := ormmocks.NewDB(t)
	schemaDBMock1.On("Close").Return(fmt.Errorf("failed to close"))
	schemaDBs.Store("schema1", schemaDBMock1)

	require.EqualError(t, Close(), "failed to close schema DB schema1: failed to close")
}

func TestClose_DBNotInitialized(t *testing.T) {
	defer cleanup()

	require.EqualError(t, Close(), "DB is not initialized")
}

func TestSetLogger(t *testing.T) {
	defer cleanup()

	logger.SetLogger(nil)
	require.NotNil(t, logger.GetLogger())

	myLogger := &myLogger{}
	SetLogger(myLogger)
	require.Equal(t, myLogger, logger.GetLogger())
}

func cleanup() {
	config.FlushAll()
	primaryDB = nil
	readerDB = nil
	schemaDBs = sync.Map{}
	once = sync.Once{}
	dbInit = dbinit.NewInitializer()
}

type myLogger struct {
}

func (l *myLogger) Infof(format string, args ...any) {
}

func (l *myLogger) Errorf(format string, args ...any) {
}

func (l *myLogger) Warnf(format string, args ...any) {
}

func (l *myLogger) Debugf(format string, args ...any) {
}
