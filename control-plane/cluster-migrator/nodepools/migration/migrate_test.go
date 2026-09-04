/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	migrator_config "github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator/database"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Migration tests", Ordered, func() {
	It("Migrate tenants", func() {
		tenants := database.GetTenantsToMigrate([]string{}, []string{})
		hadError := migrateTenants(config.Get[*migrator_config.ConfigData](), tenants)
		Expect(hadError).To(BeFalse())

		for _, tenant := range tenants {
			validateNodepoolsMigration(tenant)
		}
	})
})
