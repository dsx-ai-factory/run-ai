/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbinit

import (
	dbconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

type Factory interface {
	CreateConnector() Connector
	CreateMigrator(db orm.DB) Migrator
	CreateSessionCreator(db orm.DB) SessionCreator
}

type Connector interface {
	ConnectDB(conf *dbconfig.Config) (orm.DB, error)
	ConnectReadDB(conf *dbconfig.Config) (orm.ReadDB, error)
}

type Migrator interface {
	Migrate(conf *dbconfig.Config) error
}

type SessionCreator interface {
	Create(user, password, schema string) error
}
