/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_utils

import (
	ginkgoTesting "testing"

	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestServicesUtils(t *ginkgoTesting.T) {
	RegisterFailHandler(Fail)

	dao_utils.SetPostgresEnv("", "file://../database/migrations")
	config.Init()
	conf := config.Database()
	conf.SkipDefaultTransaction = true
	dao_utils.OpenPostgresForTesting(t, config.Database())

	RunSpecs(t, "services_utils tests")

	t.Cleanup(func() {
		testing.CleanTables()
	})
}
