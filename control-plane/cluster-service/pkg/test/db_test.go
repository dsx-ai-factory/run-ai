/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/testutils"
	cache_testing "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/testing"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
)

func TestDao(t *testing.T) {
	RegisterFailHandler(Fail)
	defer runai_config.FlushAll()

	cache_testing.EnableCache()
	// No table cleanup here: RegisterPostgresSuite closes the connection and drops
	// the container when the suite ends, so runaidb.DB() is nil by this point.
	t.Cleanup(cache_testing.DisableCache)

	testutils.RegisterPostgresSuite(config.Schema, "file://../database/migrations")

	RunSpecs(t, "DB tests")
}

// CleanTables is called from specs, while the suite's connection is still open.
func CleanTables() {
	db := runaidb.DB().WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true})
	for _, model := range []any{&models.NodeNuma{}, &models.Node{}, &models.Cluster{}, &models.ClusterHistory{}} {
		_, err := db.Delete(model)
		Expect(err).NotTo(HaveOccurred())
	}
}
