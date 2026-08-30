package config_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/config"
)

func createSecretsDir(secrets map[string]string) string {
	tmpDir, err := os.MkdirTemp("", "provisioner-service-secrets-*")
	Expect(err).NotTo(HaveOccurred())

	for key, value := range secrets {
		secretPath := filepath.Join(tmpDir, key)
		err := os.WriteFile(secretPath, []byte(value), 0o600)
		Expect(err).NotTo(HaveOccurred())
	}

	DeferCleanup(func() {
		os.RemoveAll(tmpDir)
	})

	return tmpDir
}

func initConfigForTesting(secretsDir string) {
	runai_config.FlushAll()

	if secretsDir != "" {
		err := runai_config.SetKbSecretsFolder(secretsDir)
		Expect(err).NotTo(HaveOccurred())
	}

	err := runai_config.Register[*config.Server]()
	Expect(err).NotTo(HaveOccurred(), "Failed to register server configuration")

	err = runai_config.Register[*config.TLS]()
	Expect(err).NotTo(HaveOccurred(), "Failed to register TLS configuration")

	err = runai_config.Register[*config.JFrog]()
	Expect(err).NotTo(HaveOccurred(), "Failed to register JFrog configuration")

	config.Init("")
}

var _ = Describe("Config Initialization", func() {
	Describe("Custom Environment Variables", func() {
		BeforeEach(func() {
			os.Setenv("PORT", "9090")
			os.Setenv("RUNAI_TLS_CERT_PATH", "/custom/tls/cert")
			os.Setenv("RUNAI_TLS_KEY_PATH", "/custom/tls/key")

			initConfigForTesting("")
		})

		AfterEach(func() {
			os.Unsetenv("PORT")
			os.Unsetenv("RUNAI_TLS_CERT_PATH")
			os.Unsetenv("RUNAI_TLS_KEY_PATH")
		})

		It("should load custom port from environment variable", func() {
			serverConfig := runai_config.Get[*config.Server]()
			Expect(serverConfig.Port).To(Equal(9090))
		})

		It("should load custom TLS paths from environment variables", func() {
			tlsConfig := runai_config.Get[*config.TLS]()
			Expect(tlsConfig.CertPath).To(Equal("/custom/tls/cert"))
			Expect(tlsConfig.KeyPath).To(Equal("/custom/tls/key"))
		})
	})

	Describe("Installation Version Defaults", func() {
		BeforeEach(func() {
			initConfigForTesting("")
		})

		It("should default CpVersion to the pinned Run:AI version", func() {
			installConfig := runai_config.Get[*config.Installation]()
			Expect(installConfig.CpVersion).To(Equal(config.DefaultRunaiVersion))
		})

		It("should default ClusterVersion to the pinned Run:AI version", func() {
			installConfig := runai_config.Get[*config.Installation]()
			Expect(installConfig.ClusterVersion).To(Equal(config.DefaultRunaiVersion))
		})
	})

	Describe("Installation Version from Environment Variables", func() {
		BeforeEach(func() {
			os.Setenv("RUNAI_CP_VERSION", "3.0.0")
			os.Setenv("RUNAI_CLUSTER_VERSION", ">0.0.0-0")

			initConfigForTesting("")
		})

		AfterEach(func() {
			os.Unsetenv("RUNAI_CP_VERSION")
			os.Unsetenv("RUNAI_CLUSTER_VERSION")
		})

		It("should override CpVersion from environment variable", func() {
			installConfig := runai_config.Get[*config.Installation]()
			Expect(installConfig.CpVersion).To(Equal("3.0.0"))
		})

		It("should override ClusterVersion from environment variable", func() {
			installConfig := runai_config.Get[*config.Installation]()
			Expect(installConfig.ClusterVersion).To(Equal(">0.0.0-0"))
		})
	})

	Describe("Secrets from Files", func() {
		var secretsDir string

		BeforeEach(func() {
			secrets := map[string]string{
				"runai_jfrog_token": "test-token-123",
			}
			secretsDir = createSecretsDir(secrets)

			os.Setenv("PORT", "8080")

			initConfigForTesting(secretsDir)
		})

		AfterEach(func() {
			os.Unsetenv("PORT")
		})

		It("should load JFrog token from secrets file", func() {
			jfrogConfig := runai_config.Get[*config.JFrog]()
			Expect(jfrogConfig.Token).To(Equal("test-token-123"))
		})
	})
})
