/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	ginkgoTesting "testing"

	"github.com/google/uuid"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authorization_config "github.com/run-ai/runai/control-plane/authorization/pkg/config"
	testing "github.com/run-ai/runai/control-plane/authorization/pkg/test"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
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
