/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/testutils"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authorization_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	authorizationclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	cache_testing "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/testing"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
)

var clusterNameGenerator = NewClusterNameGenerator()

// Reset the process-global setting manager after every spec so a mock set by one spec
// (e.g. settings/metrics tests) cannot leak into others. With it unset, setting_pkg.Get
// returns an error and the feature-flag helpers safely default to "off".
var _ = AfterEach(func() {
	setting_pkg.SetMock(nil)
})

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)
	defer runai_config.FlushAll()

	cache_testing.EnableCache()
	defer cache_testing.DisableCache()

	// Register config types needed for tests
	_ = runai_config.Register[*authorization_client.Config]()

	testutils.RegisterPostgresSuite(config.Schema, "file://../database/migrations")

	authConfig := runai_config.Get[*authorizationclient.Config]()

	authorization_client.Init(authConfig)

	RunSpecs(t, "Services Suite")
}

type ClusterNameGenerator struct {
	counterFn func() int
}

func NewClusterNameGenerator() *ClusterNameGenerator {
	counter := 0
	counterFn := func() int {
		counter++
		return counter
	}
	return &ClusterNameGenerator{counterFn: counterFn}
}

func (c *ClusterNameGenerator) Generate(testSuite string) string {
	return fmt.Sprintf("cluster-%s-%d", testSuite, c.counterFn())
}

func addAuditToCtx(ctx context.Context) context.Context {
	req, _ := http.NewRequest(http.MethodPost, "http://whatever.com", nil)
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "1.1.1.1"

	al := audit.NewAuditLog().WithRequest(req).WithCtx(ctx)
	return audit.CtxWithAudit(ctx, al)
}
