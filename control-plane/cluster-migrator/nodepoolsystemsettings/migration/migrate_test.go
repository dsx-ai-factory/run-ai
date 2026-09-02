/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	migrator_config "github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator/database/models"
	clusterServiceModels "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
)

var _ = Describe("Migration tests", Ordered, func() {
	BeforeEach(func() {
		Expect(deleteNPSystemSettingsCompleted()).To(BeNil())
	})

	It("Migrate tenants", func() {
		tenants := database.GetTenantsToMigrate([]string{"1", "2", "3"}, []string{})

		// validate before the migration - just to make sure it's not a redundant test
		for _, tenant := range tenants {
			validateBeforeNodepoolSystemSettingsMigration(tenant)
			Expect(database.IsNodepoolSystemSettingsMigrationCompletedForTenant(tenant.Id)).To(BeFalse())
		}
		Expect(database.IsNodepoolSystemSettingsMigrationCompleted()).To(BeFalse())

		hadError := migrateTenants(config.Get[*migrator_config.ConfigData](), tenants)
		Expect(hadError).To(BeFalse())

		for _, tenant := range tenants {
			validateNodepoolSystemSettingsMigration(tenant)
			Expect(database.IsNodepoolSystemSettingsMigrationCompletedForTenant(tenant.Id)).To(BeTrue())
		}
		Expect(database.IsNodepoolSystemSettingsMigrationCompleted()).To(BeTrue())

		// now try to run migration again for another tenant -
		// should be no-op, because the whole migration is already marked as completed
		tenants = database.GetTenantsToMigrate([]string{"4"}, []string{})
		hadError = migrateTenants(config.Get[*migrator_config.ConfigData](), tenants)
		Expect(hadError).To(BeFalse())
		Expect(database.IsNodepoolSystemSettingsMigrationCompletedForTenant(4)).To(BeFalse())
	})
})

func validateBeforeNodepoolSystemSettingsMigration(tenant models.Tenant) {
	nodepools, err := database.GetTargetTenantNodepools(tenant.Id)
	Expect(err).To(BeNil())
	Expect(len(nodepools)).To(Equal(database.TestNodepoolsCountPerTenant))

	for _, nodepool := range nodepools {
		Expect(nodepool.AllowAdjustableWeights).To(Equal(clusterServiceModels.DefaultAllowAdjustableWeights))
		Expect(nodepool.AllowCpuQuota).To(Equal(clusterServiceModels.DefaultAllowCpuQuota))
	}

	// just make sure that at least one of those is changed - Default value different from value coming from settings
	Expect(clusterServiceModels.DefaultAllowCpuQuota != database.TestCpuQuotaSetting || clusterServiceModels.DefaultAllowAdjustableWeights != database.TestWeightsSetting).To(BeTrue())
}

func validateNodepoolSystemSettingsMigration(tenant models.Tenant) {
	nodepools, err := database.GetTargetTenantNodepools(tenant.Id)
	Expect(err).To(BeNil())
	Expect(len(nodepools)).To(Equal(database.TestNodepoolsCountPerTenant))

	for _, nodepool := range nodepools {
		Expect(nodepool.AllowAdjustableWeights).To(Equal(database.TestWeightsSetting))
		Expect(nodepool.AllowCpuQuota).To(Equal(database.TestCpuQuotaSetting))
	}
}
