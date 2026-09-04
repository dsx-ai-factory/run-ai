/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/database"
)

var _ = Describe("Tenants tests", Ordered, func() {
	It("Get all tenants for migration", func() {
		tenants := database.GetTenantsToMigrate([]string{}, []string{})
		Expect(len(tenants)).To(Equal(database.TestTenantsCount))
	})

	It("Get only included tenants", func() {
		tenants := database.GetTenantsToMigrate([]string{"1", "2"}, []string{})
		Expect(len(tenants)).To(Equal(2))
		Expect(tenants[0].Id).To(Equal(int32(1)))
		Expect(tenants[1].Id).To(Equal(int32(2)))
	})

	It("Get without excluded tenants", func() {
		tenants := database.GetTenantsToMigrate([]string{}, []string{"3", "4"})
		Expect(len(tenants)).To(Equal(database.TestTenantsCount - 2))
		Expect(tenants[0].Id).To(Equal(int32(1)))
		Expect(tenants[1].Id).To(Equal(int32(2)))
	})
})
