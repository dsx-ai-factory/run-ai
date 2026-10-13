/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	migrator_config "github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator/database"
	nodepoolsystemsettings_migration "github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator/nodepoolsystemsettings/migration"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

var conf *migrator_config.ConfigData

func init() {
	logging.Init(nil)
	migrator_config.Init()
	conf = config.Get[*migrator_config.ConfigData]()
	runaidb.SetLogger(logging.Actual())
	if err := runaidb.Init(nil); err != nil {
		logging.Fatalf("Failed to initialize database: %v", err)
	}
	if err := database.Init(conf); err != nil {
		logging.Fatalf("Failed to initialize database clients: %v", err)
	}
}

func main() {
	defer func() {
		if err := runaidb.Close(); err != nil {
			logging.Errorf("Failed to close database: %v", err)
		}
	}()

	logging.Infof("Starting cluster-migrator job. <<< trigger build")

	logging.Infof("Starting nodepool system settings migration.")
	hadFailure := nodepoolsystemsettings_migration.Migrate(conf)
	if hadFailure {
		logging.Fatalf("Nodepool-system-settings migration encountered failures.")
		return
	}
	logging.Infof("Nodepool-system-settings migration completed.")

	logging.Infof("Cluster migrator job completed.")
}
