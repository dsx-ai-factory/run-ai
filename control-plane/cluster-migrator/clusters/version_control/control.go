/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package version_control

import (
	migrator_config "github.com/run-ai/runai/control-plane/cluster-migrator/config"
	"github.com/run-ai/runai/control-plane/cluster-migrator/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

var mc *models.ClustersMigrationControl

func Init(conf *migrator_config.ConfigData) {
	mc = &models.ClustersMigrationControl{
		Version: conf.MigrationVersion,
	}
}

func IsMigrationDone(client orm.DB) bool {
	if mc == nil {
		return false
	}

	if err := client.First(mc); err != nil {
		logging.Infof("Failed to get version %s: %v", mc.Version, err)
		return false
	}
	return true
}

func MarkMigrationDone(client orm.DB) {
	if _, err := client.Create(mc); err != nil {
		logging.Errorf("Failed to save version %s: %v", mc.Version, err)
	}
}
