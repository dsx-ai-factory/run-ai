/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"fmt"
	ginkgoTesting "testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/go_migrations"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/integdb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"

	authClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	intervalMillisWhenWaitingForView = 20
	IntervalWhenWaitingForView       = intervalMillisWhenWaitingForView * time.Millisecond
	timeoutMillisWhenWaitingForView  = 5000
	TimeoutWhenWaitingForView        = timeoutMillisWhenWaitingForView * time.Millisecond
)

func TestApi(t *ginkgoTesting.T) {
	RegisterFailHandler(Fail)

	config.Init()
	mockConfigForWorkloads()
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

	err = go_migrations.Migrate()
	if err != nil {
		t.Fatalf("Failed to run go DB migrations: %v", err)
	}

	authClient.SetMockAuthorizationClient(&testing.SimpleMockAuthService{})

	stopper := make(chan struct{})
	defer func() {
		close(stopper)
	}()
	refreshMaterializedView(stopper)

	RunSpecs(t, fmt.Sprintf("Api tests, db port: %d", conf.Connection.Port))
	t.Cleanup(func() {
		testing.CleanTables()
	})
}

func refreshMaterializedView(stopCh <-chan struct{}) {
	database.InitRefresher(stopCh, runaidb.DB(), runai_config.Get[*config.ServerConfig]())
}

func mockConfigForWorkloads() {
	runai_config.Set(&config.ServerConfig{
		ViewRefreshRateMilliseconds: 20,
	})
}
