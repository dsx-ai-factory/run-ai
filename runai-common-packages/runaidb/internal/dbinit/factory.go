/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbinit

import (
	"github.com/run-ai/runai/runai-common-packages/runaidb/internal/postgres"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

type postgresFactory struct{}

func (f *postgresFactory) CreateConnector() Connector {
	return postgres.NewConnector()
}

func (f *postgresFactory) CreateMigrator(db orm.DB) Migrator {
	return postgres.NewMigrator(db)
}

func (f *postgresFactory) CreateSessionCreator(db orm.DB) SessionCreator {
	return postgres.NewRWSessionCreator(db)
}
