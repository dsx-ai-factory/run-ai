/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	authClient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"testing"

	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"

	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const DB_SCHEMA_FOR_TEST = "workloads_manager"

var authMock = auth.NewSimpleMockAuthService()

func TestCrud(t *testing.T) {
	RegisterFailHandler(Fail)

	dao_utils.SetPostgresEnv("", "file://../database/migrations")

	config.Init()
	RegisterSuspendDefCacheStore()

	conf := dao_utils.ConstructConfig(DB_SCHEMA_FOR_TEST)
	conf.SkipDefaultTransaction = true
	dao_utils.OpenPostgresForTesting(t, conf)

	authClient.SetMockAuthorizationClient(authMock)

	RunSpecs(t, "workloads-manager")

	t.Cleanup(func() {
		cleanTables()
	})
}
