/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbinit_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/smithy-go/ptr"
	dbconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/internal/dbinit"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/internal/dbinit/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/internal/logger"
	loggermocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/internal/logger/mocks"
	pgerrors "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/internal/postgres/errors"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func defaultConfig() *dbconfig.Config {
	return &dbconfig.Config{
		Connection: dbconfig.ConnectionConfig{
			User:     "admin",
			Password: "secret",
			Schema:   "public",
		},
		Migration: dbconfig.MigrationConfig{
			Direction: dbconfig.MigrationDirectionUp,
		},
		ServiceName: "myservice",
	}
}

func TestNewInitializer(t *testing.T) {
	init := dbinit.NewInitializer()
	require.NotNil(t, init)
}

func TestInitialize(t *testing.T) {
	adminDB := orm.NewDB(nil, nil)
	conf := defaultConfig()

	conn := mocks.NewConnector(t)
	conn.On("ConnectDB", conf).Return(adminDB, nil)

	factory := mocks.NewFactory(t)
	factory.On("CreateConnector", mock.Anything).Return(conn)

	init := &dbinit.Initializer{Factory: factory}

	db, err := init.InitializeDB(conf)
	require.NoError(t, err)
	require.Equal(t, adminDB, db)
}

func TestInitializeReadDB(t *testing.T) {
	readDB := orm.AsReadDB(orm.NewDB(nil, nil))
	conf := defaultConfig()
	conf.ReadConnection = dbconfig.ReadConnectionConfig{
		Host:     "read-host",
		User:     "read-user",
		Password: "read-password",
	}

	conn := mocks.NewConnector(t)
	conn.On("ConnectReadDB", conf).Return(readDB, nil)

	factory := mocks.NewFactory(t)
	factory.On("CreateConnector", mock.Anything).Return(conn)

	init := &dbinit.Initializer{Factory: factory}

	db, err := init.InitializeReadDB(conf)
	require.NoError(t, err)
	require.Equal(t, readDB, db)
}

func TestInitializeReadDB_ConnectFailed(t *testing.T) {
	conf := defaultConfig()
	conf.ReadConnection = dbconfig.ReadConnectionConfig{
		Host:     "read-host",
		User:     "read-user",
		Password: "read-password",
	}

	conn := mocks.NewConnector(t)
	conn.On("ConnectReadDB", conf).Return(nil, errors.New("read connection refused"))

	factory := mocks.NewFactory(t)
	factory.On("CreateConnector", mock.Anything).Return(conn)

	init := &dbinit.Initializer{Factory: factory}

	db, err := init.InitializeReadDB(conf)
	require.Nil(t, db)
	require.EqualError(t, err, "read connection refused")
}

func TestInitialize_ConnectFailed(t *testing.T) {
	conf := defaultConfig()

	conn := mocks.NewConnector(t)
	conn.On("ConnectDB", conf).Return(nil, errors.New("connection refused"))

	factory := mocks.NewFactory(t)
	factory.On("CreateConnector", mock.Anything).Return(conn)

	init := &dbinit.Initializer{Factory: factory}

	db, err := init.InitializeDB(conf)
	require.Nil(t, db)
	require.EqualError(t, err, "connection refused")
}

func TestInitialize_WithAutoMigrate(t *testing.T) {
	adminDB := orm.NewDB(nil, nil)
	conf := defaultConfig()
	conf.Migration.Auto = ptr.Bool(true)

	conn := mocks.NewConnector(t)
	conn.On("ConnectDB", conf).Return(adminDB, nil)

	mig := mocks.NewMigrator(t)
	mig.On("Migrate", conf).Return(nil)

	factory := mocks.NewFactory(t)
	factory.On("CreateConnector", mock.Anything).Return(conn)
	factory.On("CreateMigrator", adminDB, mock.Anything).Return(mig)

	init := &dbinit.Initializer{Factory: factory}

	db, err := init.InitializeDB(conf)
	require.NoError(t, err)
	require.Equal(t, adminDB, db)
}

func TestInitialize_MigrateFailed(t *testing.T) {
	adminDB := orm.NewDB(nil, nil)
	conf := defaultConfig()
	conf.Migration.Auto = ptr.Bool(true)

	conn := mocks.NewConnector(t)
	conn.On("ConnectDB", conf).Return(adminDB, nil)

	mig := mocks.NewMigrator(t)
	mig.On("Migrate", conf).Return(errors.New("migration failed"))

	factory := mocks.NewFactory(t)
	factory.On("CreateConnector", mock.Anything).Return(conn)
	factory.On("CreateMigrator", adminDB, mock.Anything).Return(mig)

	init := &dbinit.Initializer{Factory: factory}

	db, err := init.InitializeDB(conf)
	require.Nil(t, db)
	require.EqualError(t, err, "migration failed")
}

func TestInitialize_WithServiceSession(t *testing.T) {
	adminDB := orm.NewDB(nil, nil)
	serviceDB := orm.NewDB(nil, nil)
	conf := defaultConfig()
	conf.UseServiceSession = ptr.Bool(true)

	serviceConf := *conf
	serviceConf.Connection.User = conf.ServiceName

	conn := mocks.NewConnector(t)
	conn.On("ConnectDB", conf).Return(adminDB, nil)
	conn.On("ConnectDB", &serviceConf).Return(serviceDB, nil)

	sc := mocks.NewSessionCreator(t)
	sc.On("Create", "myservice", "secret", "public").Return(nil)

	factory := mocks.NewFactory(t)
	factory.On("CreateConnector", mock.Anything).Return(conn)
	factory.On("CreateSessionCreator", adminDB).Return(sc)

	init := &dbinit.Initializer{Factory: factory}

	db, err := init.InitializeDB(conf)
	require.NoError(t, err)
	require.Equal(t, serviceDB, db)
}

