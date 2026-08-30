//go:build e2e

package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-installer/installer/e2e/utils"
	"github.com/run-ai/runai/runai-installer/installer/internal/installconfig"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	pkgvalidation "github.com/run-ai/runai/runai-installer/installer/pkg/validation"
)

var _ = Describe("Installer New Flow Airgap E2E", Label(newFlowLabel), func() {
	BeforeEach(func() {
		Expect(utils.Uninstall(testEnv.TempDir)).To(Succeed())
	})

	It("should install an air-gapped self-hosted environment from a spec file", func() {
		version := airgapVersion()
		tarball := downloadOrReuseAirgapPackage(testEnv.JfrogToken, version, testEnv.TempDir, "E2E_AIRGAP_TARBALL")

		workingDir := filepath.Join(testEnv.TempDir, "airgap-extracted")
		Expect(os.MkdirAll(workingDir, 0o755)).To(Succeed())
		DeferCleanup(func() {
			cleanupAirgapArtifacts(workingDir)
			if os.Getenv("E2E_AIRGAP_TARBALL") == "" {
				if err := os.Remove(tarball); err != nil && !os.IsNotExist(err) {
					Expect(err).ToNot(HaveOccurred())
				}
			}
		})

		cfgPath := writeInstallNewConfig(airgapFullInstallConfig(tarball, workingDir, version))

		args := []string{
			"install-new",
			"--config", cfgPath,
			"--yes",
			"--env", utils.RunaiEnv().String(),
			"--step-timeout", "60m",
			"--execution-timeout", "120m",
		}

		cmd := exec.Command(installerBinaryPath, args...)
		fmt.Fprintf(GinkgoWriter, "Running: %s %s\n", installerBinaryPath, strings.Join(args, " "))
		cmd.Stdout = GinkgoWriter
		cmd.Stderr = GinkgoWriter
		Expect(cmd.Run()).To(Succeed())

		By("Waiting for the runai cluster to stabilize")
		Expect(utils.WaitForRunaiClusterStable(10 * time.Minute)).To(Succeed())

		By("Verifying CP and cluster Helm release versions")
		cpVer, err := utils.GetHelmChartVersion(step.CPNamespace, step.CPReleaseName)
		Expect(err).ToNot(HaveOccurred())
		Expect(cpVer).To(Equal(version))

		clusterVer, err := utils.GetHelmChartVersion(step.ClusterNamespace, step.ClusterReleaseName)
		Expect(err).ToNot(HaveOccurred())
		Expect(clusterVer).To(Equal(version))
	})
})

// airgapFullInstallConfig builds an air-gapped full-installation spec wired to
// the shared TestEnv and the per-cluster Harbor mirror. Prerequisites are
// omitted on purpose: their charts pull from public registries unreachable in
// air-gap, so the run assumes the cluster already provides them (matching the
// legacy airgap e2e).
func airgapFullInstallConfig(bundlePath, workingDir, version string) installconfig.Config {
	tls := &installconfig.TLS{
		CertPath:    testEnv.TLSCertPath,
		KeyPath:     testEnv.TLSKeyPath,
		CAChainPath: testEnv.CAChainPath(),
	}

	cfg := installconfig.Config{
		Install: installconfig.Install{Mode: "full_installation", Connectivity: "air_gapped"},
		Airgap: &installconfig.Airgap{
			BundlePath: bundlePath,
			WorkingDir: workingDir,
		},
		Registry: &installconfig.Registry{
			Host:     utils.HarborRegistryURL(testEnv.ClusterName, testEnv.KubernetesDistribution),
			Username: utils.HarborUser(),
			Password: utils.HarborPass(),
		},
		ControlPlane: &installconfig.ControlPlane{
			Enabled: true,
			Version: version,
			Domain:  testEnv.Domain,
			TLS:     tls,
		},
		Cluster: &installconfig.Cluster{
			Enabled: true,
			Name:    testEnv.ClusterName,
			Version: version,
			Domain:  testEnv.Domain,
			TLS:     tls,
		},
	}

	if testEnv.KubernetesDistribution != "" {
		cfg.Install.KubernetesDistribution = testEnv.KubernetesDistribution
	}

	// External routing is required on non-OpenShift installs (OpenShift uses
	// Routes and must leave it unset). The e2e cluster is fronted by the nginx
	// ingress controller.
	if testEnv.KubernetesDistribution != pkgvalidation.KubernetesDistributionOpenShift {
		cfg.ExternalConnection = &installconfig.ExternalConnection{
			Type:    "ingress",
			Ingress: &installconfig.IngressConnection{ClassName: "nginx"},
		}
	}

	return cfg
}
