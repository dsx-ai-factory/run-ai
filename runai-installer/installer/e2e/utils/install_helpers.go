//go:build e2e

package utils

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-installer/installer/internal/cli"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator"
	installerk8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
)

func RunaiEnv() config.RunaiEnvType {
	v := os.Getenv("E2E_RUNAI_ENV")
	switch v {
	case "", "prod":
		return config.EnvProd
	case "test":
		return config.EnvTest
	case "lab":
		return config.EnvLab
	case "staging":
		return config.EnvStaging
	default:
		Fail(fmt.Sprintf("invalid E2E_RUNAI_ENV value %q: must be prod, test, staging, or lab", v))
		return config.EnvProd // unreachable
	}
}

func InstallVersion(version string, testEnv *TestEnv, extraConfigValues config.MapFlags) {
	By(fmt.Sprintf("Installing version %s", version))
	opts := &config.InstallConfig{
		Mode:               config.ModeCLI,
		DevMode:            true,
		RunaiInstallFormat: config.SelfHosted,
		GPUType:            config.GPUTypeFake,
		Env:                RunaiEnv(),
		CpVersion:          version,
		ClusterVersion:     version,
		StepTimeout:        30 * time.Minute,
		ExecutionTimeout:   90 * time.Minute,
		LogsDir:            testEnv.TempDir,
		ConfigValues: config.MapFlags{
			"fqdn":          testEnv.Domain,
			"cluster_name":  testEnv.ClusterName,
			"ssl_cert_path": testEnv.TLSCertPath,
			"ssl_key_path":  testEnv.TLSKeyPath,
		},
	}
	for k, v := range extraConfigValues {
		opts.ConfigValues[k] = v
	}
	if testEnv.JfrogToken != "" {
		opts.ConfigValues["jfrog_creds"] = testEnv.JfrogToken
	}
	if testEnv.KubernetesDistribution != "" {
		opts.ConfigValues["kubernetes_distribution"] = testEnv.KubernetesDistribution
	}
	if caPath := testEnv.CAChainPath(); caPath != "" {
		if _, present := opts.ConfigValues["ca_chain_path"]; !present {
			opts.ConfigValues["ca_chain_path"] = caPath
		}
	}

	runaiData, err := model.NewRunAIInstallData(opts)
	Expect(err).ToNot(HaveOccurred())

	k8sClients, err := installerk8s.GetK8sClients()
	Expect(err).ToNot(HaveOccurred())

	cli.AlignInstallData(runaiData, k8sClients)
	// E2E assumes the cluster already provides mandatory prerequisites (ingress,
	// Prometheus / monitoring.coreos.com CRDs, GPU operator, network operator, etc.).
	// Clearing Prerequisites skips installer prereq flows so runs stay fast and match
	// CI or local scripts that provision the cluster ahead of the test.
	runaiData.Prerequisites = nil

	deps := orchestrator.ProductionInstallDeps(k8sClients, opts.CpVersion)
	err = orchestrator.RunInstall(runaiData, deps)
	Expect(err).ToNot(HaveOccurred())

	By("Waiting for the runai cluster to stabilize before returning")
	Expect(WaitForRunaiClusterStable(10 * time.Minute)).To(Succeed())
}
