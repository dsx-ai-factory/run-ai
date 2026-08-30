package config

import (
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	MockConfig()
	Init()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Config", func() {
	Describe("GetConfig", func() {
		Context("with mock values", func() {
			It("should create a valid mocked configuration", func() {
				// Validate ServerConfig mocked values
				serverConf := runai_config.Get[*ServerConfig]()
				Expect(serverConf.ServicePort).To(Equal(8080))
				Expect(serverConf.PoliciesPerBulk).To(Equal(50))
				Expect(serverConf.RetryMaxIntervalSec).To(Equal(10))
				Expect(serverConf.RetryMaxElapsedSec).To(Equal(60))
				Expect(serverConf.DaysLimitForFirstSync).To(Equal(7))

				// Validate ServicesConfig mocked values
				servicesConf := runai_config.Get[*ServicesConfig]()
				Expect(servicesConf.AssetsServiceUrl).To(Equal("http://runai-backend-assets-service:8080"))
				Expect(servicesConf.ClusterServiceUrl).To(Equal("http://runai-backend-cluster-service:8080"))
			})
		})
	})
})
