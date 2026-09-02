/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbinit

import (
	dbconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/internal/logger"
	pgerrors "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/internal/postgres/errors"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

type Initializer struct {
	Factory Factory
}

func NewInitializer() *Initializer {
	return &Initializer{
		Factory: &postgresFactory{},
	}
}

func (i *Initializer) InitializeDB(conf *dbconfig.Config) (orm.DB, error) {
	conn := i.Factory.CreateConnector()

	adminDB, err := conn.ConnectDB(conf)
	if err != nil {
		return nil, err
	}

	if conf.Migration.Auto != nil && *conf.Migration.Auto {
		if err := i.migrate(conf, adminDB); err != nil {
			return nil, err
		}
	}

	if conf.UseServiceSession != nil && *conf.UseServiceSession {
		return i.initServiceSession(conf, adminDB, conn)
	}

	return adminDB, nil
}

func (i *Initializer) InitializeReadDB(conf *dbconfig.Config) (orm.ReadDB, error) {
	conn := i.Factory.CreateConnector()

	adminDB, err := conn.ConnectReadDB(conf)
	if err != nil {
		return nil, err
	}

	return adminDB, nil
}

func (i *Initializer) migrate(conf *dbconfig.Config, adminDB orm.DB) error {
	migrator := i.Factory.CreateMigrator(adminDB)
	return migrator.Migrate(conf)
}

func (i *Initializer) initServiceSession(
	conf *dbconfig.Config,
	adminDB orm.DB,
	conn Connector,
) (orm.DB, error) {
	sessionCreator := i.Factory.CreateSessionCreator(adminDB)

	if err := sessionCreator.Create(conf.ServiceName, conf.Connection.Password, conf.Connection.Schema); err != nil {
		if pgerrors.HasCode(err, pgerrors.ErrCodeInsufficientPrivilege) {
			logger.GetLogger().Warnf("insufficient database privileges to create service user %q: %v, defaulting to configured user", conf.ServiceName, err)
			return adminDB, nil
		}
		return nil, err
	}

	serviceConf := *conf
	serviceConf.Connection.User = conf.ServiceName
	serviceDB, err := conn.ConnectDB(&serviceConf)
	defer func() {
		_ = adminDB.Close()
	}()

	if err != nil {
		return nil, err
	}

	return serviceDB, nil
}
