/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	clusters_migration "github.com/run-ai/runai/control-plane/cluster-migrator/clusters/migration"
	"github.com/run-ai/runai/control-plane/cluster-migrator/clusters/version_control"
	migrator_config "github.com/run-ai/runai/control-plane/cluster-migrator/config"
	"github.com/run-ai/runai/control-plane/cluster-migrator/database"
	nodepools_migration "github.com/run-ai/runai/control-plane/cluster-migrator/nodepools/migration"
	nodepoolsystemsettings_migration "github.com/run-ai/runai/control-plane/cluster-migrator/nodepoolsystemsettings/migration"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

var (
	conf         *migrator_config.ConfigData
	sourceClient orm.DB
	targetClient orm.DB
)

func init() {
	logging.Init(nil)
	migrator_config.Init()
	conf = config.Get[*migrator_config.ConfigData]()
	runaidb.SetLogger(logging.Actual())
	if err := runaidb.Init(nil); err != nil {
		logging.Fatalf("Failed to initialize database: %v", err)
	}
	var err error
	sourceClient, targetClient, _, err = database.Init(conf)
	if err != nil {
		logging.Fatalf("Failed to initialize database clients: %v", err)
	}
	version_control.Init(conf)
}

func main() {
	defer func() {
		if err := runaidb.Close(); err != nil {
			logging.Errorf("Failed to close database: %v", err)
		}
	}()

	logging.Infof("Starting cluster-migrator job. <<< trigger build")

	if conf.MigrationType == migrator_config.MigrationTypeClusters || conf.MigrationType == migrator_config.MigrationTypeAll {
		logging.Infof("Starting clusters migration.")
		clusters_migration.Migrate(sourceClient, targetClient)
		logging.Infof("Clusters migration completed.")
	}

	if conf.MigrationType == migrator_config.MigrationTypeNodepools || conf.MigrationType == migrator_config.MigrationTypeAll {
		logging.Infof("Starting nodepools migration.")

		nodepools_migration.Migrate(conf)
		logging.Infof("Nodepools migration completed.")
	}

	if conf.MigrationType == migrator_config.MigrationTypeNodepoolSystemSettings || conf.MigrationType == migrator_config.MigrationTypeAll {
		logging.Infof("Starting nodepool system settings migration.")

		hadFailure := nodepoolsystemsettings_migration.Migrate(conf)
		if hadFailure {
			logging.Fatalf("Nodepool-system-settings migration encountered failures.")
			return
		}
		logging.Infof("Nodepool-system-settings migration completed.")
	}

	logging.Infof("Cluster migrator job completed.")
}
