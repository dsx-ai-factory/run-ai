/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/cluster-migrator/database"
	"github.com/run-ai/runai/control-plane/cluster-migrator/database/models"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

var _ = Describe("Version tests", Ordered, func() {
	BeforeEach(func() {
		Expect(deleteNPSystemSettingsCompleted()).To(BeNil())
	})

	It("Tenant happy flow phases", func() {
		tenants := []int32{100, 200, 300}

		for _, tenant := range tenants {
			err := database.SetNodepoolSystemSettingsMigrationStatusForTenant(tenant, models.STARTED)
			Expect(err).To(BeNil())

			err = database.SetNodepoolSystemSettingsMigrationStatusForTenant(tenant, models.COMPLETED)
			Expect(err).To(BeNil())

			Expect(database.IsNodepoolSystemSettingsMigrationCompleted()).To(BeFalse())

			Expect(database.IsNodepoolSystemSettingsMigrationCompletedForTenant(tenant)).To(BeTrue())
		}

		err := database.SetNodepoolSystemSettingsMigrationFinished(false)
		Expect(err).To(BeNil())

		Expect(database.IsNodepoolSystemSettingsMigrationCompleted()).To(BeTrue())
	})

	It("Tenant fail flow phases", func() {
		tenantId := int32(100)

		err := database.SetNodepoolSystemSettingsMigrationStatusForTenant(tenantId, models.STARTED)
		Expect(err).To(BeNil())

		err = database.SetNodepoolSystemSettingsMigrationStatusForTenant(tenantId, models.FAILED)
		Expect(err).To(BeNil())

		Expect(database.IsNodepoolSystemSettingsMigrationCompletedForTenant(tenantId)).To(BeFalse())

		err = database.SetNodepoolSystemSettingsMigrationFinished(true)
		Expect(err).To(BeNil())

		Expect(database.IsNodepoolSystemSettingsMigrationCompleted()).To(BeFalse())
	})
})

func deleteNPSystemSettingsCompleted() error {
	_, err := runaidb.DB().
		Where("version = ?", "v1").
		Delete(&models.NodepoolSystemSettingsMigrationControl{})
	return err
}
