/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbtesting

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/integdb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"gorm.io/gorm"
)

func CreateTables() func() {
	config.MockConfig()
	config.Init()

	conf, cleanup, err := integdb.SetupPostgresDB(&integdb.IntegdbConfig{
		Schema:           "k8s_objects_tracker",
		MigrationDirPath: "file://../database/migrations",
	})
	if err != nil {
		logging.Fatalf("Failed to set up postgres database for testing: %v", err)
	}

	conf.UseServiceSession = new(bool)
	if err := runaidb.Init(conf); err != nil {
		logging.Fatalf("Failed to initialize database for testing: %v", err)
	}

	return cleanup
}

func CleanTables() {
	db := runaidb.DB().Actual()
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Pvc{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Secret{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.ConfigMap{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Object{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.ResourceIssue{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.StorageClassConfiguration{})
}