func TestInitialize_ServiceSessionCreationFailed(t *testing.T) {
	adminDB := orm.NewDB(nil, nil)
	conf := defaultConfig()
	conf.UseServiceSession = ptr.Bool(true)

	conn := mocks.NewConnector(t)
	conn.On("ConnectDB", mock.Anything).Return(adminDB, nil)

	sc := mocks.NewSessionCreator(t)
	sc.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("grant failed"))

	factory := mocks.NewFactory(t)
	factory.On("CreateConnector", mock.Anything).Return(conn)
	factory.On("CreateSessionCreator", adminDB).Return(sc)

	init := &dbinit.Initializer{Factory: factory}

	db, err := init.InitializeDB(conf)
	require.Nil(t, db)
	require.EqualError(t, err, "grant failed")
}

func TestInitialize_ServiceSessionCreation_InsufficientPrivilege(t *testing.T) {
	adminDB := orm.NewDB(nil, nil)
	conf := defaultConfig()
	conf.UseServiceSession = ptr.Bool(true)

	insufficientPrivilegeErr := &pgconn.PgError{Code: pgerrors.ErrCodeInsufficientPrivilege, Message: "insufficient privilege"}

	loggerMock := loggermocks.NewLogger(t)
	loggerMock.On("Warnf", "insufficient database privileges to create service user %q: %v, defaulting to configured user", conf.ServiceName, insufficientPrivilegeErr).Return()
	logger.SetLogger(loggerMock)
	t.Cleanup(func() {
		logger.SetLogger(logger.NewDummyLogger())
	})

	conn := mocks.NewConnector(t)
	conn.On("ConnectDB", mock.Anything).Return(adminDB, nil)

	sc := mocks.NewSessionCreator(t)
	sc.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(insufficientPrivilegeErr)

	factory := mocks.NewFactory(t)
	factory.On("CreateConnector", mock.Anything).Return(conn)
	factory.On("CreateSessionCreator", adminDB).Return(sc)

	init := &dbinit.Initializer{Factory: factory}

	db, err := init.InitializeDB(conf)
	require.Equal(t, adminDB, db)
	require.NoError(t, err)
}

func TestInitialize_ServiceConnectFailed(t *testing.T) {
	adminDB := orm.NewDB(nil, nil)
	conf := defaultConfig()
	conf.UseServiceSession = ptr.Bool(true)

	serviceConf := *conf
	serviceConf.Connection.User = conf.ServiceName

	conn := mocks.NewConnector(t)
	conn.On("ConnectDB", conf).Return(adminDB, nil)
	conn.On("ConnectDB", &serviceConf).Return(nil, errors.New("service connect failed"))

	sc := mocks.NewSessionCreator(t)
	sc.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	factory := mocks.NewFactory(t)
	factory.On("CreateConnector", mock.Anything).Return(conn)
	factory.On("CreateSessionCreator", adminDB).Return(sc)

	init := &dbinit.Initializer{Factory: factory}

	db, err := init.InitializeDB(conf)
	require.Nil(t, db)
	require.EqualError(t, err, "service connect failed")
}

func TestInitialize_FullFlow(t *testing.T) {
	adminDB := orm.NewDB(nil, nil)
	serviceDB := orm.NewDB(nil, nil)
	conf := defaultConfig()
	conf.Migration.Auto = ptr.Bool(true)
	conf.UseServiceSession = ptr.Bool(true)

	serviceConf := *conf
	serviceConf.Connection.User = conf.ServiceName

	conn := mocks.NewConnector(t)
	conn.On("ConnectDB", conf).Return(adminDB, nil)
	conn.On("ConnectDB", &serviceConf).Return(serviceDB, nil)

	mig := mocks.NewMigrator(t)
	mig.On("Migrate", conf).Return(nil)

	sc := mocks.NewSessionCreator(t)
	sc.On("Create", "myservice", "secret", "public").Return(nil)

	factory := mocks.NewFactory(t)
	factory.On("CreateConnector", mock.Anything).Return(conn)
	factory.On("CreateMigrator", adminDB, mock.Anything).Return(mig)
	factory.On("CreateSessionCreator", adminDB).Return(sc)

	init := &dbinit.Initializer{Factory: factory}

	db, err := init.InitializeDB(conf)
	require.NoError(t, err)
	require.Equal(t, serviceDB, db)
}

func TestInitialize_ServiceSessionCreation_LockUnavailable(t *testing.T) {
	adminDB := orm.NewDB(nil, nil)
	conf := defaultConfig()
	conf.UseServiceSession = ptr.Bool(true)

	lockErr := fmt.Errorf("service user lock for schema %q is already held", "public")

	conn := mocks.NewConnector(t)
	conn.On("ConnectDB", mock.Anything).Return(adminDB, nil)

	sc := mocks.NewSessionCreator(t)
	sc.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(lockErr)

	factory := mocks.NewFactory(t)
	factory.On("CreateConnector", mock.Anything).Return(conn)
	factory.On("CreateSessionCreator", adminDB).Return(sc)

	init := &dbinit.Initializer{Factory: factory}

	db, err := init.InitializeDB(conf)
	require.Nil(t, db)
	require.ErrorContains(t, err, "is already held")
}
