package config

import (
	"os"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-common-packages/config"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
)

var _ = ginkgo.Describe("Config", func() {
	ginkgo.BeforeEach(func() {
		os.Setenv(dao_utils.EnvSecretsMountPath, ginkgo.GinkgoT().TempDir())
		os.Unsetenv(dao_utils.EnvDbName)
		os.Unsetenv(dao_utils.EnvDbHost)
		os.Unsetenv("SERVICE_PORT")
		os.Unsetenv(dao_utils.EnvDbSslMode)
	})

	ginkgo.Describe("GetConfig", func() {
		ginkgo.Context("with default values", func() {
			ginkgo.It("should return default server and services configuration", func() {
				config.FlushAll()
				Init()

				serverConfig := config.Get[*Server]()
				servicesConfig := config.Get[*Services]()

				gomega.Expect(serverConfig.ServicePort).To(gomega.Equal(8080))
				gomega.Expect(servicesConfig.AuthorizationUrl).To(gomega.Equal("http://runai-backend-authorization:8080"))
			})
		})

		ginkgo.Context("with specific values", func() {
			ginkgo.It("should return configuration with custom service port", func() {
				// Set environment variables BEFORE Init()
				gomega.Expect(os.Setenv("SERVICE_PORT", "9001")).To(gomega.Succeed())

				config.FlushAll()
				Init()

				serverConfig := config.Get[*Server]()
				servicesConfig := config.Get[*Services]()

				gomega.Expect(serverConfig.ServicePort).To(gomega.Equal(9001))
				gomega.Expect(servicesConfig.AuthorizationUrl).To(gomega.Equal("http://runai-backend-authorization:8080"))
			})
		})
	})

	ginkgo.Describe("ServerConfig", func() {
		ginkgo.Context("with defaults", func() {
			ginkgo.It("should have correct default values", func() {
				config.FlushAll()
				Init()

				serverConfig := config.Get[*Server]()

				gomega.Expect(serverConfig.ServicePort).To(gomega.Equal(8080))
			})
		})
	})

	ginkgo.Describe("ServicesConfig", func() {
		ginkgo.Context("with defaults", func() {
			ginkgo.It("should have correct default authorization URL", func() {
				config.FlushAll()
				Init()

				servicesConfig := config.Get[*Services]()

				gomega.Expect(servicesConfig.AuthorizationUrl).To(gomega.Equal("http://runai-backend-authorization:8080"))
			})
		})
	})

})
