/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"fmt"
	"time"

	"github.com/run-ai/runai/control-plane/cluster-migrator/database"
	"github.com/run-ai/runai/control-plane/cluster-migrator/database/models"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nodepools tests", Ordered, func() {
	It("Should get all nodepools", func() {
		tenants := database.GetTenantsToMigrate([]string{}, []string{})
		for _, tenant := range tenants {
			nodepools, err := database.GetTenantNodePools(tenant.Id)
			Expect(err).To(BeNil())
			Expect(len(nodepools)).To(Equal(database.TestNodepoolsCountPerTenant))
		}
	})
})

func validateNodepoolsMigration(tenant models.Tenant) {
	nodepools, err := database.GetTargetTenantNodepools(tenant.Id)
	Expect(err).To(BeNil())
	Expect(len(nodepools)).To(Equal(database.TestNodepoolsCountPerTenant))

	for _, nodepool := range nodepools {
		Expect(nodepool.Name).To(BeElementOf("default", fmt.Sprintf("nodepool-%d", nodepool.ID)))
		Expect(nodepool.TenantId).To(Equal(tenant.Id))
		Expect(nodepool.ClusterId.String()).To(BeElementOf(database.MapTenantIdToClusterIds[tenant.Id]))
		Expect(nodepool.CreatedAt).To(Not(Equal(time.Time{})))
		Expect(nodepool.UpdatedAt).To(Not(Equal(time.Time{})))
		Expect(nodepool.DeletedAt.Valid).To(BeFalse())
		Expect(nodepool.DeletedAt.Time).To(Equal(time.Time{}))
	}
}
