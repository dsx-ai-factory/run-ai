/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	ginkgoTesting "testing"

	service_config "github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/database"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/integdb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestApi(t *ginkgoTesting.T) {
	RegisterFailHandler(Fail)

	service_config.Init()
	dbConf, cleanup, err := integdb.SetupPostgresDB(&integdb.IntegdbConfig{
		Schema:           "org_unit_service",
		MigrationDirPath: "file://../../org-unit-service/pkg/database/migrations",
	})
	require.NoError(t, err)
	t.Cleanup(cleanup)
	dbConf.UseServiceSession = new(bool)
	dbConf.Migration.Auto = ptr.To(true)

	runaidb.SetLogger(logging.Actual())
	require.NoError(t, runaidb.Init(dbConf))
	t.Cleanup(func() { _ = runaidb.Close() })

	database.Init()
	database.SetupTables()

	RunSpecs(t, "Migration tests")
	t.Cleanup(func() {
		database.CleanTables()
	})
}
