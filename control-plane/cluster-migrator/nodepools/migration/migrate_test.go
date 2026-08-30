package migration

import (
	migrator_config "github.com/run-ai/runai/control-plane/cluster-migrator/config"
	"github.com/run-ai/runai/control-plane/cluster-migrator/database"
	"github.com/run-ai/runai/runai-common-packages/config"

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
