/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	authorization_config "github.com/run-ai/runai/control-plane/authorization/pkg/config"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/integdb"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

type TestingT interface {
	Cleanup(func())
	Fatalf(format string, args ...any)
}

var (
	TestTenant           uint = 1
	ManagementTestTenant uint = 2
	DGXCTestTenant       uint = 3
	CustomRolesEnabled        = true
	ClusterVersion            = "2.21.0"
)

func SetupPostgres(t TestingT, migrationPath string) {
	conf, cleanup, err := integdb.SetupPostgresDB(&integdb.IntegdbConfig{
		Schema:           authorization_config.AuthorizationDBSchema,
		MigrationDirPath: migrationPath,
	})
	if err != nil {
		t.Fatalf("failed to set up test postgres: %v", err)
	}
	if err := runaidb.Init(conf); err != nil {
		cleanup()
		t.Fatalf("failed to initialize database: %v", err)
	}
	t.Cleanup(func() {
		_ = runaidb.Close()
		cleanup()
	})
}

func CreateTables() {}

func CleanTables() {
	_, _ = runaidb.DB().Unscoped().Where("1 = 1").Delete(&models.AccessRule{})
	_, _ = runaidb.DB().Unscoped().Where("1 = 1").Delete(&models.Permission{})
	_, _ = runaidb.DB().Unscoped().Where("1 = 1").Delete(&models.Role{})
}
