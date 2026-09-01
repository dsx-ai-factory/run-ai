//go:build e2e

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package e2e_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-installer/installer/e2e/utils"
	installercli "github.com/run-ai/runai/runai-installer/installer/internal/cli"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/apiclient"
	installerk8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
	pkgvalidation "github.com/run-ai/runai/runai-installer/installer/pkg/validation"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

var _ = Describe("Installer E2E", func() {
	BeforeEach(func() {
		apiclient.ResetClients()
		err := utils.Uninstall(testEnv.TempDir)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Self-Hosted Installation", func() {
		It("should successfully install a self-hosted environment", func() {
			By("Configuring the installer")
			opts := &config.InstallConfig{
				Mode:               config.ModeCLI,
				DevMode:            true,
				RunaiInstallFormat: config.SelfHosted,
				GPUType:            config.GPUTypeFake,
				Env:                utils.RunaiEnv(),
				CpVersion:          installVersion(),
				ClusterVersion:     installVersion(),
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
			if testEnv.JfrogToken != "" {
				opts.ConfigValues["jfrog_creds"] = testEnv.JfrogToken
			}
			if testEnv.KubernetesDistribution != "" {
				opts.ConfigValues["kubernetes_distribution"] = testEnv.KubernetesDistribution
			}
			if caPath := testEnv.CAChainPath(); caPath != "" {
				opts.ConfigValues["ca_chain_path"] = caPath
			}

			By("Initializing install data and K8s client")
			runaiData, err := model.NewRunAIInstallData(opts)
			Expect(err).ToNot(HaveOccurred())

			k8sClients, err := installerk8s.GetK8sClients()
			Expect(err).ToNot(HaveOccurred())

			installercli.AlignInstallData(runaiData, k8sClients)
			// E2E assumes the cluster already has mandatory prerequisites installed; see
			// e2e/utils/install_helpers.go for rationale.
			runaiData.Prerequisites = nil

			By("Running the installer")
			deps := orchestrator.ProductionInstallDeps(k8sClients, opts.CpVersion)
			err = orchestrator.RunInstall(runaiData, deps)
			Expect(err).ToNot(HaveOccurred())

			By("Waiting for the runai cluster to stabilize")
			Expect(utils.WaitForRunaiClusterStable(10 * time.Minute)).To(Succeed())
		})
	})

	Describe("Self-Hosted Installation with Host-Based Routing", func() {
		It("should successfully install with host-based routing", func() {
			By("Configuring the installer with host-based routing")
			opts := &config.InstallConfig{
				Mode:               config.ModeCLI,
				DevMode:            true,
				RunaiInstallFormat: config.SelfHosted,
				GPUType:            config.GPUTypeFake,
				Env:                utils.RunaiEnv(),
				CpVersion:          installVersion(),
				ClusterVersion:     installVersion(),
				StepTimeout:        30 * time.Minute,
				ExecutionTimeout:   90 * time.Minute,
				LogsDir:            testEnv.TempDir,
				ConfigValues: config.MapFlags{
					"fqdn":                   testEnv.Domain,
					"cluster_name":           testEnv.ClusterName,
					"ssl_cert_path":          testEnv.TLSCertPath,
					"ssl_key_path":           testEnv.TLSKeyPath,
					"workload_routing":       pkgvalidation.WorkloadRoutingHost,
					"host_routing_cert_path": testEnv.WildcardCertPath(),
					"host_routing_key_path":  testEnv.WildcardKeyPath(),
				},
			}
			if testEnv.JfrogToken != "" {
				opts.ConfigValues["jfrog_creds"] = testEnv.JfrogToken
			}
			if testEnv.KubernetesDistribution != "" {
				opts.ConfigValues["kubernetes_distribution"] = testEnv.KubernetesDistribution
			}
			if caPath := testEnv.CAChainPath(); caPath != "" {
				opts.ConfigValues["ca_chain_path"] = caPath
			}

			By("Initializing install data and K8s client")
			runaiData, err := model.NewRunAIInstallData(opts)
			Expect(err).ToNot(HaveOccurred())

			k8sClients, err := installerk8s.GetK8sClients()
			Expect(err).ToNot(HaveOccurred())

			installercli.AlignInstallData(runaiData, k8sClients)
			// E2E assumes the cluster already has mandatory prerequisites installed; see
			// e2e/utils/install_helpers.go for rationale.
			runaiData.Prerequisites = nil

			By("Running the installer")
			deps := orchestrator.ProductionInstallDeps(k8sClients, opts.CpVersion)
			err = orchestrator.RunInstall(runaiData, deps)
			Expect(err).ToNot(HaveOccurred())

			By("Waiting for the runai cluster to stabilize")
			Expect(utils.WaitForRunaiClusterStable(10 * time.Minute)).To(Succeed())

			By("Verifying clusterConfig.global.subdomainSupport=true in cluster helm values")
			actionCfg := &action.Configuration{}
			settings := cli.New()
			settings.SetNamespace(step.ClusterNamespace)
			Expect(actionCfg.Init(settings.RESTClientGetter(), step.ClusterNamespace, "", nil)).To(Succeed())
			getValues := action.NewGetValues(actionCfg)
			getValues.AllValues = true
			vals, err := getValues.Run(step.ClusterReleaseName)
			Expect(err).ToNot(HaveOccurred())
			clusterCfg, ok := vals["clusterConfig"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "expected clusterConfig key in helm values")
			globalCfg, ok := clusterCfg["global"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "expected clusterConfig.global key in helm values")
			Expect(globalCfg["subdomainSupport"]).To(Equal(true))
		})
	})
})
