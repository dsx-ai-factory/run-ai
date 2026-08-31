/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"github.com/run-ai/runai/runai-common-packages/config"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

type ConfigData struct {
	SchemaSource        string   `mapstructure:"db_schema_source"`
	SchemaTarget        string   `mapstructure:"db_schema_target"`
	MigrationVersion    string   `mapstructure:"db_migration_version"`
	ShouldMarkCompleted bool     `mapstructure:"should_mark_completed"`
	TenantsToMigrate    []string `mapstructure:"tenants_to_migrate"`
	TenantsToNotMigrate []string `mapstructure:"tenants_to_not_migrate"`
}

func (c *ConfigData) SetDefaults() {
	c.SchemaSource = "backend"
	c.SchemaTarget = "org_unit_service"
	c.MigrationVersion = "v1"
	c.ShouldMarkCompleted = false
	c.TenantsToMigrate = []string{}
	c.TenantsToNotMigrate = []string{}
}

func Init() {
	logging.Debugf("initializing config")
	if err := config.Register[*ConfigData](); err != nil {
		logging.Fatalf("failed to register config: %v", err)
	}
	dbConfig := dao_utils.ConstructConfig(config.Get[*ConfigData]().SchemaTarget)
	dbConfig.MigrationCommand = "None"
	config.Set(dbConfig)
}
