/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	ginkgoTesting "testing"

	"github.com/google/uuid"

	authorization_config "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/config"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var RolesMap = map[string]uint{}
var PermissionSetsMap = map[string]uuid.UUID{}

func TestDao(t *ginkgoTesting.T) {
	RegisterFailHandler(Fail)

	testing.SetupPostgres(t, "file://../migrations")
	if err := authorization_config.Init(); err != nil {
		t.Fatalf("failed to initialize config: %v", err)
	}
	InitStaticPermissionSets()
	InitStaticLegacyRoles()
	InitStaticRoles()
	InitAccessRulesForTests(testing.TestTenant, testing.ManagementTestTenant)
	InitRolesMap(&RolesMap)
	InitPermissionSetsMap(&PermissionSetsMap)

	RunSpecs(t, "Dao tests")
	t.Cleanup(func() {
		testing.CleanTables()
		cache.ClearCache()
		cache.Close()
	})
}
