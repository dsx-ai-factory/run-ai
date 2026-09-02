/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	ginkgoTesting "testing"

	service_config "github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/database"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestApi(t *ginkgoTesting.T) {
	RegisterFailHandler(Fail)

	dao_utils.SetPostgresEnv("", "file://../../org-unit-service/pkg/database/migrations")
	service_config.Init()
	dbConfig := config.Get[*dao_utils.DataBaseConfig]()
	dbConfig.MigrationCommand = ""
	dao_utils.OpenPostgresForTesting(t, dbConfig)
	database.Init()
	database.SetupTables()

	RunSpecs(t, "Migration tests")
	t.Cleanup(func() {
		database.CleanTables()
	})
}
