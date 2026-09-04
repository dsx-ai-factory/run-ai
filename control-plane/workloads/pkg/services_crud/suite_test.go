/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	ginkgoTesting "testing"

	authClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/mocks"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/integdb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"

	k8sObjectsClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/k8s-objects-tracker"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cache_testing "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/testing"
)

func TestCrud(t *ginkgoTesting.T) {
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

	cache_testing.EnableCache()

	k8sObjectsClient.SetMock(mocks.NewK8SObjectTrackerMock())

	authClient.SetMockAuthorizationClient(auth.SimpleMockAuthService{})

	RunSpecs(t, "Api tests")

	t.Cleanup(func() {
		testing.CleanTables()
		cache_testing.DisableCache()
	})
}
