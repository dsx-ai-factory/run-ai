/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"fmt"
	ginkgoTesting "testing"
	"time"

	"github.com/run-ai/runai/control-plane/workloads/pkg/database/go_migrations"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"

	authClient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database"
	testing "github.com/run-ai/runai/control-plane/workloads/pkg/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	intervalMillisWhenWaitingForView = 20
	IntervalWhenWaitingForView       = intervalMillisWhenWaitingForView * time.Millisecond
	timeoutMillisWhenWaitingForView  = 200
	TimeoutWhenWaitingForView        = timeoutMillisWhenWaitingForView * time.Millisecond
)

func TestApi(t *ginkgoTesting.T) {
	RegisterFailHandler(Fail)

	dao_utils.SetPostgresEnv("", "file://../database/migrations")
	config.Init()
	mockConfigForWorkloads()
	conf := config.Database()
	conf.SkipDefaultTransaction = true
	dao_utils.OpenPostgresForTesting(t, conf)

	err := go_migrations.Migrate()
	if err != nil {
		t.Fatalf("Failed to run go DB migrations: %v", err)
	}

	authClient.SetMockAuthorizationClient(&testing.SimpleMockAuthService{})

	stopper := make(chan struct{})
	defer func() {
		close(stopper)
	}()
	refreshMaterializedView(stopper)

	RunSpecs(t, fmt.Sprintf("Api tests, db port: %s", conf.Port))
	t.Cleanup(func() {
		testing.CleanTables()
	})
}

func refreshMaterializedView(stopCh <-chan struct{}) {
	database.InitRefresher(stopCh, dao_utils.DB(), runai_config.Get[*config.ServerConfig]())
}

func mockConfigForWorkloads() {
	runai_config.Set(&config.ServerConfig{
		ViewRefreshRateMilliseconds: 20,
	})
}
