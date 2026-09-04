/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_utils

import (
	ginkgoTesting "testing"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/integdb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestServicesUtils(t *ginkgoTesting.T) {
	RegisterFailHandler(Fail)

	config.Init()
	conf, cleanup, err := integdb.SetupPostgresDB(&integdb.IntegdbConfig{
		Schema:           "workloads",
		MigrationDirPath: "file://../database/migrations",
	})
	if err != nil {
		t.Fatalf("failed to set up test postgres: %v", err)
	}
	t.Cleanup(cleanup)
	conf.UseServiceSession = new(bool)
	conf.SkipDefaultTransaction = true
	if err := runaidb.Init(conf); err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	t.Cleanup(func() { _ = runaidb.Close() })

	RunSpecs(t, "services_utils tests")

	t.Cleanup(func() {
		testing.CleanTables()
	})
}
