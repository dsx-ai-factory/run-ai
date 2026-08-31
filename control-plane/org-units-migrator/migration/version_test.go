/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/org-units-migrator/database"
	"github.com/run-ai/runai/control-plane/org-units-migrator/database/models"
)

var _ = Describe("Version tests", Ordered, func() {
	It("Tenant happy flow phases", func() {
		tenants := []int32{100, 200, 300}

		for _, tenant := range tenants {
			err := database.SetMigrationStatusForTenant(tenant, models.STARTED)
			Expect(err).To(BeNil())

			err = database.SetMigrationStatusForTenant(tenant, models.COMPLETED)
			Expect(err).To(BeNil())

			err = database.SetMigrationCompleted()
			Expect(err).To(BeNil())
		}
	})

	It("Tenant fail flow phases", func() {
		tenantId := int32(100)

		err := database.SetMigrationStatusForTenant(tenantId, models.STARTED)
		Expect(err).To(BeNil())

		err = database.SetMigrationStatusForTenant(tenantId, models.FAILED)
		Expect(err).To(BeNil())
	})
})
