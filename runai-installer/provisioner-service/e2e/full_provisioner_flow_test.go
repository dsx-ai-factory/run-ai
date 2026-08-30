package e2e_test

import (
	"net/http/httptest"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-installer/installer/pkg/api"
	"github.com/run-ai/runai/runai-installer/provisioner-service/e2e/utils"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/config"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/router"
)

var _ = Describe("Full Provisioner Flow E2E", func() {
	var server *httptest.Server

	BeforeEach(func() {
		api.Reset()

		err := utils.Uninstall(testEnv.TempDir)
		Expect(err).ToNot(HaveOccurred())

		handler := router.Create()
		server = httptest.NewServer(handler)
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})

	Describe("Control Plane Installation", func() {
		Context("when creating a control-plane-only environment", func() {
			It("should successfully provision the control plane", func() {
				By("Creating a new environment via POST /environments")
				requestBody := map[string]interface{}{
					"provisioningMode": "control-plane-only",
					"domain":           utils.TestDomain,
					"wildcarddomain":   utils.TestWildcardDomain,
				}
				if utils.TestK8sDistribution != "" && utils.TestK8sDistribution != "kind" {
					requestBody["kubernetesDistribution"] = utils.TestK8sDistribution
				}

				accepted := utils.CreateEnvironment(server.URL, requestBody)
				Expect(accepted.Mode).To(Equal("control-plane-only"))

				By("Polling GET /environments/{id} until installation completes")
				finalStatus := utils.WaitForEnvironmentCompletion(server.URL, accepted.Id, 15*time.Minute, 5*time.Second)
				Expect(finalStatus.Phase).To(Equal("succeeded"))

				By("Asserting the installed RunAI version matches the default")
				utils.AssertInstalledVersion(utils.TestDomain, os.Getenv("RUNAI_CA_CHAIN_PATH"))
			})
		})

		Context("when creating a full environment", func() {
			It("should successfully provision the full environment", func() {
				By("Creating a new environment via POST /environments")
				requestBody := map[string]interface{}{
					"provisioningMode": "full",
					"domain":           utils.TestDomain,
					"wildcarddomain":   utils.TestWildcardDomain,
					"userEmail":        "test@example.com",
				}
				if utils.TestK8sDistribution != "" && utils.TestK8sDistribution != "kind" {
					requestBody["kubernetesDistribution"] = utils.TestK8sDistribution
				}

				accepted := utils.CreateEnvironment(server.URL, requestBody)
				Expect(accepted.Mode).To(Equal("full"))

				By("Polling GET /environments/{id} until installation completes")
				finalStatus := utils.WaitForEnvironmentCompletion(server.URL, accepted.Id, 15*time.Minute, 5*time.Second)
				Expect(finalStatus.Phase).To(Equal("succeeded"))

				By("Waiting for the runai cluster to stabilize")
				Expect(utils.WaitForRunaiClusterStable(10 * time.Minute)).To(Succeed())

				By("Asserting the installed RunAI version matches the default")
				utils.AssertInstalledVersion(utils.TestDomain, os.Getenv("RUNAI_CA_CHAIN_PATH"))
			})
		})
	})

	Describe("Installation Timeout", func() {
		var (
			timeoutServer   *httptest.Server
			originalTimeout string
		)

		BeforeEach(func() {
			api.Reset()

			err := utils.Uninstall(testEnv.TempDir)
			Expect(err).ToNot(HaveOccurred())

			originalTimeout = os.Getenv("RUNAI_EXECUTION_TIMEOUT_MINUTES")
			os.Setenv("RUNAI_EXECUTION_TIMEOUT_MINUTES", "1")
			config.ReloadInstallation()

			handler := router.Create()
			timeoutServer = httptest.NewServer(handler)
		})

		AfterEach(func() {
			if timeoutServer != nil {
				timeoutServer.Close()
			}
			if originalTimeout != "" {
				os.Setenv("RUNAI_EXECUTION_TIMEOUT_MINUTES", originalTimeout)
			} else {
				os.Unsetenv("RUNAI_EXECUTION_TIMEOUT_MINUTES")
			}
			config.ReloadInstallation()
		})

		Context("when execution timeout is set to 1 minute", func() {
			It("should fail the installation due to timeout", func() {
				By("Creating a new environment via POST /environments")
				requestBody := map[string]interface{}{
					"provisioningMode": "control-plane-only",
					"domain":           utils.TestDomain,
					"wildcarddomain":   utils.TestWildcardDomain,
				}
				if utils.TestK8sDistribution != "" && utils.TestK8sDistribution != "kind" {
					requestBody["kubernetesDistribution"] = utils.TestK8sDistribution
				}

				accepted := utils.CreateEnvironment(timeoutServer.URL, requestBody)
				Expect(accepted.Mode).To(Equal("control-plane-only"))

				By("Polling GET /environments/{id} until installation fails due to timeout")
				finalStatus := utils.WaitForEnvironmentFailure(timeoutServer.URL, accepted.Id, 2*time.Minute, 1*time.Second)
				Expect(finalStatus.Message).To(ContainSubstring("failed"))
			})
		})
	})

	Describe("Self-Signed CA Certificate Installation", func() {
		var caServer *httptest.Server

		BeforeEach(func() {
			if utils.TestK8sDistribution == "openshift" {
				Skip("Self-signed CA tests are skipped on OpenShift - OpenShift handles TLS with the default cert on the route level")
			}

			api.Reset()

			err := utils.Uninstall(testEnv.TempDir)
			Expect(err).ToNot(HaveOccurred())

			By("Enabling self-signed CA certificates")
			err = testEnv.EnableSelfSignedCA()
			Expect(err).ToNot(HaveOccurred())

			handler := router.Create()
			caServer = httptest.NewServer(handler)
		})

		AfterEach(func() {
			if caServer != nil {
				caServer.Close()
			}
			testEnv.DisableSelfSignedCA()
		})

		Context("when creating a full environment with self-signed CA", func() {
			It("should successfully provision the full environment using custom CA", func() {
				By("Creating a new environment via POST /environments")
				requestBody := map[string]interface{}{
					"provisioningMode": "full",
					"domain":           utils.TestDomain,
					"wildcarddomain":   utils.TestWildcardDomain,
					"userEmail":        "e2e-selfsigned@run.ai",
				}
				if utils.TestK8sDistribution != "" && utils.TestK8sDistribution != "kind" {
					requestBody["kubernetesDistribution"] = utils.TestK8sDistribution
				}

				accepted := utils.CreateEnvironment(caServer.URL, requestBody)
				Expect(accepted.Mode).To(Equal("full"))

				By("Polling GET /environments/{id} until installation completes")
				caHint := "HINT: If only this test failed, you might have added a new HTTP client that does not use the CA chain. " +
					"Look at existing clients in clients/pkg/tls and adjust yours to use NewHTTPClientWithCA or NewTransportWithCA."
				finalStatus := utils.WaitForEnvironmentCompletion(caServer.URL, accepted.Id, 15*time.Minute, 5*time.Second, caHint)
				Expect(finalStatus.Phase).To(Equal("succeeded"), caHint)

				By("Waiting for the runai cluster to stabilize")
				Expect(utils.WaitForRunaiClusterStable(10 * time.Minute)).To(Succeed())

				By("Asserting the installed RunAI version matches the default")
				utils.AssertInstalledVersion(utils.TestDomain, os.Getenv("RUNAI_CA_CHAIN_PATH"))
			})
		})
	})

})
