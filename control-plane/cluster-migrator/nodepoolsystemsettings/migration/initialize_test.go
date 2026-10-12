/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	ginkgoTesting "testing"

	migrator_config "github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator/database"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
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

	migrator_config.Init()
	conf := config.Get[*migrator_config.ConfigData]()
	dbConf, cleanup, err := integdb.SetupPostgresDB(&integdb.IntegdbConfig{
		Schema:           conf.SchemaTarget,
		MigrationDirPath: "file://../../../cluster-service/pkg/database/migrations",
	})
	require.NoError(t, err)
	t.Cleanup(cleanup)
	dbConf.UseServiceSession = ptr.To(false)

	runaidb.SetLogger(logging.Actual())
	require.NoError(t, runaidb.Init(dbConf))
	t.Cleanup(func() { _ = runaidb.Close() })

	conf.ShouldMarkCompleted = true
	require.NoError(t, database.Init(conf))
	database.SetupTables(conf)

	RunSpecs(t, "Migration tests")
}
