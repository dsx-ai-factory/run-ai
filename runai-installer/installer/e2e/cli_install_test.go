//go:build e2e

package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-installer/installer/e2e/utils"
)

var _ = Describe("Installer CLI E2E", func() {
	BeforeEach(func() {
		Expect(utils.Uninstall(testEnv.TempDir)).To(Succeed())
	})

	It("should install self-hosted with self-signed certs via the CLI binary", func() {
		if testEnv.KubernetesDistribution == "openshift" {
			Skip("Self-signed CA tests are skipped on OpenShift - OpenShift handles TLS with the default cert on the route level")
		}

		configPairs := []string{
			fmt.Sprintf("fqdn=%s", testEnv.Domain),
			fmt.Sprintf("cluster_name=%s", testEnv.ClusterName),
			fmt.Sprintf("ssl_cert_path=%s", testEnv.TLSCertPath),
			fmt.Sprintf("ssl_key_path=%s", testEnv.TLSKeyPath),
			fmt.Sprintf("ca_chain_path=%s", testEnv.CAChainPath()),
		}
		if testEnv.JfrogToken != "" {
			configPairs = append(configPairs, fmt.Sprintf("jfrog_creds=%s", testEnv.JfrogToken))
		}
		if testEnv.KubernetesDistribution != "" {
			configPairs = append(configPairs, fmt.Sprintf("kubernetes_distribution=%s", testEnv.KubernetesDistribution))
		}

		args := []string{
			"install",
			"--gpu", "fake",
			"--dev",
			"--runai-install-format", "self-hosted",
			"--logs-dir", testEnv.TempDir,
			"--step-timeout", "30m",
			"--execution-timeout", "90m",
			"--pre-req", "none",
			"--config-values", strings.Join(configPairs, ";"),
		}

		cmd := exec.Command(installerBinaryPath, args...)
		fmt.Fprintf(GinkgoWriter, "Running: %s %s\n", installerBinaryPath, strings.Join(args, " "))
		cmd.Stdout = GinkgoWriter
		cmd.Stderr = GinkgoWriter
		err := cmd.Run()
		Expect(err).ToNot(HaveOccurred())
	})
})

func buildInstallerBinary() (string, error) {
	binPath := filepath.Join(os.TempDir(), "installer-e2e-test")
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/")
	cmd.Dir = filepath.Join("..")
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to build installer binary: %w", err)
	}
	return binPath, nil
}
