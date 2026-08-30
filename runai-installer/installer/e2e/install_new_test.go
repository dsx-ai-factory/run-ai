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
	pkgvalidation "github.com/run-ai/runai/runai-installer/installer/pkg/validation"
	"gopkg.in/yaml.v3"
)

// newFlowLabel tags every spec that drives the new install-new / orchestrator2
// flow so they can be run in isolation, e.g.
// go test -tags e2e ./e2e -args -ginkgo.label-filter=install-new
const newFlowLabel = "install-new"

var _ = Describe("Installer New Flow E2E", Label(newFlowLabel), func() {
	BeforeEach(func() {
		Expect(utils.Uninstall(testEnv.TempDir)).To(Succeed())
		// On OpenShift the prerequisites (GPU/network operators, monitoring) are
		// provided by the platform and must not be torn down; everywhere else we
		// wipe them so the new flow installs them from scratch and the run
		// actually exercises the prereq install path (not just Detect).
		if testEnv.KubernetesDistribution != pkgvalidation.KubernetesDistributionOpenShift {
			Expect(utils.CleanupPrerequisites()).To(Succeed())
		}
	})

	It("should install a connected self-hosted environment from a spec file", func() {
		cfg := connectedFullInstallConfig()
		// On OpenShift prerequisites are not torn down between runs, so the
		// clean-slate check does not apply.
		if testEnv.KubernetesDistribution != pkgvalidation.KubernetesDistributionOpenShift {
			ensurePrereqsNotInstalled(cfg)
		}
		cfgPath := writeInstallNewConfig(cfg)

		args := []string{
			"install-new",
			"--config", cfgPath,
			"--yes",
			"--env", utils.RunaiEnv().String(),
			"--step-timeout", "30m",
			"--execution-timeout", "90m",
		}

		cmd := exec.Command(installerBinaryPath, args...)
		fmt.Fprintf(GinkgoWriter, "Running: %s %s\n", installerBinaryPath, strings.Join(args, " "))
		cmd.Stdout = GinkgoWriter
		cmd.Stderr = GinkgoWriter
		Expect(cmd.Run()).To(Succeed())

		By("Waiting for the runai cluster to stabilize")
		Expect(utils.WaitForRunaiClusterStable(10 * time.Minute)).To(Succeed())

		By("Verifying the requested prerequisites were actually installed")
		ensurePrereqsInstalled(cfg)
	})
})

// connectedFullInstallConfig builds a connected full-installation spec (control
// plane + cluster) wired to the shared TestEnv: domain, cluster name, generated
// self-signed certs and the JFrog token from RUNAI_SECRETS_DIR. Mandatory
// prerequisites are requested so the run exercises the new prereq flow; the GPU
// operator is fake to avoid real drivers.
func connectedFullInstallConfig() installconfig.Config {
	tls := &installconfig.TLS{
		CertPath:    testEnv.TLSCertPath,
		KeyPath:     testEnv.TLSKeyPath,
		CAChainPath: testEnv.CAChainPath(),
	}

	cfg := installconfig.Config{
		Install: installconfig.Install{Mode: "full_installation", Connectivity: "connected"},
		ControlPlane: &installconfig.ControlPlane{
			Enabled: true,
			Version: installVersion(),
			Domain:  testEnv.Domain,
			TLS:     tls,
		},
		Cluster: &installconfig.Cluster{
			Enabled: true,
			Name:    testEnv.ClusterName,
			Version: installVersion(),
			Domain:  testEnv.Domain,
			TLS:     tls,
		},
		Prerequisites: prerequisitesFor(testEnv.KubernetesDistribution),
		Features: &installconfig.Features{
			LWS: &installconfig.Dependency{Mode: "install"},
			// The NIM operator's admission webhook defaults to cert-manager for
			// its TLS certs; cert-manager is not present in the e2e cluster, so
			// disable the admission controller to keep the install self-contained.
			NIMOperator: &installconfig.Dependency{
				Mode:             "install",
				CustomHelmValues: map[string]any{"operator.admissionController.enabled": false},
			},
			Grove: &installconfig.Dependency{Mode: "install"},
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

	if testEnv.JfrogToken != "" {
		cfg.Registry = &installconfig.Registry{
			Type:      "jfrog",
			TokenPath: filepath.Join(testEnv.TempDir, "runai_jfrog_token"),
		}
	}

	return cfg
}

// prerequisitesFor returns the prerequisite set to install. On OpenShift the
// cluster already provides the prerequisites (GPU/network operators, monitoring),
// so they are skipped and left untouched; LWS is still installed as a feature.
// Everywhere else the mandatory prerequisites are installed from scratch.
func prerequisitesFor(distribution string) *installconfig.Prerequisites {
	if distribution == pkgvalidation.KubernetesDistributionOpenShift {
		return &installconfig.Prerequisites{
			Prometheus:      &installconfig.Dependency{Mode: "skip"},
			GPUOperator:     &installconfig.Dependency{Mode: "skip"},
			NetworkOperator: &installconfig.Dependency{Mode: "skip"},
		}
	}
	return &installconfig.Prerequisites{
		Prometheus:      &installconfig.Dependency{Mode: "install"},
		GPUOperator:     &installconfig.Dependency{Mode: "install", Type: "fake"},
		NetworkOperator: &installconfig.Dependency{Mode: "install"},
	}
}

type prereqRelease struct{ namespace, release string }

func installablePrereqReleases(cfg installconfig.Config) []prereqRelease {
	var releases []prereqRelease
	add := func(d *installconfig.Dependency, namespace, release string) {
		if d != nil && d.Mode == "install" {
			releases = append(releases, prereqRelease{namespace, release})
		}
	}
	if p := cfg.Prerequisites; p != nil {
		add(p.Prometheus, "monitoring", "kube-prometheus-stack")
		add(p.GPUOperator, "gpu-operator", "gpu-operator")
		add(p.NetworkOperator, "nvidia-network-operator", "network-operator")
	}
	if f := cfg.Features; f != nil {
		add(f.LWS, "lws-system", "lws")
		add(f.NIMOperator, "nim-operator", "nim-operator")
		add(f.Grove, "grove-system", "grove")
	}
	return releases
}

func ensurePrereqsNotInstalled(cfg installconfig.Config) {
	for _, r := range installablePrereqReleases(cfg) {
		_, err := utils.GetHelmChartVersion(r.namespace, r.release)
		Expect(err).To(HaveOccurred(), fmt.Sprintf(
			"prerequisite helm release %q in namespace %q must not exist before install (clean slate)",
			r.release, r.namespace,
		))
	}
}

func ensurePrereqsInstalled(cfg installconfig.Config) {
	for _, r := range installablePrereqReleases(cfg) {
		_, err := utils.GetHelmChartVersion(r.namespace, r.release)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf(
			"prerequisite helm release %q in namespace %q must be installed after install",
			r.release, r.namespace,
		))
	}
}

// writeInstallNewConfig marshals the config to YAML, prints it, and writes it to
// a temp file whose path is returned for --config.
func writeInstallNewConfig(cfg installconfig.Config) string {
	out, err := yaml.Marshal(cfg)
	Expect(err).ToNot(HaveOccurred())

	fmt.Fprintf(GinkgoWriter, "\n--- install-new spec ---\n%s------------------------\n", string(out))

	cfgPath := filepath.Join(testEnv.TempDir, "install-new-spec.yaml")
	Expect(os.WriteFile(cfgPath, out, 0o600)).To(Succeed())
	return cfgPath
}
