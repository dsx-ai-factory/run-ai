/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"database/sql"

	service_config "github.com/run-ai/runai/control-plane/authorization-migrator/config"
	"github.com/run-ai/runai/control-plane/authorization-migrator/keycloak"
	"github.com/run-ai/runai/control-plane/authorization-migrator/migration"
	"github.com/run-ai/runai/control-plane/authorization-migrator/version_control"
	config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/metrics"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

var (
	conf         *service_config.ConfigData
	targetClient orm.DB
)

func init() {
	logging.Init(nil)
	metrics.Init(nil)
	service_config.Init()
	conf = config.Get[*service_config.ConfigData]()
	keycloak.Init(conf)

	runaidb.SetLogger(logging.Actual())
	if err := runaidb.Init(nil); err != nil {
		logging.Fatalf("failed to initialize database: %v", err)
	}

	var err error
	targetClient, err = runaidb.NewDB(conf.SchemaTarget)
	if err != nil {
		logging.Fatalf("failed to open target schema %s: %v", conf.SchemaTarget, err)
	}
	version_control.Init(conf, targetClient)
}

func main() {
	logging.Infof("Starting authorization migration job. <<< trigger build")
	defer func() {
		if err := runaidb.Close(); err != nil {
			logging.Errorf("failed to close database connections: %v", err)
		}
	}()

	if version_control.IsMigrationDone(targetClient) {
		logging.Infof("Migrate done already, finishing.")
		return
	}

	sourceClientBackend := openSourceSchemaOrPanic(conf.SchemaSourceBackend)
	sourceClientTenants := openSourceSchemaOrPanic(conf.SchemaSourceTenants)

	ticker := keycloak.SetRefreshTokenTicker()
	defer ticker.Stop()

	migration.Migrate(sourceClientBackend, sourceClientTenants, targetClient)

	version_control.MarkMigrationDone(targetClient)
}

func openSourceSchemaOrPanic(schema string) *sql.DB {
	db, err := runaidb.NewDB(schema)
	if err != nil {
		logging.Fatalf("failed to open source schema %s: %v", schema, err)
	}
	client, err := db.Actual().DB()
	if err != nil {
		logging.Errorf("Failed to start the source schema %s connection", schema)
		panic(err)
	}
	return client
}
