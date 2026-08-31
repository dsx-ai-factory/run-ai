/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	service_config "github.com/run-ai/runai/control-plane/org-units-migrator/config"
	"github.com/run-ai/runai/control-plane/org-units-migrator/database"
	"github.com/run-ai/runai/runai-common-packages/config"
)

var _ = Describe("Migration tests", Ordered, func() {
	It("Migrate tenants", func() {
		service_config.Init()
		tenants := database.GetTenantsToMigrate([]string{}, []string{})
		hadError := Migrate(config.Get[*service_config.ConfigData](), tenants)
		Expect(hadError).To(BeFalse())

		for _, tenant := range tenants {
			setting, err := database.GetTenantSettingsMap(tenant.Id)
			Expect(err).To(BeNil())

			validateDepartmentsMigration(tenant, setting)
			validateNodeTypesMigration(tenant)
			validateProjectsMigration(tenant, setting)
		}
	})
})
