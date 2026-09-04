/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Settings tests", Ordered, func() {
	It("Tenants should have all combinations of settings", func() {
		for i := 0; i < database.TestTenantsCount; i++ {
			tenantId := int32(i + 1)
			settings, err := database.GetTenantSettingsMap(tenantId)
			Expect(err).To(BeNil())

			Expect(settings[database.ProjectManualOverQuota]).To(Equal((i/2)%2 == 0))
			Expect(settings[database.CPUResourcesQuota]).To(Equal((i/4)%2 == 0))
		}
	})
})
