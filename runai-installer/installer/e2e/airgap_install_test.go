//go:build e2e

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/e2e/utils"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/cli"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/upgrade"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/apiclient"
	installerk8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
)

const (
	defaultAirgapVersion        = "2.25.6"
	defaultAirgapUpgradeVersion = "2.25.9"
)

func airgapVersion() string {
	if v := os.Getenv("E2E_AIRGAP_VERSION"); v != "" {
		return v
	}
	return defaultAirgapVersion
}

func airgapUpgradeVersion() string {
	if v := os.Getenv("E2E_AIRGAP_UPGRADE_VERSION"); v != "" {
		return v
	}
	return defaultAirgapUpgradeVersion
}

func downloadOrReuseAirgapPackage(token, version, destDir, envOverride string) string {
	tarballPath := os.Getenv(envOverride)
	if tarballPath != "" {
		By(fmt.Sprintf("Using existing airgap package: %s", tarballPath))
		return tarballPath
	}
	By(fmt.Sprintf("Downloading airgap package version %s", version))
	path, err := utils.DownloadAirgapPackage(token, version, destDir)
	Expect(err).ToNot(HaveOccurred())
	return path
}

func cleanupAirgapArtifacts(extractedDir string) {
	if extractedDir != "" {
		By(fmt.Sprintf("Cleaning up extracted dir: %s", extractedDir))
		Expect(os.RemoveAll(extractedDir)).To(Succeed())
	}
	if os.Getenv("E2E_AIRGAP_PRUNE_IMAGES") == "true" {
		By("Pruning unused Docker images")
		cmd := exec.Command("docker", "system", "prune", "-af")
		out, err := cmd.CombinedOutput()
		Expect(err).ToNot(HaveOccurred(), string(out))
	}
}

func airgapInstall(tarballPath, version string) string {
	registryURL := utils.HarborRegistryURL(testEnv.ClusterName, testEnv.KubernetesDistribution)

	By(fmt.Sprintf("Configuring airgap install v%s (registry: %s)", version, registryURL))
	opts := &config.InstallConfig{
		Mode:               config.ModeCLI,
		DevMode:            true,
		RunaiInstallFormat: config.Airgap,
		GPUType:            config.GPUTypeFake,
		StepTimeout:        30 * time.Minute,
		ExecutionTimeout:   60 * time.Minute,
		LogsDir:            testEnv.TempDir,
		ConfigValues: config.MapFlags{
			"fqdn":                testEnv.Domain,
			"cluster_name":        testEnv.ClusterName,
			"ssl_cert_path":       testEnv.TLSCertPath,
			"ssl_key_path":        testEnv.TLSKeyPath,
			"airgap_package_path": tarballPath,
			"registry_url":        registryURL,
			"registry_username":   utils.HarborUser(),
			"registry_password":   utils.HarborPass(),
		},
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

	cli.AlignInstallData(runaiData, k8sClients)
	runaiData.Prerequisites = nil

	By("Running the airgap installer")
	deps := orchestrator.ProductionInstallDeps(k8sClients, opts.CpVersion)
	err = orchestrator.RunInstall(runaiData, deps)
	Expect(err).ToNot(HaveOccurred())

	By("Waiting for the runai cluster to stabilize")
	Expect(utils.WaitForRunaiClusterStable(10 * time.Minute)).To(Succeed())

	By("Verifying CP Helm release version")
	cpVer, err := utils.GetHelmChartVersion(step.CPNamespace, step.CPReleaseName)
	Expect(err).ToNot(HaveOccurred())
	Expect(cpVer).To(Equal(version))

	By("Verifying cluster Helm release version")
	clusterVer, err := utils.GetHelmChartVersion(step.ClusterNamespace, step.ClusterReleaseName)
	Expect(err).ToNot(HaveOccurred())
	Expect(clusterVer).To(Equal(version))

	return runaiData.Config.AirgapExtractedDir
}

func airgapUpgrade(tarballPath, version string) string {
	registryURL := utils.HarborRegistryURL(testEnv.ClusterName, testEnv.KubernetesDistribution)

	By(fmt.Sprintf("Configuring airgap upgrade to v%s (registry: %s)", version, registryURL))
	upgradeCfg := &config.UpgradeConfig{
		DevMode:             true,
		LogsDir:             testEnv.TempDir,
		StepTimeout:         30 * time.Minute,
		ExecutionTimeout:    60 * time.Minute,
		UpgradeControlPlane: true,
		UpgradeCluster:      true,
		AirgapPackagePath:   tarballPath,
		RegistryURL:         registryURL,
		RegistryUsername:    utils.HarborUser(),
		RegistryPassword:    utils.HarborPass(),
		CACertPath:          testEnv.CAChainPath(),
	}

	By("Running the airgap upgrade")
	k8sClients, err := installerk8s.GetK8sClients()
	Expect(err).ToNot(HaveOccurred())
	deps := upgrade.ProductionDeps(k8sClients, version)
	err = orchestrator.RunUpgrade(upgradeCfg, deps)
	Expect(err).ToNot(HaveOccurred())

	By("Verifying CP Helm release version after upgrade")
	cpVer, err := utils.GetHelmChartVersion(step.CPNamespace, step.CPReleaseName)
	Expect(err).ToNot(HaveOccurred())
	Expect(cpVer).To(Equal(version))

	By("Verifying cluster Helm release version after upgrade")
	clusterVer, err := utils.GetHelmChartVersion(step.ClusterNamespace, step.ClusterReleaseName)
	Expect(err).ToNot(HaveOccurred())
	Expect(clusterVer).To(Equal(version))

	return upgradeCfg.AirgapExtractedDir
}

var _ = Describe("Airgap Installation E2E", func() {
	BeforeEach(func() {
		apiclient.ResetClients()
		err := utils.Uninstall(testEnv.TempDir)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should install and upgrade in airgap mode", func() {
		installVersion := airgapVersion()
		upgradeVersion := airgapUpgradeVersion()

		installTarball := downloadOrReuseAirgapPackage(
			testEnv.JfrogToken, installVersion, testEnv.TempDir, "E2E_AIRGAP_TARBALL")

		installExtractedDir := airgapInstall(installTarball, installVersion)

		By("Cleaning up install artifacts before upgrade")
		cleanupAirgapArtifacts(installExtractedDir)
		if os.Getenv("E2E_AIRGAP_TARBALL") == "" {
			By(fmt.Sprintf("Removing install tarball: %s", installTarball))
			err := os.Remove(installTarball)
			if err != nil && !os.IsNotExist(err) {
				Expect(err).ToNot(HaveOccurred())
			}
		}

		upgradeTarball := downloadOrReuseAirgapPackage(
			testEnv.JfrogToken, upgradeVersion, testEnv.TempDir, "E2E_AIRGAP_UPGRADE_TARBALL")

		upgradeExtractedDir := airgapUpgrade(upgradeTarball, upgradeVersion)

		By("Cleaning up upgrade artifacts")
		cleanupAirgapArtifacts(upgradeExtractedDir)
		if os.Getenv("E2E_AIRGAP_UPGRADE_TARBALL") == "" {
			By(fmt.Sprintf("Removing upgrade tarball: %s", upgradeTarball))
			err := os.Remove(upgradeTarball)
			if err != nil && !os.IsNotExist(err) {
				Expect(err).ToNot(HaveOccurred())
			}
		}
	})
})
