package e2e_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-installer/provisioner-service/e2e/utils"
)

var _ = Describe("Docker Compose Provisioner E2E", func() {
	var (
		dcEnv      *utils.DockerComposeEnv
		secretsDir string
	)

	BeforeEach(func() {
		secretsDir = os.Getenv("RUNAI_SECRETS_DIR")
		Expect(secretsDir).ToNot(BeEmpty(), "RUNAI_SECRETS_DIR environment variable is required")

		By("Destroying any existing RunAI installation")
		err := utils.Uninstall(testEnv.TempDir)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if dcEnv != nil {
			By("Cleaning up docker-compose environment")
			_ = dcEnv.Cleanup()
		}
	})

	Describe("Standard Deployment", func() {
		BeforeEach(func() {
			By("Setting up docker-compose environment")
			var err error
			dcEnv, err = utils.SetupDockerComposeEnv(secretsDir)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when deploying a full environment via docker-compose", func() {
			It("should complete successfully with exit code 0", func() {
				By("Running docker-compose with full provisioning mode")
				exitCode, _, err := dcEnv.RunDockerCompose("full", "e2e-docker-compose@run.ai")

				Expect(err).ToNot(HaveOccurred(), "docker-compose command failed to execute")
				Expect(exitCode).To(Equal(0), "Poller container exited with non-zero code. Check logs above for details.")

				By("Asserting the installed RunAI version matches the default")
				utils.AssertInstalledVersion(utils.TestDomain, dcEnv.GetCAChainPath())
			})
		})
	})

	Describe("Self-Signed CA Certificate Deployment", func() {
		BeforeEach(func() {
			if utils.TestK8sDistribution == "openshift" {
				Skip("Self-signed CA tests are skipped on OpenShift - OpenShift handles TLS with the default cert on the route level")
			}

			By("Setting up docker-compose environment")
			var err error
			dcEnv, err = utils.SetupDockerComposeEnv(secretsDir)
			Expect(err).ToNot(HaveOccurred())

			By("Enabling self-signed CA certificates")
			err = dcEnv.EnableSelfSignedCA()
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when deploying with self-signed CA certificates via docker-compose", func() {
			It("should complete successfully with exit code 0", func() {
				By("Running docker-compose with full provisioning mode and self-signed CA")
				caHint := "HINT: If only this test failed, you might have added a new HTTP client that does not use the CA chain. " +
					"Look at existing clients in clients/pkg/tls and adjust yours to use NewHTTPClientWithCA or NewTransportWithCA."

				exitCode, _, err := dcEnv.RunDockerCompose("full", "e2e-docker-compose-ca@run.ai")

				Expect(err).ToNot(HaveOccurred(), "docker-compose command failed to execute")
				Expect(exitCode).To(Equal(0), "Poller container exited with non-zero code. %s\nCheck logs above for details.", caHint)

				By("Asserting the installed RunAI version matches the default")
				utils.AssertInstalledVersion(utils.TestDomain, dcEnv.GetCAChainPath())
			})
		})
	})
})
