/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	ginkgoTesting "testing"

	"github.com/google/uuid"

	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authorization_config "github.com/run-ai/runai/control-plane/authorization/pkg/config"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/dao"
	testing "github.com/run-ai/runai/control-plane/authorization/pkg/test"
	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"
	cacheTesting "github.com/run-ai/runai/runai-common-packages/cache/v2/testing"
)

var (
	RolesMap          = map[string]uint{}
	PermissionSetsMap = map[string]uuid.UUID{}
)

func TestApi(t *ginkgoTesting.T) {
	RegisterFailHandler(Fail)

	testing.SetupPostgres(t, "file://../database/migrations")
	if err := authorization_config.Init(); err != nil {
		t.Fatalf("failed to initialize config: %v", err)
	}
	dao.InitStaticPermissionSets()
	dao.InitStaticLegacyRoles()
	dao.InitStaticRoles()
	dao.InitAccessRulesForTests(testing.TestTenant, testing.ManagementTestTenant)
	dao.InitRolesMap(&RolesMap)
	dao.InitPermissionSetsMap(&PermissionSetsMap)

	SetMockBackendServer()

	cacheTesting.EnableCache()

	checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
		WithCluster("", buildDisplayedCluster("", "2.24.0")).
		WithCluster(ClusterId, buildDisplayedCluster(ClusterId, "2.22.0")).
		WithCluster(TestCluster1Id, buildDisplayedCluster(TestCluster1Id, "2.22.0")).
		WithCluster(TestCluster2Id, buildDisplayedCluster(TestCluster2Id, "2.21.0")))
	checker.InitClusterCapabilitiesChecker("", getDefaultClustersCacheConfig(), "")

	RunSpecs(t, "Api tests")
	t.Cleanup(func() {
		testing.CleanTables()
		cache.ClearCache()
		cache.Close()
	})
}

func getDefaultClustersCacheConfig() *checker.ClustersCacheConfig {
	return &checker.ClustersCacheConfig{ExpirationSec: 600, CleanupSec: 10}
}

func buildDisplayedCluster(clusterId string, version string) openapi.DisplayedCluster {
	result := openapi.DisplayedCluster{
		Uuid: clusterId,
	}
	result.Version.Set(&version)
	return result
}
