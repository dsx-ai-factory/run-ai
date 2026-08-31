/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"context"
	"testing"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/config"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/runai-common-packages/integdb"
	"github.com/run-ai/runai/runai-common-packages/runaidb"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDao(t *testing.T) {
	config.MockConfig()
	config.Init()
	conf, cleanup, setupErr := integdb.SetupPostgresDB(&integdb.IntegdbConfig{Schema: "assets", MigrationDirPath: "file://../migrations"})
	if setupErr != nil {
		t.Fatalf("failed to set up test postgres: %v", setupErr)
	}
	t.Cleanup(cleanup)
	if initErr := runaidb.Init(conf); initErr != nil {
		t.Fatalf("failed to init database: %v", initErr)
	}
	t.Cleanup(func() { _ = runaidb.Close() })

	RegisterFailHandler(Fail)

	db := runaidb.DB().Actual().WithContext(context.Background())

	db.AutoMigrate(&dbmodel.WorkloadSupportedTypesRec{})
	db.AutoMigrate(&dbmodel.AssetMetaRec{})
	db.AutoMigrate(&dbmodel.ComputeAssetRec{})
	db.AutoMigrate(&dbmodel.EnvironmentAssetRec{})
	db.AutoMigrate(&dbmodel.AssetNameLockRecord{})

	RunSpecs(t, "Dao tests")
}
