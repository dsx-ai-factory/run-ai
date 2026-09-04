/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"os"

	service_config "github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/migration"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

func init() {
	logging.Init(nil)
	service_config.Init()
	database.Init()
}

func main() {
	logging.Infof("Starting org-units migration job.")
	conf := config.Get[*service_config.ConfigData]()
	if database.IsMigrationCompleted() {
		logging.Infof("Migrate done already, finishing.")
		return
	}

	tenantsToMigrate := database.GetTenantsToMigrate(conf.TenantsToMigrate, conf.TenantsToNotMigrate)

	hadFailure := migration.Migrate(conf, tenantsToMigrate)

	if !hadFailure {
		database.SetMigrationCompleted()
	}

	os.Exit(0)
}
