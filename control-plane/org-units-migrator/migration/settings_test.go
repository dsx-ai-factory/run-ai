package migration

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/org-units-migrator/database"
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
