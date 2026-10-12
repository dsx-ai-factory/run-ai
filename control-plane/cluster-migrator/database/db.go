/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"fmt"

	migrator_config "github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

var (
	conf         *migrator_config.ConfigData
	sourceClient orm.DB
	globalClient orm.DB
	targetClient orm.DB
)

func Init(c *migrator_config.ConfigData) error {
	conf = c
	targetClient = runaidb.DB()
	if targetClient == nil {
		return fmt.Errorf("database is not initialized")
	}

	var err error
	sourceClient, err = runaidb.NewDB(conf.SchemaSource)
	if err != nil {
		return err
	}
	globalClient, err = runaidb.NewDB("")
	if err != nil {
		return err
	}

	return autoMigrates()
}

func autoMigrates() error {
	err := targetClient.Actual().AutoMigrate(&models.NodepoolSystemSettingsMigrationControl{})
	if err != nil {
		logging.Errorf("Failed to run nodepool system settings migration control auto-migrate: %v", err)
		return err
	}

	return nil
}
