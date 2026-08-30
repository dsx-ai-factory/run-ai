package config

import (
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	MockConfig()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Config", func() {
	Describe("GetConfig", func() {
		Context("with specific values", func() {
			It("should create valid configuration", func() {
				Expect(os.Setenv("SERVICE_PORT", "9001")).To(BeNil())
				Expect(os.Setenv("SECRET_DATA_KEY_USER", "some_secret_user")).To(BeNil())
				Expect(os.Setenv("SECRET_DATA_KEY_PASSWORD", "some_secret_password")).To(BeNil())

				Expect(os.Setenv("ORG_UNIT_CACHE_EXPIRATION_SEC", "118")).To(BeNil())
				Expect(os.Setenv("TENANT_KEYS_CACHE_EXPIRATION_SEC", "1518")).To(BeNil())
				Expect(os.Setenv("CLUSTER_CACHE_EXPIRATION_SEC", "1528")).To(BeNil())
				Expect(os.Setenv("PROJECTS_CACHE_EXPIRATION_SEC", "1536")).To(BeNil())

				registerConfigs() // must call registerConfigs to apply the new config
				sc := runai_config.Get[*SecretConfig]()
				Expect(sc.DataKeyUser).To(Equal("some_secret_user"))
				Expect(sc.DataKeyPassword).To(Equal("some_secret_password"))
				Expect(runai_config.Get[*OrgUnitCache]().ExpirationSec).To(Equal(118))
				Expect(runai_config.Get[*TenantKeysCache]().ExpirationSec).To(Equal(1518))
				Expect(runai_config.Get[*ClusterCache]().ExpirationSec).To(Equal(1528))
				Expect(runai_config.Get[*ProjectsCache]().ExpirationSec).To(Equal(1536))
			})
		})
	})
})
